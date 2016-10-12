package structconf

import (
	"reflect"

	"fmt"

	"strings"
	"unicode"

	"encoding"

	"errors"

	"github.com/hashicorp/go-multierror"
)

// StructMapperDefaultTagName defines the default tag name used by StructMapper
const StructMapperDefaultTagName = "config"

// An option that configures StructMapper
type StructMapperOption func(*StructMapper) error

func StructMapperOptionTagName(tagName string) StructMapperOption {
	return func(m *StructMapper) error {
		if tagName == "" {
			return ErrTagNameEmpty
		}

		m.tagName = tagName
		return nil
	}
}

// Default options for StructMapper
var structMapperDefaultOptions = []StructMapperOption{
	StructMapperOptionTagName(StructMapperDefaultTagName),
}

var _ error = (*InvalidTag)(nil)

// InvalidTag is an error that indicates that the tag value was invalid
type InvalidTag struct {
	tag string
}

func (it *InvalidTag) Error() string {
	return fmt.Sprintf("Invalid tag: '%s'", it.tag)
}

func (it *InvalidTag) Tag() string {
	return it.tag
}

func newErrorInvalidTag(tag string) error {
	return &InvalidTag{
		tag: tag,
	}
}

// IsInvalidTag checks if the given error is an InvalidTag error
// and returns the InvalidTag error along with a boolean that defines
// if it is indeed an invalid tag error.
func IsInvalidTag(err error) (*InvalidTag, bool) {
	it, ok := err.(*InvalidTag)
	return it, ok
}

func parseTag(tag string) (name string, omitEmpty bool, err error) {
	name = tag

	// Handle the "ignore me" tag value
	if name == "-" {
		return
	}

	if strings.HasSuffix(tag, ",omitempty") {
		omitEmpty = true
		name = strings.TrimSuffix(tag, ",omitempty")
	}

	for _, letter := range name {
		if unicode.IsSymbol(letter) {
			err = newErrorInvalidTag(tag)
			return
		}
	}

	return
}

// StructMapper implements struct to map encoding
type StructMapper struct {
	tagName string
}

func (sm *StructMapper) mapMap(v reflect.Value, prefix ...string) (m map[interface{}]interface{}, err error) {
	keys := v.MapKeys()
	m = make(map[interface{}]interface{}, len(keys))

	for i := 0; i < len(keys); i++ {
		keyV := keys[i]
		keyI := keyV.Interface()
		valueV := v.MapIndex(keyV)

		valueI, mapErr := sm.mapValue(valueV.Interface(), valueV, prefix...)

		if mapErr != nil {
			err = multierror.Append(err, mapErr)
			continue
		}
		m[keyI] = valueI
	}

	return
}

func (sm *StructMapper) mapSlice(v reflect.Value, prefix ...string) (s []interface{}, err error) {
	s = make([]interface{}, 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		valueV := v.Index(i)
		valueI := valueV.Interface()

		mappedValueI, mapErr := sm.mapValue(valueI, valueV, prefix...)
		if mapErr != nil {
			err = multierror.Append(err, mapErr)
			continue
		}
		s = append(s, mappedValueI)
	}

	return
}

func (sm *StructMapper) mapValue(i interface{}, v reflect.Value, prefix ...string) (value interface{}, err error) {
	// Check if the passed interface implements encoding.TextMarshaler, in which case we use the marshaler
	// for generating the value
	if marshaler, ok := i.(encoding.TextMarshaler); ok && marshaler != nil {
		value, err = marshaler.MarshalText()
		return
	}

	// At this point it is safe to get rid of a possible pointer...
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	} else if v.Kind() == reflect.Ptr {
		// No-op for nil-pointers
		return
	}

	// Per-type handling
	switch v.Kind() {
	case reflect.Struct:
		// Handle struct
		value, err = sm.mapStruct(v, prefix...)
	case reflect.Slice, reflect.Array:
		value, err = sm.mapSlice(v, prefix...)
	case reflect.Map:
		value, err = sm.mapMap(v, prefix...)
	default:
		// All other types are mapped as-is
		value = i
	}

	return
}

func (sm *StructMapper) isNilOrEmpty(i interface{}, v reflect.Value) bool {
	// Simple case: interface is nil
	if i == nil {
		return true
	}

	// Hard case: check if interface has "zero" value (ie. empty string, zero integer, etc.)
	return reflect.DeepEqual(i, reflect.Zero(v.Type()).Interface())
}

func (sm *StructMapper) mapStruct(v reflect.Value, prefix ...string) (m map[string]interface{}, err error) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return
	}

	t := v.Type()

	// Create a new map that is pre-allocated with the number of fields v contains
	m = make(map[string]interface{}, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		fieldD := t.Field(i)
		fieldV := v.Field(i)

		if fieldD.Anonymous {
			// Skip anonymous fields...
			continue
		}

		fieldName := fieldD.Name

		if !unicode.IsUpper([]rune(fieldName)[0]) {
			// Ignore private fields
			continue
		}

		tagValue := fieldD.Tag.Get(sm.tagName)
		omitEmpty := false

		// Handle the tag, if it was present
		if tagValue != "" {
			var tagErr error
			fieldName, omitEmpty, tagErr = parseTag(tagValue)
			if tagErr != nil {
				// Parsing the tag failed, ignore the field and carry on
				err = multierror.Append(err, tagErr)
				continue
			}

			if fieldName == "-" {
				// Tag defines that the field shall be ignored, so carry on
				continue
			}
		}

		fieldI := fieldV.Interface()

		if omitEmpty && sm.isNilOrEmpty(fieldI, fieldV) {
			// If omitEmpty is set and the field is nil or empty carry on
			continue
		} else if fieldI != nil {
			// If field is non-nil, map it...
			mappedFieldI, mappingErr := sm.mapValue(fieldI, fieldV, append(prefix, fieldName)...)
			if mappingErr != nil {
				// If mapping failed, add an error
				fullFieldName := strings.Join(append(prefix, fieldName), ".")
				err = multierror.Append(err, multierror.Prefix(mappingErr, fullFieldName))
				continue
			}

			if omitEmpty && sm.isNilOrEmpty(mappedFieldI, reflect.ValueOf(mappedFieldI)) {
				// If omitEmpty is set and the mapped value is nil or zero carry on
				continue
			}
			// Override fieldI with the mapped value
			fieldI = mappedFieldI
		}

		m[fieldName] = fieldI
	}

	return
}

func (sm *StructMapper) generateMap(s interface{}, prefix ...string) (map[string]interface{}, error) {
	if s == nil {
		// If the input struct is nil, return an empty map
		return map[string]interface{}{}, nil
	}

	// Verify that we are working on a struct...
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}

	return sm.mapStruct(v, prefix...)
}

func (sm *StructMapper) unmapPtr(in interface{}, out reflect.Value, t reflect.Type) error {
	child := reflect.New(t.Elem())
	if err := sm.unmapValue(in, child.Elem(), child.Elem().Type()); err != nil {
		return err
	}
	out.Set(child)
	return nil
}

func (sm *StructMapper) unmapSlice(in interface{}, out reflect.Value, t reflect.Type) (err error) {
	inSlice := reflect.ValueOf(in)
	if inSlice.Kind() != reflect.Slice {
		return errors.New("Not a slice")
	}

	outSlice := reflect.MakeSlice(t, inSlice.Len(), inSlice.Cap())
	for i := 0; i < inSlice.Len(); i++ {
		inElem := inSlice.Index(i)
		outElem := outSlice.Index(i)

		inV := reflect.ValueOf(inElem.Interface())
		elemV := reflect.New(outElem.Type()).Elem()

		if unmapErr := sm.unmapValue(inV.Interface(), elemV, elemV.Type()); unmapErr != nil {
			err = multierror.Append(err, multierror.Prefix(unmapErr, fmt.Sprintf("@%d", i)))
			continue
		}

		outSlice.Index(i).Set(elemV)
	}

	if err == nil {
		out.Set(outSlice)
	}

	return
}

func (sm *StructMapper) unmapMap(in interface{}, out reflect.Value, t reflect.Type) (err error) {
	inMap := reflect.ValueOf(in)
	if inMap.Kind() != reflect.Map {
		return errors.New("Not a map")
	}

	outMap := reflect.MakeMap(t)

	for _, keyElem := range inMap.MapKeys() {

		keyV := reflect.ValueOf(keyElem.Interface())
		key := keyV.Interface()
		outKey := reflect.New(keyV.Type()).Elem()
		if unmapErr := sm.unmapValue(key, outKey, outKey.Type()); unmapErr != nil {
			err = multierror.Append(err, multierror.Prefix(unmapErr, fmt.Sprintf("@%+v (key)", key)))
			continue
		}
		valueElem := inMap.MapIndex(keyV)
		valueV := reflect.ValueOf(valueElem.Interface())
		value := valueV.Interface()
		outValue := reflect.New(valueV.Type()).Elem()

		if unmapErr := sm.unmapValue(value, outValue, outValue.Type()); unmapErr != nil {
			err = multierror.Append(err, multierror.Prefix(unmapErr, fmt.Sprintf("@%+v (value)", key)))
			continue
		}

		outMap.SetMapIndex(outKey, outValue)
	}

	if err == nil {
		// Special case: out may be a struct or struct pointer...
		out.Set(outMap)
	}

	return
}

func (sm *StructMapper) unmapArray(in interface{}, out reflect.Value, t reflect.Type) (err error) {
	inArray := reflect.ValueOf(in)

	if inArray.Kind() != reflect.Array && inArray.Kind() != reflect.Slice {
		return errors.New("Not an array or slice")
	}

	outArray := reflect.New(t).Elem()

	for i := 0; i < inArray.Len(); i++ {
		outElem := outArray.Index(i)
		inValue := inArray.Index(i)

		if unmapErr := sm.unmapValue(inValue.Interface(), outElem, outElem.Type()); unmapErr != nil {
			err = multierror.Append(err, multierror.Prefix(unmapErr, fmt.Sprintf("@%d", i)))
			continue
		}
	}

	if err == nil {
		out.Set(outArray)
	}

	return
}

func (sm *StructMapper) unmapValue(in interface{}, out reflect.Value, t reflect.Type) error {
	switch out.Kind() {
	case reflect.Ptr:
		return sm.unmapPtr(in, out, t)
	case reflect.Struct:
		return sm.unmapStruct(in, out, t)
	case reflect.Slice:
		return sm.unmapSlice(in, out, t)
	case reflect.Map:
		return sm.unmapMap(in, out, t)
	case reflect.Array:
		return sm.unmapArray(in, out, t)
	default:
		// TODO: type check
		out.Set(reflect.ValueOf(in))
		return nil
	}

	return errors.New("Not implemented")
}

func (sm *StructMapper) unmapStruct(in interface{}, out reflect.Value, t reflect.Type) (err error) {
	if out.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	m, ok := in.(map[string]interface{})
	if !ok {
		return ErrInvalidMap
	}

	// Hold the values of the modified fields in a map, which will be applied shortly before
	// this function returns.
	// This ensures we do not modify the target struct at all in case of an error
	modifiedFields := make(map[int]reflect.Value, t.NumField())

	// Iterate over all fields of the passed struct
	for i := 0; i < out.NumField(); i++ {
		fieldD := t.Field(i)
		fieldV := out.Field(i)

		if fieldD.Anonymous {
			// Skip anonymous fields...
			continue
		}

		fieldName := fieldD.Name

		if !unicode.IsUpper([]rune(fieldName)[0]) {
			// Ignore private fields
			continue
		}

		tagValue := fieldD.Tag.Get(sm.tagName)

		// Handle the tag, if it was present
		if tagValue != "" {
			var tagErr error
			fieldName, _, tagErr = parseTag(tagValue)
			if tagErr != nil {
				// Parsing the tag failed, ignore the field and carry on
				err = multierror.Append(err, tagErr)
				continue
			}

			if fieldName == "-" {
				// Tag defines that the field shall be ignored, so carry on
				continue
			}
		}

		// Look up value of "fieldName" in map
		mapValue, ok := m[fieldName]
		if !ok {
			// Value not in map, ignore it
			continue
		}

		if fieldV.Kind() == reflect.Interface {
			// Setting interfaces is unsupported.
			err = multierror.Append(err, multierror.Prefix(ErrFieldIsInterface, fieldName+":"))
			continue
		}

		targetV := reflect.New(fieldD.Type).Elem()
		if unmapErr := sm.unmapValue(mapValue, targetV, fieldD.Type); unmapErr != nil {
			err = multierror.Append(err, multierror.Prefix(unmapErr, fieldName+":"))
			continue
		} else {
			modifiedFields[i] = targetV
		}
	}

	// Apply changes to all modified fields in case no error happened during processing.
	if err == nil {
		// Apply changes to all modified fields
		for fieldIndex, fieldValue := range modifiedFields {
			out.Field(fieldIndex).Set(fieldValue)
		}
	}
	return
}

func (sm *StructMapper) generateStruct(m map[string]interface{}, s interface{}) error {
	if m == nil {
		return ErrMapIsNil
	}

	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		return ErrNotAStructPointer
	}

	v = v.Elem()

	return sm.unmapStruct(m, v, v.Type())
}

func (sm *StructMapper) ToStruct(m map[string]interface{}, s interface{}) error {
	return sm.generateStruct(m, s)
}

func (sm *StructMapper) ToMap(s interface{}) (map[string]interface{}, error) {
	return sm.generateMap(s)
}

// NewStructMapper initializes a new StructMapper instance
func NewStructMapper(options ...StructMapperOption) (*StructMapper, error) {
	sm := &StructMapper{}

	var err error

	// Apply default options first
	for _, opt := range structMapperDefaultOptions {
		if err := opt(sm); err != nil {
			// Panic if default option could not be applied
			panic(err)
		}
	}

	// ... and passed options afterwards.
	// This way the passed options override the default options
	for _, opt := range options {
		if optErr := opt(sm); optErr != nil {
			err = multierror.Append(err, optErr)
		}
	}

	if err != nil {
		return nil, err
	}

	return sm, nil
}
