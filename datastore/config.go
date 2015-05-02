package datastore

const (
	DYNAMODB_PROVIDER = "dynamodb"
	YAML_PROVIDER     = "yaml"
)

// Config holds per-datastore configuration properties.
type Config struct {
	Provider string
	Params   map[string]string
}
