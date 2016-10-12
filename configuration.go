package structconf

import (
	"reflect"

	"github.com/hashicorp/go-multierror"
)

// Configuration represents a configuration backed by a struct
type Configuration struct {
	config     interface{}
	configType reflect.Type

	storage  Storage
	encoding Encoding
}

func (c *Configuration) SetDefaults(defaults interface{}) error {
	// Check if a marshaller was configured
	if c.encoding == nil {
		return ErrEncodingNotConfigured
	}

	// Defaults must be set to a non-nil value
	if defaults == nil {
		return ErrConfigStructIsNil
	}

	defaultsValue := reflect.ValueOf(defaults)

	// Traverse the pointer, if defaults is a pointer
	if defaultsValue.Kind() == reflect.Ptr {
		defaultsValue = defaultsValue.Elem()
	}

	if defaultsValue.Type() != c.configType {
		return ErrConfigStructTypeMismatch
	}

	return nil
}

// NewConfiguration initializes a new configuration with the given options
func NewConfiguration(config interface{}, options ...Option) (*Configuration, error) {
	if config == nil {
		return nil, ErrConfigStructIsNil
	}

	configValue := reflect.ValueOf(config)
	if configValue.Kind() != reflect.Ptr {
		return nil, ErrNotAStructPointer
	} else if configValue.Elem().Kind() != reflect.Struct {
		return nil, ErrNotAStructPointer
	}

	vs := &Configuration{
		config:     config,
		configType: configValue.Elem().Type(),
	}

	var err error

	// Apply options
	for _, opt := range options {
		if optErr := opt(vs); optErr != nil {
			err = multierror.Append(err, optErr)
		}
	}

	// If setting any option caused an error, return it
	if err != nil {
		return nil, err
	}

	return vs, nil
}
