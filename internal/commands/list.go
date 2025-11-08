package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/DeprecatedLuar/better-curl-saul/internal"
)

// List handles the list command - lists presets directory using system commands
// Supports: list, ls, exa, lsd, tree, dir (delegates to the requested tool)
func List(cmd Command) error {
	// Get the actual command to run (ls, exa, lsd, tree, dir)
	// Default to "ls" if somehow we get "list" as the command
	command := cmd.Preset
	if command == "list" {
		command = "ls"
	}

	// Validate that it's an allowed command
	if !isListCommand(command) {
		return fmt.Errorf("invalid list command: %s", command)
	}

	// Get presets directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	presetsDir := filepath.Join(home, internal.ParentDirPath, internal.AppDirName, internal.PresetsDirName)

	// Extract additional arguments from the original command line (skip "saul" and the command itself)
	args := os.Args[2:]

	// Create and execute the command in the presets directory
	execCmd := exec.Command(command, args...)
	execCmd.Dir = presetsDir
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}
