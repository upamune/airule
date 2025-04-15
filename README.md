# airule

An interactive terminal-based tool for selectively copying rule files between directories.

![airule screenshot placeholder](https://via.placeholder.com/800x400?text=airule+Screenshot)

## Description

airule is a command-line utility that provides an interactive terminal user interface (TUI) for selecting and copying files between directories. It allows you to:

- Browse files from a source directory
- Preview file contents before copying
- Select multiple files for copying
- Copy selected files to a destination directory while preserving directory structure

Perfect for selectively copying configuration files, rules, or any other files that need to be transferred between directories with visual confirmation.

## Installation

### Prerequisites

- Go 1.23.1 or higher

### From Homebrew

```bash
brew install upamune/tap/airule
```

### From Source

```bash
# Clone the repository
git clone https://github.com/upamune/airule.git
cd airule

# Build the binary
go build -o airule ./cmd/airule

# Move to a directory in your PATH (optional)
sudo mv airule /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/upamune/airule/cmd/airule@latest
```

## Usage

Basic usage:

```bash
airule --from /path/to/source --to /path/to/destination
```

With file filtering:

```bash
airule --from /path/to/source --to /path/to/destination --include "*.json" --exclude "*.tmp"
```

### Command-line Arguments

| Argument | Short | Description | Required |
|----------|-------|-------------|----------|
| `--from` | | Source directory to copy files from. Can also be set via the `AIRULE_FROM` environment variable. | Yes |
| `--to` | | Destination directory to copy files to. Can also be set via the `AIRULE_TO` environment variable. | Yes |
| `--include` | `-i` | Patterns to include (glob syntax, e.g., '*.go') Can also be set via the `AIRULE_INCLUDE` environment variable. | No |
| `--exclude` | `-e` | Patterns to exclude (glob syntax, e.g., '*.tmp') Can also be set via the `AIRULE_EXCLUDE` environment variable. | No |
| `--select-all` | | Select all files by default. Can also be set via the `AIRULE_SELECT_ALL` environment variable. | No |
| `--pre-select` | | Patterns to pre-select (glob syntax, e.g., '*.go'). Can be specified multiple times. Can also be set via the `AIRULE_PRE_SELECT` environment variable. | No |
| `--version` | `-v` | Show version information and exit | No |

### Examples

Copy all JSON files from config directory to backup directory:

```bash
airule --from ./config --to ./backup --include "*.json"
```

Copy all files except temporary files:

```bash
airule --from ./src --to ./dest --exclude "*.tmp" --exclude "*.bak"
```

Copy all files and preselect all files by default:

```bash
airule --from ./config --to ./backup --select-all
```

Copy files with specific patterns preselected:

```bash
airule --from ./src --to ./dest --pre-select "*.json" --pre-select "config/*.yaml"
```

## Key Features

- **Interactive File Selection**: Browse and select files using a terminal user interface
- **File Preview**: View file contents before copying
- **Pattern Filtering**: Include or exclude files based on glob patterns
- **File Preselection**: Automatically select all files or files matching specific patterns
- **Directory Structure Preservation**: Maintains the original directory structure when copying
- **Binary File Detection**: Automatically detects and handles binary files
- **Large File Handling**: Provides size information for files too large to preview

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| ↑/↓ | Navigate through the file list |
| Tab | Select/deselect the current file |
| Enter | Copy selected files |
| q/Esc/Ctrl+C | Quit the application |

## Project Structure

```
airule/
├── cmd/
│   └── airule/
│       └── main.go          # Entry point
├── internal/
│   ├── app/
│   │   └── app.go           # Application logic
│   ├── cli/
│   │   └── cli.go           # CLI argument handling
│   ├── copier/
│   │   └── copier.go        # File copying logic
│   ├── finder/
│   │   └── finder.go        # File finding logic
│   └── preview/
│       └── preview.go       # File preview generation
├── go.mod                   # Go module file
└── go.sum                   # Go module checksum file
```

## Dependencies

- [github.com/alecthomas/kong](https://github.com/alecthomas/kong): CLI argument parsing
- [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss): Styling for terminal applications
- [github.com/ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder): Interactive fuzzy-finding selection interface

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgements

- [Charm](https://charm.sh/) for their excellent terminal UI libraries
- [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) for the interactive selection interface
