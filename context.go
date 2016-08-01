package luddite

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/SpirentOrion/logrus"
	"gopkg.in/SpirentOrion/trace.v2"
)

type contextKeyT int

const contextHandlerDetailsKey = contextKeyT(0)

type handlerDetails struct {
	s          Service
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

// ContextRequestId returns the current HTTP request's ID value from a
// context.Context, if possible.
func ContextRequestId(ctx context.Context) (reqId string) {
	if traceId := trace.CurrentTraceID(ctx); traceId != 0 {
		reqId = fmt.Sprint(traceId)
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
