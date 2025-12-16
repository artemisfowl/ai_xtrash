package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RestoreItem represents metadata for a single trashed item
type RestoreItem struct {
	Name         string `json:"name"`
	OriginalPath string `json:"original_path"`
	TrashedAt    string `json:"trashed_at"`
}

// RestoreMetadata represents the .restore file structure
type RestoreMetadata struct {
	Items []RestoreItem `json:"items"`
}

// GetConfigDir returns the path to the trash config directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".config", "trash")
	return configDir, nil
}

// EnsureConfigDir ensures the trash config directory exists
// Creates it if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	
	// Check if directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Create directory with appropriate permissions (0755)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		fmt.Printf("Created config directory: %s\n", configDir)
	}
	
	return nil
}

// CreateTrashTimestampDir creates a new timestamped directory in the trash config directory
// Returns the path to the created directory
func CreateTrashTimestampDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	
	// Create timestamp in format YYYYMMDD_HHMMSS
	timestamp := time.Now().Format("20060102_150405")
	trashDir := filepath.Join(configDir, timestamp)
	
	// Create the timestamped directory
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create trash directory: %w", err)
	}
	
	return trashDir, nil
}

// MoveToTrash moves a file or directory to the specified trash directory
// Returns the basename of the moved item for metadata tracking
func MoveToTrash(sourcePath, trashDir string) (string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Check if source exists
	sourceInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat source: %w", err)
	}
	
	// Get the base name of the file/directory
	baseName := filepath.Base(absPath)
	destPath := filepath.Join(trashDir, baseName)
	
	// Try to move the file/directory using rename first (fast)
	err = os.Rename(absPath, destPath)
	if err == nil {
		return baseName, nil // Success!
	}
	
	// If rename failed due to cross-device link, copy and delete instead
	if sourceInfo.IsDir() {
		// For directories, use recursive copy
		if err := CopyDir(absPath, destPath); err != nil {
			return "", fmt.Errorf("failed to copy directory %s to trash: %w", absPath, err)
		}
		// Remove original directory after successful copy
		if err := os.RemoveAll(absPath); err != nil {
			return "", fmt.Errorf("failed to remove original directory %s: %w", absPath, err)
		}
	} else {
		// For files, use simple copy
		if err := CopyFile(absPath, destPath); err != nil {
			return "", fmt.Errorf("failed to copy file %s to trash: %w", absPath, err)
		}
		// Remove original file after successful copy
		if err := os.Remove(absPath); err != nil {
			return "", fmt.Errorf("failed to remove original file %s: %w", absPath, err)
		}
	}
	
	return baseName, nil
}

// SaveRestoreMetadata saves the restore metadata to a .restore file in the trash directory
func SaveRestoreMetadata(trashDir string, metadata *RestoreMetadata) error {
	restoreFilePath := filepath.Join(trashDir, ".restore")
	
	// Marshal metadata to JSON with indentation
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	// Write to .restore file
	if err := os.WriteFile(restoreFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write .restore file: %w", err)
	}
	
	return nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	// Copy the contents
	if _, err := destFile.ReadFrom(sourceFile); err != nil {
		return err
	}
	
	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// CopyDir recursively copies a directory from src to dst
func CopyDir(src, dst string) error {
	// Get source directory info
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	// Create destination directory
	if err := os.MkdirAll(dst, sourceInfo.Mode()); err != nil {
		return err
	}
	
	// Read directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}
