package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupFile kopiuje plik do folderu backupu zorganizowanego wg daty
// Struktura: <folderKopii>/<YYYY-MM-DD>/<nazwa_pliku>
// Przy kolizji nazw dodaje timestamp: plik_150405.json
func BackupFile(srcPath string, backupRoot string) (string, error) {
	today := time.Now().Format("2006-01-02")
	dayDir := filepath.Join(backupRoot, today)

	if err := os.MkdirAll(dayDir, 0755); err != nil {
		return "", fmt.Errorf("nie można utworzyć folderu %s: %w", dayDir, err)
	}

	srcName := filepath.Base(srcPath)
	dstPath := filepath.Join(dayDir, srcName)

	// Jeśli plik już istnieje, dodaj timestamp do nazwy
	if _, err := os.Stat(dstPath); err == nil {
		ext := filepath.Ext(srcName)
		nameNoExt := strings.TrimSuffix(srcName, ext)
		timestamp := time.Now().Format("150405") // HHMMSS
		newName := fmt.Sprintf("%s_%s%s", nameNoExt, timestamp, ext)
		dstPath = filepath.Join(dayDir, newName)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		return "", fmt.Errorf("nie można skopiować pliku: %w", err)
	}

	return dstPath, nil
}

// copyFile kopiuje plik atomowo: zapis do pliku tymczasowego + rename
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Zapisz do pliku tymczasowego w tym samym katalogu
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

	// Przenieś plik tymczasowy na docelowy
	return os.Rename(tmpPath, dst)
}
