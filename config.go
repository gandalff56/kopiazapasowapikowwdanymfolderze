package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config przechowuje konfigurację programu
type Config struct {
	FolderZrodlowy     string   `json:"folder_zrodlowy"`
	FolderKopii        string   `json:"folder_kopii"`
	InterwalSekundy    int      `json:"interwal_sekundy"`
	RozszerzeniaPlikow []string `json:"rozszerzenia_plikow"`
}

// defaultConfig zwraca domyślną konfigurację
func defaultConfig() Config {
	return Config{
		FolderZrodlowy:     "D:/Polaris/data/EdgeMillData",
		FolderKopii:        "D:/KopieZapasowe/EdgeMillData",
		InterwalSekundy:    10,
		RozszerzeniaPlikow: []string{".json"},
	}
}

// exeDir zwraca katalog w którym znajduje się plik .exe
func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("nie można ustalić ścieżki programu: %w", err)
	}
	return filepath.Dir(exe), nil
}

// configPath zwraca pełną ścieżkę do config.json
func configPath() (string, error) {
	dir, err := exeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// loadConfig ładuje konfigurację z config.json
// Jeśli plik nie istnieje, tworzy domyślny i zwraca błąd z informacją
func loadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if createErr := createDefaultConfig(path); createErr != nil {
				return Config{}, fmt.Errorf("nie można utworzyć domyślnego config.json: %w", createErr)
			}
			return Config{}, fmt.Errorf(
				"utworzono domyślny plik konfiguracyjny: %s\n"+
					"Edytuj go i uruchom program ponownie", path)
		}
		return Config{}, fmt.Errorf("nie można odczytać config.json: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("błąd parsowania config.json: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// createDefaultConfig tworzy domyślny plik config.json
func createDefaultConfig(path string) error {
	cfg := defaultConfig()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// validateConfig sprawdza poprawność konfiguracji
func validateConfig(cfg *Config) error {
	if cfg.FolderZrodlowy == "" {
		return fmt.Errorf("folder_zrodlowy nie może być pusty")
	}

	info, err := os.Stat(cfg.FolderZrodlowy)
	if err != nil {
		return fmt.Errorf("folder źródłowy nie istnieje: %s", cfg.FolderZrodlowy)
	}
	if !info.IsDir() {
		return fmt.Errorf("folder_zrodlowy nie jest katalogiem: %s", cfg.FolderZrodlowy)
	}

	if cfg.FolderKopii == "" {
		return fmt.Errorf("folder_kopii nie może być pusty")
	}

	// Utwórz folder kopii jeśli nie istnieje
	if err := os.MkdirAll(cfg.FolderKopii, 0755); err != nil {
		return fmt.Errorf("nie można utworzyć folderu kopii: %w", err)
	}

	if cfg.InterwalSekundy <= 0 {
		cfg.InterwalSekundy = 10
	}

	if len(cfg.RozszerzeniaPlikow) == 0 {
		cfg.RozszerzeniaPlikow = []string{".json"}
	}

	return nil
}
