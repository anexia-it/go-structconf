package structconf

import (
	"reflect"

	"sync"

	"github.com/anexia-it/go-structconf/encoding"
	"github.com/anexia-it/go-structconf/storage"
	structmapper "github.com/anexia-it/go-structmapper"
	"github.com/hashicorp/go-multierror"
)

// Default configuration options
var defaultOptions = []Option{
	OptionTagName("config"),
}

// Configuration represents a configuration backed by a struct
type Configuration struct {
	tagName string

	config          interface{}
	pendingDefaults interface{}
	configType      reflect.Type

	storage  storage.Storage
	encoding encoding.Encoding

	mapper *structmapper.Mapper
}

func (c *Configuration) mergeAndSet(a, b map[string]interface{}) error {
	// Merge both maps...
	mergedMap, err := MergeMaps(a, b)
	if err != nil {
		return err
	}

	// If the configuration implements the sync.Locker interface, use it.
	// This allows configuration structs to ensure no race-conditions are created
	// by a write during a read.
	locker, ok := c.config.(sync.Locker)
	if ok {
		locker.Lock()
		// It is safe to defer this call inside the if block here, as it will only be executed
		// when leaving this function, not the if block.
		defer locker.Unlock()
	}
	// Apply the merged result to the config
	err = c.mapper.ToStruct(mergedMap, c.config)
	return err
}

func (c *Configuration) SetDefaults(defaults interface{}) error {
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

	// Convert both the current config, as well as the defaults to a map[string]interface{}
	defaultsMap, err := c.mapper.ToMap(defaults)
	if err != nil {
		return err
	}

	configMap, err := c.mapper.ToMap(c.config)
	if err != nil {
		return err
	}

	return c.mergeAndSet(defaultsMap, configMap)
}

// Load loads the configuration from the underlying storage
func (c *Configuration) Load() error {
	// Check if encoding and storage were configured
	if c.encoding == nil {
		return ErrEncodingNotConfigured
	} else if c.storage == nil {
		return ErrStorageNotConfigured
	}

	buf, err := c.storage.ReadConfig()
	if err != nil {
		// Storage reported error
		return err
	}

	// Decode onto map[string]interface{}
	loadedMap := make(map[string]interface{})
	if err := c.encoding.UnmarshalTo(buf, loadedMap); err != nil {
		// Encoding error
		return err
	}

	// Create a map from the current configuration
	currentMap, err := c.mapper.ToMap(c.config)

	return c.mergeAndSet(currentMap, loadedMap)
}

func (c *Configuration) Save() error {
	// Check if encoding and storage were configured
	if c.encoding == nil {
		return ErrEncodingNotConfigured
	} else if c.storage == nil {
		return ErrStorageNotConfigured
	}

	// Convert the configuration to a map
	configData, err := c.mapper.ToMap(c.config)
	if err != nil {
		return err
	}

	// Encode the configuration using the encoding
	encoded, err := c.encoding.MarshalFrom(configData)
	if err != nil {
		return err
	}

	// Write the configuration to the storage
	if err := c.storage.WriteConfig(encoded); err != nil {
		return err
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

	c := &Configuration{
		config:     config,
		configType: configValue.Elem().Type(),
	}

	var err error

	// Apply default options
	for _, opt := range defaultOptions {
		if optErr := opt(c); optErr != nil {
			panic(optErr)
		}
	}

	// Apply options
	for _, opt := range options {
		if optErr := opt(c); optErr != nil {
			err = multierror.Append(err, optErr)
		}
	}

	// If setting any option caused an error, return it
	if err != nil {
		return nil, err
	}

	// Configure the mapper...
	mapper, err := structmapper.NewMapper(structmapper.OptionTagName(c.tagName))
	if err != nil {
		return nil, err
	}

	c.mapper = mapper

	if c.pendingDefaults != nil {
		// OptionDefaults was used, apply defaults now...
		if err := c.SetDefaults(c.pendingDefaults); err != nil {
			return nil, err
		}
		// Clear pendingDefaults again
		c.pendingDefaults = nil
	}

	return c, nil
}
