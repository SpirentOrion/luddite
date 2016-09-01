package luddite

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMinApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "1")
	rw := httptest.NewRecorder()

	v := NewVersion(2, 42)

	v.HandleHTTP(rw, req, func(http.ResponseWriter, *http.Request) {
		if rw.Code != http.StatusGone {
			t.Error("expected 410/Gone response for outdated version")
		}
	})
}

func TestMaxApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "43")
	rw := httptest.NewRecorder()

	v := NewVersion(2, 42)

	v.HandleHTTP(rw, req, func(http.ResponseWriter, *http.Request) {
		if rw.Code != http.StatusNotImplemented {
			t.Error("expected 501/Not Implemented response for future version")
		}
	})
}

func TestApiVersionContext(t *testing.T) {
	req0, _ := http.NewRequest("GET", "/", nil)
	req0.Header.Add(HeaderSpirentApiVersion, "1")
	req0 = req0.WithContext(withHandlerDetails(req0.Context(), &handlerDetails{}))
	rw := httptest.NewRecorder()

	v := NewVersion(1, 1)

	v.HandleHTTP(rw, req0, func(_ http.ResponseWriter, req1 *http.Request) {
		ctx := req1.Context()
		if ContextApiVersion(ctx) != 1 {
			t.Error("missing API version in context")
		}
	})

	if _, ok := rw.HeaderMap[HeaderSpirentApiVersion]; !ok {
		t.Errorf("missing %s header in response", HeaderSpirentApiVersion)
	}
}
