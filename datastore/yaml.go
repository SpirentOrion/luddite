package datastore

import "errors"

// YAMLParams holds path properties for YAML-based datastores.
type YAMLParams struct {
	Path string
}

// NewYAMLParams extracts YAML provider parameters from a
// generic string map and returns a YAMLParams structure.
func NewYAMLParams(params map[string]string) (*YAMLParams, error) {
	p := &YAMLParams{params["path"]}
	if p.Path == "" {
		return nil, errors.New("YAML providers require a 'path' parameter")
	}

	return p, nil
}
