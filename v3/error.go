package luddite

import (
	"encoding/xml"
	"fmt"
)

const (
	EcodeUnknown               = "UNKNOWN_ERROR"
	EcodeInternal              = "INTERNAL_ERROR"
	EcodeUnsupportedMediaType  = "UNSUPPORTED_MEDIA_TYPE"
	EcodeSerializationFailed   = "SERIALIZATION_FAILED"
	EcodeDeserializationFailed = "DESERIALIZATION_FAILED"
	EcodeResourceIdMismatch    = "RESOURCE_ID_MISMATCH"
	EcodeApiVersionInvalid     = "API_VERSION_INVALID"
	EcodeApiVersionTooOld      = "API_VERSION_TOO_OLD"
	EcodeApiVersionTooNew      = "API_VERSION_TOO_NEW"
	EcodeValidationFailed      = "VALIDATION_FAILED"
	EcodeLocked                = "LOCKED"
	EcodeUpdatePreempted       = "UPDATE_PREEMPTED"
	EcodeInvalidViewName       = "INVALID_VIEW_NAME"
	EcodeMissingViewParameter  = "MISSING_VIEW_PARAMETER"
	EcodeInvalidViewParameter  = "INVALID_VIEW_PARAMETER"
	EcodeInvalidParameterValue = "INVALID_PARAMETER_VALUE"
)

var commonErrorMap = map[string]string{
	EcodeUnknown:               "Unknown error: %d",
	EcodeInternal:              "Internal error: %v",
	EcodeUnsupportedMediaType:  "Unsupported media type: %s",
	EcodeSerializationFailed:   "Serialization failed: %s",
	EcodeDeserializationFailed: "Deserialization failed: %s",
	EcodeResourceIdMismatch:    "Resource identifier in URL doesn't match value in body",
	EcodeApiVersionInvalid:     "API versions are positive integers",
	EcodeApiVersionTooOld:      "The minimum supported API version number is %d",
	EcodeApiVersionTooNew:      "The maximum supported API version number is %d",
	EcodeValidationFailed:      "Validation failed: %s",
	EcodeLocked:                "Lock error: %s",
	EcodeUpdatePreempted:       "Update was preempted: %s",
	EcodeInvalidViewName:       "Invalid view name",
	EcodeMissingViewParameter:  "Missing view parameter: %s",
	EcodeInvalidViewParameter:  "Invalid view parameter: %s",
	EcodeInvalidParameterValue: "Invalid parameter value: %s -> %s",
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

// NewError allocates and initializes an Error. If a non-nil errorMap
// map is passed, the error is built using this map. Otherwise a map
// containing common errors is used as a fallback.
func NewError(errorMap map[string]string, code string, args ...interface{}) *Error {
	var (
		format string
		ok     bool
	)

	// Lookup an error format string: first try the caller provided
	// error map with fallback to the common error map.
	if errorMap != nil {
		format, ok = errorMap[code]
	}

	if !ok {
		format, ok = commonErrorMap[code]
	}

	// If no error definition could be found, failsafe by using a
	// known-good common error along with the caller's error code.
	if !ok {
		args = []interface{}{code}
		format = commonErrorMap[EcodeUnknown]
		code = EcodeUnknown
	}

	// Optionally format the error message
	var message string
	if format != "" && len(args) != 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	return &Error{
		Code:    code,
		Message: message,
	}
}
