package generate

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
)

const (
	_defaultFilePerm   = 0755
	_postsMetadataFile = "posts.json"
	_sitemapXmlFile    = "sitemap.xml"
	_sitemapXmlMeta    = "http://www.sitemaps.org/schemas/sitemap/0.9"
)

type Urlset struct {
	XMLName xml.Name            `xml:"urlset"`
	Xmlns   string              `xml:"xmlns,attr"`
	Urls    []source.SitemapUrl `xml:"url"`
}

type generator struct {
	ctx       *config.SiteContext
	converter converter
	store     storage.Storage
}

func NewGenerator(ctx *config.SiteContext, converter converter, store storage.Storage) *generator {
	return &generator{ctx: ctx, converter: converter, store: store}
}

func (g *generator) Generate(metadata *source.SiteMetadata) error {

	baseOutPath := g.ctx.SiteRoot
	if g.ctx.SiteOutputRoot != "" {
		baseOutPath = g.ctx.SiteOutputRoot
	}

	slog.Info("generating site", "size", len(metadata.SiteContentEntities))
	if err := g.store.Mkdir(_defaultFilePerm, baseOutPath, g.ctx.PostOutputDir); err != nil {
		return fmt.Errorf("failed to create content output dir: %w", err)
	}

	for _, ce := range metadata.SiteContentEntities {
		err := g.makeContent(ce, metadata, baseOutPath)
		if err != nil {
			return err
		}
	}

	if g.ctx.MakePostMetadata && len(metadata.ContentManifest) > 0 {
		g.createMetadataFile(baseOutPath, metadata.ContentManifest[g.ctx.DefaultType])
	}
	if g.ctx.MakeSitemapXML && len(metadata.SiteMapUrlMetadata) > 0 {
		g.createSitemapFile(baseOutPath, metadata.SiteMapUrlMetadata)
	}
	if g.shouldCopyAssets() {
		g.copyAssets(g.ctx.AssetsInputDir)
	}

	return nil
}

func (g *generator) shouldCopyAssets() bool {

	if g.ctx.SiteOutputRoot != "" && g.ctx.SiteRoot != "" {
		// The asset input location has been overridden to a valid location outside of site root.
		if g.ctx.AssetsInputDir != "" && g.ctx.AssetsInputDir != g.ctx.SiteOutputRoot && g.ctx.AssetsInputDir != filepath.Join(g.ctx.SiteOutputRoot, "assets") {
			return true
		}
		// The root export location for the site is not in the same location at all.
		if g.ctx.SiteOutputRoot != g.ctx.SiteRoot {
			return true
		}
	}

	return false
}

func (g *generator) makeContent(ce *source.ContentEntity, metadata *source.SiteMetadata, baseOutPath string) error {
	fileData, err := storage.LoadSiteFile(ce.InputPath, g.store)
	if err != nil {
		slog.Warn("unable to create file, relative path is invalid", "path", ce.RelativePath, "err", err)
		return nil
	}

	relativePath := filepath.Join(baseOutPath, ce.RelativePath)
	writePath := filepath.Join(baseOutPath, ce.OutputPath)

	err = g.store.Mkdir(_defaultFilePerm, relativePath)
	if err != nil {
		return fmt.Errorf("unable to make new dir for content: %w", err)
	}

	content, err := g.converter.ToContent(fileData.Data, ce, metadata.ContentManifest)
	if err != nil {
		return fmt.Errorf("unable to convert to content: %w", err)
	}
	slog.Debug("Writing content", "path", writePath)
	err = g.store.Write(writePath, content, _defaultFilePerm)
	if err != nil {
		return fmt.Errorf("failed to write content to file: %w", err)
	}

	return nil
}

func (g *generator) copyAssets(assetsLoc string) error {

	slog.Debug("copying assets", "from", assetsLoc, "to", g.ctx.SiteOutputRoot)
	entries, err := os.ReadDir(assetsLoc)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "_") || strings.HasPrefix(e.Name(), ".git") || strings.HasPrefix(e.Name(), ".env") {
			continue
		}
		in := filepath.Clean(filepath.Join(assetsLoc, e.Name()))
		out := filepath.Clean(filepath.Join(g.ctx.SiteOutputRoot, "assets", e.Name()))
		if e.IsDir() || e.Name() != ".git" {
			err := g.store.CopyDir(_defaultFilePerm, in, out)
			if err != nil {
				slog.Error("Failed to copy dir to assets export location", "error", err)
			}
		} else {
			err := g.store.Copy(_defaultFilePerm, in, out)
			if err != nil {
				slog.Error("Failed to copy file to assets export location", "error", err)
			}
		}
	}

	return nil
}

func (g *generator) createSitemapFile(basePath string, urls []source.SitemapUrl) {

	slog.Debug("creating sitemap file", "path", basePath)
	siteMapUrlSet := Urlset{
		Xmlns:   _sitemapXmlMeta,
		XMLName: xml.Name{Local: "xmlns"},
		Urls:    urls,
	}
	var buf bytes.Buffer
	if _, err := buf.WriteString(xml.Header); err != nil {
		slog.Error("failed to write site xml header to buffer", "error", err)
		return
	}
	if err := xml.NewEncoder(&buf).Encode(siteMapUrlSet); err != nil {
		slog.Error("failed to encode site xml", "error", err)
		return
	}
	siteXmlMapPath := filepath.Join(basePath, _sitemapXmlFile)
	if err := g.store.Write(siteXmlMapPath, buf.Bytes(), _defaultFilePerm); err != nil {
		slog.Error("failed to write site xml", "error", err)
		return
	}
}

func (g *generator) createMetadataFile(basePath string, contentMetadata []*source.ContentMetadata) {

	slog.Debug("creating metadata file", "path", basePath)
	postMetadataPath := filepath.Join(basePath, "data")
	if strings.TrimSpace(g.ctx.PostMetadataDir) != "" {
		postMetadataPath = filepath.Join(basePath, g.ctx.PostMetadataDir)
	}
	err := g.store.Mkdir(_defaultFilePerm, postMetadataPath)
	if err != nil {
		slog.Error("failed to make metadata file location", "dir", postMetadataPath, "error", err)
		return
	}
	data, err := json.Marshal(contentMetadata)
	if err != nil {
		slog.Error("failed to Marshal metadata file", "file", _postsMetadataFile, "error", err)
		return
	}
	postMetaFile := filepath.Join(postMetadataPath, _postsMetadataFile)
	if err := g.store.Write(postMetaFile, data, _defaultFilePerm); err != nil {
		slog.Error("failed to write post json metadata to file", "path", postMetadataPath, "error", err)
	}
}
