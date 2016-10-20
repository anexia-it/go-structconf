package structconf_test

import (
	"testing"

	"github.com/anexia-it/go-structconf"
	"github.com/hashicorp/errwrap"
	"github.com/stretchr/testify/require"
)

func TestMergeMapsBEmpty(t *testing.T) {
	a := map[string]interface{}{
		"0": 0,
		"1": 1,
	}
	b := map[string]interface{}{}

	result, err := structconf.MergeMaps(a, b)
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

	result, err := structconf.MergeMaps(a, b)
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

	result, err := structconf.MergeMaps(a, b)
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

	result, err := structconf.MergeMaps(a, b)
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

	result, err := structconf.MergeMaps(a, b)
	require.Nil(t, result)
	require.Error(t, err)
	w, ok := err.(errwrap.Wrapper)
	require.EqualValues(t, true, ok, "Returned error is not an errwrap.Wrapper")
	wrapped := w.WrappedErrors()
	require.Len(t, wrapped, 1)
	require.EqualError(t, wrapped[0], "key a: Kind mismatch: int != string")
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

	result, err := structconf.MergeMaps(a, b)
	require.Nil(t, result)
	require.Error(t, err)
	w, ok := err.(errwrap.Wrapper)
	require.EqualValues(t, true, ok, "Returned error is not an errwrap.Wrapper")
	wrapped := w.WrappedErrors()
	require.Len(t, wrapped, 1)
	require.EqualError(t, wrapped[0], "key a: Kind mismatch: int != string")
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

	result, err := structconf.MergeMaps(a, b)
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

	result, err := structconf.MergeMaps(a, b)
	require.Nil(t, result)
	require.Error(t, err)
	w, ok := err.(errwrap.Wrapper)
	require.EqualValues(t, true, ok, "Returned error is not an errwrap.Wrapper")
	wrapped := w.WrappedErrors()
	require.Len(t, wrapped, 1)
	require.EqualError(t, wrapped[0], "key m: key c: Kind mismatch: int != string")
}

func TestMergeMapsDefaultHandling(t *testing.T) {
	a := map[string]interface{}{
		"a": 1,
	}

	b := map[string]interface{}{
		"a": 2,
	}

	result, err := structconf.MergeMaps(a, b)
	require.NoError(t, err)
	require.EqualValues(t, b, result)
}

func TestMergeMapsMapStringInterfaceMapInterfaceInterface(t *testing.T) {
	a := map[string]interface{}{
		"a": map[string]interface{}{
			"b": 1,
		},
	}

	b := map[string]interface{}{
		"a": map[interface{}]interface{}{
			"b": 2,
		},
	}

	expected := map[string]interface{}{
		"a": map[string]interface{}{
			"b": 2,
		},
	}

	require.NotPanics(t, func() {
		result, err := structconf.MergeMaps(a, b)
		require.NoError(t, err)
		require.EqualValues(t, expected, result)
	})
}

func TestMergeLoggingConfigWithMapStringString(t *testing.T) {
	a := map[string]interface{}{
		"a:": map[string]interface{}{
			"b:": "b0",
			"c:": map[string]string{
				"l1": "c0",
				"l2": "c1",
			},
		},
	}

	b := map[string]interface{}{
		"a:": map[string]interface{}{
			"b:": "b1",
			"c:": map[string]string{
				"l2": "c2",
				"l3": "c3",
			},
		},
	}

	expected := map[string]interface{}{
		"a:": map[string]interface{}{
			"b:": "b1",
			"c:": map[string]string{
				"l1": "c0",
				"l2": "c2",
				"l3": "c3",
			},
		},
	}

	require.NotPanics(t, func() {
		result, err := structconf.MergeMaps(a, b)
		require.NoError(t, err)
		require.EqualValues(t, expected, result)
	})
}

func TestMergeLoggingConfigWithListInterface(t *testing.T) {

	a := map[string]interface{}{
		"a:": map[string]interface{}{
			"b:": "b0",
			"c:": []interface{}{
				map[string]interface{}{
					"d": "c0",
					"e": "c1",
				},
			},
		},
	}

	b := map[string]interface{}{
		"a:": map[string]interface{}{
			"b:": "", // this is empty, so b0 should be in the result
			"c:": []interface{}{
				map[string]interface{}{
					"d": "c2",
					"e": "c3",
				},
				map[string]interface{}{
					"d": "c4",
					"e": "c5",
				},
			},
		},
	}

	expected := map[string]interface{}{
		"a:": map[string]interface{}{
			"b:": "b0",
			"c:": []interface{}{
				map[string]interface{}{
					"d": "c2",
					"e": "c3",
				},
				map[string]interface{}{
					"d": "c4",
					"e": "c5",
				},
			},
		},
	}

	require.NotPanics(t, func() {
		result, err := structconf.MergeMaps(a, b)
		require.NoError(t, err)
		require.EqualValues(t, expected, result)
	})
}

func TestMergeValues_ANil(t *testing.T) {
	merged, err := structconf.MergeValues(nil, "test_a_nil")
	require.NoError(t, err)
	require.EqualValues(t, "test_a_nil", merged)

}

func TestMergeValuesBNil(t *testing.T) {
	merged, err := structconf.MergeValues("test_b_nil", nil)
	require.NoError(t, err)
	require.EqualValues(t, "test_b_nil", merged)
}

func TestMergeValues_BZero(t *testing.T) {
	merged, err := structconf.MergeValues("test_b_zero", "")
	require.NoError(t, err)
	require.EqualValues(t, "test_b_zero", merged)
}

func TestMergeValues_AZero(t *testing.T) {
	merged, err := structconf.MergeValues("", "test_a_zero")
	require.NoError(t, err)
	require.EqualValues(t, "test_a_zero", merged)
}

func TestMergeValues_PtrNil(t *testing.T) {
	a := (*int)(nil)
	b := (*int)(nil)

	merged, err := structconf.MergeValues(a, b)
	require.NoError(t, err)
	require.Nil(t, merged)
}

func TestMergeValues_PtrIncompatibleValues(t *testing.T) {
	a := 0
	b := "test"

	merged, err := structconf.MergeValues(&a, &b)
	require.EqualError(t, err, "Kind mismatch: int != string")
	require.Nil(t, merged)
}

func TestMergeValues_PtrCompatible(t *testing.T) {
	a := "test0"
	b := "test1"

	merged, err := structconf.MergeValues(&a, &b)
	require.NoError(t, err)
	require.IsType(t, (*string)(nil), merged)
	require.EqualValues(t, &b, merged)
}

func TestMergeValues_AUnsupportedKind(t *testing.T) {
	a := func() {}
	b := "test0"
	merged, err := structconf.MergeValues(a, b)
	require.EqualError(t, err, "Kind func unsupported for merging")
	require.Nil(t, merged)
}

func TestMergeValues_BUnsupportedKind(t *testing.T) {
	a := "test0"
	b := make(chan bool)
	defer close(b)
	merged, err := structconf.MergeValues(a, b)
	require.EqualError(t, err, "Kind chan unsupported for merging")
	require.Nil(t, merged)
}

func TestMergeValues_MapsSimple(t *testing.T) {
	a := map[string]string{"a": "0", "b": "1"}
	b := map[string]string{"a": "2", "c": "3", "d": "4"}
	expected := map[string]string{"a": "2", "b": "1", "c": "3", "d": "4"}

	merged, err := structconf.MergeValues(a, b)
	require.NoError(t, err)
	require.EqualValues(t, expected, merged)
}

func TestMergeValues_MapsIncompatibleKeys(t *testing.T) {
	a := map[int]string{0: "0"}
	b := map[interface{}]string{"a": "1", 1: "2"}

	merged, err := structconf.MergeValues(a, b)
	require.Error(t, err)
	require.Nil(t, merged)

	w, ok := err.(errwrap.Wrapper)
	require.EqualValues(t, true, ok, "Returned error is not an errwrap.Wrapper")
	wrapped := w.WrappedErrors()
	require.Len(t, wrapped, 1)
	require.EqualError(t, wrapped[0], "key a: Kind mismatch: int != string")
}

func TestMergeValues_MapsMultipleIncompatibleKeys(t *testing.T) {
	a := map[int]string{0: "0"}
	b := map[interface{}]string{"a": "1", "b": "3", 1: "2"}

	merged, err := structconf.MergeValues(a, b)
	require.Error(t, err)
	require.Nil(t, merged)

	w, ok := err.(errwrap.Wrapper)
	require.EqualValues(t, true, ok, "Returned error is not an errwrap.Wrapper")
	wrapped := w.WrappedErrors()
	require.Len(t, wrapped, 2)
	errorStrings := make([]string, len(wrapped))
	for i, w := range wrapped {
		errorStrings[i] = w.Error()
	}

	require.Contains(t, errorStrings, "key a: Kind mismatch: int != string")
	require.Contains(t, errorStrings, "key b: Kind mismatch: int != string")
}

func TestMergeValues_Slices(t *testing.T) {
	a := []string{"a", "b"}
	b := []string{"c", "d"}

	merged, err := structconf.MergeValues(a, b)
	require.NoError(t, err)
	require.EqualValues(t, b, merged)
}

func TestMergeValues_SlicesBEmpty(t *testing.T) {
	a := []string{"a", "b"}
	b := []string{}

	merged, err := structconf.MergeValues(a, b)
	require.NoError(t, err)
	require.EqualValues(t, a, merged)
}

func TestMergeValues_Convertible(t *testing.T) {
	a := 5
	b := 8.0

	merged, err := structconf.MergeValues(a, b)
	require.NoError(t, err)
	require.EqualValues(t, 8, merged)
}

func TestMergeValues_ConvertibleCustomStringType(t *testing.T) {
	type testString string

	a := testString("a")
	b := []byte("b")

	merged, err := structconf.MergeValues(a, b)
	require.NoError(t, err)
	require.IsType(t, testString(""), merged)
	require.EqualValues(t, b, merged)
}
