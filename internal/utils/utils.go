package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

const AppName = "Dedupr"

// AppDataDir returns the application data directory.
// It creates the directory first if it does not exist.
func AppDataDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find config dir: %w", err)
	}

	appDataDir := filepath.Join(userConfigDir, AppName)
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return "", fmt.Errorf("could not create app directory: %w", err)
	}

	return appDataDir, nil
}
