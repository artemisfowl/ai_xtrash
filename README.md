# Trash

A CLI application for moving files and directories to a trash directory with timestamp-based organization.

## Features

- **Trash Management**: Move files and directories to `~/.config/trash` instead of permanently deleting
- **Timestamp Organization**: Each trash operation creates a timestamped subdirectory for easy tracking
- **Multiple Items**: Trash multiple files/directories in a single command
- **Restore Metadata**: Track original file locations in `.restore` JSON files
- **List Trashed Items**: View all items currently in trash with their original paths
- **Subcommands**: Version info and other utilities
- Built with [Cobra](https://github.com/spf13/cobra) - a powerful CLI framework

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from source

```bash
# Clone or navigate to the project directory
cd /media/oldgod/storage/storage/data/code/test_bed/skunkworks/gravity/go

# Download dependencies
go mod download

# Build the binary
go build -o trash .

# Or install to $GOPATH/bin
go install .
```

## Usage

### Basic Trash Operations

```bash
# Trash a single file
./trash /path/to/file.txt

# Trash multiple files
./trash file1.txt file2.txt file3.txt

# Trash directories (including all contents)
./trash /path/to/directory

# Trash mixed files and directories
./trash file.txt /path/to/dir another_file.log

# Use verbose mode to see details
./trash --verbose file.txt
./trash -v file1.txt file2.txt
```

### List Trashed Items

```bash
# List all trashed items
./trash list

# List with detailed information (verbose)
./trash list --verbose
```

### Subcommands

```bash
# Show version information
./trash version

# Show help
./trash --help
```

### Examples

```bash
# Trash a single file
./trash document.pdf
# Output: Successfully moved 1 item(s) to trash

# Trash multiple items with verbose output
./trash --verbose old_project/ notes.txt backup.tar.gz
# Output:
# Created trash directory: /home/user/.config/trash/20251217_005131
# Moved to trash: /path/to/old_project/
# Moved to trash: /path/to/notes.txt
# Moved to trash: /path/to/backup.tar.gz
# Successfully moved 3 item(s) to trash

# List trashed items
./trash list
# Output:
# [20251217_010006]
#   • test1.txt (from /path/to/test1.txt)
#   • test2.txt (from /path/to/test2.txt)
#   • testdir (from /path/to/testdir)
# 
# Total: 3 item(s) in trash
```

## Project Structure

```
.
├── main.go           # Entry point
├── cmd/              # Command definitions
│   ├── root.go       # Root command and global flags
│   ├── hello.go      # Example hello command
│   └── version.go    # Version command
├── go.mod            # Go module definition
├── go.sum            # Dependency checksums (generated)
└── README.md         # This file
```

## Adding New Commands

To add a new command:

1. Create a new file in the `cmd/` directory (e.g., `cmd/mycommand.go`)
2. Define your command using cobra's structure
3. Register it with the root command in the `init()` function

Example:

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Brief description",
    Long:  "Longer description",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("My command executed!")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

## Development

### Run tests

```bash
go test ./...
```

### Format code

```bash
go fmt ./...
```

### Lint code

```bash
# Install golangci-lint first
golangci-lint run
```

## License

MIT License (or your preferred license)
