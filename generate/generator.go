package generate

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
)

const (
	_defaultFilePerm   = 0755
	_postsMetadataFile = "data/posts.json"
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

	slog.Info("generating site", "size", len(metadata.SiteContentEntities))
	if _, err := g.store.Mkdir(_defaultFilePerm, g.ctx.SiteRoot, g.ctx.PostOutputDir); err != nil {
		return fmt.Errorf("failed to create content output dir: %w", err)
	}

	for _, ce := range metadata.SiteContentEntities {

		fileData, err := storage.LoadSiteFile(ce.InputPath, g.store)
		if err != nil {
			slog.Warn("unable to create file, relative path is invalid", "path", ce.RelativePath, "err", err)
			continue
		}

		_, err = g.store.Mkdir(_defaultFilePerm, g.ctx.SiteRoot, ce.RelativePath)
		if err != nil {
			return fmt.Errorf("unable to make new dir for content: %w", err)
		}
		content, err := g.converter.ToContent(fileData.Data, ce, metadata.ContentManifest)
		if err != nil {
			return fmt.Errorf("unable to convert to content: %w", err)
		}

		// write content to site
		outPath := filepath.Join(g.ctx.SiteRoot, ce.OutputPath)
		slog.Debug("Writing content", "path", outPath)
		err = g.store.Write(outPath, content, _defaultFilePerm)
		if err != nil {
			return fmt.Errorf("failed to write content to file: %w", err)
		}
	}

	if g.ctx.MakePostMetadata && len(metadata.ContentManifest) > 0 {
		g.createMetadataFile(metadata.ContentManifest[g.ctx.DefaultType])
	}
	if g.ctx.MakeSitemapXML && len(metadata.SiteMapUrlMetadata) > 0 {
		g.createSitemapFile(metadata.SiteMapUrlMetadata)
	}

	return nil
}

func (g *generator) createSitemapFile(urls []source.SitemapUrl) {

	siteMapUrlSet := Urlset{
		Xmlns:   _sitemapXmlMeta,
		XMLName: xml.Name{Local: "xmlns"},
		Urls:    urls,
	}
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	err := xml.NewEncoder(&buf).Encode(siteMapUrlSet)
	if err != nil {
		slog.Error("failed to encode site xml", "error", err)
		return
	}

	siteXmlMapPath := filepath.Join(g.ctx.SiteRoot, _sitemapXmlFile)
	if err := g.store.Write(siteXmlMapPath, buf.Bytes(), _defaultFilePerm); err != nil {
		slog.Error("failed to write site xml", "error", err)
	}
}

func (g *generator) createMetadataFile(contentMetadata []*source.ContentMetadata) {

	// TODO: handle err appropriately
	g.store.Mkdir(_defaultFilePerm, g.ctx.SiteRoot, "data")
	postMetadataPath := filepath.Join(g.ctx.SiteRoot, _postsMetadataFile)

	data, err := json.Marshal(contentMetadata)
	if err != nil {
		slog.Error("failed to Marshal metadata file", "file", _postsMetadataFile, "error", err)
		return
	}
	if err := g.store.Write(postMetadataPath, data, _defaultFilePerm); err != nil {
		slog.Error("failed to write post json metadata to file", "path", postMetadataPath, "error", err)
	}
}
