package luddite

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/SpirentOrion/httprouter"
	"github.com/SpirentOrion/trace"
	"github.com/SpirentOrion/trace/dynamorec"
	"github.com/SpirentOrion/trace/yamlrec"
	"golang.org/x/net/context"
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

	// Logger returns the service's log.Logger instance.
	Logger() *log.Logger

	// Run is a convenience function that runs the service as an
	// HTTP server. The address is taken from the ServiceConfig
	// passed to NewService.
	Run() error
}

type service struct {
	config     *ServiceConfig
	logger     *log.Logger
	router     *httprouter.Router
	handlers   []Handler
	middleware middleware
	schema     *SchemaHandler
}

func NewService(config *ServiceConfig) (Service, error) {
	// Create service
	s := &service{
		config:   config,
		logger:   log.New(os.Stderr, config.Log.Prefix, log.LstdFlags),
		router:   httprouter.New(),
		handlers: []Handler{},
	}

	s.router.NotFound = func(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}

	s.router.MethodNotAllowed = func(_ context.Context, rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}

	// Create default middleware handlers
	h, err := s.newRecoveryHandler()
	if err != nil {
		return nil, err
	}
	s.handlers = append(s.handlers, h)

	h, err = s.newTraceHandler()
	if err != nil {
		return nil, err
	}
	s.handlers = append(s.handlers, h)

	h, err = s.newLoggerHandler()
	if err != nil {
		return nil, err
	}
	s.handlers = append(s.handlers, h)

	h, err = s.newNegotiatorHandler()
	if err != nil {
		return nil, err
	}
	s.handlers = append(s.handlers, h)

	h, err = s.newContextHandler(config.Version.Max)
	if err != nil {
		return nil, err
	}
	s.handlers = append(s.handlers, h)

	// Build middleware stack
	s.middleware = buildMiddleware(s.handlers)

	// Install default http handlers
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

func (s *service) Logger() *log.Logger {
	return s.logger
}

func (s *service) Run() error {
	// Add the router as the final middleware handler
	h, err := s.newRouterHandler()
	if err != nil {
		return err
	}
	s.AddHandler(h)

	// Serve
	s.logger.Printf("listening on %s", s.config.Addr)
	return http.ListenAndServe(s.config.Addr, s.middleware)
}

func (s *service) newRecoveryHandler() (Handler, error) {
	h := NewRecovery()
	h.Logger = s.logger
	h.StacksVisible = s.config.Debug.Stacks
	return h, nil
}

func (s *service) newTraceHandler() (Handler, error) {
	if s.config.Trace.Enabled {
		var (
			rec trace.Recorder
			err error
		)

		params := strings.Split(s.config.Trace.Params, ":")
		switch s.config.Trace.Recorder {
		case "yaml":
			if len(params) != 1 {
				return nil, errors.New("yaml trace recorder expects 1 parameter (path)")
			}
			f, err := os.OpenFile(params[0], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err == nil {
				rec, err = yamlrec.New(f)
			} else {
				f.Close()
			}
			break
		case "dynamodb":
			if len(params) != 4 {
				return nil, errors.New("dynamodb trace recorder expects 4 parameters (region:table_name:access_key:secret_key)")
			}
			rec, err = dynamorec.New(params[0], params[1], params[2], params[3])
			break
		default:
			err = fmt.Errorf("unknown trace recorder: ", s.config.Trace.Recorder)
			break
		}

		if rec != nil {
			err = trace.Record(rec, s.config.Trace.Buffer, s.logger)
		}

		if err != nil {
			s.logger.Println("trace recording is not active:", err)
		} else {
			s.logger.Println("recording traces to", rec)
		}
	}

	return WrapMiddlewareHandler(trace.ServeHTTP), nil
}

func (s *service) newLoggerHandler() (Handler, error) {
	if !s.config.Debug.Requests {
		return nil, nil
	}

	return NewLogger(s.logger), nil
}

func (s *service) newNegotiatorHandler() (Handler, error) {
	return NewNegotiator([]string{ContentTypeJson, ContentTypeXml, ContentTypeHtml}), nil
}

func (s *service) newContextHandler(maxVersion int) (Handler, error) {
	if maxVersion < 1 {
		return nil, errors.New("service's maximum API version must be greater than zero")
	}

	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, next ContextHandlerFunc) {
		ctx = WithService(ctx, s)
		ctx = WithApiVersion(ctx, RequestApiVersion(r, maxVersion))
		next(ctx, rw, r)
	}), nil
}

func (s *service) newRouterHandler() (Handler, error) {
	return HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request, _ ContextHandlerFunc) {
		s.router.HandleHTTP(ctx, rw, r)
	}), nil
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

	// Optionally permanently redirect (301) the root to the base schema path, e.g. / -> /schema
	if config.Schema.RootRedirect {
		s.router.GET("/", func(_ context.Context, rw http.ResponseWriter, r *http.Request) {
			http.Redirect(rw, r, config.Schema.UriPath, http.StatusMovedPermanently)
		})
	}
}
