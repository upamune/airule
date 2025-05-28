package app

import (
	"fmt"
	"path/filepath"
	"strings"

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

// matchesAnyPattern checks if a file path matches any of the provided patterns
func matchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		// Match against the full path or just the basename if the pattern doesn't contain a separator
		base := filepath.Base(filePath)
		matchPath, _ := filepath.Match(pattern, filePath)
		matchBase := false
		if !strings.Contains(pattern, string(filepath.Separator)) {
			matchBase, _ = filepath.Match(pattern, base)
		}
		if matchPath || matchBase {
			return true
		}

		// Handle directory patterns specifically (e.g., "dir/*" or "dir/**")
		if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
			dirPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "/")
			// Ensure dirPattern is not empty and path actually starts with it + separator
			if dirPattern != "" && strings.HasPrefix(filePath, dirPattern+string(filepath.Separator)) {
				return true
			}
			// Also handle case where the pattern *is* the directory path itself
			if filePath == dirPattern {
				return true
			}
		}
	}
	return false
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

	// Create preselected indices based on SelectAll flag and PreSelect patterns
	var preselectedIndices []int
	if a.cliArgs.SelectAll {
		// If SelectAll is true, preselect all files
		for i := range files {
			preselectedIndices = append(preselectedIndices, i)
		}
	} else if len(a.cliArgs.PreSelect) > 0 {
		// If PreSelect patterns are provided, preselect matching files
		for i, file := range files {
			if matchesAnyPattern(file, a.cliArgs.PreSelect) {
				preselectedIndices = append(preselectedIndices, i)
			}
		}
	}

	// Create a map for quick lookup of preselected indices
	preselectedMap := make(map[int]bool)
	for _, idx := range preselectedIndices {
		preselectedMap[idx] = true
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
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionTop),
		fuzzyfinder.WithPreselected(func(i int) bool {
			return preselectedMap[i]
		}),
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

	if err := copier.CopyFiles(a.cliArgs.From, a.cliArgs.To, selectedFiles, a.cliArgs.Clean, a.cliArgs.CleanExclude); err != nil {
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
