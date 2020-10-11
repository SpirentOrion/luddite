package luddite

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter is a wrapper around http.ResponseWriter that
// provides extra information about the response.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker

	// Written returns true once the ResponseWriter has been written.
	Written() bool

	// Status returns the status code of the response or 0 if the
	// response has not been written.
	Status() int

	// Size returns the size of the response body or 0 if the response has
	// not been written.
	Size() int64
}

// NB: New fields added to this structure must be explicitly initialized in the
// init method below. This enables pool-based allocation.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (rw *responseWriter) init(base http.ResponseWriter) {
	rw.ResponseWriter = base
	rw.status = 0
	rw.size = 0
}

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int64 {
	return rw.size
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}
