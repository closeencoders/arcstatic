package generate

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/closeencoders/arcstatic/config"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Markdown struct {
	ctx      *config.SiteContext
	goldmark goldmark.Markdown
}

type MarkdownResult struct {
	HTML []byte
	TOC  []byte
}

type nodeTransformer struct{}

func NewMarkdown(ctx *config.SiteContext) *Markdown {
	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&nodeTransformer{}, 100),
			),
		),
	)
	return &Markdown{ctx: ctx, goldmark: gm}
}

func (m *Markdown) ConvertToHtml(content []byte) (*MarkdownResult, error) {

	var result MarkdownResult
	docNode := m.goldmark.Parser().Parse(text.NewReader(content))
	if m.ctx.MakeTableOfContents {
		tocBuffer := makeTableOfContents(content, &docNode)
		result.TOC = tocBuffer.Bytes()
	}

	var buf bytes.Buffer
	if err := m.goldmark.Renderer().Render(&buf, content, docNode); err != nil {
		return nil, err
	}

	result.HTML = buf.Bytes()
	return &result, nil
}

func (t *nodeTransformer) Transform(docNode *ast.Document, reader text.Reader, pc parser.Context) {

	_ = ast.Walk(docNode, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		var dest string

		switch typed := node.(type) {
		case *ast.Link:
			dest = string(typed.Destination)
		case *ast.AutoLink:
			dest = string(typed.URL(reader.Source()))
		default:
			return ast.WalkContinue, nil
		}

		if strings.HasPrefix(dest, "http") {
			node.SetAttributeString("target", []byte("_blank"))
			node.SetAttributeString("rel", []byte("nofollow noopener noreferrer"))
		}

		return ast.WalkContinue, nil
	})
}

// TODO:
func makeTableOfContents(content []byte, node *ast.Node) bytes.Buffer {

	var buf bytes.Buffer
	currentLevel := 0
	var hasHeader bool = false

	_ = ast.Walk(*node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*ast.Heading); ok && entering {
			if !hasHeader {
				buf.WriteString("<ul class=\"toc-list\">")
			}
			hasHeader = true
			id, _ := heading.AttributeString("id")
			title := string(n.Lines().Value(content))
			level := heading.Level

			if currentLevel == 0 {
				currentLevel = level
			}
			if level > currentLevel {
				for i := 0; i < (level - currentLevel); i++ {
					buf.WriteString("<ul>")
				}
			} else if level < currentLevel {
				for i := 0; i < (currentLevel - level); i++ {
					buf.WriteString("</ul>")
				}
			}

			buf.WriteString(fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", id, title))
			currentLevel = level
		}
		return ast.WalkContinue, nil
	})

	// Close any remaining open tags after the walk
	for i := 0; i < currentLevel-1; i++ {
		buf.WriteString("</ul>")
	}

	return buf
}
