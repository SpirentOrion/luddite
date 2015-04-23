package luddite

import (
	"net/http"
	"testing"
)

func TestDefaultApiVersion(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	if RequestApiVersion(req, 42) != 42 {
		t.Error("default API version not returned")
	}
}

func TestExplicitApiVersion(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "1")

	if RequestApiVersion(req, 42) != 1 {
		t.Error("explicit API version not returned")
	}
}

func TestInvalidApiVersion(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "blah")

	if RequestApiVersion(req, 42) != 42 {
		t.Error("default API version not returned")
	}
}
