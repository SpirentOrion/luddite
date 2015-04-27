package datastore

import "testing"

func TestGetYAMLParams(t *testing.T) {
	params := map[string]string{"path": "a"}

	p, err := GetYAMLParams(params)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if p.Path != "a" {
		t.Error("failed to parse path")
	}

	delete(params, "path")

	_, err = GetYAMLParams(params)
	if err == nil {
		t.Error("expected error for missing path")
	}
}
