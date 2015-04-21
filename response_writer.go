package luddite

import "net/http"

// ResponseWriter is a wrapper around http.ResponseWriter that
// provides extra information about the response.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher

	// Status returns the status code of the response or 0 if the
	// response has not been written.
	Status() int

	// Written returns whether or not the ResponseWriter has been
	// written.
	Written() bool

	// Size returns the size of the response body.
	Size() int
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// NewResponseWriter creates a ResponseWriter that wraps an
// http.ResponseWriter.
func NewResponseWriter(rw http.ResponseWriter) ResponseWriter {
	return &responseWriter{rw, 0, 0}
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
	rw.size += size
	return size, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}
