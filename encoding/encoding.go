// Package encoding provides common functionality for go-structconf encodings
package encoding

// Encoding defines the configuration encoding interface
// An implementation of this interface provides marshalling and unmarshalling
// of the configuration data
type Encoding interface {
	// UnmarshalTo unmarshals the passed bytes to the given destination
	UnmarshalTo(in []byte, dest map[string]interface{}) error

	// MarshalFrom marshals the given source to an array of bytes
	MarshalFrom(src map[string]interface{}) ([]byte, error)
}
