package cli

import (
	"fmt"

	"github.com/alecthomas/kong"
)

var (
	version   = "dev"     // Default value
	commit    = "none"    // Default value
	buildDate = "unknown" // Default value
)

// CLI represents the command-line interface structure
type CLI struct {
	From    string   `name:"from" help:"Source directory to copy files from." type:"path"`
	To      string   `name:"to" help:"Destination directory to copy files to." type:"path"`
	Include []string `name:"include" short:"i" help:"Patterns to include (glob syntax, e.g. '*.go')."`
	Exclude []string `name:"exclude" short:"e" help:"Patterns to exclude (glob syntax, e.g. '*.tmp')."`

	Version kong.VersionFlag `short:"v" help:"Show version and exit."`
}

// Validate validates the CLI arguments
func (c *CLI) Validate() error {
	// If version flag is set, no validation needed
	if c.Version {
		return nil
	}

	// Validate required fields when not showing version
	if c.From == "" {
		return fmt.Errorf("--from flag is required")
	}
	if c.To == "" {
		return fmt.Errorf("--to flag is required")
	}

	return nil
}

// GetVersion returns the formatted version string
func GetVersion() string {
	return fmt.Sprintf("%s (commit: %s, built at: %s)", version, commit, buildDate)
}

// SetVersionInfo sets the version information
func SetVersionInfo(ver, cmt, date string) {
	if ver != "" {
		version = ver
	}
	if cmt != "" {
		commit = cmt
	}
	if date != "" {
		buildDate = date
	}
}
