package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/artemisfowl/trash/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "trash [file/directory paths...]",
	Short: "Move files or directories to trash",
	Long: `Trash is a CLI application that moves files and directories to a trash directory.
Files are moved to ~/.config/trash in timestamped subdirectories.

When called without arguments, shows a welcome message.
When called with file/directory paths, moves them to trash.

Use subcommands for additional functionality like version info.`,
	Args:                  cobra.ArbitraryArgs,
	DisableFlagParsing:    false,
	FParseErrWhitelist:    cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, show welcome message
		if len(args) == 0 {
			fmt.Println("Welcome to Trash! Use --help to see available commands.")
			fmt.Println("Usage: trash [file/directory paths...] to move items to trash")
			return
		}

		// Handle trash operation
		verbose, _ := cmd.Flags().GetBool("verbose")
		
		// Create a timestamped directory for this trash operation
		trashDir, err := config.CreateTrashTimestampDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating trash directory: %v\n", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Printf("Created trash directory: %s\n", trashDir)
		}

		// Track success and failures
		successCount := 0
		failedPaths := []string{}
		
		// Prepare restore metadata
		metadata := &config.RestoreMetadata{
			Items: []config.RestoreItem{},
		}

		// Move each specified path to trash
		for _, path := range args {
			// Get absolute path for metadata
			absPath, err := os.Getwd()
			if err == nil {
				absPath, _ = filepath.Abs(path)
			} else {
				absPath = path
			}
			
			baseName, err := config.MoveToTrash(path, trashDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				failedPaths = append(failedPaths, path)
			} else {
				successCount++
				if verbose {
					fmt.Printf("Moved to trash: %s\n", path)
				}
				
				// Add to metadata
				metadata.Items = append(metadata.Items, config.RestoreItem{
					Name:         baseName,
					OriginalPath: absPath,
					TrashedAt:    time.Now().Format(time.RFC3339),
				})
			}
		}

		// Save restore metadata
		if len(metadata.Items) > 0 {
			if err := config.SaveRestoreMetadata(trashDir, metadata); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save restore metadata: %v\n", err)
			}
		}

		// Summary
		if successCount > 0 {
			fmt.Printf("Successfully moved %d item(s) to trash\n", successCount)
		}
		
		if len(failedPaths) > 0 {
			fmt.Fprintf(os.Stderr, "Failed to trash %d item(s)\n", len(failedPaths))
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Ensure config directory exists before executing any commands
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}
