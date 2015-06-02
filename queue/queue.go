package queue

type Message map[string]interface{}

type Queue interface {
	Send(msg *Message) error
}
