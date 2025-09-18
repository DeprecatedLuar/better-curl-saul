package executor

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"main/src/project/presets"
	"main/src/project/toml"
)

// Variable represents a detected variable in TOML values
type Variable struct {
	Name    string // Variable name without prefix
	Type    string // "soft" for ?, "hard" for @
	Current string // Current value for hard variables
}

// VariableInfo holds information about a detected variable
type VariableInfo struct {
	Key  string // TOML key path where variable was found
	Type string // "soft" or "hard"
	Name string // Custom variable name (empty if bare @ or ?)
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

// PromptForVariables prompts user for variable values and returns substitution map
func PromptForVariables(preset string, persist bool) (map[string]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	substitutions := make(map[string]string)

	// Load variables.toml to get hard variables
	variablesHandler, err := presets.LoadPresetFile(preset, "variables")
	if err != nil {
		return nil, fmt.Errorf("failed to load variables: %v", err)
	}

	// Find all variables across all TOML files
	variables, err := findAllVariables(preset)
	if err != nil {
		return nil, fmt.Errorf("failed to find variables: %v", err)
	}

	for _, variable := range variables {
		var prompt string
		var currentValue string

		if variable.Type == "soft" {
			// Soft variables: always prompt with empty input
			if variable.Name != "" {
				prompt = variable.Name + ": "
			} else {
				prompt = variable.Key + ": "
			}
		} else if variable.Type == "hard" {
			// Hard variables: show current value, only prompt if --persist
			if !persist {
				// Use existing value without prompting
				currentValue = variablesHandler.GetAsString(variable.Key)
				if currentValue != "" {
					substitutions[variable.Key] = currentValue
				}
				continue
			}

			// Prompting for hard variable with current value
			currentValue = variablesHandler.GetAsString(variable.Key)
			if variable.Name != "" {
				prompt = variable.Name + ": " + currentValue + "_"
			} else {
				prompt = variable.Key + ": " + currentValue + "_"
			}
		}

		fmt.Print(prompt)
		if scanner.Scan() {
			userInput := strings.TrimSpace(scanner.Text())

			if variable.Type == "hard" && userInput == "" && currentValue != "" {
				// Keep existing value for hard variables if user presses Enter
				substitutions[variable.Key] = currentValue
			} else if userInput != "" {
				substitutions[variable.Key] = userInput

				// Save hard variables to variables.toml
				if variable.Type == "hard" {
					variablesHandler.Set(variable.Key, userInput)
					err := presets.SavePresetFile(preset, "variables", variablesHandler)
					if err != nil {
						return nil, fmt.Errorf("failed to save variable: %v", err)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	return substitutions, nil
}

// findAllVariables scans all TOML files in preset to find variables
func findAllVariables(preset string) ([]VariableInfo, error) {
	var variables []VariableInfo
	targets := []string{"body", "headers", "query", "request"}

	for _, target := range targets {
		handler, err := presets.LoadPresetFile(preset, target)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		// Scan all keys in the TOML file
		targetVars := scanHandlerForVariables(handler, "")
		variables = append(variables, targetVars...)
	}

	return variables, nil
}

// scanHandlerForVariables recursively scans a TOML handler for variable values
func scanHandlerForVariables(handler *toml.TomlHandler, prefix string) []VariableInfo {
	var variables []VariableInfo

	for _, key := range handler.Keys() {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		value := handler.Get(key)
		if value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			// Check if this string is a variable
			if isVar, varType, varName := DetectVariableType(v); isVar {
				variables = append(variables, VariableInfo{
					Key:  fullKey,
					Type: varType,
					Name: varName,
				})
			}
		default:
			// For nested objects, we'd need recursive scanning
			// For now, we'll handle flat structures
		}
	}

	return variables
}

// SubstituteVariables replaces variables in TOML handler with actual values
func SubstituteVariables(handler *toml.TomlHandler, substitutions map[string]string) error {
	for _, key := range handler.Keys() {
		value := handler.Get(key)
		if value == nil {
			continue
		}

		if strValue, ok := value.(string); ok {
			if isVar, _, _ := DetectVariableType(strValue); isVar {
				if substitute, exists := substitutions[key]; exists {
					// Infer type for the substituted value
					typedValue := InferValueType(substitute)
					handler.Set(key, typedValue)
				}
			}
		}
	}

	return nil
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

	// Store hard variable with empty initial value (will be set during call command)
	// Simple flat structure: "path.to.field" = "current_value"
	// Note: varName can be empty (bare @) - that's fine, we store by field path
	handler.Set(key, "") // Empty value initially

	// Save variables.toml
	return presets.SavePresetFile(preset, "variables", handler)
}