package workspace

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
	"github.com/DeprecatedLuar/better-curl-saul/internal/utils"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// ===== VARIABLE TYPES =====

// VariableInfo holds information about a detected variable
type VariableInfo struct {
	Key  string // TOML key path where variable was found
	Type string // "soft" or "hard"
	Name string // Custom variable name (empty if bare @ or ?)
}

// Variable represents a detected variable in TOML values
type Variable struct {
	Name    string // Variable name without prefix
	Type    string // "soft" for ?, "hard" for @
	Current string // Current value for hard variables
}

// ===== VARIABLE DETECTION =====

// DetectVariableType checks if a value is a variable and returns its type
// Detects braced variables {@name} and {?name} to avoid URL conflicts
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

// FindAllVariables scans all TOML files in preset to find variables using simple regex
func FindAllVariables(preset string) ([]VariableInfo, error) {
	var variables []VariableInfo
	targets := []string{"body", "headers", "query", "request"}

	// Get preset path
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return variables, err
	}

	for _, target := range targets {
		filePath := presetPath + "/" + target + ".toml"

		// Read file content as text
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		// Find all variables in this file using regex
		fileVars := findVariablesInText(string(content), target)
		variables = append(variables, fileVars...)
	}

	return variables, nil
}

// findVariablesInText uses regex to find all variables in text content
func findVariablesInText(content, fileContext string) []VariableInfo {
	var variables []VariableInfo

	// Regex to find {@} and {?} patterns
	regex := regexp.MustCompile(`\{([@?])(\w*)\}`)
	matches := regex.FindAllStringSubmatch(content, -1)

	// Track unique variables to avoid duplicates
	seen := make(map[string]bool)

	for _, match := range matches {
		varSymbol := match[1] // @ or ?
		varName := match[2]   // variable name (can be empty)

		// Determine variable type
		var varType string
		if varSymbol == "@" {
			varType = "hard"
		} else {
			varType = "soft"
		}

		// Create a unique key for this variable
		var varKey string
		if varName != "" {
			varKey = fileContext + "." + varName
		} else {
			// For bare {@} or {?}, use file context as key
			varKey = fileContext + ".variable"
		}

		// Skip if we've already seen this variable
		if seen[varKey] {
			continue
		}
		seen[varKey] = true

		variables = append(variables, VariableInfo{
			Key:  varKey,
			Type: varType,
			Name: varName,
		})
	}

	return variables
}

// ===== VARIABLE PROMPTING =====

// PromptForVariables prompts user for variable values and returns substitution map
func PromptForVariables(preset string, persist bool) (map[string]string, error) {
	substitutions := make(map[string]string)

	// Load variables.toml to get hard variables
	variablesHandler, err := LoadPresetFile(preset, "variables")
	if err != nil {
		return nil, fmt.Errorf(display.ErrVariableLoadFailed)
	}

	// Find all variables across all TOML files
	variables, err := FindAllVariables(preset)
	if err != nil {
		return nil, fmt.Errorf(display.ErrVariableLoadFailed)
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
				prompt = variable.Name + ": "
			} else {
				prompt = variable.Key + ": "
			}
		}

		// Create readline instance for this prompt
		rl, err := readline.New(prompt)
		if err != nil {
			return nil, fmt.Errorf(display.ErrReadlineSetup)
		}

		// Pre-fill with current value for better UX
		if currentValue != "" {
			rl.WriteStdin([]byte(currentValue))
		}

		userInput, err := rl.Readline()
		rl.Close()

		if err != nil {
			return nil, fmt.Errorf(display.ErrInputRead)
		}

		userInput = strings.TrimSpace(userInput)

		if variable.Type == "hard" && userInput == "" && currentValue != "" {
			// Keep existing value for hard variables if user presses Enter
			substitutions[variable.Key] = currentValue
		} else if userInput != "" {
			substitutions[variable.Key] = userInput

			// Save hard variables to variables.toml
			if variable.Type == "hard" {
				variablesHandler.Set(variable.Key, userInput)
				err := SavePresetFile(preset, "variables", variablesHandler)
				if err != nil {
					return nil, fmt.Errorf(display.ErrVariableSaveFailed)
				}
			}
		}
	}

	return substitutions, nil
}

// PromptForSpecificVariables prompts only for specified variables
func PromptForSpecificVariables(preset string, variableNames []string, persist bool) (map[string]string, error) {
	substitutions := make(map[string]string)

	// Load variables.toml to get hard variables
	variablesHandler, err := LoadPresetFile(preset, "variables")
	if err != nil {
		return nil, fmt.Errorf(display.ErrVariableLoadFailed)
	}

	// Find all variables across all TOML files
	allVariables, err := FindAllVariables(preset)
	if err != nil {
		return nil, fmt.Errorf(display.ErrVariableLoadFailed)
	}

	// Filter to only requested variables (or all if empty array)
	var targetVariables []VariableInfo
	if len(variableNames) == 0 {
		// Empty array means "all variables" (-v flag used with no args)
		targetVariables = allVariables
	} else {
		// Specific variables requested
		for _, variable := range allVariables {
			for _, requestedName := range variableNames {
				if variable.Key == requestedName || variable.Name == requestedName {
					targetVariables = append(targetVariables, variable)
					break
				}
			}
		}
	}

	// Use same prompting logic as PromptForVariables but on filtered set
	for _, variable := range targetVariables {
		var prompt string
		var currentValue string

		if variable.Type == "hard" {
			// Hard variables: use stored value if exists, show for editing
			currentValue = variablesHandler.GetAsString(variable.Key)
			if variable.Name != "" {
				prompt = "@" + variable.Name + ": "
			} else {
				prompt = "@" + variable.Key + ": "
			}
		} else {
			// Soft variables: always prompt with empty input
			if variable.Name != "" {
				prompt = "?" + variable.Name + ": "
			} else {
				prompt = "?" + variable.Key + ": "
			}
		}

		// Create readline instance for this prompt
		rl, err := readline.New(prompt)
		if err != nil {
			return nil, fmt.Errorf(display.ErrReadlineSetup)
		}

		// Pre-fill with current value for hard variables
		if variable.Type == "hard" && currentValue != "" {
			rl.WriteStdin([]byte(currentValue))
		}

		userInput, err := rl.Readline()
		rl.Close()

		if err != nil {
			return nil, fmt.Errorf(display.ErrInputRead)
		}

		userInput = strings.TrimSpace(userInput)

		if variable.Type == "hard" && userInput == "" && currentValue != "" {
			// Keep existing value for hard variables if user presses Enter
			substitutions[variable.Key] = currentValue
		} else if userInput != "" {
			substitutions[variable.Key] = userInput

			// Save hard variables to variables.toml
			if variable.Type == "hard" {
				variablesHandler.Set(variable.Key, userInput)
				err := SavePresetFile(preset, "variables", variablesHandler)
				if err != nil {
					return nil, fmt.Errorf(display.ErrVariableSaveFailed)
				}
			}
		}
	}

	return substitutions, nil
}

// ===== VARIABLE STORAGE & SUBSTITUTION =====

// SubstituteVariables replaces variables in TOML handler with actual values using simple regex
func SubstituteVariables(handler *TomlHandler, substitutions map[string]string) error {
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
	// Regex to find all {@} and {?} patterns
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
	handler, err := LoadPresetFile(preset, "variables")
	if err != nil {
		return err
	}

	// Store hard variable with empty initial value (will be set during call command)
	// Simple flat structure: "path.to.field" = "current_value"
	// Note: varName can be empty (bare @) - that's fine, we store by field path
	handler.Set(key, "") // Empty value initially

	// Save variables.toml
	return SavePresetFile(preset, "variables", handler)
}
