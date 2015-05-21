package datastore

import "testing"

func TestNewPostgresParams(t *testing.T) {
	params := map[string]string{
		"user":     "a",
		"password": "b",
		"db_name":  "c",
		"host":     "d",
		"port":     "1234",
	}

	p, err := NewPostgresParams(params)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if p.User != "a" {
		t.Error("failed to parse user")
	}
	if p.Password != "b" {
		t.Error("failed to parse password")
	}
	if p.DbName != "c" {
		t.Error("failed to parse db_name")
	}
	if p.Host != "d" {
		t.Error("failed to parse host")
	}
	if p.Port != 1234 {
		t.Error("failed to parse port")
	}

	delete(params, "port")
	p, err = NewPostgresParams(params)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if p.Port != 5432 {
		t.Error("expected default port")
	}

	params = map[string]string{}
	_, err = NewDynamoParams(params)
	if err == nil {
		t.Error("expected error for missing parameters")
	}
}
