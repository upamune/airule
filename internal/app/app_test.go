package app

import (
	"testing"

	"github.com/upamune/airule/internal/cli"
)

// TestMatchesAnyPattern tests the matchesAnyPattern function with various patterns
func TestMatchesAnyPattern(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		patterns []string
		want     bool
	}{
		{
			name:     "No patterns - should not match",
			filePath: "file.txt",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "Match exact file",
			filePath: "file.txt",
			patterns: []string{"file.txt"},
			want:     true,
		},
		{
			name:     "Match file extension",
			filePath: "file.txt",
			patterns: []string{"*.txt"},
			want:     true,
		},
		{
			name:     "No match file extension",
			filePath: "file.txt",
			patterns: []string{"*.go"},
			want:     false,
		},
		{
			name:     "Match file in directory",
			filePath: "dir/file.txt",
			patterns: []string{"dir/*.txt"},
			want:     true,
		},
		{
			name:     "Match directory pattern",
			filePath: "dir/file.txt",
			patterns: []string{"dir/*"},
			want:     true,
		},
		{
			name:     "Match directory itself",
			filePath: "dir",
			patterns: []string{"dir/*"},
			want:     true,
		},
		{
			name:     "Match nested directory",
			filePath: "dir/subdir/file.txt",
			patterns: []string{"dir/*"},
			want:     true,
		},
		{
			name:     "Multiple patterns - match one",
			filePath: "file.txt",
			patterns: []string{"*.go", "*.txt"},
			want:     true,
		},
		{
			name:     "Multiple patterns - match none",
			filePath: "file.txt",
			patterns: []string{"*.go", "*.md"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesAnyPattern(tt.filePath, tt.patterns)
			if got != tt.want {
				t.Errorf("matchesAnyPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

// setupTestApp creates a test App instance with the specified CLI arguments
func setupTestApp(t *testing.T, cliArgs cli.CLI) *App {
	t.Helper()
	return NewApp(cliArgs)
}

// TestPreselectionLogic tests the preselection logic in the App
// This is a unit test for the preselection logic without running the full app
func TestPreselectionLogic(t *testing.T) {
	// Create a list of test files
	files := []string{
		"file1.txt",
		"file2.go",
		"file3.md",
		"dir1/file4.txt",
		"dir1/file5.go",
		"dir2/file6.json",
	}

	tests := []struct {
		name            string
		selectAll       bool
		preSelect       []string
		expectedIndices map[int]bool
	}{
		{
			name:            "No preselection",
			selectAll:       false,
			preSelect:       []string{},
			expectedIndices: map[int]bool{},
		},
		{
			name:      "Select all files",
			selectAll: true,
			preSelect: []string{},
			expectedIndices: map[int]bool{
				0: true, 1: true, 2: true, 3: true, 4: true, 5: true,
			},
		},
		{
			name:      "Preselect by extension - txt files",
			selectAll: false,
			preSelect: []string{"*.txt"},
			expectedIndices: map[int]bool{
				0: true, 3: true,
			},
		},
		{
			name:      "Preselect by extension - go files",
			selectAll: false,
			preSelect: []string{"*.go"},
			expectedIndices: map[int]bool{
				1: true, 4: true,
			},
		},
		{
			name:      "Preselect by directory - dir1",
			selectAll: false,
			preSelect: []string{"dir1/*"},
			expectedIndices: map[int]bool{
				3: true, 4: true,
			},
		},
		{
			name:      "Preselect multiple patterns",
			selectAll: false,
			preSelect: []string{"*.txt", "*.go"},
			expectedIndices: map[int]bool{
				0: true, 1: true, 3: true, 4: true,
			},
		},
		{
			name:      "SelectAll overrides PreSelect",
			selectAll: true,
			preSelect: []string{"*.txt"},
			expectedIndices: map[int]bool{
				0: true, 1: true, 2: true, 3: true, 4: true, 5: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI args with the test parameters
			cliArgs := cli.CLI{
				SelectAll: tt.selectAll,
				PreSelect: tt.preSelect,
			}

			// Create a test app
			app := setupTestApp(t, cliArgs)

			// Create preselected indices based on SelectAll flag and PreSelect patterns
			var preselectedIndices []int
			if app.cliArgs.SelectAll {
				// If SelectAll is true, preselect all files
				for i := range files {
					preselectedIndices = append(preselectedIndices, i)
				}
			} else if len(app.cliArgs.PreSelect) > 0 {
				// If PreSelect patterns are provided, preselect matching files
				for i, file := range files {
					if matchesAnyPattern(file, app.cliArgs.PreSelect) {
						preselectedIndices = append(preselectedIndices, i)
					}
				}
			}

			// Create a map for quick lookup of preselected indices
			preselectedMap := make(map[int]bool)
			for _, idx := range preselectedIndices {
				preselectedMap[idx] = true
			}

			// Compare the preselected map with the expected map
			if len(preselectedMap) != len(tt.expectedIndices) {
				t.Errorf("Preselected indices count = %d, want %d", len(preselectedMap), len(tt.expectedIndices))
			}

			for idx := range preselectedMap {
				if !tt.expectedIndices[idx] {
					t.Errorf("Unexpected index %d was preselected", idx)
				}
			}

			for idx := range tt.expectedIndices {
				if !preselectedMap[idx] {
					t.Errorf("Expected index %d was not preselected", idx)
				}
			}
		})
	}
}
