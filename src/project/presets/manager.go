package presets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
	// "github.com/DeprecatedLuar/toml-vars-letsgooo"
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
		return "", fmt.Errorf("failed to get home directory: %w", err)
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
	err = os.MkdirAll(presetPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create preset directory %s: %w", presetPath, err)
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
	err = os.MkdirAll(presetsDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create presets directory: %w", err)
	}

	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read presets directory: %w", err)
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
		return fmt.Errorf("preset '%s' does not exist", name)
	}

	// Remove the entire preset directory
	err = os.RemoveAll(presetPath)
	if err != nil {
		return fmt.Errorf("failed to delete preset '%s': %w", name, err)
	}

	return nil
}

// LoadPresetFile loads a specific TOML file from a preset
// Creates the file if it doesn't exist (lazy creation)
func LoadPresetFile(preset, fileType string) (*toml.TomlHandler, error) {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return nil, err
	}

	// Ensure preset directory exists
	err = os.MkdirAll(presetPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create preset directory %s: %w", presetPath, err)
	}

	filePath := filepath.Join(presetPath, fileType+".toml")
	
	// Create empty TOML file if it doesn't exist (lazy creation)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := os.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s: %w", filePath, err)
		}
	}

	return toml.NewTomlHandler(filePath)
}

// SavePresetFile saves a TOML handler to a specific preset file
func SavePresetFile(preset, fileType string, handler *toml.TomlHandler) error {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return err
	}

	filePath := filepath.Join(presetPath, fileType+".toml")
	handler.SetOutputPath(filePath)
	return handler.Write()
}

// ValidateFileType checks if the file type is valid
func ValidateFileType(fileType string) bool {
	validTypes := []string{"headers", "body", "query", "request", "variables"}
	for _, valid := range validTypes {
		if strings.ToLower(fileType) == valid {
			return true
		}
	}
	return false
}