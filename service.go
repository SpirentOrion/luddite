package luddite

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/K-Phoen/http-negotiate/negotiate"
	"github.com/SpirentOrion/trace"
	"github.com/SpirentOrion/trace/dynamorec"
	"github.com/SpirentOrion/trace/yamlrec"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

// Service is an interface that implements a standalone RESTful web service.
type Service interface {
	// AddSingletonResource registers a singleton-style resource
	// (supporting GET and PUT methods only).
	AddSingletonResource(itemPath string, r Resource)

	// AddCollectionResource registers a collection-style resource
	// (supporting GET, POST, PUT, and DELETE methods).
	AddCollectionResource(basePath string, r Resource)

	// Run is a convenience function that runs the service as an
	// HTTP server. The address is taken from the ServiceConfig
	// passed to NewService.
	Run() error
}

type service struct {
	config        *ServiceConfig
	errorMessages map[int]string
	logger        *log.Logger
	router        *mux.Router
	negroni       *negroni.Negroni
}

func NewService(config *ServiceConfig) Service {
	return &service{
		config:  config,
		logger:  log.New(os.Stderr, config.Log.Prefix, log.LstdFlags),
		router:  mux.NewRouter(),
		negroni: negroni.New(),
	}
}

func (s *service) AddSingletonResource(basePath string, r Resource) {
	// GET /basePath
	addGetRoute(s, basePath, false, r)

	// PUT /basePath
	addUpdateRoute(s, basePath, false, r)

	// POST /basePath/{action}
	addActionRoute(s, basePath, false, r)
}

func (s *service) AddCollectionResource(basePath string, r Resource) {
	// GET /basePath
	addListRoute(s, basePath, r)

	// GET /basePath/{id}
	addGetRoute(s, basePath, true, r)

	// POST /basePath
	addCreateRoute(s, basePath, r)

	// PUT /basePath/{id}
	addUpdateRoute(s, basePath, true, r)

	// DELETE /basePath
	addDeleteRoute(s, basePath, false, r)

	// DELETE /basePath/{id}
	addDeleteRoute(s, basePath, true, r)

	// POST /basePath/{action}
	addActionRoute(s, basePath, false, r)

	// POST /basePath/{id}/{action}
	addActionRoute(s, basePath, true, r)
}

func (s *service) AddErrors(errorMessages map[int]string) {
	// Merge the caller's error messages into the service's
	// map. Any duplicated mappings are replaced.
	for code, message := range errorMessages {
		s.errorMessages[code] = message
	}
}

func (s *service) Errors() map[int]string {
	return s.errorMessages
}

func (s *service) Run() error {
	// Install middleware handlers
	if err := s.useRecoveryMiddleware(); err != nil {
		return err
	}

	if err := s.useTraceMiddleware(); err != nil {
		return err
	}

	if err := s.useLoggerMiddleware(); err != nil {
		return err
	}

	if err := s.useNegotiatorMiddleware(); err != nil {
		return err
	}

	// Use our own router
	s.negroni.UseHandler(s.router)

	// Serve
	s.logger.Printf("listening on %s", s.config.Addr)
	return http.ListenAndServe(s.config.Addr, s.negroni)
}

func (s *service) useRecoveryMiddleware() error {
	r := NewRecovery()
	r.Logger = s.logger
	r.StacksVisible = s.config.Debug.Stacks
	s.negroni.Use(r)
	return nil
}

func (s *service) useTraceMiddleware() error {
	if s.config.Trace.Enabled {
		var (
			rec trace.Recorder
			err error
		)

		params := strings.Split(s.config.Trace.Params, ":")
		switch s.config.Trace.Recorder {
		case "yaml":
			if len(params) != 1 {
				return errors.New("yaml trace recorder expects 1 parameter (path)")
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
				return errors.New("dynamodb trace recorder expects 4 parameters (region:table_name:access_key:secret_key)")
			}
			rec, err = dynamorec.New(params[0], params[1], params[2], params[3])
			break
		default:
			err = errors.New(fmt.Sprint("unknown trace recorder: ", s.config.Trace.Recorder))
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

	s.negroni.Use(negroni.HandlerFunc(trace.ServeHTTP))
	return nil
}

func (s *service) useLoggerMiddleware() error {
	if !s.config.Debug.Requests {
		return nil
	}

	s.negroni.Use(NewLogger(s.logger))
	return nil
}

func (s *service) useNegotiatorMiddleware() error {
	s.negroni.Use(negotiate.FormatNegotiator([]string{"application/json", "application/xml"}))
	return nil
}
