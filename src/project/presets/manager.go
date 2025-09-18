package presets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"main/src/modules/config"
	"main/src/project/toml"
)

// Settings represents the structure of settings.toml
type Settings struct {
	Directories struct {
		ConfigDirPath  string `toml:"config_dir_path"`
		AppDirName     string `toml:"app_dir_name"`
		PresetsDirName string `toml:"presets_dir_name"`
	} `toml:"directories"`
	Defaults struct {
		TimeoutSeconds int    `toml:"timeout_seconds"`
		MaxRetries     int    `toml:"max_retries"`
		HTTPMethod     string `toml:"http_method"`
	} `toml:"defaults"`
}

// LoadSettings loads the settings.toml file
func LoadSettings() (*Settings, error) {
	settingsPath := "src/settings/settings.toml"
	handler, err := toml.NewTomlHandler(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	settings := &Settings{}
	// Use the existing config approach for compatibility
	settings.Directories.ConfigDirPath = handler.GetAsString("directories.config_dir_path")
	settings.Directories.AppDirName = handler.GetAsString("directories.app_dir_name") 
	settings.Directories.PresetsDirName = handler.GetAsString("directories.presets_dir_name")
	
	return settings, nil
}

// GetConfigDir returns the full configuration directory path
func GetConfigDir() (string, error) {
	return config.GetPresetsDir()
}

// GetPresetPath returns the full path to a specific preset directory
func GetPresetPath(name string) (string, error) {
	presetsDir, err := config.GetPresetsDir()
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

	// Create default TOML files (empty but valid TOML)
	files := []string{"headers.toml", "body.toml", "query.toml", "config.toml"}
	for _, file := range files {
		filePath := filepath.Join(presetPath, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Create empty TOML file
			err := os.WriteFile(filePath, []byte(""), 0644)
			if err != nil {
				return fmt.Errorf("failed to create %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// ListPresets returns a list of all preset names
func ListPresets() ([]string, error) {
	presetsDir, err := config.GetPresetsDir()
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
// fileType should be one of: "headers", "body", "query", "config"
func LoadPresetFile(preset, fileType string) (*toml.TomlHandler, error) {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(presetPath, fileType+".toml")
	
	// Create preset directory and file if they don't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := CreatePresetDirectory(preset)
		if err != nil {
			return nil, err
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
	validTypes := []string{"headers", "body", "query", "config"}
	for _, valid := range validTypes {
		if strings.ToLower(fileType) == valid {
			return true
		}
	}
	return false
}