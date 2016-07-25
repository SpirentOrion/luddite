package luddite

import (
	"net/http"
	"strconv"

	log "github.com/SpirentOrion/logrus"
	"golang.org/x/net/context"
)

type contextKeyT int

const contextHandlerDetailsKey = contextKeyT(0)

type handlerDetails struct {
	s          Service
	apiVersion int
	reqId      string
	respWriter http.ResponseWriter
}

func withHandlerDetails(ctx context.Context, d *handlerDetails) context.Context {
	return context.WithValue(ctx, contextHandlerDetailsKey, d)
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
	if d, ok := ctx.Value(contextHandlerDetailsKey).(*handlerDetails); ok {
		reqId = d.reqId
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

// Context is middleware that performs API version selection and enforces the service's
// min/max supported version constraints.  It also makes the Service instance, selected
// API version and request id available.
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
		WriteResponse(rw, http.StatusGone, e)
		return
	}
	if version > c.maxVersion {
		e := NewError(nil, EcodeApiVersionTooNew, c.maxVersion)
		WriteResponse(rw, http.StatusNotImplemented, e)
		return
	}

	// Add the requested API version to response headers (useful for clients when a default version was negotiated)
	rw.Header().Add(HeaderSpirentApiVersion, strconv.Itoa(version))

	// Pass the service, API version, request id, and response writer-related objects to downstream handlers via context
	d := &handlerDetails{
		s:          c.s,
		apiVersion: version,
		reqId:      RequestId(r),
		respWriter: rw,
	}
	next(withHandlerDetails(ctx, d), rw, r)
}
