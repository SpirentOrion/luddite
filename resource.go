package luddite

import (
	"net/http"
	"net/url"
	"path"

	"github.com/SpirentOrion/httprouter"
	"golang.org/x/net/context"
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
	List(ctx context.Context, req *http.Request) (int, interface{})

	// Get returns an HTTP status code and a single resource (or error).
	Get(ctx context.Context, req *http.Request, id string) (int, interface{})

	// Create returns an HTTP status code and a new resource (or error).
	Create(ctx context.Context, req *http.Request, value interface{}) (int, interface{})

	// Update returns an HTTP status code and an updated resource (or error).
	Update(ctx context.Context, req *http.Request, id string, value interface{}) (int, interface{})

	// Delete returns an HTTP status code and a deleted resource (or error).
	Delete(ctx context.Context, req *http.Request, id string) (int, interface{})

	// Action returns an HTTP status code and a response body (or error).
	Action(ctx context.Context, req *http.Request, id string, action string) (int, interface{})
}

func AddListRoute(router *httprouter.Router, basePath string, r Resource) {
	router.GET(basePath, func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		status, v := r.List(ctx, req)
		WriteResponse(rw, status, v)
	})
}

func AddGetRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, ":id")
	} else {
		itemPath = basePath
	}
	router.GET(itemPath, func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		vars := httprouter.ContextParams(ctx)
		id := vars.ByName("id")
		status, v := r.Get(ctx, req, id)
		WriteResponse(rw, status, v)
	})
}

func AddCreateRoute(router *httprouter.Router, basePath string, r Resource) {
	router.POST(basePath, func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		status, v1 := r.Create(ctx, req, v0)
		if status == http.StatusCreated {
			url := url.URL{
				Scheme: req.URL.Scheme,
				Host:   req.URL.Host,
				Path:   path.Join(basePath, r.Id(v1)),
			}
			rw.Header().Add(HeaderLocation, url.String())
		}
		WriteResponse(rw, status, v1)
	})
}

func AddUpdateRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, ":id")
	} else {
		itemPath = basePath
	}
	router.PUT(itemPath, func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		vars := httprouter.ContextParams(ctx)
		id := vars.ByName("id")
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		if withId && id != r.Id(v0) {
			WriteResponse(rw, http.StatusBadRequest, NewError(nil, EcodeResourceIdMismatch))
			return
		}
		status, v1 := r.Update(ctx, req, id, v0)
		WriteResponse(rw, status, v1)
	})
}

func AddDeleteRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var itemPath string
	if withId {
		itemPath = path.Join(basePath, ":id")
	} else {
		itemPath = basePath
	}
	router.DELETE(itemPath, func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		vars := httprouter.ContextParams(ctx)
		id := vars.ByName("id")
		status, v := r.Delete(ctx, req, id)
		WriteResponse(rw, status, v)
	})
}

func AddActionRoute(router *httprouter.Router, basePath string, withId bool, r Resource) {
	var actionPath string
	if withId {
		actionPath = path.Join(basePath, ":id", ":action")
	} else {
		actionPath = path.Join(basePath, ":action")
	}
	router.POST(actionPath, func(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
		vars := httprouter.ContextParams(ctx)
		var id string
		if withId {
			id = vars.ByName("id")
		}
		action := vars.ByName("action")
		status, v := r.Action(ctx, req, id, action)
		WriteResponse(rw, status, v)
	})
}

// NotImplementedResource returns HTTP 501 NotImplemented for all HTTP methods.
type NotImplementedResource struct {
	Resource
}

func (r *NotImplementedResource) New() interface{} {
	return &NotImplementedResource{}
}

func (r *NotImplementedResource) Id(value interface{}) string {
	return ""
}

func (r *NotImplementedResource) List(ctx context.Context, req *http.Request) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Get(ctx context.Context, req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Create(ctx context.Context, req *http.Request, value interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Update(ctx context.Context, req *http.Request, id string, value interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Delete(ctx context.Context, req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *NotImplementedResource) Action(ctx context.Context, req *http.Request, id, action string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}
