package luddite

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type contextKey int

const contextHandlerDetailsKey = contextKey(0)

// NB: New fields added to this structure must be explicitly initialized in the
// init method below. This enables pool-based allocation.
type handlerDetails struct {
	s               *Service
	rw              ResponseWriter
	request         *http.Request
	requestId       string
	requestProgress string
	apiVersion      int
	callerId        string
	skipInfoLog     bool
	details         map[interface{}]interface{}
}

func (d *handlerDetails) init(s *Service, rw ResponseWriter, request *http.Request, requestId, requestProgress string) {
	d.s = s
	d.rw = rw
	d.request = request
	d.requestId = requestId
	d.requestProgress = requestProgress
	d.apiVersion = 0
	d.callerId = ""
	d.skipInfoLog = false
	d.details = nil
}

func withHandlerDetails(ctx context.Context, d *handlerDetails) context.Context {
	return context.WithValue(ctx, contextHandlerDetailsKey, d)
}

func contextHandlerDetails(ctx context.Context) (d *handlerDetails) {
	d, _ = ctx.Value(contextHandlerDetailsKey).(*handlerDetails)
	return
}

// ContextService returns the Service instance value from a
// context.Context, if possible.
func ContextService(ctx context.Context) (s *Service) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		s = d.s
	}
	return
}

// ContextLogger returns the Service's logger instance value from a
// context.Context, if possible.
func ContextLogger(ctx context.Context) (logger *log.Logger) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		logger = d.s.Logger()
	} else {
		logger = log.New()
	}
	return
}

// ContextResponseWriter returns the current HTTP request's ResponseWriter from
// a context.Context, if possible.
func ContextResponseWriter(ctx context.Context) (rw ResponseWriter) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		rw, _ = d.rw.(ResponseWriter)
	}
	return
}

// ContextResponseHeaders returns the current HTTP response's header collection from
// a context.Context, if possible.
func ContextResponseHeaders(ctx context.Context) (header http.Header) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		header = d.rw.Header()
	}
	return
}

// ContextRequest returns the current HTTP request from a context.Context, if
// possible.
func ContextRequest(ctx context.Context) (request *http.Request) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		request = d.request
	}
	return
}

// ContextRequestId returns the current HTTP request's ID value from a
// context.Context, if possible.
func ContextRequestId(ctx context.Context) (requestId string) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		requestId = d.requestId
	}
	return
}

// ContextSessionId returns the current HTTP request's session ID value from a
// context.Context, if possible.
func ContextSessionId(ctx context.Context) (sessionId string) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		sessionId = d.request.Header.Get(HeaderSessionId)
	}
	return
}

// ContextRequestProgress returns the current HTTP request's progress trace from
// a context.Context, if possible.
func ContextRequestProgress(ctx context.Context) (reqProgress string) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		reqProgress = d.requestProgress
	}
	return
}

// SetContextRequestProgress sets the current HTTP request's progress trace in
// a context.Context.
func SetContextRequestProgress(ctx context.Context, progress string) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		d.requestProgress = progress
	}
}

// ContextApiVersion returns the current HTTP request's API version value from a
// context.Context, if possible.
func ContextApiVersion(ctx context.Context) (apiVersion int) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		apiVersion = d.apiVersion
	}
	return
}

// SetCallerId sets a service-specific caller id string in a context.Context. If
// set, this caller id will appear in the request's access log entry and trace
// data.
func SetContextCallerId(ctx context.Context, callerId string) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		d.callerId = callerId
	}
}

// SetContextSkipInfoLog sets a request-specific flag in a context.Context. If
// set, the request's access log entry and trace data will be skipped at
// InfoLevel if request was successful. Errors will always be logged.
func SetContextSkipInfoLog(ctx context.Context) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		d.skipInfoLog = true
	}
}

// SetContextDetail sets a detail in the current HTTP request's context. This
// may be used by the service's own middleware and avoids allocating a new
// request with additional context.
func SetContextDetail(ctx context.Context, key, value interface{}) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		if d.details == nil {
			d.details = make(map[interface{}]interface{})
		}
		d.details[key] = value
	}
}

// ContextDetail gets a detail from the current HTTP request's context, if
// possible.
func ContextDetail(ctx context.Context, key interface{}) (value interface{}) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok && d.details != nil {
		value = d.details[key]
	}
	return
}
