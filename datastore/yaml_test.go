package datastore

import "testing"

func TestNewYAMLParams(t *testing.T) {
	params := map[string]string{"path": "a"}

	p, err := NewYAMLParams(params)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if p.Path != "a" {
		t.Error("failed to parse path")
	}

	delete(params, "path")

	_, err = NewYAMLParams(params)
	if err == nil {
		t.Error("expected error for missing path")
	}
}
