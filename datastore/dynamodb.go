package datastore

import (
	"errors"
	"fmt"
	"time"

	"github.com/SpirentOrion/trace"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/dynamodb"
)

// DynamoParams holds AWS connection and auth properties for
// DynamoDB-based datastores.
type DynamoParams struct {
	Region    string
	TableName string
	AccessKey string
	SecretKey string
}

// NewDynamoParams extracts DynamoDB provider parameters from a
// generic string map and returns a DynamoParams structure.
func NewDynamoParams(params map[string]string) (*DynamoParams, error) {
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

type DynamoTable struct {
	*dynamodb.Table
}

func NewDynamoTable(params *DynamoParams) (*DynamoTable, error) {
	auth, err := aws.GetAuth(params.AccessKey, params.SecretKey, "", time.Time{})
	if err != nil {
		return nil, err
	}
	server := &dynamodb.Server{
		Auth:   auth,
		Region: aws.Regions[params.Region],
	}
	table := server.NewTable(params.TableName, dynamodb.PrimaryKey{KeyAttribute: dynamodb.NewStringAttribute("id", "")})
	return &DynamoTable{table}, nil
}

func (t *DynamoTable) String() string {
	return fmt.Sprintf("%s{%s:%s}", DYNAMODB_PROVIDER, t.Server.Region.Name, t.Name)
}

func (t *DynamoTable) Scan() (attrs []map[string]*dynamodb.Attribute, err error) {
	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		attrs, err = t.Table.Scan([]dynamodb.AttributeComparison{})
		if s != nil {
			data := s.Data()
			data["op"] = "Scan"
			if err != nil {
				data["error"] = err
			} else {
				data["items"] = len(attrs)
			}
		}
	})

	if err != nil {
		attrs = nil
		return
	}
	return
}

func (t *DynamoTable) GetItem(id string) (attrs map[string]*dynamodb.Attribute, ok bool, err error) {
	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		key := &dynamodb.Key{HashKey: id}
		attrs, err = t.Table.GetItem(key)
		if s != nil {
			data := s.Data()
			data["op"] = "GetItem"
			if err != nil && err != dynamodb.ErrNotFound {
				data["error"] = err
			}
		}
	})

	if err != nil {
		if err == dynamodb.ErrNotFound {
			err = nil
		}
		return
	} else {
		ok = true
	}
	return
}

func (t *DynamoTable) PutItem(id string, attrs []dynamodb.Attribute) (err error) {
	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		_, err = t.Table.PutItem(id, "", attrs)
		if s != nil {
			data := s.Data()
			data["op"] = "PutItem"
			if err != nil {
				data["error"] = err
			}
		}
	})

	return
}

func (t *DynamoTable) UpdateItem(id string, serial int64, attrs []dynamodb.Attribute) (err error) {
	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		key := &dynamodb.Key{HashKey: id}
		serialAttr := []dynamodb.Attribute{{
			Type:  dynamodb.TYPE_NUMBER,
			Name:  "serial",
			Value: fmt.Sprint(serial),
		}}
		_, err = t.Table.ConditionalUpdateAttributes(key, attrs, serialAttr)
		if s != nil {
			data := s.Data()
			data["op"] = "ConditionalUpdateAttributes"
			if err != nil {
				data["error"] = err
			}
		}
	})

	return
}

func (t *DynamoTable) DeleteItem(id string) (ok bool, err error) {
	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		key := &dynamodb.Key{HashKey: id}
		_, err = t.Table.DeleteItem(key)
		if s != nil {
			data := s.Data()
			data["op"] = "DeleteItem"
			if err != nil && err != dynamodb.ErrNotFound {
				data["error"] = err
			}
		}
	})

	if err != nil {
		if err == dynamodb.ErrNotFound {
			err = nil
		}
		return
	} else {
		ok = true
	}
	return
}
