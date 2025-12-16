package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/artemisfowl/trash/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all trashed files",
	Long:  `Display all files and directories currently in the trash, organized by when they were trashed.`,
	Run: func(cmd *cobra.Command, args []string) {
		configDir, err := config.GetConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config directory: %v\n", err)
			os.Exit(1)
		}

		// Read all timestamped directories
		entries, err := os.ReadDir(configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading trash directory: %v\n", err)
			os.Exit(1)
		}

		// Filter and sort timestamped directories
		var trashDirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				trashDirs = append(trashDirs, entry.Name())
			}
		}
		sort.Strings(trashDirs) // Chronological order due to YYYYMMDD_HHMMSS format

		if len(trashDirs) == 0 {
			fmt.Println("Trash is empty")
			return
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		totalItems := 0

		// Process each trash directory
		for _, dirName := range trashDirs {
			dirPath := filepath.Join(configDir, dirName)
			restoreFile := filepath.Join(dirPath, ".restore")

			// Check if .restore file exists
			if _, err := os.Stat(restoreFile); os.IsNotExist(err) {
				if verbose {
					fmt.Printf("\n[%s] (no metadata)\n", dirName)
				}
				continue
			}

			// Read and parse .restore file
			data, err := os.ReadFile(restoreFile)
			if err != nil {
				if verbose {
					fmt.Printf("\n[%s] Error reading metadata: %v\n", dirName, err)
				}
				continue
			}

			var metadata config.RestoreMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				if verbose {
					fmt.Printf("\n[%s] Error parsing metadata: %v\n", dirName, err)
				}
				continue
			}

			// Display items from this trash session
			if len(metadata.Items) > 0 {
				fmt.Printf("\n[%s]\n", dirName)
				for _, item := range metadata.Items {
					totalItems++
					if verbose {
						fmt.Printf("  • %s\n", item.Name)
						fmt.Printf("    Original: %s\n", item.OriginalPath)
						fmt.Printf("    Trashed:  %s\n", item.TrashedAt)
					} else {
						fmt.Printf("  • %s (from %s)\n", item.Name, item.OriginalPath)
					}
				}
			}
		}

		fmt.Printf("\nTotal: %d item(s) in trash\n", totalItems)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
