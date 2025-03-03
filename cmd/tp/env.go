package main

import (
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
)

var (
	// local flags
	envCmd = &cobra.Command{
		Use:   "env",
		Short: "env opens the configured environment file for editing",
		Long:  "env opens the configured environment file for editing in the default editor (determined by $EDITOR, defaulting to vim)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return handlers.Env(logger, c, args[0])
			} else {
				return handlers.Env(logger, c, envFile)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(envCmd)
}
