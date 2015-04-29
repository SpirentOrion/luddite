package luddite

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/SpirentOrion/luddite/datastore"
	"github.com/SpirentOrion/trace"
	"github.com/SpirentOrion/trace/dynamorec"
	"github.com/SpirentOrion/trace/yamlrec"
	"github.com/quipo/statsd"
	"golang.org/x/net/context"
)

const MAX_STACK_SIZE = 8 * 1024

// Bottom is middleware that logs the request as it goes in and the response as it goes out.
type Bottom struct {
	logger        *log.Logger
	stats         *statsd.StatsdBuffer
	logRequests   bool
	respStacks    bool
	respStackSize int
}

// NewBottom returns a new Bottom instance.
func NewBottom(config *ServiceConfig, logger *log.Logger, stats *statsd.StatsdBuffer) *Bottom {
	b := &Bottom{
		logger:        logger,
		stats:         stats,
		logRequests:   config.Debug.Requests,
		respStacks:    config.Debug.Stacks,
		respStackSize: config.Debug.StackSize,
	}

	if b.respStacks && b.respStackSize < 1 {
		b.respStackSize = MAX_STACK_SIZE
	}

	if config.Trace.Enabled {
		// Enable trace recording
		var (
			rec trace.Recorder
			err error
		)
		switch config.Trace.Recorder {
		case "yaml":
			var p *datastore.YAMLParams
			p, err = datastore.NewYAMLParams(config.Trace.Params)
			if err != nil {
				break
			}
			var f *os.File
			f, err = os.OpenFile(p.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				break
			}
			rec, err = yamlrec.New(f)
			if err != nil {
				f.Close()
				break
			}
			break
		case "dynamodb":
			var p *datastore.DynamoParams
			p, err = datastore.NewDynamoParams(config.Trace.Params)
			if err != nil {
				break
			}
			rec, err = dynamorec.New(p.Region, p.TableName, p.AccessKey, p.SecretKey)
			break
		default:
			err = fmt.Errorf("unknown trace recorder: ", config.Trace.Recorder)
			break
		}

		if rec != nil {
			err = trace.Record(rec, config.Trace.Buffer, logger)
		}

		if err != nil {
			logger.Println("trace recording is not active:", err)
		} else {
			logger.Println("recording traces to", rec)
		}
	}

	return b
}

func (b *Bottom) HandleHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request, next ContextHandlerFunc) {
	// Don't allow panics to escape the bottom handler under any circumstances!
	defer func() {
		if rcv := recover(); rcv != nil {
			if b.logger != nil {
				stack := make([]byte, MAX_STACK_SIZE)
				stack = stack[:runtime.Stack(stack, false)]
				b.logger.Printf("PANIC: %s\n%s", rcv, stack)
			}
		}
	}()

	// For now, always honor incoming id headers. If present, header must be in the form "traceId:parentId"
	traceId, parentId := b.getRequestTraceIds(req)

	// Start a new trace, either using an existing id (from the request header) or a new one
	s, _ := trace.New(traceId, TraceKindRequest, req.URL.Path)
	if s != nil {
		b.addRequestResponseTraceIds(rw, req, traceId, s.SpanId)
		s.ParentId = parentId
	}

	res := rw.(ResponseWriter)
	trace.Run(s, func() {
		// If a panic occurs in a downstream handler, log it, generate a 500 error
		// response, and annotate the trace with additional panic-related info
		defer func() {
			if rcv := recover(); rcv != nil {
				stack := make([]byte, MAX_STACK_SIZE)
				stack = stack[:runtime.Stack(stack, false)]
				if b.logger != nil {
					b.logger.Printf("PANIC: %s\n%s", rcv, stack)
				}

				resp := NewError(nil, EcodeInternal, rcv)
				if b.respStacks {
					if len(stack) > b.respStackSize {
						resp.Stack = string(stack[:b.respStackSize])
					} else {
						resp.Stack = string(stack)
					}
				}
				writeResponse(rw, http.StatusInternalServerError, resp)
				if b.logger != nil {
					b.logger.Printf("%s \"%s %s %s\" %v %d \"%s\" \"%s\"",
						req.RemoteAddr, req.Method, req.RequestURI, req.Proto,
						res.Status(), res.Size(), req.Referer(), req.UserAgent())
				}

				if s != nil {
					data := s.Data()
					data["panic"] = rcv
					data["stack"] = stack
					data["resp_status"] = res.Status()
					data["resp_size"] = res.Size()
				}
			}
		}()

		// Invoke the next handler
		next(ctx, rw, req)
		if b.logger != nil {
			b.logger.Printf("%s \"%s %s %s\" %v %d \"%s\" \"%s\"",
				req.RemoteAddr, req.Method, req.RequestURI, req.Proto,
				res.Status(), res.Size(), req.Referer(), req.UserAgent())
		}

		// Annotate the trace
		if s != nil {
			data := s.Data()
			data["req_method"] = req.Method
			data["resp_status"] = res.Status()
			data["resp_size"] = res.Size()
		}
	})

	// Update request/response metrics
	if b.stats != nil {
		b.updateRequestResponseStats(res, req)
	}
}

func (b *Bottom) getRequestTraceIds(req *http.Request) (traceId, parentId int64) {
	if hdr := req.Header.Get(HeaderRequestId); hdr != "" {
		n, _ := fmt.Sscanf(hdr, "%d:%d", &traceId, &parentId)
		if n < 2 || traceId < 1 || parentId < 1 {
			traceId = 0
			parentId = 0
		}
	}
	return
}

func (b *Bottom) addRequestResponseTraceIds(rw http.ResponseWriter, req *http.Request, traceId, parentId int64) {
	req.Header.Set(HeaderRequestId, fmt.Sprintf("%d:%d", traceId, parentId))
	rw.Header().Set(HeaderRequestId, fmt.Sprintf("%d", traceId))
}

func (b *Bottom) updateRequestResponseStats(res ResponseWriter, req *http.Request) {
	stat := fmt.Sprintf("request.http.%s%s", strings.ToLower(req.Method), strings.Replace(req.URL.Path, "/", ".", -1))
	fmt.Println(stat)
	b.stats.Incr(stat, 1)
	stat = fmt.Sprintf("response.http.%s%s.%d", strings.ToLower(req.Method), strings.Replace(req.URL.Path, "/", ".", -1), res.Status())
	fmt.Println(stat)
	b.stats.Incr(stat, 1)
}
