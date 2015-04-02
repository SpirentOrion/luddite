package luddite

import (
	"fmt"
	"testing"
)

var errorDefs = map[int]ErrorDefinition{
	EcodeUnsupportedMediaType: {"UNSUPPORTED_MEDIA_TYPE", "Error override"},
	EcodeServiceBase:          {"HELLO_WORLD", "Hello world: %s"},
}

func TestNewError(t *testing.T) {
	// Test custom error lookups
	e := NewError(errorDefs, EcodeServiceBase, "foo") // should resolve to errorDefs
	if e != nil {
		if e.Code != errorDefs[EcodeServiceBase].Code {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(errorDefs[EcodeServiceBase].Format, "foo") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	// Test common error lookups
	e = NewError(errorDefs, EcodeInternal, "oh noes!") // should resolve to commonErrorDefs
	if e != nil {
		if e.Code != commonErrorDefs[EcodeInternal].Code {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(commonErrorDefs[EcodeInternal].Format, "oh noes!") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	e = NewError(nil, EcodeInternal, "oh noes!") // should resolve to commonErrorDefs
	if e != nil {
		if e.Code != commonErrorDefs[EcodeInternal].Code {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(commonErrorDefs[EcodeInternal].Format, "oh noes!") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	e = NewError(errorDefs, EcodeUnsupportedMediaType, "blah/blah") // should resolve to errorDefs
	if e != nil {
		if e.Code != errorDefs[EcodeUnsupportedMediaType].Code {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(errorDefs[EcodeUnsupportedMediaType].Format, "blah/blah") {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}

	// Test failsafe
	e = NewError(errorDefs, EcodeServiceBase+1) // should resolve to commonErrorDefs
	if e != nil {
		if e.Code != commonErrorDefs[EcodeUnknown].Code {
			t.Error("error code not set")
		}
		if e.Message != fmt.Sprintf(commonErrorDefs[EcodeUnknown].Format, EcodeServiceBase+1) {
			t.Error("error message not formatted")
		}
	} else {
		t.Error("no error returned")
	}
}
