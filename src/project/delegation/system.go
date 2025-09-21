package delegation

import (
	"os"
	"os/exec"
	"path/filepath"
)

// allowedCommands defines the whitelist of safe system commands for delegation
var allowedCommands = []string{"ls", "exa", "lsd", "tree", "dir"}

// IsSystemCommand checks if a command is in the allowed system commands whitelist
func IsSystemCommand(command string) bool {
	for _, allowed := range allowedCommands {
		if command == allowed {
			return true
		}
	}
	return false
}

// DelegateToSystem executes a system command in the presets directory
func DelegateToSystem(command string, args []string) error {
	// Set working directory to presets folder
	presetsDir := filepath.Join(os.Getenv("HOME"), ".config", "saul", "presets")

	// Create the command with arguments
	execCmd := exec.Command(command, args...)
	execCmd.Dir = presetsDir
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Execute the command
	return execCmd.Run()
}