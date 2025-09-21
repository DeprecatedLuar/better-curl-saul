package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/parser"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

// Get retrieves values from TOML files for debugging
func Get(cmd parser.Command) (interface{}, error) {
	if cmd.Preset == "" {
		return nil, fmt.Errorf(errors.ErrPresetNameRequired)
	}
	if cmd.Target == "" {
		return nil, fmt.Errorf(errors.ErrTargetRequired)
	}

	// Normalize target aliases
	normalizedTarget := NormalizeTarget(cmd.Target)
	if normalizedTarget == "" {
		return nil, fmt.Errorf(errors.ErrInvalidTarget, cmd.Target)
	}
	cmd.Target = normalizedTarget

	// Load the TOML file for the target
	handler, err := presets.LoadPresetFile(cmd.Preset, cmd.Target)
	if err != nil {
		return nil, fmt.Errorf(errors.ErrFileLoadFailed, cmd.Target+".toml")
	}

	if len(cmd.KeyValuePairs) == 0 || cmd.KeyValuePairs[0].Key == "" {
		// Return entire TOML structure as a simple message
		return "TOML structure display not implemented yet", nil
	}

	// Get specific key
	key := cmd.KeyValuePairs[0].Key
	value := handler.Get(key)
	if value == nil {
		return nil, fmt.Errorf(errors.ErrKeyNotFound, key, cmd.Target)
	}

	return value, nil
}