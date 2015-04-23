package luddite

import (
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strconv"

	"github.com/SpirentOrion/httprouter"
	"golang.org/x/net/context"
)

type SchemaHandler struct {
	filePath        string
	filePattern     string
	httpFileHandler http.Handler
}

func NewSchemaHandler(filePath, filePattern string) *SchemaHandler {
	return &SchemaHandler{filePath, filePattern, http.FileServer(http.Dir(filePath))}
}

func (h *SchemaHandler) ServeHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	ps := httprouter.ContextParams(ctx)
	versionStr := ps.ByName("version")

	// Glob schema files for the requested API version
	version, err := strconv.Atoi(versionStr)
	if err != nil || version < 1 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	files, err := filepath.Glob(path.Join(h.filePath, fmt.Sprintf("v%d", version), h.filePattern))
	if err != nil || len(files) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	// Select an appropriate schema file based on the negotiated Content-Type
	file := selectSchemaFile(files, rw.Header().Get(HeaderContentType))
	if file == "" {
		rw.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// Build a narrowed HTTP request for a schema file
	// NB: the new URI path is absolute from the top of h.filePath
	file = fmt.Sprintf("/v%d/%s", version, filepath.Base(file))
	fileReq, err := http.NewRequest("GET", file, nil)
	if err != nil {
		panic(err)
	}

	// Handle via the file server
	h.httpFileHandler.ServeHTTP(rw, fileReq)
}

func selectSchemaFile(files []string, ct string) (file string) {
	file = ""

	// Support both JSON and HTML representations
	var ext string
	switch ct {
	case ContentTypeJson:
		ext = ".json"
		break
	case ContentTypeHtml:
		ext = ".html"
		break
	default:
		return
	}

	for _, f := range files {
		if filepath.Ext(f) == ext {
			file = f
			return
		}
	}
	return
}
