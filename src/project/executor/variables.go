package executor

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

// DetectVariableType checks if a value is a variable and returns its type (legacy function for backwards compatibility)
func DetectVariableType(value string) (isVariable bool, varType string, varName string) {
	variables := FindVariablesInString(value)
	if len(variables) == 1 && variables[0].FullMatch == value {
		// Only return true if the entire string is a single variable (backwards compatibility)
		return true, variables[0].Type, variables[0].Name
	}
	return false, "", ""
}

// VariableMatch represents a variable found within a string
type VariableMatch struct {
	FullMatch string // The complete variable syntax: {?name} or {@name}
	Type      string // "soft" or "hard"
	Name      string // Variable name (empty for bare variables)
	Start     int    // Start position in the string
	End       int    // End position in the string
}

// FindVariablesInString finds all variables embedded in a string
func FindVariablesInString(value string) []VariableMatch {
	var variables []VariableMatch

	// Regex to match {?name} or {@name} patterns
	re := regexp.MustCompile(`\{([@?])([^}]*)\}`)
	matches := re.FindAllStringSubmatch(value, -1)
	indices := re.FindAllStringIndex(value, -1)

	for i, match := range matches {
		if len(match) >= 3 {
			prefix := match[1]
			name := match[2]

			var varType string
			switch prefix {
			case "?":
				varType = "soft"
			case "@":
				varType = "hard"
			default:
				continue
			}

			variables = append(variables, VariableMatch{
				FullMatch: match[0],
				Type:      varType,
				Name:      name,
				Start:     indices[i][0],
				End:       indices[i][1],
			})
		}
	}

	return variables
}

// HasVariables checks if a string contains any variables
func HasVariables(value string) bool {
	return len(FindVariablesInString(value)) > 0
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
	variables, err := FindAllVariables(preset)
	if err != nil {
		return nil, fmt.Errorf("failed to find variables: %v", err)
	}

	for _, variable := range variables {
		var prompt string
		var currentValue string

		// Create storage key and substitution key
		var storageKey string      // Key for storing in variables.toml
		var substitutionKey string // Key for variable substitution

		// Check if this is a request field variable (special handling)
		if variable.Key == "url" || variable.Key == "method" || variable.Key == "timeout" {
			// Request variables: use "request.field.varname" for storage
			if variable.Name != "" {
				storageKey = "request." + variable.Key + "." + variable.Name
				substitutionKey = variable.Name
			} else {
				storageKey = "request." + variable.Key
				substitutionKey = variable.Key
			}
		} else {
			// Body/Header/Query variables: use variable name for storage
			if variable.Name != "" {
				storageKey = variable.Name
				substitutionKey = variable.Name
			} else {
				storageKey = variable.Key
				substitutionKey = variable.Key
			}
		}

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
				currentValue = variablesHandler.GetAsString(storageKey)
				if currentValue != "" {
					substitutions[substitutionKey] = currentValue
				}
				continue
			}

			// Prompting for hard variable with current value
			currentValue = variablesHandler.GetAsString(storageKey)
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
				substitutions[substitutionKey] = currentValue
			} else if userInput != "" {
				substitutions[substitutionKey] = userInput

				// Save hard variables to variables.toml
				if variable.Type == "hard" {
					variablesHandler.Set(storageKey, userInput)
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

// FindAllVariables scans all TOML files in preset to find variables
func FindAllVariables(preset string) ([]VariableInfo, error) {
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
			// Find all variables in this string (supports embedded variables)
			varMatches := FindVariablesInString(v)
			for _, varMatch := range varMatches {
				variables = append(variables, VariableInfo{
					Key:  fullKey,
					Type: varMatch.Type,
					Name: varMatch.Name,
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
			// Check if this string has any variables
			if HasVariables(strValue) {
				// Replace all variables in the string
				newValue := strValue
				varMatches := FindVariablesInString(strValue)

				// Replace variables from right to left to preserve indices
				for i := len(varMatches) - 1; i >= 0; i-- {
					varMatch := varMatches[i]

					// Create substitution key: for URL variables, we use the field path + variable identifier
					var substitutionKey string
					if varMatch.Name != "" {
						substitutionKey = varMatch.Name
					} else {
						substitutionKey = key
					}

					if substitute, exists := substitutions[substitutionKey]; exists {
						// Replace this specific variable occurrence in the string
						newValue = newValue[:varMatch.Start] + substitute + newValue[varMatch.End:]
					}
				}

				// If we successfully replaced variables, update the value
				if newValue != strValue {
					// For strings with embedded variables, keep as string (don't infer type)
					handler.Set(key, newValue)
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