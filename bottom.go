package luddite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/rs/cors"
	"gopkg.in/SpirentOrion/trace.v2"
)

const MAX_STACK_SIZE = 8 * 1024

// Bottom is the bottom-most middleware layer that combines CORS,
// tracing, logging, metrics and recovery actions. Tracing generates a
// unique request id and optionally records traces to a persistent
// backend.  Logging logs requests/responses in a structured JSON
// format.  Metrics increments basic request/response stats. Recovery
// handles panics that occur in HTTP method handlers and optionally
// includes stack traces in 500 responses.
type Bottom struct {
	s             Service
	ctx           context.Context
	defaultLogger *log.Logger
	accessLogger  *log.Logger
	respStacks    bool
	respStackSize int
	cors          *cors.Cors
}

// NewBottom returns a new Bottom instance.
func NewBottom(s Service, defaultLogger, accessLogger *log.Logger) *Bottom {
	config := s.Config()

	b := &Bottom{
		s:             s,
		ctx:           context.Background(),
		defaultLogger: defaultLogger,
		accessLogger:  accessLogger,
		respStacks:    config.Debug.Stacks,
		respStackSize: config.Debug.StackSize,
	}

	if b.respStacks && b.respStackSize < 1 {
		b.respStackSize = MAX_STACK_SIZE
	}

	if config.Cors.Enabled {
		// Enable CORS
		corsOptions := cors.Options{
			AllowedOrigins:   config.Cors.AllowedOrigins,
			AllowedMethods:   config.Cors.AllowedMethods,
			AllowedHeaders:   config.Cors.AllowedHeaders,
			ExposedHeaders:   config.Cors.ExposedHeaders,
			AllowCredentials: config.Cors.AllowCredentials,
		}
		if len(corsOptions.AllowedMethods) == 0 {
			corsOptions.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE"}
		}
		b.cors = cors.New(corsOptions)
	}

	if config.Trace.Enabled {
		// Enable trace recording
		var (
			rec trace.Recorder
			err error
		)
		if rec = recorders[config.Trace.Recorder]; rec == nil {
			// Automatically create JSON and YAML recorders if they are not otherwise registered
			switch config.Trace.Recorder {
			case "json":
				if p := config.Trace.Params["path"]; p != "" {
					var f *os.File
					if f, err = os.OpenFile(p, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err != nil {
						break
					}
					rec = trace.NewJSONRecorder(f)
				} else {
					err = errors.New("JSON trace recorders require a 'path' parameter")
				}
			case "yaml":
				if p := config.Trace.Params["path"]; p != "" {
					var f *os.File
					if f, err = os.OpenFile(p, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err != nil {
						break
					}
					rec = &yamlRecorder{f}
				} else {
					err = errors.New("YAML trace recorders require a 'path' parameter")
				}
			default:
				err = fmt.Errorf("unknown trace recorder: %s", config.Trace.Recorder)
			}
		}
		if rec != nil {
			ctx := trace.WithBuffer(b.ctx, config.Trace.Buffer)
			ctx = trace.WithLogger(ctx, defaultLogger)
			if ctx, err = trace.Record(ctx, rec); err == nil {
				b.ctx = ctx
			}
		}
		if err != nil {
			defaultLogger.Warn("trace recording is not active: ", err)
		}
	}

	return b
}

func (b *Bottom) HandleHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// Start duration measurement ASAP
	start := time.Now()

	// Don't allow panics to escape the bottom handler under any circumstances!
	defer func() {
		if rcv := recover(); rcv != nil {
			stack := make([]byte, MAX_STACK_SIZE)
			stack = stack[:runtime.Stack(stack, false)]
			b.defaultLogger.WithFields(log.Fields{
				"stack": string(stack),
			}).Error(rcv)
		}
	}()

	// Handle CORS prior to tracing
	if b.cors != nil {
		b.cors.HandlerFunc(rw, req)
		if req.Method == "OPTIONS" {
			return
		}
	}

	// Join the request's context to the handler's context where trace
	// recording may be enabled
	ctx0, err := trace.Join(req.Context(), b.ctx)
	if err != nil {
		ctx0 = req.Context()
	}

	// Trace using either using an existing trace id (recovered from the
	// X-Request-Id header in the form "traceId:parentId") or a newly
	// generated one
	var traceId, parentId int64
	if hdr := req.Header.Get(HeaderRequestId); hdr != "" {
		parts := strings.Split(hdr, ":")
		if len(parts) == 2 {
			traceId, _ = strconv.ParseInt(parts[0], 10, 64)
			parentId, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}
	if traceId > 0 && parentId > 0 {
		ctx0 = trace.WithTraceID(trace.WithParentID(ctx0, parentId), traceId)
	} else {
		traceId, _ = trace.GenerateID(ctx0)
		ctx0 = trace.WithTraceID(ctx0, traceId)
	}
	requestId := fmt.Sprint(traceId)
	rw.Header().Set(HeaderRequestId, requestId)

	// Also include our own handler details in the context. Note: We do this
	// in the bottom middleware to avoid having to make multiple shallow
	// copies of the HTTP request. Other handler details may be populated by
	// downstream handlers.
	ctx0 = withHandlerDetails(ctx0, &handlerDetails{
		s:          b.s,
		requestId:  requestId,
		sessionId:  rw.Header().Get(HeaderSessionId),
		request:    req,
		respWriter: rw,
	})

	// Execute the next HTTP handler in a trace span
	trace.Do(ctx0, TraceKindRequest, req.URL.Path, func(ctx1 context.Context) {
		b.handleHTTP(rw.(ResponseWriter), req.WithContext(ctx1), next, start)
	})
}

func (b *Bottom) handleHTTP(res ResponseWriter, req *http.Request, next http.HandlerFunc, start time.Time) {
	defer func() {
		var (
			latency = time.Now().Sub(start)
			status  = res.Status()
			rcv     interface{}
			stack   string
		)

		// If a panic occurs in a downstream handler generate a fail-safe response
		if rcv = recover(); rcv != nil {
			var resp *Error
			if err, ok := rcv.(error); ok && err == context.Canceled {
				// Context cancelation is not an error: use the 418 status as a log marker
				status = http.StatusTeapot
			} else {
				// Unhandled error: return a 500 response
				stackBuffer := make([]byte, MAX_STACK_SIZE)
				stack = string(stackBuffer[:runtime.Stack(stackBuffer, false)])
				b.defaultLogger.WithFields(log.Fields{"stack": stack}).Error(rcv)

				resp = NewError(nil, EcodeInternal, rcv)
				if b.respStacks {
					if len(stack) > b.respStackSize {
						resp.Stack = stack[:b.respStackSize]
					} else {
						resp.Stack = stack
					}
				}
				status = http.StatusInternalServerError
			}
			WriteResponse(res, status, resp)
		}

		// Log the request
		fields := log.Fields{
			"client_addr":   req.RemoteAddr,
			"forwarded_for": req.Header.Get(HeaderForwardedFor),
			"proto":         req.Proto,
			"method":        req.Method,
			"uri":           req.RequestURI,
			"status_code":   status,
			"size":          res.Size(),
			"user_agent":    req.UserAgent(),
			"request_id":    res.Header().Get(HeaderRequestId),
			"api_version":   res.Header().Get(HeaderSpirentApiVersion),
			"time_duration": fmt.Sprintf("%.3f", latency.Seconds()*1000),
		}
		if sessionId := res.Header().Get(HeaderSessionId); sessionId != "" {
			fields["session_id"] = sessionId
		}
		entry := b.accessLogger.WithFields(fields)
		if status/100 != 5 {
			entry.Info()
		} else {
			entry.Error()
		}

		// Annotate the trace
		ctx := req.Context()
		if data := trace.Annotate(ctx); data != nil {
			data["req_method"] = req.Method
			data["req_state"] = ContextRequestState(ctx)
			data["resp_status"] = res.Status()
			data["resp_size"] = res.Size()
			if req.URL.RawQuery != "" {
				data["query"] = req.URL.RawQuery
			}
			if rcv != nil {
				data["panic"] = rcv
				data["stack"] = stack
			}
		}
	}()

	// Invoke the next handler
	next(res, req)
}
