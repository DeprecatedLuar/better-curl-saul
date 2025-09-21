package parser

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
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
	RawOutput     bool          // For --raw flag
}

type KeyValuePair struct {
	Key   string
	Value string
}

func ParseCommand(args []string) (Command, error) {
	var cmd Command

	if len(args) < 1 {
		return cmd, fmt.Errorf(errors.ErrArgumentsNeeded)
	}

	// Parse flags and filter them out of args
	filteredArgs, err := parseFlags(args, &cmd)
	if err != nil {
		return cmd, err
	}
	args = filteredArgs

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

	// Handle regular commands with key=value syntax (supports space-separated)
	if len(args) > 2 {
		cmd.Target = args[2]
	}
	if len(args) > 3 {
		keyValueArgs := args[3:]

		// Parse space-separated key=value pairs
		pairs, err := parseSpaceSeparatedKeyValues(keyValueArgs)
		if err != nil {
			return cmd, fmt.Errorf(errors.ErrInvalidKeyValue)
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

// parseSpaceSeparatedKeyValues handles space-separated key=value arguments
func parseSpaceSeparatedKeyValues(args []string) ([]KeyValuePair, error) {
	var pairs []KeyValuePair

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf(errors.ErrInvalidKeyValue)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		pairs = append(pairs, KeyValuePair{Key: key, Value: value})
	}

	return pairs, nil
}

// parseFlags extracts flags from args and sets them in cmd, returning filtered args
func parseFlags(args []string, cmd *Command) ([]string, error) {
	var filteredArgs []string

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			// Handle long flags
			switch arg {
			case "--raw":
				cmd.RawOutput = true
			default:
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// Not a flag, keep in filtered args
			filteredArgs = append(filteredArgs, arg)
		}
	}

	return filteredArgs, nil
}

