package main

import (
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
)

var (
	// local flags
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "config opens the configured config file for editing",
		Long:  "config opens the configured config file for editing in the default editor (determined by $EDITOR, defaulting to vim)",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Config(logger, configPath)
		},
	}
)

func init() {
	rootCmd.AddCommand(configCmd)
}
