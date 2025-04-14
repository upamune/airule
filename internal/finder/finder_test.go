package finder

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// setupTestDir creates a temporary test directory with the specified files
func setupTestDir(t *testing.T) string {
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

	// Return cleanup function
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

// TestFindFiles tests the FindFiles function with various scenarios
func TestFindFiles(t *testing.T) {
	// Setup test directory
	tempDir := setupTestDir(t)

	// Helper function to normalize and sort file paths for comparison
	normalizeAndSort := func(files []string) []string {
		sort.Strings(files)
		return files
	}

	tests := []struct {
		name     string
		includes []string
		excludes []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "No patterns - should find all files",
			includes: []string{},
			excludes: []string{},
			want: []string{
				"dir1", "dir2", "dir3", "dir3/subdir", "dir1/file4.txt", "dir1/file5.go",
				"dir2/file6.json", "dir3/file7.yaml", "dir3/subdir/file8.txt",
				"dir3/subdir/file9.go", "file1.txt", "file2.go", "file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Include only .txt files",
			includes: []string{"*.txt"},
			excludes: []string{},
			want:     []string{"dir1", "dir3", "dir3/subdir", "file1.txt", "dir1/file4.txt", "dir3/subdir/file8.txt"},
			wantErr:  false,
		},
		{
			name:     "Include only .go files",
			includes: []string{"*.go"},
			excludes: []string{},
			want:     []string{"dir1", "dir3", "dir3/subdir", "file2.go", "dir1/file5.go", "dir3/subdir/file9.go"},
			wantErr:  false,
		},
		{
			name:     "Exclude .txt files",
			includes: []string{},
			excludes: []string{"*.txt"},
			want: []string{
				"dir1", "dir2", "dir3", "dir3/subdir", "dir1/file5.go",
				"dir2/file6.json", "dir3/file7.yaml", "dir3/subdir/file9.go",
				"file2.go", "file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Include .go files but exclude dir1",
			includes: []string{"*.go"},
			excludes: []string{"dir1/*"},
			want:     []string{"dir3", "dir3/subdir", "file2.go", "dir3/subdir/file9.go"},
			wantErr:  false,
		},
		{
			name:     "Include all but exclude multiple patterns",
			includes: []string{},
			excludes: []string{"*.txt", "*.go"},
			want: []string{
				"dir2",
				"dir3",
				"dir2/file6.json",
				"dir3/file7.yaml",
				"file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Directory pattern include",
			includes: []string{"dir1/*"},
			excludes: []string{},
			want:     []string{"dir1", "dir1/file4.txt", "dir1/file5.go"},
			wantErr:  false,
		},
		{
			name:     "Directory pattern exclude",
			includes: []string{},
			excludes: []string{"dir3/*"},
			want: []string{
				"dir1", "dir2", "dir1/file4.txt", "dir1/file5.go",
				"dir2/file6.json", "file1.txt", "file2.go", "file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Non-existent directory",
			includes: []string{},
			excludes: []string{},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "Verify reverse sorting behavior",
			includes: []string{},
			excludes: []string{},
			want: []string{
				"file3.md", "file2.go", "file1.txt", "dir3/subdir/file9.go",
				"dir3/subdir/file8.txt", "dir3/subdir", "dir3/file7.yaml", "dir3",
				"dir2/file6.json", "dir2", "dir1/file5.go", "dir1/file4.txt", "dir1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootDir := tempDir
			if tt.name == "Non-existent directory" {
				rootDir = filepath.Join(tempDir, "non-existent")
			}

			got, err := FindFiles(rootDir, tt.includes, tt.excludes)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("FindFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.name == "Verify reverse sorting behavior" {
				// For the reverse sorting test, we don't sort the actual results
				// to verify they come back already reverse sorted

				// Create a reverse-sorted copy of the expected results
				reverseSorted := make([]string, len(tt.want))
				copy(reverseSorted, tt.want)
				sort.Sort(sort.Reverse(sort.StringSlice(reverseSorted)))

				// First check if the results are already reverse sorted
				reverseSortedGot := make([]string, len(got))
				copy(reverseSortedGot, got)
				sort.Sort(sort.Reverse(sort.StringSlice(reverseSortedGot)))

				if !reflect.DeepEqual(got, reverseSortedGot) {
					t.Errorf("FindFiles() results are not reverse sorted: %v", got)
				}

				// Then check if they match the expected values
				if !reflect.DeepEqual(got, reverseSorted) {
					t.Errorf("FindFiles() = %v, want %v", got, reverseSorted)
				}
			} else {
				// For other tests, sort both slices for comparison
				got = normalizeAndSort(got)
				tt.want = normalizeAndSort(tt.want)

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("FindFiles() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestShouldInclude tests the shouldInclude function with various patterns
func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		includes []string
		excludes []string
		want     bool
	}{
		{
			name:     "No patterns - should include",
			path:     "file.txt",
			includes: []string{},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "Match include pattern",
			path:     "file.txt",
			includes: []string{"*.txt"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "No match include pattern",
			path:     "file.txt",
			includes: []string{"*.go"},
			excludes: []string{},
			want:     false,
		},
		{
			name:     "Match exclude pattern",
			path:     "file.txt",
			includes: []string{},
			excludes: []string{"*.txt"},
			want:     false,
		},
		{
			name:     "Match include but also exclude - exclude wins",
			path:     "file.txt",
			includes: []string{"*.txt"},
			excludes: []string{"*.txt"},
			want:     false,
		},
		{
			name:     "Match include and no match exclude",
			path:     "file.txt",
			includes: []string{"*.txt"},
			excludes: []string{"*.go"},
			want:     true,
		},
		{
			name:     "Directory pattern include match",
			path:     "dir/file.txt",
			includes: []string{"dir/*"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "Directory pattern include no match",
			path:     "other/file.txt",
			includes: []string{"dir/*"},
			excludes: []string{},
			want:     false,
		},
		{
			name:     "Directory pattern exclude match",
			path:     "dir/file.txt",
			includes: []string{},
			excludes: []string{"dir/*"},
			want:     false,
		},
		{
			name:     "Directory pattern exclude no match",
			path:     "other/file.txt",
			includes: []string{},
			excludes: []string{"dir/*"},
			want:     true,
		},
		{
			name:     "Nested directory match",
			path:     "dir/subdir/file.txt",
			includes: []string{"dir/*"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "Edge case - empty path",
			path:     "",
			includes: []string{"*.txt"},
			excludes: []string{},
			want:     false,
		},
		{
			name:     "Edge case - exact file match",
			path:     "specific.file",
			includes: []string{"specific.file"},
			excludes: []string{},
			want:     true,
		},
		{
			name:     "Edge case - pattern with special characters",
			path:     "file-with-dashes.txt",
			includes: []string{"*-*-*.txt"},
			excludes: []string{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldInclude(tt.path, tt.includes, tt.excludes)
			if got != tt.want {
				t.Errorf("shouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}
