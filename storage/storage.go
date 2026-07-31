package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Storage interface {
	fs.FS
	Write(name string, data []byte, perm int) error
	Mkdir(fileMode int, path ...string) error
	CopyDir(perm int, from string, to string) error
	Copy(perm int, from string, to string) error
	GetWd() (string, error)
}

type osFileStorage struct{}

var _ Storage = osFileStorage{}

func NewOSFileStorage() Storage {
	return osFileStorage{}
}

func (f osFileStorage) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (f osFileStorage) Write(name string, data []byte, perm int) error {
	if err := os.WriteFile(name, data, os.FileMode(perm)); err != nil {
		return fmt.Errorf("failed to write file %s perm %d: %w", name, perm, err)
	}
	return nil
}

func (f osFileStorage) Mkdir(perm int, path ...string) error {
	loc := filepath.Join(path...)
	if strings.HasPrefix(loc, "..") {
		return fmt.Errorf("traversal (..) is not allowed %s", loc)
	}
	err := os.MkdirAll(loc, os.FileMode(perm))
	if err != nil {
		return err
	}
	return nil
}

func (f osFileStorage) CopyDir(perm int, from string, to string) error {
	from = filepath.Clean(from)
	to = filepath.Clean(to)

	return filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %q: %w", path, err)
		}

		relPath, err := filepath.Rel(from, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path for %q: %w", path, err)
		}

		dstPath := filepath.Join(to, relPath)
		if info.IsDir() {
			if err := f.Mkdir(perm, dstPath); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", dstPath, err)
			}
			return nil
		}

		if err := f.Copy(perm, path, dstPath); err != nil {
			return fmt.Errorf("failed to copy file from %q to %q: %w", path, dstPath, err)
		}

		return nil
	})
}

func (f osFileStorage) Copy(perm int, from string, to string) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("failed to open source copy file %q: %w", from, err)
	}
	defer fromFile.Close()

	toFile, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(perm))
	if err != nil {
		return fmt.Errorf("failed to open/create destination for copy file %q: %w", to, err)
	}

	defer func() {
		if closeErr := toFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close destination for copy file %q: %w", to, closeErr)
		}
	}()

	if _, err = io.Copy(toFile, fromFile); err != nil {
		return fmt.Errorf("failed to write contents for copy from %q to %q: %w", from, to, err)
	}

	return nil
}

func (f osFileStorage) GetWd() (string, error) {
	return os.Getwd()
}

func ToCleanAbs(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, ".") {
		return path, nil
	}
	path = filepath.Clean(path)
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return path, nil
}
