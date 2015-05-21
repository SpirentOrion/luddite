package luddite

import (
	"net/http"
	"strconv"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite/stats"
	"golang.org/x/net/context"
)

type contextKeyT int

const (
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

// ContextLogger returns the Service's logger instance value from a
// context.Context, if possible.
func ContextLogger(ctx context.Context) (logger *log.Entry) {
	if s, _ := ctx.Value(contextServiceKey).(Service); s != nil {
		logger = s.Logger()
	} else {
		logger = log.NewEntry(log.New())
	}
	return
}

// ContextStats returns the Service's stats instance value from a
// context.Context, if possible.
func ContextStats(ctx context.Context) (stats stats.Stats) {
	if s, _ := ctx.Value(contextServiceKey).(Service); s != nil {
		stats = s.Stats()
	} else {
		stats = nullStats
	}
	return
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

// Context is middleware that
type Context struct {
	s          Service
	minVersion int
	maxVersion int
}

// NewContext returns a new Context instance.
func NewContext(s Service, minVersion, maxVersion int) *Context {
	return &Context{s, minVersion, maxVersion}
}

func (c *Context) HandleHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc) {
	defaultVersion := c.maxVersion

	// Range check the requested API version and reject requests that fall outside supported version numbers
	version := RequestApiVersion(r, defaultVersion)
	if version < c.minVersion {
		e := NewError(nil, EcodeApiVersionTooOld, c.minVersion)
		writeResponse(rw, http.StatusGone, e)
		return
	}
	if version > c.maxVersion {
		e := NewError(nil, EcodeApiVersionTooNew, c.maxVersion)
		writeResponse(rw, http.StatusNotImplemented, e)
		return
	}

	// Add the requested API version to response headers (useful for clients when a default version was negotiated)
	rw.Header().Add(HeaderSpirentApiVersion, strconv.Itoa(version))

	// Pass the service and API version to downstream handlers via context
	ctx = WithService(ctx, c.s)
	ctx = WithApiVersion(ctx, version)
	next(ctx, rw, r)
}
