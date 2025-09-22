package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/delegation"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor/commands"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// Session state: current preset memory (session-only)
var currentPreset string

// isActionCommand checks if a command is a preset action command
func isActionCommand(cmd string) bool {
	return cmd == "set" || cmd == "get" || cmd == "check" || cmd == "edit" || cmd == "call"
}

// getTTY returns a sanitized terminal device identifier for session files
func getTTY() string {
	// Try TTY environment variable first
	if tty := os.Getenv("TTY"); tty != "" {
		// Clean up path: /dev/pts/0 -> pts_0
		cleaned := strings.TrimPrefix(tty, "/dev/")
		return strings.ReplaceAll(cleaned, "/", "_")
	}

	// Fallback: read from stdin's tty
	if file, err := os.Open("/proc/self/fd/0"); err == nil {
		defer file.Close()
		if link, err := os.Readlink("/proc/self/fd/0"); err == nil {
			if strings.HasPrefix(link, "/dev/") {
				cleaned := strings.TrimPrefix(link, "/dev/")
				return strings.ReplaceAll(cleaned, "/", "_")
			}
		}
	}

	// Final fallback: use a generic identifier
	return "console"
}

// loadCurrentPreset loads the current preset from terminal session file
func loadCurrentPreset() {
	tty := getTTY()
	sessionFile := filepath.Join(os.Getenv("HOME"), ".config", "saul", fmt.Sprintf(".session_%s", tty))
	if data, err := os.ReadFile(sessionFile); err == nil {
		currentPreset = strings.TrimSpace(string(data))
	}
}

// saveCurrentPreset saves the current preset to terminal session file
func saveCurrentPreset() {
	if currentPreset == "" {
		return
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "saul")
	os.MkdirAll(configDir, 0755)

	tty := getTTY()
	sessionFile := filepath.Join(configDir, fmt.Sprintf(".session_%s", tty))
	os.WriteFile(sessionFile, []byte(currentPreset), 0644)
}

// cleanupStaleSessionFiles removes session files for terminals that no longer exist
func cleanupStaleSessionFiles() {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "saul")

	// Read all session files
	files, err := filepath.Glob(filepath.Join(configDir, ".session_*"))
	if err != nil {
		return // Silent failure - cleanup is optional
	}

	for _, sessionFile := range files {
		// Extract TTY from filename: .session_pts_0 -> pts_0
		basename := filepath.Base(sessionFile)
		if strings.HasPrefix(basename, ".session_") {
			ttyName := strings.TrimPrefix(basename, ".session_")

			// Check if TTY device exists
			if ttyName != "console" { // Don't cleanup console sessions
				devicePath := "/dev/" + strings.ReplaceAll(ttyName, "_", "/")
				if _, err := os.Stat(devicePath); os.IsNotExist(err) {
					// TTY no longer exists, remove stale session
					os.Remove(sessionFile)
				}
			}
		}
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("\nAlright, alright! Let me break it down for you, folk:\nsaul [preset] [set/rm/edit...] [url/body...] [key=value]\n\nThat's the real cha-cha-cha. Use 'saul help' for the full legal brief")
		return
	}

	// Clean up stale session files from closed terminals
	cleanupStaleSessionFiles()

	// Load current preset from terminal session
	loadCurrentPreset()

	// Inject current preset for action commands
	if len(args) > 0 && isActionCommand(args[0]) {
		if currentPreset != "" {
			// Inject preset: ["set", "body"] -> ["pokeapi", "set", "body"]
			args = append([]string{currentPreset}, args...)
		} else {
			// Error: action command but no current preset
			display.Error(errors.ErrNoCurrentPreset)
			return
		}
	}

	cmd, err := parser.ParseCommand(args)
	if err != nil {
		display.Error(err.Error())
		return
	}

	err = executeCommand(cmd)
	if err != nil {
		display.Error(err.Error())
		os.Exit(1)
	}
}

// executeCommand routes commands to appropriate handlers
func executeCommand(cmd parser.Command) error {
	// Check for system command delegation first
	if delegation.IsSystemCommand(cmd.Preset) {
		// Extract arguments from the original command line
		args := os.Args[2:] // Skip "saul" and the system command
		return delegation.DelegateToSystem(cmd.Preset, args)
	}

	// Update current preset when explicitly specified and save to session
	if cmd.Preset != "" {
		currentPreset = cmd.Preset
		saveCurrentPreset()
	}

	// Handle global commands
	if cmd.Global != "" {
		return executeGlobalCommand(cmd)
	}

	// Handle preset commands
	return executePresetCommand(cmd)
}

// executeGlobalCommand handles global commands like list, rm, version
func executeGlobalCommand(cmd parser.Command) error {
	switch cmd.Global {
	case "version":
		fmt.Println("Better-Curl (Saul) v0.1.0")
		fmt.Println("'When http gets complicated, Better Curl Saul'")
		return nil

	case "rm":
		if len(cmd.Targets) == 0 {
			return fmt.Errorf("preset name required for rm command")
		}

		// Handle multiple targets: saul rm preset1 preset2 preset3
		// Continue processing, warn about non-existent presets
		var warnings []string
		deletedCount := 0

		for _, presetName := range cmd.Targets {
			err := presets.DeletePreset(presetName)
			if err != nil {
				// Collect warnings for non-existent presets, continue processing
				warnings = append(warnings, fmt.Sprintf("Warning: preset '%s' does not exist", presetName))
			} else {
				deletedCount++
			}
		}

		// Print warnings if any
		for _, warning := range warnings {
			display.Warning(warning)
		}

		// Silent success if at least one was deleted, or no warnings
		return nil

	case "help":
		showHelp()
		return nil

	default:
		return fmt.Errorf("unknown global command: %s", cmd.Global)
	}
}

// executePresetCommand handles preset-specific commands
func executePresetCommand(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required")
	}

	// If no command specified, create the preset if it doesn't exist
	if cmd.Command == "" {
		err := presets.CreatePresetDirectory(cmd.Preset)
		if err != nil {
			return fmt.Errorf("failed to create preset '%s': %v", cmd.Preset, err)
		}
		// Silent success - Unix philosophy
		return nil
	}

	// Route preset commands
	switch cmd.Command {
	case "set":
		return commands.Set(cmd)

	case "get":
		value, err := commands.Get(cmd)
		if err != nil {
			return err
		}
		fmt.Printf("Value: %v\n", value)
		return nil

	case "check":
		return commands.Check(cmd)

	case "edit":
		return commands.Edit(cmd)

	case "call":
		return executor.ExecuteCallCommand(cmd)

	default:
		return fmt.Errorf("unknown preset command: %s", cmd.Command)
	}
}

// showHelp displays usage information
func showHelp() {
	fmt.Println("Better-Curl (Saul) - Workspace-based HTTP Client")
	fmt.Println()

	// Usage section
	usage := "  saul [preset] [command] [target] [key=value]"
	formatted := display.FormatSimpleSection("Usage", usage)
	display.Plain(formatted)

	// Global Commands section
	globalCmds := `  saul version              Show version information
  saul ls [options]         List presets directory (system ls command)
  saul rm [preset...]       Delete one or more presets
  saul help                 Show this help`
	formatted = display.FormatSimpleSection("Global Commands", globalCmds)
	display.Plain(formatted)

	// Preset Commands section
	presetCmds := `  saul [preset]             Create or switch to preset
  saul [preset] set [target] [key=value]
                            Set value in target file
  saul [preset] check [target] [key]
                            Display target contents (clean format)
  saul [preset] get [target] [key]
                            Get value from target file
  saul [preset] call        Execute HTTP request
  saul call                 Execute HTTP request (current preset)`
	formatted = display.FormatSimpleSection("Preset Commands", presetCmds)
	display.Plain(formatted)

	// Targets section
	targets := `  body      HTTP request body (JSON)
  headers   HTTP headers
  query     Query/search payload data
  request   HTTP method, URL, and settings
  variables Hard variables only (soft variables never stored)`
	formatted = display.FormatSimpleSection("Targets", targets)
	display.Plain(formatted)

	// Examples section
	examples := `  # Special request syntax (no = sign)
  saul pokeapi set url https://api.example.com
  saul pokeapi set method POST
  saul pokeapi set timeout 30

  # Regular TOML syntax (with = sign)
  saul pokeapi set body pokemon.name=pikachu
  saul pokeapi set header Content-Type=application/json
  saul pokeapi set body pokemon.level=@level

  # Check what's configured
  saul pokeapi check url
  saul pokeapi check body pokemon.name`
	formatted = display.FormatSimpleSection("Examples", examples)
	display.Plain(formatted)
}
