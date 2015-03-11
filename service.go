package luddite

import (
	"log"
	"net/http"
	"os"

	"github.com/K-Phoen/http-negotiate/negotiate"
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
	config *ServiceConfig
	router *mux.Router
}

func NewService(config *ServiceConfig) Service {
	return &service{
		config: config,
		router: mux.NewRouter(),
	}
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
	// POST /basePath/{action}
	addActionRoute(s.router, basePath, false, r)
	// POST /basePath/{id}/{action}
	addActionRoute(s.router, basePath, true, r)
}

func (s *service) Run() error {
	cfg := s.config
	logger := log.New(os.Stderr, cfg.Log.Prefix, log.LstdFlags)

	// Create a new negroni instance for the service
	n := negroni.New()

	// Install recovery middleware
	rec := NewRecovery()
	rec.Logger = logger
	rec.StacksVisible = cfg.Debug.Stacks
	n.Use(rec)

	// Optionally install logger middleware
	if cfg.Debug.Requests {
		n.Use(NewLogger(logger))
	}

	// Install negotiator middleware
	n.Use(negotiate.FormatNegotiator([]string{"application/json", "application/xml"}))

	// Use our own router
	n.UseHandler(s.router)

	// Serve
	logger.Printf("listening on %s", cfg.Addr)
	return http.ListenAndServe(cfg.Addr, n)
}
