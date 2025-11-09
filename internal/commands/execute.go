package commands

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Execute routes and executes the parsed command
func Execute(cmd parser.Command, sessionManager *workspace.SessionManager) error {
	// Handle variant switching first (before updating session)
	if cmd.Global == "switch" || cmd.Global == "switch-variant" {
		return executeVariantSwitch(cmd, sessionManager)
	}

	// Update current preset when explicitly specified and save to session
	if cmd.Preset != "" {
		// Handle variant creation if preset contains / AND base preset exists
		if strings.Contains(cmd.Preset, "/") {
			basePreset := strings.Split(cmd.Preset, "/")[0]
			// Only ensure variant structure if base preset already exists
			// If it doesn't exist, executePresetCommand will handle creation
			if workspace.PresetExists(basePreset) {
				if err := workspace.EnsureVariantStructure(cmd.Preset); err != nil {
					return err
				}
			}
		}

		err := sessionManager.SetCurrentPreset(cmd.Preset)
		if err != nil {
			// Session save failure is not critical - log but continue
			display.Warning(fmt.Sprintf(display.WarnSessionSaveFailed, err))
		}
	}

	// Handle global commands
	if cmd.Global != "" {
		return executeGlobalCommand(cmd)
	}

	// Handle preset commands
	return executePresetCommand(cmd)
}

// executeVariantSwitch handles variant switching
func executeVariantSwitch(cmd parser.Command, sessionManager *workspace.SessionManager) error {
	if !sessionManager.HasCurrentPreset() {
		return fmt.Errorf(display.ErrNoActivePreset)
	}

	currentPreset := sessionManager.GetCurrentPreset()
	basePreset := strings.Split(currentPreset, "/")[0]

	// Extract variant name from cmd.Preset
	// For "switch" command: cmd.Preset is just variant name
	// For "switch-variant" (from /variant): cmd.Preset is full path
	variantName := cmd.Preset
	if strings.Contains(variantName, "/") {
		variantName = strings.Split(variantName, "/")[1]
	}

	return workspace.SwitchVariant(basePreset, variantName, sessionManager)
}

// executeGlobalCommand handles global commands like list, rm, version
func executeGlobalCommand(cmd parser.Command) error {
	switch cmd.Global {
	case "create":
		if cmd.Preset == "" {
			return fmt.Errorf(display.ErrPresetNameRequired)
		}
		err := workspace.CreatePresetDirectory(cmd.Preset)
		if err != nil {
			return fmt.Errorf(display.ErrPresetCreateFailed, cmd.Preset, err)
		}
		return nil

	case "list":
		return List(cmd)

	case "version":
		return Version()

	case "copy":
		return Copy(cmd)

	case "rm":
		if len(cmd.Targets) == 0 {
			return fmt.Errorf(display.ErrPresetNameRequired)
		}

		// Handle multiple targets: saul rm preset1 preset2 preset3
		// Continue processing, warn about non-existent presets
		deletedCount := 0

		for _, presetName := range cmd.Targets {
			err := workspace.DeletePreset(presetName)
			if err != nil {
				// Collect warnings for non-existent presets, continue processing
				display.Warning(fmt.Sprintf(display.ErrPresetNotFound, presetName))
			} else {
				deletedCount++
			}
		}

		// Silent success if at least one was deleted, or no warnings
		return nil

	case "help":
		return Help()

	case "update":
		return Update()

	default:
		return fmt.Errorf(display.ErrCommandUnknownGlobal, cmd.Global)
	}
}

// executeHTTPieCommand handles HTTPie-style multi-target operations
func executeHTTPieCommand(cmd parser.Command) error {
	// Group KeyValuePairs by Target
	targetGroups := make(map[string][]parser.KeyValuePair)
	for _, kvp := range cmd.KeyValuePairs {
		targetGroups[kvp.Target] = append(targetGroups[kvp.Target], kvp)
	}

	// Execute Set() for each target group
	for target, pairs := range targetGroups {
		setCmd := parser.Command{
			Preset:        cmd.Preset,
			Command:       "set",
			Target:        target,
			KeyValuePairs: pairs,
		}
		if err := Set(setCmd); err != nil {
			return err
		}
	}

	return nil
}

// executePresetCommand handles preset-specific commands
func executePresetCommand(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	// Handle variant paths: check base preset exists, not full variant path
	basePreset := cmd.Preset
	isVariant := strings.Contains(cmd.Preset, "/")
	if isVariant {
		basePreset = strings.Split(cmd.Preset, "/")[0]
	}

	// Check if base preset exists
	presetExists := workspace.PresetExists(basePreset)

	// Handle preset creation requirements
	if !presetExists {
		if cmd.Create {
			// Create base preset explicitly requested via --create flag
			err := workspace.CreatePresetDirectory(basePreset)
			if err != nil {
				return fmt.Errorf(display.ErrPresetCreateFailed, basePreset, err)
			}
			// If it's a variant, ensure variant structure
			if isVariant {
				if err := workspace.EnsureVariantStructure(cmd.Preset); err != nil {
					return err
				}
			}
			// If no command specified, we're done
			if cmd.Command == "" {
				return nil
			}
			// Otherwise, continue to execute the command
		} else {
			// Preset doesn't exist and no --create flag
			return fmt.Errorf(display.ErrPresetNotFoundCreate, basePreset, basePreset, basePreset)
		}
	}

	// If preset exists and no command specified, just switch to it (silent success)
	if cmd.Command == "" {
		return nil
	}

	// Route preset commands
	var err error
	switch cmd.Command {
	case "httpie":
		err = executeHTTPieCommand(cmd)

	case "set":
		err = Set(cmd)

	case "get":
		err = Get(cmd)

	case "edit":
		err = Edit(cmd)

	case "call":
		err = Call(cmd)

	default:
		return fmt.Errorf(display.ErrCommandUnknownPreset, cmd.Command)
	}

	// If main command succeeded and --call flag is set, execute call
	if err == nil && cmd.Call {
		return Call(cmd)
	}

	return err
}
