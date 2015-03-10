package luddite

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

func readRequest(req *http.Request, r Resource) (interface{}, error) {
	switch ct := req.Header.Get("Content-Type"); ct {
	case "application/json":
		decoder := json.NewDecoder(req.Body)
		v := r.New()
		return v, decoder.Decode(v)
	case "application/xml":
		decoder := xml.NewDecoder(req.Body)
		v := r.New()
		return v, decoder.Decode(v)
	default:
		return nil, &ErrorResponse{Code: -1, Message: fmt.Sprint("Cannot decode ", ct)}
	}
}

func writeResponse(rw http.ResponseWriter, status int, v interface{}) (err error) {
	var b []byte
	if v != nil {
		switch v.(type) {
		case *ErrorResponse:
			break
		case error:
			v = &ErrorResponse{Code: -1, Message: fmt.Sprint(v)}
			break
		}
		switch rw.Header().Get("Content-Type") {
		case "application/json":
			b, err = json.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = json.Marshal(&ErrorResponse{Code: -1, Message: fmt.Sprint(err)})
				if err != nil {
					rw.Write(b)
				}
				return
			}
			break
		case "application/xml":
			b, err = xml.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = xml.Marshal(&ErrorResponse{Code: -1, Message: fmt.Sprint(err)})
				if err != nil {
					rw.Write(b)
				}
				return
			}
			break
		}
	}
	rw.WriteHeader(status)
	if b != nil {
		_, err = rw.Write(b)
	}
	return
}
