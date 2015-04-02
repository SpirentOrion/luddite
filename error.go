package luddite

import (
	"encoding/xml"
	"fmt"
)

const (
	// Common error codes
	EcodeUnknown               = 0
	EcodeInternal              = 1
	EcodeUnsupportedMediaType  = 2
	EcodeSerializationFailed   = 3
	EcodeDeserializationFailed = 4
	EcodeIdentifierMismatch    = 5

	// Service-specific error codes
	EcodeServiceBase = 1024
)

var commonErrorDefs = map[int]ErrorDefinition{
	EcodeUnknown:               {"UNKNOWN_ERROR", "Unknown error: %d"},
	EcodeInternal:              {"INTERNAL_ERROR", "Internal error: %v"},
	EcodeUnsupportedMediaType:  {"UNSUPPORTED_MEDIA_TYPE", "Unsupported media type: %s"},
	EcodeSerializationFailed:   {"SERIALIZATION_FAILED", "Serialization failed: %s"},
	EcodeDeserializationFailed: {"DESERIALIZATION_FAILED", "Deserialization failed: %s"},
	EcodeIdentifierMismatch:    {"RESOURCE_ID_MISMATCH", "Resource identifier in URL doesn't match value in body"},
}

// ErrorDefinition defines a structured error that can be returned in 4xx and 5xx responses.
type ErrorDefinition struct {
	Code   string
	Format string
}

// Error is a transfer object that is serialized as the body in 4xx and 5xx responses.
type Error struct {
	XMLName xml.Name `json:"-" xml:"error"`
	Code    string   `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Stack   string   `json:"stack,omitempty" xml:"stack,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// NewError allocates and initializes an Error. If a non-nil errors
// map is passed, the error is built using this map. Otherwise a map
// containing common errors is used as a fallback.
func NewError(errorDefs map[int]ErrorDefinition, code int, args ...interface{}) *Error {
	var (
		def ErrorDefinition
		ok  bool
	)

	// Lookup an error definition: first try the caller provided
	// error map with fallback to the common error map.
	if errorDefs != nil {
		def, ok = errorDefs[code]
	}

	if !ok {
		def, ok = commonErrorDefs[code]
	}

	// If no error definition could be found, failsafe by using a
	// known-good common error along with the caller's error code.
	if !ok {
		def = commonErrorDefs[EcodeUnknown]
		args = []interface{}{code}
	}

	// Optionally format the error message
	var message string
	if def.Format != "" && len(args) != 0 {
		message = fmt.Sprintf(def.Format, args...)
	} else {
		message = def.Format
	}

	return &Error{
		Code:    def.Code,
		Message: message,
	}
}
