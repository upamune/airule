package finder

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FindFiles searches for files in the given root directory
// and filters them based on include and exclude patterns
func FindFiles(rootDir string, includes, excludes []string) ([]string, error) {
	// Check if the root directory exists
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, err
	}

	foundFiles := make([]string, 0)
	// Keep track of parent directories of found files (using a map as a set)
	parentDirs := make(map[string]struct{})

	// Walk through the directory recursively
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Handle errors accessing files/dirs, but continue walking if possible
			if errors.Is(err, fs.ErrPermission) {
				// Can't read directory or file, skip it
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil // Skip the file
			}
			// For other errors, report them but try to continue
			// Consider logging the error here: log.Printf("Error accessing %s: %v", path, err)
			return nil // or return err to stop the walk
		}

		// Skip the root directory itself
		if path == rootDir {
			return nil
		}

		// Get the relative path from the root directory
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err // Stop walk on relative path error
		}

		// Check if the current directory should be skipped based on exclude patterns
		if d.IsDir() {
			isDirExcluded := false
			for _, pattern := range excludes {
				// Use filepath.Match for directory exclusion
				matched, _ := filepath.Match(pattern, relPath)
				if matched {
					isDirExcluded = true
					break
				}
				// Handle dir/* exclude patterns
				if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
					dirPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "/")
					if dirPattern != "" && strings.HasPrefix(relPath, dirPattern+string(filepath.Separator)) {
						isDirExcluded = true
						break
					}
					if relPath == dirPattern { // Match the directory itself
						isDirExcluded = true
						break
					}
				}
			}
			if isDirExcluded {
				return fs.SkipDir // Skip excluded directory
			}
			// Don't record directories during walk, only parents of found files later
			return nil // Continue walking into the directory
		}

		// Check if the file should be included
		if shouldInclude(relPath, includes, excludes) {
			foundFiles = append(foundFiles, relPath)

			// Add all parent directories to the set
			dir := filepath.Dir(relPath)
			for dir != "." && dir != "/" {
				parentDirs[dir] = struct{}{}
				dir = filepath.Dir(dir)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Combine found files and parent directories (if not excluded)
	finalResultsMap := make(map[string]struct{}) // Use map to auto-handle duplicates
	for _, file := range foundFiles {
		finalResultsMap[file] = struct{}{}
	}
	for dir := range parentDirs {
		// Add directory only if it's NOT excluded itself
		// We check exclusion by calling shouldInclude with only exclude patterns
		if shouldInclude(dir, []string{"*"}, excludes) { // Check if it passes *any* include (placeholder *) against excludes
			finalResultsMap[dir] = struct{}{}
		}
	}

	// Convert map keys to slice
	finalResults := make([]string, 0, len(finalResultsMap))
	for path := range finalResultsMap {
		finalResults = append(finalResults, path)
	}

	// Sort the final results in reverse alphabetical order
	sort.Sort(sort.Reverse(sort.StringSlice(finalResults)))

	return finalResults, nil
}

// shouldInclude determines if a file should be included based on
// include and exclude patterns. It performs a basic check against the given path.
func shouldInclude(path string, includes, excludes []string) bool {
	// Check exclude patterns first (they take precedence)
	for _, pattern := range excludes {
		// Match against the full path or just the basename if the pattern doesn't contain a separator
		base := filepath.Base(path)
		matchPath, _ := filepath.Match(pattern, path)
		matchBase := false
		if !strings.Contains(pattern, string(filepath.Separator)) {
			matchBase, _ = filepath.Match(pattern, base)
		}
		if matchPath || matchBase {
			return false
		}

		// Handle directory exclude patterns specifically (e.g., "dir/*" or "dir/**")
		// Check if the path is within an excluded directory pattern
		if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
			dirPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "/")
			// Ensure dirPattern is not empty and path actually starts with it + separator
			if dirPattern != "" && strings.HasPrefix(path, dirPattern+string(filepath.Separator)) {
				return false
			}
			// Also handle case where the excluded pattern *is* the directory path itself
			if path == dirPattern {
				return false
			}
		}
	}

	// If no include patterns are specified, include everything not excluded
	if len(includes) == 0 {
		return true
	}

	// Check include patterns
	isIncluded := false
	for _, pattern := range includes {
		// Match against the full path or just the basename if the pattern doesn't contain a separator
		base := filepath.Base(path)
		matchPath, _ := filepath.Match(pattern, path)
		matchBase := false
		if !strings.Contains(pattern, string(filepath.Separator)) {
			matchBase, _ = filepath.Match(pattern, base)
		}
		if matchPath || matchBase {
			isIncluded = true
			break // Found a matching include pattern
		}

		// Handle directory include patterns specifically (e.g., "dir/*" or "dir/**")
		// Check if the path is within an included directory pattern
		if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
			dirPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "/")
			// Ensure dirPattern is not empty and path actually starts with it + separator
			if dirPattern != "" && strings.HasPrefix(path, dirPattern+string(filepath.Separator)) {
				isIncluded = true
				break
			}
			// Also handle case where the included pattern *is* the directory path itself
			if path == dirPattern {
				isIncluded = true
				break
			}
		}
	}

	return isIncluded
}
