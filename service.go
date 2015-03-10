package luddite

import (
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
	// HTTP server. The addr string takes the same format as
	// http.ListenAndServe.
	Run(addr string)
}

type service struct {
	router *mux.Router
}

func NewService() Service {
	return &service{mux.NewRouter()}
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

func (s *service) Run(addr string) {
	n := negroni.New()
	n.Use(NewRecovery())
	n.Use(NewLogger())
	n.Use(negotiate.FormatNegotiator([]string{"application/json", "application/xml"}))
	n.UseHandler(s.router)
	n.Run(addr)
}
