package main

import (
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
)

var (
	// local flags
	varsCmd = &cobra.Command{
		Use:   "vars",
		Short: "vars opens the configured variables file for editing",
		Long:  "vars opens the configured variables file for editing in the default editor (determined by $EDITOR, defaulting to vim)",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.Vars(logger, c, varsPath)
		},
	}
)

func init() {
	rootCmd.AddCommand(varsCmd)
}
