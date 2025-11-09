package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/DeprecatedLuar/better-curl-saul/internal"
)


// GetConfigDir returns the full configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, internal.ParentDirPath, internal.AppDirName, internal.PresetsDirName), nil
}

// GetPresetPath returns the full path to a specific preset directory
func GetPresetPath(name string) (string, error) {
	presetsDir, err := GetConfigDir()
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
	err = os.MkdirAll(presetPath, internal.DirPermissions)
	if err != nil {
		return fmt.Errorf(display.ErrDirectoryFailed)
	}

	// Don't create any TOML files initially
	// Files will be created on-demand when data is actually added

	return nil
}

// ListPresets returns a list of all preset names
func ListPresets() ([]string, error) {
	presetsDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	// Create presets directory if it doesn't exist
	err = os.MkdirAll(presetsDir, internal.DirPermissions)
	if err != nil {
		return nil, fmt.Errorf(display.ErrDirectoryFailed)
	}

	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		return nil, fmt.Errorf(display.ErrDirectoryFailed)
	}

	var presets []string
	for _, entry := range entries {
		if entry.IsDir() {
			presets = append(presets, entry.Name())
		}
	}

	return presets, nil
}

// PresetExists checks if a preset directory exists
func PresetExists(name string) bool {
	presetPath, err := GetPresetPath(name)
	if err != nil {
		return false
	}

	_, err = os.Stat(presetPath)
	return !os.IsNotExist(err)
}

// DeletePreset removes a preset directory and all its files
func DeletePreset(name string) error {
	presetPath, err := GetPresetPath(name)
	if err != nil {
		return err
	}

	// Check if preset exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf(display.ErrPresetNotFound, name)
	}

	// Remove the entire preset directory
	err = os.RemoveAll(presetPath)
	if err != nil {
		return fmt.Errorf(display.ErrDirectoryFailed)
	}

	return nil
}

// GetActiveVariant returns the active variant name from .config file
// Returns "default" if .config doesn't exist or contains invalid variant
func GetActiveVariant(preset string) string {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return "default"
	}

	configPath := filepath.Join(presetPath, ".config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "default"
	}

	variant := filepath.Clean(filepath.Base(string(data)))
	if variant == "" || variant == "." {
		return "default"
	}

	variantsDir := filepath.Join(presetPath, "variants")
	variantPath := filepath.Join(variantsDir, variant)

	if _, err := os.Stat(variantPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: variant '%s' from .config does not exist, using 'default'\n", variant)
		return "default"
	}

	return variant
}

// SetActiveVariant writes the variant name to .config file
func SetActiveVariant(preset, variant string) error {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return err
	}

	variantsDir := filepath.Join(presetPath, "variants")
	variantPath := filepath.Join(variantsDir, variant)

	if _, err := os.Stat(variantPath); os.IsNotExist(err) {
		return fmt.Errorf("variant '%s' does not exist", variant)
	}

	configPath := filepath.Join(presetPath, ".config")
	return os.WriteFile(configPath, []byte(variant), internal.FilePermissions)
}
