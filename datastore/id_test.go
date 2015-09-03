package datastore

import (
	"strings"
	"testing"
)

func TestNewGlobalId(t *testing.T) {
	id := NewGlobalId()
	if len(id) != 32 || strings.Map(stripHexDigits, id) != "" {
		t.Errorf("expected GUID-like value, got: %s", id)
	}
}

func TestIsValidGlobalId(t *testing.T) {
	if !IsValidGlobalId(NewGlobalId()) {
		t.Error("expected valid id")
	}

	if IsValidGlobalId("") {
		t.Error("expected invalid id")
	}
	if IsValidGlobalId("foo") {
		t.Error("expected invalid id")
	}
	if IsValidGlobalId(strings.ToUpper(NewGlobalId())) {
		t.Error("expected invalid id")
	}
	if IsValidGlobalId(NewGlobalId() + "0") {
		t.Error("expected invalid id")
	}
}

func stripHexDigits(r rune) rune {
	switch {
	case r >= 'a' && r <= 'f':
		return -1
	case r >= '0' && r <= '9':
		return -1
	}
	return r
}
