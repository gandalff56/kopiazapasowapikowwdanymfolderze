package main

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Monitor watches a folder for file changes using OS-level events
type Monitor struct {
	sourceDir  string
	extensions map[string]bool
	backupRoot string
	logger     *log.Logger
	debounce   map[string]*time.Timer
	mu         sync.Mutex
}

// NewMonitor creates a new event-based monitor
func NewMonitor(sourceDir string, extensions []string, backupRoot string, logger *log.Logger) *Monitor {
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}
	return &Monitor{
		sourceDir:  sourceDir,
		extensions: extMap,
		backupRoot: backupRoot,
		logger:     logger,
		debounce:   make(map[string]*time.Timer),
	}
}

// Watch starts watching the folder for changes (blocking)
// It uses OS-level filesystem notifications (ReadDirectoryChangesW on Windows)
// which consume near-zero CPU when idle
func (m *Monitor) Watch(done <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(m.sourceDir); err != nil {
		return err
	}

	m.logger.Printf("Watching folder: %s", m.sourceDir)
	m.logger.Printf("Backups saved to: %s", m.backupRoot)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			m.handleEvent(event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			m.logger.Printf("ERROR: watcher error: %s", err)

		case <-done:
			m.logger.Printf("Stopping watcher...")
			return nil
		}
	}
}

// handleEvent processes a single filesystem event
func (m *Monitor) handleEvent(event fsnotify.Event) {
	// Only react to create and write events
	if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) {
		return
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(event.Name))
	if !m.extensions[ext] {
		return
	}

	// Debounce: wait 2 seconds after last change before copying
	// This prevents copying while the file is still being written
	m.mu.Lock()
	if timer, exists := m.debounce[event.Name]; exists {
		timer.Stop()
	}
	m.debounce[event.Name] = time.AfterFunc(2*time.Second, func() {
		m.backupFile(event.Name)
		m.mu.Lock()
		delete(m.debounce, event.Name)
		m.mu.Unlock()
	})
	m.mu.Unlock()
}

// backupFile copies a changed file to the backup folder
func (m *Monitor) backupFile(srcPath string) {
	name := filepath.Base(srcPath)
	m.logger.Printf("Change detected: %s", name)

	dstPath, err := BackupFile(srcPath, m.backupRoot)
	if err != nil {
		m.logger.Printf("ERROR copying %s: %s", name, err)
		return
	}
	m.logger.Printf("Backed up to: %s", dstPath)
}
