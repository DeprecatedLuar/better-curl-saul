package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type Command struct {
	Global        string
	Preset        string
	Command       string
	Target        string
	Targets       []string      // For bulk operations (space-separated args)
	ValueType     string
	Mode          string
	KeyValuePairs []KeyValuePair
}

type KeyValuePair struct {
	Key   string
	Value string
}

func ParseCommand(args []string) (Command, error) {
	var cmd Command

	if len(args) < 1 {
		return cmd, fmt.Errorf("you gonna need more arguments than that buddy (no pressure)")
	}

	switch args[0] {
	case "rm":
		cmd.Global = args[0]
		if len(args) >= 2 {
			// Handle multiple targets for rm: saul rm preset1 preset2 preset3
			cmd.Targets = args[1:]
			// Keep single target for backward compatibility
			cmd.Preset = args[1]
		}
		return cmd, nil
	case "list", "version", "help", "call":
		cmd.Global = args[0]
		if len(args) >= 2 {
			cmd.Preset = args[1]
		}
		return cmd, nil
	}

	cmd.Preset = args[0]

	if len(args) > 1 {
		cmd.Command = args[1]
	}

	// Handle special request commands with no-equals syntax
	if len(args) >= 4 && cmd.Command == "set" {
		if isSpecialRequestCommand(args[2]) {
			// Special syntax: "saul preset set url https://..." 
			cmd.Target = "request"
			cmd.KeyValuePairs = []KeyValuePair{
				{Key: args[2], Value: args[3]},
			}
			return cmd, nil
		}
	}

	// Handle check command (special syntax: check target [key])
	if cmd.Command == "check" {
		if len(args) > 2 {
			// Check if it's a special request field (auto-map to request target)
			if isSpecialRequestCommand(args[2]) {
				cmd.Target = "request"
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: args[3]}}
				} else {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: ""}}
				}
			} else {
				cmd.Target = args[2]
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[3], Value: ""}}
				}
			}
		}
		return cmd, nil
	}

	// Handle edit command (same syntax as check: edit target [key])
	if cmd.Command == "edit" {
		if len(args) > 2 {
			// Check if it's a special request field (auto-map to request target)
			if isSpecialRequestCommand(args[2]) {
				cmd.Target = "request"
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: args[3]}}
				} else {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: ""}}
				}
			} else {
				cmd.Target = args[2]
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[3], Value: ""}}
				}
				// Note: If no key provided (len(args) == 3), KeyValuePairs stays empty -> container editing
			}
		}
		return cmd, nil
	}

	// Handle regular commands with key=value syntax (supports comma-separated)
	if len(args) > 2 {
		cmd.Target = args[2]
	}
	if len(args) > 3 {
		keyValueInput := args[3]
		
		// Parse comma-separated key=value pairs with quote support
		pairs, err := parseCommaSeparatedKeyValues(keyValueInput)
		if err != nil {
			return cmd, fmt.Errorf("invalid key=value format: %v", err)
		}
		
		cmd.KeyValuePairs = pairs
	}

	return cmd, nil
}

// isSpecialRequestCommand checks if a command is a special request command (no = syntax)
func isSpecialRequestCommand(command string) bool {
	specialCommands := []string{"url", "method", "timeout"}
	command = strings.ToLower(command)
	
	for _, special := range specialCommands {
		if command == special {
			return true
		}
	}
	return false
}

// parseCommaSeparatedKeyValues uses simple Unix approach - right tool for each job
func parseCommaSeparatedKeyValues(input string) ([]KeyValuePair, error) {
	// Step 1: Array syntax (key=[...]) - simple string handling
	if isArraySyntax(input) {
		return parseSinglePair(input)
	}
	
	// Step 2: Check if multiple pairs exist (comma outside any quotes)
	if hasMultiplePairs(input) {
		// Multiple pairs: key1=val1,key2=val2 - use regex
		return parseMultiplePairs(input) 
	}
	
	// Step 3: Single pair - simple string split (most common case)
	return parseSinglePair(input)
}

// isArraySyntax detects array format: key=[...]
func isArraySyntax(input string) bool {
	return strings.Contains(input, "=[") && strings.HasSuffix(input, "]")
}

// hasMultiplePairs detects if input has multiple key=value pairs
func hasMultiplePairs(input string) bool {
	// Simple heuristic: count = signs
	// Multiple pairs will have multiple = signs
	return strings.Count(input, "=") > 1
}

// parseSinglePair handles single key=value (most common case)
// Works for: key=simple, key="quoted value", key=[array], key="value,with,commas"
func parseSinglePair(input string) ([]KeyValuePair, error) {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key=value format: %s", input)
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	// Remove surrounding quotes if present
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	
	return []KeyValuePair{{Key: key, Value: value}}, nil
}

// parseMultiplePairs handles comma-separated key=value pairs using regex
// Only used when multiple = signs detected: key1=val1,key2=val2
func parseMultiplePairs(input string) ([]KeyValuePair, error) {
	// Simple regex for multiple pairs
	pattern := `(\w+)=([^,=]+)`
	regex := regexp.MustCompile(pattern)
	
	matches := regex.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid key=value pairs found")
	}
	
	var pairs []KeyValuePair
	for _, match := range matches {
		key := match[1]
		value := strings.TrimSpace(match[2])
		
		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		
		pairs = append(pairs, KeyValuePair{Key: key, Value: value})
	}
	
	return pairs, nil
}
