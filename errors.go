package structconf

import "errors"

var (
	// ErrConfigStructIsNil indicates that the passed config struct is nil
	ErrConfigStructIsNil = errors.New("Config struct is nil")

	// ErrNotAStructPointer indicates that the passed value is not a pointer to a struct
	ErrNotAStructPointer = errors.New("Passed value is not a struct pointer")
	// ErrNotAStruct indicates that the passed value is not a struct
	ErrNotAStruct = errors.New("Passed value is not a struct")
	// ErrMapIsNil indicates that the passed map is nil
	ErrMapIsNil = errors.New("Map is nil")
	// ErrFieldIsInterface indicates that the field is an interface
	ErrFieldIsInterface = errors.New("Field is an interface")
	// ErrInvalidMap indicates that an invalid map was passed to a function
	ErrInvalidMap = errors.New("Invalid map")

	// ErrConfigStructTypeMismatch indicates that the passed config struct and
	// default values are not of the same type
	ErrConfigStructTypeMismatch = errors.New("Type mismatch between defaults and config struct")

	// ErrEncodingNotConfigured indicates that no encoding was configured
	ErrEncodingNotConfigured = errors.New("Encoding not configured")

	// ErrTagNameEmpty indicates that the supplied tag name is empty
	ErrTagNameEmpty = errors.New("Tag name is empty")
)
