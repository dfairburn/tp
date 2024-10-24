package main

import (
	"github.com/spf13/cobra"
)

var (
	// local flags
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init sets up the default directory structure and config",
		Long:  "Init sets up the default directory structure and config",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
}
