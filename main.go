package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	fmt.Println("=== Folder Backup Monitor ===")
	fmt.Println()

	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		fmt.Println("\nPress Enter to close...")
		fmt.Scanln()
		os.Exit(1)
	}

	logger, logFile := setupLogger()
	if logFile != nil {
		defer logFile.Close()
	}

	logger.Printf("Program started")
	logger.Printf("Watching folder: %s", cfg.FolderZrodlowy)
	logger.Printf("Backups saved to: %s", cfg.FolderKopii)
	logger.Printf("Monitored extensions: %v", cfg.RozszerzeniaPlikow)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	monitor := NewMonitor(cfg.FolderZrodlowy, cfg.RozszerzeniaPlikow, cfg.FolderKopii, logger)

	// Graceful shutdown on Ctrl+C
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Printf("Shutting down...")
		close(done)
	}()

	if err := monitor.Watch(done); err != nil {
		logger.Printf("ERROR: %s", err)
		fmt.Println("\nPress Enter to close...")
		fmt.Scanln()
		os.Exit(1)
	}

	fmt.Println("\nProgram stopped.")
}

// setupLogger configures logging to both console and backup.log file
func setupLogger() (*log.Logger, *os.File) {
	dir, err := exeDir()
	if err != nil {
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	logPath := filepath.Join(dir, "backup.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	return log.New(multiWriter, "", log.LstdFlags), logFile
}
