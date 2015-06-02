package queue

const (
	SQS_PROVIDER = "sqs"
)

// Config holds per-queue configuration properties.
type Config struct {
	Provider string
	Params   map[string]string
}
