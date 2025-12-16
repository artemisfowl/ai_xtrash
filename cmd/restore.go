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

var restoreCmd = &cobra.Command{
	Use:   "restore [item-name]",
	Short: "Restore a trashed file or directory",
	Long: `Restore a file or directory from trash back to its original location.
If multiple items with the same name exist, the most recently trashed one will be restored.
Use --all flag to see all matches and choose, or --timestamp to specify which one.

Examples:
  trash restore test1.txt
  trash restore testdir
  trash restore test1.txt --timestamp 20251217_010006`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		itemName := args[0]
		specifiedTimestamp, _ := cmd.Flags().GetString("timestamp")
		showAll, _ := cmd.Flags().GetBool("all")
		verbose, _ := cmd.Flags().GetBool("verbose")
		force, _ := cmd.Flags().GetBool("force")

		configDir, err := config.GetConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config directory: %v\n", err)
			os.Exit(1)
		}

		// Find all instances of the item in trash
		type MatchedItem struct {
			Timestamp    string
			Item         config.RestoreItem
			TrashDirPath string
		}

		var matches []MatchedItem

		// Read all timestamped directories
		entries, err := os.ReadDir(configDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading trash directory: %v\n", err)
			os.Exit(1)
		}

		// Sort directories (newest first for default behavior)
		var trashDirs []string
		for _, entry := range entries {
			if entry.IsDir() {
				trashDirs = append(trashDirs, entry.Name())
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(trashDirs)))

		// Search for the item
		for _, dirName := range trashDirs {
			// If timestamp specified, only check that directory
			if specifiedTimestamp != "" && dirName != specifiedTimestamp {
				continue
			}

			dirPath := filepath.Join(configDir, dirName)
			restoreFile := filepath.Join(dirPath, ".restore")

			// Check if .restore file exists
			if _, err := os.Stat(restoreFile); os.IsNotExist(err) {
				continue
			}

			// Read and parse .restore file
			data, err := os.ReadFile(restoreFile)
			if err != nil {
				continue
			}

			var metadata config.RestoreMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				continue
			}

			// Look for matching item
			for _, item := range metadata.Items {
				if item.Name == itemName {
					matches = append(matches, MatchedItem{
						Timestamp:    dirName,
						Item:         item,
						TrashDirPath: dirPath,
					})
				}
			}
		}

		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "Error: item '%s' not found in trash\n", itemName)
			os.Exit(1)
		}

		// Handle multiple matches
		if len(matches) > 1 {
			if showAll {
				fmt.Printf("Found %d instances of '%s':\n\n", len(matches), itemName)
				for i, match := range matches {
					fmt.Printf("%d. [%s]\n", i+1, match.Timestamp)
					fmt.Printf("   Original: %s\n", match.Item.OriginalPath)
					fmt.Printf("   Trashed:  %s\n\n", match.Item.TrashedAt)
				}
				fmt.Println("Use --timestamp flag to specify which one to restore")
				fmt.Printf("Example: trash restore %s --timestamp %s\n", itemName, matches[0].Timestamp)
				return
			}

			if specifiedTimestamp == "" {
				fmt.Printf("Found %d instances of '%s'. Restoring the most recent one.\n", len(matches), itemName)
				fmt.Printf("Use --all to see all matches or --timestamp to specify which one.\n\n")
			}
		}

		// Restore the first match (most recent if not specified)
		match := matches[0]
		timestamp := match.Timestamp
		trashDir := match.TrashDirPath
		itemToRestore := match.Item

		// Source and destination paths
		sourcePath := filepath.Join(trashDir, itemName)
		destPath := itemToRestore.OriginalPath

		// Check if destination already exists
		if _, err := os.Stat(destPath); err == nil {
			if !force {
				fmt.Fprintf(os.Stderr, "Error: destination already exists: %s\n", destPath)
				fmt.Fprintf(os.Stderr, "Use --force to overwrite\n")
				os.Exit(1)
			}
			if verbose {
				fmt.Printf("Overwriting existing file/directory: %s\n", destPath)
			}
			// Remove existing destination
			if err := os.RemoveAll(destPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing existing destination: %v\n", err)
				os.Exit(1)
			}
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(destPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating parent directory: %v\n", err)
			os.Exit(1)
		}

		// Try to move using rename first
		err = os.Rename(sourcePath, destPath)
		if err == nil {
			if verbose {
				fmt.Printf("Restored: %s -> %s\n", itemName, destPath)
			}
		} else {
			// Fallback to copy and delete for cross-device
			sourceInfo, err := os.Stat(sourcePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accessing source: %v\n", err)
				os.Exit(1)
			}

			if sourceInfo.IsDir() {
				if err := config.CopyDir(sourcePath, destPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error copying directory: %v\n", err)
					os.Exit(1)
				}
			} else {
				if err := config.CopyFile(sourcePath, destPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error copying file: %v\n", err)
					os.Exit(1)
				}
			}

			// Remove from trash after successful copy
			if err := os.RemoveAll(sourcePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove from trash: %v\n", err)
			}

			if verbose {
				fmt.Printf("Restored (copied): %s -> %s\n", itemName, destPath)
			}
		}

		// Update metadata to remove restored item
		restoreFile := filepath.Join(trashDir, ".restore")
		data, _ := os.ReadFile(restoreFile)
		var metadata config.RestoreMetadata
		json.Unmarshal(data, &metadata)

		var updatedItems []config.RestoreItem
		for _, item := range metadata.Items {
			if item.Name != itemName {
				updatedItems = append(updatedItems, item)
			}
		}

		if len(updatedItems) == 0 {
			// No items left, remove the entire trash directory
			if err := os.RemoveAll(trashDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove empty trash directory: %v\n", err)
			}
			if verbose {
				fmt.Printf("Removed empty trash directory: %s\n", timestamp)
			}
		} else {
			// Update .restore file with remaining items
			metadata.Items = updatedItems
			if err := config.SaveRestoreMetadata(trashDir, &metadata); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update metadata: %v\n", err)
			}
		}

		fmt.Printf("Successfully restored: %s\n", destPath)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolP("force", "f", false, "Overwrite destination if it exists")
	restoreCmd.Flags().String("timestamp", "", "Specify which timestamp to restore from")
	restoreCmd.Flags().Bool("all", false, "Show all matches without restoring")
}
