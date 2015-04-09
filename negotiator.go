package luddite

import (
	"net/http"

	"github.com/K-Phoen/negotiation"
)

// FormatNegotiator is a middleware handler that handles Content-Type negotiation.
type FormatNegotiator struct {
	acceptedFormats []string
}

// RegisterFormat registers a new format and associated MIME types.
func RegisterFormat(format string, mimeTypes []string) {
	negotiation.RegisterFormat(format, mimeTypes)
}

// NewNegotiator returns a new FormatNegotiator instance.
func NewNegotiator(acceptedFormats []string) *FormatNegotiator {
	return &FormatNegotiator{
		acceptedFormats: acceptedFormats,
	}
}

func (n *FormatNegotiator) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// If no Accept header was included, default to the first accepted format
	accept := req.Header.Get("Accept")
	if accept == "" {
		accept = n.acceptedFormats[0]
	}

	// Negotiate and set a Content-Type
	format, err := negotiation.NegotiateAccept(accept, n.acceptedFormats)
	if err != nil {
		rw.WriteHeader(http.StatusNotAcceptable)
		return
	}

	rw.Header().Set("Content-Type", format.Value)
	next(rw, req)
}
