package datastore

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func NewGlobalId() string {
	return strings.ToLower(strings.Replace(uuid.NewV4().String(), "-", "", -1))
}
