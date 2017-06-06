package luddite

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type SchemaHandler struct {
	fileServer http.Handler
}

func NewSchemaHandler(filePath string) *SchemaHandler {
	return &SchemaHandler{http.FileServer(http.Dir(filePath))}
}

func (h *SchemaHandler) ServeHTTP(rw http.ResponseWriter, req0 *http.Request, params httprouter.Params) {
	versionStr := params.ByName("version")

	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	filepath := params.ByName("filepath")
	file := fmt.Sprintf("/v%d/%s", version, filepath)
	req1, err := http.NewRequest("GET", file, nil)
	if err != nil {
		panic(err)
	}

	switch strings.ToLower(path.Ext(filepath)) {
	case ".yaml", ".yml":
		rw.Header().Set(HeaderContentType, ContentTypeOctetStream)
	default:
		rw.Header().Del(HeaderContentType)
	}

	h.fileServer.ServeHTTP(rw, req1)
}
