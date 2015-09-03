package datastore

import (
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var validGlobalId = regexp.MustCompile(`^([a-f0-9]{32})$`)

func NewGlobalId() string {
	return strings.ToLower(strings.Replace(uuid.NewV4().String(), "-", "", -1))
}

func IsValidGlobalId(id string) bool {
	return validGlobalId.MatchString(id)
}
