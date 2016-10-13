package file_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/anexia-it/go-structconf/storage/file"
	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "go-structconf-test-")
	require.NoError(t, err)
	require.NotNil(t, tmpFile)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	s := file.NewFileStorage(tmpFile.Name(), 0640)
	require.NotNil(t, s)

	testContents := []byte("test contents\ntest second line")

	require.NoError(t, s.WriteConfig([]byte(testContents)))

	// Check if contents were actually written...
	inBytes, err := ioutil.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	require.NotNil(t, inBytes)
	require.EqualValues(t, testContents, inBytes)

	// Write other contents to file using ioutil.
	// This ensures that the file storage does not keep those values in-memory and actually re-reads
	// them from disk
	testContents = []byte("secondary test contents")
	require.NoError(t, ioutil.WriteFile(tmpFile.Name(), testContents, 0640))

	// Read the file using our storage
	inBytes, err = s.ReadConfig()
	require.NoError(t, err)
	require.EqualValues(t, testContents, inBytes)
}
