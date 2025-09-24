package commands

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/core"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)


// Get displays TOML file contents in a clean, readable format
func Get(cmd core.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(display.ErrTargetRequired)
	}

	// Special handling for history and response targets
	if strings.ToLower(cmd.Target) == "history" {
		return getHistory(cmd)
	}
	if strings.ToLower(cmd.Target) == "response" {
		return getResponse(cmd)
	}

	// Normalize target aliases
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf(display.ErrInvalidTarget, cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
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
func getHistory(cmd core.Command) error {
	// History command only lists responses - no specific response access
	return ListHistoryResponses(cmd.Preset, cmd.RawOutput)
}

// getResponse handles response content fetching (FETCH operation)
func getResponse(cmd core.Command) error {
	var number int
	var err error

	// Check if specific response number is provided
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		numberStr := cmd.KeyValuePairs[0].Key

		number, err = ParseResponseNumber(numberStr, cmd.Preset)
		if err != nil {
			return err
		}
	} else {
		// No number provided - default to most recent response
		number, err = GetMostRecentResponseNumber(cmd.Preset)
		if err != nil {
			return fmt.Errorf("no history found for preset '%s'", cmd.Preset)
		}
	}

	return DisplayHistoryResponse(cmd.Preset, number, cmd.RawOutput)
}






