package finder

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindFiles searches for files in the given root directory
// and filters them based on include and exclude patterns
func FindFiles(rootDir string, includes, excludes []string) ([]string, error) {
	// Check if the root directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, err
	}

	var files []string

	// Walk through the directory recursively
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == rootDir {
			return nil
		}

		// Get the relative path from the root directory
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		// Check if the file should be included based on patterns
		if shouldInclude(relPath, includes, excludes) {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If no files were found, return an empty slice
	if len(files) == 0 {
		return []string{}, nil
	}

	return files, nil
}

// shouldInclude determines if a file should be included based on
// include and exclude patterns
func shouldInclude(path string, includes, excludes []string) bool {
	// Check exclude patterns first (they take precedence)
	for _, pattern := range excludes {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return false
		}

		// Handle directory patterns like "dir/*"
		if strings.HasSuffix(pattern, "/*") {
			dirPattern := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, dirPattern+"/") {
				return false
			}
		}
	}

	// If no include patterns are specified, include everything
	if len(includes) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range includes {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Handle directory patterns like "dir/*"
		if strings.HasSuffix(pattern, "/*") {
			dirPattern := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, dirPattern+"/") {
				return true
			}
		}
	}

	return false
}
