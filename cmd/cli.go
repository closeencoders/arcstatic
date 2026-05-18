package cmd

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// TODO: move to cobra and/or viper when time permits
func Run() {
	size := len(os.Args)
	if size < 2 || size > 200 {
		fmt.Println("Unable to parse input, try -help")
		return
	}
	defaultLocation, err := os.Getwd()
	if err != nil {
		fmt.Println("Unable to get working directory, try -help")
	}

	build := flag.Bool("build", false, "Builds static site from provided resources. Defaults to current worked directory.")
	verbose := flag.Bool("verbose", false, "run verbose")
	location := flag.String("in", defaultLocation, "Static site context input location. Defaults to current worked directory.")

	serve := flag.Bool("serve", false, "Serve static site from provided resources for testing. Defaults to current worked directory.")
	port := flag.Int("port", 8000, "location to file serve. If the serve command is not used, this is ignored")

	flag.Parse()

	var level slog.Level = slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	defaultLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(defaultLogger)

	slog.Debug("Running...")

	if strings.TrimSpace(*location) == "" || len(*location) > 5000 {
		slog.Info("Invalid Location provided")
		return
	}

	if *build {
		executeBuild(*location)
	}
	if *serve {
		executeServe(*location, *port)
	}
}
