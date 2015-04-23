package luddite

import "golang.org/x/net/context"

type contextKeyT int

var (
	contextServiceKey    = contextKeyT(0)
	contextApiVersionKey = contextKeyT(1)
)

// WithService returns a new context.Context instance with the Service
// included as a value.
func WithService(ctx context.Context, s Service) context.Context {
	return context.WithValue(ctx, contextServiceKey, s)
}

// ContextService returns the Service instance value from a
// context.Context, if possible.
func ContextService(ctx context.Context) Service {
	s, _ := ctx.Value(contextServiceKey).(Service)
	return s
}

// WithApiVersion returns a new context.Context instance with the current HTTP request's
// API version header included as a value.
func WithApiVersion(ctx context.Context, version int) context.Context {
	return context.WithValue(ctx, contextApiVersionKey, version)
}

// ContextApiVersion returns the current HTTP request's API version value from a
// context.Context, if possible.
func ContextApiVersion(ctx context.Context) int {
	version, _ := ctx.Value(contextApiVersionKey).(int)
	return version
}
