package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/http"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/debug"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Call handles HTTP execution for call commands
func Call(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	// Check if preset exists first
	presetPath, err := workspace.GetPresetPath(cmd.Preset)
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
	var substitutions map[string]string

	if cmd.VariableFlags != nil {
		// -v flag was used (either with args or without)
		substitutions, err = workspace.PromptForSpecificVariables(cmd.Preset, cmd.VariableFlags, persist)
	} else {
		// No -v flag used = normal variable prompting
		substitutions, err = workspace.PromptForVariables(cmd.Preset, persist)
	}
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}

	// Load each file as separate handler - no merging
	requestHandler, err := http.LoadPresetFile(cmd.Preset, "request")
	if err != nil {
		return err
	}
	headersHandler, err := http.LoadPresetFile(cmd.Preset, "headers")
	if err != nil {
		return err
	}
	bodyHandler, err := http.LoadPresetFile(cmd.Preset, "body")
	if err != nil {
		return err
	}
	queryHandler, err := http.LoadPresetFile(cmd.Preset, "query")
	if err != nil {
		return err
	}

	debug.Log("body keys after load: %v", bodyHandler.Keys())

	// Apply variable substitutions to each separately
	err = workspace.SubstituteVariables(requestHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}
	err = workspace.SubstituteVariables(headersHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}
	err = workspace.SubstituteVariables(bodyHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}
	err = workspace.SubstituteVariables(queryHandler, substitutions)
	if err != nil {
		return fmt.Errorf(display.ErrVariableLoadFailed)
	}

	debug.Log("body keys after substitution: %v", bodyHandler.Keys())
	if bodyJSON, err := bodyHandler.ToJSON(); err == nil {
		debug.Log("body JSON: %s", string(bodyJSON))
	}

	// Build HTTP request components explicitly - no guessing
	request, err := http.BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
	if err != nil {
		return fmt.Errorf(display.ErrRequestBuildFailed)
	}

	debug.Log("request.Body length: %d", len(request.Body))
	debug.Log("request.Body content: %s", string(request.Body))

	// Handle dry-run mode
	if cmd.DryRun {
		return displayDryRunRequest(request)
	}

	// Execute the HTTP request (only if not dry-run)
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
	http.DisplayResponse(response, rawMode, cmd.Preset, cmd.ResponseFormat)

	return nil
}

// storeResponseHistory stores the HTTP response in history if enabled
func storeResponseHistory(preset string, request *http.HTTPRequestConfig, response *resty.Response) error {
	// Load request.toml to check for history configuration
	requestHandler, err := workspace.LoadPresetFile(preset, "request")
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
	responseData := workspace.HistoryResponse{
		Method:   request.Method,
		URL:      request.URL,
		Status:   response.Status(),
		Duration: duration,
		Headers:  headers,
		Body:     body,
	}

	return workspace.StoreResponse(preset, responseData, historyCount)
}

// displayDryRunRequest shows request details without executing
func displayDryRunRequest(request *http.HTTPRequestConfig) error {
	fmt.Printf("%s %s\n", request.Method, request.URL)

	if len(request.Headers) > 0 {
		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	if request.Body != nil && len(request.Body) > 0 {
		fmt.Println("Body:")
		fmt.Println("  " + strings.Replace(string(request.Body), "\n", "\n  ", -1))
	}

	if len(request.Query) > 0 {
		fmt.Println("Query Parameters:")
		for key, value := range request.Query {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("\n(Request not sent - dry run mode)")
	return nil
}
