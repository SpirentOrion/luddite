package luddite

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"golang.org/x/net/context"
)

// Recovery is middleware that recovers from panics and sends a 500 error response containing an optional stack trace.
type Recovery struct {
	Logger        *log.Logger
	StackAll      bool
	StackSize     int
	StacksVisible bool
}

// Verify that Recovery implements Handler.
var _ Handler = &Recovery{}

// NewRecovery returns a new Recovery instance.
func NewRecovery() *Recovery {
	return &Recovery{
		StackAll:      false,
		StackSize:     1024 * 8,
		StacksVisible: true,
	}
}

func (rec *Recovery) HandleHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request, next ContextHandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]

			if rec.Logger != nil {
				rec.Logger.Printf("PANIC: %s\n%s", err, stack)
			}

			resp := NewError(nil, EcodeInternal, err)
			if rec.StacksVisible {
				resp.Stack = fmt.Sprintf("%s\n%s", err, stack)
			}
			writeResponse(rw, http.StatusInternalServerError, resp)
		}
	}()

	next(ctx, rw, req)
}
