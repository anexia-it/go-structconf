// Package aferofile provides file-based configuration storage for go-structconf accessed through an afero.Fs
package aferofile

import (
	"github.com/spf13/afero"
	"os"
	"sync"

	"github.com/anexia-it/go-structconf/storage"
)

var _ storage.Storage = (*aferoFileStorage)(nil)

// file-based storage implementation with afero
type aferoFileStorage struct {
	fs    afero.Fs
	path  string
	mode  os.FileMode
	mutex sync.Mutex
}

func (s *aferoFileStorage) WriteConfig(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return afero.WriteFile(s.fs, s.path, data, s.mode)
}

func (s *aferoFileStorage) ReadConfig() ([]byte, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return afero.ReadFile(s.fs, s.path)
}

// NewAferoFileStorage initializes a new file-based configuration storage accessed through an afero.Fs
func NewAferoFileStorage(fs afero.Fs, path string, mode os.FileMode) storage.Storage {
	return &aferoFileStorage{
		fs:   fs,
		path: path,
		mode: mode,
	}
}
