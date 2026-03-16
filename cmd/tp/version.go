package main

import (
	"fmt"
	dbg "runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// -ldflags sets this at build time, if not then falls back to module version from go install
var version = "dev"

func getVersion() string {
	if version != "dev" {
		return strings.TrimPrefix(version, "v")
	}

	if info, ok := dbg.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}

	return "dev"
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", getVersion())
	},
}
