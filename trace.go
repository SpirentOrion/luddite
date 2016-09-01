package luddite

import (
	"fmt"
	"io"

	"gopkg.in/SpirentOrion/trace.v2"
	"gopkg.in/yaml.v2"
)

const (
	TraceKindAWS     = "aws"
	TraceKindProcess = "process"
	TraceKindRequest = "request"
	TraceKindWorker  = "worker"
)

var recorders = make(map[string]trace.Recorder)

func RegisterTraceRecorder(name string, recorder trace.Recorder) {
	if name == "" {
		panic("empty trace recorder name")
	}
	if recorder == nil {
		panic("nil trace recorder")
	}
	if _, ok := recorders[name]; ok {
		panic(fmt.Sprintf("%s trace recorder registered twice", name))
	}
	recorders[name] = recorder
}

// yamlRecorder is included in luddite for backwards compatibility with v1 of
// github.com/SpirentOrion/trace since v2 of that package no longer supports
// YAML directly.
type yamlRecorder struct {
	io.Writer
}

func (r *yamlRecorder) Record(s *trace.Span) error {
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	fmt.Fprintln(r, "---") // document separator
	_, err = r.Write(buf)
	return err
}
