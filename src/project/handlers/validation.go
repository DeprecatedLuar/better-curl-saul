// Package handlers provides command execution logic and validation for Better-Curl-Saul.
// This package orchestrates HTTP requests, variable processing, validation,
// and integrates all components to execute user commands.
package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/src/modules/errors"
)

// ValidateRequestField validates special request field values
func ValidateRequestField(key, value string) error {
	switch strings.ToLower(key) {
	case "method":
		return validateHTTPMethod(value)
	case "url":
		return validateURL(value)
	case "timeout":
		return validateTimeout(value)
	case "history", "history_count":
		return validateHistoryCount(value)
	default:
		return nil
	}
}

// validateHTTPMethod checks if the HTTP method is valid
func validateHTTPMethod(method string) error {
	validMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH",
		"HEAD", "OPTIONS", "TRACE", "CONNECT",
	}

	methodUpper := strings.ToUpper(method)
	for _, valid := range validMethods {
		if methodUpper == valid {
			return nil
		}
	}

	return fmt.Errorf(errors.ErrInvalidMethod, method)
}

// validateURL performs basic URL validation
func validateURL(url string) error {
	if url == "" {
		return fmt.Errorf(errors.ErrMissingURL)
	}
	// Basic check - should start with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf(errors.ErrInvalidURL)
	}
	return nil
}

// validateTimeout validates timeout value
func validateTimeout(timeout string) error {
	if _, err := strconv.Atoi(timeout); err != nil {
		return fmt.Errorf(errors.ErrInvalidTimeout)
	}
	return nil
}

// validateHistoryCount validates history count value
func validateHistoryCount(count string) error {
	historyCount, err := strconv.Atoi(count)
	if err != nil {
		return fmt.Errorf("invalid history count: %s (must be a number)", count)
	}
	if historyCount < 0 {
		return fmt.Errorf("history count cannot be negative: %d", historyCount)
	}
	if historyCount > 100 {
		return fmt.Errorf("history count cannot exceed 100: %d", historyCount)
	}
	return nil
}

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