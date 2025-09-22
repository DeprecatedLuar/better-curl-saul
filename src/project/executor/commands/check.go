package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
	httpModule "github.com/DeprecatedLuar/better-curl-saul/src/project/executor/http"
)


// Check displays TOML file contents in a clean, readable format
func Check(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(errors.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(errors.ErrTargetRequired)
	}

	// Special handling for history and response targets
	if strings.ToLower(cmd.Target) == "history" {
		return checkHistory(cmd)
	}
	if strings.ToLower(cmd.Target) == "response" {
		return checkResponse(cmd)
	}

	// Normalize target aliases
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf(errors.ErrInvalidTarget, cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf(errors.ErrFileLoadFailed, cmd.Target+".toml")
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
			return fmt.Errorf(errors.ErrKeyNotFound, cmd.KeyValuePairs[0].Key, cmd.Target)
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
			return fmt.Errorf(errors.ErrKeyNotFound, key, cmd.Target)
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
	return displayTOMLFile(handler, cmd.Target, cmd.Preset, cmd.RawOutput)
}

// displayTOMLFile shows the entire TOML file in a clean format
func displayTOMLFile(handler *toml.TomlHandler, target string, preset string, rawOutput bool) error {
	// Get the file path and read raw contents
	presetPath, err := presets.GetPresetPath(preset)
	if err != nil {
		// Silent failure - no file exists (Unix philosophy)
		return nil
	}

	filePath := filepath.Join(presetPath, target+".toml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Silent failure - no file content (Unix philosophy)
		return nil
	}

	// Always display raw file contents (Unix philosophy - like cat)
	fmt.Print(string(content))
	
	return nil
}

// checkHistory handles history listing (LIST operation only)
func checkHistory(cmd parser.Command) error {
	// History command only lists responses - no specific response access
	return listHistoryResponses(cmd.Preset, cmd.RawOutput)
}

// checkResponse handles response content fetching (FETCH operation)
func checkResponse(cmd parser.Command) error {
	var number int
	var err error

	// Check if specific response number is provided
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		numberStr := cmd.KeyValuePairs[0].Key

		// Handle "last" alias for most recent response
		if strings.ToLower(numberStr) == "last" {
			responses, err := presets.ListHistoryResponses(cmd.Preset)
			if err != nil {
				return fmt.Errorf("failed to load history: %v", err)
			}
			if len(responses) == 0 {
				return fmt.Errorf("no history found for preset '%s'", cmd.Preset)
			}
			number = len(responses) // Most recent is highest number
		} else {
			number, err = strconv.Atoi(numberStr)
			if err != nil {
				return fmt.Errorf("invalid response number: %s", numberStr)
			}
		}
	} else {
		// No number provided - default to most recent response (1)
		number = 1
	}

	return displayHistoryResponse(cmd.Preset, number, cmd.RawOutput)
}

// listHistoryResponses shows all available history responses
func listHistoryResponses(preset string, rawOutput bool) error {
	responses, err := presets.ListHistoryResponses(preset)
	if err != nil {
		return fmt.Errorf("failed to load history: %v", err)
	}

	if len(responses) == 0 {
		if rawOutput {
			// Silent in raw mode (Unix philosophy)
			return nil
		}
		fmt.Printf("No history found for preset '%s'\n", preset)
		return nil
	}

	if rawOutput {
		// Raw mode: just print numbers space-separated
		for i := range responses {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(i + 1)
		}
		fmt.Println()
		return nil
	}

	// Formatted mode: clean tabular output (reverse chronological order)
	for i := len(responses) - 1; i >= 0; i-- {
		displayIndex := len(responses) - i
		response := responses[i]

		// Extract path from URL for cleaner display
		path := extractPath(response.URL)

		// Parse status for status code
		statusCode := extractStatusCode(response.Status)

		// Format relative time
		relativeTime := formatRelativeTime(response.Timestamp)

		// Clean tabular format: "  1  POST /api/users    201  0.234s  2m ago"
		fmt.Printf("  %-2d %-4s %-20s %-3s %-8s %s\n",
			displayIndex,
			response.Method,
			path,
			statusCode,
			response.Duration,
			relativeTime)
	}

	return nil
}

// displayHistoryResponse shows a specific history response with formatting
func displayHistoryResponse(preset string, number int, rawOutput bool) error {
	response, err := presets.LoadHistoryResponse(preset, number)
	if err != nil {
		return err
	}

	if rawOutput {
		// Raw mode: just print the response body
		switch v := response.Body.(type) {
		case string:
			fmt.Print(v)
		default:
			// Try to marshal as JSON
			if jsonData, err := json.Marshal(v); err == nil {
				fmt.Print(string(jsonData))
			} else {
				fmt.Print(v)
			}
		}
		return nil
	}

	// Get JSON data and format using same logic as live responses
	jsonStr := response.Body.(string)
	content := httpModule.FormatResponseContent([]byte(jsonStr), preset, rawOutput)

	if rawOutput {
		fmt.Print(content)
	} else {
		formatted := display.FormatSection(
			fmt.Sprintf("History Response %d", number),
			content,
			fmt.Sprintf("%s • %s", response.Status, formatRelativeTime(response.Timestamp)))
		display.Plain(formatted)
	}

	return nil
}

// extractPath extracts the path component from a URL for clean display
func extractPath(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// If parsing fails, try to extract manually
		if idx := strings.Index(urlStr, "://"); idx != -1 {
			remaining := urlStr[idx+3:]
			if slashIdx := strings.Index(remaining, "/"); slashIdx != -1 {
				return remaining[slashIdx:]
			}
		}
		return urlStr // Return original if all parsing fails
	}

	path := parsedURL.Path
	if path == "" || path == "/" {
		path = "/"
	}

	// Add query parameters if they exist
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	return path
}

// extractStatusCode extracts the numeric status code from status string like "200 OK"
func extractStatusCode(status string) string {
	parts := strings.Fields(status)
	if len(parts) > 0 {
		return parts[0]
	}
	return status
}

// formatRelativeTime formats timestamp into relative time like "2m ago"
func formatRelativeTime(timestamp string) string {
	// Parse the timestamp
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "unknown"
	}

	// Calculate duration since then
	duration := time.Since(t)

	// Format into human-readable relative time
	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}

