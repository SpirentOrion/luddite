package luddite

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"

	"github.com/SpirentOrion/httprouter"
	log "github.com/SpirentOrion/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
)

const (
	DEFAULT_METRICS_URI_PATH = "/metrics"
)

// Service is an interface that implements a standalone RESTful web service.
type Service interface {
	// AddHandler adds a context-aware middleware handler to the
	// middleware stack. All handlers must be added before Run is
	// called.
	AddHandler(h Handler)

	// AddSingletonResource registers a singleton-style resource
	// (supporting GET and PUT methods only).
	AddSingletonResource(itemPath string, r Resource)

	// AddCollectionResource registers a collection-style resource
	// (supporting GET, POST, PUT, and DELETE methods).
	AddCollectionResource(basePath string, r Resource)

	// Config returns the service's ServiceConfig instance.
	Config() *ServiceConfig

	// Logger returns the service's log.Logger instance.
	Logger() *log.Logger

	// Router returns the service's httprouter.Router instance.
	Router() *httprouter.Router

	// Run is a convenience function that runs the service as an
	// HTTP server. The address is taken from the ServiceConfig
	// passed to NewService.
	Run() error
}

type service struct {
	config        *ServiceConfig
	defaultLogger *log.Logger
	accessLogger  *log.Logger
	router        *httprouter.Router
	handlers      []Handler
	middleware    *middleware
	schema        *SchemaHandler
}

// Verify that service implements Service.
var _ Service = &service{}

func NewService(config *ServiceConfig) (Service, error) {
	var err error

	// Create the service
	s := &service{
		config: config,
		router: httprouter.New(),
	}

	s.defaultLogger = log.New()
	if config.Log.ServiceLogPath != "" {
		openLogFile(s.defaultLogger, config.Log.ServiceLogPath)
		s.defaultLogger.Formatter = &log.JSONFormatter{}
	} else {
		s.defaultLogger.Out = os.Stdout
	}

	switch strings.ToLower(config.Log.ServiceLogLevel) {
	case "debug":
		s.defaultLogger.Level = log.DebugLevel
	default:
		fallthrough
	case "info":
		s.defaultLogger.Level = log.InfoLevel
	case "warn":
		s.defaultLogger.Level = log.WarnLevel
	case "error":
		s.defaultLogger.Level = log.ErrorLevel
	}

	// Add handler to log stacktrace
	addStackTraceHandler(s.defaultLogger)

	s.accessLogger = log.New()
	if config.Log.AccessLogPath != "" {
		openLogFile(s.accessLogger, config.Log.AccessLogPath)
		s.accessLogger.Formatter = &log.JSONFormatter{}
	} else {
		s.accessLogger.Out = os.Stdout
		s.accessLogger.Level = log.DebugLevel
	}

	s.router.NotFound = func(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}

	s.router.MethodNotAllowed = func(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

	// Create default middleware handlers
	bottom, err := s.newBottomHandler()
	if err != nil {
		return nil, err
	}

	negotiator, err := s.newNegotiatorHandler()
	if err != nil {
		return nil, err
	}

	context, err := s.newContextHandler()
	if err != nil {
		return nil, err
	}

	// Build middleware stack
	s.handlers = []Handler{bottom, negotiator, context}
	s.middleware = buildMiddleware(s.handlers)

	// Install default http handlers
	if s.config.Metrics.Enabled {
		s.addMetricsRoute()
	}
	if config.Schema.Enabled {
		s.addSchemaRoutes()
	}

	return s, nil
}

func (s *service) AddHandler(h Handler) {
	s.handlers = append(s.handlers, h)
	s.middleware = buildMiddleware(s.handlers)
}

func (s *service) AddSingletonResource(basePath string, r Resource) {
	// GET /basePath
	addGetRoute(s.router, basePath, false, r)

	// PUT /basePath
	addUpdateRoute(s.router, basePath, false, r)

	// POST /basePath/{action}
	addActionRoute(s.router, basePath, false, r)
}

func (s *service) AddCollectionResource(basePath string, r Resource) {
	// GET /basePath
	addListRoute(s.router, basePath, r)

	// GET /basePath/{id}
	addGetRoute(s.router, basePath, true, r)

	// POST /basePath
	addCreateRoute(s.router, basePath, r)

	// PUT /basePath/{id}
	addUpdateRoute(s.router, basePath, true, r)

	// DELETE /basePath
	addDeleteRoute(s.router, basePath, false, r)

	// DELETE /basePath/{id}
	addDeleteRoute(s.router, basePath, true, r)

	// POST /basePath/{id}/{action}
	addActionRoute(s.router, basePath, true, r)
}

func (s *service) Config() *ServiceConfig {
	return s.config
}

func (s *service) Logger() *log.Logger {
	return s.defaultLogger
}

func (s *service) Router() *httprouter.Router {
	return s.router
}

func (s *service) Run() error {
	// Add the router as the final middleware handler
	h, err := s.newRouterHandler()
	if err != nil {
		return err
	}
	s.AddHandler(h)

	var middleware http.Handler = s.middleware
	if s.config.Metrics.Enabled {
		middleware = prometheus.InstrumentHandler("service", middleware)
	}

	// Serve HTTP or HTTPS, depending on config. Use stoppable listener
	// so we can exit gracefully if signaled to do so.
	var stoppableListener net.Listener
	if s.config.Transport.TLS {
		s.defaultLogger.Debugf("HTTPS listening on %s", s.config.Addr)
		stoppableListener, err = NewStoppableTLSListener(s.config.Addr, true, s.config.Transport.CertFilePath, s.config.Transport.KeyFilePath)
	} else {
		s.defaultLogger.Debugf("HTTP listening on %s", s.config.Addr)
		stoppableListener, err = NewStoppableTCPListener(s.config.Addr, true)
	}
	if err != nil {
		return err
	}
	err = http.Serve(stoppableListener, middleware)
	if _, ok := err.(*ListenerStoppedError); ok {
		return nil
	}
	return err

}

func (s *service) newBottomHandler() (Handler, error) {
	return NewBottom(s.config, s.defaultLogger, s.accessLogger), nil
}

func (s *service) newNegotiatorHandler() (Handler, error) {
	return NewNegotiator([]string{ContentTypeJson, ContentTypeXml, ContentTypeHtml, ContentTypeOctetStream}), nil
}

func (s *service) newContextHandler() (Handler, error) {
	if s.config.Version.Min < 1 {
		return nil, errors.New("service's minimum API version must be greater than zero")
	}
	if s.config.Version.Max < 1 {
		return nil, errors.New("service's maximum API version must be greater than zero")
	}

	return NewContext(s, s.config.Version.Min, s.config.Version.Max), nil
}

func (s *service) newRouterHandler() (Handler, error) {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, _ ContextHandlerFunc) {
		// No more middleware handlers: further dispatch happens via httprouter
		s.router.HandleHTTP(ctx, rw, r)
	}), nil
}

func (s *service) addMetricsRoute() {
	uriPath := s.config.Metrics.UriPath
	if uriPath == "" {
		uriPath = DEFAULT_METRICS_URI_PATH
	}

	h := prometheus.UninstrumentedHandler()
	s.router.GET(uriPath, func(_ context.Context, rw http.ResponseWriter, r *http.Request) { h.ServeHTTP(rw, r) })
}

func (s *service) addSchemaRoutes() {
	config := s.config

	// Serve the various schemas, e.g. /schema/v1, /schema/v2, etc.
	s.schema = NewSchemaHandler(config.Schema.FilePath, config.Schema.FilePattern)
	s.router.GET(path.Join(config.Schema.UriPath, "/v:version"), s.schema.ServeHTTP)

	// Temporarily redirect (307) the base schema path to the default schema, e.g. /schema -> /schema/v2
	defaultSchemaPath := path.Join(config.Schema.UriPath, fmt.Sprintf("v%d", config.Version.Max))
	s.router.GET(config.Schema.UriPath, func(_ context.Context, rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, defaultSchemaPath, http.StatusTemporaryRedirect)
	})

	// Optionally temporarily redirect (307) the root to the base schema path, e.g. / -> /schema
	if config.Schema.RootRedirect {
		s.router.GET("/", func(_ context.Context, rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r, config.Schema.UriPath, http.StatusTemporaryRedirect)
		})
	}
}

func openLogFile(logger *log.Logger, logPath string) {
	sigs := make(chan os.Signal, 1)
	logging := make(chan bool, 1)

	go func() {
		var curLog, priorLog *os.File
		for {
			// Open and begin using a new log file
			curLog, _ = os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			logger.SetOutput(curLog)

			if priorLog == nil {
				// First log, signal the outer goroutine that we're running
				logging <- true
			} else {
				// Follow-on log, close the prior log file
				priorLog.Close()
				priorLog = nil
			}

			// Wait for a SIGHUP
			<-sigs

			// Setup for the next iteration
			priorLog = curLog
		}
	}()

	signal.Notify(sigs, syscall.SIGHUP)
	<-logging
}

func addStackTraceHandler(logger *log.Logger) {
	sigs := make(chan os.Signal, 1)
	go func() {
		for {
			<-sigs
			buf := make([]byte, 1<<16)
			size := runtime.Stack(buf, true)
			logger.Infof("*** goroutine dump ***\n%s", buf[:size])
		}
	}()
	signal.Notify(sigs, syscall.SIGUSR1)
}
