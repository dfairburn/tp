package main

import (
	"errors"
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
)

var (
	// local flags
	newCmd = &cobra.Command{
		Use:   "open",
		Short: "Open creates a new template or opens an existing one",
		Long:  "Open creates a new template, or opens an existing one and loads it into your configured editor (configured by $EDITOR, default vim)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("need new path to be returned for new template")
			}

			filename := args[0]
			return handlers.Open(logger, c.TemplatesDirectoryPath, filename)
		},
	}
)

func init() {
	newCmd.Flags().StringSliceVarP(&overrides, "overrides", "o", []string{}, overrideUsage)

	rootCmd.AddCommand(newCmd)
}
