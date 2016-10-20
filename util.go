package structconf

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-multierror"
)

// MergeMaps merges the passed maps
// The resulting map contains all keys that existed in either of the passed maps.
// For keys that exist in both maps, the value from the "b" map takes precedence,
// iff it is not a zero-value.
// Keys that are present in both maps are expected to be of the same type. If this is not the case,
// an error will be returned.
func MergeMaps(a, b map[string]interface{}) (map[string]interface{}, error) {
	merged, err := mergeMaps(reflect.ValueOf(a), reflect.ValueOf(b))
	if err != nil {
		return nil, err
	}

	// As the inputs are map[string]interface{} values, the result will always
	// be a map[string]interface{}, which means we can cast here without any checks
	return merged.(map[string]interface{}), nil
}

// reflect.Kind values which cannot be merged
var mergeUnsupportedKinds = []reflect.Kind{
	reflect.Invalid,
	reflect.Chan,
	reflect.Func,
	reflect.Interface,
	reflect.UnsafePointer,
}

// isMergeUnsupportedKind checks if a given kind is unsupported for merging
func isMergeUnsupportedKind(k reflect.Kind) error {
	for _, unsupportedKind := range mergeUnsupportedKinds {
		if unsupportedKind == k {
			return fmt.Errorf("Kind %s unsupported for merging", k.String())
		}
	}
	return nil
}

// MergeValues "merges" two values
//
// The internal logic is as follows:
//
// - If b is zero, return a
// - If a is zero, return b
// - If both values are non-zero and scalar types, convert b with a's type, if possible.
// - If both values are non-zero and slices or arrays, return b
// - If both values are non-zero and maps, merge each element of the map using the above logic
func MergeValues(a, b interface{}) (interface{}, error) {
	// Simple case: a is nil
	if a == nil {
		return b, nil
	}

	// Simple case: b is nil
	if b == nil {
		return a, nil
	}

	valueB := reflect.ValueOf(b)

	// Simple case: b is zero
	if reflect.DeepEqual(b, reflect.Zero(valueB.Type()).Interface()) {
		return a, nil
	}

	valueA := reflect.ValueOf(a)

	// At this point both a and b are non-nil and non-zero
	kindA := valueA.Kind()
	kindB := valueB.Kind()

	// Special case: both values are pointers
	if kindA == reflect.Ptr && kindA == kindB {
		return mergePointers(valueA, valueB)
	}

	// Check if both kinds are supported for merging
	if err := isMergeUnsupportedKind(kindA); err != nil {
		return nil, err
	}

	if err := isMergeUnsupportedKind(kindB); err != nil {
		return nil, err
	}

	// Special case: both are maps
	if kindA == reflect.Map && kindA == kindB {
		// Special case: maps need merging
		return mergeMaps(valueA, valueB)
	}

	// Special case: both are slices or arrays
	if kindA == kindB && (kindA == reflect.Slice || kindA == reflect.Array) {
		return mergeSlices(valueA, valueB)
	}

	// Finally: try merging scalar values
	return convertScalarValues(valueA, valueB)
}

// mergePointers handles merging of pointer values
func mergePointers(a reflect.Value, b reflect.Value) (interface{}, error) {
	merged, err := MergeValues(a.Elem().Interface(), b.Elem().Interface())
	if err != nil {
		return nil, err
	} else if merged == nil {
		return nil, nil
	}

	// Ensure we do return a pointer again
	res := reflect.New(a.Type().Elem())
	res.Elem().Set(reflect.ValueOf(merged))
	return res.Interface(), nil
}

// mergeMaps handles merging of map values
func mergeMaps(a reflect.Value, b reflect.Value) (m interface{}, err error) {
	mValue := reflect.MakeMap(a.Type())
	var sampleKeyValue reflect.Value

	// Initialize result map with all values from a
	for _, k := range a.MapKeys() {
		mValue.SetMapIndex(k, a.MapIndex(k))
		sampleKeyValue = k
	}

	// Now iterate over all keys from b and set them on our result map
	for _, k := range b.MapKeys() {
		// Convert key
		convertedKeyIntf, convertErr := convertScalarValues(sampleKeyValue, k)
		if convertErr != nil {
			err = multierror.Append(err, multierror.Prefix(convertErr, fmt.Sprintf("key %v:", k.Interface())))
			continue
		}

		key := reflect.ValueOf(convertedKeyIntf)
		v := b.MapIndex(k)

		// Check if key exists
		existingV := mValue.MapIndex(key)
		if !existingV.IsValid() {
			// Key does not exist.
			mValue.SetMapIndex(key, v)
			continue
		}

		// The key does already exist, in which case we need to merge the existing value and the
		// value from b
		mergedV, mergeErr := MergeValues(existingV.Interface(), v.Interface())
		if mergeErr != nil {
			err = multierror.Append(err, multierror.Prefix(mergeErr, fmt.Sprintf("key %v:", convertedKeyIntf)))
			continue
		}
		mValue.SetMapIndex(key, reflect.ValueOf(mergedV))
	}

	if err != nil {
		return
	}
	m = mValue.Interface()
	return
}

// mergeSlices merges two slices
//
// The logic is very simple: if b is not empty return b, otherwise, return a
func mergeSlices(a reflect.Value, b reflect.Value) (interface{}, error) {
	if b.Len() != 0 {
		return b.Interface(), nil
	}

	return a.Interface(), nil
}

// mergeScalarValues merges two scalar values
//
// This function guarantees that the returned value is of the same type as a, so conversion is attempted.
// If conversion fails, an error is returned
func convertScalarValues(a reflect.Value, b reflect.Value) (interface{}, error) {
	// Detect the actual types of both a and b
	aType := reflect.TypeOf(a.Interface())
	bType := reflect.TypeOf(b.Interface())

	// Simple case: both values are of the same type, so simply return b, as b always overrides a
	if aType == bType {
		return b.Interface(), nil
	}

	// If b can be converted to a, just convert the value
	if bType.ConvertibleTo(aType) {
		return reflect.ValueOf(b.Interface()).Convert(aType).Interface(), nil
	}

	return nil, fmt.Errorf("Kind mismatch: %s != %s", aType.Kind(), bType.Kind())
}
