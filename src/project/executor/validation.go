package executor

import (
	"fmt"
	"strconv"
	"strings"
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

	return fmt.Errorf("sorry champ \"%s\" isn't really a thing, but i'll let you try again", method)
}

// validateURL performs basic URL validation
func validateURL(url string) error {
	if url == "" {
		return fmt.Errorf("listen pal, at least put in the URL. Come on")
	}
	// Basic check - should start with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("alright, so the \"U R L\" needs to start with one of these two here: 'http://' or 'https://'. Go get'em tiger")
	}
	return nil
}

// validateTimeout validates timeout value
func validateTimeout(timeout string) error {
	if _, err := strconv.Atoi(timeout); err != nil {
		return fmt.Errorf("timeout must be a number (seconds)")
	}
	return nil
}

// InferValueType converts string values to appropriate Go types for TOML
func InferValueType(value string) interface{} {
	// For now, keep everything as strings to avoid TOML handler issues
	// TODO: Implement proper type handling once TOML handler supports all types

	// Try to parse as boolean (keep this as it's simple)
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Check for array notation (comma-separated values)
	if strings.Contains(value, ",") {
		parts := strings.Split(value, ",")
		var result []string
		for _, part := range parts {
			result = append(result, strings.TrimSpace(part))
		}
		return result
	}

	// Default to string (including numbers for now)
	return value
}