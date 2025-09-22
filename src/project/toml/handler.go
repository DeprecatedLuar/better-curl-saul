package toml

import (
	"fmt"

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



// load parses the raw TOML data into a tree structure
func (t *TomlHandler) load() error {
	var err error
	t.tree, err = lib.LoadBytes(t.raw)
	return err
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



