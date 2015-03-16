package luddite

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

const (
	// Framework's error codes
	EcodeUnknown               = 0
	EcodeInternal              = 1
	EcodeUnsupportedMediaType  = 2
	EcodeSerializationFailed   = 3
	EcodeDeserializationFailed = 4
	EcodeIdentifierMismatch    = 5

	// Service-specific error codes
	EcodeServiceBase = 100
)

var errorMessages = map[int]string{
	EcodeUnknown:               "Unknown error",
	EcodeInternal:              "Internal error",
	EcodeUnsupportedMediaType:  "Unsupported media type: %s",
	EcodeSerializationFailed:   "Serialization failed: %s",
	EcodeDeserializationFailed: "Deserialization failed: %s",
	EcodeIdentifierMismatch:    "Resource identifier in URL doesn't match value in body",
}

// Error is a structured error that is returned as the body in all 4xx and 5xx responses.
type Error struct {
	XMLName xml.Name `json:"-" xml:"error"`
	Code    int      `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Stack   string   `json:"stack,omitempty" xml:"stack,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(req *http.Request, code int, args ...interface{}) *Error {
	var (
		s      *service
		format string
		ok     bool
	)

	if code >= EcodeServiceBase {
		// Lookup a service-specific error message
		if s, ok = serviceContext(req); ok {
			format, ok = s.errorMessages[code]
		}
	} else {
		// Lookup a framework error message
		format, ok = errorMessages[code]
	}

	// If no message was found, use a safe error message along with the caller's error code
	if !ok {
		format = errorMessages[EcodeUnknown]
		args = nil
	}

	var message string
	if len(args) != 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	return &Error{
		Code:    code,
		Message: message,
	}
}
