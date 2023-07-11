package luddite

import (
	"net/http"
)

type notImplementedResource struct{}

type notImplementedType struct {
	Id string
}

func NewNotImplementedResource() *notImplementedResource {
	return &notImplementedResource{}
}

func (r *notImplementedResource) New() interface{} {
	return new(notImplementedType)
}

func (r *notImplementedResource) Id(value interface{}) string {
	v := value.(*notImplementedType)
	return v.Id
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
