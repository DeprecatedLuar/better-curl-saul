package handlers

import (
	"github.com/DeprecatedLuar/better-curl-saul/src/project/handlers/variables"
)

// Re-export types for backward compatibility
type Variable = variables.Variable
type VariableInfo = variables.VariableInfo

// Re-export functions for backward compatibility
var DetectVariableType = variables.DetectVariableType
var PromptForVariables = variables.PromptForVariables
var PromptForSpecificVariables = variables.PromptForSpecificVariables
var SubstituteVariables = variables.SubstituteVariables
var StoreVariableInfo = variables.StoreVariableInfo

// findAllVariables is redirected to the new package
func findAllVariables(preset string) ([]variables.VariableInfo, error) {
	return variables.FindAllVariables(preset)
}