// Package workspace provides workspace and TOML file management for Better-Curl-Saul.
// This package handles TOML parsing, modification, JSON conversion, preset directories,
// and history management for the 5-file structure (body, headers, query, request, variables).
package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lib "github.com/pelletier/go-toml"
	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/utils"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// ===== TOML HANDLER TYPE =====

// TomlHandler provides TOML file manipulation capabilities
type TomlHandler struct {
	path string
	out  string
	raw  []byte
	tree *lib.Tree
}

// ===== CONSTRUCTORS =====

// NewTomlHandler creates a new TOML handler from file path
func NewTomlHandler(path string) (*TomlHandler, error) {
	handler := &TomlHandler{path: path}

	if err := handler.readFile(); err != nil {
		return nil, err
	}

	if err := handler.load(); err != nil {
		return nil, err
	}

	return handler, nil
}

// NewTomlHandlerFromBytes creates a TOML handler from raw bytes
func NewTomlHandlerFromBytes(data []byte) (*TomlHandler, error) {
	handler := &TomlHandler{raw: data}

	if err := handler.load(); err != nil {
		return nil, err
	}

	return handler, nil
}

// ===== FILE I/O METHODS =====

// readFile reads the TOML file from disk
func (t *TomlHandler) readFile() error {
	var err error
	t.raw, err = os.ReadFile(t.path)
	return err
}

// load parses the raw TOML data into a tree structure
func (t *TomlHandler) load() error {
	var err error
	t.tree, err = lib.LoadBytes(t.raw)
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

	return utils.AtomicWriteFile(path, []byte(tomlString), internal.FilePermissions)
}

// ToBytes returns the TOML data as bytes
func (t *TomlHandler) ToBytes() ([]byte, error) {
	tomlString, err := t.tree.ToTomlString()
	if err != nil {
		return nil, err
	}
	return []byte(tomlString), nil
}

// ===== DATA STRUCTURE METHODS =====

// Get retrieves a value using dot notation (e.g., "server.port", "database.host")
func (t *TomlHandler) Get(query string) interface{} {
	return t.tree.Get(query)
}

// Set updates a value using dot notation
// Creates the key if it doesn't exist
func (t *TomlHandler) Set(query string, data interface{}) {
	t.tree.Set(query, data)
}

// Has checks if a key exists using dot notation
func (t *TomlHandler) Has(query string) bool {
	return t.tree.Has(query)
}

// Delete removes a key using dot notation
func (t *TomlHandler) Delete(query string) error {
	if !t.tree.Has(query) {
		return fmt.Errorf("key %s does not exist", query)
	}
	t.tree.Delete(query)
	return nil
}

// Keys returns all top-level keys in the TOML
func (t *TomlHandler) Keys() []string {
	return t.tree.Keys()
}

// ===== TYPE CONVERSION METHODS =====

// GetAsString gets a value and converts it to string
func (t *TomlHandler) GetAsString(query string) string {
	val := t.Get(query)
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}

// GetAsInt gets a value and converts it to int64
func (t *TomlHandler) GetAsInt(query string) (int64, error) {
	val := t.Get(query)
	if val == nil {
		return 0, fmt.Errorf("key %s not found", query)
	}

	switch v := val.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", val)
	}
}

// ToJSON converts the TOML data to JSON
func (t *TomlHandler) ToJSON() ([]byte, error) {
	goMap := t.tree.ToMap()
	return json.Marshal(goMap)
}

// ===== MERGE OPERATIONS =====

// Merge combines another TOML handler into this one
// The other handler's values will override this handler's values for conflicts
// Nested objects are merged recursively, arrays are replaced entirely
func (t *TomlHandler) Merge(other *TomlHandler) error {
	return t.mergeTree(t.tree, other.tree)
}

// mergeTree recursively merges source tree into target tree
func (t *TomlHandler) mergeTree(target, source *lib.Tree) error {
	for _, key := range source.Keys() {
		sourceValue := source.Get(key)

		if target.Has(key) {
			targetValue := target.Get(key)

			// If both are trees (nested objects), merge recursively
			if sourceTree, ok := sourceValue.(*lib.Tree); ok {
				if targetTree, ok := targetValue.(*lib.Tree); ok {
					if err := t.mergeTree(targetTree, sourceTree); err != nil {
						return err
					}
					continue
				}
			}
		}

		// For all other cases (primitives, arrays, or new keys), overwrite
		target.Set(key, sourceValue)
	}
	return nil
}

// MergeMultiple merges multiple TOML handlers into this one
// Later handlers override earlier ones for conflicts
func (t *TomlHandler) MergeMultiple(others ...*TomlHandler) error {
	for _, other := range others {
		if err := t.Merge(other); err != nil {
			return err
		}
	}
	return nil
}

// Clone creates a copy of the TOML handler
func (t *TomlHandler) Clone() (*TomlHandler, error) {
	data, err := t.ToBytes()
	if err != nil {
		return nil, err
	}
	return NewTomlHandlerFromBytes(data)
}

// ===== PRESET FILE OPERATIONS =====

// LoadPresetFile loads a specific TOML file from a preset
// If file doesn't exist, returns an empty handler that will create the file on first Write()
// Supports variants: handles preset paths like "myapi/submit" or "myapi"
func LoadPresetFile(preset, fileType string) (*TomlHandler, error) {
	// Extract base preset if variant path provided
	basePreset := preset
	if strings.Contains(preset, "/") {
		basePreset = strings.Split(preset, "/")[0]
	}

	presetPath, err := GetPresetPath(basePreset)
	if err != nil {
		return nil, err
	}

	// Ensure preset directory exists
	err = os.MkdirAll(presetPath, internal.DirPermissions)
	if err != nil {
		return nil, fmt.Errorf(display.ErrDirectoryFailed)
	}

	variantsDir := filepath.Join(presetPath, "variants")
	var filePath string

	// Check if variants folder exists
	if _, err := os.Stat(variantsDir); err == nil {
		activeVariant := GetActiveVariant(basePreset)
		variantPath := filepath.Join(variantsDir, activeVariant)

		// Ensure variant directory exists
		err = os.MkdirAll(variantPath, internal.DirPermissions)
		if err != nil {
			return nil, fmt.Errorf(display.ErrDirectoryFailed)
		}

		filePath = filepath.Join(variantPath, fileType+".toml")
	} else {
		// Fallback to root files (backward compatible)
		filePath = filepath.Join(presetPath, fileType+".toml")
	}

	// If file doesn't exist, create an empty handler (file created on Write())
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		handler := &TomlHandler{path: filePath, raw: []byte("")}
		if err := handler.load(); err != nil {
			return nil, err
		}
		return handler, nil
	}

	return NewTomlHandler(filePath)
}

// SavePresetFile saves a TOML handler to a specific preset file
// Supports variants: handles preset paths like "myapi/submit" or "myapi"
func SavePresetFile(preset, fileType string, handler *TomlHandler) error {
	// Extract base preset if variant path provided
	basePreset := preset
	if strings.Contains(preset, "/") {
		basePreset = strings.Split(preset, "/")[0]
	}

	presetPath, err := GetPresetPath(basePreset)
	if err != nil {
		return err
	}

	variantsDir := filepath.Join(presetPath, "variants")
	var filePath string

	// Check if variants folder exists
	if _, err := os.Stat(variantsDir); err == nil {
		activeVariant := GetActiveVariant(basePreset)
		variantPath := filepath.Join(variantsDir, activeVariant)

		// Ensure variant directory exists
		err = os.MkdirAll(variantPath, internal.DirPermissions)
		if err != nil {
			return fmt.Errorf(display.ErrDirectoryFailed)
		}

		filePath = filepath.Join(variantPath, fileType+".toml")
	} else {
		// Fallback to root files (backward compatible)
		filePath = filepath.Join(presetPath, fileType+".toml")
	}

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

// ===== UTILITY FUNCTIONS =====

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
