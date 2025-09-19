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

// DetectVariableType checks if a value is a variable and returns its type
// NEW: Detects braced variables {@name} and {?name} to avoid URL conflicts
func DetectVariableType(value string) (isVariable bool, varType string, varName string) {
	if len(value) < 3 { // Minimum: {?} or {@}
		return false, "", ""
	}

	// Check for hard variable: {@name} or bare {@}
	hardRegex := regexp.MustCompile(`^\{@(\w*)\}$`)
	if matches := hardRegex.FindStringSubmatch(value); matches != nil {
		return true, "hard", matches[1] // matches[1] is the captured name (empty if bare {@})
	}

	// Check for soft variable: {?name} or bare {?}
	softRegex := regexp.MustCompile(`^\{\?(\w*)\}$`)
	if matches := softRegex.FindStringSubmatch(value); matches != nil {
		return true, "soft", matches[1] // matches[1] is the captured name (empty if bare {?})
	}

	return false, "", ""
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
			// Hard variables: use stored value if exists, otherwise prompt
			currentValue = variablesHandler.GetAsString(variable.Key)
			if !persist && currentValue != "" {
				// Use existing value without prompting (only if value exists)
				substitutions[variable.Key] = currentValue
				continue
			}

			// Prompting for hard variable with current value
			currentValue = variablesHandler.GetAsString(variable.Key)
			if variable.Name != "" {
				prompt = variable.Name + " [" + currentValue + "]: "
			} else {
				prompt = variable.Key + " [" + currentValue + "]: "
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
// NEW: Handles partial variables within strings and nested structures
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
			// Check for both full string variables and partial variables within strings
			if isVar, varType, varName := DetectVariableType(v); isVar {
				// Full string is a variable: {@username}
				variables = append(variables, VariableInfo{
					Key:  fullKey,
					Type: varType,
					Name: varName,
				})
			} else {
				// Check for partial variables within the string: https://api.com/{@user}/posts
				variables = append(variables, extractPartialVariables(v, fullKey)...)
			}
		default:
			// For now, handle flat structures only
			// TODO: Add recursive scanning for nested objects if needed
		}
	}

	return variables
}

// extractPartialVariables finds variables within a string using regex
func extractPartialVariables(str, keyPath string) []VariableInfo {
	var variables []VariableInfo

	// Find all hard variables: {@name}
	hardRegex := regexp.MustCompile(`\{@(\w*)\}`)
	hardMatches := hardRegex.FindAllStringSubmatch(str, -1)
	for i, match := range hardMatches {
		varName := match[1] // Captured variable name (empty if bare {@})
		// Create unique key for each occurrence: keyPath.varName.occurrence
		varKey := keyPath
		if varName != "" {
			varKey += "." + varName
		} else {
			// For bare {@}, use the keyPath with occurrence number
			varKey += fmt.Sprintf(".var%d", i)
		}
		variables = append(variables, VariableInfo{
			Key:  varKey,
			Type: "hard",
			Name: varName,
		})
	}

	// Find all soft variables: {?name}
	softRegex := regexp.MustCompile(`\{\?(\w*)\}`)
	softMatches := softRegex.FindAllStringSubmatch(str, -1)
	for i, match := range softMatches {
		varName := match[1] // Captured variable name (empty if bare {?})
		// Create unique key for each occurrence: keyPath.varName.occurrence
		varKey := keyPath
		if varName != "" {
			varKey += "." + varName
		} else {
			// For bare {?}, use the keyPath with occurrence number
			varKey += fmt.Sprintf(".var%d", i)
		}
		variables = append(variables, VariableInfo{
			Key:  varKey,
			Type: "soft",
			Name: varName,
		})
	}

	return variables
}

// SubstituteVariables replaces variables in TOML handler with actual values
// NEW: Handles both full string variables and partial substitution within strings
func SubstituteVariables(handler *toml.TomlHandler, substitutions map[string]string) error {
	for _, key := range handler.Keys() {
		value := handler.Get(key)
		if value == nil {
			continue
		}

		if strValue, ok := value.(string); ok {
			if isVar, _, varName := DetectVariableType(strValue); isVar {
				// Full string is a variable - replace entire value
				// Construct proper variable key for lookup (same as extractPartialVariables logic)
				varKey := key
				if varName != "" {
					varKey += "." + varName
				} else {
					// For bare {@} or {?}, use the keyPath
					varKey += ".var0"
				}

				if substitute, exists := substitutions[varKey]; exists {
					// Infer type for the substituted value
					typedValue := InferValueType(substitute)
					handler.Set(key, typedValue)
				}
			} else {
				// Check for partial variables within the string and replace them
				newValue := substitutePartialVariables(strValue, substitutions, key)
				if newValue != strValue {
					// String was modified, update it
					typedValue := InferValueType(newValue)
					handler.Set(key, typedValue)
				}
			}
		}
	}

	return nil
}

// substitutePartialVariables replaces variables within a string
func substitutePartialVariables(str string, substitutions map[string]string, keyPath string) string {
	result := str

	// Replace all hard variables: {@name}
	hardRegex := regexp.MustCompile(`\{@(\w*)\}`)
	hardOccurrence := 0
	result = hardRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable name from {@name}
		varName := hardRegex.FindStringSubmatch(match)[1]
		
		// Create key to match what we stored during detection
		varKey := keyPath
		if varName != "" {
			varKey += "." + varName
		} else {
			varKey += fmt.Sprintf(".var%d", hardOccurrence)
		}
		hardOccurrence++
		
		if substitute, exists := substitutions[varKey]; exists {
			return substitute
		}
		return match // No substitution found, keep original
	})

	// Replace all soft variables: {?name}
	softRegex := regexp.MustCompile(`\{\?(\w*)\}`)
	softOccurrence := 0
	result = softRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable name from {?name}
		varName := softRegex.FindStringSubmatch(match)[1]
		
		// Create key to match what we stored during detection
		varKey := keyPath
		if varName != "" {
			varKey += "." + varName
		} else {
			varKey += fmt.Sprintf(".var%d", softOccurrence)
		}
		softOccurrence++
		
		if substitute, exists := substitutions[varKey]; exists {
			return substitute
		}
		return match // No substitution found, keep original
	})

	return result
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