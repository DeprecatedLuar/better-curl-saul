package executor

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor/http"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// ExecuteCallCommand handles HTTP execution for call commands
func ExecuteCallCommand(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(errors.ErrPresetNameRequired)
	}

	// Check if preset exists first
	presetPath, err := presets.GetPresetPath(cmd.Preset)
	if err != nil {
		return fmt.Errorf(errors.ErrDirectoryFailed)
	}

	// Check if preset directory exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf(errors.ErrPresetNotFound, cmd.Preset)
	}

	// Check for flags
	persist := false
	rawMode := cmd.RawOutput

	// Prompt for variables and get substitution map
	substitutions, err := PromptForVariables(cmd.Preset, persist)
	if err != nil {
		return fmt.Errorf(errors.ErrVariableLoadFailed)
	}

	// Load each file as separate handler - no merging
	requestHandler := http.LoadPresetFile(cmd.Preset, "request")
	headersHandler := http.LoadPresetFile(cmd.Preset, "headers")
	bodyHandler := http.LoadPresetFile(cmd.Preset, "body")
	queryHandler := http.LoadPresetFile(cmd.Preset, "query")

	// Apply variable substitutions to each separately
	err = SubstituteVariables(requestHandler, substitutions)
	if err != nil {
		return fmt.Errorf(errors.ErrVariableLoadFailed)
	}
	err = SubstituteVariables(headersHandler, substitutions)
	if err != nil {
		return fmt.Errorf(errors.ErrVariableLoadFailed)
	}
	err = SubstituteVariables(bodyHandler, substitutions)
	if err != nil {
		return fmt.Errorf(errors.ErrVariableLoadFailed)
	}
	err = SubstituteVariables(queryHandler, substitutions)
	if err != nil {
		return fmt.Errorf(errors.ErrVariableLoadFailed)
	}

	// Build HTTP request components explicitly - no guessing
	request, err := http.BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
	if err != nil {
		return fmt.Errorf(errors.ErrRequestBuildFailed)
	}

	// Execute the HTTP request
	response, err := http.ExecuteHTTPRequest(request)
	if err != nil {
		return fmt.Errorf(errors.ErrHTTPRequestFailed)
	}

	// Display response with filtering support
	http.DisplayResponse(response, rawMode, cmd.Preset)

	return nil
}

