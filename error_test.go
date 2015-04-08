package luddite

import (
	"fmt"
	"testing"
)

const (
	EcodeHelloWorld = "HELLO_WORLD"
)

var errorMap = map[string]string{
	EcodeUnsupportedMediaType: "Error override",
	EcodeHelloWorld:           "Hello world: %s",
}

func TestNewError(t *testing.T) {
	// Test custom error lookups
	e := NewError(errorMap, EcodeHelloWorld, "foo") // should resolve to errorMap
	if e != nil {
		if e.Code != EcodeHelloWorld {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(errorMap[EcodeHelloWorld], "foo") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	// Test common error lookups
	e = NewError(errorMap, EcodeInternal, "oh noes!") // should resolve to commonErrorMap
	if e != nil {
		if e.Code != EcodeInternal {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(commonErrorMap[EcodeInternal], "oh noes!") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	e = NewError(nil, EcodeInternal, "oh noes!") // should resolve to commonErrorMap
	if e != nil {
		if e.Code != EcodeInternal {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(commonErrorMap[EcodeInternal], "oh noes!") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	e = NewError(errorMap, EcodeUnsupportedMediaType, "blah/blah") // should resolve to errorMap
	if e != nil {
		if e.Code != EcodeUnsupportedMediaType {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(errorMap[EcodeUnsupportedMediaType], "blah/blah") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	// Test failsafe
	e = NewError(errorMap, "SOMETHING_INVALID") // should resolve to commonErrorMap
	if e != nil {
		if e.Code != EcodeUnknown {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(commonErrorMap[EcodeUnknown], "SOMETHING_INVALID") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}
}
