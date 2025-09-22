package presets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/config"
)

// getPresetsDir returns the presets directory path using environment variables with defaults
func getPresetsDir() (string, error) {
	// Use environment variables with defaults (replaces TOMV)
	configDirPath := getEnvOrDefault("SAUL_CONFIG_DIR_PATH", ".config")
	appDirName := getEnvOrDefault("SAUL_APP_DIR_NAME", "saul")
	presetsDirName := getEnvOrDefault("SAUL_PRESETS_DIR_NAME", "presets")

	// Build full path relative to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf(errors.ErrDirectoryFailed)
	}

	return filepath.Join(homeDir, configDirPath, appDirName, presetsDirName), nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

// GetConfigDir returns the full configuration directory path
func GetConfigDir() (string, error) {
	return getPresetsDir()
}

// GetPresetPath returns the full path to a specific preset directory
func GetPresetPath(name string) (string, error) {
	presetsDir, err := getPresetsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(presetsDir, name), nil
}

// CreatePresetDirectory creates a new preset directory with default TOML files
func CreatePresetDirectory(name string) error {
	presetPath, err := GetPresetPath(name)
	if err != nil {
		return err
	}

	// Create preset directory
	err = os.MkdirAll(presetPath, config.DirPermissions)
	if err != nil {
		return fmt.Errorf(errors.ErrDirectoryFailed)
	}

	// Don't create any TOML files initially
	// Files will be created on-demand when data is actually added

	return nil
}

// ListPresets returns a list of all preset names
func ListPresets() ([]string, error) {
	presetsDir, err := getPresetsDir()
	if err != nil {
		return nil, err
	}

	// Create presets directory if it doesn't exist
	err = os.MkdirAll(presetsDir, config.DirPermissions)
	if err != nil {
		return nil, fmt.Errorf(errors.ErrDirectoryFailed)
	}

	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		return nil, fmt.Errorf(errors.ErrDirectoryFailed)
	}

	var presets []string
	for _, entry := range entries {
		if entry.IsDir() {
			presets = append(presets, entry.Name())
		}
	}

	return presets, nil
}

// DeletePreset removes a preset directory and all its files
func DeletePreset(name string) error {
	presetPath, err := GetPresetPath(name)
	if err != nil {
		return err
	}

	// Check if preset exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf(errors.ErrPresetNotFound, name)
	}

	// Remove the entire preset directory
	err = os.RemoveAll(presetPath)
	if err != nil {
		return fmt.Errorf(errors.ErrDirectoryFailed)
	}

	return nil
}

