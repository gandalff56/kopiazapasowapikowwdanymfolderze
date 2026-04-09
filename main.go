package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== Kopia Zapasowa - Monitor Folderu ===")
	fmt.Println()

	// Załaduj konfigurację
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("BLAD: %s\n", err)
		fmt.Println("\nNacisnij Enter aby zamknac...")
		fmt.Scanln()
		os.Exit(1)
	}

	// Skonfiguruj logowanie do pliku i konsoli
	logger, logFile := setupLogger()
	if logFile != nil {
		defer logFile.Close()
	}

	logger.Printf("Program uruchomiony")
	logger.Printf("Monitorowanie folderu: %s", cfg.FolderZrodlowy)
	logger.Printf("Kopie zapasowe w: %s", cfg.FolderKopii)
	logger.Printf("Interwal skanowania: %d sekund", cfg.InterwalSekundy)
	logger.Printf("Monitorowane rozszerzenia: %v", cfg.RozszerzeniaPlikow)
	fmt.Println()
	fmt.Println("Nacisnij Ctrl+C aby zatrzymac program")
	fmt.Println()

	// Utwórz monitor
	monitor := NewMonitor(cfg.FolderZrodlowy, cfg.RozszerzeniaPlikow)

	// Pierwsze skanowanie - zapamiętaj aktualny stan plików bez kopiowania
	logger.Printf("Pierwsze skanowanie - zapisywanie aktualnego stanu plikow...")
	monitor.Scan()
	// Oznacz wszystkie pliki jako stabilne po pierwszym skanie
	// żeby nie kopiować istniejących plików przy starcie
	for path, state := range monitor.knownFiles {
		state.Stable = true
		monitor.knownFiles[path] = state
	}
	logger.Printf("Znaleziono %d plikow do monitorowania", len(monitor.knownFiles))

	// Obsługa Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Pętla główna
	ticker := time.NewTicker(time.Duration(cfg.InterwalSekundy) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			changedFiles := monitor.Scan()
			for _, srcPath := range changedFiles {
				name := filepath.Base(srcPath)
				logger.Printf("Wykryto zmiane: %s", name)

				dstPath, err := BackupFile(srcPath, cfg.FolderKopii)
				if err != nil {
					logger.Printf("BLAD kopiowania %s: %s", name, err)
					continue
				}
				logger.Printf("Skopiowano do: %s", dstPath)
			}

		case <-sigChan:
			logger.Printf("Zatrzymywanie programu...")
			fmt.Println("\nProgram zatrzymany.")
			return
		}
	}
}

// setupLogger konfiguruje logowanie do konsoli i pliku backup.log
func setupLogger() (*log.Logger, *os.File) {
	dir, err := exeDir()
	if err != nil {
		// Jeśli nie można ustalić katalogu exe, loguj tylko do konsoli
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	logPath := filepath.Join(dir, "backup.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return log.New(os.Stdout, "", log.LstdFlags), nil
	}

	// Loguj do konsoli i pliku jednocześnie
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	return log.New(multiWriter, "", log.LstdFlags), logFile
}
