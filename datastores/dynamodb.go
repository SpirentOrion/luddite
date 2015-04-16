package datastores

import (
	"errors"
	"strings"
)

// DynamoParams holds connection and auth properties for
// DynamoDB-based datastores.
type DynamoParams struct {
	Region    string
	TableName string
	AccessKey string
	SecretKey string
}

// ParseDynamoParams parses a parameter string for the DynamoDB
// datastore provider and returns a DynamoParams structure.
func ParseDynamoParams(s string) (*DynamoParams, error) {
	params := strings.Split(s, ":")
	if len(params) != 4 {
		return nil, errors.New("DynamoDB provider params have 4 parameters (region:table_name:access_key:secret_key)")
	}
	return &DynamoParams{
		Region:    params[0],
		TableName: params[1],
		AccessKey: params[2],
		SecretKey: params[3],
	}, nil
}
