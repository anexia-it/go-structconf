package structconf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeMapsBEmpty(t *testing.T) {
	a := map[string]interface{}{
		"0": 0,
		"1": 1,
	}
	b := map[string]interface{}{}

	result, err := MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, a, result)
}

func TestMergeMapsDifferentKeys(t *testing.T) {
	a := map[string]interface{}{
		"0": 0,
		"1": 1,
	}
	b := map[string]interface{}{
		"2": 2,
		"3": 3,
	}

	expectedResult := make(map[string]interface{}, len(a)+len(b))
	for k, v := range a {
		expectedResult[k] = v
	}

	for k, v := range b {
		expectedResult[k] = v
	}

	result, err := MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, expectedResult, result)
}

func TestMergeMapsNilValueA(t *testing.T) {
	a := map[string]interface{}{
		"a": nil,
		"b": 0,
		"c": "",
	}

	b := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": "c",
	}

	result, err := MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, b, result)
}

func TestMergeMapsNilValueB(t *testing.T) {
	a := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": "c",
	}

	b := map[string]interface{}{
		"a": nil,
		"b": 0,
		"c": "",
	}

	result, err := MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, a, result)
}

func TestMergeMapsKindMismatch(t *testing.T) {
	a := map[string]interface{}{
		"a": 1,
	}
	b := map[string]interface{}{
		"a": "a",
	}

	result, err := MergeMaps(a, b)
	require.Nil(t, result)
	require.EqualError(t, err, "Kind mismatch for key a: int != string")
}

func TestMergeMapsTypeMismatch(t *testing.T) {
	integer := 1
	str := "a"

	a := map[string]interface{}{
		"a": &integer,
	}
	b := map[string]interface{}{
		"a": &str,
	}

	result, err := MergeMaps(a, b)
	require.Nil(t, result)
	require.EqualError(t, err, "Type mismatch for key a: *int != *string")
}

func TestMergeMapsRecursion(t *testing.T) {
	a := map[string]interface{}{
		"a": 1,
		"m": map[string]interface{}{
			"ma": 11,
		},
	}

	b := map[string]interface{}{
		"b": 2,
		"m": map[string]interface{}{
			"mb": 21,
		},
	}

	expectedResult := map[string]interface{}{
		"a": 1,
		"b": 2,
		"m": map[string]interface{}{
			"ma": 11,
			"mb": 21,
		},
	}

	result, err := MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, expectedResult, result)
}

func TestMergeMapsRecursionError(t *testing.T) {
	a := map[string]interface{}{
		"a": 1,
		"m": map[string]interface{}{
			"c": 11,
		},
	}

	b := map[string]interface{}{
		"b": 2,
		"m": map[string]interface{}{
			"c": "21",
		},
	}

	result, err := MergeMaps(a, b)
	require.Nil(t, result)
	require.EqualError(t, err, "Kind mismatch for key m.c: int != string")
}

func TestMergeMapsDefaultHandling(t *testing.T) {
	a := map[string]interface{}{
		"a": 1,
	}

	b := map[string]interface{}{
		"a": 2,
	}

	result, err := MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, b, result)
}
