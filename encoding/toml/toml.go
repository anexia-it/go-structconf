// Package toml provides the TOML encoding for go-structconf
package toml

import (
	"bytes"

	"github.com/BurntSushi/toml"
	"github.com/anexia-it/go-structconf/encoding"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/anexia-it/go-structmapper.v1"
)

var _ encoding.Encoding = (*tomlEncoding)(nil)

// Option defines the function type TOML encoding options use
type Option func(*tomlEncoding) error

type tomlEncoding struct {
}

func (e *tomlEncoding) UnmarshalTo(in []byte, dest map[string]interface{}) error {
	_, err := toml.Decode(string(in), &dest)
	return err
}

func (e *tomlEncoding) MarshalFrom(src map[string]interface{}) ([]byte, error) {
	buf := bytes.NewBufferString("")
	enc := toml.NewEncoder(buf)
	var err error
	src, err = structmapper.ForceStringMapKeys(src)
	if err != nil {
		return nil, err
	}
	if err := enc.Encode(src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NewTOMLEncoding returns a new TOML encoding instance
func NewTOMLEncoding(options ...Option) (encoding.Encoding, error) {
	enc := &tomlEncoding{}

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
