// Package commands provides command parsing, execution logic and validation for Better-Curl-Saul.
// This package handles all command-related operations including argument parsing,
// command execution, and request validation.
package commands

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Command represents a parsed command with all its components
type Command struct {
	// Core command parsing
	Global        string
	Preset        string
	Command       string
	Target        string
	Targets       []string      // For bulk operations (space-separated args)
	ValueType     string
	Mode          string
	KeyValuePairs []KeyValuePair

	// Flags
	RawOutput        bool
	VariableFlags    []string
	ResponseFormat   string
	DryRun          bool
	Call            bool
	Create          bool
}

// KeyValuePair represents a key-value pair from command arguments
type KeyValuePair struct {
	Key   string
	Value string
}

// SessionProvider interface for getting current preset
type SessionProvider interface {
	HasCurrentPreset() bool
	GetCurrentPreset() string
}

func ParseCommand(args []string) (Command, error) {
	return ParseCommandWithSession(args, nil)
}

func ParseCommandWithSession(args []string, session SessionProvider) (Command, error) {
	var cmd Command

	if len(args) < 1 {
		return cmd, fmt.Errorf(display.ErrArgumentsNeeded)
	}

	// Check for list command aliases FIRST - skip flag parsing for them
	if isListCommand(args[0]) {
		cmd.Global = "list"
		cmd.Preset = args[0]  // Store the actual command (ls/exa/etc) for delegation
		return cmd, nil
	}

	// Parse flags and filter them out of args (only for non-system commands)
	filteredArgs, err := parseFlags(args, &cmd)
	if err != nil {
		return cmd, err
	}
	args = filteredArgs

	switch args[0] {
	case "create":
		cmd.Global = "create"
		if len(args) >= 2 {
			cmd.Preset = args[1]
			cmd.Create = true
		}
		return cmd, nil
	case "rm":
		cmd.Global = args[0]
		if len(args) >= 2 {
			// Handle multiple targets for rm: saul rm preset1 preset2 preset3
			cmd.Targets = args[1:]
			// Keep single target for backward compatibility
			cmd.Preset = args[1]
		}
		return cmd, nil
	case "version", "help", "update":
		cmd.Global = args[0]
		if len(args) >= 2 {
			cmd.Preset = args[1]
		}
		return cmd, nil
	case "switch":
		// Handle explicit switch command: saul switch variant
		if len(args) < 2 {
			return cmd, fmt.Errorf("variant name required for switch command")
		}
		if session == nil || !session.HasCurrentPreset() {
			return cmd, fmt.Errorf("no active preset")
		}
		cmd.Global = "switch"
		cmd.Preset = args[1]
		return cmd, nil
	}

	// Handle relative variant path: /variant
	if strings.HasPrefix(args[0], "/") {
		if session == nil || !session.HasCurrentPreset() {
			return cmd, fmt.Errorf("no active preset for relative variant path")
		}
		variantName := args[0][1:]
		if variantName == "" {
			return cmd, fmt.Errorf("variant name required")
		}

		// Extract base preset (strip existing variant if present)
		basePreset := strings.Split(session.GetCurrentPreset(), "/")[0]
		cmd.Preset = basePreset + "/" + variantName

		// If only the variant path is provided, treat as switch
		if len(args) == 1 {
			cmd.Global = "switch-variant"
			return cmd, nil
		}

		// Otherwise, continue with command processing
		if len(args) > 1 {
			cmd.Command = args[1]
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
			cmd.Target = normalizeTarget("request")
			cmd.KeyValuePairs = []KeyValuePair{
				{Key: args[2], Value: args[3]},
			}
			return cmd, nil
		}
	}

	// Handle get command (special syntax: get target [key])
	if cmd.Command == "get" {
		if len(args) > 2 {
			// Special cases: "history" and "response" as targets should not be treated as request fields
			if strings.ToLower(args[2]) == "history" || strings.ToLower(args[2]) == "response" {
				cmd.Target = args[2]
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[3], Value: ""}}
				}
			} else if isSpecialRequestCommand(args[2]) {
				// Check if it's a special request field (auto-map to request target)
				cmd.Target = normalizeTarget("request")
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: args[3]}}
				} else {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: ""}}
				}
			} else {
				cmd.Target = normalizeTarget(args[2])
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
				cmd.Target = normalizeTarget("request")
				if len(args) > 3 {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: args[3]}}
				} else {
					cmd.KeyValuePairs = []KeyValuePair{{Key: args[2], Value: ""}}
				}
			} else {
				cmd.Target = normalizeTarget(args[2])
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
		cmd.Target = normalizeTarget(args[2])
	}
	if len(args) > 3 {
		keyValueArgs := args[3:]

		// Special handling for filters - just field names, no key=value
		if cmd.Target == "filters" {
			var pairs []KeyValuePair
			for _, fieldName := range keyValueArgs {
				pairs = append(pairs, KeyValuePair{Key: "", Value: fieldName})
			}
			cmd.KeyValuePairs = pairs
		} else {
			// Parse space-separated key=value pairs for other targets
			pairs, err := parseSpaceSeparatedKeyValues(keyValueArgs)
			if err != nil {
				return cmd, fmt.Errorf(display.ErrInvalidKeyValue)
			}
			cmd.KeyValuePairs = pairs
		}
	}

	return cmd, nil
}

// isSpecialRequestCommand checks if a command is a special request command (no = syntax)
func isSpecialRequestCommand(command string) bool {
	specialCommands := []string{"url", "method", "timeout", "history"}
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
			return nil, fmt.Errorf(display.ErrInvalidKeyValue)
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
	skip := 0

	for i, arg := range args {
		if skip > 0 {
			skip--
			continue
		}

		if strings.HasPrefix(arg, "--") {
			// Handle long flags
			switch arg {
			case "--raw":
				cmd.RawOutput = true
			case "--headers-only":
				cmd.ResponseFormat = "headers-only"
			case "--body-only":
				cmd.ResponseFormat = "body-only"
			case "--status-only":
				cmd.ResponseFormat = "status-only"
			case "--dry-run":
				cmd.DryRun = true
			case "--call":
				cmd.Call = true
			case "--create":
				cmd.Create = true
			default:
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
			// Handle short flags
			flagPart := arg[1:] // Remove leading -
			if flagPart == "v" {
				// -v flag: collect all following non-flag arguments as variable names
				varCount := 0
				for j := i + 1; j < len(args); j++ {
					if strings.HasPrefix(args[j], "-") {
						break // Stop at next flag
					}
					cmd.VariableFlags = append(cmd.VariableFlags, args[j])
					varCount++
				}
				skip = varCount // Skip these args in main loop
				// If no variables specified, empty slice signals "all variables"
				if len(cmd.VariableFlags) == 0 {
					cmd.VariableFlags = []string{}
				}
			} else {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// Not a flag, keep in filtered args
			filteredArgs = append(filteredArgs, arg)
		}
	}

	return filteredArgs, nil
}

// isListCommand checks if a command is a list command alias
func isListCommand(command string) bool {
	listCommands := []string{"list", "ls", "exa", "lsd", "tree", "dir"}
	for _, cmd := range listCommands {
		if command == cmd {
			return true
		}
	}
	return false
}

// normalizeTarget converts target aliases to canonical names
func normalizeTarget(target string) string {
	switch strings.ToLower(target) {
	case "body":
		return "body"
	case "headers", "header":
		return "headers"
	case "query", "queries":
		return "query"
	case "request", "req", "url":
		return "request"
	case "variables", "vars", "var":
		return "variables"
	case "filters", "filter":
		return "filters"
	default:
		return target // Return as-is if not recognized (for special cases like "history", "response")
	}
}
