package source

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

var (
	isoDateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)
)

type SitemapUrl struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type Manifest map[string][]*ContentMetadata

type SiteMetadata struct {
	// content entity struct to hold all original content metadata loaded from source material until I find a better pattern
	SiteContentEntities []*ContentEntity
	// Represent the state of every categorized content file without holding the entire file in memory
	ContentManifest Manifest
	// Used for creating xml representations for the site for search engines, like a sitemap.xml
	SiteMapUrlMetadata []SitemapUrl
}

type ContentEntity struct {
	// OG file name: "2006-01-02-some_post.md"
	FileName string
	// Modified name that will be exported: "some-post"
	ArtificialFileName string
	// Data extracted via a header or similar from a content file
	ContentMetadata ContentMetadata
	// Sub-path from root to the content: /posts/
	RelativePath string
	// Full path to file with rendered content: pcroot/siteroot/posts/some-post.html
	OutputPath string
	// Full path to original content file to be rendered:  pcroot/siteroot/posts/2006-01-02-some-post.md
	InputPath string
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
	metadata.ContentManifest = Manifest{}
	for _, loc := range paths {
		if err := m.readSiteMetadataFiles(loc, &metadata); err != nil {
			return nil, fmt.Errorf("failed to load metadata for %s: %w", loc, err)
		}
	}

	if m.ctx.AllowNamelessDateSort {
		slog.Debug("sorting content by date")
		slices.SortFunc(metadata.SiteContentEntities, func(a, b *ContentEntity) int {
			return b.ContentMetadata.Date.Compare(a.ContentMetadata.Date)
		})
	}

	return &metadata, nil
}

func (m *metadata) readSiteMetadataFiles(root string, metadata *SiteMetadata) error {

	baseLog := slog.With("root", root)
	baseLog.Debug("searching for source files")

	return fs.WalkDir(m.store, root, func(path string, dirEntry os.DirEntry, err error) error {

		log := baseLog.With("path", path)

		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				log.Warn("file does not exist, skipping")
				return nil
			}
			return fmt.Errorf("failed to walk dir %s: %w", path, err)
		}
		if dirEntry == nil || dirEntry.IsDir() {
			log.Debug("next directory entry found")
			return nil
		}

		fileData, err := storage.LoadSiteFile(path, m.store)
		if err != nil {
			if errors.Is(err, storage.ErrUnsupported) {
				log.Debug("unsupported file for metadata extraction, skipping")
				return nil
			}
			return fmt.Errorf("failed to load file data: %w", err)
		}
		if len(fileData.Data) < 3 {
			log.Warn("not enough file data viable for conversion, skipping")
			return nil
		}
		if len(fileData.Data) > _maxInputSize {
			log.Warn("file data size exceeds current max, skipping")
			return nil
		}

		content, err := m.getContentMetadata(fileData.Data, fileData.Name, root)
		if err != nil {
			return fmt.Errorf("failed to convert to content %s: %w", path, err)
		}
		if content.ContentMetadata.Draft {
			log.Debug("is draft")
			return nil
		}
		content.InputPath = path

		if m.ctx.AllowManifest {
			m.updateManifest(content, metadata)
		}
		if m.ctx.MakeSitemapXML {
			m.updateSitemap(content, metadata)
		}

		metadata.SiteContentEntities = append(metadata.SiteContentEntities, content)

		return err
	})
}

func (m *metadata) updateManifest(content *ContentEntity, metadata *SiteMetadata) {
	cm := &content.ContentMetadata
	id := strings.TrimSpace(cm.TemplateId)
	if id == _defaultPostTemplate || (id == "" && m.ctx.PostInputDir == content.InputPath) {
		if cm.Type != "" {
			metadata.ContentManifest[cm.Type] = append(metadata.ContentManifest[cm.Type], cm)
		} else {
			cm.Type = m.ctx.DefaultType
		}
		// Will contain all posts
		metadata.ContentManifest[m.ctx.DefaultType] = append(metadata.ContentManifest[m.ctx.DefaultType], cm)
		metadata.ContentManifest = sortTypes(cm, metadata.ContentManifest, cm.Tags...)
		metadata.ContentManifest = sortTypes(cm, metadata.ContentManifest, cm.Categories...)
	}
}

func sortTypes(cm *ContentMetadata, data map[string][]*ContentMetadata, types ...string) map[string][]*ContentMetadata {
	if len(types) == 0 {
		return data
	}
	for _, id := range types {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		data[id] = append(data[id], cm)
	}
	return data
}

// TODO: the input dir is not a dir, its the path to the content file. this needs to be corrected so the artificial name is
// the output file name after filtering and the input dir is field name is correct, InPath or something...
func (m *metadata) getContentMetadata(fileData []byte, fileName string, root string) (*ContentEntity, error) {

	frontmatter, bodyData, err := SplitFileContent(fileData, m.ctx.FrontmatterToken)
	if err != nil {
		slog.Warn("unable to extract frontmatter, continuing with defaults", "file", fileName)
	}
	frontmatter.Description = m.extractDescription(frontmatter, bodyData)
	fullFileName := strings.ReplaceAll(strings.ToLower(fileName), " ", "-")

	ce := ContentEntity{
		FileName:           fileName,
		ArtificialFileName: fullFileName,
		// TODO: transformation to metadata around here or when split from the file should allow the frontmatter to be dynamically set
		ContentMetadata: frontmatter,
	}

	if !m.ctx.AllowNamelessDateSort && root != m.ctx.PageInputDir {
		datePrefix := isoDateRegex.FindStringIndex(fileName)
		if datePrefix == nil || datePrefix[0] != 0 {
			return nil, fmt.Errorf("file not prefixed with valid date, this can be disabled in configuration at the cost of performance")
		}
		ce.ArtificialFileName = strings.TrimLeft(fileName[datePrefix[1]:], "_- ")
	}

	subDir := ""
	if strings.TrimSpace(ce.ContentMetadata.TemplateId) == "" {
		switch root {

		case m.ctx.PostInputDir:
			ce.ContentMetadata.TemplateId = _defaultPostTemplate
			subDir = m.ctx.PostOutputDir

		case m.ctx.PageInputDir:
			ce.ContentMetadata.TemplateId = _defaultPageTemplate
		}
	}

	// TODO: There are a few places I have taken shortcuts like this that need to be fixed to reduce complexity and lines of code when time permits
	usePrettyUrl := !m.ctx.FullHtmlPaths && fullFileName != _indexHtmlFile
	usePermalink := len(strings.TrimSpace(ce.ContentMetadata.Permalink)) > 1
	if usePrettyUrl {
		if usePermalink {
			ce.OutputPath = filepath.Join(subDir, ce.ContentMetadata.Permalink, _indexHtmlFile)
			ce.RelativePath = path.Join(m.ctx.Base, subDir, ce.ContentMetadata.Permalink)
		} else {
			fileName := strings.TrimSuffix(fullFileName, filepath.Ext(ce.ArtificialFileName))
			ce.OutputPath = filepath.Join(subDir, fileName, _indexHtmlFile)
			ce.RelativePath = path.Join(m.ctx.Base, subDir, fileName)
		}
	} else {
		if usePermalink {
			ce.OutputPath = filepath.Join(subDir, ce.ContentMetadata.Permalink)
		} else {
			ce.OutputPath = filepath.Join(subDir, ce.ArtificialFileName)
		}
		ce.RelativePath = path.Join(m.ctx.Base, subDir)
	}

	ce.ContentMetadata.Url = ce.RelativePath

	return &ce, nil
}

func (m *metadata) updateSitemap(content *ContentEntity, metadata *SiteMetadata) {

	siteUrl, _ := url.Parse(m.ctx.SiteURL)
	if m.ctx.FullHtmlPaths {
		siteUrl.Path = path.Join(siteUrl.Path, content.OutputPath)
	} else {
		siteUrl.Path = path.Join(siteUrl.Path, content.RelativePath)
	}

	xmlDate := content.ContentMetadata.Date
	if xmlDate.IsZero() {
		xmlDate = time.Now()
	}
	xmlUrl := SitemapUrl{
		Loc:     siteUrl.String(),
		LastMod: xmlDate.Format(_YYYYMMDD_RFC3339),
	}

	metadata.SiteMapUrlMetadata = append(metadata.SiteMapUrlMetadata, xmlUrl)
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
