package luddite

import (
	"net/http"
	"strconv"
)

const (
	HeaderAccept            = "Accept"
	HeaderAuthorization     = "Authorization"
	HeaderContentType       = "Content-Type"
	HeaderForwardedFor      = "X-Forwarded-For"
	HeaderLocation          = "Location"
	HeaderRequestId         = "X-Request-Id"
	HeaderSpirentApiVersion = "X-Spirent-Api-Version"
	HeaderSpirentNextLink   = "X-Spirent-Next-Link"
	HeaderSpirentPrevLink   = "X-Spirent-Prev-Link"
)

func RequestBearerToken(r *http.Request) (token string) {
	if authStr := r.Header.Get(HeaderAuthorization); authStr != "" && authStr[:7] == "Bearer " {
		token = authStr[7:]
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
