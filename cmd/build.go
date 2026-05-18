package cmd

import (
	"log/slog"

	"github.com/closeencoders/arcstatic/generate"
	"github.com/closeencoders/arcstatic/source"
)

func executeBuild(location string) {

	slog.Debug("loading site context and configuration")
	ctx, err := source.LoadSiteCtx(location)
	if err != nil {
		slog.Error("failed to load site context and configuration", "error", err)
		panic(err)
	}

	slog.Debug("loading metadata")
	ml := source.NewMetadata(ctx)
	metadata, err := ml.LoadMetadata(ctx.PostInputDir, ctx.PageInputDir)
	if err != nil {
		slog.Error("failed to load source material for site generation", "error", err)
		panic(err)
	}

	t, err := generate.NewTemplater(ctx.ComponentMap, nil)
	if err != nil {
		slog.Error("failed to load components for templating", "error", err)
		panic(err)
	}

	c := *generate.NewConverter(ctx, *generate.NewMarkdown(ctx), *t)
	sg := generate.NewGenerator(ctx, c)
	if err := sg.Generate(*metadata); err != nil {
		slog.Error("failed to generate site", "error", err)
		panic(err)
	}
}
