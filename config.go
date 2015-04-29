package luddite

import (
	"io/ioutil"
	"time"

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
		// StackSize sets an upper limit on the length of stack traces that appear in 500 error responses.
		StackSize int `yaml:"stack_size"`
	}
	Log struct {
		// Prefix sets the log prefix string.
		Prefix string
	}
	Metrics struct {
		// Enabled, when true, enables the service's statsd client.
		Enabled bool
		// Server is the address:port of the statsd server. Defaults to "127.0.0.1:8125".
		Server string
		// Prefix sets the statsd client prefix. Defaults to "%HOST%.".
		Prefix string
		// Interval sets the statsd client flush interface. Defaults to 2s.
		Interval time.Duration
	}
	Schema struct {
		// Enabled, when true, self-serve the service's own schema.
		Enabled bool
		// UriPath sets the URI path for the schema.
		UriPath string `yaml:"uri_path"`
		// FilePath sets the base file path for the schema.
		FilePath string `yaml:"file_path"`
		// FilePattern sets the schema file glob pattern.
		FilePattern string `yaml:"file_pattern"`
		// RootRedirect, when true, redirects the service's root to the default schema.
		RootRedirect bool `yaml:"root_redirect"`
	}
	Trace struct {
		// Enabled, when true, enables trace recording.
		Enabled bool
		// Buffer sets the trace package's buffer size.
		Buffer int
		// Recorder selects the trace recorder implementation: yaml | dynamodb.
		Recorder string
		// Params is a map of trace recorder parameters.
		Params map[string]string
	}
	Transport struct {
		// Tls, causes the service to listen using HTTPS.
		TLS bool `yaml:"tls"`
		// CertFilePath sets the path to the server's certificate file.
		CertFilePath string `yaml:"cert_file_path"`
		// KeyFilePath sets the path to the server's key file.
		KeyFilePath string `yaml:"key_file_path"`
	}
	Version struct {
		// Min sets the minimum API version that the service supports.
		Min int
		// Max sets the maximum API version that the service supports.
		Max int
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
