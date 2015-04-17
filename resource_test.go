package luddite

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNotImplementedResourceId(t *testing.T) {
	ni := &NotImplementedResource{}
	ret := ni.Id(nil)
	if ret != "" {
		t.Error("failed to retrieve Id")
	}
}

func TestNotImplementedResourceList(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.List(nil, nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceGet(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Get(nil, nil, "")
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceCreate(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Create(nil, nil, nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceUpdate(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Update(nil, nil, "", nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceDelete(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Delete(nil, nil, "")
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceAction(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Action(nil, nil, "", "")
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}
