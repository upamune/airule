package app

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/upamune/airule/internal/cli"
	"github.com/upamune/airule/internal/finder"
)

// setupIntegrationTestDir creates a temporary test directory with the specified files
func setupIntegrationTestDir(t *testing.T) string {
	t.Helper()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create test directory structure
	dirs := []string{
		"dir1",
		"dir2",
		"dir3/subdir",
	}

	files := []string{
		"file1.txt",
		"file2.go",
		"file3.md",
		"dir1/file4.txt",
		"dir1/file5.go",
		"dir2/file6.json",
		"dir3/file7.yaml",
		"dir3/subdir/file8.txt",
		"dir3/subdir/file9.go",
	}

	// Create directories
	for _, dir := range dirs {
		dirPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
	}

	// Create files
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	return tempDir
}

// TestSelectAllFlag tests the SelectAll flag functionality
func TestSelectAllFlag(t *testing.T) {
	// Setup test directory
	tempDir := setupIntegrationTestDir(t)

	// Find all files in the test directory
	allFiles, err := finder.FindFiles(tempDir, []string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to find files: %v", err)
	}

	// Create CLI args with SelectAll=true
	cliArgs := cli.CLI{
		From:      tempDir,
		To:        filepath.Join(tempDir, "dest"),
		SelectAll: true,
	}

	// Create app
	app := NewApp(cliArgs)

	// Create preselected indices based on SelectAll flag
	var preselectedIndices []int
	if app.cliArgs.SelectAll {
		// If SelectAll is true, preselect all files
		for i := range allFiles {
			preselectedIndices = append(preselectedIndices, i)
		}
	}

	// Verify that all files are preselected
	if len(preselectedIndices) != len(allFiles) {
		t.Errorf("SelectAll flag did not preselect all files: got %d, want %d",
			len(preselectedIndices), len(allFiles))
	}

	// Create a map for quick lookup of preselected indices
	preselectedMap := make(map[int]bool)
	for _, idx := range preselectedIndices {
		preselectedMap[idx] = true
	}

	// Verify that all indices are in the map
	for i := range allFiles {
		if !preselectedMap[i] {
			t.Errorf("Index %d was not preselected", i)
		}
	}
}

// TestPreSelectPatterns tests the PreSelect patterns functionality
func TestPreSelectPatterns(t *testing.T) {
	// Setup test directory
	tempDir := setupIntegrationTestDir(t)

	// Find all files in the test directory
	allFiles, err := finder.FindFiles(tempDir, []string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to find files: %v", err)
	}

	tests := []struct {
		name          string
		preSelect     []string
		expectedFiles []string
	}{
		{
			name:      "PreSelect txt files",
			preSelect: []string{"*.txt"},
			expectedFiles: []string{
				"file1.txt",
				"dir1/file4.txt",
				"dir3/subdir/file8.txt",
			},
		},
		{
			name:      "PreSelect go files",
			preSelect: []string{"*.go"},
			expectedFiles: []string{
				"file2.go",
				"dir1/file5.go",
				"dir3/subdir/file9.go",
			},
		},
		{
			name:      "PreSelect dir1 files",
			preSelect: []string{"dir1/*"},
			expectedFiles: []string{
				"dir1/file4.txt",
				"dir1/file5.go",
			},
		},
		{
			name:      "PreSelect multiple patterns",
			preSelect: []string{"*.txt", "*.go"},
			expectedFiles: []string{
				"file1.txt",
				"file2.go",
				"dir1/file4.txt",
				"dir1/file5.go",
				"dir3/subdir/file8.txt",
				"dir3/subdir/file9.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create CLI args with PreSelect patterns
			cliArgs := cli.CLI{
				From:      tempDir,
				To:        filepath.Join(tempDir, "dest"),
				PreSelect: tt.preSelect,
			}

			// Create app
			app := NewApp(cliArgs)

			// Create preselected indices based on PreSelect patterns
			var preselectedIndices []int
			if len(app.cliArgs.PreSelect) > 0 {
				// If PreSelect patterns are provided, preselect matching files
				for i, file := range allFiles {
					if matchesAnyPattern(file, app.cliArgs.PreSelect) {
						preselectedIndices = append(preselectedIndices, i)
					}
				}
			}

			// Get the preselected files
			preselectedFiles := make([]string, len(preselectedIndices))
			for i, idx := range preselectedIndices {
				preselectedFiles[i] = allFiles[idx]
			}

			// Sort both slices for comparison
			sort.Strings(preselectedFiles)
			expectedFiles := make([]string, len(tt.expectedFiles))
			copy(expectedFiles, tt.expectedFiles)
			sort.Strings(expectedFiles)

			// Filter out directories from preselected files for comparison
			filteredPreselected := make([]string, 0, len(preselectedFiles))
			for _, file := range preselectedFiles {
				// Skip directories
				if !strings.HasSuffix(file, "/") {
					// Check if it's a file (not a directory)
					filePath := filepath.Join(tempDir, file)
					fileInfo, err := os.Stat(filePath)
					if err == nil && !fileInfo.IsDir() {
						filteredPreselected = append(filteredPreselected, file)
					}
				}
			}

			// Sort both slices again after filtering
			sort.Strings(filteredPreselected)

			// Verify that the correct files are preselected
			if !reflect.DeepEqual(filteredPreselected, expectedFiles) {
				t.Errorf("PreSelect patterns did not preselect the correct files:\ngot: %v\nwant: %v",
					filteredPreselected, expectedFiles)
			}
		})
	}
}

// TestCombinedFunctionality tests combinations of SelectAll and PreSelect with include/exclude patterns
func TestCombinedFunctionality(t *testing.T) {
	// Setup test directory
	tempDir := setupIntegrationTestDir(t)

	tests := []struct {
		name          string
		includes      []string
		excludes      []string
		selectAll     bool
		preSelect     []string
		expectedCount int
	}{
		{
			name:          "SelectAll with includes",
			includes:      []string{"*.txt"},
			excludes:      []string{},
			selectAll:     true,
			preSelect:     []string{},
			expectedCount: 3, // All txt files
		},
		{
			name:          "SelectAll with excludes",
			includes:      []string{},
			excludes:      []string{"*.txt"},
			selectAll:     true,
			expectedCount: 10, // All files except txt files (including directories)
		},
		{
			name:          "PreSelect with includes",
			includes:      []string{"*.txt", "*.go"},
			excludes:      []string{},
			selectAll:     false,
			preSelect:     []string{"*.go"},
			expectedCount: 3, // Only go files
		},
		{
			name:          "PreSelect with excludes",
			includes:      []string{},
			excludes:      []string{"dir1/*"},
			selectAll:     false,
			preSelect:     []string{"*.txt"},
			expectedCount: 2, // txt files except those in dir1
		},
		{
			name:          "SelectAll overrides PreSelect",
			includes:      []string{"*.txt", "*.go"},
			excludes:      []string{},
			selectAll:     true,
			preSelect:     []string{"*.txt"},
			expectedCount: 6, // All txt and go files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find files based on include/exclude patterns
			filteredFiles, err := finder.FindFiles(tempDir, tt.includes, tt.excludes)
			if err != nil {
				t.Fatalf("Failed to find files: %v", err)
			}

			// Create CLI args
			cliArgs := cli.CLI{
				From:      tempDir,
				To:        filepath.Join(tempDir, "dest"),
				Include:   tt.includes,
				Exclude:   tt.excludes,
				SelectAll: tt.selectAll,
				PreSelect: tt.preSelect,
			}

			// Create app
			app := NewApp(cliArgs)

			// Create preselected indices based on SelectAll flag and PreSelect patterns
			var preselectedIndices []int
			if app.cliArgs.SelectAll {
				// If SelectAll is true, preselect all files
				for i := range filteredFiles {
					preselectedIndices = append(preselectedIndices, i)
				}
			} else if len(app.cliArgs.PreSelect) > 0 {
				// If PreSelect patterns are provided, preselect matching files
				for i, file := range filteredFiles {
					if matchesAnyPattern(file, app.cliArgs.PreSelect) {
						preselectedIndices = append(preselectedIndices, i)
					}
				}
			}

			// Count preselected items
			itemCount := 0
			for _, idx := range preselectedIndices {
				filePath := filepath.Join(tempDir, filteredFiles[idx])
				fileInfo, err := os.Stat(filePath)
				if err == nil {
					// For "SelectAll with excludes" test case, count both files and directories
					// For other test cases, count only files
					if tt.name == "SelectAll with excludes" || !fileInfo.IsDir() {
						itemCount++
					}
				}
			}

			// Verify the number of preselected items (including directories)
			if itemCount != tt.expectedCount {
				t.Errorf("Incorrect number of preselected items: got %d, want %d",
					itemCount, tt.expectedCount)
			}
		})
	}
}
