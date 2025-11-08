package workspace

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/better-curl-saul/internal/config"
	"github.com/DeprecatedLuar/better-curl-saul/internal/utils"
)

// readFile reads the TOML file from disk
func (t *TomlHandler) readFile() error {
	var err error
	t.raw, err = os.ReadFile(t.path)
	return err
}

// SetOutputPath sets the output file path for writing
func (t *TomlHandler) SetOutputPath(path string) {
	t.out = path
}

// Write saves the TOML data to file
// Uses output path if set, otherwise overwrites original file
func (t *TomlHandler) Write() error {
	path := t.out
	if path == "" {
		path = t.path
	}

	tomlString, err := t.tree.ToTomlString()
	if err != nil {
		return err
	}

	return utils.AtomicWriteFile(path, []byte(tomlString), config.FilePermissions)
}

// ToBytes returns the TOML data as bytes
func (t *TomlHandler) ToBytes() ([]byte, error) {
	tomlString, err := t.tree.ToTomlString()
	if err != nil {
		return nil, err
	}
	return []byte(tomlString), nil
}

// MergeTomlFiles merges multiple TOML files and returns JSON
// Perfect for combining method.toml + headers.toml + body.toml
func MergeTomlFiles(filePaths ...string) ([]byte, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// Load first file as base
	base, err := NewTomlHandler(filePaths[0])
	if err != nil {
		return nil, fmt.Errorf("failed to load base file %s: %v", filePaths[0], err)
	}

	// Merge remaining files
	for i := 1; i < len(filePaths); i++ {
		overlay, err := NewTomlHandler(filePaths[i])
		if err != nil {
			return nil, fmt.Errorf("failed to load file %s: %v", filePaths[i], err)
		}

		if err := base.Merge(overlay); err != nil {
			return nil, fmt.Errorf("failed to merge file %s: %v", filePaths[i], err)
		}
	}

	// Convert to JSON for HTTP requests
	return base.ToJSON()
}

// UpdateTomlValue updates a value in a TOML file using dot notation
func UpdateTomlValue(filePath, key string, value interface{}) error {
	handler, err := NewTomlHandler(filePath)
	if err != nil {
		return err
	}

	handler.Set(key, value)
	return handler.Write()
}