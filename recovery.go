package luddite

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
)

// Recovery is middleware handler that recovers from panics and sends a 500 error response containing an optional stack trace.
type Recovery struct {
	Logger       *log.Logger
	StackVisible bool
	StackAll     bool
	StackSize    int
}

// NewRecovery returns a new Recovery instance.
func NewRecovery() *Recovery {
	return &Recovery{
		Logger:       log.New(os.Stderr, "[negroni] ", 0),
		StackVisible: true,
		StackAll:     false,
		StackSize:    1024 * 8,
	}
}

func (rec *Recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]

			rec.Logger.Printf("PANIC: %s\n%s", err, stack)

			resp := &ErrorResponse{Code: -1, Message: fmt.Sprint(err)}
			if rec.StackVisible {
				resp.Stack = fmt.Sprintf("%s\n%s", err, stack)
			}
			writeResponse(rw, http.StatusInternalServerError, resp)
		}
	}()

	next(rw, r)
}
