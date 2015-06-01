package queue

import "testing"

func TestNewSqsParams(t *testing.T) {
	params := map[string]string{
		"region":     "a",
		"queue_name": "b",
		"access_key": "c",
		"secret_key": "d",
	}

	p, err := NewSqsParams(params)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if p.Region != "a" {
		t.Error("failed to parse region")
	}
	if p.QueueName != "b" {
		t.Error("failed to parse queue_name")
	}
	if p.AccessKey != "c" {
		t.Error("failed to parse access_key")
	}
	if p.SecretKey != "d" {
		t.Error("failed to parse secret_key")
	}

	delete(params, "access_key")
	delete(params, "secret_key")
	_, err = NewSqsParams(params)
	if err != nil {
		t.Error("unexpected error for missing access_key and secret_key: %s", err)
	}

	delete(params, "region")
	delete(params, "queue_name")
	_, err = NewSqsParams(params)
	if err == nil {
		t.Error("expected error for missing region and queue_name")
	}
}
