package datastores

import "testing"

func TestParseDynamoParams(t *testing.T) {
	if _, err := ParseDynamoParams(""); err == nil {
		t.Error("expected error parsing empty string")
	}

	if _, err := ParseDynamoParams("a:b"); err == nil {
		t.Error("expected error parsing short string")
	}

	if _, err := ParseDynamoParams("a:b:c"); err == nil {
		t.Error("expected error parsing short string")
	}

	if _, err := ParseDynamoParams("a:b:c:d:e"); err == nil {
		t.Error("expected error parsing long string")
	}

	p, err := ParseDynamoParams("a:b:c:d")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if p.Region != "a" {
		t.Error("failed to parse region")
	}
	if p.TableName != "b" {
		t.Error("failed to parse table_name")
	}
	if p.AccessKey != "c" {
		t.Error("failed to parse access_key")
	}
	if p.SecretKey != "d" {
		t.Error("failed to parse secret_key")
	}
}
