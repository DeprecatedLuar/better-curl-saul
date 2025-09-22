package toml

import (
	"encoding/json"
	"fmt"

	lib "github.com/pelletier/go-toml"
)

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

// LoadTomlAsJSON loads a single TOML file and converts to JSON
func LoadTomlAsJSON(filePath string) ([]byte, error) {
	handler, err := NewTomlHandler(filePath)
	if err != nil {
		return nil, err
	}
	return handler.ToJSON()
}