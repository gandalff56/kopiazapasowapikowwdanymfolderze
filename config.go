package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds program configuration
type Config struct {
	FolderZrodlowy     string   `json:"folder_zrodlowy"`
	FolderKopii        string   `json:"folder_kopii"`
	InterwalSekundy    int      `json:"interwal_sekundy"`
	RozszerzeniaPlikow []string `json:"rozszerzenia_plikow"`
}

func defaultConfig() Config {
	return Config{
		FolderZrodlowy:     "D:/Polaris/data/EdgeMillData",
		FolderKopii:        "D:/KopieZapasowe/EdgeMillData",
		InterwalSekundy:    10,
		RozszerzeniaPlikow: []string{".json"},
	}
}

func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot determine executable path: %w", err)
	}
	return filepath.Dir(exe), nil
}

func configPath() (string, error) {
	dir, err := exeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if createErr := createDefaultConfig(path); createErr != nil {
				return Config{}, fmt.Errorf("cannot create default config.json: %w", createErr)
			}
			return Config{}, fmt.Errorf(
				"created default config file: %s\n"+
					"Edit it and restart the program", path)
		}
		return Config{}, fmt.Errorf("cannot read config.json: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config.json parse error: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func createDefaultConfig(path string) error {
	cfg := defaultConfig()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func validateConfig(cfg *Config) error {
	if cfg.FolderZrodlowy == "" {
		return fmt.Errorf("folder_zrodlowy cannot be empty")
	}

	info, err := os.Stat(cfg.FolderZrodlowy)
	if err != nil {
		return fmt.Errorf("source folder does not exist: %s", cfg.FolderZrodlowy)
	}
	if !info.IsDir() {
		return fmt.Errorf("folder_zrodlowy is not a directory: %s", cfg.FolderZrodlowy)
	}

	if cfg.FolderKopii == "" {
		return fmt.Errorf("folder_kopii cannot be empty")
	}

	if err := os.MkdirAll(cfg.FolderKopii, 0755); err != nil {
		return fmt.Errorf("cannot create backup folder: %w", err)
	}

	if len(cfg.RozszerzeniaPlikow) == 0 {
		cfg.RozszerzeniaPlikow = []string{".json"}
	}

	return nil
}
