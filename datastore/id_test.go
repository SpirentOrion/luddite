package datastore

import (
	"strings"
	"testing"
)

func TestGlobalId(t *testing.T) {
	id := NewGlobalId()
	if len(id) != 32 || strings.Map(stripHexDigits, id) != "" {
		t.Errorf("expected GUID-like value, got: %s", id)
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
