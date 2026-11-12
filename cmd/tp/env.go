package main

import (
	"github.com/dfairburn/tp/config"
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
			explicitPath := envFile
			if len(args) == 1 {
				explicitPath = args[0]
			}

			path, err := config.ResolveEnvironmentFile(logger, c.EnvironmentFile, env, explicitPath)
			if err != nil {
				return err
			}

			return handlers.Env(logger, path)
		},
	}
)

func init() {
	rootCmd.AddCommand(envCmd)
}
