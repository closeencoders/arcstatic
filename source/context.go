package source

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/storage"
	"gopkg.in/yaml.v3"
)

const (
	_postsLoc      = "_posts"
	_pagesLoc      = "_pages"
	_componentsLoc = "_components"
	_templatesLoc  = "_templates"
	_configName    = "config.yml"
	_defaultUrl    = "http://yourdomain.com/#"
)

func LoadSiteContext(path string, store fs.FS) (*config.SiteContext, error) {

	ctx, err := createSiteContext(path, store)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("invalid config: %w", err)
		}
		slog.Warn("no valid configuration file, using defaults", "path", path, "err", err)
		ctx = createDefaultContext(path)
	}

	// TODO: Embedded defaults/themes
	componentsPath := filepath.Join(path, _componentsLoc)
	componentsMap, err := storage.LoadFilesToMap(componentsPath, store)
	if err != nil {
		slog.Warn("components not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", componentsPath)
	}
	templatesPath := filepath.Join(path, _templatesLoc)
	templatesMap, err := storage.LoadFilesToMap(templatesPath, store)
	if err != nil {
		slog.Warn("templates not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", templatesPath)
	}

	ctx.ComponentMap = componentsMap
	ctx.TemplateMap = templatesMap
	return ctx, nil
}

func createSiteContext(path string, store fs.FS) (*config.SiteContext, error) {

	var configPath string = path
	if !strings.HasSuffix(path, _configName) {
		configPath = filepath.Join(path, _configName)
	}

	fileData, err := storage.LoadSiteFile(configPath, store)
	if err != nil {
		return nil, err
	}

	// Create defaults which should be the minimal operational required data, then apply config file overrides.
	ctx := createDefaultContext(path)
	if err := yaml.Unmarshal(fileData.Data, ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal site config file: %w", err)
	}

	return ctx, nil
}

func createDefaultContext(path string) *config.SiteContext {
	ctx := config.NewContext(path)
	ctx.SiteURL = _defaultUrl
	ctx.PostInputDir = filepath.Join(path, _postsLoc)
	ctx.PageInputDir = filepath.Join(path, _pagesLoc)
	return ctx
}
