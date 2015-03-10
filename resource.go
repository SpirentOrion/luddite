package luddite

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

// Resource is a set of REST-oriented HTTP method handlers.
//
// All resources must implement New.
//
// Singleton-style resources may implement Get and Update.  For
// read-only singleton resources, Update may hardcode a 405 response.
// The Id, List, Create, and Delete methods will never be called.
//
// Collection-style resources must also implement Id.  They may
// implement List, Get, Create, Update, and Delete.  For read-only
// collection resources, Create, Update, and Delete may hardcode a 405
// response.
//
// Any resource may implement POST actions.  These are dispatched to
// the Action method, which should switch on the action name.
//
// NB: because golang's type system sucks, any implementation must
// take care to handle interface{} types correctly. In particular, Id,
// Create, and Update must be able to handle value parameters that
// have the same concrete type as returned by New.
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
	// Action returns an HTTP status code and a response body (or error).
	Action(req *http.Request, id string, action string) (int, interface{})
}

type ResourceActionHandler func(*http.Request) (int, interface{})

func addListRoute(router *mux.Router, basePath string, r Resource) {
	router.HandleFunc(basePath, func(rw http.ResponseWriter, req *http.Request) {
		status, v := r.List(req)
		writeResponse(rw, status, v)
	}).Methods("GET")
}

func addGetRoute(router *mux.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, "{id}")
	} else {
		itemPath = basePath
	}
	route := router.HandleFunc(itemPath, func(rw http.ResponseWriter, req *http.Request) {
		id, _ := mux.Vars(req)["id"]
		status, v := r.Get(req, id)
		writeResponse(rw, status, v)
	}).Methods("GET")
	if withId {
		route.Name(itemPath)
	}
}

func addCreateRoute(router *mux.Router, basePath string, r Resource) {
	itemPath := path.Join(basePath, "{id}")
	router.HandleFunc(basePath, func(rw http.ResponseWriter, req *http.Request) {
		v0, err := readRequest(req, r)
		if err != nil {
			writeResponse(rw, http.StatusBadRequest, err)
			return
		}
		status, v1 := r.Create(req, v0)
		if status == http.StatusCreated {
			url, err := router.Get(itemPath).URL("id", r.Id(v1))
			if err == nil {
				rw.Header().Add("Location", fmt.Sprint(url))
			}
		}
		writeResponse(rw, status, v1)
	}).Methods("POST")
}

func addUpdateRoute(router *mux.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, "{id}")
	} else {
		itemPath = basePath
	}
	router.HandleFunc(itemPath, func(rw http.ResponseWriter, req *http.Request) {
		id, _ := mux.Vars(req)["id"]
		v0, err := readRequest(req, r)
		if err != nil {
			writeResponse(rw, http.StatusBadRequest, err)
			return
		}
		if withId && id != r.Id(v0) {
			writeResponse(rw, http.StatusBadRequest, NewError(EcodeIdentifierMismatch))
			return
		}
		status, v1 := r.Update(req, id, v0)
		writeResponse(rw, status, v1)
	}).Methods("PUT")
}

func addDeleteRoute(router *mux.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, "{id}")
	} else {
		itemPath = basePath
	}
	router.HandleFunc(itemPath, func(rw http.ResponseWriter, req *http.Request) {
		id, _ := mux.Vars(req)["id"]
		status, v := r.Delete(req, id)
		writeResponse(rw, status, v)
	}).Methods("DELETE")
}

func addActionRoute(router *mux.Router, basePath string, withId bool, r Resource) {
	var actionPath string
	if withId {
		actionPath = path.Join(basePath, "{id}", "{action}")
	} else {
		actionPath = path.Join(basePath, "{action}")
	}
	router.HandleFunc(actionPath, func(rw http.ResponseWriter, req *http.Request) {
		id, _ := mux.Vars(req)["id"]
		action, _ := mux.Vars(req)["action"]
		status, v := r.Action(req, id, action)
		writeResponse(rw, status, v)
	}).Methods("POST")
}
