package generate

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
)

const (
	_defaultFilePerm   = 0755
	_defaultPostsItem  = "Posts"
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

func (g *generator) Generate(metadata source.SiteMetadata) error {

	slog.Info("generating site", "size", len(metadata.SiteContentEntities))
	if err := mkDirIfNotExists(_defaultFilePerm, g.ctx.SiteRoot, g.ctx.PostOutDir); err != nil {
		return fmt.Errorf("failed to create content output dir: %w", err)
	}

	for _, ce := range metadata.SiteContentEntities {

		fileData, err := storage.LoadSiteFile(ce.InPath, g.store)
		if err != nil {
			slog.Warn("unable to create file, relative path is invalid %s, %w", ce.RelativePath, err)
			continue
		}

		err = mkDirIfNotExists(_defaultFilePerm, g.ctx.SiteRoot, ce.RelativePath)
		if err != nil {
			return fmt.Errorf("unable to make new dir for content: %w", err)
		}
		content, err := g.converter.ConvertToContent(fileData.Data, ce, metadata.ContentManifest)
		if err != nil {
			return err
		}

		// write content to site
		outPath := filepath.Join(g.ctx.SiteRoot, ce.OutPath)
		slog.Debug("Writing content", "path", outPath)
		err = g.store.Write(outPath, content, _defaultFilePerm)
		if err != nil {
			return fmt.Errorf("failed to content to file: %w", err)
		}
	}

	if g.ctx.MakePostMetadata && len(metadata.ContentManifest) > 0 {
		g.createMetadataFile(metadata.ContentManifest[_defaultPostsItem])
	}
	if g.ctx.MakeSitemapXml && len(metadata.SiteMapUrlMetadata) > 0 {
		g.createSitemapFile(metadata.SiteMapUrlMetadata)
	}

	return nil
}

func mkDirIfNotExists(fileMode os.FileMode, path ...string) error {
	pathToCreate := filepath.Join(path...)
	_, err := os.Stat(pathToCreate)
	if err != nil && !os.IsExist(err) {
		return os.MkdirAll(pathToCreate, fileMode)
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

	siteXmlMapPath := filepath.Join(g.ctx.SiteInputRoot, _sitemapXmlFile)
	if err := os.WriteFile(siteXmlMapPath, buf.Bytes(), _defaultFilePerm); err != nil {
		slog.Error("failed to write site xml", "error", err)
	}
}

func (g *generator) createMetadataFile(siteMetadata []*source.ContentMetadata) {

	mkDirIfNotExists(_defaultFilePerm, g.ctx.SiteRoot, "data")
	postMetadataPath := filepath.Join(g.ctx.SiteRoot, _postsMetadataFile)

	data, err := json.Marshal(siteMetadata)
	if err != nil {
		slog.Error("failed to Marshal metadata file", "file", _postsMetadataFile, "error", err)
		return
	}
	if err := os.WriteFile(postMetadataPath, data, _defaultFilePerm); err != nil {
		slog.Error("failed to write post json metadata to file", "path", postMetadataPath, "error", err)
	}
}
