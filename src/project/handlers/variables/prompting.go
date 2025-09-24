package variables

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// Variable represents a detected variable in TOML values
type Variable struct {
	Name    string // Variable name without prefix
	Type    string // "soft" for ?, "hard" for @
	Current string // Current value for hard variables
}

// PromptForVariables prompts user for variable values and returns substitution map
func PromptForVariables(preset string, persist bool) (map[string]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	substitutions := make(map[string]string)

	// Load variables.toml to get hard variables
	variablesHandler, err := presets.LoadPresetFile(preset, "variables")
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
						return nil, fmt.Errorf(display.ErrVariableSaveFailed)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(display.ErrVariableLoadFailed)
	}

	return substitutions, nil
}