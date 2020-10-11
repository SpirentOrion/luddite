package luddite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonPositiveApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "0")
	rw := httptest.NewRecorder()
	rw.Header().Set(HeaderContentType, ContentTypeJson)

	v := &versionHandler{
		minVersion: 2,
		maxVersion: 42,
	}
	v.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	res := rw.Result()
	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestMinApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "1")
	rw := httptest.NewRecorder()
	rw.Header().Set(HeaderContentType, ContentTypeJson)

	v := &versionHandler{
		minVersion: 2,
		maxVersion: 42,
	}
	v.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	res := rw.Result()
	require.Equal(t, http.StatusGone, res.StatusCode)
}

func TestMaxApiVersionConstraint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "43")
	rw := httptest.NewRecorder()
	rw.Header().Set(HeaderContentType, ContentTypeJson)

	v := &versionHandler{
		minVersion: 2,
		maxVersion: 42,
	}
	v.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	res := rw.Result()
	require.Equal(t, http.StatusNotImplemented, res.StatusCode)
}

func TestApiVersionContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add(HeaderSpirentApiVersion, "1")
	req = req.WithContext(withHandlerDetails(req.Context(), &handlerDetails{}))
	rw := httptest.NewRecorder()
	rw.Header().Set(HeaderContentType, ContentTypeJson)

	v := &versionHandler{
		minVersion: 1,
		maxVersion: 1,
	}
	v.ServeHTTP(rw, req, func(_ http.ResponseWriter, _ *http.Request) {})

	require.Equal(t, 1, ContextApiVersion(req.Context()))
	res := rw.Result()
	require.Equal(t, "1", res.Header.Get(HeaderSpirentApiVersion))
}
