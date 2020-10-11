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
	ContentTypeCss               = "text/css"
	ContentTypeCsv               = "text/csv"
	ContentTypeGif               = "image/gif"
	ContentTypeHtml              = "text/html"
	ContentTypeJson              = "application/json"
	ContentTypeMsgpack           = "application/msgpack"
	ContentTypeMultipartFormData = "multipart/form-data"
	ContentTypeOctetStream       = "application/octet-stream"
	ContentTypePlain             = "text/plain"
	ContentTypePng               = "image/png"
	ContentTypeProtobuf          = "application/protobuf"
	ContentTypeWwwFormUrlencoded = "application/x-www-form-urlencoded"
	ContentTypeXml               = "application/xml"

	maxFormDataMemoryUsage = 10 * 1024 * 1024
)

var FormDecoder = schema.NewDecoder()

func init() {
	t := time.Time{}
	FormDecoder.RegisterConverter(t, convertTime)
}

func convertTime(value string) reflect.Value {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return reflect.ValueOf(t)
	}
	return reflect.Value{}
}

// ReadRequest deserializes a request body according to the Content-Type header.
func ReadRequest(req *http.Request, v interface{}) error {
	ct := req.Header.Get(HeaderContentType)
	switch mt, _, _ := mime.ParseMediaType(ct); mt {
	case ContentTypeMultipartFormData:
		if err := req.ParseMultipartForm(maxFormDataMemoryUsage); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		if err := FormDecoder.Decode(v, req.PostForm); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		return nil
	case ContentTypeWwwFormUrlencoded:
		if err := req.ParseForm(); err != nil {
			return NewError(nil, EcodeDeserializationFailed, err)
		}
		if err := FormDecoder.Decode(v, req.PostForm); err != nil {
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
	case "":
		return nil
	default:
		return NewError(nil, EcodeUnsupportedMediaType, ct)
	}
}

// WriteResponse serializes a response body according to the negotiated Content-Type.
func WriteResponse(rw http.ResponseWriter, status int, v interface{}) (err error) {
	var inhibitResp bool
	if rw.Header().Get(HeaderSpirentInhibitResponse) != "" {
		if status/100 == 2 {
			inhibitResp = true
		} else {
			rw.Header().Del(HeaderSpirentInhibitResponse)
		}
	}
	var b []byte
	if v != nil {
		switch v.(type) {
		case *Error:
		case error:
			v = NewError(nil, EcodeInternal, v)
		}
		switch ct := rw.Header().Get(HeaderContentType); ct {
		case ContentTypeJson:
			b, err = json.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = json.Marshal(NewError(nil, EcodeSerializationFailed, err))
				if err != nil {
					_, _ = rw.Write(b)
				}
				return
			}
		case ContentTypeXml:
			b, err = xml.Marshal(v)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				b, err = xml.Marshal(NewError(nil, EcodeSerializationFailed, err))
				if err != nil {
					_, _ = rw.Write(b)
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
						_, _ = rw.Write(b)
					}
					return
				}
				esc := new(bytes.Buffer)
				json.HTMLEscape(esc, b)
				b = esc.Bytes()
			}
		default:
			switch v.(type) {
			case []byte:
				b = v.([]byte)
				if ct == "" {
					rw.Header().Set(HeaderContentType, ContentTypeOctetStream)
				}
			case string:
				b = []byte(v.(string))
				if ct == "" {
					rw.Header().Set(HeaderContentType, ContentTypePlain)
				}
			default:
				rw.WriteHeader(http.StatusNotAcceptable)
				return
			}
		}
	}
	if inhibitResp {
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	rw.WriteHeader(status)
	if b != nil {
		_, err = rw.Write(b)
	}
	return
}
