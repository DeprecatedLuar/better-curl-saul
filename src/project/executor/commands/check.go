package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
)


// Check displays TOML file contents in a clean, readable format
func Check(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(errors.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(errors.ErrTargetRequired)
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
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf(errors.ErrKeyNotFound, key, cmd.Target)
		}
		
		if cmd.RawOutput {
			// Raw mode: print value only (no newline for scriptability)
			fmt.Print(value)
		} else {
			// Normal mode: formatted output
			fmt.Println(value)
		}
		return nil
	}

	// Get specific key if provided
	if len(cmd.KeyValuePairs) > 0 && cmd.KeyValuePairs[0].Key != "" {
		key := cmd.KeyValuePairs[0].Key
		value := handler.Get(key)
		if value == nil {
			return fmt.Errorf(errors.ErrKeyNotFound, key, cmd.Target)
		}

		if cmd.RawOutput {
			// Raw mode: print value only (no formatting, no newline for scriptability)
			switch v := value.(type) {
			case []interface{}:
				// For arrays, print as space-separated values
				for i, item := range v {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(item)
				}
			default:
				// Simple value
				fmt.Print(value)
			}
		} else {
			// Normal mode: formatted output
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
		if rawOutput {
			// Raw mode: silent failure for scriptability
			return nil
		}
		// Fall back to simple display if we can't get the preset path
		displayTarget := strings.ToUpper(target[:1]) + target[1:]
		content := "(Unable to display full file contents)"
		formatted := display.FormatSimpleSection(displayTarget, content)
		display.Plain(formatted)
		return nil
	}

	filePath := filepath.Join(presetPath, target+".toml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		if rawOutput {
			// Raw mode: silent failure for scriptability
			return nil
		}
		displayTarget := strings.ToUpper(target[:1]) + target[1:]
		emptyContent := "(File is empty or doesn't exist)"
		formatted := display.FormatSimpleSection(displayTarget, emptyContent)
		display.Plain(formatted)
		return nil
	}

	if rawOutput {
		// Raw mode: print file contents as-is (like cat)
		fmt.Print(string(content))
		return nil
	}

	// Normal mode: formatted display
	displayTarget := strings.ToUpper(target[:1]) + target[1:]
	size := formatFileSize(len(content))
	entryCount := calculateEntryCount(string(content))
	
	fileContent := strings.TrimSpace(string(content))
	formatted := display.FormatFileDisplay(displayTarget, size, entryCount, fileContent)
	display.Plain(formatted)
	
	return nil
}

// formatFileSize converts byte count to human-readable format
func formatFileSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d bytes", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	}
}

// calculateEntryCount estimates the number of entries in TOML content
func calculateEntryCount(content string) string {
	if content == "" {
		return "0"
	}
	
	lines := strings.Split(content, "\n")
	entryCount := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Count lines that contain assignments (key = value)
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			entryCount++
		}
	}
	
	if entryCount == 0 {
		// If no assignments found, count non-empty, non-comment lines
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				entryCount++
			}
		}
	}
	
	return fmt.Sprintf("%d", entryCount)
}