package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/config"
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
	if err := os.MkdirAll(filepath.Dir(sessionFile), config.DirPermissions); err != nil {
		return fmt.Errorf("failed to create session directory: %v", err)
	}

	return os.WriteFile(sessionFile, []byte(s.currentPreset), config.FilePermissions)
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

// getConfigPath returns the saul configuration directory path using centralized config
func getConfigPath() (string, error) {
	cfg := config.LoadConfig()
	base, err := config.GetConfigBase()
	if err != nil {
		return "", fmt.Errorf("failed to get config base: %v", err)
	}
	return filepath.Join(base, cfg.ConfigDirPath, cfg.AppDirName), nil
}