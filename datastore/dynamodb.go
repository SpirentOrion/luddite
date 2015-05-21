package datastore

import (
	"errors"
	"fmt"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite/stats"
	"github.com/SpirentOrion/trace"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/dynamodb"
)

const (
	statDynamoGetItemSuffix             = ".get_item"
	statDynamoGetItemLatencySuffix      = ".get_item_latency"
	statDynamoPutItemSuffix             = ".put_item"
	statDynamoPutItemLatencySuffix      = ".put_item_latency"
	statDynamoUpdateItemSuffix          = ".update_item"
	statDynamoUpdateItemLatencySuffix   = ".update_item_latency"
	statDynamoDeleteItemSuffix          = ".delete_item"
	statDynamoDeleteItemLatencySuffix   = ".delete_item_latency"
	statDynamoScanSuffix                = ".scan"
	statDynamoScanLatencySuffix         = ".scan_latency"
	statDynamoQueryOnIndexSuffix        = ".query_on_index"
	statDynamoQueryOnIndexLatencySuffix = ".query_on_index_latency"
	statDynamoErrorSuffix               = ".error."
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
	logger      *log.Entry
	stats       stats.Stats
	statsPrefix string
	*dynamodb.Table
}

func NewDynamoTable(params *DynamoParams, logger *log.Entry, stats stats.Stats) (*DynamoTable, error) {
	auth, err := aws.GetAuth(params.AccessKey, params.SecretKey, "", time.Time{})
	if err != nil {
		return nil, err
	}
	server := &dynamodb.Server{
		Auth:   auth,
		Region: aws.Regions[params.Region],
	}
	table := server.NewTable(params.TableName, dynamodb.PrimaryKey{KeyAttribute: dynamodb.NewStringAttribute("id", "")})
	return &DynamoTable{
		logger:      logger,
		stats:       stats,
		statsPrefix: fmt.Sprintf("datastore.%s.%s.%s.", DYNAMODB_PROVIDER, params.Region, params.TableName),
		Table:       table,
	}, nil
}

func (t *DynamoTable) String() string {
	return fmt.Sprintf("%s{%s:%s}", DYNAMODB_PROVIDER, t.Server.Region.Name, t.Name)
}

func (t *DynamoTable) GetItem(id string) (attrs map[string]*dynamodb.Attribute, ok bool, err error) {
	var latency time.Duration
	const op = "GetItem"

	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		key := &dynamodb.Key{HashKey: id}
		start := time.Now()
		attrs, err = t.Table.GetItem(key)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			if err != nil && err != dynamodb.ErrNotFound {
				data["error"] = err
			}
		}
	})

	t.stats.Incr(t.statsPrefix+statDynamoGetItemSuffix, 1)
	t.stats.PrecisionTiming(t.statsPrefix+statDynamoGetItemLatencySuffix, latency)
	if err != nil {
		if err == dynamodb.ErrNotFound {
			err = nil
		} else {
			t.handleError(op, err)
		}
		return
	} else {
		ok = true
	}
	return
}

func (t *DynamoTable) PutItem(id string, attrs []dynamodb.Attribute, condAttrs []dynamodb.Attribute) (err error) {
	var (
		latency time.Duration
		op      string
	)

	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		if len(condAttrs) != 0 {
			op = "ConditionalPutItem"
			start := time.Now()
			_, err = t.Table.ConditionalPutItem(id, "", attrs, condAttrs)
			latency = time.Since(start)
			if s != nil {
				data := s.Data()
				data["op"] = op
				if err != nil {
					data["error"] = err
				}
			}
		} else {
			op = "PutItem"
			start := time.Now()
			_, err = t.Table.PutItem(id, "", attrs)
			latency = time.Since(start)
			if s != nil {
				data := s.Data()
				data["op"] = op
				if err != nil {
					data["error"] = err
				}
			}
		}
	})

	t.stats.Incr(t.statsPrefix+statDynamoPutItemSuffix, 1)
	t.stats.PrecisionTiming(t.statsPrefix+statDynamoPutItemLatencySuffix, latency)
	if err != nil {
		t.handleError(op, err)
	}
	return
}

func (t *DynamoTable) UpdateItem(id string, attrs []dynamodb.Attribute, condAttrs []dynamodb.Attribute) (err error) {
	var (
		latency time.Duration
		op      string
	)

	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		key := &dynamodb.Key{HashKey: id}
		if len(condAttrs) != 0 {
			op = "ConditionalUpdateAttributes"
			start := time.Now()
			_, err = t.Table.ConditionalUpdateAttributes(key, attrs, condAttrs)
			latency = time.Since(start)
			if s != nil {
				data := s.Data()
				data["op"] = op
				if err != nil {
					data["error"] = err
				}
			}
		} else {
			op = "UpdateAttributes"
			start := time.Now()
			_, err = t.Table.UpdateAttributes(key, attrs)
			latency = time.Since(start)
			if s != nil {
				data := s.Data()
				data["op"] = op
				if err != nil {
					data["error"] = err
				}
			}
		}
	})

	t.stats.Incr(t.statsPrefix+statDynamoUpdateItemSuffix, 1)
	t.stats.PrecisionTiming(t.statsPrefix+statDynamoUpdateItemLatencySuffix, latency)
	if err != nil {
		t.handleError(op, err)
	}
	return
}

func (t *DynamoTable) DeleteItem(id string) (ok bool, err error) {
	var latency time.Duration
	const op = "DeleteItem"

	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		key := &dynamodb.Key{HashKey: id}
		start := time.Now()
		_, err = t.Table.DeleteItem(key)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			if err != nil && err != dynamodb.ErrNotFound {
				data["error"] = err
			}
		}
	})

	t.stats.Incr(t.statsPrefix+statDynamoDeleteItemSuffix, 1)
	t.stats.PrecisionTiming(t.statsPrefix+statDynamoDeleteItemLatencySuffix, latency)
	if err != nil {
		if err == dynamodb.ErrNotFound {
			err = nil
		} else {
			t.handleError(op, err)
		}
		return
	} else {
		ok = true
	}
	return
}

func (t *DynamoTable) Scan(comps []dynamodb.AttributeComparison) (attrs []map[string]*dynamodb.Attribute, err error) {
	var latency time.Duration
	const op = "Scan"

	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		start := time.Now()
		attrs, err = t.Table.Scan(comps)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			if err != nil {
				data["error"] = err
			} else {
				data["items"] = len(attrs)
			}
		}
	})

	t.stats.Incr(t.statsPrefix+statDynamoScanSuffix, 1)
	t.stats.PrecisionTiming(t.statsPrefix+statDynamoScanSuffix, latency)
	if err != nil {
		t.handleError(op, err)
		attrs = nil
		return
	}
	return
}

func (t *DynamoTable) QueryOnIndex(comps []dynamodb.AttributeComparison, indexName string) (attrs []map[string]*dynamodb.Attribute, err error) {
	var latency time.Duration
	const op = "QueryOnIndex"

	s, _ := trace.Continue(DYNAMODB_PROVIDER, t.String())
	trace.Run(s, func() {
		start := time.Now()
		attrs, err = t.Table.QueryOnIndex(comps, indexName)
		latency = time.Since(start)
		if s != nil {
			data := s.Data()
			data["op"] = op
			data["index"] = indexName
			if err != nil {
				data["error"] = err
			} else {
				data["items"] = len(attrs)
			}
		}
	})

	t.stats.Incr(t.statsPrefix+statDynamoQueryOnIndexSuffix, 1)
	t.stats.PrecisionTiming(t.statsPrefix+statDynamoQueryOnIndexSuffix, latency)
	if err != nil {
		t.handleError(op, err)
		return
	}
	return
}

func (t *DynamoTable) handleError(op string, err error) {
	t.logger.WithFields(log.Fields{
		"provider":   DYNAMODB_PROVIDER,
		"region":     t.Server.Region,
		"table_name": t.Name,
		"op":         op,
		"error":      err,
	}).Error()

	dynErr, ok := err.(*dynamodb.Error)
	if ok {
		t.stats.Incr(t.statsPrefix+statDynamoErrorSuffix+dynErr.Code, 1)
	} else {
		t.stats.Incr(t.statsPrefix+statDynamoErrorSuffix+"other", 1)
	}
}
