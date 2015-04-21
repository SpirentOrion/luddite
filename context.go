package luddite

import "golang.org/x/net/context"

type contextKeyT int

var contextServiceKey = contextKeyT(0)

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
