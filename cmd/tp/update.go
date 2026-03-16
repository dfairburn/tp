package main

import (
	"fmt"
	"os"
	"os/exec"
	dbg "runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		gobin, err := exec.LookPath("go")
		if err != nil {
			return fmt.Errorf("go toolchain not found in PATH: %w\nInstall Go from https://go.dev/dl/ or download a release manually", err)
		}

		var modulePath string
		if info, ok := dbg.ReadBuildInfo(); ok && info.Main.Path != "" {
			modulePath = info.Main.Path + "/cmd/tp"
		} else {
			modulePath = "github.com/dfairburn/tp/cmd/tp"
		}

		currentVersion := getVersion()
		fmt.Printf("Current version: v%s\n", currentVersion)

		// ask for confirmation
		fmt.Print("Do you want to update tp to the latest version? (y/N): ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Update cancelled.")
			return nil
		}

		fmt.Println("Updating tp to latest version...")

		installCmd := exec.Command(gobin, "install", modulePath+"@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to update tp: %w\nYou may need to install Go from https://go.dev/dl/", err)
		}

		fmt.Println("Update complete! Restart your shell or run 'tp version' to verify the new version.")
		return nil
	},
}
