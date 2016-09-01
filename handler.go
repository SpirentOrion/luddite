package luddite

import "net/http"

// Handler is an interface that objects can implement to be registered to serve
// as middleware in a service's middleware stack. HandleHTTP should yield to the
// next middleware in the chain by invoking the next http.HandlerFunc passed in.
//
// If the Handler writes to the ResponseWriter, the next http.HandlerFunc should
// not be invoked.
type Handler interface {
	HandleHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

// HandlerFunc is an adapter to allow the use of ordinary functions as
// middleware handlers. If f is a function with the appropriate signature,
// HandlerFunc(f) is a Handler object that calls f.
type HandlerFunc func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

func (h HandlerFunc) HandleHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	h(rw, r, next)
}

// WrapHttpHandler converts an http.Handler into a Handler.
func WrapHttpHandler(h http.Handler) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		h.ServeHTTP(rw, r)
		next(rw, r)
	})
}
