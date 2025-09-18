package parser

import (
	"fmt"
	"strings"
)

type Command struct {
	Global    string
	Preset    string
	Command   string
	Target    string
	ValueType string
	Key       string
	Value     string
	Mode      string
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
			cmd.Key = args[2]
			cmd.Value = args[3]
			return cmd, nil
		}
	}

	// Handle check command (special syntax: check target [key])
	if cmd.Command == "check" {
		if len(args) > 2 {
			// Check if it's a special request field (auto-map to request target)
			if isSpecialRequestCommand(args[2]) {
				cmd.Target = "request"
				cmd.Key = args[2]
			} else {
				cmd.Target = args[2]
				if len(args) > 3 {
					cmd.Key = args[3] // Optional key for specific field
				}
			}
		}
		return cmd, nil
	}

	// Handle regular commands with key=value syntax
	if len(args) > 2 {
		cmd.Target = args[2]
	}
	if len(args) > 3 {
		keyValue := args[3]
		parts := strings.SplitN(keyValue, "=", 2)
		if len(parts) != 2 {
			return cmd, fmt.Errorf("what am I even supposed to do with: %s? Value=Key c'mon not that hard buddy", keyValue)
		}

		cmd.Key = parts[0]
		cmd.Value = parts[1]
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
