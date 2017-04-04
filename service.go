package luddite

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	log "github.com/SpirentOrion/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
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

func NewService(config *ServiceConfig) (Service, error) {
	var err error

	// Create the service
	s := &service{
		config: config,
		router: httprouter.New(),
	}

	s.defaultLogger = log.New()
	s.defaultLogger.SetFormatter(&log.JSONFormatter{})
	if config.Log.ServiceLogPath != "" {
		openLogFile(s.defaultLogger, config.Log.ServiceLogPath)
	} else {
		s.defaultLogger.SetOutput(os.Stdout)
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

	s.accessLogger = log.New()
	s.accessLogger.SetFormatter(&log.JSONFormatter{})
	if config.Log.AccessLogPath != "" {
		openLogFile(s.accessLogger, config.Log.AccessLogPath)
	} else {
		s.accessLogger.SetOutput(os.Stdout)
		s.accessLogger.SetLevel(log.DebugLevel)
	}

	s.router.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	})

	s.router.MethodNotAllowed = http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Create default middleware handlers
	bottom, err := s.newBottomHandler()
	if err != nil {
		return nil, err
	}

	negotiator, err := s.newNegotiatorHandler()
	if err != nil {
		return nil, err
	}

	version, err := s.newVersionHandler()
	if err != nil {
		return nil, err
	}

	// Build middleware stack
	s.handlers = []Handler{bottom, negotiator, version}
	s.middleware = buildMiddleware(s.handlers)

	// Install default http handlers
	if s.config.Metrics.Enabled {
		s.addMetricsRoute()
	}
	if config.Schema.Enabled {
		s.addSchemaRoutes()
	}

	// Dump goroutine stacks on demand
	dumpGoroutineStacks(s.defaultLogger)
	return s, nil
}

func (s *service) AddHandler(h Handler) {
	s.handlers = append(s.handlers, h)
	s.middleware = buildMiddleware(s.handlers)
}

func (s *service) AddSingletonResource(basePath string, r Resource) {
	// GET /basePath
	AddGetRoute(s.router, basePath, false, r)

	// PUT /basePath
	AddUpdateRoute(s.router, basePath, false, r)

	// POST /basePath/{action}
	AddActionRoute(s.router, basePath, false, r)
}

func (s *service) AddCollectionResource(basePath string, r Resource) {
	// GET /basePath
	AddListRoute(s.router, basePath, r)

	// GET /basePath/{id}
	AddGetRoute(s.router, basePath, true, r)

	// POST /basePath
	AddCreateRoute(s.router, basePath, r)

	// PUT /basePath/{id}
	AddUpdateRoute(s.router, basePath, true, r)

	// DELETE /basePath
	AddDeleteRoute(s.router, basePath, false, r)

	// DELETE /basePath/{id}
	AddDeleteRoute(s.router, basePath, true, r)

	// POST /basePath/{id}/{action}
	AddActionRoute(s.router, basePath, true, r)
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

	if err = http.Serve(stoppableListener, middleware); err != nil {
		if _, ok := err.(*ListenerStoppedError); !ok {
			return err
		}
	}
	return nil
}

func (s *service) newBottomHandler() (Handler, error) {
	return NewBottom(s, s.defaultLogger, s.accessLogger), nil
}

func (s *service) newNegotiatorHandler() (Handler, error) {
	return NewNegotiator([]string{
		ContentTypeJson,
		ContentTypeCss,
		ContentTypePlain,
		ContentTypeXml,
		ContentTypeHtml,
		ContentTypeOctetStream},
	), nil
}

func (s *service) newVersionHandler() (Handler, error) {
	if s.config.Version.Min < 1 {
		return nil, errors.New("service's minimum API version must be greater than zero")
	}
	if s.config.Version.Max < 1 {
		return nil, errors.New("service's maximum API version must be greater than zero")
	}

	return NewVersion(s.config.Version.Min, s.config.Version.Max), nil
}

func (s *service) newRouterHandler() (Handler, error) {
	// No more middleware handlers: remaining dispatch happens via httprouter
	return WrapHttpHandler(s.router), nil
}

func (s *service) addMetricsRoute() {
	uriPath := s.config.Metrics.UriPath
	if uriPath == "" {
		uriPath = DEFAULT_METRICS_URI_PATH
	}

	h := prometheus.UninstrumentedHandler()
	s.router.GET(uriPath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) { h.ServeHTTP(rw, req) })
}

func (s *service) addSchemaRoutes() {
	config := s.config

	// Serve the various schemas, e.g. /schema/v1, /schema/v2, etc.
	s.schema = NewSchemaHandler(config.Schema.FilePath)

	s.router.GET(path.Join(config.Schema.UriPath, "/v:version/", "*filepath"), s.schema.ServeHTTP)

	// Temporarily redirect (307) the base schema path to the default schema file, e.g. /schema -> /schema/v2/fileName
	defaultSchemaPath := path.Join(config.Schema.UriPath, fmt.Sprintf("v%d", config.Version.Max), config.Schema.FileName)

	s.router.GET(config.Schema.UriPath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		http.Redirect(rw, req, defaultSchemaPath, http.StatusTemporaryRedirect)
	})

	// Temporarily redirect (307) the version schema path to the default schema file, e.g. /schema/v2 -> /schema/v2/fileName
	s.router.GET(path.Join(config.Schema.UriPath, "/v:version/"), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		http.Redirect(rw, req, defaultSchemaPath, http.StatusTemporaryRedirect)
	})

	// Optionally temporarily redirect (307) the root to the base schema path, e.g. / -> /schema
	if config.Schema.RootRedirect {
		s.router.GET("/", func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
			http.Redirect(rw, req, config.Schema.UriPath, http.StatusTemporaryRedirect)
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
