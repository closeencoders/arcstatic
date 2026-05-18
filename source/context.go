package source

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/closeencoders/arcstatic/config"
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

func LoadSiteCtx(location string) (config.SiteContext, error) {

	confFileLoc := filepath.Join(location, _configName)
	var ctx config.SiteContext

	_, err := os.Stat(confFileLoc)
	if err != nil {
		slog.Warn("no valid configuration file, using defaults", "err", err)
		ctx = newDefaultContext(location)

	} else {
		configFile, err := os.Open(confFileLoc)
		if err != nil {
			return ctx, err
		}
		defer configFile.Close()

		fileSiteCtx, err := loadSiteCtx(location, configFile)
		if err != nil {
			return ctx, nil
		}
		ctx = fileSiteCtx
	}

	componentPath := filepath.Join(location, _componentsLoc)
	componentsMap, err := LoadFileBytesToMap(componentPath)
	if err != nil {
		return ctx, fmt.Errorf("failed to load components: %w", err)
	}

	templatesPath := filepath.Join(location, _templatesLoc)
	templatesMap, err := LoadFileBytesToMap(templatesPath)
	if err != nil {
		return ctx, fmt.Errorf("failed to load templates: %w", err)
	}

	ctx.ComponentMap = componentsMap
	ctx.TemplateMap = templatesMap

	return ctx, nil
}

func loadSiteCtx(location string, r io.Reader) (config.SiteContext, error) {

	var ctx config.SiteContext
	configData, err := io.ReadAll(r)
	if err != nil {
		return ctx, err
	}

	ctx = newDefaultContext(location)
	err = yaml.Unmarshal(configData, &ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to load site context: %w", err)
	}

	return ctx, nil
}

func newDefaultContext(location string) config.SiteContext {
	return config.SiteContext{
		// TODO: the context can be different from the site root which is what is generated
		SiteInputRoot: location,
		SiteRoot:      location,
		Base:          "/",
		SiteUrl:       _defaultUrl,

		PostInputDir: filepath.Join(location, _postsLoc),
		PostOutDir:   "/",

		PageInputDir: filepath.Join(location, _pagesLoc),

		FrontmatterToken: []byte("---"),

		FullHtmlPath:         false,
		GenerateSitemapXml:   true,
		GeneratePostMetadata: true,
		MakeTableOfContents:  false,
	}
}
