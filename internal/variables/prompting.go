package variables

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/chzyer/readline"
)

// Variable represents a detected variable in TOML values
type Variable struct {
	Name    string // Variable name without prefix
	Type    string // "soft" for ?, "hard" for @
	Current string // Current value for hard variables
}

// PromptForVariables prompts user for variable values and returns substitution map
func PromptForVariables(preset string, persist bool) (map[string]string, error) {
	substitutions := make(map[string]string)

	// Load variables.toml to get hard variables
	variablesHandler, err := workspace.LoadPresetFile(preset, "variables")
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
				err := workspace.SavePresetFile(preset, "variables", variablesHandler)
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
	variablesHandler, err := workspace.LoadPresetFile(preset, "variables")
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
				err := workspace.SavePresetFile(preset, "variables", variablesHandler)
				if err != nil {
					return nil, fmt.Errorf(display.ErrVariableSaveFailed)
				}
			}
		}
	}

	return substitutions, nil
}
