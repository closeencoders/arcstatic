package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
)

var (
	errUnsupported = errors.New("unsupported file extension")

	supportedContentTypes  = map[string]struct{}{".md": {}, ".markdown": {}, ".html": {}, ".htm": {}}
	supportedDataFileTypes = map[string]struct{}{".json": {}, ".xml": {}, ".yml": {}, ".yaml": {}}
)

type FileData struct {
	Name      string
	Extension string
	Data      []byte
}

func SupportedContentFile(path string) bool {
	return hasExtension(path, supportedContentTypes)
}

func SupportedFile(path string) bool {
	return hasExtension(path, supportedDataFileTypes, supportedContentTypes)
}

func hasExtension(path string, targetMaps ...map[string]struct{}) bool {

	if len(path) < 3 {
		slog.Debug("unsupported path length, must go to content file")
		return false
	}

	ext := filepath.Ext(path)
	if ext == "" || ext == "." || len(targetMaps) == 0 {
		return false
	}
	for _, t := range targetMaps {
		if _, exists := t[strings.ToLower(ext)]; exists {
			return true
		}
	}
	return false
}

func LoadFilesToMap(root string, fsys fs.FS) (map[string][]byte, error) {

	fileMap := make(map[string][]byte)
	slog.Debug("loading file data to map", "path", root)

	err := fs.WalkDir(fsys, root, func(path string, entry fs.DirEntry, err error) error {

		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		fileData, err := LoadSiteFile(path, fsys)
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

func LoadSiteFile(path string, fsys fs.FS) (FileData, error) {

	if !SupportedFile(path) {
		slog.Debug("unsupported file extension", "path", path)
		return FileData{}, errUnsupported
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return FileData{}, fmt.Errorf("failed to load site file %s: %w", path, err)
	}
	return FileData{Extension: filepath.Ext(path), Data: data, Name: filepath.Base(path)}, nil
}
