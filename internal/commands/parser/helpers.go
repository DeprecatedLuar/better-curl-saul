// Package parser provides command parsing and routing logic for Better-Curl-Saul.
// This package handles detecting command types, extracting arguments, and building
// Command structs for execution by the command handlers.
package parser

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// IsListCommand checks if a command is a list command alias
func IsListCommand(command string) bool {
	listCommands := []string{"list", "ls", "exa", "lsd", "tree", "dir"}
	for _, cmd := range listCommands {
		if command == cmd {
			return true
		}
	}
	return false
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

// isExplicitCommand checks if a string is an explicit command (set, get, edit, call)
func isExplicitCommand(cmd string) bool {
	explicitCommands := []string{"set", "get", "edit", "call", "check"}
	cmdLower := strings.ToLower(cmd)
	for _, explicit := range explicitCommands {
		if cmdLower == explicit {
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
