package luddite

import (
	"net/http"
	"strconv"
)

// Version is middleware that performs API version selection and enforces the service's
// min/max supported version constraints.
type Version struct {
	minVersion int
	maxVersion int
}

// NewVersion returns a new Version instance.
func NewVersion(minVersion, maxVersion int) *Version {
	return &Version{
		minVersion: minVersion,
		maxVersion: maxVersion,
	}
}

func (v *Version) HandleHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	ctx := req.Context()
	SetContextRequestProgress(ctx, "Version", "HandleHTTP", "begin")

	defaultVersion := v.maxVersion

	// Range check the requested API version and reject requests that fall outside supported version numbers
	version := RequestApiVersion(req, defaultVersion)
	if version < v.minVersion {
		e := NewError(nil, EcodeApiVersionTooOld, v.minVersion)
		WriteResponse(rw, ctx, http.StatusGone, e)
		return
	}
	if version > v.maxVersion {
		e := NewError(nil, EcodeApiVersionTooNew, v.maxVersion)
		WriteResponse(rw, ctx, http.StatusNotImplemented, e)
		return
	}

	// Add the requested API version to response headers (useful for clients when a default version was negotiated)
	rw.Header().Add(HeaderSpirentApiVersion, strconv.Itoa(version))

	// Add the requested API version to handler context so that downstream handlers can access
	d := contextHandlerDetails(req.Context())
	d.apiVersion = version

	next(rw, req)
}
