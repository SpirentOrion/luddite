package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite/stats"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/sqs"
)

const (
	statSqsSendSuffix    = ".send"
	statSqsReceiveSuffix = ".receive"
	statSqsDeleteSuffix  = ".delete"
	statSqsErrorSuffix   = ".error"
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
	queue       *sqs.Queue
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
		queue:       queue,
	}, nil
}

func (q *SqsQueue) String() string {
	return q.queue.Name
}

func (q *SqsQueue) Send(msg *Message) error {
	const op = "SendMessage"

	buf, err := json.Marshal(msg)
	if err != nil {
		q.logger.WithFields(log.Fields{
			"provider":   SQS_PROVIDER,
			"region":     q.queue.Region,
			"queue_name": q.queue.Name,
			"op":         op,
			"error":      err,
		}).Error(err.Error())
		q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
		return err
	}

	resp, err := q.queue.SendMessage(string(buf))
	if err != nil {
		q.logger.WithFields(log.Fields{
			"provider":   SQS_PROVIDER,
			"region":     q.queue.Region,
			"queue_name": q.queue.Name,
			"op":         op,
			"error":      err,
		}).Error(err.Error())
		q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
		return err
	}

	q.logger.WithFields(log.Fields{
		"provider":   SQS_PROVIDER,
		"region":     q.queue.Region,
		"queue_name": q.queue.Name,
		"op":         op,
		"message_id": resp.Id,
	}).Info()
	q.stats.Incr(q.statsPrefix+statSqsSendSuffix, 1)
	return nil
}

// SqsMessage holds Messages and extra properties for messages
// received from an SQS queue.
type SqsMessage struct {
	MessageId     string
	ReceiptHandle string
	Message
}

func (q *SqsQueue) Receive(maxMessages int, waitTime time.Duration) ([]SqsMessage, error) {
	const op = "ReceiveMessage"

	params := map[string]string{
		"MaxNumberOfMessages": strconv.Itoa(maxMessages),
		"WaitTimeSeconds":     strconv.Itoa(int(waitTime.Seconds())),
	}

	resp, err := q.queue.ReceiveMessageWithParameters(params)
	if err != nil {
		q.logger.WithFields(log.Fields{
			"provider":   SQS_PROVIDER,
			"region":     q.queue.Region,
			"queue_name": q.queue.Name,
			"op":         op,
			"error":      err,
		}).Error(err.Error())
		q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
		return nil, err
	}
	if len(resp.Messages) == 0 {
		return nil, nil
	}

	msgs := make([]SqsMessage, 0, len(resp.Messages))
	for _, respMsg := range resp.Messages {
		msg := new(SqsMessage)
		if err := json.Unmarshal([]byte(respMsg.Body), &msg.Message); err != nil {
			q.logger.WithFields(log.Fields{
				"provider":   SQS_PROVIDER,
				"region":     q.queue.Region,
				"queue_name": q.queue.Name,
				"op":         op,
				"message_id": respMsg.MessageId,
				"error":      err,
			}).Error(err.Error())
			q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
			defer q.queue.DeleteMessageUsingReceiptHandle(respMsg.ReceiptHandle)
		}

		msg.MessageId = respMsg.MessageId
		msg.ReceiptHandle = respMsg.ReceiptHandle
		msgs = append(msgs, *msg)

		q.logger.WithFields(log.Fields{
			"provider":       SQS_PROVIDER,
			"region":         q.queue.Region,
			"queue_name":     q.queue.Name,
			"op":             op,
			"message_id":     msg.MessageId,
			"receipt_handle": msg.ReceiptHandle,
		}).Info()
		q.stats.Incr(q.statsPrefix+statSqsReceiveSuffix, 1)
	}

	return msgs, nil
}

func (q *SqsQueue) Delete(receiptHandle string) error {
	const op = "DeleteMessage"

	if _, err := q.queue.DeleteMessageUsingReceiptHandle(receiptHandle); err != nil {
		q.logger.WithFields(log.Fields{
			"provider":   SQS_PROVIDER,
			"region":     q.queue.Region,
			"queue_name": q.queue.Name,
			"op":         op,
			"error":      err,
		}).Error()
		q.stats.Incr(q.statsPrefix+statSqsErrorSuffix, 1)
		return err
	}

	q.logger.WithFields(log.Fields{
		"provider":       SQS_PROVIDER,
		"region":         q.queue.Region,
		"queue_name":     q.queue.Name,
		"op":             op,
		"receipt_handle": receiptHandle,
	}).Info()
	q.stats.Incr(q.statsPrefix+statSqsDeleteSuffix, 1)
	return nil
}
