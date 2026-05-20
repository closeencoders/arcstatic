package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

var errUnsupported = errors.New("unsupported file extension")

type Storage interface {
	fs.FS
	Write(name string, data []byte, perm int) error
}

type FileData struct {
	Name      string
	Extension string
	Data      []byte
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
