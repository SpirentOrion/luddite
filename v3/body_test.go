package luddite

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type sample struct {
	XMLName   xml.Name  `json:"-" xml:"sample"`
	Id        int       `json:"id" xml:"id"`
	Name      string    `json:"name" xml:"name"`
	Flag      bool      `json:"flag" xml:"flag"`
	Data      []byte    `json:"data" xml:"data"`
	Timestamp time.Time `json:"timestamp" xml:"timestamp"`
}

const (
	sampleId             = 1234
	sampleName           = "dave"
	sampleData           = "Hello world"
	sampleJsonBody       = "{\"id\":1234,\"name\":\"dave\",\"flag\":true,\"data\":\"SGVsbG8gd29ybGQ=\",\"timestamp\":\"2015-03-18T14:30:00Z\"}"
	sampleXmlBody        = "<sample><id>1234</id><name>dave</name><flag>true</flag><data>Hello world</data><timestamp>2015-03-18T14:30:00Z</timestamp></sample>"
	sampleUrlencodedBody = "id=1234&name=dave&flag=true&timestamp=2015-03-18T14:30:00Z"
)

var (
	sampleTimestamp = time.Date(2015, 3, 18, 14, 30, 0, 0, time.UTC)
)

func TestReadJson(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", strings.NewReader(sampleJsonBody))
	req.Header[HeaderContentType] = []string{ContentTypeJson + "; charset=UTF-8"}

	v := &sample{}
	if err := ReadRequest(req, v); err != nil {
		t.Fatal(err)
	}

	if v.Id != sampleId {
		t.Error("JSON int deserialization failed")
	}
	if v.Name != sampleName {
		t.Error("JSON string deserialization failed")
	}
	if !v.Flag {
		t.Error("JSON bool deserialization failed")
	}
	if !bytes.Equal(v.Data, []byte(sampleData)) {
		t.Error("JSON binary deserialization failed")
	}
	if v.Timestamp != sampleTimestamp {
		t.Error("JSON date deserialization failed")
	}
}

func TestWriteJson(t *testing.T) {
	s := &sample{
		Id:        sampleId,
		Name:      sampleName,
		Flag:      true,
		Data:      []byte(sampleData),
		Timestamp: sampleTimestamp,
	}

	rw := httptest.NewRecorder()
	rw.Header().Add(HeaderContentType, ContentTypeJson)

	if err := WriteResponse(rw, http.StatusOK, s); err != nil {
		t.Fatal(err)
	}

	if rw.Code != http.StatusOK {
		t.Error("status code never written")
	}

	if rw.Body != nil {
		if body := string(rw.Body.String()); body != sampleJsonBody {
			t.Errorf("JSON serialization failed, got: %s, expected: %s\n", body, sampleJsonBody)
		}
	} else {
		t.Error("body never written")
	}
}

func TestReadXml(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", strings.NewReader(sampleXmlBody))
	req.Header[HeaderContentType] = []string{ContentTypeXml + "; charset=UTF-8"}

	v := &sample{}
	if err := ReadRequest(req, v); err != nil {
		t.Fatal(err)
	}

	if v.Id != sampleId {
		t.Error("XML int deserialization failed")
	}
	if v.Name != sampleName {
		t.Error("XML string deserialization failed")
	}
	if !v.Flag {
		t.Error("XML bool deserialization failed")
	}
	if !bytes.Equal(v.Data, []byte(sampleData)) {
		t.Error("XML binary deserialization failed")
	}
	if v.Timestamp != sampleTimestamp {
		t.Error("XML date deserialization failed")
	}
}

func TestWriteXml(t *testing.T) {
	s := &sample{
		Id:        sampleId,
		Name:      sampleName,
		Flag:      true,
		Data:      []byte(sampleData),
		Timestamp: sampleTimestamp,
	}

	rw := httptest.NewRecorder()
	rw.Header().Add(HeaderContentType, ContentTypeXml)

	if err := WriteResponse(rw, http.StatusOK, s); err != nil {
		t.Fatal(err)
	}

	if rw.Code != http.StatusOK {
		t.Error("status code never written")
	}

	if rw.Body != nil {
		if body := string(rw.Body.String()); body != sampleXmlBody {
			t.Errorf("XML serialization failed, got: %s, expected: %s\n", body, sampleXmlBody)
		}
	} else {
		t.Error("body never written")
	}
}

func TestWriteHtml(t *testing.T) {
	// Write []byte
	rw := httptest.NewRecorder()
	rw.Header().Add(HeaderContentType, ContentTypeHtml)

	if err := WriteResponse(rw, http.StatusOK, []byte(sampleData)); err != nil {
		t.Fatal(err)
	}

	if rw.Code != http.StatusOK {
		t.Error("status code never written")
	}

	if rw.Body != nil {
		if body := string(rw.Body.String()); body != sampleData {
			t.Errorf("HTML body write failed, got: %s, expected: %s\n", body, sampleData)
		}
	} else {
		t.Error("body never written")
	}

	// Write string
	rw = httptest.NewRecorder()
	rw.Header().Add(HeaderContentType, ContentTypeHtml)

	if err := WriteResponse(rw, http.StatusOK, sampleData); err != nil {
		t.Fatal(err)
	}

	if rw.Code != http.StatusOK {
		t.Error("status code never written")
	}

	if rw.Body != nil {
		if body := string(rw.Body.String()); body != sampleData {
			t.Errorf("HTML body write failed, got: %s, expected: %s\n", body, sampleData)
		}
	} else {
		t.Error("body never written")
	}

	// Write other type
	s := &sample{
		Id:        sampleId,
		Name:      sampleName,
		Flag:      true,
		Data:      []byte(sampleData),
		Timestamp: sampleTimestamp,
	}

	rw = httptest.NewRecorder()
	rw.Header().Add(HeaderContentType, ContentTypeHtml)

	if err := WriteResponse(rw, http.StatusOK, s); err != nil {
		t.Fatal(err)
	}

	if rw.Code != http.StatusOK {
		t.Error("status code never written")
	}
}

func TestReadUrlencoded(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(sampleUrlencodedBody))
	req.Header[HeaderContentType] = []string{ContentTypeWwwFormUrlencoded}

	v := &sample{}
	if err := ReadRequest(req, v); err != nil {
		t.Fatal(err)
	}

	if v.Id != sampleId {
		t.Error("Urlencoded int deserialization failed")
	}
	if v.Name != sampleName {
		t.Error("Urlencoded string deserialization failed")
	}
	if !v.Flag {
		t.Error("Urlencoded bool deserialization failed")
	}
	if v.Timestamp != sampleTimestamp {
		t.Error("Urlencoded date deserialization failed")
	}
}
