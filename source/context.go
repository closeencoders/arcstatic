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

func LoadSiteContext(path string, store fs.FS) (config.SiteContext, error) {

	ctx, err := loadSiteCtx(path, store)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return config.SiteContext{}, fmt.Errorf("failed to load site config: %w", err)
		}
		slog.Warn("no valid configuration file, using defaults", "path", path, "err", err)
		ctx = newDefaultContext(path)
	}

	// TODO: Embedded defaults/themes with correct error handling
	componentsPath := filepath.Join(path, _componentsLoc)
	componentsMap, err := storage.LoadFilesToMap(componentsPath, store)
	if err != nil {
		slog.Warn("components not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", componentsPath)
	}
	templatesPath := filepath.Join(path, _templatesLoc)
	templatesMap, err := storage.LoadFilesToMap(templatesPath, store)
	if err != nil {
		slog.Warn("templates not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", componentsPath)
	}

	ctx.ComponentMap = componentsMap
	ctx.TemplateMap = templatesMap
	return ctx, nil
}

func loadSiteCtx(path string, store fs.FS) (config.SiteContext, error) {

	var configPath string = path
	if !strings.HasSuffix(path, _configName) {
		configPath = filepath.Join(path, _configName)
	}

	fileData, err := storage.LoadSiteFile(configPath, store)
	if err != nil {
		return config.SiteContext{}, err
	}

	ctx := newDefaultContext(path)
	if err := yaml.Unmarshal(fileData.Data, &ctx); err != nil {
		return config.SiteContext{}, fmt.Errorf("failed to unmarshal site config: %w", err)
	}

	return ctx, nil
}

func newDefaultContext(location string) config.SiteContext {
	return config.SiteContext{
		SiteInputRoot:        location,
		SiteRoot:             location,
		Base:                 "/",
		SiteUrl:              _defaultUrl,
		PostInputDir:         filepath.Join(location, _postsLoc),
		PostOutDir:           "/",
		PageInputDir:         filepath.Join(location, _pagesLoc),
		FrontmatterToken:     []byte("---"),
		FullHtmlPath:         false,
		GenerateSitemapXml:   true,
		GeneratePostMetadata: true,
		MakeTableOfContents:  false,
	}
}
