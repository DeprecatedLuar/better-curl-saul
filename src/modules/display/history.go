package display

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
)

// DisplayTOMLFile shows the entire TOML file in a clean format
func DisplayTOMLFile(handler *toml.TomlHandler, target string, preset string, rawOutput bool) error {
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

// ListHistoryResponses shows all available history responses
func ListHistoryResponses(preset string, rawOutput bool) error {
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
		path := ExtractPath(response.URL)

		// Parse status for status code
		statusCode := ExtractStatusCode(response.Status)

		// Format relative time
		relativeTime := FormatRelativeTime(response.Timestamp)

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

// DisplayHistoryResponse shows a specific history response with formatting
func DisplayHistoryResponse(preset string, number int, rawOutput bool) error {
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

	// Get JSON data and format for display
	jsonStr := response.Body.(string)
	content := formatHistoryContent([]byte(jsonStr), rawOutput)

	if rawOutput {
		fmt.Print(content)
	} else {
		formatted := FormatSection(
			fmt.Sprintf("History Response %d", number),
			content,
			fmt.Sprintf("%s • %s", response.Status, FormatRelativeTime(response.Timestamp)))
		Plain(formatted)
	}

	return nil
}

// ExtractPath extracts the path component from a URL for clean display
func ExtractPath(urlStr string) string {
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

// ExtractStatusCode extracts the numeric status code from status string like "200 OK"
func ExtractStatusCode(status string) string {
	parts := strings.Fields(status)
	if len(parts) > 0 {
		return parts[0]
	}
	return status
}

// FormatRelativeTime formats timestamp into relative time like "2m ago"
func FormatRelativeTime(timestamp string) string {
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

// GetMostRecentResponseNumber returns the number of the most recent response
func GetMostRecentResponseNumber(preset string) (int, error) {
	responses, err := presets.ListHistoryResponses(preset)
	if err != nil {
		return 0, fmt.Errorf("failed to load history: %v", err)
	}
	if len(responses) == 0 {
		return 0, fmt.Errorf("no history found for preset '%s'", preset)
	}
	return len(responses), nil // Most recent is highest number
}

// ParseResponseNumber parses response number from string, handling "last" alias
func ParseResponseNumber(numberStr string, preset string) (int, error) {
	// Handle "last" alias for most recent response
	if strings.ToLower(numberStr) == "last" {
		return GetMostRecentResponseNumber(preset)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return 0, fmt.Errorf("invalid response number: %s", numberStr)
	}
	return number, nil
}

// formatHistoryContent formats JSON content for history display
func formatHistoryContent(jsonData []byte, rawMode bool) string {
	if rawMode {
		// Raw mode: just pretty-print JSON
		var jsonObj interface{}
		if json.Unmarshal(jsonData, &jsonObj) == nil {
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				return string(prettyJSON)
			}
		}
		return string(jsonData)
	}

	// Non-raw mode: pretty-print JSON (simplified version without TOML conversion)
	var jsonObj interface{}
	if json.Unmarshal(jsonData, &jsonObj) == nil {
		if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
			return string(prettyJSON)
		}
	}
	return string(jsonData)
}