package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// List handles the list command - lists presets directory using system commands
// Supports: list, ls, exa, lsd, tree, dir (delegates to the requested tool)
func List(cmd parser.Command) error {
	// Get the actual command to run (ls, exa, lsd, tree, dir)
	// Default to "ls" if somehow we get "list" as the command
	command := cmd.Preset
	if command == "list" {
		command = "ls"
	}

	// Validate that it's an allowed command
	if !parser.IsListCommand(command) {
		return fmt.Errorf(display.ErrInvalidListCommand, command)
	}

	// Get presets directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf(display.ErrHomeDirFailed, err)
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
