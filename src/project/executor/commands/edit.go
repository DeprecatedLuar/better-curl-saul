package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/executor"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// Edit handles both field-level and container-level editing
func Edit(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(errors.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return fmt.Errorf(errors.ErrTargetRequired)
	}

	// Normalize target aliases
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return fmt.Errorf(errors.ErrInvalidTarget, cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Distinguish between field-level and container-level editing
	if len(cmd.KeyValuePairs) == 0 || cmd.KeyValuePairs[0].Key == "" {
		// Container-level editing: edit the entire TOML file in editor
		return executeContainerEdit(cmd)
	} else {
		// Field-level editing: edit a specific field with readline
		return executeFieldEdit(cmd)
	}
}

// executeFieldEdit handles field-level editing with pre-filled prompts (existing functionality)
func executeFieldEdit(cmd parser.Command) error {
	// Use first key-value pair for field editing
	key := cmd.KeyValuePairs[0].Key
	
	// Load current value using existing patterns
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return fmt.Errorf(errors.ErrFileLoadFailed, cmd.Target+".toml")
	}

	// Get current value (empty string if doesn't exist)
	currentValue := handler.GetAsString(key)

	// Pre-filled interactive editing with readline
	rl, err := readline.New(fmt.Sprintf("%s: ", key))
	if err != nil {
		return fmt.Errorf(errors.ErrReadlineSetup)
	}
	defer rl.Close()

	// Pre-fill the prompt with current value
	rl.WriteStdin([]byte(currentValue))

	// Get new value from user
	newValue, err := rl.Readline()
	if err != nil {
		return fmt.Errorf(errors.ErrInputRead)
	}

	// Special validation for request fields
	if cmd.Target == "request" {
		if err := executor.ValidateRequestField(key, newValue); err != nil {
			return err
		}
	}

	// Save using existing validation and patterns
	valueToStore := newValue
	if cmd.Target == "request" && strings.ToLower(key) == "method" {
		// Store HTTP methods in uppercase
		valueToStore = strings.ToUpper(newValue)
	}
	inferredValue := executor.InferValueType(valueToStore)
	handler.Set(key, inferredValue)

	err = presets.SavePresetFile(cmd.Preset, cmd.Target, handler)
	if err != nil {
		return fmt.Errorf(errors.ErrFileSaveFailed, cmd.Target+".toml")
	}

	// Silent success - Unix philosophy
	return nil
}

// executeContainerEdit handles container-level editing (open file in editor)
func executeContainerEdit(cmd parser.Command) error {
	// Get the file path for the target
	presetPath, err := presets.GetPresetPath(cmd.Preset)
	if err != nil {
		return fmt.Errorf(errors.ErrDirectoryFailed)
	}

	filePath := filepath.Join(presetPath, cmd.Target+".toml")

	// Ensure the file exists (create empty file if it doesn't)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create empty TOML file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf(errors.ErrFileSaveFailed, cmd.Target+".toml")
		}
		file.Close()
	}

	// Detect and launch editor
	editor := detectEditor()
	if editor == "" {
		return fmt.Errorf(errors.ErrEditorNotFound)
	}

	// Launch editor with the file
	editorCmd := exec.Command(editor, filePath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	err = editorCmd.Run()
	if err != nil {
		return fmt.Errorf(errors.ErrEditorFailed, err)
	}

	// Silent success - Unix philosophy
	return nil
}

// detectEditor finds the best available editor
func detectEditor() string {
	// 1. Check $EDITOR environment variable first
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// 2. Fall back to common editors (in order of preference)
	editors := []string{"nano", "vim", "vi", "emacs"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return ""
}