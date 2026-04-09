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
	"time"
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

	// Create stop signal file path and clean up any leftover
	stopFile := stopFilePath()
	os.Remove(stopFile)

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

	done := make(chan struct{})

	// Shutdown on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Printf("Shutting down (signal)...")
		close(done)
	}()

	// Shutdown on stop file creation (for StopBackup.bat)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				if _, err := os.Stat(stopFile); err == nil {
					os.Remove(stopFile)
					logger.Printf("Shutting down (stop file detected)...")
					close(done)
					return
				}
				time.Sleep(1 * time.Second)
			}
		}
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

// stopFilePath returns the path to the stop signal file
func stopFilePath() string {
	dir, err := exeDir()
	if err != nil {
		return "backup.stop"
	}
	return filepath.Join(dir, "backup.stop")
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
		return log.New(logFile, "", log.LstdFlags), logFile
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	return log.New(multiWriter, "", log.LstdFlags), logFile
}
