package luddite

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var (
	responseWriterPool = sync.Pool{New: func() interface{} { return new(responseWriter) }}
	handlerDetailsPool = sync.Pool{New: func() interface{} { return new(handlerDetails) }}
)

type bottomHandler struct {
	s             *Service
	defaultLogger *log.Logger
	accessLogger  *log.Logger
	tracerKind    TracerKind
	tracer        opentracing.Tracer
	cors          *cors.Cors
}

func (b *bottomHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	var (
		start = time.Now()
		res   *responseWriter
		d     *handlerDetails
	)

	// Don't allow panics to escape under any circumstances!
	defer func() {
		if rcv := recover(); rcv != nil {
			buf := goroutineStack(false)
			b.defaultLogger.WithField("stack", string(buf)).Error(rcv)
		}
		if res != nil {
			responseWriterPool.Put(res)
		}
		if d != nil {
			handlerDetailsPool.Put(d)
		}
	}()

	// Handle CORS prior to tracing
	if b.cors != nil {
		b.cors.HandlerFunc(rw, req)
		if req.Method == "OPTIONS" {
			return
		}
	}

	// If possible, recover the client's span information from HTTP
	// headers. If this fails, a new root span will be created.
	clientSpan, _ := b.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	serverSpan, ctx := opentracing.StartSpanFromContextWithTracer(req.Context(), b.tracer, "request", ext.RPCServerOption(clientSpan))
	defer serverSpan.Finish()

	requestId := strconv.FormatUint(b.tracerKind.TraceId(serverSpan), 10)
	rw.Header().Set(HeaderRequestId, requestId)

	// Create a new response writer
	res = responseWriterPool.Get().(*responseWriter)
	res.init(rw)

	// Create new handler details and to the request context
	d = handlerDetailsPool.Get().(*handlerDetails)
	d.init(b.s, res, req, requestId, "luddite.bottomHandler.begin")
	ctx = withHandlerDetails(ctx, d)

	// Create a shallow copy of the request so that it references the final
	// context
	req = req.WithContext(ctx)
	d.request = req

	defer func() {
		var (
			latency = time.Since(start)
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
				buf := goroutineStack(false)
				stack = string(buf)
				b.defaultLogger.WithField("stack", stack).Error(rcv)

				resp = NewError(nil, EcodeInternal, rcv)
				if b.s.config.Debug.Stacks {
					resp.Stack = stack
				}
				status = http.StatusInternalServerError
			}
			if !res.Written() {
				_ = WriteResponse(res, status, resp)
			}
		}

		if !d.skipInfoLog || (status >= 400) || (b.accessLogger.Level != log.InfoLevel) {

			// Log the request
			fields := log.Fields{
				"client_addr":   req.RemoteAddr,
				"forwarded_for": req.Header.Get(HeaderForwardedFor),
				"proto":         req.Proto,
				"method":        req.Method,
				"uri":           req.RequestURI,
				"status":        status,
				"size":          res.Size(),
				"user_agent":    req.UserAgent(),
				"request_id":    requestId,
				"api_version":   d.apiVersion,
				"latency":       fmt.Sprintf("%.6f", latency.Seconds()),
			}
			sessionId := req.Header.Get(HeaderSessionId)
			if sessionId != "" {
				fields["session_id"] = sessionId
			}
			callerId := d.callerId
			if callerId != "" {
				fields["caller_id"] = callerId
			}
			entry := b.accessLogger.WithFields(fields)
			if status/100 != 5 {
				entry.Info()
			} else {
				entry.Error()
			}

			// Decorate trace span
			if !b.tracerKind.IsNoop() {
				serverSpan.SetTag("luddite.progress", ContextRequestProgress(ctx))
				serverSpan.SetTag("http.method", req.Method)
				serverSpan.SetTag("http.url", req.URL.String())
				serverSpan.SetTag("http.status_code", res.Status())
				serverSpan.SetTag("http.response_size", res.Size())
				serverSpan.SetTag("http.request_id", requestId)
				serverSpan.SetTag("api_version", d.apiVersion)
				if sessionId != "" {
					serverSpan.SetTag("session_id", sessionId)
				}
				if callerId != "" {
					serverSpan.SetTag("caller_id", callerId)
				}
				if rcv != nil {
					serverSpan.SetTag("panic", rcv)
					serverSpan.SetTag("stack", stack)
				}
			}
		}
	}()

	// Invoke the next handler
	next(res, req)
	SetContextRequestProgress(ctx, "luddite.bottomHandler.end")
}
