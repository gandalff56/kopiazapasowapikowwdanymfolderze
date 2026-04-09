package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var silent bool

func main() {
	flag.BoolVar(&silent, "silent", false, "run in background without console output")
	flag.Parse()

	if !silent {
		fmt.Println("=== Folder Backup Monitor ===")
		fmt.Println()
	}

	cfg, err := loadConfig()
	if err != nil {
		if silent {
			// In silent mode, log error to file and exit
			if logger, _ := setupLogger(); logger != nil {
				logger.Printf("ERROR: %s", err)
			}
		} else {
			fmt.Printf("ERROR: %s\n", err)
			fmt.Println("\nPress Enter to close...")
			fmt.Scanln()
		}
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
	if silent {
		logger.Printf("Running in silent mode (no console)")
	} else {
		fmt.Println("Press Ctrl+C to stop")
		fmt.Println()
	}

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
		if !silent {
			fmt.Println("\nPress Enter to close...")
			fmt.Scanln()
		}
		os.Exit(1)
	}

	if !silent {
		fmt.Println("\nProgram stopped.")
	}
}

// setupLogger configures logging to file (and console if not silent)
func setupLogger() (*log.Logger, *os.File) {
	dir, err := exeDir()
	if err != nil {
		if silent {
			return log.New(io.Discard, "", log.LstdFlags), nil
		}
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	logPath := filepath.Join(dir, "backup.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		if silent {
			return log.New(io.Discard, "", log.LstdFlags), nil
		}
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	if silent {
		// Silent mode: log only to file
		return log.New(logFile, "", log.LstdFlags), logFile
	}

	// Normal mode: log to console + file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	return log.New(multiWriter, "", log.LstdFlags), logFile
}
