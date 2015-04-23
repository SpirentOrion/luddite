package luddite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

func TestMinApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "1")
	rw := httptest.NewRecorder()

	c := NewContext(nil, 2, 42)

	c.HandleHTTP(context.Background(), rw, req, func(context.Context, http.ResponseWriter, *http.Request) {
		if rw.Code != http.StatusGone {
			t.Error("expected 410/Gone response for outdated version")
		}
	})
}

func TestMaxApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "43")
	rw := httptest.NewRecorder()

	c := NewContext(nil, 2, 42)

	c.HandleHTTP(context.Background(), rw, req, func(context.Context, http.ResponseWriter, *http.Request) {
		if rw.Code != http.StatusNotImplemented {
			t.Error("expected 501/Not Implemented response for future version")
		}
	})
}

func TestApiVersionContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "1")
	rw := httptest.NewRecorder()

	c := NewContext(nil, 1, 1)

	c.HandleHTTP(context.Background(), rw, req, func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) {
		if ContextApiVersion(ctx) != 1 {
			t.Error("missing API version in context")
		}
	})

	if _, ok := rw.HeaderMap[HeaderSpirentApiVersion]; !ok {
		t.Errorf("missing %s header in response", HeaderSpirentApiVersion)
	}
}
