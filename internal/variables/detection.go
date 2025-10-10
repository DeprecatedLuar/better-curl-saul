package variables

import (
	"os"
	"regexp"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/workspace"
)

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

// FindAllVariables scans all TOML files in preset to find variables using simple regex
func FindAllVariables(preset string) ([]VariableInfo, error) {
	var variables []VariableInfo
	targets := []string{"body", "headers", "query", "request"}

	// Get preset path
	presetPath, err := workspace.GetPresetPath(preset)
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

	// Regex to find {@ } and {?} patterns
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
			// For bare {@ } or {?}, use file context as key
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