package luddite

import (
	"bytes"
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

func ReadRequest(req *http.Request, v interface{}) error {
	switch ct := req.Header.Get(HeaderContentType); ct {
	case ContentTypeWwwFormUrlencoded:
		if err := req.ParseForm(); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		if err := formDecoder.Decode(v, req.PostForm); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		return nil
	case ContentTypeJson:
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(v)
		if err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		return nil
	case ContentTypeXml:
		decoder := xml.NewDecoder(req.Body)
		err := decoder.Decode(v)
		if err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		return nil
	default:
		return NewError(nil, EcodeUnsupportedMediaType, ct)
	}
}

func WriteResponse(rw http.ResponseWriter, status int, v interface{}) (err error) {
	var b []byte
	if v != nil {
		switch v.(type) {
		case *Error:
		case error:
			v = NewError(nil, EcodeInternal, v)
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
		case ContentTypeHtml:
			switch v.(type) {
			case []byte:
				b = v.([]byte)
			case string:
				b = []byte(v.(string))
			default:
				b, err = json.Marshal(v)
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					b, err = json.Marshal(NewError(nil, EcodeSerializationFailed, err))
					if err != nil {
						rw.Write(b)
					}
					return
				}
				esc := new(bytes.Buffer)
				json.HTMLEscape(esc, b)
				b = esc.Bytes()
			}
		}
	}
	rw.WriteHeader(status)
	if b != nil {
		_, err = rw.Write(b)
	}
	return
}
