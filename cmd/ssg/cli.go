package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inputLocation string
	build         bool
	serve         bool
	port          int
	verbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "arcstatic",
	Short: "A Simple SSG",
	Long:  "Arcstatic is a simple static site generator",
	Args:  cobra.ArbitraryArgs,
	Run:   runSsg,
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
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "run verbose with debug logs")
}

// TODO: move to cobra and/or viper when time permits
func runSsg(cmd *cobra.Command, args []string) {
	size := len(os.Args)
	if size < 2 || size > 200 {
		fmt.Println("Unable to parse input, try -help")
		return
	}

	var level slog.Level = slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	defaultLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(defaultLogger)

	slog.Debug("Running...")

	if strings.TrimSpace(inputLocation) == "" || len(inputLocation) > 5000 {
		slog.Info("Invalid Input Location Provided")
		return
	}

	if build {
		executeBuild(inputLocation)
	}
	if serve {
		executeServe(inputLocation, port)
	}
}
