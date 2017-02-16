package luddite

import (
	"context"
	"net/http"

	log "github.com/SpirentOrion/logrus"
)

type contextKeyT int

const contextHandlerDetailsKey = contextKeyT(0)

type handlerDetails struct {
	s          Service
	request    *http.Request
	requestId  string
	sessionId  string
	apiVersion int
	respWriter http.ResponseWriter
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
func ContextService(ctx context.Context) (s Service) {
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

// ContextApiVersion returns the current HTTP request's API version value from a
// context.Context, if possible.
func ContextApiVersion(ctx context.Context) (apiVersion int) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		apiVersion = d.apiVersion
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
		sessionId = d.sessionId
	}
	return
}

// ContextResponseHeaders returns the current HTTP response's header collection from
// a context.Context, if possible.
func ContextResponseHeaders(ctx context.Context) (respHeaders http.Header) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		respHeaders = d.respWriter.Header()
	}
	return
}

// ContextCloseNotify returns a channel that receives at most a single value
// (true) when the client connection has gone away, if possible.
func ContextCloseNotify(ctx context.Context) (closeNotify <-chan bool) {
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		closeNotify = d.respWriter.(http.CloseNotifier).CloseNotify()
	}
	return
}
