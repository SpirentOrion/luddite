package luddite

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"

	"github.com/dimfeld/httptreemux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Service implements a standalone RESTful web service.
type Service struct {
	config        *ServiceConfig
	globalRouter  *httptreemux.ContextMux
	apiRouters    map[int]*httptreemux.ContextMux
	defaultLogger *log.Logger
	accessLogger  *log.Logger
	tracerKind    TracerKind
	tracer        opentracing.Tracer
	schemas       http.FileSystem
	cors          *cors.Cors
	handlers      []Handler
	once          sync.Once
}

// NewService creates a new Service instance based on the given config.
// Middleware handlers and resources should be added before the service is run.
// The service may be run one time.
func NewService(config *ServiceConfig, configExt ...*ServiceConfigExt) (*Service, error) {
	var serviceLogWriter, accessLogWriter io.Writer
	switch len(configExt) {
	case 0:
	case 1:
		serviceLogWriter = configExt[0].ServiceLogWriter
		accessLogWriter = configExt[0].AccessLogWriter
	default:
		return nil, fmt.Errorf("Only single instance of ServiceConfigExt allowed atmost")
	}

	// Normalize and validate config
	config.Normalize()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create the service
	s := &Service{
		config:        config,
		defaultLogger: &log.Logger{Formatter: new(log.JSONFormatter)},
		apiRouters:    make(map[int]*httptreemux.ContextMux, config.Version.Max-config.Version.Min+1),
	}
	s.globalRouter = s.newRouter()
	for v := config.Version.Min; v <= config.Version.Max; v++ {
		s.apiRouters[v] = s.newRouter()
	}

	// Configure logging
	if serviceLogWriter != nil {
		s.defaultLogger.Out = serviceLogWriter
	} else if config.Log.ServiceLogPath != "" {
		// Service log to file
		openLogFile(s.defaultLogger, config.Log.ServiceLogPath)
	} else {
		// Service log to stdout
		s.defaultLogger.Out = os.Stdout
	}

	switch strings.ToLower(config.Log.ServiceLogLevel) {
	case "debug":
		s.defaultLogger.SetLevel(log.DebugLevel)
	default:
		fallthrough
	case "info":
		s.defaultLogger.SetLevel(log.InfoLevel)
	case "warn":
		s.defaultLogger.SetLevel(log.WarnLevel)
	case "error":
		s.defaultLogger.SetLevel(log.ErrorLevel)
	}

	s.accessLogger = &log.Logger{
		Formatter: new(log.JSONFormatter),
		Level:     log.InfoLevel,
	}
	if accessLogWriter != nil {
		s.accessLogger.Out = accessLogWriter
	} else if config.Log.AccessLogPath != "" {
		// Access log to file
		openLogFile(s.accessLogger, config.Log.AccessLogPath)
	} else if config.Log.ServiceLogPath != "" {
		// Access log to stdout
		s.accessLogger.Out = os.Stdout
	} else {
		// Both service log and access log to stdout (sharing a logger)
		s.accessLogger = s.defaultLogger
	}

	// Configure tracing
	s.tracerKind, s.tracer = s.newTracer()

	// Dump goroutine stacks on demand
	dumpGoroutineStacks()

	// Configure CORS
	if config.CORS.Enabled {
		opts := cors.Options{
			AllowedOrigins:   config.CORS.AllowedOrigins,
			AllowedMethods:   config.CORS.AllowedMethods,
			AllowedHeaders:   config.CORS.AllowedHeaders,
			ExposedHeaders:   config.CORS.ExposedHeaders,
			AllowCredentials: config.CORS.AllowCredentials,
		}
		s.cors = cors.New(opts)
	}

	// Create the default schema filesystem
	if config.Schema.Enabled {
		s.schemas = http.Dir(config.Schema.FilePath)
	}

	// Add default middleware handlers, beginning with "bottom"
	s.AddHandler(&bottomHandler{
		s:             s,
		defaultLogger: s.defaultLogger,
		accessLogger:  s.accessLogger,
		tracerKind:    s.tracerKind,
		tracer:        s.tracer,
		cors:          s.cors,
	})

	s.AddHandler(new(negotiatorHandler))

	s.AddHandler(&versionHandler{
		minVersion: s.config.Version.Min,
		maxVersion: s.config.Version.Max,
	})

	return s, nil
}

// Config returns the service's ServiceConfig instance.
func (s *Service) Config() *ServiceConfig {
	return s.config
}

// Logger returns the service's log.Logger instance.
func (s *Service) Logger() *log.Logger {
	return s.defaultLogger
}

// Logger returns the service's access logger instance.
func (s *Service) AccessLogger() *log.Logger {
	return s.accessLogger
}

// Tracer returns the service's opentracing.Tracer instance.
func (s *Service) Tracer() opentracing.Tracer {
	return s.tracer
}

// Router returns the service's router instance for the given API version.
func (s *Service) Router(version int) (*httptreemux.ContextMux, error) {
	if version < s.config.Version.Min || version > s.config.Version.Max {
		return nil, fmt.Errorf("API version is out of range (min: %d, max: %d)", s.config.Version.Min, s.config.Version.Max)
	}
	router := s.apiRouters[version]
	return router, nil
}

// AppendHandler appends a middleware handler to the service's middleware stack.
// All handlers must be added before Run is called.
func (s *Service) AppendHandler(h Handler) {
	s.handlers = append(s.handlers, h)
}

// PrependHandler prepends a middleware handler to the service's middleware
// stack. All handlers must be added before Run is called.
func (s *Service) PrependHandler(h Handler) {
	s.handlers = append([]Handler{h}, s.handlers...)
}

// AddHandler forwards to AppendHandler. It is provided for backward
// compatibility.
func (s *Service) AddHandler(h Handler) {
	s.AppendHandler(h)
}

// AddResource is a convenience method that performs runtime type assertions on
// a resource handler and adds routes as appropriate based on what interfaces
// are implemented. The same effect can be achieved by calling the various
// "Add*CollectionResource" and "Add*SingletonResource" functions with the
// appropriate router instance.
func (s *Service) AddResource(version int, basePath string, r interface{}) error {
	router, err := s.Router(version)
	if err != nil {
		return err
	}

	s.addCollectionRoutes(router, basePath, r)
	s.addSingletonRoutes(router, basePath, r)
	return nil
}

// SetSchemas allows a service to provide its own HTTP filesystem to be used for
// schema assets. This overrides the use of the local filesystem and paths given
// in the service config.
func (s *Service) SetSchemas(schemas http.FileSystem) {
	s.schemas = schemas
}

// Run starts the service's HTTP server and runs it forever or until SIGINT is
// received. This method should be invoked once per service.
func (s *Service) Run() error {
	err := errors.New("service instances may only be run one time")
	s.once.Do(func() { err = s.run() })
	return err
}

func (s *Service) newRouter() *httptreemux.ContextMux {
	router := httptreemux.NewContextMux()
	router.NotFoundHandler = func(rw http.ResponseWriter, _ *http.Request) { rw.WriteHeader(http.StatusNotFound) }
	if prefix := s.config.Prefix; prefix != "" {
		router.ContextGroup = router.NewGroup(prefix)
	}
	return router
}

func (s *Service) newTracer() (kind TracerKind, tracer opentracing.Tracer) {
	var (
		name string
		err  error
	)
	defer func() {
		if err != nil {
			s.defaultLogger.WithFields(log.Fields{
				"tracer": name,
				"error":  err.Error(),
			}).Warn("tracing is not active")
		}
		if tracer == nil {
			kind = tracerKinds["noop"]
			tracer, _ = kind.New(nil, s.defaultLogger)
			s.defaultLogger.Warn("using no-op tracer")
		}
	}()

	config := s.config
	if !config.Trace.Enabled {
		return
	}

	name = config.Trace.Tracer
	kind, ok := tracerKinds[name]
	if !ok {
		err = errors.New("unknown tracer")
		return
	}

	tracer, err = kind.New(config.Trace.Params, s.defaultLogger)
	return
}

func (s *Service) addMetricsRoute() {
	h := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	s.globalRouter.GET(s.config.Metrics.URIPath, h.ServeHTTP)
}

func (s *Service) addProfilerRoutes() {
	router := s.globalRouter
	uriPath := path.Clean(s.config.Profiler.URIPath)
	router.GET(strings.TrimRight(uriPath, "/")+"/", pprof.Index)
	router.GET(path.Join(uriPath, "/allocs"), pprof.Handler("allocs").ServeHTTP)
	router.GET(path.Join(uriPath, "/block"), pprof.Handler("block").ServeHTTP)
	router.GET(path.Join(uriPath, "/cmdline"), pprof.Cmdline)
	router.GET(path.Join(uriPath, "/goroutine"), pprof.Handler("goroutine").ServeHTTP)
	router.GET(path.Join(uriPath, "/heap"), pprof.Handler("heap").ServeHTTP)
	router.GET(path.Join(uriPath, "/mutex"), pprof.Handler("mutex").ServeHTTP)
	router.GET(path.Join(uriPath, "/profile"), pprof.Profile)
	router.POST(path.Join(uriPath, "/profile"), pprof.Profile)
	router.GET(path.Join(uriPath, "/symbol"), pprof.Symbol)
	router.POST(path.Join(uriPath, "/symbol"), pprof.Symbol)
	router.GET(path.Join(uriPath, "/threadcreate"), pprof.Handler("threadcreate").ServeHTTP)
	router.GET(path.Join(uriPath, "/trace"), pprof.Trace)
	router.POST(path.Join(uriPath, "/trace"), pprof.Trace)
}

func (s *Service) addSchemaRoutes() {
	config := s.config
	router := s.globalRouter

	// Serve the various schemas, e.g. /schema/v1, /schema/v2, etc.
	h := newSchemaHandler(s.schemas)
	router.GET(path.Join(config.Schema.URIPath, ":version/*filepath"), h.ServeHTTP)

	// Temporarily redirect (307) the base schema path to the default schema file, e.g. /schema -> /schema/v2/fileName
	defaultSchemaPath := path.Join(config.Prefix, config.Schema.URIPath, fmt.Sprintf("v%d", config.Version.Max), config.Schema.FileName)
	router.GET(config.Schema.URIPath, func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, defaultSchemaPath, http.StatusTemporaryRedirect)
	})

	// Temporarily redirect (307) the version schema path to the default schema file, e.g. /schema/v2 -> /schema/v2/fileName
	router.GET(path.Join(config.Schema.URIPath, ":version"), func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, defaultSchemaPath, http.StatusTemporaryRedirect)
	})

	// Optionally temporarily redirect (307) the root to the base schema path, e.g. / -> /schema
	if config.Schema.RootRedirect {
		router.GET("/", func(rw http.ResponseWriter, req *http.Request) {
			http.Redirect(rw, req, defaultSchemaPath, http.StatusTemporaryRedirect)
		})
	}
}

func (s *Service) addCollectionRoutes(router *httptreemux.ContextMux, basePath string, r interface{}) {
	if x, ok := r.(CollectionLister); ok {
		AddListCollectionRoute(router, basePath, x)
	}
	if x, ok := r.(CollectionCounter); ok {
		AddCountCollectionRoute(router, basePath, x)
	}
	if x, ok := r.(CollectionGetter); ok {
		AddGetCollectionRoute(router, basePath, x)
	}
	if x, ok := r.(CollectionCreator); ok {
		AddCreateCollectionRoute(router, basePath, x)
	}
	if x, ok := r.(CollectionUpdater); ok {
		AddUpdateCollectionRoute(router, basePath, x)
	}
	if x, ok := r.(CollectionDeleter); ok {
		AddDeleteCollectionRoute(router, basePath, x)
	}
	if x, ok := r.(CollectionActioner); ok {
		AddActionCollectionRoute(router, basePath, x)
	}
}

func (s *Service) addSingletonRoutes(router *httptreemux.ContextMux, basePath string, r interface{}) {
	if x, ok := r.(SingletonGetter); ok {
		AddGetSingletonRoute(router, basePath, x)
	}
	if x, ok := r.(SingletonUpdater); ok {
		AddUpdateSingletonRoute(router, basePath, x)
	}
	if x, ok := r.(SingletonActioner); ok {
		AddActionSingletonRoute(router, basePath, x)
	}
}

func (s *Service) run() error {
	config := s.config

	// Add optional HTTP handlers
	if config.Metrics.Enabled {
		s.addMetricsRoute()
	}
	if config.Profiler.Enabled {
		s.addProfilerRoutes()
	}
	if config.Schema.Enabled {
		s.addSchemaRoutes()
	}

	// Add "top" as the final middleware handler
	s.AddHandler(&topHandler{
		globalRouter: s.globalRouter,
		apiRouters:   s.apiRouters,
	})
	middleware := buildMiddleware(s.handlers)

	// If metrics are enabled let Prometheus have a look at the request first
	var h http.HandlerFunc
	if config.Metrics.Enabled {
		h = instrumentHTTPHandler(middleware).ServeHTTP
	} else {
		h = middleware.ServeHTTP
	}

	// Serve HTTP or HTTPS, depending on config. Use stoppable listener so
	// we can exit gracefully if signaled to do so.
	if config.Transport.TLS {
		s.defaultLogger.Debugf("HTTPS listening on %s", config.Addr)
		l, err := NewStoppableTLSListener(config.Addr, true, config.Transport.CertFilePath, config.Transport.KeyFilePath)
		if err != nil {
			return err
		}

		if err = http.Serve(l, h); err != nil {
			if _, ok := err.(*ListenerStoppedError); ok {
				err = nil
			}
		}
	} else {
		s.defaultLogger.Debugf("HTTP listening on %s", config.Addr)
		l, err := NewStoppableTCPListener(config.Addr, true)
		if err != nil {
			return err
		}

		h2s := new(http2.Server)
		if err = http.Serve(l, h2c.NewHandler(h, h2s)); err != nil {
			if _, ok := err.(*ListenerStoppedError); ok {
				err = nil
			}
		}
	}

	return nil
}

func openLogFile(logger *log.Logger, logPath string) {
	sigs := make(chan os.Signal, 1)
	logging := make(chan struct{})

	go func() {
		var curLog *os.File
		for {
			// Open and begin using a new log file
			newLog, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				panic(err)
			}

			logger.Out = newLog
			if curLog == nil {
				// First log, signal the outer goroutine that we're running
				close(logging)
			} else {
				// Follow-on log, close the current log file
				_ = curLog.Close()
			}
			curLog = newLog

			// Wait for a SIGHUP
			<-sigs
		}
	}()

	signal.Notify(sigs, syscall.SIGHUP)
	<-logging
}
