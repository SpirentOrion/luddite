package luddite

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/gorilla/schema"
)

const (
	ContentTypeWwwFormUrlencoded = "application/x-www-form-urlencoded"
	ContentTypeJson              = "application/json"
	ContentTypeXml               = "application/xml"
	ContentTypeHtml              = "text/html"
)

var formDecoder = schema.NewDecoder()

func readRequest(req *http.Request, r Resource) (interface{}, error) {
	switch ct := req.Header.Get(HeaderContentType); ct {
	case ContentTypeWwwFormUrlencoded:
		if err := req.ParseForm(); err != nil {
			return nil, NewError(nil, EcodeDeserializationFailed, err)
		}
		v := r.New()
		if err := formDecoder.Decode(v, req.PostForm); err != nil {
			return nil, NewError(nil, EcodeDeserializationFailed, err)
		}
		return v, nil
	case ContentTypeJson:
		decoder := json.NewDecoder(req.Body)
		v := r.New()
		err := decoder.Decode(v)
		if err != nil {
			return nil, NewError(nil, EcodeDeserializationFailed, err)
		}
		return v, nil
	case ContentTypeXml:
		decoder := xml.NewDecoder(req.Body)
		v := r.New()
		err := decoder.Decode(v)
		if err != nil {
			return nil, NewError(nil, EcodeDeserializationFailed, err)
		}
		return v, nil
	default:
		return nil, NewError(nil, EcodeUnsupportedMediaType, ct)
	}
}

func writeResponse(rw http.ResponseWriter, status int, v interface{}) (err error) {
	var b []byte
	if v != nil {
		switch v.(type) {
		case *Error:
			break
		case error:
			v = NewError(nil, EcodeInternal, v)
			break
		}
		switch rw.Header().Get(HeaderContentType) {
		case ContentTypeJson:
			b, err = json.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = json.Marshal(NewError(nil, EcodeSerializationFailed, err))
				if err != nil {
					rw.Write(b)
				}
				return
			}
			break
		case ContentTypeXml:
			b, err = xml.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = xml.Marshal(NewError(nil, EcodeSerializationFailed, err))
				if err != nil {
					rw.Write(b)
				}
				return
			}
			break
		case ContentTypeHtml:
			switch v.(type) {
			case []byte:
				b = v.([]byte)
				break
			case string:
				b = []byte(v.(string))
				break
			default:
				rw.WriteHeader(http.StatusNotAcceptable)
				break
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
