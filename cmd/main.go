package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands"
	"github.com/DeprecatedLuar/better-curl-saul/internal/utils"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// SessionManager encapsulates session state and file operations
type SessionManager struct {
	currentPreset string
	ttyID         string
	configPath    string
}

// NewSessionManager creates a new session manager with TTY-based session isolation
func NewSessionManager() (*SessionManager, error) {
	ttyID, err := getTTYID()
	if err != nil {
		return nil, fmt.Errorf(display.ErrSessionTTYFailed, err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf(display.ErrSessionConfigFailed, err)
	}

	sm := &SessionManager{
		ttyID:      ttyID,
		configPath: configPath,
	}

	// Load existing session
	if err := sm.LoadSession(); err != nil {
		// Session load failure is not critical - continue with empty session
		sm.currentPreset = ""
	}

	return sm, nil
}

// GetCurrentPreset returns the current preset for this session
func (s *SessionManager) GetCurrentPreset() string {
	return s.currentPreset
}

// SetCurrentPreset sets the current preset and saves the session
func (s *SessionManager) SetCurrentPreset(preset string) error {
	s.currentPreset = preset
	return s.SaveSession()
}

// LoadSession loads the session from the TTY-specific session file
func (s *SessionManager) LoadSession() error {
	sessionFile := s.getSessionFilePath()

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		// Session file doesn't exist - not an error
		s.currentPreset = ""
		return nil
	}

	s.currentPreset = strings.TrimSpace(string(data))
	return nil
}

// SaveSession saves the current session to the TTY-specific session file
func (s *SessionManager) SaveSession() error {
	sessionFile := s.getSessionFilePath()

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(sessionFile), internal.DirPermissions); err != nil {
		return fmt.Errorf(display.ErrSessionDirFailed, err)
	}

	return utils.AtomicWriteFile(sessionFile, []byte(s.currentPreset), internal.FilePermissions)
}

// HasCurrentPreset returns true if a current preset is set
func (s *SessionManager) HasCurrentPreset() bool {
	return s.currentPreset != ""
}

// getSessionFilePath returns the TTY-specific session file path
func (s *SessionManager) getSessionFilePath() string {
	return filepath.Join(s.configPath, fmt.Sprintf(".session_%s", s.ttyID))
}

// getTTYID gets the current TTY identifier for session isolation
func getTTYID() (string, error) {
	tty := os.Getenv("TTY")
	if tty == "" {
		// Fallback to simpler TTY detection
		tty = "default"
	} else {
		// Extract just the TTY number/name for filename safety
		tty = filepath.Base(tty)
		// Replace any unsafe characters
		tty = strings.ReplaceAll(tty, "/", "_")
	}
	return tty, nil
}

// getConfigPath returns the saul configuration directory path
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, internal.ParentDirPath, internal.AppDirName), nil
}

// isActionCommand checks if a command is a preset action command
func isActionCommand(cmd string) bool {
	return cmd == "set" || cmd == "get" || cmd == "edit" || cmd == "call"
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
	sessionManager, err := NewSessionManager()
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

	cmd, err := commands.ParseCommandWithSession(args, sessionManager)
	if err != nil {
		display.Error(err.Error())
		return
	}

	err = executeCommand(cmd, sessionManager)
	if err != nil {
		display.Error(err.Error())
		os.Exit(1)
	}
}

// executeCommand routes commands to appropriate handlers
func executeCommand(cmd commands.Command, sessionManager *SessionManager) error {

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
				if err := ensureVariantStructure(cmd.Preset); err != nil {
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
func executeVariantSwitch(cmd commands.Command, sessionManager *SessionManager) error {
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

	// Ensure variants folder and variant directory exist
	fullPresetPath := basePreset + "/" + variantName
	if err := ensureVariantStructure(fullPresetPath); err != nil {
		return err
	}

	// Update .config file
	if err := workspace.SetActiveVariant(basePreset, variantName); err != nil {
		return err
	}

	// Update session
	if err := sessionManager.SetCurrentPreset(fullPresetPath); err != nil {
		display.Warning(fmt.Sprintf(display.WarnSessionSaveFailed, err))
	}

	display.Success(fmt.Sprintf("Switched to variant: %s", variantName))
	return nil
}

// ensureVariantStructure creates variants/ folder and variant directory if preset contains /
// On first variant creation, migrates root TOML files into the variant
func ensureVariantStructure(presetWithVariant string) error {
	parts := strings.Split(presetWithVariant, "/")
	if len(parts) != 2 {
		return fmt.Errorf(display.ErrInvalidVariantPath, presetWithVariant)
	}

	basePreset := parts[0]
	variantName := parts[1]

	// Ensure base preset exists
	if !workspace.PresetExists(basePreset) {
		return fmt.Errorf(display.ErrVariantPresetMissing, basePreset)
	}

	presetPath, err := workspace.GetPresetPath(basePreset)
	if err != nil {
		return err
	}

	variantsDir := filepath.Join(presetPath, "variants")
	isFirstVariant := false

	// Check if this is first variant creation
	if _, err := os.Stat(variantsDir); os.IsNotExist(err) {
		isFirstVariant = true
	}

	// Create variants/ folder
	if err := os.MkdirAll(variantsDir, internal.DirPermissions); err != nil {
		return fmt.Errorf(display.ErrVariantsDirFailed, err)
	}

	// Create variant directory
	variantPath := filepath.Join(variantsDir, variantName)
	if err := os.MkdirAll(variantPath, internal.DirPermissions); err != nil {
		return fmt.Errorf(display.ErrVariantDirFailed, err)
	}

	// Migrate root TOML files to first variant
	if isFirstVariant {
		tomlFiles := []string{"request.toml", "body.toml", "headers.toml", "query.toml", "variables.toml", "filters.toml"}
		for _, file := range tomlFiles {
			rootFile := filepath.Join(presetPath, file)
			if _, err := os.Stat(rootFile); err == nil {
				// File exists, move it
				variantFile := filepath.Join(variantPath, file)
				if err := os.Rename(rootFile, variantFile); err != nil {
					return fmt.Errorf(display.ErrVariantMigrateFailed, file, err)
				}
			}
		}
	}

	// Create/update .config file to point to this variant
	configPath := filepath.Join(presetPath, ".config")
	if err := os.WriteFile(configPath, []byte(variantName), internal.FilePermissions); err != nil {
		return fmt.Errorf(display.ErrVariantConfigFailed, err)
	}

	return nil
}

// executeGlobalCommand handles global commands like list, rm, version
func executeGlobalCommand(cmd commands.Command) error {
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
		return commands.List(cmd)

	case "version":
		return commands.Version()

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
		return commands.Help()

	case "update":
		return commands.Update()

	default:
		return fmt.Errorf(display.ErrCommandUnknownGlobal, cmd.Global)
	}
}

// executePresetCommand handles preset-specific commands
func executePresetCommand(cmd commands.Command) error {
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
				if err := ensureVariantStructure(cmd.Preset); err != nil {
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
	case "set":
		err = commands.Set(cmd)

	case "get":
		err = commands.Get(cmd)

	case "edit":
		err = commands.Edit(cmd)

	case "call":
		err = commands.Call(cmd)

	default:
		return fmt.Errorf(display.ErrCommandUnknownPreset, cmd.Command)
	}

	// If main command succeeded and --call flag is set, execute call
	if err == nil && cmd.Call {
		return commands.Call(cmd)
	}

	return err
}
