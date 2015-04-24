package luddite

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/SpirentOrion/trace"
	"github.com/SpirentOrion/trace/dynamorec"
	"github.com/SpirentOrion/trace/yamlrec"
	"golang.org/x/net/context"
)

const (
	TraceKindRequest  = "request"
	TraceKindDynamodb = "dynamodb"
)

// Trace is middleware that logs the request as it goes in and the response as it goes out.
type Trace struct {
	*log.Logger
}

// Verify that Trace implements Handler.
var _ Handler = &Trace{}

// NewTrace returns a new Trace instance.
func NewTrace(enabled bool, buffer int, recorder, paramsStr string, logger *log.Logger) (t *Trace, err error) {
	if enabled {
		params := strings.Split(paramsStr, ":")

		var rec trace.Recorder
		switch recorder {
		case "yaml":
			if len(params) != 1 {
				err = errors.New("yaml trace recorder expects 1 parameter (path)")
				return
			}
			var f *os.File
			f, err = os.OpenFile(params[0], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				return
			}
			rec, err = yamlrec.New(f)
			if err != nil {
				f.Close()
				return
			}
			break
		case "dynamodb":
			if len(params) != 4 {
				err = errors.New("dynamodb trace recorder expects 4 parameters (region:table_name:access_key:secret_key)")
				return
			}
			rec, err = dynamorec.New(params[0], params[1], params[2], params[3])
			if err != nil {
				return
			}
			break
		default:
			err = fmt.Errorf("unknown trace recorder: ", recorder)
			return
		}

		if err = trace.Record(rec, buffer, logger); err != nil {
			return
		}

		logger.Println("recording traces to", rec)
	}

	t = &Trace{logger}
	return
}

func (t *Trace) HandleHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request, next ContextHandlerFunc) {
	// Don't allow panics to escape
	defer func() {
		if err := recover(); err != nil {
			if t.Logger != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]
				t.Printf("PANIC: %s\n%s", err, stack)
			}
		}
	}()

	// For now, always honor incoming id headers. If present, header must be in the form "traceId:parentId".
	var traceId, parentId int64
	if true {
		if hdr := req.Header.Get(HeaderRequestId); hdr != "" {
			var traceId, parentId int64
			n, _ := fmt.Sscanf(hdr, "%d:%d", &traceId, &parentId)
			if n < 2 || traceId < 1 || parentId < 1 {
				traceId = 0
				parentId = 0
			}
		}
	}

	// Start a new trace, either using an existing id (from the request header) or a new one
	s, err := trace.New(traceId, TraceKindRequest, fmt.Sprintf("%s %s", req.Method, req.URL.Path))
	if err == nil {
		s.ParentId = parentId

		// Add headers
		req.Header.Set(HeaderRequestId, fmt.Sprintf("%d:%d", s.TraceId, s.SpanId))
		rw.Header().Set(HeaderRequestId, fmt.Sprintf("%d", s.TraceId))

		trace.Run(s, func() {
			// Invoke the next handler
			next(ctx, rw, req)

			// Annotate the trace with additional response data
			res := rw.(ResponseWriter)
			data := s.Data()
			data["resp_status"] = res.Status()
			data["resp_size"] = res.Size()
		})
	} else {
		// Invoke the next handler
		next(ctx, rw, req)
	}
}
