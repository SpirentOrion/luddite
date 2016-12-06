package luddite

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type SchemaHandler struct {
	filePath string
	fileName string
}

func NewSchemaHandler(filePath, fileName string) *SchemaHandler {
	return &SchemaHandler{filePath, fileName}
}

func (h *SchemaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	versionStr := params.ByName("version")

	// Glob schema files for the requested API version
	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	fs := http.FileServer(http.Dir(path.Join(h.filePath, "v"+versionStr)))
	req.URL.Path = params.ByName("filepath")

	if strings.HasSuffix(req.URL.Path, ".yaml") {
		rw.Header().Set("Content-Type", ContentTypePlain)
	}

	fs.ServeHTTP(rw, req)
}
