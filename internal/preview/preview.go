package preview

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MaxPreviewSize is the maximum size of a file to preview (100KB)
const MaxPreviewSize = 100 * 1024

// GeneratePreview generates a preview of the file at the given path
// This function is designed to work with go-fuzzyfinder's preview window
func GeneratePreview(baseDir, relPath string, width, height int) (string, error) {
	fullPath := filepath.Join(baseDir, relPath)

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Handle directory
	if info.IsDir() {
		return generateDirectoryPreview(fullPath, width, height)
	}

	// Check file size
	if info.Size() > MaxPreviewSize {
		return fmt.Sprintf("File too large to preview (%.2f MB)", float64(info.Size())/1024/1024), nil
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Check if it's a binary file
	if isBinaryContent(content) || isBinaryFilename(fullPath) {
		return fmt.Sprintf("Binary file (%s, %.2f KB)", filepath.Base(fullPath), float64(info.Size())/1024), nil
	}

	// Format the content for display
	return formatContentForDisplay(string(content), width, height), nil
}

// generateDirectoryPreview generates a preview of the directory contents
func generateDirectoryPreview(dirPath string, width, height int) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Directory: %s\n\n", dirPath))
	buf.WriteString("Contents:\n")

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Format: [D] dirname/ or [F] filename (size)
		if entry.IsDir() {
			buf.WriteString(fmt.Sprintf("[D] %s/\n", entry.Name()))
		} else {
			buf.WriteString(fmt.Sprintf("[F] %s (%.2f KB)\n", entry.Name(), float64(info.Size())/1024))
		}
	}

	return buf.String(), nil
}

// formatContentForDisplay formats the content for display in the preview window
func formatContentForDisplay(content string, width, height int) string {
	// Split content into lines
	lines := strings.Split(content, "\n")

	// Limit the number of lines to display based on height
	if len(lines) > height-2 { // Leave some space for borders
		lines = lines[:height-2]
		lines = append(lines, "... (truncated)")
	}

	// Truncate long lines based on width
	for i, line := range lines {
		if len(line) > width-4 { // Leave some space for borders
			lines[i] = line[:width-7] + "..."
		}
	}

	return strings.Join(lines, "\n")
}

// isBinaryFilename checks if the filename suggests a binary file
func isBinaryFilename(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	binaryExts := []string{
		".exe", ".dll", ".so", ".dylib", ".bin", ".obj",
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".ico",
		".zip", ".tar", ".gz", ".rar", ".7z",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	}

	for _, binaryExt := range binaryExts {
		if ext == binaryExt {
			return true
		}
	}

	return false
}

// isBinaryContent checks if the content appears to be binary
func isBinaryContent(content []byte) bool {
	// Check for null bytes or too many non-printable characters
	nullCount := bytes.Count(content, []byte{0})
	if nullCount > 0 {
		return true
	}

	// Check a sample of the content for non-printable characters
	sampleSize := 1024
	if len(content) < sampleSize {
		sampleSize = len(content)
	}

	nonPrintableCount := 0
	for i := 0; i < sampleSize; i++ {
		c := content[i]
		if (c < 32 || c > 126) && !isAllowedNonPrintable(c) {
			nonPrintableCount++
		}
	}

	// If more than 10% are non-printable, consider it binary
	return nonPrintableCount > sampleSize/10
}

// isAllowedNonPrintable checks if a non-printable character is allowed in text files
func isAllowedNonPrintable(c byte) bool {
	// Allow common whitespace characters: tab, newline, carriage return
	return c == 9 || c == 10 || c == 13
}
