package handlers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/handlers/http"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/core"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// ExecuteCallCommand handles HTTP execution for call commands
func ExecuteCallCommand(cmd core.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	// Check if preset exists first
	presetPath, err := presets.GetPresetPath(cmd.Preset)
	if err != nil {
		return fmt.Errorf(display.ErrDirectoryFailed)
	}

	// Check if preset directory exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf(display.ErrPresetNotFound, cmd.Preset)
	}

	// Check for flags
	persist := false
	rawMode := cmd.RawOutput

	// Prompt for variables and get substitution map
	substitutions, err := PromptForVariables(cmd.Preset, persist)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}

	// Load each file as separate handler - no merging
	requestHandler := http.LoadPresetFile(cmd.Preset, "request")
	headersHandler := http.LoadPresetFile(cmd.Preset, "headers")
	bodyHandler := http.LoadPresetFile(cmd.Preset, "body")
	queryHandler := http.LoadPresetFile(cmd.Preset, "query")

	// Apply variable substitutions to each separately
	err = SubstituteVariables(requestHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}
	err = SubstituteVariables(headersHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}
	err = SubstituteVariables(bodyHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}
	err = SubstituteVariables(queryHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}

	// Build HTTP request components explicitly - no guessing
	request, err := http.BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
	if err != nil {
		return fmt.Errorf(display.ErrRequestBuildFailed)
	}

	// Execute the HTTP request
	response, err := http.ExecuteHTTPRequest(request)
	if err != nil {
		return fmt.Errorf(display.ErrHTTPRequestFailed)
	}

	// Check if history is enabled and store response
	err = storeResponseHistory(cmd.Preset, request, response)
	if err != nil {
		// Don't fail the whole request if history storage fails
		display.Warning(display.WarnHistoryFailed)
	}

	// Display response with filtering support
	http.DisplayResponse(response, rawMode, cmd.Preset)

	return nil
}

// storeResponseHistory stores the HTTP response in history if enabled
func storeResponseHistory(preset string, request *http.HTTPRequestConfig, response *resty.Response) error {
	// Load request.toml to check for history configuration
	requestHandler, err := presets.LoadPresetFile(preset, "request")
	if err != nil {
		return nil // If we can't load request config, skip history
	}

	// Get history count from request.toml
	historyCountValue := requestHandler.Get("history_count")
	if historyCountValue == nil {
		return nil // No history configured
	}

	// Convert to int
	var historyCount int
	switch v := historyCountValue.(type) {
	case int:
		historyCount = v
	case int64:
		historyCount = int(v)
	case string:
		historyCount, err = strconv.Atoi(v)
		if err != nil {
			return nil // Invalid history count, skip
		}
	default:
		return nil // Invalid type, skip
	}

	if historyCount <= 0 {
		return nil // History disabled
	}

	// Convert response headers to map for storage
	headers := make(map[string]string)
	for key, values := range response.Header() {
		if len(values) > 0 {
			headers[key] = values[0] // Store first value
		}
	}

	// Parse response body as interface{} for JSON storage
	var body interface{}
	if response.Body() != nil && len(response.Body()) > 0 {
		// Try to unmarshal as JSON first
		if err := response.Result(); err == nil {
			body = string(response.Body()) // Store as string if JSON parsing fails
		} else {
			body = string(response.Body())
		}
	}

	// Format duration as a human-readable string
	duration := fmt.Sprintf("%.3fs", response.Time().Seconds())

	// Store the response
	responseData := presets.HistoryResponse{
		Method:   request.Method,
		URL:      request.URL,
		Status:   response.Status(),
		Duration: duration,
		Headers:  headers,
		Body:     body,
	}

	return presets.StoreResponse(preset, responseData, historyCount)
}

