package luddite

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
)

const (
	HeaderAccept               = "Accept"
	HeaderAcceptEncoding       = "Accept-Encoding"
	HeaderAuthorization        = "Authorization"
	HeaderCacheControl         = "Cache-Control"
	HeaderContentDisposition   = "Content-Disposition"
	HeaderContentEncoding      = "Content-Encoding"
	HeaderContentLength        = "Content-Length"
	HeaderContentType          = "Content-Type"
	HeaderETag                 = "ETag"
	HeaderExpect               = "Expect"
	HeaderForwardedFor         = "X-Forwarded-For"
	HeaderForwardedHost        = "X-Forwarded-Host"
	HeaderIfNoneMatch          = "If-None-Match"
	HeaderLocation             = "Location"
	HeaderRequestId            = "X-Request-Id"
	HeaderSessionId            = "X-Session-Id"
	HeaderSpirentApiVersion    = "X-Spirent-Api-Version"
	HeaderSpirentNextLink      = "X-Spirent-Next-Link"
	HeaderSpirentPageSize      = "X-Spirent-Page-Size"
	HeaderSpirentResourceNonce = "X-Spirent-Resource-Nonce"
	HeaderUserAgent            = "User-Agent"
)

func RequestId(r *http.Request) string {
	return r.Header.Get(HeaderRequestId)
}

func RequestBearerToken(r *http.Request) (token string) {
	if authStr := r.Header.Get(HeaderAuthorization); len(authStr) >= 7 && authStr[:7] == "Bearer " {
		token = authStr[7:]
	} else {
		token = r.URL.Query().Get("access_token")
	}
	return
}

func RequestApiVersion(r *http.Request, defaultVersion int) (version int) {
	version = defaultVersion
	if versionStr := r.Header.Get(HeaderSpirentApiVersion); versionStr != "" {
		if i, err := strconv.Atoi(versionStr); err == nil && i > 0 {
			version = i
		}
	}
	return
}

func RequestExternalHost(r *http.Request) string {
	if host := r.Header.Get(HeaderForwardedHost); host != "" {
		return host
	}
	return r.Host
}

func RequestQueryCursor(r *http.Request) string {
	return r.URL.Query().Get("cursor")
}

func RequestNextLink(r *http.Request, cursor string) *url.URL {
	next := *r.URL
	v := next.Query()
	v.Set("cursor", cursor)
	next.RawQuery = v.Encode()
	return &next
}

func RequestPageSize(r *http.Request) int {
	var (
		pageSize int
		err      error
	)
	if pageSize, err = strconv.Atoi(r.Header.Get(HeaderSpirentPageSize)); err != nil {
		return math.MaxInt32
	}
	return pageSize
}

func RequestResourceNonce(r *http.Request) string {
	return r.Header.Get(HeaderSpirentResourceNonce)
}
