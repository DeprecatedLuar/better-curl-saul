package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Remove handles both preset deletion and target file deletion
func Remove(cmd parser.Command) error {
	// Global mode: Remove entire presets
	if cmd.Global == "rm" {
		return removePresets(cmd.Targets)
	}

	// Preset mode: Remove target files from preset
	if cmd.Command == "rm" {
		return removeTargets(cmd.Preset, cmd.Targets)
	}

	return fmt.Errorf(display.ErrInvalidRemoveCommand)
}

// removePresets deletes entire preset directories
func removePresets(presetNames []string) error {
	if len(presetNames) == 0 {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	deletedCount := 0
	for _, presetName := range presetNames {
		if err := workspace.DeletePreset(presetName); err != nil {
			display.Warning(fmt.Sprintf(display.ErrPresetNotFound, presetName))
		} else {
			deletedCount++
		}
	}

	return nil
}

// removeTargets deletes specific target TOML files from a preset
// Supports variants: uses GetVariantPath for variant-aware file resolution
func removeTargets(presetName string, targets []string) error {
	if len(targets) == 0 {
		return fmt.Errorf(display.ErrRemoveTargetRequired)
	}

	// Extract base preset for existence check
	basePreset := presetName
	if filepath.Base(presetName) != presetName {
		basePreset = filepath.Dir(presetName)
	}

	if !workspace.PresetExists(basePreset) {
		return fmt.Errorf(display.ErrPresetNotFound, basePreset)
	}

	validTargets := []string{"body", "headers", "query", "request", "variables"}
	validMap := make(map[string]bool)
	for _, t := range validTargets {
		validMap[t] = true
	}

	deletedCount := 0
	for _, target := range targets {
		if !validMap[target] {
			display.Warning(fmt.Sprintf("Invalid target: %s (valid: body, headers, query, request, variables)", target))
			continue
		}

		// Use GetVariantPath for variant-aware file resolution
		filePath, err := workspace.GetVariantPath(presetName, target)
		if err != nil {
			display.Warning(fmt.Sprintf("Failed to resolve path for %s: %v", target, err))
			continue
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			display.Warning(fmt.Sprintf("%s.toml does not exist in preset '%s'", target, presetName))
			continue
		}

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			display.Warning(fmt.Sprintf("Failed to remove %s: %v", target, err))
			continue
		}

		deletedCount++
		display.Success(fmt.Sprintf("Removed %s from preset '%s'", target, presetName))
	}

	if deletedCount == 0 {
		return fmt.Errorf(display.ErrNoTargetsRemoved)
	}

	return nil
}
