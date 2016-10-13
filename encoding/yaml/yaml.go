// Package yaml provides the YAML encoding for go-structconf
package yaml

import (
	"github.com/anexia-it/go-structconf/encoding"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v2"
)

var _ encoding.Encoding = (*yamlEncoding)(nil)

// Option defines the function type YAML encoding options use
type Option func(*yamlEncoding) error

type yamlEncoding struct {
}

func (e *yamlEncoding) UnmarshalTo(in []byte, dest map[string]interface{}) error {
	return yaml.Unmarshal(in, dest)
}

func (e *yamlEncoding) MarshalFrom(src map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(src)
}

// NewYAMLEncoding returns a new YAML encoding instance
func NewYAMLEncoding(options ...Option) (encoding.Encoding, error) {
	enc := &yamlEncoding{}

	var err error
	for _, opt := range options {
		if optErr := opt(enc); optErr != nil {
			err = multierror.Append(err, optErr)
		}
	}

	if err != nil {
		return nil, err
	}

	return enc, nil
}
