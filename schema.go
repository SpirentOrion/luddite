package luddite

import (
	"fmt"
	"net/http"
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

func (h *SchemaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
	versionStr := params.ByName("version")

	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if strings.HasSuffix(req.URL.Path, ".yaml") {
		rw.Header().Set("Content-Type", ContentTypePlain)
	}

	file := fmt.Sprintf("/v%d/%s", version, params.ByName("filepath"))
	fileReq, err := http.NewRequest("GET", file, nil)
	if err != nil {
		panic(err)
	}

	h.fileServer.ServeHTTP(rw, fileReq)
}
