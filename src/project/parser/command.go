package parser

import (
	"fmt"
	"strings"
)

type Command struct {
	Global        string
	Preset        string
	Command       string
	Target        string
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
	case "rm", "list", "version", "help", "call":
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

// parseCommaSeparatedKeyValues parses comma-separated key=value pairs with quote support
// Examples: "key1=value1,key2=value2" or "auth=Bearer123,type=\"application/json, charset=utf-8\""
func parseCommaSeparatedKeyValues(input string) ([]KeyValuePair, error) {
	var pairs []KeyValuePair
	var currentPair strings.Builder
	inQuotes := false
	
	for _, char := range input {
		switch char {
		case '"':
			inQuotes = !inQuotes
			currentPair.WriteRune(char)
		case ',':
			if inQuotes {
				currentPair.WriteRune(char)
			} else {
				// End of current pair
				pairStr := strings.TrimSpace(currentPair.String())
				if pairStr != "" {
					kvp, err := parseKeyValuePair(pairStr)
					if err != nil {
						return nil, err
					}
					pairs = append(pairs, kvp)
				}
				currentPair.Reset()
			}
		default:
			currentPair.WriteRune(char)
		}
	}
	
	// Handle the last pair
	pairStr := strings.TrimSpace(currentPair.String())
	if pairStr != "" {
		kvp, err := parseKeyValuePair(pairStr)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, kvp)
	}
	
	if len(pairs) == 0 {
		return nil, fmt.Errorf("no valid key=value pairs found")
	}
	
	return pairs, nil
}

// parseKeyValuePair parses a single key=value pair and handles quotes
func parseKeyValuePair(input string) (KeyValuePair, error) {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return KeyValuePair{}, fmt.Errorf("invalid key=value format: %s", input)
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	// Remove surrounding quotes if present
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	
	return KeyValuePair{Key: key, Value: value}, nil
}
