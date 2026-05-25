package source

import (
	"strings"

	"github.com/closeencoders/arcstatic/config"
)

type Manifest map[string][]*ContentMetadata

func NewManifest(ctx config.SiteContext, contents []*ContentEntity) Manifest {
	manifest := Manifest{}
	for _, ce := range contents {
		if ce == nil {
			continue
		}
		fm := &ce.ContentMetadata
		id := strings.TrimSpace(fm.TemplateId)
		if id == _defaultPostTemplate || (id == "" && ctx.PostInputDir == ce.InputDir) {
			// Posts only of Subtype
			if fm.Type != "" {
				manifest[fm.Type] = append(manifest[fm.Type], fm)
			}
			// All posts
			manifest[ctx.DefaultType] = append(manifest[ctx.DefaultType], fm)
			manifest = renderTypes(fm, manifest, fm.Tags...)
			manifest = renderTypes(fm, manifest, fm.Categories...)
		}
	}
	return manifest
}

func renderTypes(fm *ContentMetadata, data map[string][]*ContentMetadata, types ...string) map[string][]*ContentMetadata {
	if len(types) == 0 {
		return data
	}
	for _, id := range types {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		data[id] = append(data[id], fm)
	}
	return data
}
