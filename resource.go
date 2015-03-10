package luddite

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Resource is a set of REST-oriented HTTP method handlers.
//
// All implementations must implement New.
//
// Singleton-style resources may implement Get and Update.  For
// read-only resources, Update may hardcode a 405 response.  Other
// methods will never be called.
//
// Collection-style resources must also implement Id.  They may
// implement List, Get, Create, Update, and Delete.  For read-only
// resources, Create, Update, and Delete may hardcode a 405 response.
//
// NB: because golang's type system sucks, any implementation must
// take care to handle interface{} types correctly. In particular, Id,
// Create, and Update must be able to handle parameters that have the
// same concrete type as returned by New.
type Resource interface {
	// New returns a new instance of the resource.
	New() interface{}
	// Id returns a resource's identifier as a string.
	Id(value interface{}) string
	// List returns an HTTP status code and a slice of resources (or error).
	List(req *http.Request) (int, interface{})
	// Get returns an HTTP status code and a single resource (or error).
	Get(req *http.Request, id string) (int, interface{})
	// Create returns an HTTP status code and a new resource (or error).
	Create(req *http.Request, value interface{}) (int, interface{})
	// Update returns an HTTP status code and an updated resource (or error).
	Update(req *http.Request, id string, value interface{}) (int, interface{})
	// Delete returns an HTTP status code and a deleted resource (or error).
	Delete(req *http.Request, id string) (int, interface{})
}

func addListRoute(router *mux.Router, basePath string, r Resource) {
	router.HandleFunc(basePath, func(rw http.ResponseWriter, req *http.Request) {
		status, v := r.List(req)
		writeResponse(rw, status, v)
	}).Methods("GET").Name(basePath)
}

func addGetRoute(router *mux.Router, itemPath string, r Resource) {
	router.HandleFunc(itemPath, func(rw http.ResponseWriter, req *http.Request) {
		id := mux.Vars(req)["id"]
		status, v := r.Get(req, id)
		writeResponse(rw, status, v)
	}).Methods("GET").Name(itemPath)
}

func addCreateRoute(router *mux.Router, basePath string, itemPath string, r Resource) {
	router.HandleFunc(basePath, func(rw http.ResponseWriter, req *http.Request) {
		v0, err := readRequest(req, r)
		if err != nil {
			writeResponse(rw, http.StatusBadRequest, err)
			return
		}
		status, v1 := r.Create(req, v0)
		if status/100 == 2 {
			url, err := router.Get(itemPath).URL("id", r.Id(v1))
			if err == nil {
				rw.Header().Add("Location", fmt.Sprint(url))
			}
		}
		writeResponse(rw, status, v1)
	}).Methods("POST")
}

func addUpdateRoute(router *mux.Router, itemPath string, r Resource) {
	router.HandleFunc(itemPath, func(rw http.ResponseWriter, req *http.Request) {
		id := mux.Vars(req)["id"]
		v0, err := readRequest(req, r)
		if err != nil {
			writeResponse(rw, http.StatusBadRequest, err)
			return
		}
		if id != r.Id(v0) {
			writeResponse(rw, http.StatusBadRequest, &ErrorResponse{Code: -1, Message: "URL and body identifiers do not match"})
			return
		}
		status, v1 := r.Update(req, id, v0)
		writeResponse(rw, status, v1)
	}).Methods("PUT")
}

func addDeleteRoute(router *mux.Router, itemPath string, r Resource) {
	router.HandleFunc(itemPath, func(rw http.ResponseWriter, req *http.Request) {
		id := mux.Vars(req)["id"]
		status, v := r.Delete(req, id)
		writeResponse(rw, status, v)
	}).Methods("DELETE")
}
