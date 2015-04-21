package luddite

import (
	"net/http"

	"golang.org/x/net/context"
)

type middleware struct {
	handler Handler
	next    *middleware
}

// Verify that middleware implements http.Handler.
var _ http.Handler = middleware{}

func (m middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.handler.HandleHTTP(context.Background(), NewResponseWriter(rw), r, m.next.dispatchHandler)
}

func (m middleware) dispatchHandler(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	m.handler.HandleHTTP(ctx, rw, r, m.next.dispatchHandler)
}

func voidMiddleware() middleware {
	return middleware{
		handler: HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc) {}),
		next:    &middleware{},
	}
}

func buildMiddleware(handlers []Handler) middleware {
	var next middleware

	if len(handlers) == 0 {
		return voidMiddleware()
	} else if len(handlers) > 1 {
		next = buildMiddleware(handlers[1:])
	} else {
		next = voidMiddleware()
	}

	return middleware{handlers[0], &next}
}
