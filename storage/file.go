package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
)

func LoadFilesToMap(root string, store fs.FS) (map[string][]byte, error) {
	fileMap := make(map[string][]byte)
	slog.Debug("loading file data to map", "path", root)

	err := fs.WalkDir(store, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		fileData, err := LoadSiteFile(path, store)
		if errors.Is(err, errUnsupported) {
			return nil
		}
		if err != nil {
			return err
		}

		fileMap[entry.Name()] = fileData.Data
		return nil
	})

	return fileMap, err
}

func LoadSiteFile(path string, store fs.FS) (FileData, error) {

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	// TODO: This is dumb
	case ".md", ".markdown", ".html", ".htm", ".txt", ".yml":
		data, err := fs.ReadFile(store, path)
		if err != nil {
			return FileData{}, fmt.Errorf("failed to load site file %s: %w", path, err)
		}

		return FileData{Extension: ext, Data: data}, nil
	default:
		slog.Debug("unsupported file extension", "path", path)
		return FileData{}, errUnsupported
	}
}
