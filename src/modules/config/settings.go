package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/toml-vars-letsgooo"
)

func GetPresetsDir() (string, error) {
	configDir := tomv.GetOr("directories.config_dir", ".config")
	appDir := tomv.GetOr("directories.app_dir", "saul")
	presetsDir := tomv.GetOr("directories.presets_dir", "presets")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, configDir, appDir, presetsDir), nil
}