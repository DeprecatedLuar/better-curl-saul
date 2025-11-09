package main

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/better-curl-saul/internal/commands"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// isActionCommand checks if a command is a preset action command
func isActionCommand(cmd string) bool {
	return cmd == "set" || cmd == "get" || cmd == "edit" || cmd == "call"
}

// isValidTarget checks if a string is a valid target name
func isValidTarget(s string) bool {
	validTargets := map[string]bool{
		"body": true, "headers": true, "query": true,
		"request": true, "variables": true, "history": true,
		"response": true, "filters": true,
	}
	return validTargets[s]
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		display.Info("Alright, alright! Let me break it down for you, folk:")
		display.Plain("saul [preset] [set/rm/edit...] [url/body...] [key=value]")
		display.Tip("That's the real cha-cha-cha. Use 'saul help' for the full legal brief")
		return
	}

	// Initialize session manager
	sessionManager, err := workspace.NewSessionManager()
	if err != nil {
		display.Error(fmt.Sprintf("failed to initialize session: %v", err))
		return
	}

	// Inject current preset for action commands
	if len(args) > 0 && isActionCommand(args[0]) {
		if sessionManager.HasCurrentPreset() {
			// Inject preset: ["set", "body"] -> ["pokeapi", "set", "body"]
			args = append([]string{sessionManager.GetCurrentPreset()}, args...)
		} else {
			// Error: action command but no current preset
			display.Error(display.ErrNoCurrentPreset)
			return
		}
	}

	// Special handling for rm command: dual-mode operation
	// If current preset exists AND args look like targets, inject preset
	// Otherwise, treat as global preset deletion
	if len(args) > 0 && args[0] == "rm" {
		if sessionManager.HasCurrentPreset() && len(args) > 1 {
			// Check if first argument after rm looks like a target name
			if isValidTarget(args[1]) {
				// Inject preset: ["rm", "body"] -> ["pokeapi", "rm", "body"]
				args = append([]string{sessionManager.GetCurrentPreset()}, args...)
			}
			// Otherwise, treat as global preset removal (no injection)
		}
	}

	cmd, err := parser.ParseCommandWithSession(args, sessionManager)
	if err != nil {
		display.Error(err.Error())
		return
	}

	err = commands.Execute(cmd, sessionManager)
	if err != nil {
		display.Error(err.Error())
		os.Exit(1)
	}
}
