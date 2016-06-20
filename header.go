package luddite

import (
	"net/http"
	"net/url"
	"strconv"
)

const (
	HeaderAccept               = "Accept"
	HeaderAuthorization        = "Authorization"
	HeaderContentDisposition   = "Content-Disposition"
	HeaderContentLength        = "Content-Length"
	HeaderContentType          = "Content-Type"
	HeaderETag                 = "ETag"
	HeaderExpect               = "Expect"
	HeaderForwardedFor         = "X-Forwarded-For"
	HeaderIfNoneMatch          = "If-None-Match"
	HeaderLocation             = "Location"
	HeaderRequestId            = "X-Request-Id"
	HeaderSpirentApiVersion    = "X-Spirent-Api-Version"
	HeaderSpirentInhibitPaging = "X-Spirent-Inhibit-Paging"
	HeaderSpirentMetaPrefix    = "X-Spirent-Meta-"
	HeaderSpirentNextLink      = "X-Spirent-Next-Link"
	HeaderSpirentResourceNonce = "X-Spirent-Resource-Nonce"

	CursorNever = "never"
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
	if host := r.Header.Get(HeaderForwardedFor); host != "" {
		return host
	}
	return r.Host
}

func RequestQueryCursor(r *http.Request) string {
	if _, ok := r.Header[HeaderSpirentInhibitPaging]; ok {
		return CursorNever
	}
	return r.URL.Query().Get("cursor")
}

func RequestNextLink(r *http.Request, cursor string) *url.URL {
	next := *r.URL
	v := next.Query()
	v.Set("cursor", cursor)
	next.RawQuery = v.Encode()
	return &next
}

func RequestResourceNonce(r *http.Request) string {
	return r.Header.Get(HeaderSpirentResourceNonce)
}
