package source

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadFileBytesToMap(root string) (map[string][]byte, error) {

	fileMap := make(map[string][]byte)
	slog.Debug("loading file data to map", "path", root)
	err := filepath.WalkDir(root, func(currentPath string, dirEntry os.DirEntry, err error) error {
		if dirEntry == nil || dirEntry.IsDir() {
			slog.Debug("not a file that can be loaded into map", "dir", dirEntry, "path", root, "currentPath", currentPath)
			return nil
		}
		var fileData []byte
		if fileData, err = LoadFileBytes(currentPath); err == nil {
			fileMap[dirEntry.Name()] = fileData
		}
		return err
	})
	return fileMap, err
}

func LoadFileBytes(filePathToLoad string) ([]byte, error) {

	ext := strings.ToLower(filepath.Ext(filePathToLoad))

	var fileData []byte
	var err error
	switch ext {
	case ".md", ".markdown", ".html", ".htm", ".txt":
		fileData, err = os.ReadFile(filePathToLoad)
	default:
		slog.Info("unsupported file extension", "ext", filePathToLoad)
	}
	return fileData, err
}

// TODO: move
func SplitFileContent(content []byte, token []byte) (ContentMetadata, []byte, error) {

	var fm ContentMetadata
	if len(token) < 3 {
		return fm, content, fmt.Errorf("invalid frontmatter token: minimum length 3 required")
	}

	content = bytes.TrimSpace(content)
	if !bytes.HasPrefix(content, token) {
		return fm, content, fmt.Errorf("content missing starting frontmatter token")
	}
	start := len(token)
	end := bytes.Index(content[start:], token)
	if end == -1 {
		return fm, content, fmt.Errorf("closing frontmatter token not found")
	}

	fmData := content[start : start+end]
	body := content[start+end+len(token):]
	if err := yaml.Unmarshal(fmData, &fm); err != nil {
		return fm, body, fmt.Errorf("yaml unmarshal error: %w", err)
	}
	return fm, bytes.TrimSpace(body), nil
}
