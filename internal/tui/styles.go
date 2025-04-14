package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - using adaptive colors that work well in both light and dark terminal themes
var (
	primaryColor    = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#9D76FF"} // Purple
	secondaryColor  = lipgloss.AdaptiveColor{Light: "#5A5A5A", Dark: "#BBBBBB"} // Gray
	accentColor     = lipgloss.AdaptiveColor{Light: "#3B82F6", Dark: "#60A5FA"} // Blue
	successColor    = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"} // Green
	errorColor      = lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"} // Red
	warningColor    = lipgloss.AdaptiveColor{Light: "#F59E0B", Dark: "#FBBF24"} // Amber
	infoColor       = lipgloss.AdaptiveColor{Light: "#6366F1", Dark: "#818CF8"} // Indigo
	backgroundColor = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1F2937"} // White/Dark Gray
	textColor       = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"} // Dark Gray/White
)

var (
	// Styles for different UI elements

	// titleStyle - For main application title and section headers
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(primaryColor).
			Padding(0, 1).
			MarginBottom(1)

	// subtitleStyle - For secondary headers
	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// errorStyle - For error messages
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// successStyle - For success messages
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	// warningStyle - For warning messages
	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	// infoStyle - For informational messages
	infoStyle = lipgloss.NewStyle().
			Foreground(infoColor)

	// helpStyle - For help text and keyboard shortcuts
	helpStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// selectedItemStyle - For highlighting selected items
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true)

	// Default viewport style - For content preview
	defaultViewportStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2)

	// List style - For file list
	defaultListStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2)

	// Checkbox styles - For selection indicators
	checkboxCheckedStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	checkboxUncheckedStyle = lipgloss.NewStyle().
				Foreground(secondaryColor)
)
