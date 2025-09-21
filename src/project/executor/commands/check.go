package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
)


// Check displays TOML file contents in a clean, readable format
func Check(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(errors.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(errors.ErrTargetRequired)
	}

	// Normalize target aliases
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf(errors.ErrInvalidTarget, cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf(errors.ErrFileLoadFailed, cmd.Target+".toml")
	}

	// Special handling for request fields (single values)
	if cmd.Target == "request" && len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf(errors.ErrKeyNotFound, key, cmd.Target)
		}
		
		// Always print raw value (Unix philosophy)
		fmt.Println(value)
		return nil
	}

	// Get specific key if provided
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf(errors.ErrKeyNotFound, key, cmd.Target)
		}

		// Always print raw value (Unix philosophy)
		switch v := value.(type) {
		case []interface{}:
			// For arrays, print as space-separated values
			for i, item := range v {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(item)
			}
			fmt.Println() // Add newline after array
		default:
			// Simple value
			fmt.Println(value)
		}
		return nil
	}

	// Display entire file contents
	return displayTOMLFile(handler, cmd.Target, cmd.Preset, cmd.RawOutput)
}

// displayTOMLFile shows the entire TOML file in a clean format
func displayTOMLFile(handler *toml.TomlHandler, target string, preset string, rawOutput bool) error {
	// Get the file path and read raw contents
	presetPath, err := presets.GetPresetPath(preset)
	if err != nil {
		// Silent failure - no file exists (Unix philosophy)
		return nil
	}

	filePath := filepath.Join(presetPath, target+".toml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Silent failure - no file content (Unix philosophy)
		return nil
	}

	// Always display raw file contents (Unix philosophy - like cat)
	fmt.Print(string(content))
	
	return nil
}

// Helper functions removed - using raw display only (Unix philosophy)