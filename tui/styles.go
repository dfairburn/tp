package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palettes for dark and light backgrounds
var (
	darkPalette = struct {
		primaryColor   lipgloss.Color
		secondaryColor lipgloss.Color
		accentColor    lipgloss.Color
		errorColor     lipgloss.Color
		mutedColor     lipgloss.Color
		textColor      lipgloss.Color
		borderColor    lipgloss.Color
		highlightBg    lipgloss.Color
		methodColors   map[string]lipgloss.Color
		urlColor       lipgloss.Color
	}{
		primaryColor:   lipgloss.Color("#7C3AED"), // Purple
		secondaryColor: lipgloss.Color("#10B981"), // Green
		accentColor:    lipgloss.Color("#F59E0B"), // Orange
		errorColor:     lipgloss.Color("#EF4444"), // Red
		mutedColor:     lipgloss.Color("#6B7280"), // Gray
		textColor:      lipgloss.Color("#F9FAFB"), // White
		borderColor:    lipgloss.Color("#374151"), // Medium gray
		highlightBg:    lipgloss.Color("#374151"), // Selection background
		methodColors: map[string]lipgloss.Color{
			"GET":     lipgloss.Color("#10B981"), // Green
			"POST":    lipgloss.Color("#F59E0B"), // Orange
			"PUT":     lipgloss.Color("#3B82F6"), // Blue
			"PATCH":   lipgloss.Color("#8B5CF6"), // Purple
			"DELETE":  lipgloss.Color("#EF4444"), // Red
			"HEAD":    lipgloss.Color("#6B7280"), // Gray
			"OPTIONS": lipgloss.Color("#EC4899"), // Pink
		},
		urlColor: lipgloss.Color("#60A5FA"), // Light blue
	}

	lightPalette = struct {
		primaryColor   lipgloss.Color
		secondaryColor lipgloss.Color
		accentColor    lipgloss.Color
		errorColor     lipgloss.Color
		mutedColor     lipgloss.Color
		textColor      lipgloss.Color
		borderColor    lipgloss.Color
		highlightBg    lipgloss.Color
		methodColors   map[string]lipgloss.Color
		urlColor       lipgloss.Color
	}{
		primaryColor:   lipgloss.Color("#7C3AED"), // Purple
		secondaryColor: lipgloss.Color("#059669"), // Darker green
		accentColor:    lipgloss.Color("#D97706"), // Darker orange
		errorColor:     lipgloss.Color("#DC2626"), // Darker red
		mutedColor:     lipgloss.Color("#9CA3AF"), // Lighter gray
		textColor:      lipgloss.Color("#1F2937"), // Near-black
		borderColor:    lipgloss.Color("#D1D5DB"), // Light gray
		highlightBg:    lipgloss.Color("#E5E7EB"), // Light selection
		methodColors: map[string]lipgloss.Color{
			"GET":     lipgloss.Color("#059669"), // Green
			"POST":    lipgloss.Color("#D97706"), // Orange
			"PUT":     lipgloss.Color("#2563EB"), // Blue
			"PATCH":   lipgloss.Color("#7C3AED"), // Purple
			"DELETE":  lipgloss.Color("#DC2626"), // Red
			"HEAD":    lipgloss.Color("#9CA3AF"), // Gray
			"OPTIONS": lipgloss.Color("#DB2777"), // Pink
		},
		urlColor: lipgloss.Color("#2563EB"), // Blue
	}
)

// Selected palette colors
var (
	primaryColor   lipgloss.Color
	secondaryColor lipgloss.Color
	accentColor    lipgloss.Color
	errorColor     lipgloss.Color
	mutedColor     lipgloss.Color
	textColor      lipgloss.Color
	borderColor    lipgloss.Color
	highlightBg    lipgloss.Color
	methodColors   map[string]lipgloss.Color
	urlColor       lipgloss.Color
)

func init() {
	if lipgloss.HasDarkBackground() {
		primaryColor = darkPalette.primaryColor
		secondaryColor = darkPalette.secondaryColor
		accentColor = darkPalette.accentColor
		errorColor = darkPalette.errorColor
		mutedColor = darkPalette.mutedColor
		textColor = darkPalette.textColor
		borderColor = darkPalette.borderColor
		highlightBg = darkPalette.highlightBg
		methodColors = darkPalette.methodColors
		urlColor = darkPalette.urlColor
	} else {
		primaryColor = lightPalette.primaryColor
		secondaryColor = lightPalette.secondaryColor
		accentColor = lightPalette.accentColor
		errorColor = lightPalette.errorColor
		mutedColor = lightPalette.mutedColor
		textColor = lightPalette.textColor
		borderColor = lightPalette.borderColor
		highlightBg = lightPalette.highlightBg
		methodColors = lightPalette.methodColors
		urlColor = lightPalette.urlColor
	}
}

// Muted text style (for rendering muted text)
var mutedStyle = lipgloss.NewStyle().Foreground(mutedColor)

// Panel styles - initialized as functions to use correct colors after init()
func getPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

func getActivePanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(primaryColor).
		Padding(0, 1)
}

// Styles
var (
	// Base styles (used for future enhancements)
	_ = lipgloss.NewStyle().Foreground(textColor) // baseStyle

	// Panel styles (kept for backwards compatibility, but prefer functions above)
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 1).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// List styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Background(highlightBg).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Method badge styles
	methodBadgeStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1)

	// URL styles
	urlStyle = lipgloss.NewStyle().
			Foreground(urlColor).
			Bold(true)

	// Header styles
	headerKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	headerValueStyle = lipgloss.NewStyle().
				Foreground(textColor)

	// Status styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	// Response styles
	successStatusStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	errorStatusStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Input styles
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textColor).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 2)

	// Variable override styles
	overrideKeyStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)
)

// Helper function to get method color
func getMethodStyle(method string) lipgloss.Style {
	color, ok := methodColors[method]
	if !ok {
		color = mutedColor
	}
	return methodBadgeStyle.Foreground(color)
}

// Helper function to render status code with appropriate color
func renderStatusCode(code int) string {
	style := successStatusStyle
	if code >= 400 {
		style = errorStatusStyle
	} else if code >= 300 {
		style = lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	}
	return style.Render(fmt.Sprintf("%d", code))
}
