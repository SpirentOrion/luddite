package luddite

import (
	"net/http"

	"github.com/dimfeld/httptreemux"
)

type topHandler struct {
	globalRouter *httptreemux.ContextMux
	apiRouters   map[int]*httptreemux.ContextMux
}

func (t *topHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, _ http.HandlerFunc) {
	ctx := req.Context()
	SetContextRequestProgress(ctx, "luddite.topHandler.begin")

	// Try a route lookup using the global router. Routes registered
	// here have preference over API version-specific routes and are
	// served w/o regard to requested API version number.
	if lr, ok := t.globalRouter.Lookup(nil, req); ok {
		t.globalRouter.ServeLookupResult(rw, req, lr)
		return
	}

	// Otherwise, dispatch via an API router
	d := contextHandlerDetails(req.Context())
	router := t.apiRouters[d.apiVersion]
	router.ServeHTTP(rw, req)

	SetContextRequestProgress(ctx, "luddite.topHandler.end")
}
