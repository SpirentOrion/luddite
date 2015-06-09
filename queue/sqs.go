package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite/stats"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"
)

const (
	statSqsSendSuffix  = ".send"
	statSqsErrorSuffix = ".error."
)

// Verify that SqsQueue implements Queue.
var _ Queue = &SqsQueue{}

// SqsParams holds properties for SQS-based queues.
type SqsParams struct {
	Region    string
	QueueName string
	AccessKey string
	SecretKey string
}

// NewSqsParams extracts SQS provider parameters from a
// generic string map and returns a SqsParams structure.
func NewSqsParams(params map[string]string) (*SqsParams, error) {
	p := &SqsParams{
		Region:    params["region"],
		QueueName: params["queue_name"],
		AccessKey: params["access_key"],
		SecretKey: params["secret_key"],
	}

	if p.Region == "" {
		return nil, errors.New("SQS providers require a 'region' parameter")
	}
	if p.QueueName == "" {
		return nil, errors.New("SQS providers require a 'queue_name' parameter")
	}

	return p, nil
}

type SqsQueue struct {
	logger      *log.Entry
	stats       stats.Stats
	statsPrefix string
	*sqs.Queue
}

func NewSqsQueue(params *SqsParams, logger *log.Entry, stats stats.Stats) (*SqsQueue, error) {
	auth, err := aws.GetAuth(params.AccessKey, params.SecretKey, "", time.Time{})
	if err != nil {
		return nil, err
	}
	sqs := sqs.New(auth, aws.Regions[params.Region])
	queue, err := sqs.GetQueue(params.QueueName)
	if err != nil {
		return nil, err
	}
	return &SqsQueue{
		logger:      logger,
		stats:       stats,
		statsPrefix: fmt.Sprintf("queue.%s.%s.%s.", SQS_PROVIDER, params.Region, params.QueueName),
		Queue:       queue,
	}, nil
}

func (q *SqsQueue) Send(msg *Message) error {
	const op = "SendMessage"

	buf, err := json.Marshal(msg)
	if err != nil {
		q.logger.WithFields(log.Fields{
			"provider":   SQS_PROVIDER,
			"region":     q.Queue.Region,
			"queue_name": q.Queue.Name,
			"op":         op,
			"error":      err,
		}).Error()
		q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
		return err
	}

	resp, err := q.SendMessage(string(buf))
	if err != nil {
		q.logger.WithFields(log.Fields{
			"provider":   SQS_PROVIDER,
			"region":     q.Queue.Region,
			"queue_name": q.Queue.Name,
			"op":         op,
			"error":      err,
		}).Error()
		q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
		return err
	}

	q.logger.WithFields(log.Fields{
		"provider":   SQS_PROVIDER,
		"region":     q.Queue.Region,
		"queue_name": q.Queue.Name,
		"op":         op,
		"message_id": resp.Id,
	}).Info()
	q.stats.Incr(q.statsPrefix+statSqsSendSuffix, 1)
	return nil
}
