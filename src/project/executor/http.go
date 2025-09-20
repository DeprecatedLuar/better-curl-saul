package executor

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor/http"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// ExecuteCallCommand handles HTTP execution for call commands
func ExecuteCallCommand(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required for call command")
	}

	// Check if preset exists first
	presetPath, err := presets.GetPresetPath(cmd.Preset)
	if err != nil {
		return fmt.Errorf("failed to get preset path: %v", err)
	}

	// Check if preset directory exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf("preset '%s' does not exist. Create it first with: saul %s", cmd.Preset, cmd.Preset)
	}

	// Check for flags (simple flag parsing for now)
	persist := false
	rawMode := false
	// TODO: Implement proper flag parsing in parser package

	// Prompt for variables and get substitution map
	substitutions, err := PromptForVariables(cmd.Preset, persist)
	if err != nil {
		return fmt.Errorf("variable prompting failed: %v", err)
	}

	// Load each file as separate handler - no merging
	requestHandler := http.LoadPresetFile(cmd.Preset, "request")
	headersHandler := http.LoadPresetFile(cmd.Preset, "headers")
	bodyHandler := http.LoadPresetFile(cmd.Preset, "body")
	queryHandler := http.LoadPresetFile(cmd.Preset, "query")

	// Apply variable substitutions to each separately
	err = SubstituteVariables(requestHandler, substitutions)
	if err != nil {
		return fmt.Errorf("request variable substitution failed: %v", err)
	}
	err = SubstituteVariables(headersHandler, substitutions)
	if err != nil {
		return fmt.Errorf("headers variable substitution failed: %v", err)
	}
	err = SubstituteVariables(bodyHandler, substitutions)
	if err != nil {
		return fmt.Errorf("body variable substitution failed: %v", err)
	}
	err = SubstituteVariables(queryHandler, substitutions)
	if err != nil {
		return fmt.Errorf("query variable substitution failed: %v", err)
	}

	// Build HTTP request components explicitly - no guessing
	request, err := http.BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
	if err != nil {
		return fmt.Errorf("failed to build HTTP request: %v", err)
	}

	// Execute the HTTP request
	response, err := http.ExecuteHTTPRequest(request)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}

	// Display response with filtering support
	http.DisplayResponse(response, rawMode, cmd.Preset)

	return nil
}

