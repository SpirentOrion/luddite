package luddite

import (
	"net/http"
	"strconv"

	"github.com/K-Phoen/negotiation"
)

var acceptedContentTypes = []string{
	ContentTypeJson,
	ContentTypeCss,
	ContentTypePlain,
	ContentTypeXml,
	ContentTypeHtml,
	ContentTypeGif,
	ContentTypePng,
	ContentTypeOctetStream,
}

// RegisterFormat registers a new format and associated MIME types.
func RegisterFormat(format string, mimeTypes []string) {
	negotiation.RegisterFormat(format, mimeTypes)
}

type negotiatorHandler struct{}

func (n *negotiatorHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// If no Accept header was included, default to the first accepted format
	accept := req.Header.Get(HeaderAccept)
	if accept == "" {
		accept = acceptedContentTypes[0]
	}

	// Negotiate and set a Content-Type
	//
	// Note: Negotation failures do not return 406 errors here. This allows
	// resource handlers to potentially inspect/handle certain rarely-used
	// content types on their own. If a negotiation failure has occurred and
	// the resource handler doesn't deal with it, then we can expect a 406
	// from WriteResponse.
	if format, _ := negotiation.NegotiateAccept(accept, acceptedContentTypes); format != nil {
		SetHeader(rw, HeaderContentType, format.Value)
	}

	// If the X-Spirent-Inhibit-Response header is set and true-ish, then
	// set the same response header. This will cause subsequent calls to
	// WriteResponse to omit the response body for 2xx responses and also
	// makes the behavior obvious to clients (i.e. response header shows
	// intention beyond the 204 status).
	if inhibitResp, _ := strconv.ParseBool(req.Header.Get(HeaderSpirentInhibitResponse)); inhibitResp {
		SetHeader(rw, HeaderSpirentInhibitResponse, "1")
	}

	next(rw, req)
}
