package copier

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// setupTestDir creates a temporary test directory with both hidden and non-hidden files
func setupTestDir(t *testing.T) (string, string) {
	t.Helper()

	// Create a temporary directory
	tempDir := t.TempDir()

	// Create source directory
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Create destination directory
	dstDir := filepath.Join(tempDir, "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}

	// Create source files
	srcFiles := []string{
		"file1.txt",
		"file2.go",
		".hidden1",
		"dir1/file3.txt",
		"dir1/.hidden2",
		"dir2/file4.go",
	}

	// Create source directories and files
	for _, file := range srcFiles {
		filePath := filepath.Join(srcDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("source content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Create destination files (including hidden files that should be preserved)
	dstFiles := []string{
		"old_file.txt",
		".hidden_preserve1",
		"dir1/.hidden_preserve2",
		"dir2/old_file.go",
	}

	// Create destination directories and files
	for _, file := range dstFiles {
		filePath := filepath.Join(dstDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("destination content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	return srcDir, dstDir
}

// listFiles returns a sorted list of all files in the given directory
func listFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == dir {
			return nil
		}

		// Get the relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

// TestCopyFilesPreservesHiddenFiles tests that the CopyFiles function correctly
// preserves hidden files in the destination directory
func TestCopyFilesPreservesHiddenFiles(t *testing.T) {
	// Setup test directories
	srcDir, dstDir := setupTestDir(t)

	// Files to copy (all non-hidden files from source)
	filesToCopy := []string{
		"file1.txt",
		"file2.go",
		"dir1/file3.txt",
		"dir2/file4.go",
	}

	// Perform the copy operation
	err := CopyFiles(srcDir, dstDir, filesToCopy, true, nil) // Use cleanDest=true
	if err != nil {
		t.Fatalf("CopyFiles failed: %v", err)
	}

	// List all files in the destination directory after copy
	dstFiles, err := listFiles(dstDir)
	if err != nil {
		t.Fatalf("Failed to list destination files: %v", err)
	}

	// Expected files in destination after copy:
	// 1. All copied non-hidden files from source
	// 2. All hidden files that were already in destination (including nested)
	expectedFiles := []string{
		".hidden_preserve1",
		"dir1",                   // Directory containing preserved hidden file and copied file
		"dir1/.hidden_preserve2", // Preserved by clearDestinationDir
		"dir1/file3.txt",         // Copied
		"dir2",                   // Directory containing copied file
		"dir2/file4.go",          // Copied
		"file1.txt",              // Copied
		"file2.go",               // Copied
	}
	sort.Strings(expectedFiles)

	// Verify the destination directory contains the expected files
	if !reflect.DeepEqual(dstFiles, expectedFiles) {
		t.Errorf("Destination directory has incorrect files after CopyFiles(cleanDest=true): Got:  %v Want: %v", dstFiles, expectedFiles)
	}

	// Verify content of copied files
	for _, file := range filesToCopy {
		dstPath := filepath.Join(dstDir, file)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", dstPath, err)
			continue
		}

		if string(content) != "source content" {
			t.Errorf("File %s has incorrect content: got %q, want %q", file, string(content), "source content")
		}
	}

	// Verify content of preserved hidden files
	hiddenFiles := []string{
		".hidden_preserve1",
		"dir1/.hidden_preserve2", // Should be preserved
	}

	for _, file := range hiddenFiles {
		dstPath := filepath.Join(dstDir, file)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read preserved hidden file %s: %v", dstPath, err)
			continue
		}

		if string(content) != "destination content" {
			t.Errorf("Hidden file %s has incorrect content: got %q, want %q", file, string(content), "destination content")
		}
	}
}

// TestClearDestinationDir tests the clearDestinationDir function directly
func TestClearDestinationDir(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create files in the temporary directory
	files := []string{
		"regular_file.txt",
		".hidden_file",
		"dir1/nested_file.txt",
		"dir1/.hidden_nested_file",
		".hidden_dir/file.txt",
	}

	// Create directories and files
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Call clearDestinationDir
	err := clearDestinationDir(tempDir, nil)
	if err != nil {
		t.Fatalf("clearDestinationDir failed: %v", err)
	}

	// List remaining files
	remainingFiles, err := listFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to list remaining files: %v", err)
	}

	// Expected remaining files (only hidden files/dirs and dirs containing hidden files)
	expectedFiles := []string{
		".hidden_dir",
		".hidden_dir/file.txt",
		".hidden_file",
		"dir1",                     // Directory containing preserved hidden file
		"dir1/.hidden_nested_file", // Preserved hidden file
	}
	sort.Strings(expectedFiles)

	// Verify only hidden files and directories remain
	if !reflect.DeepEqual(remainingFiles, expectedFiles) {
		t.Errorf("clearDestinationDir did not preserve hidden files correctly: Got:  %v Want: %v", remainingFiles, expectedFiles)
	}
}

// TestCopyFilesWithoutCleaning tests that the CopyFiles function correctly
// preserves existing files in the destination directory when cleanDest is false
func TestCopyFilesWithoutCleaning(t *testing.T) {
	// Setup test directories
	srcDir, dstDir := setupTestDir(t)

	// Files to copy (all non-hidden files from source)
	filesToCopy := []string{
		"file1.txt",
		"file2.go",
		"dir1/file3.txt",
		"dir2/file4.go",
	}

	// First, create some additional files in the destination that should be preserved
	// when cleanDest is false
	preserveFiles := []string{
		"preserve_file1.txt",
		"preserve_file2.go",
		"preserve_dir/preserve_file3.txt",
	}

	for _, file := range preserveFiles {
		filePath := filepath.Join(dstDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("preserved content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Perform the copy operation with cleanDest=false
	err := CopyFiles(srcDir, dstDir, filesToCopy, false, nil)
	if err != nil {
		t.Fatalf("CopyFiles failed: %v", err)
	}

	// List all files in the destination directory after copy
	dstFiles, err := listFiles(dstDir)
	if err != nil {
		t.Fatalf("Failed to list destination files: %v", err)
	}

	// Expected files in destination after copy:
	// 1. All copied non-hidden files from source
	// 2. All files that were already in destination (including hidden and non-hidden)
	// 3. All preserved files that were added before copying
	expectedFilesOnly := []string{
		".hidden_preserve1",               // Original hidden file
		"dir1/.hidden_preserve2",          // Original nested hidden file
		"dir1/file3.txt",                  // Copied
		"dir2/file4.go",                   // Copied
		"dir2/old_file.go",                // Original file
		"file1.txt",                       // Copied
		"file2.go",                        // Copied
		"old_file.txt",                    // Original file
		"preserve_dir/preserve_file3.txt", // Added preserved file
		"preserve_file1.txt",              // Added preserved file
		"preserve_file2.go",               // Added preserved file
	}
	sort.Strings(expectedFilesOnly)

	// Check that all the expected files exist in the destination
	// Extract only files from the actual list for precise comparison
	actualFilesOnly := []string{}
	for _, f := range dstFiles {
		// Construct full path to check if it's a directory
		fullPath := filepath.Join(dstDir, f)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			actualFilesOnly = append(actualFilesOnly, f)
		} else if err != nil {
			// If stat fails, it might be a file that was expected, include it for comparison
			// This handles cases where listFiles might include paths that don't exist after potential removals (though not expected here)
			// A more robust approach might involve checking specific error types if needed.
			// For now, assume non-stat-able entries listed by listFiles are files for comparison purposes.
			// Check if the path looks like a file (has an extension or no slash at the end)
			if !strings.HasSuffix(f, string(filepath.Separator)) && strings.Contains(filepath.Base(f), ".") {
				actualFilesOnly = append(actualFilesOnly, f)
			}
		}
	}
	sort.Strings(actualFilesOnly)

	if !reflect.DeepEqual(actualFilesOnly, expectedFilesOnly) {
		t.Errorf(`Destination directory has incorrect files after CopyFiles(cleanDest=false):
Got Files: %v
Want Files: %v
All listed entries: %v`, actualFilesOnly, expectedFilesOnly, dstFiles)
	}

	// Log the actual files for debugging
	// t.Logf("Files in destination directory: %v", dstFiles) // Keep commented unless debugging

	// Verify content of copied files
	for _, file := range filesToCopy {
		dstPath := filepath.Join(dstDir, file)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", dstPath, err)
			continue
		}

		if string(content) != "source content" {
			t.Errorf("File %s has incorrect content: got %q, want %q", file, string(content), "source content")
		}
	}

	// Verify content of preserved files
	for _, file := range preserveFiles {
		dstPath := filepath.Join(dstDir, file)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read preserved file %s: %v", dstPath, err)
			continue
		}

		if string(content) != "preserved content" {
			t.Errorf("Preserved file %s has incorrect content: got %q, want %q", file, string(content), "preserved content")
		}
	}
}

// TestClearDestinationDirWithExclusions tests that the clearDestinationDir function correctly
// preserves files matching the exclusion patterns
func TestClearDestinationDirWithExclusions(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create files in the temporary directory
	files := []string{
		"regular_file.txt",
		".hidden_file",
		"dir1/nested_file.txt",
		"dir1/.hidden_nested_file",
		".hidden_dir/file.txt",
		"keep_this_file.txt",
		"dir2/keep_this_too.txt",
		"config/important.json",
	}

	// Create directories and files
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Define exclusion patterns
	excludePatterns := []string{
		"keep_this_file.txt",
		"dir2/*",
		"config/*.json",
	}

	// Call clearDestinationDir with exclusion patterns
	err := clearDestinationDir(tempDir, excludePatterns)
	if err != nil {
		t.Fatalf("clearDestinationDir failed: %v", err)
	}

	// List remaining files
	remainingFiles, err := listFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to list remaining files: %v", err)
	}

	// Expected remaining files (hidden files/dirs, files matching patterns, and dirs containing them)
	expectedFiles := []string{
		".hidden_dir",
		".hidden_dir/file.txt",
		".hidden_file",
		"config",                   // Preserved by pattern "config/*.json"
		"config/important.json",    // Preserved by pattern "config/*.json"
		"dir1",                     // Contains preserved hidden file
		"dir1/.hidden_nested_file", // Preserved hidden file
		"dir2",                     // Preserved by pattern "dir2/*"
		"dir2/keep_this_too.txt",   // Preserved by pattern "dir2/*"
		"keep_this_file.txt",       // Preserved by pattern "keep_this_file.txt"
	}
	sort.Strings(expectedFiles)

	// Verify only hidden files and files matching exclusion patterns remain
	if !reflect.DeepEqual(remainingFiles, expectedFiles) {
		t.Errorf("clearDestinationDir did not preserve files correctly with exclusions: Got:  %v Want: %v", remainingFiles, expectedFiles)
	}
}

// TestCopyFilesWithCleanExclusions tests that the CopyFiles function correctly
// preserves files matching the clean-exclude patterns when cleaning the destination directory
func TestCopyFilesWithCleanExclusions(t *testing.T) {
	// Setup test directories
	srcDir, dstDir := setupTestDir(t)

	// Files to copy (all non-hidden files from source)
	filesToCopy := []string{
		"file1.txt",
		"file2.go",
		"dir1/file3.txt",
		"dir2/file4.go",
	}

	// Create additional files in the destination that should be preserved
	// based on clean-exclude patterns
	preserveFiles := []string{
		"keep_this_file.txt",
		"dir3/keep_this_too.txt",
		"config/important.json",
	}

	for _, file := range preserveFiles {
		filePath := filepath.Join(dstDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("preserved content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Define clean-exclude patterns
	cleanExcludePatterns := []string{
		"keep_this_file.txt",
		"dir3/*",
		"config/*.json",
	}

	// Perform the copy operation with cleanDest=true and clean-exclude patterns
	err := CopyFiles(srcDir, dstDir, filesToCopy, true, cleanExcludePatterns)
	if err != nil {
		t.Fatalf("CopyFiles failed: %v", err)
	}

	// List all files in the destination directory after copy
	dstFiles, err := listFiles(dstDir)
	if err != nil {
		t.Fatalf("Failed to list destination files: %v", err)
	}

	// Expected files in destination after copy:
	// 1. All copied non-hidden files from source
	// 2. All hidden files that were already in destination (including nested)
	// 3. All files/dirs matching the clean-exclude patterns
	expectedFiles := []string{
		".hidden_preserve1",      // Preserved top-level hidden file
		"config",                 // Preserved by exclude pattern "config/*.json"
		"config/important.json",  // Preserved by exclude pattern "config/*.json"
		"dir1",                   // Directory containing preserved hidden file and copied file
		"dir1/.hidden_preserve2", // Preserved nested hidden file
		"dir1/file3.txt",         // Copied
		"dir2",                   // Directory containing copied file
		"dir2/file4.go",          // Copied
		"file1.txt",              // Copied
		"file2.go",               // Copied
		"dir3",                   // Preserved by exclude pattern "dir3/*"
		"dir3/keep_this_too.txt", // Preserved by exclude pattern "dir3/*"
		"keep_this_file.txt",     // Preserved by exclude pattern
	}
	sort.Strings(expectedFiles)

	// Verify the destination directory contains the expected files
	if !reflect.DeepEqual(dstFiles, expectedFiles) {
		t.Errorf("Destination directory has incorrect files after CopyFiles(cleanDest=true, with exclusions): Got:  %v Want: %v", dstFiles, expectedFiles)
	}

	// Verify content of copied files
	for _, file := range filesToCopy {
		dstPath := filepath.Join(dstDir, file)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read copied file %s: %v", dstPath, err)
			continue
		}

		if string(content) != "source content" {
			t.Errorf("File %s has incorrect content: got %q, want %q", file, string(content), "source content")
		}
	}

	// Verify content of preserved files
	// Restore check for nested preserved files
	preserveFilesToCheck := preserveFiles       // Use the original preserveFiles list
	for _, file := range preserveFilesToCheck { // Changed back from preserveFilesToCheck
		dstPath := filepath.Join(dstDir, file)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read preserved file %s: %v", dstPath, err)
			continue
		}

		if string(content) != "preserved content" {
			t.Errorf("Preserved file %s has incorrect content: got %q, want %q", file, string(content), "preserved content")
		}
	}
}

// TestDefaultCleanExcludePattern tests that .gitkeep files are preserved by default
func TestDefaultCleanExcludePattern(t *testing.T) {
	// Setup test directories
	srcDir, dstDir := setupTestDir(t)

	// Create .gitkeep files in destination directory
	gitkeepFiles := []string{
		".gitkeep",
		"empty_dir/.gitkeep",
		"another_dir/.gitkeep",
	}

	for _, file := range gitkeepFiles {
		filePath := filepath.Join(dstDir, file)
		dirPath := filepath.Dir(filePath)

		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}

		if err := os.WriteFile(filePath, []byte("gitkeep content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	// Files to copy
	filesToCopy := []string{
		"file1.txt",
		"file2.go",
	}

	// Perform the copy operation with cleanDest=true and default clean-exclude pattern (.gitkeep)
	defaultCleanExclude := []string{".gitkeep"}
	err := CopyFiles(srcDir, dstDir, filesToCopy, true, defaultCleanExclude)
	if err != nil {
		t.Fatalf("CopyFiles failed: %v", err)
	}

	// List all files in the destination directory after copy
	dstFiles, err := listFiles(dstDir)
	if err != nil {
		t.Fatalf("Failed to list destination files: %v", err)
	}

	// Check that .gitkeep files are preserved
	expectedGitkeepFiles := []string{
		".gitkeep",
		"another_dir/.gitkeep",
		"empty_dir/.gitkeep",
	}

	for _, gitkeepFile := range expectedGitkeepFiles {
		found := false
		for _, actualFile := range dstFiles {
			if actualFile == gitkeepFile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(".gitkeep file %s was not preserved", gitkeepFile)
		}
	}

	// Check that copied files exist
	expectedCopiedFiles := []string{
		"file1.txt",
		"file2.go",
	}

	for _, copiedFile := range expectedCopiedFiles {
		found := false
		for _, actualFile := range dstFiles {
			if actualFile == copiedFile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Copied file %s was not found", copiedFile)
		}
	}

	// Verify content of .gitkeep files
	for _, gitkeepFile := range expectedGitkeepFiles {
		filePath := filepath.Join(dstDir, gitkeepFile)
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read .gitkeep file %s: %v", filePath, err)
			continue
		}

		if string(content) != "gitkeep content" {
			t.Errorf(".gitkeep file %s has incorrect content: got %q, want %q", gitkeepFile, string(content), "gitkeep content")
		}
	}
}
