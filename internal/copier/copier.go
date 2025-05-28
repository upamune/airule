package copier

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

		// Handle glob patterns with filepath.Match for more complex patterns
		if strings.Contains(pattern, "*") {
			matched, _ := filepath.Match(pattern, filePath)
			if matched {
				return true
			}
		}
	}
	return false
}

// shouldPreserve checks if a path should be preserved based on exclusion patterns
// It returns true if:
// 1. The path is hidden (starts with a dot)
// 2. The path matches any of the exclusion patterns
// 3. The path is a directory that contains files matching any of the exclusion patterns
func shouldPreserve(path string, isDir bool, excludePatterns []string) bool {
	// Check if it's a hidden file/directory
	name := filepath.Base(path)
	if len(name) > 0 && name[0] == '.' {
		return true
	}

	// Check if the path matches any of the exclusion patterns
	if matchesAnyPattern(path, excludePatterns) {
		return true
	}

	// For directories, check if any exclusion pattern would match files inside this directory
	if isDir {
		for _, pattern := range excludePatterns {
			// Check if this is a directory pattern (e.g., "dir/*" or "dir/**")
			if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
				dirPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "/")
				// If the pattern directory is a subdirectory of the current directory, preserve it
				if dirPattern != "" && strings.HasPrefix(dirPattern, path+string(filepath.Separator)) {
					return true
				}
				// If the current directory is the pattern directory itself, preserve it
				if path == dirPattern {
					return true
				}
			}
		}
	}

	// Special case: Check if any pattern directly targets a file in this directory
	// This handles patterns like "config/*.json" which should preserve the "config" directory
	if isDir {
		dirPrefix := path + string(filepath.Separator)
		for _, pattern := range excludePatterns {
			// Skip directory wildcard patterns as they're handled above
			if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
				continue
			}

			// Check if the pattern targets a file in this directory
			if strings.Contains(pattern, string(filepath.Separator)) {
				patternDir := filepath.Dir(pattern)
				if patternDir == path || strings.HasPrefix(patternDir, dirPrefix) {
					return true
				}
			}
		}
	}

	return false
}

// checkPreservationRecursive checks if a path or any item within it (if it's a directory)
// should be preserved based on hidden status or exclusion patterns.
// It returns true if the path itself should be preserved OR if it's a directory
// containing at least one item that should be preserved recursively OR if any parent
// directory should be preserved.
func checkPreservationRecursive(path string, excludePatterns []string) (bool, error) {
	return checkPreservationRecursiveWithBase(path, "", excludePatterns)
}

// checkPreservationRecursiveWithBase is the internal implementation that tracks the base directory
func checkPreservationRecursiveWithBase(path, baseDir string, excludePatterns []string) (bool, error) {
	info, err := os.Lstat(path) // Use Lstat to handle symlinks if they were ever supported
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Path doesn't exist, definitely not preserved
		}
		return false, err // Other stat error
	}

	// Calculate relative path for pattern matching
	var relPath string
	if baseDir != "" {
		var err error
		relPath, err = filepath.Rel(baseDir, path)
		if err != nil {
			relPath = path // Fallback to absolute path
		}
	} else {
		relPath = path
	}

	// 1. Check if the item itself is hidden or matches exclude patterns
	name := filepath.Base(path)
	isDir := info.IsDir()
	isHidden := len(name) > 0 && name[0] == '.'
	matchesExclusion := matchesAnyPattern(relPath, excludePatterns)

	if isHidden || matchesExclusion {
		return true, nil // Item itself should be preserved
	}

	// 2. Check if any parent directory is hidden or matches exclude patterns
	currentPath := path
	for {
		parent := filepath.Dir(currentPath)
		if parent == currentPath || parent == "." || parent == "/" {
			break
		}

		// Calculate relative path for parent
		var parentRelPath string
		if baseDir != "" {
			var err error
			parentRelPath, err = filepath.Rel(baseDir, parent)
			if err != nil {
				parentRelPath = parent // Fallback to absolute path
			}
		} else {
			parentRelPath = parent
		}

		parentName := filepath.Base(parent)
		parentIsHidden := len(parentName) > 0 && parentName[0] == '.'
		parentMatchesExclusion := matchesAnyPattern(parentRelPath, excludePatterns)

		if parentIsHidden || parentMatchesExclusion {
			return true, nil // Parent directory should be preserved, so this item should too
		}

		currentPath = parent
	}

	// 3. If it's a directory, check if any exclusion pattern would match files inside this directory
	if isDir {
		for _, pattern := range excludePatterns {
			// Check if this is a directory pattern (e.g., "dir/*" or "dir/**")
			if strings.HasSuffix(pattern, "/*") || strings.HasSuffix(pattern, "/**") {
				dirPattern := strings.TrimSuffix(strings.TrimSuffix(pattern, "*"), "/")
				if relPath == dirPattern {
					return true, nil // This directory is targeted by a wildcard pattern
				}
			}

			// Check if any pattern targets files in this directory (e.g., "config/*.json")
			if strings.Contains(pattern, string(filepath.Separator)) && strings.Contains(pattern, "*") {
				patternDir := filepath.Dir(pattern)
				if relPath == patternDir {
					return true, nil // This directory contains files that match the pattern
				}
			}
		}

		// Check its contents recursively
		entries, err := os.ReadDir(path)
		if err != nil {
			// Handle cases like permission denied reading directory
			// If we can't read it, we can't know if it needs preservation, err on the side of caution?
			// Or assume it doesn't need preservation if unreadable? Let's return error for now.
			return false, fmt.Errorf("failed to read directory %s for preservation check: %w", path, err)
		}
		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			// Recursively check child. If any child needs preservation, this dir needs it too.
			preserveChild, err := checkPreservationRecursiveWithBase(childPath, baseDir, excludePatterns)
			if err != nil {
				return false, err // Propagate error from recursive call
			}
			if preserveChild {
				return true, nil // Found a child that needs preservation, so this directory must be kept
			}
		}
	}

	// If we reach here, neither the item itself nor any of its children (if dir) need preservation.
	return false, nil
}

// clearDestinationDir selectively removes files and subdirectories in the destination directory
// while preserving files/directories that are hidden or match exclude patterns,
// including items nested within directories and the parent directories needed to hold them.
func clearDestinationDir(dir string, excludePatterns []string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, create it
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
			return nil
		}
		// Some other error occurred
		return fmt.Errorf("failed to check destination directory: %w", err)
	}

	// Use filepath.WalkDir to traverse the directory.
	// We need to remove items *after* traversing their children if the parent directory
	// itself doesn't need preservation but some children do. This is tricky with WalkDir.
	// A simpler approach might be to collect all paths to potentially remove first,
	// then iterate through them and check preservation *again* before removing.

	pathsToRemove := []string{}
	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Propagate walk errors
		}
		// Skip root
		if path == dir {
			return nil
		}

		// Tentatively add all paths for removal check later
		pathsToRemove = append(pathsToRemove, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking destination directory %s: %w", dir, err)
	}

	// Sort paths in reverse order so children are processed before parents
	sort.Slice(pathsToRemove, func(i, j int) bool {
		return len(pathsToRemove[i]) > len(pathsToRemove[j])
	})

	// Now, check each path for preservation and remove if necessary
	for _, path := range pathsToRemove {
		// Check if the path still exists (might have been removed as part of a parent dir)
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue // Already removed
		}

		preserve, err := checkPreservationRecursiveWithBase(path, dir, excludePatterns)
		if err != nil {
			// Log or handle error during check, maybe skip removal?
			fmt.Fprintf(os.Stderr, "Warning: error checking preservation for %s, skipping removal: %v\n", path, err)
			continue
		}

		if !preserve {
			// Attempt to remove. Use RemoveAll for directories.
			if err := os.RemoveAll(path); err != nil {
				// Log or handle error during removal
				fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", path, err)
				// Decide whether to continue or return error. Let's continue for now.
			}
		}
	}

	// Ensure the root directory still exists and has correct permissions
	// (It shouldn't have been added to pathsToRemove, but double-check)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// This is unexpected if removal logic is correct, recreate it.
			fmt.Fprintf(os.Stderr, "Warning: destination directory %s was unexpectedly removed, recreating.\n", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to recreate destination directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to stat destination directory after clear: %w", err)
		}
	} else if info.Mode().Perm() != 0755 {
		if err := os.Chmod(dir, 0755); err != nil {
			return fmt.Errorf("failed to set directory permissions after clear: %w", err)
		}
	}

	return nil
}

// CopyFiles copies files from the source directory to the destination directory
// If cleanDest is true, it will clear the destination directory before copying,
// while preserving hidden files (those starting with a dot) and files matching cleanExcludePatterns.
// If cleanDest is false, it will not clear the destination directory.
func CopyFiles(fromDir, toDir string, relativePaths []string, cleanDest bool, cleanExcludePatterns []string) error {
	// Clear the destination directory before copying if cleanDest is true
	if cleanDest {
		if err := clearDestinationDir(toDir, cleanExcludePatterns); err != nil {
			return err
		}
	} else {
		// Ensure the destination directory exists
		if err := os.MkdirAll(toDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	// Copy each file
	for _, relPath := range relativePaths {
		srcPath := filepath.Join(fromDir, relPath)
		dstPath := filepath.Join(toDir, relPath)

		// Get file info
		info, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", srcPath, err)
		}

		// Handle directories and files differently
		if info.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", relPath, err)
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", relPath, err)
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dstDir, err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// copyDir copies a directory recursively from src to dst
// while preserving hidden files in the destination directory
func copyDir(src, dst string) error {
	// Check if destination directory exists
	_, err := os.Stat(dst)
	if err == nil {
		// Directory exists, selectively remove non-hidden contents
		entries, err := os.ReadDir(dst)
		if err != nil {
			return fmt.Errorf("failed to read destination directory: %w", err)
		}

		// Remove each non-hidden entry
		for _, entry := range entries {
			name := entry.Name()
			// Skip files/directories that start with a dot (hidden)
			if len(name) > 0 && name[0] == '.' {
				continue
			}

			// Remove non-hidden file/directory
			path := filepath.Join(dst, name)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	} else if os.IsNotExist(err) {
		// Directory doesn't exist, create it
		if err := os.MkdirAll(dst, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dst, err)
		}
	} else {
		// Some other error occurred
		return fmt.Errorf("failed to check destination directory: %w", err)
	}

	// Get source directory info for permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	// Set the same permissions on the destination directory
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set directory permissions: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
