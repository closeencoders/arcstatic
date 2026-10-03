package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
	"github.com/spf13/cobra"
)

func createShowCommand(ssg *ssg) {
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "show or print functions",
	}

	config := false
	size := false

	showCmd.Flags().BoolVarP(&config, "config", "c", false, "builds static site from provided resources")
	showCmd.Flags().BoolVarP(&size, "size", "s", false, "builds static site from provided resources")

	showCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if config {
			return executeShowConfig(ssg.ctx)
		}
		if size {
			return executeShowSize(ssg.ctx)
		}
		return nil
	}

	ssg.cmd.AddCommand(showCmd)
}

func executeShowConfig(ctx *config.SiteContext) error {

	if ctx == nil {
		return fmt.Errorf("No valid configuration found to print")
	}

	content, err := json.MarshalIndent(ctx, "", " ")
	if err != nil {
		return fmt.Errorf("unable to marshal configuration: %w", err)
	}

	fmt.Println(string(content))

	return nil
}

// TODO: Currently shares the same process the verify command will have.
func executeShowSize(ctx *config.SiteContext) error {

	if ctx == nil {
		return fmt.Errorf("No valid configuration found to search and print size")
	}

	slog.Debug("loading metadata")
	ml := source.NewMetadata(ctx, storage.NewOSFileStorage())
	metadata, err := ml.LoadMetadata(ctx.PostInputDir, ctx.PageInputDir)
	if err != nil {
		return fmt.Errorf("failed to load source material for site generation: %w", err)
	}

	slog.Info("content files", "size", len(metadata.SiteContentEntities))

	return nil
}
