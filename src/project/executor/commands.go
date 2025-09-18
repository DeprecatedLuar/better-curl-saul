package executor

import (
	"fmt"
	"strconv"
	"strings"

	"main/src/project/parser"
	"main/src/project/presets"
	"main/src/project/toml"
)


// ExecuteSetCommand handles set operations for TOML files
func ExecuteSetCommand(cmd parser.Command) error {
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
	normalizedTarget := normalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf("invalid target '%s'. Use: body, headers/header, query, request, variables", cmd.Target)
	}

	// Use normalized target for file operations
	cmd.Target = normalizedTarget

	// Special validation for request fields
	if cmd.Target == "request" {
		if err := validateRequestField(cmd.Key, cmd.Value); err != nil {
			return err
		}
	}

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf("failed to load %s.toml: %v", cmd.Target, err)
	}

	// Detect if value is a variable
	isVar, varType, varName := DetectVariableType(cmd.Value)
	if isVar {
		// Store variable info in config.toml for later resolution
		err := storeVariableInfo(cmd.Preset, cmd.Key, varType, varName)
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
		inferredValue := InferValueType(valueToStore)
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

// ExecuteGetCommand retrieves values from TOML files for debugging
func ExecuteGetCommand(cmd parser.Command) (interface{}, error) {
	if cmd.Preset == "" {
		return nil, fmt.Errorf("preset name required for get command")
	}
	if cmd.Target == "" {
		return nil, fmt.Errorf("target required (body, headers, query, request, variables)")
	}

	// Normalize target aliases
	normalizedTarget := normalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return nil, fmt.Errorf("invalid target '%s'. Use: body, headers/header, query, request, variables", cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s.toml: %v", cmd.Target, err)
	}

	if cmd.Key == "" {
		// Return entire TOML structure as a simple message
		return "TOML structure display not implemented yet", nil
	}

	// Get specific key
	value := handler.Get(cmd.Key)
	if value == nil {
		return nil, fmt.Errorf("key '%s' not found", cmd.Key)
	}

	return value, nil
}

// ExecuteCheckCommand displays TOML file contents in a clean, readable format
func ExecuteCheckCommand(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required for check command")
	}
	if cmd.Target == "" {
		return fmt.Errorf("target required (body, headers, query, request, variables)")
	}

	// Normalize target aliases
	normalizedTarget := normalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf("invalid target '%s'. Use: body, headers/header, query, request, variables", cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf("failed to load %s.toml: %v", cmd.Target, err)
	}

	// Special handling for request fields (single values)
	if cmd.Target == "request" && cmd.Key != "" {
		value := handler.Get(cmd.Key)
		if value == nil {
			return fmt.Errorf("'%s' not set in request", cmd.Key)
		}
		fmt.Println(value)
		return nil
	}

	// Get specific key if provided
	if cmd.Key != "" {
		value := handler.Get(cmd.Key)
		if value == nil {
			return fmt.Errorf("key '%s' not found in %s", cmd.Key, cmd.Target)
		}

		// Format based on type
		switch v := value.(type) {
		case []interface{}:
			// Array format
			fmt.Printf("%s = [", cmd.Key)
			for i, item := range v {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("\"%v\"", item)
			}
			fmt.Println("]")
		default:
			// Simple value
			fmt.Printf("%s = \"%v\"\n", cmd.Key, value)
		}
		return nil
	}

	// Display entire file contents
	return displayTOMLFile(handler, cmd.Target)
}

// displayTOMLFile shows the entire TOML file in a clean format
func displayTOMLFile(handler *toml.TomlHandler, target string) error {
	// For now, use a simple approach - we'll enhance this as needed
	// This is a basic implementation that can be improved

	fmt.Printf("# %s.toml contents:\n", target)

	// Get the raw TOML content and display it
	// Note: This is a simplified version - we might need to enhance formatting later
	fmt.Println("(Full file display not yet implemented - use specific keys for now)")

	return nil
}


// InferValueType converts string values to appropriate Go types for TOML
func InferValueType(value string) interface{} {
	// For now, keep everything as strings to avoid TOML handler issues
	// TODO: Implement proper type handling once TOML handler supports all types

	// Try to parse as boolean (keep this as it's simple)
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Check for array notation (comma-separated values)
	if strings.Contains(value, ",") {
		parts := strings.Split(value, ",")
		var result []string
		for _, part := range parts {
			result = append(result, strings.TrimSpace(part))
		}
		return result
	}

	// Default to string (including numbers for now)
	return value
}

// normalizeTarget converts target aliases to canonical names
func normalizeTarget(target string) string {
	switch strings.ToLower(target) {
	case "body":
		return "body"
	case "headers", "header":
		return "headers"
	case "query", "queries":
		return "query"
	case "request", "req", "url":
		return "request"
	case "variables", "vars", "var":
		return "variables"
	default:
		return ""
	}
}

