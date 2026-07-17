package generate

import (
	"fmt"
	"log/slog"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
)

type converter struct {
	ctx      *config.SiteContext
	markdown Markdown
	renderer Renderer
}

func NewConverter(ctx *config.SiteContext, markdown Markdown, renderer Renderer) *converter {
	return &converter{ctx: ctx, markdown: markdown, renderer: renderer}
}

func (c *converter) ToContent(rawFile []byte, content *source.ContentEntity, manifest source.Manifest) ([]byte, error) {

	_, body, err := source.SplitFileContent(rawFile, c.ctx.FrontmatterToken)
	if err != nil {
		slog.Warn("unable to extract frontmatter, continuing with defaults", "name", content.FileName, "err", err)
	}

	canonicalUrl, _ := url.Parse(c.ctx.SiteURL)
	canonicalUrl.Path = path.Join(canonicalUrl.Path, content.ContentMetadata.Url)
	renderMap := map[string]interface{}{
		"Metadata":     content.ContentMetadata,
		"CanonicalURL": canonicalUrl.String(),
	}
	for k, v := range manifest {
		renderMap[k] = v
	}

	// TODO: Type detection should happen in the source phase
	fileExt := strings.ToLower(filepath.Ext(content.FileName))
	if fileExt == ".md" {
		htmlResult, err := c.markdown.ToHtml(body)
		if err != nil {
			return body, fmt.Errorf("markdown conversion for %s failed: %w", content.FileName, err)
		}
		body = htmlResult.HTML
		renderMap["TOC"] = string(htmlResult.TOC)
	}

	body, err = c.renderer.Render(renderMap, string(body))
	if err != nil {
		return body, fmt.Errorf("unable to render body: %w", err)
	}

	templateId := content.ContentMetadata.TemplateId
	templateStr := string(c.ctx.TemplateMap[templateId])

	// re-render with template loaded only if template exists
	if strings.TrimSpace(templateStr) != "" {
		renderMap["Body"] = string(body)
		body, err = c.renderer.Render(renderMap, templateStr)
		if err != nil {
			return body, fmt.Errorf("unable to render body with template %s for entity %s: %w", templateId, content.FileName, err)
		}
	}

	return body, nil
}
