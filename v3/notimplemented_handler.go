package luddite

import (
	"net/http"
)

type notImplementedResource struct{}

func NewNotImplementedResource() *notImplementedResource {
	return &notImplementedResource{}
}

func (r *notImplementedResource) New() interface{} {
	return &struct{}{}
}

func (r *notImplementedResource) Id(value interface{}) string {
	return ""
}

func (r *notImplementedResource) List(req *http.Request) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *notImplementedResource) Get(req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *notImplementedResource) Create(req *http.Request, v interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *notImplementedResource) Update(req *http.Request, id string, v interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *notImplementedResource) Delete(req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *notImplementedResource) Action(req *http.Request, id, action string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}
