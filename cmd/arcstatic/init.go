package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"
)

func createInitCommand(ssg *ssg) {

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "create basic project",
	}

	initCmd.RunE = func(cmd *cobra.Command, args []string) error {
		slog.Info("Will eventually initialize a default configuration files and basic project setup before being able to run the build command.")
		return nil
	}

	ssg.cmd.AddCommand(initCmd)
}

func initProject(ssg *ssg) error {

	configLoc := filepath.Join(ssg.outputLocation, "arcconfig.yml")
	if ssg.store.Exists(configLoc) {
		slog.Info("Config already exists")
		return nil
	}

	data, err := getConfigData(ssg.ctx)
	if err != nil {
		return fmt.Errorf("failed to get data to write config file %w", err)
	}
	if err = ssg.store.Write(configLoc, data, 0755); err != nil {
		return fmt.Errorf("failed to write config file %w", err)
	}

	return nil
}
