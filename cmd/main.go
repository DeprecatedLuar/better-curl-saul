package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor/commands"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("\nOkay, so let me break it down to you buddy:\nsaul [preset] [set/rm/edit...] [url/body...] [key=value]\n ")
		return
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
		fmt.Println("The workspace-based HTTP client that makes curl simple")
		return nil

	case "list":
		presets, err := presets.ListPresets()
		if err != nil {
			return fmt.Errorf("failed to list presets: %v", err)
		}
		
		if cmd.RawOutput {
			// Raw mode: space-separated preset names (Unix style)
			if len(presets) > 0 {
				fmt.Println(strings.Join(presets, " "))
			}
			return nil
		}
		
		// Normal mode: formatted display
		if len(presets) == 0 {
			content := "Create one with: saul [preset-name]"
			formatted := display.FormatSimpleSection("No Presets Found", content)
			display.Plain(formatted)
			return nil
		}
		
		var content strings.Builder
		for _, preset := range presets {
			content.WriteString(fmt.Sprintf("  %s\n", preset))
		}
		
		formatted := display.FormatSimpleSection("Available Presets", strings.TrimSpace(content.String()))
		display.Plain(formatted)
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

	case "call":
		if cmd.Preset == "" {
			return fmt.Errorf("preset name required for call command")
		}
		return executor.ExecuteCallCommand(cmd)

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
  saul list                 List all presets
  saul rm [preset...]       Delete one or more presets
  saul call [preset]        Execute HTTP request
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
                            Get value from target file`
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
