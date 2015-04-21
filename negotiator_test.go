package luddite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"
)

func TestDefaultContentType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	n := NewNegotiator([]string{"application/json", "application/xml"})

	n.HandleHTTP(context.Background(), rw, req, func(context.Context, http.ResponseWriter, *http.Request) {
		if rw.Header().Get("Content-Type") != "application/json" {
			t.Error("default Content-Type not negotiated")
		}
	})
}
