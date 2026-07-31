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

func LoadSiteContext(inputPath string, outputPath string, store fs.FS) (*config.SiteContext, error) {

	ctx, err := createSiteContext(inputPath, outputPath, store)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			slog.Warn("no valid configuration file found, using defaults", "path", inputPath, "err", err)
			ctx = CreateDefaultContext(inputPath)
		} else {
			return nil, fmt.Errorf("%w: %w", errInvalidConfig, err)
		}
	}

	// Override config with passed in if available and not in config
	if outputPath != "" {
		slog.Debug("setting output override", "path", outputPath)
		ctx.SiteOutputRoot = outputPath
	} else if strings.TrimSpace(ctx.SiteOutputRoot) == "" {
		slog.Debug("SiteOutputRoot not set in configuration", "path", inputPath)
		ctx.SiteOutputRoot = inputPath
	}

	// TODO: Embedded defaults/themes
	componentsPath := filepath.Join(inputPath, _componentsLoc)
	componentsMap, err := storage.LoadFilesToMap(componentsPath, store)
	if err != nil {
		slog.Warn("components not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", componentsPath)
	}
	templatesPath := filepath.Join(inputPath, _templatesLoc)
	templatesMap, err := storage.LoadFilesToMap(templatesPath, store)
	if err != nil {
		slog.Warn("templates not loaded, embedded defaults/themes are not implemented yet, content may not be rendered properly", "path", templatesPath)
	}

	ctx.ComponentMap = componentsMap
	ctx.TemplateMap = templatesMap
	return ctx, nil
}

// TODO: This is still being overridden by the config file in some cases for now.
func createSiteContext(inputPath string, outputPath string, store fs.FS) (*config.SiteContext, error) {

	var configPath string = inputPath
	if !strings.HasSuffix(inputPath, ConfigName) {
		configPath = filepath.Join(inputPath, ConfigName)
	}

	configPath = strings.TrimSuffix(configPath, "/")
	fileData, err := storage.LoadSiteFile(configPath, store)
	if err != nil {
		return nil, err
	}

	// Create defaults which should be the minimal operational required data, then apply config file overrides.
	ctx := CreateDefaultContext(inputPath)
	if err := yaml.Unmarshal(fileData.Data, ctx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
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
