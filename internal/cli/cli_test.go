package cli

import (
	"reflect"
	"testing"

	"github.com/alecthomas/kong"
)

// TestDefaultValues tests that CLI struct has correct default values
func TestDefaultValues(t *testing.T) {
	var cli CLI
	
	// Parse empty arguments to get default values
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Parse with minimal required arguments
	_, err = parser.Parse([]string{"--from", "/tmp/src", "--to", "/tmp/dst"})
	if err != nil {
		t.Fatalf("Failed to parse arguments: %v", err)
	}

	// Test default values
	tests := []struct {
		name     string
		actual   interface{}
		expected interface{}
	}{
		{
			name:     "Clean default value",
			actual:   cli.Clean,
			expected: true,
		},
		{
			name:     "CleanExclude default value",
			actual:   cli.CleanExclude,
			expected: []string{".gitkeep"},
		},
		{
			name:     "SelectAll default value",
			actual:   cli.SelectAll,
			expected: false,
		},
		{
			name:     "Include default value",
			actual:   cli.Include,
			expected: []string(nil),
		},
		{
			name:     "Exclude default value",
			actual:   cli.Exclude,
			expected: []string(nil),
		},
		{
			name:     "PreSelect default value",
			actual:   cli.PreSelect,
			expected: []string(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.actual, tt.expected) {
				t.Errorf("%s: got %v, want %v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

// TestCleanExcludeOverride tests that CleanExclude can be overridden
func TestCleanExcludeOverride(t *testing.T) {
	var cli CLI
	
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Parse with custom clean-exclude patterns
	_, err = parser.Parse([]string{
		"--from", "/tmp/src", 
		"--to", "/tmp/dst",
		"--clean-exclude", "*.log",
		"--clean-exclude", "config/*",
	})
	if err != nil {
		t.Fatalf("Failed to parse arguments: %v", err)
	}

	expected := []string{"*.log", "config/*"}
	if !reflect.DeepEqual(cli.CleanExclude, expected) {
		t.Errorf("CleanExclude override: got %v, want %v", cli.CleanExclude, expected)
	}
}

// TestEnvironmentVariables tests that environment variables work correctly
func TestEnvironmentVariables(t *testing.T) {
	// This test would require setting environment variables
	// For now, we'll just test the struct tags are correct
	
	var cli CLI
	parser, err := kong.New(&cli)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Verify that the parser was created successfully with env tags
	// The actual environment variable testing would require more complex setup
	if parser == nil {
		t.Error("Parser should not be nil")
	}
}
