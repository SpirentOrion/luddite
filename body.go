package luddite

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"mime"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/schema"
)

const (
	ContentTypeMultipartFormData = "multipart/form-data"
	ContentTypeWwwFormUrlencoded = "application/x-www-form-urlencoded"
	ContentTypeCss               = "text/css"
	ContentTypePlain             = "text/plain"
	ContentTypeJson              = "application/json"
	ContentTypeOctetStream       = "application/octet-stream"
	ContentTypeXml               = "application/xml"
	ContentTypeHtml              = "text/html"
	ContentTypeYaml              = "application/yaml"

	maxFormDataMemoryUsage = 10 * 1024 * 1024
)

var formDecoder = schema.NewDecoder()

func init() {
	t := time.Time{}
	formDecoder.RegisterConverter(t, ConvertTime)
}

func ConvertTime(value string) reflect.Value {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return reflect.ValueOf(t)
	}
	return reflect.Value{}
}

func ReadRequest(req *http.Request, v interface{}) error {
	ct := req.Header.Get(HeaderContentType)
	switch mt, _, _ := mime.ParseMediaType(ct); mt {
	case ContentTypeMultipartFormData:
		if err := req.ParseMultipartForm(maxFormDataMemoryUsage); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		if err := formDecoder.Decode(v, req.PostForm); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		return nil
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
		case ContentTypeOctetStream:
			switch v.(type) {
			case []byte:
				b = v.([]byte)
			case string:
				b = []byte(v.(string))
			default:
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	rw.WriteHeader(status)
	if b != nil {
		_, err = rw.Write(b)
	}
	return
}
