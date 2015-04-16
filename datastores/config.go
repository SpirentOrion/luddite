package datastores

const (
	DYNAMODB_PROVIDER = "dynamodb"
)

// Config holds per-datastore configuration properties.
type Config struct {
	Provider string
	Params   string
}
