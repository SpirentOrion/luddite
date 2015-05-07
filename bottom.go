package luddite

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/SpirentOrion/luddite/datastore"
	"github.com/SpirentOrion/trace"
	"github.com/SpirentOrion/trace/dynamorec"
	"github.com/SpirentOrion/trace/yamlrec"
	"golang.org/x/net/context"
)

const MAX_STACK_SIZE = 8 * 1024

// Bottom is the bottom-most middleware layer that combines tracing,
// logging, metrics and recovery actions. Tracing generates a unique
// request id and optionally records traces to a persistent backend.
// Logging logs requests/responses in a structured JSON format.
// Metrics increments basic request/response stats. Recovery handles
// panics that occur in HTTP method handlers and optionally includes
// stack traces in 500 responses.
type Bottom struct {
	defaultLogger *log.Entry
	accessLogger  *log.Entry
	stats         Stats
	respStacks    bool
	respStackSize int
}

// NewBottom returns a new Bottom instance.
func NewBottom(config *ServiceConfig, defaultLogger, accessLogger *log.Entry, stats Stats) *Bottom {
	b := &Bottom{
		defaultLogger: defaultLogger,
		accessLogger:  accessLogger,
		stats:         stats,
		respStacks:    config.Debug.Stacks,
		respStackSize: config.Debug.StackSize,
	}

	if b.respStacks && b.respStackSize < 1 {
		b.respStackSize = MAX_STACK_SIZE
	}

	if config.Trace.Enabled {
		// Enable trace recording
		var (
			rec trace.Recorder
			err error
		)
		switch config.Trace.Recorder {
		case datastore.YAML_PROVIDER:
			var p *datastore.YAMLParams
			p, err = datastore.NewYAMLParams(config.Trace.Params)
			if err != nil {
				break
			}
			var f *os.File
			f, err = os.OpenFile(p.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				break
			}
			rec, err = yamlrec.New(f)
			if err != nil {
				f.Close()
			}
		case datastore.DYNAMODB_PROVIDER:
			var p *datastore.DynamoParams
			p, err = datastore.NewDynamoParams(config.Trace.Params)
			if err != nil {
				break
			}
			rec, err = dynamorec.New(p.Region, p.TableName, p.AccessKey, p.SecretKey)
		default:
			err = fmt.Errorf("unknown trace recorder: %s", config.Trace.Recorder)
		}

		if rec != nil {
			err = trace.Record(rec, config.Trace.Buffer, defaultLogger)
		}

		if err != nil {
			defaultLogger.Warn("trace recording is not active: ", err)
		} else {
			defaultLogger.Debug("recording traces to ", rec)
		}
	}

	return b
}

func (b *Bottom) HandleHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request, next ContextHandlerFunc) {
	// Don't allow panics to escape the bottom handler under any circumstances!
	defer func() {
		if rcv := recover(); rcv != nil {
			stack := make([]byte, MAX_STACK_SIZE)
			stack = stack[:runtime.Stack(stack, false)]
			b.defaultLogger.WithFields(log.Fields{
				"error": rcv,
				"stack": string(stack),
			}).Error("PANIC")
		}
	}()

	// For now, always honor incoming id headers. If present, header must be in the form "traceId:parentId"
	traceId, parentId := b.getRequestTraceIds(req)

	// Start a new trace, either using an existing id (from the request header) or a new one
	if traceId == 0 {
		traceId, _ = trace.GenerateId()
	}
	s, _ := trace.New(traceId, TraceKindRequest, req.URL.Path)
	if s != nil {
		b.addRequestResponseTraceIds(rw, req, traceId, s.SpanId)
		s.ParentId = parentId
	} else {
		b.addResponseTraceId(rw, traceId)
	}

	res := rw.(ResponseWriter)
	trace.Run(s, func() {
		// If a panic occurs in a downstream handler, log it, generate a 500 error
		// response, and annotate the trace with additional panic-related info
		defer func() {
			if rcv := recover(); rcv != nil {
				stack := make([]byte, MAX_STACK_SIZE)
				stack = stack[:runtime.Stack(stack, false)]
				b.defaultLogger.WithFields(log.Fields{
					"error": rcv,
					"stack": string(stack),
				}).Error("PANIC")

				resp := NewError(nil, EcodeInternal, rcv)
				if b.respStacks {
					if len(stack) > b.respStackSize {
						resp.Stack = string(stack[:b.respStackSize])
					} else {
						resp.Stack = string(stack)
					}
				}
				writeResponse(rw, http.StatusInternalServerError, resp)

				b.accessLogger.WithFields(log.Fields{
					"client_addr":   req.RemoteAddr,
					"forwarded_for": req.Header.Get(HeaderForwardedFor),
					"proto":         req.Proto,
					"method":        req.Method,
					"uri":           req.RequestURI,
					"status_code":   res.Status(),
					"size":          res.Size(),
					"user_agent":    req.UserAgent(),
				}).Error()

				if s != nil {
					data := s.Data()
					data["panic"] = rcv
					data["stack"] = string(stack)
					data["resp_status"] = res.Status()
					data["resp_size"] = res.Size()
				}
			}
		}()

		// Invoke the next handler
		next(ctx, rw, req)

		// Log the request
		status := res.Status()
		entry := b.accessLogger.WithFields(log.Fields{
			"client_addr":   req.RemoteAddr,
			"forwarded_for": req.Header.Get(HeaderForwardedFor),
			"proto":         req.Proto,
			"method":        req.Method,
			"uri":           req.RequestURI,
			"status_code":   status,
			"size":          res.Size(),
			"user_agent":    req.UserAgent(),
		})
		if status/100 != 5 {
			entry.Info()
		} else {
			entry.Error()
		}

		// Annotate the trace
		if s != nil {
			data := s.Data()
			data["req_method"] = req.Method
			data["resp_status"] = res.Status()
			data["resp_size"] = res.Size()
		}
	})

	// Update request/response metrics
	stat := fmt.Sprintf("request.http.%s%s", strings.ToLower(req.Method), strings.Replace(req.URL.Path, "/", ".", -1))
	b.stats.Incr(stat, 1)
	stat = fmt.Sprintf("response.http.%s%s.%d", strings.ToLower(req.Method), strings.Replace(req.URL.Path, "/", ".", -1), res.Status())
	b.stats.Incr(stat, 1)
}

func (b *Bottom) getRequestTraceIds(req *http.Request) (traceId, parentId int64) {
	if hdr := req.Header.Get(HeaderRequestId); hdr != "" {
		n, _ := fmt.Sscanf(hdr, "%d:%d", &traceId, &parentId)
		if n < 2 || traceId < 1 || parentId < 1 {
			traceId = 0
			parentId = 0
		}
	}
	return
}

func (b *Bottom) addRequestResponseTraceIds(rw http.ResponseWriter, req *http.Request, traceId, parentId int64) {
	req.Header.Set(HeaderRequestId, fmt.Sprintf("%d:%d", traceId, parentId))
	rw.Header().Set(HeaderRequestId, fmt.Sprint(traceId))
}

func (b *Bottom) addResponseTraceId(rw http.ResponseWriter, traceId int64) {
	rw.Header().Set(HeaderRequestId, fmt.Sprint(traceId))
}
