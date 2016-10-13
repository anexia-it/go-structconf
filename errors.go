package structconf

import "errors"

var (
	// ErrConfigStructIsNil indicates that the passed config struct is nil
	ErrConfigStructIsNil = errors.New("Config struct is nil")

	// ErrNotAStructPointer indicates that the passed value is not a pointer to a struct
	ErrNotAStructPointer = errors.New("Passed value is not a struct pointer")

	// ErrConfigStructTypeMismatch indicates that the passed config struct and
	// default values are not of the same type
	ErrConfigStructTypeMismatch = errors.New("Type mismatch between defaults and config struct")

	// ErrEncodingNotConfigured indicates that no encoding was configured
	ErrEncodingNotConfigured = errors.New("Encoding not configured")

	// ErrStorageNotConfigured indicates that no storage was configured
	ErrStorageNotConfigured = errors.New("Storage not configured")
)
