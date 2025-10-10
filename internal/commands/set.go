// Package commands provides individual command implementations for Better-Curl-Saul.
// This package contains the specific logic for set, get, check, edit, call, and history commands,
// implementing the core functionality of the HTTP client workspace system.
package commands

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/DeprecatedLuar/better-curl-saul/internal/core"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/internal/variables"
)

// Set handles set operations for TOML files
func Set(cmd core.Command) error {
	// Handle --raw flag for curl import via editor
	if cmd.RawOutput {
		if cmd.Preset == "" {
			return fmt.Errorf(display.ErrPresetNameRequired)
		}
		return workspace.ImportCurlViaEditor(cmd.Preset)
	}

	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(display.ErrTargetRequired)
	}
	if len(cmd.KeyValuePairs) == 0 {
		return fmt.Errorf(display.ErrKeyValueRequired)
	}

	// Normalize target aliases for better UX
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf(display.ErrInvalidTarget, cmd.Target)
	}

	// Use normalized target for file operations
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := workspace.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf(display.ErrFileLoadFailed, cmd.Target+".toml")
	}

	// Special handling for filters target - store values as array
	if cmd.Target == "filters" {
		var fields []string
		for _, kvp := range cmd.KeyValuePairs {
			fields = append(fields, kvp.Value)
		}
		handler.Set("fields", fields)
	} else {
		// Process all key-value pairs for other targets
		for _, kvp := range cmd.KeyValuePairs {
			// Special validation for request fields
			if cmd.Target == "request" {
				if err := ValidateRequestField(kvp.Key, kvp.Value); err != nil {
					return err
				}
			}

			// Detect if value is a variable
			isVar, varType, varName := variables.DetectVariableType(kvp.Value)
			if isVar {
				// Store variable info in config.toml for later resolution
				err := variables.StoreVariableInfo(cmd.Preset, kvp.Key, varType, varName)
				if err != nil {
					return fmt.Errorf(display.ErrVariableSaveFailed)
				}

				// Set the raw variable in the target file for now
				handler.Set(kvp.Key, kvp.Value)
			} else {
				// Infer type and set value, with special handling for request fields
				valueToStore := kvp.Value
				keyToStore := kvp.Key

				if cmd.Target == "request" {
					if strings.ToLower(kvp.Key) == "method" {
						// Store HTTP methods in uppercase
						valueToStore = strings.ToUpper(kvp.Value)
					} else if strings.ToLower(kvp.Key) == "history" {
						// Map "history" to "history_count" for storage
						keyToStore = "history_count"
					}
				}

				inferredValue := InferValueType(valueToStore)
				handler.Set(keyToStore, inferredValue)
			}
		}
	}

	// Save the updated TOML file (once after all operations)
	err = workspace.SavePresetFile(cmd.Preset, cmd.Target, handler)
	if err != nil {
		return fmt.Errorf(display.ErrFileSaveFailed, cmd.Target+".toml")
	}

	// Silent success - Unix philosophy
	return nil
}