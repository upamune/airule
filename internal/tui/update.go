package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/upamune/airule/internal/copier"
	"github.com/upamune/airule/internal/preview"
)

// CopyCompleteMsg is the message sent when copying is complete
type CopyCompleteMsg struct {
	Error error
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height

		// Calculate dimensions for list and viewport
		listWidth := (m.width * 2) / 5           // Changed from 1/3 to 2/5
		viewportWidth := m.width - listWidth - 4 // Account for borders and padding
		viewportHeight := m.height - 4           // Account for borders and help text

		// Update list dimensions
		m.list.SetWidth(listWidth)
		m.list.SetHeight(m.height - 2) // Account for help text

		// Update viewport dimensions
		m.viewport.Width = viewportWidth
		m.viewport.Height = viewportHeight

		return m, nil

	case tea.KeyMsg:
		// Handle key presses
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			// Quit
			return m, tea.Quit

		case "ctrl+n":
			// Map Ctrl+N to down arrow functionality
			if m.state == StateNormal {
				// Simulate down arrow key press
				downMsg := tea.KeyMsg{Type: tea.KeyDown}
				return m.Update(downMsg)
			}
			return m, nil

		case "ctrl+p":
			// Map Ctrl+P to up arrow functionality
			if m.state == StateNormal {
				// Simulate up arrow key press
				upMsg := tea.KeyMsg{Type: tea.KeyUp}
				return m.Update(upMsg)
			}
			return m, nil

		case " ":
			// Toggle selection of current item
			if m.state != StateNormal {
				return m, nil
			}

			index := m.list.Index()
			if _, selected := m.selected[index]; selected {
				// Deselect
				delete(m.selected, index)
			} else {
				// Select
				m.selected[index] = true
			}
			return m, nil

		case "enter":
			// Copy selected files
			if m.state != StateNormal || len(m.selected) == 0 {
				return m, nil
			}

			// Get selected files
			var selectedFiles []string
			for i := range m.selected {
				if i < len(m.list.Items()) {
					item := m.list.Items()[i].(fileItem)
					selectedFiles = append(selectedFiles, item.path)
				}
			}

			// Update state and start copying
			m.state = StateCopying
			return m, copyFilesCmd(m.cliArgs.From, m.cliArgs.To, selectedFiles)
		}

	case preview.PreviewLoadedMsg:
		// Handle preview loaded message
		if msg.Err != nil {
			m.previewError = msg.Err
		} else {
			m.previewError = nil
			m.viewport.SetContent(msg.Content)
			m.viewport.GotoTop()
		}
		return m, nil

	case CopyCompleteMsg:
		// Handle copy complete message
		m.state = StateCopyComplete
		m.copyError = msg.Error
		return m, nil
	}

	// Handle list updates
	if m.state == StateNormal {
		newListModel, cmd := m.list.Update(msg)
		m.list = newListModel
		cmds = append(cmds, cmd)

		// If the selected item changed, load its preview
		if newIndex := m.list.Index(); newIndex != -1 && newIndex < len(m.list.Items()) {
			item := m.list.Items()[newIndex].(fileItem)
			cmds = append(cmds, preview.LoadPreview(m.cliArgs.From, item.path))
		}
	}

	// Handle viewport updates
	if m.state == StateNormal {
		newViewport, cmd := m.viewport.Update(msg)
		m.viewport = newViewport
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// copyFilesCmd returns a command that copies files and sends a CopyCompleteMsg
func copyFilesCmd(fromDir, toDir string, files []string) tea.Cmd {
	return func() tea.Msg {
		err := copier.CopyFiles(fromDir, toDir, files)
		return CopyCompleteMsg{
			Error: err,
		}
	}
}
