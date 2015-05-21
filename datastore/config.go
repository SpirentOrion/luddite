package datastore

const (
	DYNAMODB_PROVIDER = "dynamodb"
	POSTGRES_PROVIDER = "postgres"
	YAML_PROVIDER     = "yaml"
)

// Config holds per-datastore configuration properties.
type Config struct {
	Provider string
	Params   map[string]string
}
