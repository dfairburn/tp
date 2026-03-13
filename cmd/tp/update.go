package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

const modulePath = "github.com/dfairburn/tp/cmd/tp"

func init() {
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update tp to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		gobin, err := exec.LookPath("go")
		if err != nil {
			return fmt.Errorf("go toolchain not found in PATH: %w\nInstall Go from https://go.dev/dl/ or download a release manually", err)
		}

		fmt.Printf("Current version: %s\n", getVersion())
		fmt.Println("Updating tp to latest version...")

		install := exec.Command(gobin, "install", modulePath+"@latest")
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr

		if err := install.Run(); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Println("Update complete. Restart your shell or run 'tp version' to verify.")
		return nil
	},
}
