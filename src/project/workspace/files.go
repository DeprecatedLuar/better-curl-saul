package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/config"
)

// LoadPresetFile loads a specific TOML file from a preset
// Creates the file if it doesn't exist (lazy creation)
func LoadPresetFile(preset, fileType string) (*TomlHandler, error) {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return nil, err
	}

	// Ensure preset directory exists
	err = os.MkdirAll(presetPath, config.DirPermissions)
	if err != nil {
		return nil, fmt.Errorf(display.ErrDirectoryFailed)
	}

	filePath := filepath.Join(presetPath, fileType+".toml")

	// Create empty TOML file if it doesn't exist (lazy creation)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err := os.WriteFile(filePath, []byte(""), config.FilePermissions)
		if err != nil {
			return nil, fmt.Errorf(display.ErrFileSaveFailed, filePath)
		}
	}

	return NewTomlHandler(filePath)
}

// SavePresetFile saves a TOML handler to a specific preset file
func SavePresetFile(preset, fileType string, handler *TomlHandler) error {
	presetPath, err := GetPresetPath(preset)
	if err != nil {
		return err
	}

	filePath := filepath.Join(presetPath, fileType+".toml")
	handler.SetOutputPath(filePath)
	return handler.Write()
}

// ValidateFileType checks if the file type is valid
func ValidateFileType(fileType string) bool {
	validTypes := []string{"headers", "body", "query", "request", "variables", "filters"}
	for _, valid := range validTypes {
		if strings.ToLower(fileType) == valid {
			return true
		}
	}
	return false
}