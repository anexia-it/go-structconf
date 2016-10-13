package structconf

import (
	"github.com/anexia-it/go-structconf/encoding"
	"github.com/anexia-it/go-structconf/storage"
)

// Option defines the function type of Configuration options
type Option func(*Configuration) error

// OptionStorage configures a configuration storage
func OptionStorage(storage storage.Storage) Option {
	return func(c *Configuration) error {
		c.storage = storage
		return nil
	}
}

// OptionEncoding configures a configuration encoding
func OptionEncoding(encoding encoding.Encoding) Option {
	return func(c *Configuration) error {
		c.encoding = encoding
		return nil
	}
}

// OptionTagName configures the tag names used when encoding the config struct
func OptionTagName(tagName string) Option {
	return func(c *Configuration) error {
		c.tagName = tagName
		return nil
	}
}

// OptionDefaults configures the default values from a struct
// This requires an encoding to be configured before-hand and will
// return an error if no encoding was configured
func OptionDefaults(defaults interface{}) Option {
	return func(c *Configuration) error {
		c.pendingDefaults = defaults
		return nil
	}
}
