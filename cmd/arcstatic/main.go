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

var (
	inputLocation string
	build         bool
	serve         bool
	port          int
	verbose       bool
	ctx           *config.SiteContext
)

var rootCmd = &cobra.Command{
	Use:           "arcstatic",
	Short:         "A Simple SSG",
	Long:          "Arcstatic is a simple static site generator",
	Args:          cobra.ArbitraryArgs,
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		config, err := source.LoadSiteContext(inputLocation, storage.NewOSFileStorage())
		if err != nil {
			return fmt.Errorf("failed to load site context and configuration: %w", err)
		}
		// TODO: split the config and context into different structs and have the context contain the config. this might be a large refactor, waiting to see if this is the right choice
		ctx = config

		var level slog.Level = slog.LevelInfo
		if verbose {
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
	},
	RunE: runSsg,
}

func init() {
	defaultLocation, err := os.Getwd()
	if err != nil {
		defaultLocation = "."
	}
	rootCmd.Flags().StringVarP(&inputLocation, "in", "i", defaultLocation, "override default site context current working directory input location")
	rootCmd.Flags().BoolVarP(&build, "build", "b", false, "builds static site from provided resources")
	rootCmd.Flags().BoolVarP(&serve, "serve", "s", false, "serve static site from provided resources, currently only for testing")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8000, "port number to file serve the site, if the serve command is not used, this is ignored")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "run verbose with debug logs, this will attempt to override any config file settings")
}

func runSsg(cmd *cobra.Command, args []string) error {

	slog.Debug("Running Commands", "cmds", os.Args)

	inputLocation = strings.TrimSpace(inputLocation)
	if inputLocation == "" || len(inputLocation) > 5000 {
		return fmt.Errorf("invalid input location provided")
	}

	if build {
		if err := executeBuild(ctx, storage.NewOSFileStorage()); err != nil {
			return fmt.Errorf("build command failed: %w", err)
		}
	}
	if serve {
		if err := executeServe(inputLocation, port); err != nil {
			return fmt.Errorf("serve command failed: %w", err)
		}
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Execution halted", "error", err)
		os.Exit(1)
	}
}
