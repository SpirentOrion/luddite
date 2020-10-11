package luddite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultContentType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	n := new(negotiatorHandler)
	n.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	res := rw.Result()
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, acceptedContentTypes[0], rw.Header().Get(HeaderContentType))
}

func TestSupportedContentType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderAccept, ContentTypePng)
	rw := httptest.NewRecorder()

	n := new(negotiatorHandler)
	n.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	res := rw.Result()
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, ContentTypePng, rw.Header().Get(HeaderContentType))
}

func TestUnsupportedContentType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderAccept, "x-application/foo")
	rw := httptest.NewRecorder()

	n := new(negotiatorHandler)
	n.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	res := rw.Result()
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NotContains(t, rw.Header(), HeaderContentType)
}
