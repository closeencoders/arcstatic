package source

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/closeencoders/arcstatic/config"
)

const (
	_indexHtmlFile             = "index.html"
	_maxInputSize              = 1_000_000
	_defaultMaxDescriptionSize = 156
	_YYYYMMDD_RFC3339          = "2006-01-02T15:04:05Z07:00"

	_defaultPostsItem    = "Posts"
	_defaultPostTemplate = "post.html"
	_defaultPageTemplate = "page.html"
)

type ContentManifest map[string][]*ContentMetadata

type SitemapUrl struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type SiteMetadata struct {
	// content entity struct to hold all original content metadata loaded from source material until I find a better pattern
	SiteContentEntities []*ContentEntity
	// Represent the state of every categorized set of metadata
	ContentManifest ContentManifest
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
	InPath string
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

	IsDraft bool `json:"-" yaml:"is_draft"`

	Type       string   `json:"type" yaml:"type"`
	Categories []string `json:"categories" yaml:"categories"`
	Tags       []string `json:"tags" yaml:"tags"`
}

type Metadata struct {
	ctx config.SiteContext
}

func NewMetadata(ctx config.SiteContext) *Metadata {
	return &Metadata{ctx: ctx}
}

func (m *Metadata) LoadMetadata(locations ...string) (*SiteMetadata, error) {

	var metadata SiteMetadata
	for _, loc := range locations {
		if err := m.readSiteMetadataFiles(loc, &metadata); err != nil {
			return &metadata, fmt.Errorf("failed to load metadata for %s, %w", loc, err)
		}
	}
	slices.SortFunc(metadata.SiteContentEntities, func(a, b *ContentEntity) int {
		return b.ContentMetadata.Date.Compare(a.ContentMetadata.Date)
	})

	// TODO: maintain hierarchy and flatten appropriately
	manifest := ContentManifest{}
	for _, ce := range metadata.SiteContentEntities {

		fm := &ce.ContentMetadata
		id := strings.TrimSpace(fm.TemplateId)

		if id == _defaultPostTemplate || (id == "" && m.ctx.PostInputDir == ce.InPath) {

			if fm.Type == "" {
				if m.ctx.DefaultType != "" {
					manifest[m.ctx.DefaultType] = append(manifest[m.ctx.DefaultType], fm)
				}
			} else {
				manifest[fm.Type] = append(manifest[fm.Type], fm)
			}

			manifest[_defaultPostsItem] = append(manifest[_defaultPostsItem], fm)
			manifest = renderTypes(fm, manifest, fm.Tags...)
			manifest = renderTypes(fm, manifest, fm.Categories...)
		}
	}
	metadata.ContentManifest = manifest

	return &metadata, nil
}

func (m *Metadata) readSiteMetadataFiles(root string, metadata *SiteMetadata) error {

	slog.Debug("rendering content", "path", root)
	return filepath.WalkDir(root, func(currentPath string, dirEntry os.DirEntry, err error) error {

		if dirEntry == nil || dirEntry.IsDir() {
			slog.Debug("not a file that can be used for metadata extraction", "dir", dirEntry, "path", root, "currentPath", currentPath)
			return nil
		}
		rawFile, err := LoadFileBytes(currentPath)
		if err != nil {
			return fmt.Errorf("failed to load file data: %w", err)
		}
		if len(rawFile) < 3 {
			slog.Debug("no file data viable for conversion found", "dir", dirEntry, "path", root)
			return nil
		}

		content, err := m.getContentMetadata(rawFile, dirEntry.Name(), root)
		if err != nil {
			return fmt.Errorf("failed to convert to content: %w", err)
		}

		content.InPath = currentPath
		metadata.SiteContentEntities = append(metadata.SiteContentEntities, content)

		if m.ctx.GenerateSitemapXml {
			xmlUrl := makeSitemapEntry(m.ctx, content)
			metadata.SiteMapUrlMetadata = append(metadata.SiteMapUrlMetadata, xmlUrl)
		}

		return err
	})
}

func makeSitemapEntry(ctx config.SiteContext, ce *ContentEntity) SitemapUrl {
	siteUrl, _ := url.Parse(ctx.SiteUrl)
	if ctx.FullHtmlPath {
		siteUrl.Path = path.Join(siteUrl.Path, ce.OutPath)
	} else {
		siteUrl.Path = path.Join(siteUrl.Path, ce.RelativePath)
	}
	loc := siteUrl.String()

	xmlDate := ce.ContentMetadata.Date
	if xmlDate.IsZero() {
		xmlDate = time.Now()
	}

	xmlUrl := SitemapUrl{
		Loc:     loc,
		LastMod: xmlDate.Format(_YYYYMMDD_RFC3339),
	}
	slog.Debug("sitemap", "url", xmlUrl, "site", ctx.SiteUrl)
	return xmlUrl
}

// TODO: There are a few places I have taken shortcuts like this function that need to be fixed to reduce complexity and lines of code when time permits
func (m *Metadata) getContentMetadata(fileData []byte, fileName string, contentRoot string) (*ContentEntity, error) {

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
	frontmatter.Description = extractDescription(frontmatter, bodyData)

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
			subDir = m.ctx.PostOutDir

		case m.ctx.PageInputDir:
			ce.ContentMetadata.TemplateId = _defaultPageTemplate
		}
	}

	usePrettyUrl := !m.ctx.FullHtmlPath && fullFileName != _indexHtmlFile
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

func renderTypes(fm *ContentMetadata, data map[string][]*ContentMetadata, types ...string) map[string][]*ContentMetadata {
	if len(types) == 0 {
		return data
	}
	for _, id := range types {
		id = strings.TrimSpace(id)
		if id == "" || id == _defaultPostsItem {
			continue
		}
		data[id] = append(data[id], fm)
	}
	return data
}

func extractDescription(fm ContentMetadata, body []byte) string {
	desc := strings.TrimSpace(fm.Description)
	if desc == "" {
		desc = truncateBytes(body, _defaultMaxDescriptionSize)
	} else if len(desc) > _defaultMaxDescriptionSize {
		desc = truncateBytes([]byte(desc), _defaultMaxDescriptionSize)
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
