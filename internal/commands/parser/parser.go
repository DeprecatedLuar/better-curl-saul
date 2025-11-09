package parser

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
	Target string // Target file: "body", "headers", "query", "request" (optional, for HTTPie syntax)
	Key    string
	Value  string
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
	if IsListCommand(args[0]) {
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
	case "copy", "cp":
		cmd.Global = "copy"
		if len(args) >= 3 {
			// Handle copy command: saul cp source dest
			cmd.Targets = args[1:]
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

	// Check for HTTPie syntax: preset followed by non-explicit command that looks like HTTPie
	if len(args) > 1 && !isExplicitCommand(args[1]) {
		// Check if it looks like HTTPie syntax (URL, method, or HTTPie arg)
		if LooksLikeURL(args[1]) || IsHTTPMethod(args[1]) || IsHTTPieArg(args[1]) {
			return ParseHTTPieSyntax(args[1:], cmd.Preset)
		}
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
