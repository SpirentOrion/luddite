package luddite

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	HeaderAccept                 = "Accept"
	HeaderAcceptEncoding         = "Accept-Encoding"
	HeaderAuthorization          = "Authorization"
	HeaderCacheControl           = "Cache-Control"
	HeaderContentDisposition     = "Content-Disposition"
	HeaderContentEncoding        = "Content-Encoding"
	HeaderContentLength          = "Content-Length"
	HeaderContentType            = "Content-Type"
	HeaderETag                   = "ETag"
	HeaderExpect                 = "Expect"
	HeaderForwardedFor           = "X-Forwarded-For"
	HeaderForwardedHost          = "X-Forwarded-Host"
	HeaderIfNoneMatch            = "If-None-Match"
	HeaderLocation               = "Location"
	HeaderRequestId              = "X-Request-Id"
	HeaderSessionId              = "X-Session-Id"
	HeaderSpirentApiVersion      = "X-Spirent-Api-Version"
	HeaderSpirentInhibitResponse = "X-Spirent-Inhibit-Response"
	HeaderSpirentNextLink        = "X-Spirent-Next-Link"
	HeaderSpirentPageSize        = "X-Spirent-Page-Size"
	HeaderSpirentResourceNonce   = "X-Spirent-Resource-Nonce"
	HeaderUserAgent              = "User-Agent"
)

// RequestBearerToken returns the bearer token from an http.Request
// Authorization header. If the header isn't present or it doesn't start with
// the string "Bearer", then an empty string is returned.
func RequestBearerToken(r *http.Request) string {
	if s := r.Header.Get(HeaderAuthorization); len(s) >= 7 && s[:7] == "Bearer " {
		return s[7:]
	}
	return r.URL.Query().Get("access_token")
}

// RequestExternalHost returns the best estimation of the service's host name.
// If a reverse proxy forwarded the request and populated the X-Forwarded-Host
// header, then that value is returned. Otherwise, the Host member from the
// http.Request is returned.
func RequestExternalHost(r *http.Request) string {
	if host := r.Header.Get(HeaderForwardedHost); host != "" {
		return host
	}
	return r.Host
}

// RequestNextLink returns a url.URL value suitable for use in a response header
// as the X-Spirent-Next-Link value. It combines the current http.Request URI
// together with a "next page" cursor value.
func RequestNextLink(r *http.Request, cursor string) *url.URL {
	next := *r.URL
	v := next.Query()
	v.Set("cursor", cursor)
	next.RawQuery = v.Encode()
	return &next
}

// RequestPageSize returns the client's requested page size, defaulting to
// math.MaxInt32 in cases where the X-Spirent-Page-Size header wasn't included
// in the original request or when the header's value is <= 0.
func RequestPageSize(r *http.Request) (pageSize int) {
	var err error
	if pageSize, err = strconv.Atoi(r.Header.Get(HeaderSpirentPageSize)); err != nil || pageSize <= 0 {
		pageSize = math.MaxInt32
	}
	return
}

// RequestQueryCursor returns the "cursor" query string value from the
// http.Request.
func RequestQueryCursor(r *http.Request) string {
	return r.URL.Query().Get("cursor")
}

// RequestResourceNonce returns the X-Spirent-Resource-Nonce header value from
// the http.Request.
func RequestResourceNonce(r *http.Request) string {
	return r.Header.Get(HeaderSpirentResourceNonce)
}

// SetHeader sets the header key with sanitized value to a http response writer
func SetHeader(rw http.ResponseWriter, key string, value string) {
	// remove /r/n(CRLF) to avoid http response splitting
	sanitizedString := sanitizeString(value)
	rw.Header().Set(key, sanitizedString)
}

// AddHeader adds the header key with sanitized value to a http response writer
func AddHeader(rw http.ResponseWriter, key string, value string) {
	// remove /r/n(CRLF) to avoid http response splitting
	sanitizedString := sanitizeString(value)
	rw.Header().Add(key, sanitizedString)
}

// sanitizeString removes "/r/n"(CRLF) to avoid http response splitting
func sanitizeString(value string) (sanitizedString string) {
	sanitizedString = strings.ReplaceAll(value, "\r", "")
	sanitizedString = strings.ReplaceAll(sanitizedString, "\n", "")
	return
}
