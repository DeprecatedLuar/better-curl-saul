package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/display"
)


// Check displays TOML file contents in a clean, readable format
func Check(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required for check command")
	}
	if cmd.Target == "" {
		return fmt.Errorf("target required (body, headers, query, request, variables, filters)")
	}

	// Normalize target aliases
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf("invalid target '%s'. Use: body, headers/header, query, request, variables, filters", cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf("failed to load %s.toml: %v", cmd.Target, err)
	}

	// Special handling for request fields (single values)
	if cmd.Target == "request" && len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf("'%s' not set in request", key)
		}
		fmt.Println(value)
		return nil
	}

	// Get specific key if provided
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf("key '%s' not found in %s", key, cmd.Target)
		}

		// Format based on type
		switch v := value.(type) {
		case []interface{}:
			// Array format
			fmt.Printf("%s = [", key)
			for i, item := range v {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("\"%v\"", item)
			}
			fmt.Println("]")
		default:
			// Simple value
			fmt.Printf("%s = \"%v\"\n", key, value)
		}
		return nil
	}

	// Display entire file contents
	return displayTOMLFile(handler, cmd.Target, cmd.Preset)
}

// displayTOMLFile shows the entire TOML file in a clean format
func displayTOMLFile(handler *toml.TomlHandler, target string, preset string) error {
	// Capitalize target for display
	displayTarget := strings.ToUpper(target[:1]) + target[1:]
	fmt.Println(display.SectionHeader(displayTarget))
	fmt.Println(display.SectionFooter())

	// Get the file path and read raw contents
	presetPath, err := presets.GetPresetPath(preset)
	if err != nil {
		// Fall back to simple display if we can't get the preset path
		fmt.Println("(Unable to display full file contents)")
		fmt.Println(display.SectionFooter())
		return nil
	}

	filePath := filepath.Join(presetPath, target+".toml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("(File is empty or doesn't exist)")
		fmt.Println(display.SectionFooter())
		return nil
	}

	// Display raw TOML content
	fmt.Print(string(content))
	fmt.Println(display.SectionFooter())
	return nil
}