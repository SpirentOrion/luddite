package queue

// Verify that NullQueue implements Queue.
var _ Queue = &NullQueue{}

type NullQueue struct{}

func (q *NullQueue) Send(msg *Message) error {
	return nil
}
