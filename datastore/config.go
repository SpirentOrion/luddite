package datastore

const (
	DYNAMODB_PROVIDER = "dynamodb"
)

// Config holds per-datastore configuration properties.
type Config struct {
	Provider string
	Params   map[string]string
}
