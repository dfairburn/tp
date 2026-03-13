package main

import (
	"fmt"
	dbg "runtime/debug"

	"github.com/spf13/cobra"
)

// -ldflags sets this at build time; falls back to module version from go install
var version = "dev"

func getVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := dbg.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of tp",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", getVersion())
	},
}
