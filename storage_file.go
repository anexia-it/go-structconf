package structconf

import (
	"io/ioutil"
	"os"
	"sync"
)

var _ Storage = (*fileStorage)(nil)

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

// NewStorageFile initializes a new file-based configuration storage
func NewStorageFile(path string, mode os.FileMode) Storage {
	return &fileStorage{
		path: path,
		mode: mode,
	}
}
