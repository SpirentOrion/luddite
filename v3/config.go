package luddite

import (
	"errors"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	defaultMetricsURIPath  = "/metrics"
	defaultProfilerURIPath = "/debug/pprof"
)

var (
	// ErrInvalidMinApiVersion occurs when a service's minimum API version is <= 0.
	ErrInvalidMinApiVersion = errors.New("service's minimum API version must be greater than zero")

	// ErrInvalidMaxApiVersion occurs when a service's maximum API version is <= 0.
	ErrInvalidMaxApiVersion = errors.New("service's maximum API version must be greater than zero")

	// ErrMismatchedApiVersions occurs when a service's minimum API version > its maximum API version.
	ErrMismatchedApiVersions = errors.New("service's maximum API version must be greater than or equal to the minimum API version")

	// ErrMissingTLSConfig occurs when TLS is enabled without required file paths
	ErrMissingTLSConfig = errors.New("must set both CertFilePath and KeyFilePath to enable TLS transport")

	defaultCORSAllowedMethods = []string{"GET", "POST", "PUT", "DELETE"}
)

// ServiceConfig holds a service's config values.
type ServiceConfig struct {
	// Addr is the address:port pair that the HTTP server listens on.
	Addr string

	// Prefix is a prefix to add to every path
	Prefix string

	CORS struct {
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

	Debug struct {
		// Stacks, when true, causes stack traces to appear in 500 error responses.
		Stacks bool
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
		URIPath string `yaml:"uri_path"`
	}

	Profiler struct {
		// Enabled, when true, enables the service's profiling endpoints.
		Enabled bool

		// UriPath sets the profiler path. Defaults to "/debug/pprof".
		URIPath string `yaml:"uri_path"`
	}

	Schema struct {
		// Enabled, when true, self-serve the service's own schema.
		Enabled bool

		// URIPath sets the URI path for the schema.
		URIPath string `yaml:"uri_path"`

		// FilePath sets the base file path for the schema.
		FilePath string `yaml:"file_path"`

		// FileName sets the schema file name.
		FileName string `yaml:"file_name"`

		// RootRedirect, when true, redirects the service's root to the default schema.
		RootRedirect bool `yaml:"root_redirect"`
	}

	Trace struct {
		// Enabled, when true, enables distributed tracing using the
		// OpenTracing framework.
		Enabled bool

		// Tracer selects the tracer implementation. Built-in
		// Tracer: json, yaml.
		Tracer string

		// Params is a map of tracer-specific parameters.
		Params map[string]string
	}

	Transport struct {
		// Tls, when true, causes the service to listen using HTTPS.
		TLS bool `yaml:"tls"`

		// CertFilePath sets the path to the server's certificate file.
		CertFilePath string `yaml:"cert_file_path"`

		// KeyFilePath sets the path to the server's key file.
		KeyFilePath string `yaml:"key_file_path"`

		// CertWatcher monitor CertFilePath and KeyFilePath for changes
		CertWatcher struct {
			// Disabled disable monitoring and automatic reloads when cert/key files are changed
			Disabled bool `yaml:"disabled,omitempty"`
			// ScanMinutes when monitoring for file changes, how often to scan for changes (default is 5)
			ScanMinutes int `yaml:"scan_minutes,omitempty"`
		} `yaml:"cert_watcher"`
	}

	Version struct {
		// Min sets the minimum API version that the service supports.
		Min int

		// Max sets the maximum API version that the service supports.
		Max int
	}
}

// Normalize applies sensible defaults to service config values when they are
// otherwise unspecified or invalid.
func (config *ServiceConfig) Normalize() {
	if config.CORS.Enabled && len(config.CORS.AllowedMethods) == 0 {
		config.CORS.AllowedMethods = defaultCORSAllowedMethods
	}

	if config.Metrics.Enabled && config.Metrics.URIPath == "" {
		config.Metrics.URIPath = defaultMetricsURIPath
	}

	if config.Profiler.Enabled && config.Profiler.URIPath == "" {
		config.Profiler.URIPath = defaultProfilerURIPath
	}
}

// Validate sanity-checks service config values.
func (config *ServiceConfig) Validate() error {
	if config.Version.Min < 1 {
		return ErrInvalidMinApiVersion
	}

	if config.Version.Max < 1 {
		return ErrInvalidMaxApiVersion
	}

	if config.Version.Min > config.Version.Max {
		return ErrMismatchedApiVersions
	}

	if config.Transport.TLS && (config.Transport.CertFilePath == "" || config.Transport.KeyFilePath == "") {
		return ErrMissingTLSConfig
	}

	return nil
}

// NewConfig parses a YAML config file from an io.Reader. The file is strictly
// parsed (any fields that are found in the data that do not have corresponding
// struct members, or mapping keys that are duplicates, will result in an error)
// into the struct pointed to by cfg.
func NewConfig(r io.Reader, cfg interface{}) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return yaml.UnmarshalStrict(buf, cfg)
}

// ReadConfig reads a YAML config file from path. The file is strictly parsed
// (any fields that are found in the data that do not have corresponding struct
// members, or mapping keys that are duplicates, will result in an error) into
// the struct pointed to by cfg.
func ReadConfig(path string, cfg interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	err = NewConfig(f, cfg)
	_ = f.Close()
	return err
}
