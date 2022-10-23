package configfile

import (
	"errors"

	"github.com/benchttp/sdk/configparse"
)

type extension string

const (
	extYML  extension = ".yml"
	extYAML extension = ".yaml"
	extJSON extension = ".json"
)

// configParser exposes a method parse to read bytes as a raw config.
type configParser interface {
	// parse parses a raw bytes input as a raw config and stores
	// the resulting value into dst.
	Parse(in []byte, dst *configparse.Representation) error
}

// newParser returns an appropriate parser according to ext, or a non-nil
// error if ext is not an expected extension.
func newParser(ext extension) (configParser, error) {
	switch ext {
	case extYML, extYAML:
		return configparse.YAMLParser{}, nil
	case extJSON:
		return configparse.JSONParser{}, nil
	default:
		return nil, errors.New("unsupported config format")
	}
}
