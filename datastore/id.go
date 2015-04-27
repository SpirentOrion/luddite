package datastore

import (
	"strings"

	"code.google.com/p/go-uuid/uuid"
)

func NewGlobalId() string {
	return strings.ToLower(strings.Replace(uuid.New(), "-", "", -1))
}
