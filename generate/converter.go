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
	ctx       *config.SiteContext
	markdown  Markdown
	templater templater
}

func NewConverter(ctx *config.SiteContext, markdown Markdown, templater templater) *converter {
	return &converter{ctx: ctx, markdown: markdown, templater: templater}
}

func (c *converter) ConvertToContent(rawFile []byte, content *source.ContentEntity, manifest source.ContentManifest) ([]byte, error) {

	_, body, err := source.SplitFileContent(rawFile, c.ctx.FrontmatterToken)
	if err != nil {
		slog.Warn("unable to extract frontmatter, continuing with defaults", "name", content.Name, "err", err)
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
	fileExt := strings.ToLower(filepath.Ext(content.Name))
	if fileExt == ".md" {
		htmlResult, err := c.markdown.ConvertToHtml(body)
		if err != nil {
			return body, fmt.Errorf("markdown conversion for %s failed: %w", content.Name, err)
		}
		body = htmlResult.HTML
		renderMap["TOC"] = string(htmlResult.TOC)
	}

	body, err = c.templater.Render(renderMap, string(body))
	if err != nil {
		return body, fmt.Errorf("unable to render body: %w", err)
	}

	renderMap["Body"] = string(body)
	templateId := content.ContentMetadata.TemplateId
	templateStr := string(c.ctx.TemplateMap[templateId])
	body, err = c.templater.Render(renderMap, templateStr)
	if err != nil {
		return body, fmt.Errorf("unable to render body with template %s for entity %s: %w", templateId, content.Name, err)
	}

	return body, nil
}
