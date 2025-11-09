package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
)


// Get displays TOML file contents in a clean, readable format
func Get(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	// Early return for curl export: --raw flag with no target
	if cmd.RawOutput && cmd.Target == "" {
		// Load preset files
		requestHandler, _ := workspace.LoadPresetFile(cmd.Preset, "request")
		headersHandler, _ := workspace.LoadPresetFile(cmd.Preset, "headers")
		queryHandler, _ := workspace.LoadPresetFile(cmd.Preset, "query")
		bodyHandler, _ := workspace.LoadPresetFile(cmd.Preset, "body")

		// Extract data
		method := requestHandler.GetAsString("method")
		baseURL := requestHandler.GetAsString("url")

		headers := make(map[string]string)
		for _, key := range headersHandler.Keys() {
			headers[key] = headersHandler.GetAsString(key)
		}

		query := make(map[string]string)
		for _, key := range queryHandler.Keys() {
			query[key] = queryHandler.GetAsString(key)
		}

		var body []byte
		if len(bodyHandler.Keys()) > 0 {
			body, _ = bodyHandler.ToJSON()
		}

		curlCmd, err := internal.ExportToCurl(method, baseURL, headers, query, body)
		if err != nil {
			return err
		}
		fmt.Print(curlCmd)
		return nil
	}

	if cmd.Target == "" {
		return fmt.Errorf(display.ErrTargetRequired)
	}

	// Special handling for history and response targets
	if strings.ToLower(cmd.Target) == "history" {
		return getHistory(cmd)
	}
	if strings.ToLower(cmd.Target) == "response" {
		// Check if this is field extraction on most recent response
		if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
			keyStr := cmd.KeyValuePairs[0].Key
			// Try to parse as response number first (integers, "last")
			if _, err := ParseResponseNumber(keyStr, cmd.Preset); err == nil {
				// It's a valid response number, use existing response logic
				return getResponse(cmd)
			}
			// Parsing failed - assume it's a field name and validate
			if !isFieldName(keyStr) {
				return fmt.Errorf(display.ErrUnknownResponseField, keyStr)
			}
			// Valid field name - extract from most recent response
			return getResponseFieldMostRecent(cmd)
		}
		// No KeyValuePairs - show most recent response
		return getResponse(cmd)
	} else if strings.HasPrefix(strings.ToLower(cmd.Target), "response") && len(cmd.Target) > 8 {
		// Handle response1, response2, etc. with field extraction
		return getResponseWithField(cmd)
	}

	// Load the TOML file for the target
	handler, err := workspace.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf(display.ErrFileLoadFailed, cmd.Target+".toml")
	}

	// Special handling for request fields (single values)
	if cmd.Target == "request" && len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key

		// Map "history" to "history_count" for lookup
		if strings.ToLower(key) == "history" {
			key = "history_count"
		}

		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf(display.ErrKeyNotFound, cmd.KeyValuePairs[0].Key, cmd.Target)
		}

		// Always print raw value (Unix philosophy)
		fmt.Println(value)
		return nil
	}

	// Get specific key if provided
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf(display.ErrKeyNotFound, key, cmd.Target)
		}

		// Always print raw value (Unix philosophy)
		switch v := value.(type) {
		case []interface{}:
			// For arrays, print as space-separated values
			for i, item := range v {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(item)
			}
			fmt.Println() // Add newline after array
		default:
			// Simple value
			fmt.Println(value)
		}
		return nil
	}

	// Display entire file contents
	return DisplayTOMLFile(handler, cmd.Target, cmd.Preset, cmd.RawOutput)
}


// getHistory handles history listing (LIST operation only)
func getHistory(cmd parser.Command) error {
	// History command only lists responses - no specific response access
	return ListHistoryResponses(cmd.Preset, cmd.RawOutput)
}

// getResponse handles response content fetching for most recent response only
// Space-separated format (response 1) is no longer supported - use response1 instead
func getResponse(cmd parser.Command) error {
	// Always get most recent response - no space-separated number support
	number, err := GetMostRecentResponseNumber(cmd.Preset)
	if err != nil {
		return fmt.Errorf(display.ErrNoHistory, cmd.Preset)
	}

	return DisplayHistoryResponse(cmd.Preset, number, cmd.RawOutput)
}

// getResponseWithField handles response field extraction (e.g., response1 body, response2 headers)
// Also handles whole response display (e.g., response1 with no field specified)
func getResponseWithField(cmd parser.Command) error {
	// Extract response number from target (response1 -> 1)
	numberStr := cmd.Target[8:] // Remove "response" prefix
	number, err := ParseResponseNumber(numberStr, cmd.Preset)
	if err != nil {
		return err
	}

	// Check if field is specified
	if len(cmd.KeyValuePairs) == 0 || cmd.KeyValuePairs[0].Key == "" {
		// No field specified - show whole response (single-line support)
		return DisplayHistoryResponse(cmd.Preset, number, cmd.RawOutput)
	}

	fieldName := strings.ToLower(cmd.KeyValuePairs[0].Key)

	// Load the response
	response, err := workspace.LoadHistoryResponse(cmd.Preset, number)
	if err != nil {
		return err
	}

	// Extract and display the requested field
	return displayResponseField(response, fieldName, cmd.RawOutput, cmd.Preset)
}

// displayResponseField extracts and displays a specific field from response
func displayResponseField(response *workspace.HistoryResponse, fieldName string, rawOutput bool, preset string) error {
	switch fieldName {
	case "body":
		// Use EXACT same filtering logic as live API calls
		if response.Body == nil {
			fmt.Println("(empty body)")
			return nil
		}

		// Body is stored as string - convert to bytes exactly like live API uses response.Body()
		var bodyBytes []byte
		switch v := response.Body.(type) {
		case string:
			bodyBytes = []byte(v) // Convert stored string to bytes (same as response.Body())
		default:
			// Fallback: marshal to JSON if not stored as string
			var err error
			bodyBytes, err = json.Marshal(v)
			if err != nil {
				return fmt.Errorf(display.ErrResponseProcessFailed, err)
			}
		}

		// Use EXACT same function as live API: FormatResponseContent (applies filtering + TOML conversion)
		formattedContent := internal.FormatResponseContent(bodyBytes, preset, rawOutput)
		fmt.Print(formattedContent)

	case "headers":
		// Headers should NOT be filtered (like live API) - show as-is with TOML formatting
		if response.Headers == nil {
			fmt.Println("(no headers)")
			return nil
		}

		headersJSON, err := json.Marshal(response.Headers)
		if err != nil {
			return fmt.Errorf(display.ErrResponseProcessFailed, err)
		}

		// Format headers without filtering (direct TOML conversion)
		if tomlFormatted := internal.FormatAsToml(headersJSON); tomlFormatted != "" {
			fmt.Print(tomlFormatted)
		} else {
			// Fallback to pretty JSON if TOML conversion fails
			if prettyJSON, err := json.MarshalIndent(response.Headers, "", "  "); err == nil {
				fmt.Print(string(prettyJSON))
			} else {
				fmt.Print(response.Headers)
			}
		}

	case "status":
		fmt.Println(response.Status)

	case "url":
		fmt.Println(response.URL)

	case "method":
		fmt.Println(response.Method)

	case "duration":
		fmt.Println(response.Duration)

	default:
		return fmt.Errorf(display.ErrUnknownResponseField, fieldName)
	}

	return nil
}

// isFieldName checks if a string is a valid field name for response extraction
func isFieldName(s string) bool {
	validFields := []string{"body", "headers", "status", "url", "method", "duration"}
	s = strings.ToLower(s)
	for _, field := range validFields {
		if s == field {
			return true
		}
	}
	return false
}

// getResponseFieldMostRecent handles field extraction from most recent response (e.g., "response body")
func getResponseFieldMostRecent(cmd parser.Command) error {
	// Get most recent response number
	number, err := GetMostRecentResponseNumber(cmd.Preset)
	if err != nil {
		return fmt.Errorf(display.ErrNoHistory, cmd.Preset)
	}

	// Get the field name
	fieldName := strings.ToLower(cmd.KeyValuePairs[0].Key)

	// Load the response
	response, err := workspace.LoadHistoryResponse(cmd.Preset, number)
	if err != nil {
		return err
	}

	// Extract and display the requested field
	return displayResponseField(response, fieldName, cmd.RawOutput, cmd.Preset)
}






