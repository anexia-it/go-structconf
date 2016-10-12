package structconf

// An option configures a Configuration option
type Option func(*Configuration) error

// OptionStorage configures a configuration storage
func OptionStorage(storage Storage) Option {
	return func(c *Configuration) error {
		c.storage = storage
		return nil
	}
}

// OptionEncoding configures a configuration encoding
func OptionEncoding(encoding Encoding) Option {
	return func(c *Configuration) error {
		c.encoding = encoding
		return nil
	}
}

// OptionDefaults configures the default values from a struct
// This requires an encoding to be configured before-hand and will
// return an error if no encoding was configured
func OptionDefaults(defaults interface{}) Option {
	return func(c *Configuration) error {
		if c.encoding == nil {
			return ErrEncodingNotConfigured
		}
		return c.SetDefaults(defaults)
	}
}
