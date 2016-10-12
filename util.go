package structconf

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// generateMapFromStruct takes a struct and encoding and generates a map[string]interface{}
// from that struct.
func generateMapFromStruct(s interface{}, e Encoding) (map[string]interface{}, error) {
	// Marshal to byte array
	data, err := e.MarshalFrom(s)
	if err != nil {
		return nil, err
	}

	// Unmarshal from byte array to map
	m := make(map[string]interface{})

	if err := e.UnmarshalTo(data, m); err != nil {
		return nil, err
	}

	return m, nil
}

// mergeMaps merges the passed maps
// The resulting map contains all keys that existed in either of the passed maps.
// For keys that exist in both maps, the value from the "b" map takes precedence,
// iff it is not a zero-value.
// Keys that are present in both maps are expected to be of the same type. If this is not the case,
// an error will be returned.
func mergeMaps(a, b map[string]interface{}, prefix ...string) (map[string]interface{}, error) {
	// Allocate the resulting map with a sane default
	m := make(map[string]interface{}, int(math.Max(float64(len(a)), float64(len(b)))))

	// Iterate over map a and set all values in result map
	for key, value := range a {
		// Copy only if value is non-nil
		m[key] = value
	}

	// Now merge in the values from map b
	for key, valueB := range b {
		valB := reflect.ValueOf(valueB)
		valueA, exists := m[key]

		// If the key does not yet exist, or the value in the merged map is nil,
		// just copy over the new value
		if !exists || valueA == nil {
			m[key] = valueB
			continue
		}

		// If the value in B is nil, keep the value from A
		if valueB == nil {
			continue
		}

		valA := reflect.ValueOf(valueA)

		// Check if the A value is zero-ish, in which case we use the value from B
		if reflect.DeepEqual(valueA, reflect.Zero(valA.Type()).Interface()) {
			m[key] = valueB
			continue
		}

		// Check if the B value is zero-ish, in which case we skip further processing,
		// and keep the value from A
		if reflect.DeepEqual(valueB, reflect.Zero(valB.Type()).Interface()) {
			continue
		}

		// Check if both values are of the same

		if valA.Kind() != valB.Kind() {
			return nil, fmt.Errorf("Kind mismatch for key %s: %s != %s",
				strings.Join(append(prefix, key), "."), valA.Kind().String(), valB.Kind().String())
		} else if valA.Type() != valB.Type() {
			return nil, fmt.Errorf("Type mismatch for key %s: %s != %s",
				strings.Join(append(prefix, key), "."), valA.Type().String(), valB.Type().String())
		}

		// At this point we are sure that both values are of the same type.
		// If these are map[string]interface{} instances, we start a recursion.
		if mapA, ok := valueA.(map[string]interface{}); ok {
			mapB := valueB.(map[string]interface{})
			mergedMap, err := mergeMaps(mapA, mapB, append(prefix, key)...)
			if err != nil {
				return nil, err
			}
			m[key] = mergedMap
			continue
		}

		// In the default case, use the value from B
		m[key] = valueB
	}

	return m, nil
}
