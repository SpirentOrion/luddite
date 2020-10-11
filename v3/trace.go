package luddite

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	basictracer "github.com/opentracing/basictracer-go"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	ddopentracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/yaml.v2"
)

var (
	tracerKinds = make(map[string]TracerKind)
	traceIdGen  = rand.New(rand.NewSource(time.Now().UnixNano()))
	traceIdLock sync.Mutex
)

type TracerKind interface {
	// New constructs a concrete opentracing.Tracer implementation, given a
	// map of tracer-specific parameters.
	New(params map[string]string, logger *log.Logger) (opentracing.Tracer, error)

	// TraceId returns a span's trace id.
	TraceId(span opentracing.Span) uint64

	// IsNoop returns true when the tracer is a no-op tracer. This may be
	// useful for short-circuiting span tagging and logging.
	IsNoop() bool
}

func init() {
	RegisterTracerKind("noop", new(noopTracerKind))
	RegisterTracerKind("json", new(jsonTracerKind))
	RegisterTracerKind("yaml", new(yamlTracerKind))
	RegisterTracerKind("datadog", new(datadogTracerKind))
}

func RegisterTracerKind(name string, kind TracerKind) {
	if name == "" {
		panic("empty tracer name")
	}
	if kind == nil {
		panic("mising tracer kind")
	}
	if _, ok := tracerKinds[name]; ok {
		panic(fmt.Sprintf("%s tracer kind registered twice", name))
	}
	tracerKinds[name] = kind
}

type noopTracerKind struct{}

func (*noopTracerKind) New(_ map[string]string, _ *log.Logger) (opentracing.Tracer, error) {
	return new(opentracing.NoopTracer), nil
}

func (*noopTracerKind) TraceId(_ opentracing.Span) uint64 {
	traceIdLock.Lock()
	defer traceIdLock.Unlock()
	return traceIdGen.Uint64()
}

func (*noopTracerKind) IsNoop() bool {
	return true
}

type jsonTracerKind struct{}

func (*jsonTracerKind) New(params map[string]string, _ *log.Logger) (opentracing.Tracer, error) {
	path := params["path"]
	if path == "" {
		return nil, errors.New("the json tracer requires a 'path' parameter")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	opts := basictracer.Options{
		ShouldSample:   func(_ uint64) bool { return true },
		Recorder:       &jsonSpanRecorder{json.NewEncoder(f)},
		MaxLogsPerSpan: 100,
		EnableSpanPool: true,
	}
	return basictracer.NewWithOptions(opts), nil
}

func (*jsonTracerKind) TraceId(span opentracing.Span) uint64 {
	ctx := span.Context().(basictracer.SpanContext)
	return uint64(ctx.TraceID)
}

func (*jsonTracerKind) IsNoop() bool {
	return false
}

type jsonSpanRecorder struct {
	*json.Encoder
}

func (r *jsonSpanRecorder) RecordSpan(span basictracer.RawSpan) {
	r.Encode(spanToBlob(&span))
}

type yamlTracerKind struct{}

func (*yamlTracerKind) New(params map[string]string, _ *log.Logger) (opentracing.Tracer, error) {
	path := params["path"]
	if path == "" {
		return nil, errors.New("the yaml tracer requires a 'path' parameter")
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	opts := basictracer.Options{
		ShouldSample:   func(_ uint64) bool { return true },
		Recorder:       &yamlSpanRecorder{f},
		MaxLogsPerSpan: 100,
		EnableSpanPool: true,
	}
	return basictracer.NewWithOptions(opts), nil
}

func (*yamlTracerKind) TraceId(span opentracing.Span) uint64 {
	ctx := span.Context().(basictracer.SpanContext)
	return uint64(ctx.TraceID)
}

func (*yamlTracerKind) IsNoop() bool {
	return false
}

type yamlSpanRecorder struct {
	io.Writer
}

func (r *yamlSpanRecorder) RecordSpan(span basictracer.RawSpan) {
	_, err := fmt.Fprintln(r, "---") // document separator
	if err != nil {
		return
	}

	buf, err := yaml.Marshal(spanToBlob(&span))
	if err != nil {
		return
	}

	_, _ = r.Write(buf)
}

type datadogTracerKind struct{}

func (*datadogTracerKind) New(params map[string]string, logger *log.Logger) (opentracing.Tracer, error) {
	opts := []ddtracer.StartOption{ddtracer.WithLogger(&datadogLogger{logger})}

	if agentAddr := params["agent_addr"]; agentAddr != "" {
		opts = append(opts, ddtracer.WithAgentAddr(agentAddr))
	}

	return ddopentracer.New(opts...), nil
}

func (*datadogTracerKind) TraceId(span opentracing.Span) uint64 {
	ctx := span.Context().(ddtrace.SpanContext)
	return uint64(ctx.TraceID())
}

func (*datadogTracerKind) IsNoop() bool {
	return false
}

type datadogLogger struct {
	*log.Logger
}

func (l *datadogLogger) Log(msg string) {
	l.Debug(msg)
}

func spanToBlob(span *basictracer.RawSpan) map[string]interface{} {
	blob := map[string]interface{}{
		"trace_id":       strconv.FormatUint(span.Context.TraceID, 10),
		"span_id":        strconv.FormatUint(span.Context.SpanID, 10),
		"parent_span_id": strconv.FormatUint(span.ParentSpanID, 10),
		"operation":      span.Operation,
		"start":          span.Start,
		"end":            span.Start.Add(span.Duration),
		"duration":       fmt.Sprintf("%.6f", span.Duration.Seconds()),
		"tags":           span.Tags,
	}

	if n := len(span.Logs); n > 0 {
		logs := make([]map[string]interface{}, n)
		for i := range span.Logs {
			l := &span.Logs[i]
			m := make(map[string]interface{}, len(l.Fields)+1)
			m["timestamp"] = l.Timestamp
			for j := range l.Fields {
				f := &l.Fields[j]
				m[f.Key()] = f.Value()
			}
			logs[i] = m
		}
		blob["logs"] = logs
	}

	return blob
}
