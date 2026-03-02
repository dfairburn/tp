package main

import (
	"github.com/dfairburn/tp/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI",
	Long: `Launch an interactive terminal user interface for managing and executing HTTP request templates.

The TUI provides a Postman-like experience with:
  - Template browser with folder navigation
  - Request preview with headers, body, and parameters
  - Response viewer with JSON formatting
  - Variable override management

Keyboard Shortcuts:
  Navigation:
    tab/shift+tab  - Switch between panels
    1/2/3/4        - Jump to specific panel
    j/k or arrows  - Move up/down in lists
    enter          - Select template / expand folder
    space          - Toggle folder expansion

  Actions:
    x, ctrl+enter  - Execute selected request
    /              - Search templates
    o              - Add variable override
    C              - Clear all overrides
    r              - Refresh template list
    e              - Edit selected template in $EDITOR
    n              - Create new template

  View:
    h              - Show headers tab
    b              - Show body tab
    p              - Show params tab (request only)
    ?              - Toggle help

  General:
    q, ctrl+c      - Quit
    esc            - Exit current mode
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run(logger, c, envFile)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
