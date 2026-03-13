package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// -ldflags sets this to Vx.x.x at build time
var version = "dev"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of tp",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", version)
	},
}
