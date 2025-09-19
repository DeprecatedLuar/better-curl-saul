package main

import (
	"fmt"
	"os"

	"main/src/project/executor"
	"main/src/project/executor/commands"
	"main/src/project/parser"
	"main/src/project/presets"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("\nOkay, so let me break it down to you buddy:\nsaul [preset] [set/rm/edit...] [url/body...] [key=value]\n ")
		return
	}

	cmd, err := parser.ParseCommand(args)
	if err != nil {
		fmt.Printf("Oopsies: %v\n", err)
		return
	}

	err = executeCommand(cmd)
	if err != nil {
		fmt.Printf("Command failed: %v\n", err)
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
		if len(presets) == 0 {
			fmt.Println("No presets found. Create one with: saul [preset-name]")
			return nil
		}
		fmt.Println("Available presets:")
		for _, preset := range presets {
			fmt.Printf("  %s\n", preset)
		}
		return nil

	case "rm":
		if cmd.Preset == "" {
			return fmt.Errorf("preset name required for rm command")
		}
		err := presets.DeletePreset(cmd.Preset)
		if err != nil {
			return fmt.Errorf("failed to delete preset '%s': %v", cmd.Preset, err)
		}
		// Silent success
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
	fmt.Println("USAGE:")
	fmt.Println("  saul [preset] [command] [target] [key=value]")
	fmt.Println()
	fmt.Println("GLOBAL COMMANDS:")
	fmt.Println("  saul version              Show version information")
	fmt.Println("  saul list                 List all presets")
	fmt.Println("  saul rm [preset]          Delete a preset")
	fmt.Println("  saul call [preset]        Execute HTTP request")
	fmt.Println("  saul help                 Show this help")
	fmt.Println()
	fmt.Println("PRESET COMMANDS:")
	fmt.Println("  saul [preset]             Create or switch to preset")
	fmt.Println("  saul [preset] set [target] [key=value]")
	fmt.Println("                            Set value in target file")
	fmt.Println("  saul [preset] check [target] [key]")
	fmt.Println("                            Display target contents (clean format)")
	fmt.Println("  saul [preset] get [target] [key]")
	fmt.Println("                            Get value from target file")
	fmt.Println()
	fmt.Println("TARGETS:")
	fmt.Println("  body      HTTP request body (JSON)")
	fmt.Println("  headers   HTTP headers")
	fmt.Println("  query     Query/search payload data")
	fmt.Println("  request   HTTP method, URL, and settings")
	fmt.Println("  variables Hard variables only (soft variables never stored)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Special request syntax (no = sign)")
	fmt.Println("  saul pokeapi set url https://api.example.com")
	fmt.Println("  saul pokeapi set method POST")
	fmt.Println("  saul pokeapi set timeout 30")
	fmt.Println()
	fmt.Println("  # Regular TOML syntax (with = sign)")
	fmt.Println("  saul pokeapi set body pokemon.name=pikachu")
	fmt.Println("  saul pokeapi set header Content-Type=application/json")
	fmt.Println("  saul pokeapi set body pokemon.level=@level")
	fmt.Println()
	fmt.Println("  # Check what's configured")
	fmt.Println("  saul pokeapi check url")
	fmt.Println("  saul pokeapi check body pokemon.name")
}
