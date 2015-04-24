package datastores

import "errors"

// DynamoParams holds AWS connection and auth properties for
// DynamoDB-based datastores.
type DynamoParams struct {
	Region    string
	TableName string
	AccessKey string
	SecretKey string
}

// GetDynamoParams extracts DynamoDB provider parameters from a
// generic string map and returns a DynamoParams structure.
func GetDynamoParams(params map[string]string) (*DynamoParams, error) {
	p := &DynamoParams{
		Region:    params["region"],
		TableName: params["table_name"],
		AccessKey: params["access_key"],
		SecretKey: params["secret_key"],
	}

	if p.Region == "" {
		return nil, errors.New("DynamoDB providers require a 'region' parameter")
	}
	if p.TableName == "" {
		return nil, errors.New("DynamoDB providers require a 'table_name' parameter")
	}

	return p, nil
}
