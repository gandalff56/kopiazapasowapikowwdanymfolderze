package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupFile copies a file to a date-organized backup folder
// Structure: <backupRoot>/<YYYY-MM-DD>/<filename>
// On name collision appends timestamp: file_150405.json
func BackupFile(srcPath string, backupRoot string) (string, error) {
	// Skip empty files (0 bytes)
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	if info.Size() == 0 {
		return "", fmt.Errorf("skipped (0 bytes)")
	}

	today := time.Now().Format("2006-01-02")
	dayDir := filepath.Join(backupRoot, today)

	if err := os.MkdirAll(dayDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create folder %s: %w", dayDir, err)
	}

	srcName := filepath.Base(srcPath)
	dstPath := filepath.Join(dayDir, srcName)

	// If file already exists, append timestamp to name
	if _, err := os.Stat(dstPath); err == nil {
		ext := filepath.Ext(srcName)
		nameNoExt := strings.TrimSuffix(srcName, ext)
		timestamp := time.Now().Format("150405") // HHMMSS
		newName := fmt.Sprintf("%s_%s%s", nameNoExt, timestamp, ext)
		dstPath = filepath.Join(dayDir, newName)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		return "", fmt.Errorf("cannot copy file: %w", err)
	}

	return dstPath, nil
}

// CleanOldBackups removes backup folders older than maxDays
func CleanOldBackups(backupRoot string, maxDays int, logger interface{ Printf(string, ...any) }) {
	cutoff := time.Now().AddDate(0, 0, -maxDays)

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse folder name as date (YYYY-MM-DD)
		folderDate, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue // skip folders that aren't date-named
		}

		if folderDate.Before(cutoff) {
			path := filepath.Join(backupRoot, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				logger.Printf("ERROR deleting old backup %s: %s", entry.Name(), err)
			} else {
				logger.Printf("Deleted old backup: %s", entry.Name())
			}
		}
	}
}

// copyFile copies a file atomically: write to temp file + rename
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Write to temp file in the same directory
	tmpPath := dst + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(tmpFile, srcFile)
	if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Rename temp file to destination
	return os.Rename(tmpPath, dst)
}
