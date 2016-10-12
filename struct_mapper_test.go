package structconf

import (
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

func TestStructMapperTagName(t *testing.T) {
	sm, err := NewStructMapper(StructMapperOptionTagName("test"))

	require.NoError(t, err)
	require.NotNil(t, sm)

	// Check if the supplied tag name was set
	require.EqualValues(t, "test", sm.tagName)

	// Try setting an empty tag name

	sm, err = NewStructMapper(StructMapperOptionTagName(""))
	require.Error(t, err)
	require.Nil(t, sm)

	multiErr, ok := err.(*multierror.Error)
	require.EqualValues(t, ok, true, "Returned error is not a multierror.Error")
	require.Len(t, multiErr.WrappedErrors(), 1)
	require.EqualError(t, multiErr.WrappedErrors()[0], ErrTagNameEmpty.Error())
}

func TestNewStructMapper(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	// Check if default tag name is set
	require.EqualValues(t, StructMapperDefaultTagName, sm.tagName)
}

type structMapperTestStructInner struct {
	A string `config:"eff,omitempty"`
}

type structMapperTestStruct struct {
	// Even though a tag is set, this should be ignored
	privateTest string                       `config:"private"`
	A           string                       `config:"a"`
	B           int                          `config:"b"`
	C           float64                      `config:"c"`
	D           uint64                       `config:"dee,omitempty"`
	E           *structMapperTestStructInner `config:"e"`
}

type structMapperTestStructArraySlice struct {
	A []string                       `config:"a"`
	B []*structMapperTestStructInner `config:"b,omitempty"`
	C [2]string                      `config:"c"`
}

func TestStructMapper_ToMap(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	// Call ToMap with nil value
	m, err := sm.ToMap(nil)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Len(t, m, 0)

	// Call ToMap with non-struct
	m, err = sm.ToMap("test")
	require.EqualError(t, err, ErrNotAStruct.Error())
	require.Nil(t, m)

	// Call ToMap with pointer to non-struct
	testValue := "test"
	m, err = sm.ToMap(&testValue)
	require.EqualError(t, err, ErrNotAStruct.Error())
	require.Nil(t, m)

	// Construct simple test case: all fields present
	testStruct := &structMapperTestStruct{
		A: "0",
		B: 1,
		C: 2.1,
		D: 3,
		E: &structMapperTestStructInner{
			A: "4",
		},
	}

	expectedMap := map[string]interface{}{
		"a":   "0",
		"b":   1,
		"c":   2.1,
		"dee": uint64(3),
		"e": map[string]interface{}{
			"eff": "4",
		},
	}

	m, err = sm.ToMap(testStruct)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, m)

	// Test if omission of fields works
	testStruct = &structMapperTestStruct{
		A: "0",
		B: 1,
		C: 2.1,
		E: &structMapperTestStructInner{},
	}

	expectedMap = map[string]interface{}{
		"a": "0",
		"b": 1,
		"c": 2.1,
		"e": map[string]interface{}{},
	}

	m, err = sm.ToMap(testStruct)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, m)

	testStructArraySlice := &structMapperTestStructArraySlice{
		A: []string{"0.0", "0.1"},
		B: []*structMapperTestStructInner{
			{
				A: "1.0",
			},
			{
				A: "",
			},
		},
		C: [2]string{"2.0", ""},
	}

	expectedMap = map[string]interface{}{
		"a": []interface{}{"0.0", "0.1"},
		"b": []interface{}{
			map[string]interface{}{
				"eff": "1.0",
			},
			map[string]interface{}{},
		},
		"c": []interface{}{"2.0", ""},
	}

	m, err = sm.ToMap(testStructArraySlice)
	require.NoError(t, err)
	require.EqualValues(t, expectedMap, m)
}

func TestStructMapper_ToStruct(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	// Call ToStruct with nil map
	require.EqualError(t, sm.ToStruct(nil, &structMapperTestStructInner{}), ErrMapIsNil.Error())

	testValue := "test"

	// Call ToStruct with non-struct pointer
	require.EqualError(t, sm.ToStruct(make(map[string]interface{}), &testValue), ErrNotAStruct.Error())

	// Call ToStruct with non-struct pointer
	require.EqualError(t, sm.ToStruct(make(map[string]interface{}), structMapperTestStructInner{}),
		ErrNotAStructPointer.Error())
}

type structMapperTestInterfaceField struct {
	A interface{} `config:"x"`
}

func TestStructMapper_ToStruct_InterfaceField(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	m := map[string]interface{}{
		"x": "test",
	}

	target := &structMapperTestInterfaceField{}

	err = sm.ToStruct(m, target)
	require.Error(t, err)
	me, ok := err.(*multierror.Error)
	require.EqualValues(t, true, ok, "Returned error is not a *multierror.Error")
	require.Len(t, me.Errors, 1)
	e := me.Errors[0]
	// Test if the error is correct...
	require.Error(t, e, multierror.Prefix(ErrFieldIsInterface, "x: ").Error())
}

func TestStructMapper_ToStruct_Simple(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	// Simple test case: single field, no nesting
	expected := &structMapperTestStructInner{
		A: "test",
	}

	m := map[string]interface{}{
		"eff": "test",
	}

	target := &structMapperTestStructInner{}

	require.NoError(t, sm.ToStruct(m, target))
	require.EqualValues(t, expected, target)
}

func TestStructMapper_ToStruct_NestedSimple(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	// Construct simple test case: all fields present
	expected := &structMapperTestStruct{
		A: "0",
		B: 1,
		C: 2.1,
		D: 3,
		E: &structMapperTestStructInner{
			A: "4",
		},
	}

	m := map[string]interface{}{
		"a":   "0",
		"b":   1,
		"c":   2.1,
		"dee": uint64(3),
		"e": map[string]interface{}{
			"eff": "4",
		},
	}

	target := &structMapperTestStruct{}

	require.NoError(t, sm.ToStruct(m, target))
	require.EqualValues(t, expected, target)
}

func TestStructMapper_ToStruct_ArraySlice(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	expected := &structMapperTestStructArraySlice{
		A: []string{"0.0", "0.1"},
		B: []*structMapperTestStructInner{
			{
				A: "1.0",
			},
			{
				A: "",
			},
		},
		C: [2]string{"2.0", ""},
	}

	m := map[string]interface{}{
		"a": []interface{}{"0.0", "0.1"},
		"b": []interface{}{
			map[string]interface{}{
				"eff": "1.0",
			},
			map[string]interface{}{},
		},
		"c": []interface{}{"2.0", ""},
	}

	target := &structMapperTestStructArraySlice{}

	require.NoError(t, sm.ToStruct(m, target))
	require.EqualValues(t, expected, target)
}

func TestStructMapper_Roundtrip(t *testing.T) {
	// Initialize StructMapper without options
	sm, err := NewStructMapper()
	require.NoError(t, err)
	require.NotNil(t, sm)

	source := &structMapperTestStruct{
		A: "0",
		B: 1,
		C: 2.1,
		D: 3,
		E: &structMapperTestStructInner{
			A: "4",
		},
	}

	target := &structMapperTestStruct{}

	// Convert struct to map
	m, err := sm.ToMap(source)
	require.NoError(t, err)
	require.NotNil(t, m)

	// Convert back to struct
	require.NoError(t, sm.ToStruct(m, target))

	// Check if source and target are equal
	require.EqualValues(t, source, target)

	// Define second source
	source2 := &structMapperTestStructArraySlice{
		A: []string{"0.0", "0.1"},
		B: []*structMapperTestStructInner{
			{
				A: "1.0",
			},
			{
				A: "",
			},
		},
		C: [2]string{"2.0", ""},
	}

	target2 := &structMapperTestStructArraySlice{}

	// Convert struct to map
	m, err = sm.ToMap(source2)
	require.NoError(t, err)
	require.NotNil(t, m)

	// Convert back to struct
	require.NoError(t, sm.ToStruct(m, target2))

	// Check if source and target are equal
	require.EqualValues(t, source2, target2)

}
