package luddite

import "testing"

var errorMessages = map[int]string{
	EcodeUnsupportedMediaType: "Error string override",
	EcodeServiceBase:          "Hello world: %s",
}

func TestNewError(t *testing.T) {
	// Test custom error string lookups
	e := NewError(errorMessages, EcodeServiceBase, "foo") // should resolve to errorMessages
	if e != nil {
		if e.Code != EcodeServiceBase {
			t.Error("error code not set")
		}
		if e.Message != "Hello world: foo" {
			t.Error("error message not set")
		}
	}

	// Test common error string lookups
	e = NewError(errorMessages, EcodeInternal) // should resolve to commonErrorMessages
	if e != nil {
		if e.Code != EcodeInternal {
			t.Error("error code not set")
		}
		if e.Message != commonErrorMessages[EcodeInternal] {
			t.Error("error message not set")
		}
	}

	e = NewError(nil, EcodeInternal) // should resolve to errorMessages
	if e != nil {
		if e.Code != EcodeInternal {
			t.Error("error code not set")
		}
		if e.Message != commonErrorMessages[EcodeInternal] {
			t.Error("error message not set")
		}
	}

	e = NewError(errorMessages, EcodeUnsupportedMediaType) // should resolve to errorMessages
	if e != nil {
		if e.Code != EcodeUnsupportedMediaType {
			t.Error("error code not set")
		}
		if e.Message != errorMessages[EcodeUnsupportedMediaType] {
			t.Error("error message not set")
		}
	}

	// Test failsafe
	e = NewError(errorMessages, EcodeServiceBase+1) // should resolve to commonErrorMessages
	if e != nil {
		if e.Code != EcodeServiceBase+1 {
			t.Error("error code not set")
		}
		if e.Message != commonErrorMessages[EcodeUnknown] {
			t.Error("error message not set")
		}
	}
}
