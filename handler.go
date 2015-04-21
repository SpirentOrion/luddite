package luddite

import (
	"net/http"

	"golang.org/x/net/context"
)

// ContextHandlerFunc is the context-aware equivalent of http.HandlerFunc.
type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// Handler is an interface that objects can implement to be registered
// to serve as middleware in a service's middleware stack.  HandleHTTP
// should yield to the next middleware in the chain by invoking the
// next ContextHandlerFunc passed in.
//
// If the Handler writes to the ResponseWriter, the next
// ContextHandlerFunc should not be invoked.
type Handler interface {
	HandleHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as
// middleware handlers.  If f is a function with the appropriate
// signature, HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc)

func (h HandlerFunc) HandleHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc) {
	h(ctx, rw, r, next)
}

// WrapHttpHandler converts a non-context-aware http.Handler into a Handler.
func WrapHttpHandler(h http.Handler) Handler {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc) {
		h.ServeHTTP(rw, r)
		next(ctx, rw, r)
	})
}

// WrapMiddlewareHandler converts a non-context-aware http.Handler into a Handler.
func WrapMiddlewareHandler(h func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)) Handler {
	return HandlerFunc(func(ctx context.Context, rw0 http.ResponseWriter, r0 *http.Request, next ContextHandlerFunc) {
		h(rw0, r0, func(rw1 http.ResponseWriter, r1 *http.Request) {
			next(ctx, rw1, r1)
		})
	})
}
