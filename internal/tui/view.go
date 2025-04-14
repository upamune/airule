package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the model to a string
func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.state {
	case StateNormal:
		// Render normal view with list and preview
		return renderNormalView(m)
	case StateCopying:
		// Render copying view
		return renderCopyingView(m)
	case StateCopyComplete:
		// Render copy complete view
		return renderCopyCompleteView(m)
	case StateError:
		// Render error view
		return renderErrorView(m)
	default:
		return errorStyle.Render("Unknown state")
	}
}

// renderNormalView renders the normal view with list and preview
func renderNormalView(m Model) string {
	// Get the list view and add selection indicators
	listView := addSelectionIndicators(m)

	// Get the preview view
	previewView := getPreviewView(m)

	// Style the views
	styledListView := defaultListStyle.Render(listView)
	styledPreviewView := defaultViewportStyle.Render(previewView)

	// Add title to the top
	title := titleStyle.Render("airule - Rule File Selector")

	// Add subtitle with instructions
	subtitle := subtitleStyle.Render("Select files to copy")

	// Combine the views horizontally
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, styledListView, styledPreviewView)

	// Add help text at the bottom
	helpText := helpStyle.Render("↑/↓: Navigate • Space: Select/Deselect • Enter: Copy • q: Quit")

	// Join everything vertically
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		subtitle,
		mainView,
		helpText)
}

// addSelectionIndicators adds selection indicators to the list view
func addSelectionIndicators(m Model) string {
	listView := m.list.View()
	lines := strings.Split(listView, "\n")

	for i := range lines {
		// Skip header lines
		if i < 2 {
			continue
		}

		// Calculate the index in the list model
		listIndex := i - 2
		if listIndex >= 0 && listIndex < len(m.list.Items()) {
			if _, selected := m.selected[listIndex]; selected {
				// Use styled checkbox for selected items
				lines[i] = checkboxCheckedStyle.Render("[✓]") + " " + lines[i]
			} else {
				// Use styled checkbox for unselected items
				lines[i] = checkboxUncheckedStyle.Render("[ ]") + " " + lines[i]
			}
		}
	}

	return strings.Join(lines, "\n")
}

// getPreviewView returns the preview view with error handling
func getPreviewView(m Model) string {
	if m.previewError != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading preview: %v", m.previewError))
	}
	return m.viewport.View()
}

// renderCopyingView renders the copying view
func renderCopyingView(m Model) string {
	// Count selected files
	selectedCount := len(m.selected)

	// Create a more informative copying message
	message := fmt.Sprintf("Copying %d file(s) from %s to %s...\n\nPlease wait.",
		selectedCount,
		m.cliArgs.From,
		m.cliArgs.To)

	return lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Copying Files"),
		warningStyle.Render(message))
}

// renderCopyCompleteView renders the copy complete view
func renderCopyCompleteView(m Model) string {
	if m.copyError != nil {
		return lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("Copy Failed"),
			errorStyle.Render(fmt.Sprintf("Error copying files: %v", m.copyError)),
			helpStyle.Render("Press q to quit."))
	}

	// Count copied files
	copiedCount := len(m.selected)

	// Create a success message with details
	message := fmt.Sprintf("%d file(s) copied successfully from %s to %s!",
		copiedCount,
		m.cliArgs.From,
		m.cliArgs.To)

	return lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Copy Complete"),
		successStyle.Render(message),
		helpStyle.Render("Press q to quit."))
}

// renderErrorView renders the error view
func renderErrorView(m Model) string {
	return lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render("Error"),
		errorStyle.Render(fmt.Sprintf("%v", m.err)),
		helpStyle.Render("Press q to quit."))
}
