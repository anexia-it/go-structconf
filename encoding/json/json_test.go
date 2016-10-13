package json_test

import (
	"testing"

	"strings"

	"github.com/anexia-it/go-structconf/encoding/json"
	"github.com/stretchr/testify/require"
)

func TestNewJSONEncoding_Init(t *testing.T) {
	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)
}

func TestJSONEncoding_MarshalFrom(t *testing.T) {
	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	// Source map
	source := map[string]interface{}{
		"a": 5,
		"b": "test b",
		"c": 1024,
	}

	// Expected JSON string
	expected := `{"a":5,"b":"test b","c":1024}`

	encoded, err := enc.MarshalFrom(source)
	require.NoError(t, err)
	require.NotNil(t, encoded)
	require.EqualValues(t, expected, strings.TrimSuffix(string(encoded), "\n"))
}

func TestJSONEncoding_UnmarshalTo(t *testing.T) {
	enc, err := json.NewJSONEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	// Source JSON string
	source := `{"a":5,"b":"test b","c":1024}`

	// Expected map
	expected := map[string]interface{}{
		"a": float64(5),
		"b": "test b",
		"c": float64(1024),
	}

	target := make(map[string]interface{})
	require.NoError(t, enc.UnmarshalTo([]byte(source), target))
	require.EqualValues(t, expected, target)
}
