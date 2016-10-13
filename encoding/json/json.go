// Package json provides the JSON encoding for go-structconf
package json

import (
	"bytes"
	"encoding/json"

	"github.com/anexia-it/go-structconf/encoding"
	"github.com/hashicorp/go-multierror"
)

var _ encoding.Encoding = (*jsonEncoding)(nil)

// Option defines the function type JSON encoding options use
type Option func(*jsonEncoding) error

type jsonEncoding struct {
}

func (e *jsonEncoding) UnmarshalTo(in []byte, dest map[string]interface{}) error {
	buf := bytes.NewBuffer(in)
	dec := json.NewDecoder(buf)
	return dec.Decode(&dest)
}

func (e *jsonEncoding) MarshalFrom(src map[string]interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)

	if err := enc.Encode(src); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// NewJSONEncoding returns a new JSON encoding instance
func NewJSONEncoding(options ...Option) (encoding.Encoding, error) {
	enc := &jsonEncoding{}

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
