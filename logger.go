package luddite

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// Logger is middleware that logs the request as it goes in and the response as it goes out.
type Logger struct {
	*log.Logger
}

// Verify that Logger implements Handler.
var _ Handler = &Logger{}

// NewLogger returns a new Logger instance.
func NewLogger(logger *log.Logger) *Logger {
	return &Logger{logger}
}

func (l *Logger) HandleHTTP(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc) {
	start := time.Now()
	l.Printf("Started %s %s", r.Method, r.URL.Path)

	next(ctx, rw, r)

	res := rw.(ResponseWriter)
	l.Printf("Completed %s %s -> %v %s in %v", r.Method, r.URL.Path, res.Status(), http.StatusText(res.Status()), time.Since(start))
}
