package executor

import (
	"fmt"
	"strconv"
	"strings"

	"main/src/project/parser"
	"main/src/project/presets"
	"main/src/project/toml"
)

// Variable represents a detected variable in TOML values
type Variable struct {
	Name    string // Variable name without prefix
	Type    string // "soft" for ?, "hard" for $
	Current string // Current value for hard variables
}

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

// DetectVariableType checks if a value is a variable and returns its type
func DetectVariableType(value string) (isVariable bool, varType string, varName string) {
	if len(value) < 1 {
		return false, "", ""
	}

	switch value[0] {
	case '?':
		if len(value) == 1 {
			// Bare ? - no custom name, will use field path
			return true, "soft", ""
		}
		// ?customname - has custom name
		return true, "soft", value[1:]
	case '@':
		if len(value) == 1 {
			// Bare @ - no custom name, will use field path
			return true, "hard", ""
		}
		// @customname - has custom name
		return true, "hard", value[1:]
	default:
		return false, "", ""
	}
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

// storeVariableInfo stores hard variables in variables.toml (only hard variables, no soft variables)
func storeVariableInfo(preset, key, varType, varName string) error {
	// Only store hard variables - soft variables are never stored
	if varType != "hard" {
		return nil // Don't store soft variables
	}

	// Load variables.toml
	handler, err := presets.LoadPresetFile(preset, "variables")
	if err != nil {
		return err
	}

	// Store hard variable with empty initial value (will be set during fire command)
	// Simple flat structure: "path.to.field" = "current_value"
	// Note: varName can be empty (bare $) - that's fine, we store by field path
	handler.Set(key, "") // Empty value initially

	// Save variables.toml
	return presets.SavePresetFile(preset, "variables", handler)
}

// validateRequestField validates special request field values
func validateRequestField(key, value string) error {
	switch strings.ToLower(key) {
	case "method":
		return validateHTTPMethod(value)
	case "url":
		return validateURL(value)
	case "timeout":
		return validateTimeout(value)
	default:
		return nil
	}
}

// validateHTTPMethod checks if the HTTP method is valid
func validateHTTPMethod(method string) error {
	validMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH",
		"HEAD", "OPTIONS", "TRACE", "CONNECT",
	}

	methodUpper := strings.ToUpper(method)
	for _, valid := range validMethods {
		if methodUpper == valid {
			return nil
		}
	}

	return fmt.Errorf("sorry champ \"%s\" isn't really a thing, but i'll let you try again", method)
}

// validateURL performs basic URL validation
func validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("listen pal, at least put in the URL. Come on")
	}
	// Basic check - should start with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("alright, so the \"U R L\" needs to start with one of these two here: 'http://' or 'https://'. Go get'em tiger")
	}
	return nil
}

// validateTimeout validates timeout value
func validateTimeout(timeout string) error {
	if _, err := strconv.Atoi(timeout); err != nil {
		return fmt.Errorf("timeout must be a number (seconds)")
	}
	return nil
}
