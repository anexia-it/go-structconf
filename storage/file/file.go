// Package file provides file-based configuration storage for go-structconf
package file

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/anexia-it/go-structconf/storage"
)

var _ storage.Storage = (*fileStorage)(nil)

// file-based storage implementation
type fileStorage struct {
	path  string
	mode  os.FileMode
	mutex sync.Mutex
}

func (fs *fileStorage) WriteConfig(data []byte) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	return ioutil.WriteFile(fs.path, data, fs.mode)
}

func (fs *fileStorage) ReadConfig() ([]byte, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	return ioutil.ReadFile(fs.path)
}

// NewFileStorage initializes a new file-based configuration storage
func NewFileStorage(path string, mode os.FileMode) storage.Storage {
	return &fileStorage{
		path: path,
		mode: mode,
	}
}
