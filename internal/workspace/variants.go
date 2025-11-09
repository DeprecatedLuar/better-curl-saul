package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

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

// EnsureVariantStructure creates variants/ folder and variant directory if preset contains /
// On first variant creation, migrates root TOML files into the variant
func EnsureVariantStructure(presetWithVariant string) error {
	parts := strings.Split(presetWithVariant, "/")
	if len(parts) != 2 {
		return fmt.Errorf(display.ErrInvalidVariantPath, presetWithVariant)
	}

	basePreset := parts[0]
	variantName := parts[1]

	// Ensure base preset exists
	if !PresetExists(basePreset) {
		return fmt.Errorf(display.ErrVariantPresetMissing, basePreset)
	}

	presetPath, err := GetPresetPath(basePreset)
	if err != nil {
		return err
	}

	variantsDir := filepath.Join(presetPath, "variants")
	isFirstVariant := false

	// Check if this is first variant creation
	if _, err := os.Stat(variantsDir); os.IsNotExist(err) {
		isFirstVariant = true
	}

	// Create variants/ folder
	if err := os.MkdirAll(variantsDir, internal.DirPermissions); err != nil {
		return fmt.Errorf(display.ErrVariantsDirFailed, err)
	}

	// Create variant directory
	variantPath := filepath.Join(variantsDir, variantName)
	if err := os.MkdirAll(variantPath, internal.DirPermissions); err != nil {
		return fmt.Errorf(display.ErrVariantDirFailed, err)
	}

	// Migrate root TOML files to first variant
	if isFirstVariant {
		tomlFiles := []string{"request.toml", "body.toml", "headers.toml", "query.toml", "variables.toml", "filters.toml"}
		for _, file := range tomlFiles {
			rootFile := filepath.Join(presetPath, file)
			if _, err := os.Stat(rootFile); err == nil {
				// File exists, move it
				variantFile := filepath.Join(variantPath, file)
				if err := os.Rename(rootFile, variantFile); err != nil {
					return fmt.Errorf(display.ErrVariantMigrateFailed, file, err)
				}
			}
		}
	}

	// Create/update .config file to point to this variant
	configPath := filepath.Join(presetPath, ".config")
	if err := os.WriteFile(configPath, []byte(variantName), internal.FilePermissions); err != nil {
		return fmt.Errorf(display.ErrVariantConfigFailed, err)
	}

	return nil
}

// SwitchVariant switches to a different variant and updates session
func SwitchVariant(basePreset, variantName string, sessionManager *SessionManager) error {
	// Ensure variants folder and variant directory exist
	fullPresetPath := basePreset + "/" + variantName
	if err := EnsureVariantStructure(fullPresetPath); err != nil {
		return err
	}

	// Update .config file
	if err := SetActiveVariant(basePreset, variantName); err != nil {
		return err
	}

	// Update session
	if err := sessionManager.SetCurrentPreset(fullPresetPath); err != nil {
		display.Warning(fmt.Sprintf(display.WarnSessionSaveFailed, err))
	}

	display.Success(fmt.Sprintf("Switched to variant: %s", variantName))
	return nil
}

// ListVariants returns all variant names for a preset
func ListVariants(preset string) ([]string, error) {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return nil, err
	}

	variantsDir := filepath.Join(presetPath, "variants")
	if _, err := os.Stat(variantsDir); os.IsNotExist(err) {
		// No variants exist
		return []string{}, nil
	}

	entries, err := os.ReadDir(variantsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read variants directory: %v", err)
	}

	var variants []string
	for _, entry := range entries {
		if entry.IsDir() {
			variants = append(variants, entry.Name())
		}
	}

	return variants, nil
}

// DeleteVariant removes a specific variant directory
func DeleteVariant(preset, variant string) error {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return err
	}

	variantsDir := filepath.Join(presetPath, "variants")
	variantPath := filepath.Join(variantsDir, variant)

	// Check if variant exists
	if _, err := os.Stat(variantPath); os.IsNotExist(err) {
		return fmt.Errorf("variant '%s' does not exist in preset '%s'", variant, preset)
	}

	// Remove the variant directory
	if err := os.RemoveAll(variantPath); err != nil {
		return fmt.Errorf("failed to delete variant '%s': %v", variant, err)
	}

	return nil
}

// GetVariantPath returns the full file path for a preset, considering variants
// For variant paths like "myapi/submit", returns the variant directory path
// For regular presets, returns the preset root directory
func GetVariantPath(preset, fileType string) (string, error) {
	// Extract base preset if variant path provided
	basePreset := preset
	if strings.Contains(preset, "/") {
		basePreset = strings.Split(preset, "/")[0]
	}

	presetPath, err := GetPresetPath(basePreset)
	if err != nil {
		return "", err
	}

	variantsDir := filepath.Join(presetPath, "variants")

	// Check if variants folder exists
	if _, err := os.Stat(variantsDir); err == nil {
		activeVariant := GetActiveVariant(basePreset)
		variantPath := filepath.Join(variantsDir, activeVariant)

		// Ensure variant directory exists
		if err := os.MkdirAll(variantPath, internal.DirPermissions); err != nil {
			return "", fmt.Errorf(display.ErrDirectoryFailed)
		}

		return filepath.Join(variantPath, fileType+".toml"), nil
	}

	// Fallback to root files (backward compatible)
	return filepath.Join(presetPath, fileType+".toml"), nil
}
