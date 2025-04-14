package finder

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// setupTestDir creates a temporary test directory with the specified files
func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "finder-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

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
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestFindFiles tests the FindFiles function with various scenarios
func TestFindFiles(t *testing.T) {
	// Setup test directory
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

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
				"dir1", "dir2", "dir3", "dir1/file4.txt", "dir1/file5.go",
				"dir2/file6.json", "dir3/file7.yaml", "dir3/subdir",
				"dir3/subdir/file8.txt", "dir3/subdir/file9.go",
				"file1.txt", "file2.go", "file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Include only .txt files",
			includes: []string{"*.txt"},
			excludes: []string{},
			want:     []string{"file1.txt", "dir1/file4.txt", "dir3/subdir/file8.txt"},
			wantErr:  false,
		},
		{
			name:     "Include only .go files",
			includes: []string{"*.go"},
			excludes: []string{},
			want:     []string{"file2.go", "dir1/file5.go", "dir3/subdir/file9.go"},
			wantErr:  false,
		},
		{
			name:     "Exclude .txt files",
			includes: []string{},
			excludes: []string{"*.txt"},
			want: []string{
				"dir1", "dir2", "dir3", "dir1/file5.go",
				"dir2/file6.json", "dir3/file7.yaml", "dir3/subdir",
				"dir3/subdir/file9.go", "file2.go", "file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Include .go files but exclude dir1",
			includes: []string{"*.go"},
			excludes: []string{"dir1/*"},
			want:     []string{"file2.go", "dir3/subdir/file9.go"},
			wantErr:  false,
		},
		{
			name:     "Include all but exclude multiple patterns",
			includes: []string{},
			excludes: []string{"*.txt", "*.go"},
			want: []string{
				"dir1", "dir2", "dir3", "dir2/file6.json",
				"dir3/file7.yaml", "dir3/subdir", "file3.md",
			},
			wantErr: false,
		},
		{
			name:     "Directory pattern include",
			includes: []string{"dir1/*"},
			excludes: []string{},
			want:     []string{"dir1/file4.txt", "dir1/file5.go"},
			wantErr:  false,
		},
		{
			name:     "Directory pattern exclude",
			includes: []string{},
			excludes: []string{"dir3/*"},
			want: []string{
				"dir1", "dir2", "dir3", "dir1/file4.txt", "dir1/file5.go",
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

			// Sort both slices for comparison
			got = normalizeAndSort(got)
			tt.want = normalizeAndSort(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindFiles() = %v, want %v", got, tt.want)
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
