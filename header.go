package luddite

import (
	"net/http"
	"strconv"
)

const (
	HeaderAccept            = "Accept"
	HeaderContentType       = "Content-Type"
	HeaderLocation          = "Location"
	HeaderSpirentApiVersion = "X-Spirent-Api-Version"
)

func RequestApiVersion(r *http.Request, defaultVersion int) (version int) {
	version = defaultVersion
	if versionStr := r.Header.Get(HeaderSpirentApiVersion); versionStr != "" {
		if i, err := strconv.Atoi(versionStr); err == nil && i > 0 {
			version = i
		}
	}
	return
}
