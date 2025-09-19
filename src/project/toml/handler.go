package toml

import (
	"encoding/json"
	"fmt"
	"os"

	lib "github.com/pelletier/go-toml"
)

// TomlHandler provides TOML file manipulation capabilities
type TomlHandler struct {
	path string
	out  string
	raw  []byte
	tree *lib.Tree
}

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

// NewTomlHandlerFromJSON creates a TOML handler from JSON bytes
func NewTomlHandlerFromJSON(jsonData []byte) (*TomlHandler, error) {
	// Unmarshal JSON to Go map
	var goMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &goMap); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	// Convert Go map to TOML tree
	tree, err := lib.TreeFromMap(goMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create TOML tree: %v", err)
	}

	// Create handler with the tree
	handler := &TomlHandler{
		tree: tree,
		raw:  jsonData, // Keep original JSON as raw data
	}

	return handler, nil
}

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

	return os.WriteFile(path, []byte(tomlString), 0644)
}

// ToBytes returns the TOML data as bytes
func (t *TomlHandler) ToBytes() ([]byte, error) {
	tomlString, err := t.tree.ToTomlString()
	if err != nil {
		return nil, err
	}
	return []byte(tomlString), nil
}

// ToJSON converts the TOML data to JSON
// Perfect for HTTP request payloads
func (t *TomlHandler) ToJSON() ([]byte, error) {
	// Convert TOML tree to Go map
	goMap := t.tree.ToMap()
	
	// Convert Go map to JSON
	return json.Marshal(goMap)
}

// ToJSONPretty converts the TOML data to pretty-printed JSON
func (t *TomlHandler) ToJSONPretty() ([]byte, error) {
	goMap := t.tree.ToMap()
	return json.MarshalIndent(goMap, "", "  ")
}

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

// Clone creates a copy of the TOML handler
func (t *TomlHandler) Clone() (*TomlHandler, error) {
	data, err := t.ToBytes()
	if err != nil {
		return nil, err
	}
	return NewTomlHandlerFromBytes(data)
}

// Keys returns all top-level keys in the TOML
func (t *TomlHandler) Keys() []string {
	return t.tree.Keys()
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

// LoadTomlAsJSON loads a single TOML file and converts to JSON
func LoadTomlAsJSON(filePath string) ([]byte, error) {
	handler, err := NewTomlHandler(filePath)
	if err != nil {
		return nil, err
	}
	return handler.ToJSON()
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