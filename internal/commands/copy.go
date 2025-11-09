package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Copy handles the copy command - copies presets or variants
// Supports: saul cp source dest, saul copy source dest
func Copy(cmd parser.Command) error {
	if len(cmd.Targets) < 2 {
		return fmt.Errorf("copy requires source and destination arguments")
	}

	source := cmd.Targets[0]
	dest := cmd.Targets[1]

	// Determine if source is a variant path
	sourceIsVariant := strings.Contains(source, "/")
	destIsVariant := strings.Contains(dest, "/")

	// Handle variant to variant copy
	if sourceIsVariant && destIsVariant {
		return copyVariantToVariant(source, dest)
	}

	// Handle preset to variant copy
	if !sourceIsVariant && destIsVariant {
		return copyPresetToVariant(source, dest)
	}

	// Handle variant to preset copy
	if sourceIsVariant && !destIsVariant {
		return copyVariantToPreset(source, dest)
	}

	// Handle preset to preset copy
	return copyPresetToPreset(source, dest)
}

// copyPresetToPreset copies an entire preset directory
func copyPresetToPreset(source, dest string) error {
	// Check source exists
	if !workspace.PresetExists(source) {
		return fmt.Errorf(display.ErrPresetNotFound, source)
	}

	// Check destination doesn't exist
	if workspace.PresetExists(dest) {
		return fmt.Errorf("destination preset '%s' already exists", dest)
	}

	sourcePath, err := workspace.GetPresetPath(source)
	if err != nil {
		return err
	}

	destPath, err := workspace.GetPresetPath(dest)
	if err != nil {
		return err
	}

	// Copy entire directory recursively
	if err := copyDir(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy preset: %v", err)
	}

	return nil
}

// copyVariantToVariant copies one variant to another
func copyVariantToVariant(source, dest string) error {
	sourceParts := strings.Split(source, "/")
	destParts := strings.Split(dest, "/")

	if len(sourceParts) != 2 || len(destParts) != 2 {
		return fmt.Errorf("invalid variant path format")
	}

	sourceBase := sourceParts[0]
	sourceVariant := sourceParts[1]
	destBase := destParts[0]
	destVariant := destParts[1]

	// Check source base preset exists
	if !workspace.PresetExists(sourceBase) {
		return fmt.Errorf(display.ErrPresetNotFound, sourceBase)
	}

	// Check destination base preset exists
	if !workspace.PresetExists(destBase) {
		return fmt.Errorf(display.ErrPresetNotFound, destBase)
	}

	// Get source variant path
	sourcePresetPath, err := workspace.GetPresetPath(sourceBase)
	if err != nil {
		return err
	}
	sourceVariantPath := filepath.Join(sourcePresetPath, "variants", sourceVariant)

	// Check source variant exists
	if _, err := os.Stat(sourceVariantPath); os.IsNotExist(err) {
		return fmt.Errorf("source variant '%s' does not exist", source)
	}

	// Ensure destination variant structure
	if err := workspace.EnsureVariantStructure(dest); err != nil {
		return err
	}

	// Get destination variant path
	destPresetPath, err := workspace.GetPresetPath(destBase)
	if err != nil {
		return err
	}
	destVariantPath := filepath.Join(destPresetPath, "variants", destVariant)

	// Copy variant directory
	if err := copyDir(sourceVariantPath, destVariantPath); err != nil {
		return fmt.Errorf("failed to copy variant: %v", err)
	}

	return nil
}

// copyPresetToVariant copies a preset to a variant
func copyPresetToVariant(source, dest string) error {
	destParts := strings.Split(dest, "/")
	if len(destParts) != 2 {
		return fmt.Errorf("invalid variant path format")
	}

	destBase := destParts[0]
	destVariant := destParts[1]

	// Check source preset exists
	if !workspace.PresetExists(source) {
		return fmt.Errorf(display.ErrPresetNotFound, source)
	}

	// Check destination base preset exists, create if not
	if !workspace.PresetExists(destBase) {
		if err := workspace.CreatePresetDirectory(destBase); err != nil {
			return err
		}
	}

	// Ensure destination variant structure
	if err := workspace.EnsureVariantStructure(dest); err != nil {
		return err
	}

	// Get source preset path (use root files)
	sourcePath, err := workspace.GetPresetPath(source)
	if err != nil {
		return err
	}

	// Get destination variant path
	destPresetPath, err := workspace.GetPresetPath(destBase)
	if err != nil {
		return err
	}
	destVariantPath := filepath.Join(destPresetPath, "variants", destVariant)

	// Copy TOML files from source to destination variant
	tomlFiles := []string{"request.toml", "body.toml", "headers.toml", "query.toml", "variables.toml", "filters.toml"}
	for _, file := range tomlFiles {
		sourceFile := filepath.Join(sourcePath, file)
		destFile := filepath.Join(destVariantPath, file)

		if _, err := os.Stat(sourceFile); err == nil {
			if err := copyFile(sourceFile, destFile); err != nil {
				return fmt.Errorf("failed to copy %s: %v", file, err)
			}
		}
	}

	return nil
}

// copyVariantToPreset copies a variant to a preset
func copyVariantToPreset(source, dest string) error {
	sourceParts := strings.Split(source, "/")
	if len(sourceParts) != 2 {
		return fmt.Errorf("invalid variant path format")
	}

	sourceBase := sourceParts[0]
	sourceVariant := sourceParts[1]

	// Check source base preset exists
	if !workspace.PresetExists(sourceBase) {
		return fmt.Errorf(display.ErrPresetNotFound, sourceBase)
	}

	// Check destination doesn't exist
	if workspace.PresetExists(dest) {
		return fmt.Errorf("destination preset '%s' already exists", dest)
	}

	// Get source variant path
	sourcePresetPath, err := workspace.GetPresetPath(sourceBase)
	if err != nil {
		return err
	}
	sourceVariantPath := filepath.Join(sourcePresetPath, "variants", sourceVariant)

	// Check source variant exists
	if _, err := os.Stat(sourceVariantPath); os.IsNotExist(err) {
		return fmt.Errorf("source variant '%s' does not exist", source)
	}

	// Create destination preset
	if err := workspace.CreatePresetDirectory(dest); err != nil {
		return err
	}

	// Get destination preset path
	destPath, err := workspace.GetPresetPath(dest)
	if err != nil {
		return err
	}

	// Copy TOML files from variant to preset root
	tomlFiles := []string{"request.toml", "body.toml", "headers.toml", "query.toml", "variables.toml", "filters.toml"}
	for _, file := range tomlFiles {
		sourceFile := filepath.Join(sourceVariantPath, file)
		destFile := filepath.Join(destPath, file)

		if _, err := os.Stat(sourceFile); err == nil {
			if err := copyFile(sourceFile, destFile); err != nil {
				return fmt.Errorf("failed to copy %s: %v", file, err)
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// copyDir recursively copies a directory tree
func copyDir(src, dst string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, sourceInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(sourcePath, destPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(sourcePath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}
