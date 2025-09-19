package commands

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// Set handles set operations for TOML files
func Set(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required for set command")
	}
	if cmd.Target == "" {
		return fmt.Errorf("target required (body, headers, query, config)")
	}
	if cmd.Key == "" {
		return fmt.Errorf("key required for set operation")
	}
	if cmd.Value == "" {
		return fmt.Errorf("value required for set operation")
	}

	// Normalize target aliases for better UX
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf("invalid target '%s'. Use: body, headers/header, query, request, variables", cmd.Target)
	}

	// Use normalized target for file operations
	cmd.Target = normalizedTarget

	// Special validation for request fields
	if cmd.Target == "request" {
		if err := executor.ValidateRequestField(cmd.Key, cmd.Value); err != nil {
			return err
		}
	}

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf("failed to load %s.toml: %v", cmd.Target, err)
	}

	// Detect if value is a variable
	isVar, varType, varName := executor.DetectVariableType(cmd.Value)
	if isVar {
		// Store variable info in config.toml for later resolution
		err := executor.StoreVariableInfo(cmd.Preset, cmd.Key, varType, varName)
		if err != nil {
			return fmt.Errorf("failed to store variable info: %v", err)
		}

		// Set the raw variable in the target file for now
		handler.Set(cmd.Key, cmd.Value)
	} else {
		// Infer type and set value, with special handling for request fields
		valueToStore := cmd.Value
		if cmd.Target == "request" && strings.ToLower(cmd.Key) == "method" {
			// Store HTTP methods in uppercase
			valueToStore = strings.ToUpper(cmd.Value)
		}
		inferredValue := executor.InferValueType(valueToStore)
		handler.Set(cmd.Key, inferredValue)
	}

	// Save the updated TOML file
	err = presets.SavePresetFile(cmd.Preset, cmd.Target, handler)
	if err != nil {
		return fmt.Errorf("failed to save %s.toml: %v", cmd.Target, err)
	}

	// Silent success - Unix philosophy
	return nil
}