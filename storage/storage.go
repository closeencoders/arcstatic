package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var errTraversalUnsupported = errors.New("traversal (..) is not allowed")

type Storage interface {
	fs.FS
	Write(name string, data []byte, perm int) error
	Mkdir(fileMode int, path ...string) (string, error)
}

type osFileStorage struct{}

var _ Storage = osFileStorage{}

func NewOSFileStorage() Storage {
	return osFileStorage{}
}

func (osFileStorage) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (osFileStorage) Write(name string, data []byte, perm int) error {
	if err := os.WriteFile(name, data, os.FileMode(perm)); err != nil {
		return fmt.Errorf("failed to write file %s perm %d: %w", name, perm, err)
	}
	return nil
}

func (osFileStorage) Mkdir(perm int, path ...string) (string, error) {
	loc := filepath.Join(path...)
	if strings.HasPrefix(loc, "..") {
		return loc, fmt.Errorf("directory not created: %w", errTraversalUnsupported)
	}
	err := os.MkdirAll(loc, os.FileMode(perm))
	if err != nil {
		return loc, err
	}
	return loc, err
}
