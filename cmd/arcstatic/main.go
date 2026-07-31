package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/closeencoders/arcstatic/config"
	"github.com/closeencoders/arcstatic/source"
	"github.com/closeencoders/arcstatic/storage"
	"github.com/spf13/cobra"
)

type ssg struct {
	inputLocation  string
	outputLocation string
	build          bool
	serve          bool
	port           int
	verbose        bool

	store storage.Storage
	ctx   *config.SiteContext
	cmd   *cobra.Command
}

func NewSsg(cmd *cobra.Command, store storage.Storage) *ssg {

	defaultLocation, err := store.GetWd()
	if err != nil {
		defaultLocation = "."
	}
	ssg := ssg{store: store, cmd: cmd}

	cmd.Flags().StringVarP(&ssg.inputLocation, "in", "i", defaultLocation, "override default site context current working directory input location, defaults to current location")
	cmd.Flags().StringVarP(&ssg.outputLocation, "out", "o", "", "override default site context current working directory out location, defaults to input location")
	cmd.Flags().BoolVarP(&ssg.build, "build", "b", false, "builds static site from provided resources")
	cmd.Flags().BoolVarP(&ssg.serve, "serve", "s", false, "serve static site from provided resources, currently only for testing")
	cmd.Flags().IntVarP(&ssg.port, "port", "p", 8000, "port number to file serve the site, if the serve command is not used, this is ignored")
	cmd.Flags().BoolVarP(&ssg.verbose, "verbose", "v", false, "run verbose with debug logs, this will attempt to override any config file settings")

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		inAbs, err := storage.ToCleanAbs(ssg.inputLocation)
		if err != nil {
			return fmt.Errorf("invalid input path: %s, %w", ssg.inputLocation, err)
		}
		ssg.inputLocation = inAbs

		outAbs, err := storage.ToCleanAbs(ssg.outputLocation)
		if err != nil {
			return fmt.Errorf("invalid output path: %s, %w", ssg.outputLocation, err)
		}
		ssg.outputLocation = outAbs

		config, err := source.LoadSiteContext(ssg.inputLocation, ssg.outputLocation, ssg.store)
		if err != nil {
			return fmt.Errorf("failed to load site context and configuration: %w", err)
		}
		// TODO: split the config and context into different structs and have the context contain the config. this might be a large refactor, waiting to see if this is the right choice
		ssg.ctx = config

		var level slog.Level = slog.LevelInfo
		if ssg.verbose {
			level = slog.LevelDebug
		} else if strings.TrimSpace(config.LogLevel) == "" {
			level = slog.LevelInfo
		} else if err := level.UnmarshalText([]byte(config.LogLevel)); err != nil {
			level = slog.LevelInfo
		}

		opts := &slog.HandlerOptions{Level: level}
		var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
		if config.JsonLog {
			handler = slog.NewJSONHandler(os.Stdout, opts)
		}
		slog.SetDefault(slog.New(handler))

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		slog.Debug("Running Commands", "cmds", os.Args)
		ssg.inputLocation = strings.TrimSpace(ssg.inputLocation)
		if ssg.inputLocation == "" || len(ssg.inputLocation) > 5000 {
			return fmt.Errorf("invalid input location provided")
		}

		if ssg.build {
			if err := executeBuild(ssg.ctx, ssg.store); err != nil {
				return fmt.Errorf("build command failed: %w", err)
			}
		}
		if ssg.serve {
			if err := executeServe(ssg.inputLocation, ssg.port); err != nil {
				return fmt.Errorf("serve command failed: %w", err)
			}
		}
		return nil
	}

	return &ssg
}

func main() {
	var rootCmd = &cobra.Command{
		Use:           "arcstatic",
		Short:         "A Simple SSG",
		Long:          "Arcstatic is a simple static site generator",
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	ssg := NewSsg(rootCmd, storage.NewOSFileStorage())
	cobra.CheckErr(ssg.cmd.Execute())
}
