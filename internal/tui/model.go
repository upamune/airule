package tui

import (
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/upamune/airule/internal/cli"
	"github.com/upamune/airule/internal/finder"
	"github.com/upamune/airule/internal/preview"
)

// truncateWithEllipsis truncates a string if it's longer than maxLength
// and adds an ellipsis at the end
func truncateWithEllipsis(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// UIState represents the current state of the UI
type UIState int

const (
	// StateNormal is the default state
	StateNormal UIState = iota
	// StateCopying is the state when files are being copied
	StateCopying
	// StateCopyComplete is the state when copying is complete
	StateCopyComplete
	// StateError is the state when an error occurs
	StateError
)

// fileItem represents a file in the list
type fileItem struct {
	path     string
	selected bool
}

// FilterValue implements list.Item interface
func (i fileItem) FilterValue() string {
	return i.path
}

// Title implements list.Item interface
func (i fileItem) Title() string {
	// Get the base filename without the directory path
	filename := filepath.Base(i.path)
	// Get the directory path
	dir := filepath.Dir(i.path)
	if dir == "." {
		dir = ""
	} else {
		dir = dir + "/"
	}

	// Truncate the filename if it's too long
	maxFilenameLength := 30
	truncatedFilename := truncateWithEllipsis(filename, maxFilenameLength)

	// If we have a directory path, show it with the truncated filename
	if dir != "" {
		return dir + truncatedFilename
	}
	return truncatedFilename
}

// Description implements list.Item interface
func (i fileItem) Description() string {
	return ""
}

// Model represents the TUI model
type Model struct {
	cliArgs      cli.CLI
	list         list.Model
	viewport     viewport.Model
	selected     map[int]bool
	state        UIState
	err          error
	copyError    error
	previewError error
	width        int
	height       int
}

// InitialModel creates a new model with initial values
func InitialModel(cliArgs cli.CLI) Model {
	// Find files
	files, err := finder.FindFiles(cliArgs.From, cliArgs.Include, cliArgs.Exclude)
	if err != nil {
		return Model{
			cliArgs: cliArgs,
			state:   StateError,
			err:     err,
		}
	}

	// Create list items
	items := make([]list.Item, len(files))
	for i, file := range files {
		items[i] = fileItem{path: file}
	}

	// Initialize list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	// Initialize viewport for preview
	vp := viewport.New(0, 0)
	vp.Style = defaultViewportStyle
	vp.SetContent("Select a file to preview its contents")
	vp.Style = defaultViewportStyle

	// Initialize selected map
	selected := make(map[int]bool)

	return Model{
		cliArgs:  cliArgs,
		list:     l,
		viewport: vp,
		selected: selected,
		state:    StateNormal,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// If there are files, load the preview of the first one
	if len(m.list.Items()) > 0 {
		item := m.list.Items()[0].(fileItem)
		return preview.LoadPreview(m.cliArgs.From, item.path)
	}
	return nil
}

// Error returns the current error
func (m Model) Error() error {
	if m.err != nil {
		return m.err
	}
	if m.copyError != nil {
		return m.copyError
	}
	if m.previewError != nil {
		return m.previewError
	}
	return nil
}
