package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
)


// Check displays TOML file contents in a clean, readable format
func Check(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(errors.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(errors.ErrTargetRequired)
	}

	// Special handling for history target
	if strings.ToLower(cmd.Target) == "history" {
		return checkHistory(cmd)
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

// checkHistory handles history-related check commands
func checkHistory(cmd parser.Command) error {
	// Check if specific history number is requested
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		// Handle specific history number (e.g., saul api check history 1)
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
			// Use the last (highest numbered) response
			numberStr = strconv.Itoa(len(responses))
		}

		number, err := strconv.Atoi(numberStr)
		if err != nil {
			return fmt.Errorf("invalid history number: %s", numberStr)
		}

		return displayHistoryResponse(cmd.Preset, number, cmd.RawOutput)
	}

	// List all history responses (interactive menu)
	return listHistoryResponses(cmd.Preset, cmd.RawOutput)
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

	// Formatted mode: show interactive menu
	content := ""
	for i, response := range responses {
		content += fmt.Sprintf("%d. %s %s (%s)\n",
			i+1,
			response.Method,
			response.URL,
			response.Timestamp)
	}

	formatted := display.FormatSection("History", content, fmt.Sprintf("%d responses", len(responses)))
	display.Plain(formatted)

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

	// Formatted mode: use the Phase 4B response formatting
	content := ""

	// Add request metadata
	content += fmt.Sprintf("Method: %s\nURL: %s\nTimestamp: %s\nStatus: %s\n\n",
		response.Method, response.URL, response.Timestamp, response.Status)

	// Add response body with smart formatting
	switch v := response.Body.(type) {
	case string:
		// Try to parse as JSON for pretty formatting
		var jsonObj interface{}
		if json.Unmarshal([]byte(v), &jsonObj) == nil {
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				content += string(prettyJSON)
			} else {
				content += v
			}
		} else {
			content += v
		}
	default:
		// Marshal as JSON
		if jsonData, err := json.MarshalIndent(v, "", "  "); err == nil {
			content += string(jsonData)
		} else {
			content += fmt.Sprintf("%v", v)
		}
	}

	formatted := display.FormatSection(
		fmt.Sprintf("History Response %d", number),
		content,
		fmt.Sprintf("%s • %s", response.Status, response.Timestamp))
	display.Plain(formatted)

	return nil
}