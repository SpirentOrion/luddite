package luddite

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ServiceConfig is a struct that holds config values relevant to the service framework.
type ServiceConfig struct {
	// Addr is the address:port pair that the HTTP server listens on.
	Addr string
	Cors struct {
		// Enabled, when true, enables CORS.
		Enabled bool
		// AllowedOrigins contains the list of origins a cross-domain request can be executed from. Defaults to "*" on an empty list.
		AllowedOrigins []string `yaml:"allowed_origins"`
		// AllowedMethods contains the list of methods the client is allowed to use with cross-domain requests. Defaults to "GET", "POST", "PUT" and "DELETE" on an empty list.
		AllowedMethods []string `yaml:"allowed_methods"`
		// AllowedHeaders contains the list of non-simple headers clients are allowed to use in cross-origin requests.  An empty list is interpreted literally however "Origin" is always appended.
		AllowedHeaders []string `yaml:"allowed_headers"`
		// ExposedHeaders contains the list of non-simple headers clients are allowed to access in cross-origin responses.  An empty list is interpreted literally.
		ExposedHeaders []string `yaml:"exposed_headers"`
		// AllowCredentials indicates whether the request can include user credentials like cookies or HTTP auth.
		AllowCredentials bool `yaml:"allow_credentials"`
	}
	// Credentials is a generic map of strings that may be used to store tokens, AWS keys, etc.
	Credentials map[string]string
	Debug       struct {
		// Stacks, when true, causes stack traces to appear in 500 error responses.
		Stacks bool
		// StackSize sets an upper limit on the length of stack traces that appear in 500 error responses.
		StackSize int `yaml:"stack_size"`
	}
	Log struct {
		// ServiceLogPath sets the file path for the service log (written as JSON). If unset, defaults to stdout (written as text).
		ServiceLogPath string `yaml:"service_log_path"`
		// ServiceLogLevel sets the minimum log level for the service log, If unset, defaults to INFO.
		ServiceLogLevel string `yaml:"service_log_level"`
		// AccessLogPath sets the file path for the access log (written as JSON). If unset, defaults to stdout (written as text).
		AccessLogPath string `yaml:"access_log_path"`
	}
	Metrics struct {
		// Enabled, when true, enables the service's prometheus client.
		Enabled bool
		// UriPath sets the metrics path. Defaults to "/metrics".
		UriPath string `yaml:"uri_path"`
	}
	Profiler struct {
		// Enabled, when true, enables the service's profiling endpoints.
		Enabled bool
		// UriPath sets the profiler path. Defaults to "/debug/pprof".
		UriPath string `yaml:"uri_path"`
	}
	Schema struct {
		// Enabled, when true, self-serve the service's own schema.
		Enabled bool
		// UriPath sets the URI path for the schema.
		UriPath string `yaml:"uri_path"`
		// FilePath sets the base file path for the schema.
		FilePath string `yaml:"file_path"`
		// FileName sets the schema file name.
		FileName string `yaml:"file_name"`
		// RootRedirect, when true, redirects the service's root to the default schema.
		RootRedirect bool `yaml:"root_redirect"`
	}
	Trace struct {
		// Enabled, when true, enables trace recording.
		Enabled bool
		// Buffer sets the trace package's buffer size.
		Buffer int
		// Recorder selects the trace recorder implementation: json | other.
		Recorder string
		// Params is a map of trace recorder parameters.
		Params map[string]string
	}
	Transport struct {
		// Tls, when true, causes the service to listen using HTTPS.
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
