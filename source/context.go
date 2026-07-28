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
	_defaultUrl    = "http://yourdomain.com/#"
	ConfigName     = "arcconfig.yml"
)

var (
	errInvalidConfig = errors.New("invalid site config file format")
)

func LoadSiteContext(inPath string, outPath string, store fs.FS) (*config.SiteContext, error) {

	ctx, err := createSiteContext(inPath, outPath, store)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			slog.Warn("no valid configuration file found, using defaults", "path", inPath, "err", err)
			ctx = CreateDefaultContext(inPath)
		} else {
			return nil, fmt.Errorf("%w: %w", errInvalidConfig, err)
		}
	}

	// TODO: Embedded defaults/themes
	componentsPath := filepath.Join(inPath, _componentsLoc)
	componentsMap, err := storage.LoadFilesToMap(componentsPath, store)
	if err != nil {
		slog.Warn("components not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", componentsPath)
	}
	templatesPath := filepath.Join(inPath, _templatesLoc)
	templatesMap, err := storage.LoadFilesToMap(templatesPath, store)
	if err != nil {
		slog.Warn("templates not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", templatesPath)
	}

	ctx.ComponentMap = componentsMap
	ctx.TemplateMap = templatesMap
	return ctx, nil
}

func createSiteContext(inPath string, outPath string, store fs.FS) (*config.SiteContext, error) {

	var configPath string = inPath
	if !strings.HasSuffix(inPath, ConfigName) {
		configPath = filepath.Join(inPath, ConfigName)
	}

	fileData, err := storage.LoadSiteFile(configPath, store)
	if err != nil {
		return nil, err
	}

	// Create defaults which should be the minimal operational required data, then apply config file overrides.
	ctx := CreateDefaultContext(inPath)
	if err := yaml.Unmarshal(fileData.Data, ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Override config with passed in if available
	outPath = strings.TrimSpace(outPath)
	if outPath != "" && outPath != inPath {
		ctx.SiteOutputRoot = outPath
	}

	return ctx, nil
}

func CreateDefaultContext(root string) *config.SiteContext {
	ctx := config.NewContext(root)
	ctx.SiteURL = _defaultUrl
	ctx.PostInputDir = filepath.Join(root, _postsLoc)
	ctx.PageInputDir = filepath.Join(root, _pagesLoc)
	return ctx
}
