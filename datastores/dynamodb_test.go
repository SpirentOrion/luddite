package datastores

import "testing"

func TestGetDynamoParams(t *testing.T) {
	params := map[string]string{
		"region":     "a",
		"table_name": "b",
		"access_key": "c",
		"secret_key": "d",
	}

	p, err := GetDynamoParams(params)
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

	delete(params, "access_key")
	delete(params, "secret_key")
	_, err = GetDynamoParams(params)
	if err != nil {
		t.Error("unexpected error for missing access_key and secret_key: %s", err)
	}

	delete(params, "region")
	delete(params, "table_name")
	_, err = GetDynamoParams(params)
	if err == nil {
		t.Error("expected error for missing region and table_name")
	}
}
