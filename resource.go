package luddite

import (
	"net/http"
	"net/url"
	"path"

	"github.com/julienschmidt/httprouter"
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

	// Count returns an HTTP status code and a count of resources (or error).
	Count(req *http.Request) (int, interface{})

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

func AddListRoute(router *httprouter.Router, basePath string, r Resource) {
	router.GET(basePath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		SetContextRequestProgress(req.Context(), "luddite", "Router.List", "begin")
		if status, v := r.List(req); status > 0 {
			SetContextRequestProgress(req.Context(), "luddite", "Router.List", "write")
			WriteResponse(rw, status, v)
		}
	})
}

func AddCountRoute(router *httprouter.Router, basePath string, r Resource) {
	router.GET(path.Join(basePath, "all", "count"), func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		SetContextRequestProgress(req.Context(), "luddite", "Router.Count", "begin")
		if status, v := r.Count(req); status > 0 {
			SetContextRequestProgress(req.Context(), "luddite", "Router.Count", "write")
			WriteResponse(rw, status, v)
		}
	})
}

func AddGetRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, ":id")
	} else {
		itemPath = basePath
	}
	router.GET(itemPath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		SetContextRequestProgress(req.Context(), "luddite", "Router.Get", "begin")
		if status, v := r.Get(req, id); status > 0 {
			SetContextRequestProgress(req.Context(), "luddite", "Router.Get", "write")
			WriteResponse(rw, status, v)
		}
	})
}

func AddCreateRoute(router *httprouter.Router, basePath string, r Resource) {
	router.POST(basePath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		SetContextRequestProgress(req.Context(), "luddite", "Router.Create", "begin")
		if status, v1 := r.Create(req, v0); status > 0 {
			if status == http.StatusCreated {
				url := url.URL{
					Scheme: req.URL.Scheme,
					Host:   req.URL.Host,
					Path:   path.Join(basePath, r.Id(v1)),
				}
				rw.Header().Add(HeaderLocation, url.String())
			}
			SetContextRequestProgress(req.Context(), "luddite", "Router.Create", "write")
			WriteResponse(rw, status, v1)
		}
	})
}

func AddUpdateRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, ":id")
	} else {
		itemPath = basePath
	}
	router.PUT(itemPath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		if withId && id != r.Id(v0) {
			WriteResponse(rw, http.StatusBadRequest, NewError(nil, EcodeResourceIdMismatch))
			return
		}
		SetContextRequestProgress(req.Context(), "luddite", "Router.Update", "begin")
		if status, v1 := r.Update(req, id, v0); status > 0 {
			SetContextRequestProgress(req.Context(), "luddite", "Router.Update", "write")
			WriteResponse(rw, status, v1)
		}
	})
}

func AddDeleteRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, ":id")
	} else {
		itemPath = basePath
	}
	router.DELETE(itemPath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		SetContextRequestProgress(req.Context(), "luddite", "Router.Delete", "begin")
		if status, v := r.Delete(req, id); status > 0 {
			SetContextRequestProgress(req.Context(), "luddite", "Router.Delete", "write")
			WriteResponse(rw, status, v)
		}
	})
}

func AddActionRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var actionPath string
	if withId {
		actionPath = path.Join(basePath, ":id", ":action")
	} else {
		actionPath = path.Join(basePath, ":action")
	}
	router.POST(actionPath, func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		var id string
		if withId {
			id = params.ByName("id")
		}
		action := params.ByName("action")
		SetContextRequestProgress(req.Context(), "luddite", "Router.Action", "begin")
		if status, v := r.Action(req, id, action); status > 0 {
			SetContextRequestProgress(req.Context(), "luddite", "Router.Action", "write")
			WriteResponse(rw, status, v)
		}
	})
}

// NotImplementedResource returns HTTP 501 NotImplemented for all HTTP methods.
type NotImplementedResource struct {
}

func (r *NotImplementedResource) New() interface{} {
	return &NotImplementedResource{}
}

func (r *NotImplementedResource) Id(value interface{}) string {
	return ""
}

func (r *NotImplementedResource) List(req *http.Request) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Count(req *http.Request) (int, interface{}) {
	return http.StatusNotImplemented, 0
}

func (r *NotImplementedResource) Get(req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Create(req *http.Request, value interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Update(req *http.Request, id string, value interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Delete(req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Action(req *http.Request, id, action string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}
