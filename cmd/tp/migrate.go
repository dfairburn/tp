package main

import (
	"github.com/dfairburn/tp/handlers"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:                   "migrate [insomnia|postman]",
	Short:                 "Migrate from other API dev tools",
	Long:                  "",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"insomnia"},
	Args:                  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "insomnia":
			return handlers.HandleInsomniaMigration(args[1])
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
