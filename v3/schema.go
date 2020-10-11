package luddite

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/dimfeld/httptreemux"
)

type schemaHandler struct {
	fileServer http.Handler
}

func newSchemaHandler(fs http.FileSystem) http.Handler {
	return &schemaHandler{
		fileServer: http.FileServer(fs),
	}
}

func (h *schemaHandler) ServeHTTP(rw http.ResponseWriter, req0 *http.Request) {
	// Transform the request path to a path compatible with the schema directory
	params := httptreemux.ContextParams(req0.Context())

	versionStr := params["version"]
	if len(versionStr) < 2 || versionStr[0] != 'v' {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	version, err := strconv.Atoi(versionStr[1:])
	if err != nil || version < 1 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	filepath := params["filepath"]
	req1, err := http.NewRequest("GET", fmt.Sprintf("/v%d/%s", version, filepath), nil)
	if err != nil {
		panic(err)
	}

	switch strings.ToLower(path.Ext(filepath)) {
	case ".yaml", ".yml":
		rw.Header().Set(HeaderContentType, ContentTypeOctetStream)
	default:
		rw.Header().Del(HeaderContentType)
	}

	// Delegate request handling to the standard fileserver
	h.fileServer.ServeHTTP(rw, req1)
}
