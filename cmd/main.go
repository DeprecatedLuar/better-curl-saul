package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/http"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/internal/utils"
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
		return nil, fmt.Errorf("failed to get TTY ID: %v", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %v", err)
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
		return fmt.Errorf("failed to create session directory: %v", err)
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

	cmd, err := commands.ParseCommand(args)
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

	// Update current preset when explicitly specified and save to session
	if cmd.Preset != "" {
		err := sessionManager.SetCurrentPreset(cmd.Preset)
		if err != nil {
			// Session save failure is not critical - log but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
		}
	}

	// Handle global commands
	if cmd.Global != "" {
		return executeGlobalCommand(cmd)
	}

	// Handle preset commands
	return executePresetCommand(cmd)
}

// executeGlobalCommand handles global commands like list, rm, version
func executeGlobalCommand(cmd commands.Command) error {
	switch cmd.Global {
	case "list":
		return commands.List(cmd)

	case "version":
		return commands.Version()

	case "rm":
		if len(cmd.Targets) == 0 {
			return fmt.Errorf("preset name required for rm command")
		}

		// Handle multiple targets: saul rm preset1 preset2 preset3
		// Continue processing, warn about non-existent presets
		var warnings []string
		deletedCount := 0

		for _, presetName := range cmd.Targets {
			err := workspace.DeletePreset(presetName)
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

	case "update":
		return commands.Update()

	default:
		return fmt.Errorf("unknown global command: %s", cmd.Global)
	}
}

// executePresetCommand handles preset-specific commands
func executePresetCommand(cmd commands.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required")
	}

	// If no command specified, create the preset if it doesn't exist
	if cmd.Command == "" {
		err := workspace.CreatePresetDirectory(cmd.Preset)
		if err != nil {
			return fmt.Errorf("failed to create preset '%s': %v", cmd.Preset, err)
		}
		// Silent success - Unix philosophy
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
		err = http.ExecuteCallCommand(cmd)

	default:
		return fmt.Errorf("unknown preset command: %s", cmd.Command)
	}

	// If main command succeeded and --call flag is set, execute call
	if err == nil && cmd.Call {
		return http.ExecuteCallCommand(cmd)
	}

	return err
}

// showHelp displays usage information
func showHelp() {
	display.Info("Better-Curl (Saul) - Workspace-based HTTP Client")
	display.Plain("")

	// Usage section
	usage := "  saul [preset] [command] [target] [key=value]"
	formatted := display.FormatSimpleSection("Usage", usage)
	display.Plain(formatted)

	// Global Commands section
	globalCmds := `  saul version              Show version information
  saul update               Check for updates
  saul ls [options]         List presets directory (system ls command)
  saul rm [preset...]       Delete one or more presets
  saul help                 Show this help`
	formatted = display.FormatSimpleSection("Global Commands", globalCmds)
	display.Plain(formatted)

	// Preset Commands section
	presetCmds := `  saul [preset]             Create or switch to preset
  saul [preset] set [target] [key=value]
                            Set value in target file
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
  saul pokeapi get url
  saul pokeapi get body pokemon.name`
	formatted = display.FormatSimpleSection("Examples", examples)
	display.Plain(formatted)
}
