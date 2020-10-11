package luddite

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// TestDispatch allows external code to test its own handlers, complete with
// mocked luddite handler details.
func TestDispatch(rw http.ResponseWriter, req *http.Request, h http.Handler) {
	s := &Service{
		config:        new(ServiceConfig),
		defaultLogger: log.New(),
	}

	res := responseWriterPool.Get().(*responseWriter)
	defer responseWriterPool.Put(res)
	res.init(rw)

	d := &handlerDetails{
		s:          s,
		rw:         res,
		request:    req,
		apiVersion: 1,
	}

	ctx := withHandlerDetails(req.Context(), d)
	h.ServeHTTP(res, req.WithContext(ctx))
}
