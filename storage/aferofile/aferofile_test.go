package aferofile_test

import (
	"github.com/spf13/afero"
	"testing"

	"github.com/anexia-it/go-structconf/storage/aferofile"
	"github.com/stretchr/testify/require"
)

func TestAferoFileStorage(t *testing.T) {
	fs := afero.NewMemMapFs()

	configPath := "config.txt"

	s := aferofile.NewAferoFileStorage(fs, configPath, 0640)
	require.NotNil(t, s)

	testContents := []byte("test contents\ntest second line")

	require.NoError(t, s.WriteConfig(testContents))

	// Check if contents were actually written...
	inBytes, err := afero.ReadFile(fs, configPath)
	require.NoError(t, err)
	require.NotNil(t, inBytes)
	require.EqualValues(t, testContents, inBytes)

	testContents = []byte("secondary test contents")
	require.NoError(t, afero.WriteFile(fs, configPath, testContents, 0640))

	// Read the file using our storage
	inBytes, err = s.ReadConfig()
	require.NoError(t, err)
	require.EqualValues(t, testContents, inBytes)
}
