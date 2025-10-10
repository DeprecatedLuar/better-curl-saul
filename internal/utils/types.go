// Package utils provides shared utility functions for Better-Curl-Saul
package utils

import (
	"strconv"
	"strings"
)

// InferValueType converts string values to appropriate Go types for TOML
func InferValueType(value string) interface{} {
	// Check for explicit array notation with brackets: [item1,item2,item3]
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		// Remove brackets and parse as array
		content := strings.TrimSpace(value[1 : len(value)-1])
		if content == "" {
			// Empty array
			return []string{}
		}

		// Split by comma and clean up each item
		parts := strings.Split(content, ",")
		var result []string
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			// Remove quotes if present
			if len(trimmed) >= 2 && trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
				trimmed = trimmed[1 : len(trimmed)-1]
			}
			result = append(result, trimmed)
		}
		return result
	}

	// Try to parse as boolean
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Default to string (no automatic comma-to-array conversion)
	return value
}