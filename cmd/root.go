package cmd

import (
  "errors"

	"github.com/spf13/cobra"
)

var (
  // global flags
  rootCmd = &cobra.Command{
    Use:   "tp",
    Short: "tp is a configurable api client",
    Long:  `some text about tp`,
    RunE: func(cmd *cobra.Command, args []string) error {
      return errors.New("No subcommand given")
    },
  }
)

func Execute() error {
  return rootCmd.Execute()
}

func init() {
  handleFlags()
}

func handleFlags() {}
