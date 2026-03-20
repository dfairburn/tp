package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	buildinfo "runtime/debug"
)

func getVersion() string {
	if info, ok := buildinfo.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}

	return "devel"
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
