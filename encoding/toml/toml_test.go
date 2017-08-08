package toml_test

import (
	"strings"
	"testing"

	"github.com/anexia-it/go-structconf/encoding/toml"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTOMLEncoding_Init(t *testing.T) {
	enc, err := toml.NewTOMLEncoding()
	require.NoError(t, err)
	require.NotNil(t, enc)
}

func TestTOMLEncoding_MarshalFrom(t *testing.T) {
	enc, err := toml.NewTOMLEncoding()
	require.NoError(t, err)

	// Source map
	source := map[string]interface{}{
		"a": 5,
		"b": "test b",
		"c": 1024,
	}

	// Expected TOML string
	expected := `a = 5
b = "test b"
c = 1024`

	encoded, err := enc.MarshalFrom(source)
	require.NoError(t, err)
	require.NotNil(t, encoded)
	require.EqualValues(t, expected, strings.TrimSuffix(string(encoded), "\n"))

	// Source map
	source = map[string]interface{}{
		"a": 5,
		"b": "test b",
		"c": map[interface{}]interface{}{
			"d": "test d",
		},
	}

	// Expected TOML string
	expected = `a = 5
b = "test b"

[c]
  d = "test d"`

	encoded, err = enc.MarshalFrom(source)
	require.NoError(t, err)
	require.NotNil(t, encoded)
	require.EqualValues(t, expected, strings.TrimSuffix(string(encoded), "\n"))

	source = map[string]interface{}{
		"title": "TOML Example",
		"owner": map[string]interface{}{
			"name":         "Tom Preston-Werner",
			"organization": "GitHub",
			"bio":          "GitHub Cofounder & CEO\nLikes tater tots and beer.",
		},
		"database": map[string]interface{}{
			"server": "192.168.1.1",
			"ports": []interface{}{
				8001, 8001, 8002,
			},
			"connection_max": 5000,
			"enabled":        true,
		},
		"servers": map[string]interface{}{
			"alpha": map[string]interface{}{
				"ip": "10.0.0.1",
				"dc": "eqdc10",
			},
			"beta": map[string]interface{}{
				"ip": "10.0.0.2",
				"dc": "eqdc10",
			},
		},
		"clients": map[string]interface{}{
			"data": []interface{}{
				[]interface{}{
					"gamma", "delta",
				},
				[]interface{}{
					1, 2,
				},
			},
		},
		"host": []interface{}{
			"alpha",
			"omega",
		},
	}

	expected = `title = "TOML Example"
[owner]
name = "Tom Preston-Werner"
organization = "GitHub"
bio = "GitHub Cofounder & CEO\nLikes tater tots and beer."

[database]
server = "192.168.1.1"
ports = [8001, 8001, 8002]
connection_max = 5000
enabled = true

[servers]

[servers.alpha]
ip = "10.0.0.1"
dc = "eqdc10"

[servers.beta]
ip = "10.0.0.2"
dc = "eqdc10"

[clients]
data = [["gamma", "delta"], [1, 2]]

host = ["alpha", "omega"]
`
	encoded, err = enc.MarshalFrom(source)
	//spew.Dump(expected)
	spew.Dump(string(encoded))
	require.NoError(t, err)
	require.NotNil(t, encoded)

	// Testing MarshalFrom by checking each line from the encoded output against an map of expected lines.
	// Due to undeterministic ordering (and strange string parsing behavior) maps cannot be compared directly
	encodedMap := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSuffix(string(encoded), "\n"), "\n") {
		if line == "" {
			continue
		}
		line = strings.TrimSpace(line)
		encodedMap[line] = true
	}

	for _, line := range strings.Split(expected, "\n") {
		if line == "" {
			continue
		}
		line = strings.TrimSpace(line)

		_, ok := encodedMap[line]
		assert.True(t, ok, "Line '%s' missing from encoded", line)
	}

	//	require.EqualValues(t, encodedMap, encodedMap)

}

func TestTOMLEncoding_UnmarshalTo(t *testing.T) {
	enc, err := toml.NewTOMLEncoding()
	require.NoError(t, err)

	// Source TOML string
	source := `a = 5
b = "test b"
c = 1024`

	// Expected map
	expected := map[string]interface{}{
		"a": int64(5),
		"b": "test b",
		"c": int64(1024),
	}

	target := make(map[string]interface{})
	require.NoError(t, enc.UnmarshalTo([]byte(source), target))
	require.EqualValues(t, expected, target)

}
