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
	ret, _ := ni.List(nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceCount(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Count(nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceGet(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Get(nil, "")
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceCreate(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Create(nil, nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceUpdate(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Update(nil, "", nil)
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceDelete(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Delete(nil, "")
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestNotImplementedResourceAction(t *testing.T) {
	ni := &NotImplementedResource{}
	ret, _ := ni.Action(nil, "", "")
	if ret != http.StatusNotImplemented {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusNotImplemented, ret))
	}
}

func TestMethodNotAllowedResourceId(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret := ni.Id(nil)
	if ret != "" {
		t.Error("failed to retrieve Id")
	}
}

func TestMethodNotAllowedResourceList(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.List(nil)
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}

func TestMethodNotAllowedResourceCount(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.Count(nil)
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}

func TestMethodNotAllowedResourceGet(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.Get(nil, "")
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}

func TestMethodNotAllowedResourceCreate(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.Create(nil, nil)
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}

func TestMethodNotAllowedResourceUpdate(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.Update(nil, "", nil)
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}

func TestMethodNotAllowedResourceDelete(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.Delete(nil, "")
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}

func TestMethodNotAllowedResourceAction(t *testing.T) {
	ni := &MethodNotAllowedResource{}
	ret, _ := ni.Action(nil, "", "")
	if ret != http.StatusMethodNotAllowed {
		t.Error(fmt.Sprintf("failed, expected %d but was %d", http.StatusMethodNotAllowed, ret))
	}
}
