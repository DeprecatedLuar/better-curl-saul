package variables

import (
	"regexp"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/utils"
)

// SubstituteVariables replaces variables in TOML handler with actual values using simple regex
func SubstituteVariables(handler *toml.TomlHandler, substitutions map[string]string) error {
	for _, key := range handler.Keys() {
		value := handler.Get(key)
		if value == nil {
			continue
		}

		if strValue, ok := value.(string); ok {
			// Use simple regex to replace all variables in the string
			newValue := substituteVariablesInText(strValue, substitutions)
			if newValue != strValue {
				// String was modified, update it
				typedValue := utils.InferValueType(newValue)
				handler.Set(key, typedValue)
			}
		}
	}

	return nil
}

// substituteVariablesInText replaces all variables in text using regex
// The function doesn't need to know which file the variable came from
// because substitutions map already contains the full key (e.g., "body.pokename")
func substituteVariablesInText(text string, substitutions map[string]string) string {
	// Regex to find all {@ } and {?} patterns
	regex := regexp.MustCompile(`\{([@?])(\w*)\}`)

	return regex.ReplaceAllStringFunc(text, func(match string) string {
		// Parse the variable from the match
		submatches := regex.FindStringSubmatch(match)
		varName := submatches[2] // variable name (can be empty)

		// Try to find this variable in our substitutions map
		// The substitutions map contains keys like "body.pokename", "headers.auth", etc.
		for key, substitute := range substitutions {
			// Extract the variable name from the key (e.g., "pokename" from "body.pokename")
			parts := strings.Split(key, ".")
			if len(parts) >= 2 {
				keyVarName := parts[1]
				if (varName != "" && keyVarName == varName) || (varName == "" && keyVarName == "variable") {
					return substitute
				}
			}
		}

		// If no substitution found, return original match unchanged
		return match
	})
}

// StoreVariableInfo stores hard variables in variables.toml (only hard variables, no soft variables)
func StoreVariableInfo(preset, key, varType, varName string) error {
	// Only store hard variables - soft variables are never stored
	if varType != "hard" {
		return nil // Don't store soft variables
	}

	// Load variables.toml
	handler, err := presets.LoadPresetFile(preset, "variables")
	if err != nil {
		return err
	}

	// Store hard variable with empty initial value (will be set during call command)
	// Simple flat structure: "path.to.field" = "current_value"
	// Note: varName can be empty (bare @) - that's fine, we store by field path
	handler.Set(key, "") // Empty value initially

	// Save variables.toml
	return presets.SavePresetFile(preset, "variables", handler)
}

