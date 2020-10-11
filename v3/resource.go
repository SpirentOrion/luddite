package luddite

import (
	"net/http"
	"net/url"
	"path"

	"github.com/dimfeld/httptreemux"
)

const (
	RouteTagSeg1 = "seg1"
	RouteTagSeg2 = "seg2"

	RouteParamAction = RouteTagSeg2 // e.g. in `POST /resource/id/action`
	RouteParamId     = RouteTagSeg1 // e.g. in `GET /resource/id`
)

// CollectionLister is a collection-style resource that returns all its elements
// in response to `GET /resource`.
type CollectionLister interface {
	// List returns an HTTP status code and a slice of resources (or error).
	List(req *http.Request) (int, interface{})
}

// AddListCollectionRoute adds a route for a CollectionLister.
func AddListCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionLister) {
	router.GET(basePath, func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.ListCollectionRoute.begin")
		if status, v := r.List(req); status > 0 {
			SetContextRequestProgress(ctx, "luddite.ListCollectionRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}

// CollectionCounter is a collection-style resource that returns a count of its
// elements in response to `GET /resource/all/count`.
type CollectionCounter interface {
	// Count returns an HTTP status code and a count of resources (or error).
	Count(req *http.Request) (int, interface{})
}

// AddCountCollectionRoute adds a route for a CollectionCounter.
func AddCountCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionCounter) {
	router.GET(path.Join(basePath, "all", "count"), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.CountCollectionRoute.begin")
		if status, v := r.Count(req); status > 0 {
			SetContextRequestProgress(ctx, "luddite.CountCollectionRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}

// CollectionGetter is a collection-style resource that returns a specific element
// in response to `GET /resource/id`.
type CollectionGetter interface {
	// Get returns an HTTP status code and a single resource (or error).
	Get(req *http.Request, id string) (int, interface{})
}

// AddGetCollectionRoute adds a route for a CollectionGetter.
func AddGetCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionGetter) {
	router.GET(path.Join(basePath, ":"+RouteParamId), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.GetCollectionRoute.begin")
		params := httptreemux.ContextParams(ctx)
		if status, v := r.Get(req, params[RouteParamId]); status > 0 {
			SetContextRequestProgress(ctx, "luddite.GetCollectionRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}

// CollectionCreator is a collection-style resource that creates a new element
// in response to `POST /resource`.
type CollectionCreator interface {
	// New returns a new instance of the resource.
	New() interface{}

	// Id returns a resource's identifier as a string.
	Id(value interface{}) string

	// Create returns an HTTP status code and a new resource (or error).
	Create(req *http.Request, value interface{}) (int, interface{})
}

// AddCreateCollectionRoute adds a route for a CollectionCreator.
func AddCreateCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionCreator) {
	router.POST(basePath, func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.CreateCollectionRoute.begin")
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			SetContextRequestProgress(ctx, "luddite.CreateCollectionRoute.body_error")
			_ = WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		if status, v1 := r.Create(req, v0); status > 0 {
			if status == http.StatusCreated {
				url := url.URL{
					Scheme: req.URL.Scheme,
					Host:   req.URL.Host,
					Path:   path.Join(req.URL.Path, r.Id(v1)),
				}
				rw.Header().Add(HeaderLocation, url.String())
			}
			SetContextRequestProgress(ctx, "luddite.CreateCollectionRoute.write")
			_ = WriteResponse(rw, status, v1)
		}
	})
}

// CollectionUpdater is a collection-style resource that updates a specific
// element in response to `PUT /resource/id`.
type CollectionUpdater interface {
	// New returns a new instance of the resource.
	New() interface{}

	// Id returns a resource's identifier as a string.
	Id(value interface{}) string

	// Update returns an HTTP status code and an updated resource (or error).
	Update(req *http.Request, id string, value interface{}) (int, interface{})
}

// AddUpdateCollectionRoute adds a route for a CollectionUpdater.
func AddUpdateCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionUpdater) {
	router.PUT(path.Join(basePath, ":"+RouteParamId), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.UpdateCollectionRoute.begin")
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			SetContextRequestProgress(ctx, "luddite.UpdateCollectionRoute.body_error")
			_ = WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		params := httptreemux.ContextParams(ctx)
		id := params[RouteParamId]
		if id != r.Id(v0) {
			SetContextRequestProgress(ctx, "luddite.UpdateCollectionRoute.id_error")
			_ = WriteResponse(rw, http.StatusBadRequest, NewError(nil, EcodeResourceIdMismatch))
			return
		}
		if status, v1 := r.Update(req, id, v0); status > 0 {
			SetContextRequestProgress(ctx, "luddite.UpdateCollectionRoute.write")
			_ = WriteResponse(rw, status, v1)
		}
	})
}

// CollectionDeleter is a collection-style resource that deletes a specific
// element in response to `DELETE /resource/id`. It may also optionally delete
// the entire collection in response to `DELETE /resource`.
type CollectionDeleter interface {
	// Delete returns an HTTP status code and a deleted resource (or error).
	Delete(req *http.Request, id string) (int, interface{})
}

// AddDeleteCollectionRoute adds routes for a CollectionDeleter.
func AddDeleteCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionDeleter) {
	router.DELETE(path.Join(basePath, ":"+RouteParamId), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.DeleteCollectionRoute.begin")
		params := httptreemux.ContextParams(ctx)
		if status, v := r.Delete(req, params[RouteParamId]); status > 0 {
			SetContextRequestProgress(ctx, "luddite.DeleteCollectionRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
	router.DELETE(basePath, func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.DeleteCollectionRoute.begin")
		if status, v := r.Delete(req, ""); status > 0 {
			SetContextRequestProgress(ctx, "luddite.DeleteCollectionRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}

// CollectionActioner is a collection-style resource that executes an action in
// response to `POST /resource/id/action`.
type CollectionActioner interface {
	// Action returns an HTTP status code and a response body (or error).
	Action(req *http.Request, id string, action string) (int, interface{})
}

// AddActionCollectionRoute adds a route for a CollectionActioner.
func AddActionCollectionRoute(router *httptreemux.ContextMux, basePath string, r CollectionActioner) {
	router.POST(path.Join(basePath, ":"+RouteParamId, ":"+RouteParamAction), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.ActionCollectionRoute.begin")
		params := httptreemux.ContextParams(ctx)
		if status, v := r.Action(req, params[RouteParamId], params[RouteParamAction]); status > 0 {
			SetContextRequestProgress(ctx, "luddite.ActionCollectionRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}

// SingletonGetter is a singleton-style resource that returns a response to `GET
// /resource`.
type SingletonGetter interface {
	// Get returns an HTTP status code and a single resource (or error).
	Get(req *http.Request) (int, interface{})
}

// AddGetSingletonRoute adds a route for a SingletonGetter.
func AddGetSingletonRoute(router *httptreemux.ContextMux, basePath string, r SingletonGetter) {
	router.GET(basePath, func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.GetSingletonRoute.begin")
		if status, v := r.Get(req); status > 0 {
			SetContextRequestProgress(ctx, "luddite.GetSingletonRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}

// SingletonUpdater is a singleton-style resource that is updated in response to
// `PUT /resource`.
type SingletonUpdater interface {
	// New returns a new instance of the resource.
	New() interface{}

	// Update returns an HTTP status code and an updated resource (or error).
	Update(req *http.Request, value interface{}) (int, interface{})
}

// AddUpdateSingletonRoute adds a route for a SingletonUpdater.
func AddUpdateSingletonRoute(router *httptreemux.ContextMux, basePath string, r SingletonUpdater) {
	router.PUT(basePath, func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.UpdateSingletonRoute.begin")
		v0 := r.New()
		if err := ReadRequest(req, v0); err != nil {
			SetContextRequestProgress(ctx, "luddite.UpdateSingletonRoute.body_error")
			_ = WriteResponse(rw, http.StatusBadRequest, err)
			return
		}
		if status, v1 := r.Update(req, v0); status > 0 {
			SetContextRequestProgress(ctx, "luddite.UpdateSingletonRoute.write")
			_ = WriteResponse(rw, status, v1)
		}
	})
}

// SingletonActioner is a singleton-style resource that executes an action in
// response to `POST /resource/action`.
type SingletonActioner interface {
	// Action returns an HTTP status code and a response body (or error).
	Action(req *http.Request, action string) (int, interface{})
}

// AddActionSingletonRoute adds a route for a SingletonActioner.
func AddActionSingletonRoute(router *httptreemux.ContextMux, basePath string, r SingletonActioner) {
	router.POST(path.Join(basePath, ":"+RouteParamAction), func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		SetContextRequestProgress(ctx, "luddite.ActionSingletonRoute.begin")
		params := httptreemux.ContextParams(ctx)
		if status, v := r.Action(req, params[RouteParamAction]); status > 0 {
			SetContextRequestProgress(ctx, "luddite.ActionSingletonRoute.write")
			_ = WriteResponse(rw, status, v)
		}
	})
}
