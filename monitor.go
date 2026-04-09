package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileState przechowuje stan pliku z ostatniego skanowania
type FileState struct {
	ModTime time.Time
	Size    int64
	Stable  bool // true jeśli plik nie zmienił się między dwoma skanami
}

// Monitor śledzi zmiany w folderze źródłowym
type Monitor struct {
	sourceDir  string
	extensions map[string]bool
	knownFiles map[string]FileState
}

// NewMonitor tworzy nowy monitor
func NewMonitor(sourceDir string, extensions []string) *Monitor {
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}
	return &Monitor{
		sourceDir:  sourceDir,
		extensions: extMap,
		knownFiles: make(map[string]FileState),
	}
}

// Scan skanuje folder i zwraca listę ścieżek plików gotowych do backupu
// Plik jest "gotowy" dopiero gdy jego ModTime i Size nie zmieniły się
// między dwoma kolejnymi skanami (zabezpieczenie przed kopiowaniem w trakcie zapisu)
func (m *Monitor) Scan() []string {
	var ready []string

	entries, err := os.ReadDir(m.sourceDir)
	if err != nil {
		return nil
	}

	currentFiles := make(map[string]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !m.extensions[ext] {
			continue
		}

		fullPath := filepath.Join(m.sourceDir, name)
		currentFiles[fullPath] = true

		info, err := entry.Info()
		if err != nil {
			continue
		}

		modTime := info.ModTime()
		size := info.Size()

		prev, exists := m.knownFiles[fullPath]

		if !exists {
			// Nowy plik - zapamiętaj stan, będzie gotowy przy następnym skanie
			m.knownFiles[fullPath] = FileState{
				ModTime: modTime,
				Size:    size,
				Stable:  false,
			}
			continue
		}

		if prev.ModTime.Equal(modTime) && prev.Size == size {
			// Plik się nie zmienił od ostatniego skanu
			if !prev.Stable {
				// Pierwszy raz stabilny - oznacz jako gotowy do backupu
				m.knownFiles[fullPath] = FileState{
					ModTime: modTime,
					Size:    size,
					Stable:  true,
				}
				ready = append(ready, fullPath)
			}
			// Jeśli już był stabilny - nic nie robimy (już skopiowany)
		} else {
			// Plik się zmienił - zresetuj stan
			m.knownFiles[fullPath] = FileState{
				ModTime: modTime,
				Size:    size,
				Stable:  false,
			}
		}
	}

	// Usuń pliki które już nie istnieją
	for path := range m.knownFiles {
		if !currentFiles[path] {
			delete(m.knownFiles, path)
		}
	}

	return ready
}
