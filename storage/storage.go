package storage

// Storage defines the interface configuration storages implement
type Storage interface {
	// WriteConfig writes the configuration bytes to the storage
	WriteConfig([]byte) error
	// ReadConfig reads the configuration bytes from the storage
	ReadConfig() ([]byte, error)
}
