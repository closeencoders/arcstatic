package source

import (
	"strings"

	"github.com/closeencoders/arcstatic/config"
)

type Manifest map[string][]*ContentMetadata

func NewManifest(ctx config.SiteContext, contents []*ContentEntity) Manifest {
	manifest := Manifest{}
	for _, ce := range contents {

		fm := &ce.ContentMetadata
		id := strings.TrimSpace(fm.TemplateId)

		if id == _defaultPostTemplate || (id == "" && ctx.PostInputDir == ce.InPath) {

			if fm.Type == "" {
				if ctx.DefaultType != "" {
					manifest[ctx.DefaultType] = append(manifest[ctx.DefaultType], fm)
				}
			} else {
				manifest[fm.Type] = append(manifest[fm.Type], fm)
			}

			manifest[_defaultPostsItem] = append(manifest[_defaultPostsItem], fm)
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
		if id == "" || id == _defaultPostsItem {
			continue
		}
		data[id] = append(data[id], fm)
	}
	return data
}
