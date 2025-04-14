package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/upamune/airule/internal/cli"
	"github.com/upamune/airule/internal/copier"
	"github.com/upamune/airule/internal/finder"
	"github.com/upamune/airule/internal/preview"
)

// App represents the main application
type App struct {
	cliArgs cli.CLI
}

// NewApp creates a new App instance
func NewApp(cliArgs cli.CLI) *App {
	return &App{
		cliArgs: cliArgs,
	}
}

// Run executes the application
func (a *App) Run() error {
	// Find files based on include/exclude patterns
	files, err := finder.FindFiles(a.cliArgs.From, a.cliArgs.Include, a.cliArgs.Exclude)
	if err != nil {
		return fmt.Errorf("error finding files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found matching the criteria")
	}

	// Use go-fuzzyfinder to select files
	indices, err := fuzzyfinder.FindMulti(
		files,
		func(i int) string {
			return files[i]
		},
		fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
			if i == -1 {
				return "Select a file to preview its contents"
			}
			// Use the preview package to generate preview content
			previewContent, err := preview.GeneratePreview(a.cliArgs.From, files[i], width, height)
			if err != nil {
				return fmt.Sprintf("Error loading preview: %v", err)
			}
			return previewContent
		}),
		fuzzyfinder.WithPromptString("Select files to copy (Tab to select, Enter to confirm): "),
		fuzzyfinder.WithHeader("airule - Rule File Selector"),
	)

	// Handle cancellation (Esc key)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Operation cancelled")
			return nil
		}
		return fmt.Errorf("error selecting files: %w", err)
	}

	// No files selected
	if len(indices) == 0 {
		fmt.Println("No files selected")
		return nil
	}

	// Get the selected files
	selectedFiles := make([]string, len(indices))
	for i, idx := range indices {
		selectedFiles[i] = files[idx]
	}

	// Define styles for output
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63"))

	// Display selected files with styling
	title := titleStyle.Render(fmt.Sprintf("Selected %d file(s):", len(selectedFiles)))
	fmt.Println(title)

	for _, file := range selectedFiles {
		bullet := bulletStyle.Render("  • ")
		fmt.Printf("%s%s\n", bullet, file)
	}

	// Define path style
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Italic(true)

	// Confirm copy operation with styling
	fmt.Printf("\nCopying from %s to %s\n",
		pathStyle.Render(a.cliArgs.From),
		pathStyle.Render(a.cliArgs.To))

	fmt.Print("Proceed with copy? (y/n): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("203")).
			Bold(true)
		fmt.Println(cancelStyle.Render("Copy operation cancelled"))
		return nil
	}

	// Copy the selected files
	copyingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("105"))
	fmt.Println(copyingStyle.Render("Copying files..."))

	if err := copier.CopyFiles(a.cliArgs.From, a.cliArgs.To, selectedFiles); err != nil {
		return fmt.Errorf("error copying files: %w", err)
	}

	// Success message with styling
	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("42"))

	checkmark := successStyle.Render("✓")

	// Create a styled box for the success message
	messageBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(fmt.Sprintf("%s Successfully copied %d file(s) to %s",
			checkmark,
			len(selectedFiles),
			pathStyle.Render(a.cliArgs.To)))

	fmt.Println("\n" + messageBox)
	return nil
}
