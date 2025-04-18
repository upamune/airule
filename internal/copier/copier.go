package copier

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// clearDestinationDir selectively removes files and subdirectories in the destination directory
// while preserving files and directories that start with a dot (hidden files/directories).
// It ensures the directory exists after clearing.
func clearDestinationDir(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err == nil {
		// Directory exists, selectively remove contents
		entries, err := os.ReadDir(dir)
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
			path := filepath.Join(dir, name)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	} else if os.IsNotExist(err) {
		// Directory doesn't exist, create it
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
		return nil
	} else {
		// Some other error occurred
		return fmt.Errorf("failed to check destination directory: %w", err)
	}

	// Ensure directory permissions are set correctly if it was modified
	if info != nil && info.Mode().Perm() != 0755 {
		if err := os.Chmod(dir, 0755); err != nil {
			return fmt.Errorf("failed to set directory permissions: %w", err)
		}
	}

	return nil
}

// CopyFiles copies files from the source directory to the destination directory
// If cleanDest is true, it will clear the destination directory before copying,
// while preserving hidden files (those starting with a dot).
// If cleanDest is false, it will not clear the destination directory.
func CopyFiles(fromDir, toDir string, relativePaths []string, cleanDest bool) error {
	// Clear the destination directory before copying if cleanDest is true
	if cleanDest {
		if err := clearDestinationDir(toDir); err != nil {
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
