package luddite

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

func readRequest(req *http.Request, r Resource) (interface{}, error) {
	switch ct := req.Header.Get("Content-Type"); ct {
	case "application/json":
		decoder := json.NewDecoder(req.Body)
		v := r.New()
		err := decoder.Decode(v)
		if err != nil {
			err = NewError(EcodeDeserializationFailed, err)
		}
		return v, err
	case "application/xml":
		decoder := xml.NewDecoder(req.Body)
		v := r.New()
		err := decoder.Decode(v)
		if err != nil {
			err = NewError(EcodeDeserializationFailed, err)
		}
		return v, err
	default:
		return nil, NewError(EcodeUnsupportedMediaType, ct)
	}
}

func writeResponse(rw http.ResponseWriter, status int, v interface{}) (err error) {
	var b []byte
	if v != nil {
		switch v.(type) {
		case *Error:
			break
		case error:
			v = NewError(EcodeInternal, v)
			break
		}
		switch rw.Header().Get("Content-Type") {
		case "application/json":
			b, err = json.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = json.Marshal(NewError(EcodeSerializationFailed, err))
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
				b, err = xml.Marshal(NewError(EcodeSerializationFailed, err))
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
