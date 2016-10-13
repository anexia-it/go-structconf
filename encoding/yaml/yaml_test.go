package yaml_test

import (
	"strings"
	"testing"

	"github.com/anexia-it/go-structconf/encoding/yaml"
	"github.com/stretchr/testify/require"
)

func TestNewYAMLEncoding_Init(t *testing.T) {
	enc, err := yaml.NewYAMLEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)
}

func TestYAMLEncoding_MarshalFrom(t *testing.T) {
	enc, err := yaml.NewYAMLEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	// Source map
	source := map[string]interface{}{
		"a": 5,
		"b": "test b",
		"c": 1024,
	}

	// Expected YAML string
	expected := `a: 5
b: test b
c: 1024`

	encoded, err := enc.MarshalFrom(source)
	require.NoError(t, err)
	require.NotNil(t, encoded)
	require.EqualValues(t, expected, strings.TrimSuffix(string(encoded), "\n"))
}

func TestYAMLEncoding_UnmarshalTo(t *testing.T) {
	enc, err := yaml.NewYAMLEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)

	// Source YAML string
	source := `a: 5
b: test b
c: 1024`

	// Expected map
	expected := map[string]interface{}{
		"a": 5,
		"b": "test b",
		"c": 1024,
	}

	target := make(map[string]interface{})
	require.NoError(t, enc.UnmarshalTo([]byte(source), target))
	require.EqualValues(t, expected, target)
}
