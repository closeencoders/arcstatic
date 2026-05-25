package source

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/storage"
	"gopkg.in/yaml.v3"
)

const (
	_indexHtmlFile    = "index.html"
	_maxInputSize     = 1_000_000
	_YYYYMMDD_RFC3339 = "2006-01-02T15:04:05Z07:00"

	_defaultPostTemplate = "post.html"
	_defaultPageTemplate = "page.html"
)

type SitemapUrl struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type SiteMetadata struct {
	// content entity struct to hold all original content metadata loaded from source material until I find a better pattern
	SiteContentEntities []*ContentEntity
	// Represent the state of every categorized set of metadata
	ContentManifest Manifest
	// Used for creating xml representations for the site for search engines, like a sitemap.xml
	SiteMapUrlMetadata []SitemapUrl
}

type ContentEntity struct {
	Name string
	// Data extracted via a header or similar from a content file
	ContentMetadata ContentMetadata
	// Sub-path from root to the content. e.g. /posts/
	RelativePath string
	// Full path to file with rendered content
	OutPath string
	// Full path to original content to be rendered
	InputDir string
}

// TODO: This should be more dynamic
type ContentMetadata struct {
	Title       string `json:"title" yaml:"title"`
	Image       string `json:"image" yaml:"image"`
	Url         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`

	AltImage        string `json:"-" yaml:"alt_image"`
	MetaDescription string `json:"-" yaml:"meta_description"`
	TemplateId      string `json:"-" yaml:"template_id"`
	Permalink       string `json:"-" yaml:"permalink"`

	Date time.Time `json:"date" yaml:"date"`

	Draft bool `json:"-" yaml:"draft"`

	Type       string   `json:"type" yaml:"type"`
	Categories []string `json:"categories" yaml:"categories"`
	Tags       []string `json:"tags" yaml:"tags"`
}

type metadata struct {
	ctx   *config.SiteContext
	store storage.Storage
}

func NewMetadata(ctx *config.SiteContext, store storage.Storage) *metadata {
	return &metadata{ctx: ctx, store: store}
}

func (m *metadata) LoadMetadata(paths ...string) (*SiteMetadata, error) {

	var metadata SiteMetadata
	for _, loc := range paths {
		if err := m.readSiteMetadataFiles(loc, &metadata); err != nil {
			return &metadata, fmt.Errorf("failed to load metadata for %s: %w", loc, err)
		}
	}
	slices.SortFunc(metadata.SiteContentEntities, func(a, b *ContentEntity) int {
		return b.ContentMetadata.Date.Compare(a.ContentMetadata.Date)
	})

	metadata.ContentManifest = NewManifest(*m.ctx, metadata.SiteContentEntities)
	return &metadata, nil
}

func (m *metadata) readSiteMetadataFiles(root string, metadata *SiteMetadata) error {

	slog.Debug("rendering content", "path", root)
	return fs.WalkDir(m.store, root, func(path string, dirEntry os.DirEntry, err error) error {

		if dirEntry == nil || dirEntry.IsDir() {
			slog.Debug("not a file that can be used for metadata extraction", "dir", dirEntry, "path", root, "currentPath", path)
			return nil
		}
		fileData, err := storage.LoadSiteFile(path, m.store)
		if err != nil {
			return fmt.Errorf("failed to load file data: %w", err)
		}
		if len(fileData.Data) < 3 {
			slog.Debug("no file data viable for conversion found", "dir", dirEntry, "path", root)
			return nil
		}
		if !storage.SupportedContentFile(fileData.Extension) {
			slog.Debug("unsupported content file extension", "path", path)
			return nil
		}

		content, err := m.getContentMetadata(fileData.Data, dirEntry.Name(), root)
		if err != nil {
			return fmt.Errorf("failed to convert to content: %w", err)
		}
		content.InputDir = path

		if content.ContentMetadata.Draft {
			slog.Debug("is draft", "file", path)
			return nil
		}

		metadata.SiteContentEntities = append(metadata.SiteContentEntities, content)

		if m.ctx.MakeSitemapXML {
			xmlUrl := m.makeSitemapEntry(content)
			metadata.SiteMapUrlMetadata = append(metadata.SiteMapUrlMetadata, xmlUrl)
		}
		return err
	})
}

func (m *metadata) makeSitemapEntry(ce *ContentEntity) SitemapUrl {
	// TODO:
	siteUrl, _ := url.Parse(m.ctx.SiteURL)
	if m.ctx.FullHtmlPaths {
		siteUrl.Path = path.Join(siteUrl.Path, ce.OutPath)
	} else {
		siteUrl.Path = path.Join(siteUrl.Path, ce.RelativePath)
	}

	xmlDate := ce.ContentMetadata.Date
	if xmlDate.IsZero() {
		xmlDate = time.Now()
	}
	xmlUrl := SitemapUrl{
		Loc:     siteUrl.String(),
		LastMod: xmlDate.Format(_YYYYMMDD_RFC3339),
	}
	return xmlUrl
}

// TODO: There are a few places I have taken shortcuts like this function that need to be fixed to reduce complexity and lines of code when time permits
func (m *metadata) getContentMetadata(fileData []byte, fileName string, contentRoot string) (*ContentEntity, error) {

	if len(fileData) == 0 {
		return nil, fmt.Errorf("no data provided for file to be rendered to content")
	}
	if len(fileData) > _maxInputSize {
		return nil, fmt.Errorf("file data size exceeds current max")
	}

	frontmatter, bodyData, err := SplitFileContent(fileData, m.ctx.FrontmatterToken)
	if err != nil {
		slog.Warn("unable to extract frontmatter, continuing with defaults", "file", fileName)
	}
	frontmatter.Description = m.extractDescription(frontmatter, bodyData)

	fullFileName := strings.ReplaceAll(strings.ToLower(fileName), " ", "-")
	ce := ContentEntity{
		ContentMetadata: frontmatter,
		Name:            fullFileName,
	}

	subDir := ""
	if strings.TrimSpace(ce.ContentMetadata.TemplateId) == "" {
		switch contentRoot {

		case m.ctx.PostInputDir:
			ce.ContentMetadata.TemplateId = _defaultPostTemplate
			subDir = m.ctx.PostOutputDir

		case m.ctx.PageInputDir:
			ce.ContentMetadata.TemplateId = _defaultPageTemplate
		}
	}

	usePrettyUrl := !m.ctx.FullHtmlPaths && fullFileName != _indexHtmlFile
	usePermalink := len(strings.TrimSpace(ce.ContentMetadata.Permalink)) > 1
	if usePrettyUrl {
		if usePermalink {
			ce.OutPath = filepath.Join(subDir, ce.ContentMetadata.Permalink, _indexHtmlFile)
			ce.RelativePath = path.Join(m.ctx.Base, subDir, ce.ContentMetadata.Permalink)
		} else {
			fileName := strings.TrimSuffix(fullFileName, filepath.Ext(fullFileName))
			ce.OutPath = filepath.Join(subDir, fileName, _indexHtmlFile)
			ce.RelativePath = path.Join(m.ctx.Base, subDir, fileName)
		}
	} else {
		if usePermalink {
			ce.OutPath = filepath.Join(subDir, ce.ContentMetadata.Permalink)
		} else {
			ce.OutPath = filepath.Join(subDir, fullFileName)
		}
		ce.RelativePath = path.Join(m.ctx.Base, subDir)
	}

	ce.ContentMetadata.Url = ce.RelativePath
	return &ce, nil
}

func (m *metadata) extractDescription(fm ContentMetadata, body []byte) string {
	desc := strings.TrimSpace(fm.Description)
	if desc == "" {
		desc = truncateBytes(body, m.ctx.MaxDescriptionLen)
	} else if len(desc) > m.ctx.MaxDescriptionLen {
		desc = truncateBytes([]byte(desc), m.ctx.MaxDescriptionLen)
	}
	return desc
}

func truncateBytes(data []byte, limit int) string {
	if len(data) == 0 {
		return ""
	}
	end := limit
	if len(data) < limit {
		end = len(data)
	}
	summary := strings.Join(strings.Fields(string(data[:end])), " ")

	if len(data) > limit {
		summary += "..."
	}
	return summary
}

// TODO: move
func SplitFileContent(content []byte, token string) (ContentMetadata, []byte, error) {

	tok := []byte(token)
	var fm ContentMetadata
	if len(token) < 3 {
		return fm, content, fmt.Errorf("invalid frontmatter token: minimum length 3 required")
	}

	content = bytes.TrimSpace(content)
	if !bytes.HasPrefix(content, tok) {
		return fm, content, fmt.Errorf("content missing starting frontmatter token")
	}
	start := len(token)
	end := bytes.Index(content[start:], tok)
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
