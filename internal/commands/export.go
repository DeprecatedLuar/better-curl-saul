package commands

import (
	"encoding/json"
	"fmt"

	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Export handles exporting preset TOML files to curl command
func Export(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	// Check if preset exists
	basePreset := cmd.Preset
	if !workspace.PresetExists(basePreset) {
		return fmt.Errorf(display.ErrPresetNotFound, basePreset)
	}

	// Load request.toml
	reqHandler, err := workspace.LoadPresetFile(cmd.Preset, "request")
	if err != nil {
		return fmt.Errorf("failed to load request.toml: %v", err)
	}

	method := reqHandler.GetAsString("method")
	if method == "" {
		method = "GET"
	}

	baseURL := reqHandler.GetAsString("url")
	if baseURL == "" {
		return fmt.Errorf("no URL found in request.toml - cannot export")
	}

	// Load headers.toml
	headers := make(map[string]string)
	headersHandler, err := workspace.LoadPresetFile(cmd.Preset, "headers")
	if err == nil {
		headersMap := headersHandler.ToMap()
		for key, value := range headersMap {
			headers[key] = fmt.Sprintf("%v", value)
		}
	}

	// Load query.toml
	query := make(map[string]string)
	queryHandler, err := workspace.LoadPresetFile(cmd.Preset, "query")
	if err == nil {
		queryMap := queryHandler.ToMap()
		for key, value := range queryMap {
			query[key] = fmt.Sprintf("%v", value)
		}
	}

	// Load body.toml
	var bodyBytes []byte
	bodyHandler, err := workspace.LoadPresetFile(cmd.Preset, "body")
	if err == nil {
		bodyMap := bodyHandler.ToMap()
		if len(bodyMap) > 0 {
			// Convert TOML back to JSON
			bodyBytes, err = json.Marshal(bodyMap)
			if err != nil {
				return fmt.Errorf("failed to convert body to JSON: %v", err)
			}
		}
	}

	// Generate curl command using existing exporter
	curlCmd, err := internal.ExportToCurl(method, baseURL, headers, query, bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to generate curl command: %v", err)
	}

	// Display the curl command
	fmt.Println(curlCmd)

	return nil
}
