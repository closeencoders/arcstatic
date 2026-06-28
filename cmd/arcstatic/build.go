package main

import (
	"fmt"
	"log/slog"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/generate"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
)

func executeBuild(ctx *config.SiteContext, store storage.Storage, path string) error {

	slog.Debug("loading metadata")
	ml := source.NewMetadata(ctx, store)
	metadata, err := ml.LoadMetadata(ctx.PostInputDir, ctx.PageInputDir)
	if err != nil {
		return fmt.Errorf("failed to load source material for site generation: %w", err)
	}

	t, err := generate.NewTemplater(ctx.ComponentMap, nil)
	if err != nil {
		return fmt.Errorf("failed to load components for templating: %w", err)
	}

	c := *generate.NewConverter(ctx, *generate.NewMarkdown(ctx), *t)
	sg := generate.NewGenerator(ctx, c, store)
	if _, err := sg.Generate(metadata); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	slog.Debug("build complete")
	return nil
}
