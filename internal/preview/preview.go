package preview

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MaxPreviewSize is the maximum size of a file to preview (100KB)
const MaxPreviewSize = 100 * 1024

// PreviewLoadedMsg is the message sent when a preview is loaded
type PreviewLoadedMsg struct {
	Content string
	Err     error
}

// LoadPreview loads a preview of the file at the given path
func LoadPreview(baseDir, relPath string) tea.Cmd {
	return func() tea.Msg {
		fullPath := filepath.Join(baseDir, relPath)

		// Get file info
		info, err := os.Stat(fullPath)
		if err != nil {
			return PreviewLoadedMsg{
				Err: fmt.Errorf("failed to get file info: %w", err),
			}
		}

		// Handle directory
		if info.IsDir() {
			return loadDirectoryPreview(fullPath)
		}

		// Check file size
		if info.Size() > MaxPreviewSize {
			return PreviewLoadedMsg{
				Content: fmt.Sprintf("File too large to preview (%.2f MB)", float64(info.Size())/1024/1024),
			}
		}

		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return PreviewLoadedMsg{
				Err: fmt.Errorf("failed to read file: %w", err),
			}
		}

		// Check if it's a binary file
		if isBinaryContent(content) || isBinaryFilename(fullPath) {
			return PreviewLoadedMsg{
				Content: fmt.Sprintf("Binary file (%s, %.2f KB)", filepath.Base(fullPath), float64(info.Size())/1024),
			}
		}

		// Return the content
		return PreviewLoadedMsg{
			Content: string(content),
		}
	}
}

// loadDirectoryPreview loads a preview of the directory contents
func loadDirectoryPreview(dirPath string) PreviewLoadedMsg {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return PreviewLoadedMsg{
			Err: fmt.Errorf("failed to read directory: %w", err),
		}
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

	return PreviewLoadedMsg{
		Content: buf.String(),
	}
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
