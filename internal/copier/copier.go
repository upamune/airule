package copier

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// clearDestinationDir removes all files and subdirectories in the destination directory
// and ensures the directory exists after clearing
func clearDestinationDir(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); err == nil {
		// Directory exists, remove all contents
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to clear destination directory: %w", err)
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("failed to check destination directory: %w", err)
	}

	// Recreate the directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to recreate destination directory: %w", err)
	}

	return nil
}

// CopyFiles copies files from the source directory to the destination directory
func CopyFiles(fromDir, toDir string, relativePaths []string) error {
	// Clear the destination directory before copying
	if err := clearDestinationDir(toDir); err != nil {
		return err
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
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dst, err)
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
