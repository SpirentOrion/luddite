package luddite

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ServiceConfig is a struct that holds config values relevant to the service framework.
type ServiceConfig struct {
	// Addr is the address:port pair that the HTTP server listens on.
	Addr  string
	Debug struct {
		// Requests, when true, causes logging of requests and responses.
		Requests bool
		// Stacks, when true, causes stack traces to appear in 500 error responses.
		Stacks bool
	}
	Log struct {
		// Prefix sets the log prefix string.
		Prefix string
	}
	Trace struct {
		// Enabled, when true, enables use of the trace package
		Enabled bool
		// Buffer sets the trace package's buffer size
		Buffer int
		// Recorder selects the trace recorder implementation: yaml | dynamodb
		Recorder string
		// Params is a colon-separated list of trace recorder constructor parameters
		Params string
	}
}

// ReadConfig reads a YAML config file from path. The file is parsed into the struct pointed to by cfg.
func ReadConfig(path string, cfg interface{}) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(buf, cfg)
}
