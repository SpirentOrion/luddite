package luddite

import (
	"encoding/xml"
	"fmt"
)

const (
	// Framework's error codes
	EcodeUnknown               = 0
	EcodeInternal              = 1
	EcodeUnsupportedMediaType  = 2
	EcodeSerializationFailed   = 3
	EcodeDeserializationFailed = 4
	EcodeIdentifierMismatch    = 5

	// Service's error codes
	EcodeApplicationBase = 100
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

func NewError(code int, args ...interface{}) *Error {
	format, ok := errorMessages[code]
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
