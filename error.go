package luddite

import "encoding/xml"

// ErrorResponse is a structured error response that is returned as the body in all 4xx and 5xx responses.
type ErrorResponse struct {
	XMLName xml.Name `json:"-" xml:"error"`
	Code    int      `json:"code" xml:"code"`
	Message string   `json:"message" xml:"message"`
	Stack   string   `json:"stack,omitempty" xml:"stack,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}
