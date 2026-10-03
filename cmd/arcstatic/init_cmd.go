package main

import (
	"log/slog"

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
