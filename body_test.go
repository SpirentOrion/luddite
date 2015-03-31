package luddite

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type sample struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Flag      bool      `json:"flag"`
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	sampleId       = 1234
	sampleName     = "dave"
	sampleData     = "Hello world"
	sampleJsonBody = "{\"id\":1234,\"name\":\"dave\",\"flag\":true,\"data\":\"SGVsbG8gd29ybGQ=\",\"timestamp\":\"2015-03-18T14:30:00Z\"}"
)

var (
	sampleTimestamp = time.Date(2015, 3, 18, 14, 30, 0, 0, time.UTC)
)

type sampleResource struct {
}

func (r *sampleResource) New() interface{} {
	return &sample{}
}

func (r *sampleResource) Id(value interface{}) string {
	return strconv.Itoa(value.(*sample).Id)
}

func (r *sampleResource) List(s Service, req *http.Request) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *sampleResource) Get(s Service, req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *sampleResource) Create(s Service, req *http.Request, value interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *sampleResource) Update(s Service, req *http.Request, id string, value interface{}) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *sampleResource) Delete(s Service, req *http.Request, id string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func (r *sampleResource) Action(s Service, req *http.Request, id string, action string) (int, interface{}) {
	return http.StatusNotImplemented, nil
}

func TestReadJson(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", strings.NewReader(sampleJsonBody))
	req.Header["Content-Type"] = []string{"application/json"}

	v, err := readRequest(req, &sampleResource{})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	s := v.(*sample)
	if s.Id != sampleId {
		t.Error("JSON int deserialization failed")
	}
	if s.Name != sampleName {
		t.Error("JSON string deserialization failed")
	}
	if !s.Flag {
		t.Error("JSON bool deserialization failed")
	}
	if !bytes.Equal(s.Data, []byte(sampleData)) {
		t.Error("JSON binary deserialization failed")
	}
	if s.Timestamp != sampleTimestamp {
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
	rw.Header().Add("Content-Type", "application/json")

	err := writeResponse(rw, 200, s)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if rw.Code != 200 {
		t.Error("status code never written")
	}

	if rw.Body != nil {
		body := string(rw.Body.String())
		if body != sampleJsonBody {
			t.Errorf("JSON serialization failed, got: %s, expected: %s\n", body, sampleJsonBody)
		}
	} else {
		t.Error("body never written")
	}
}
