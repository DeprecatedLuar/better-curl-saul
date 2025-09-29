package workspace

import (
	"fmt"
	"net/url"
	"strings"
)

// ExportToCurl exports a preset to a curl command string
// Variables are preserved as-is ({@token}, {?name}) for documentation/sharing
func ExportToCurl(preset string) (string, error) {
	// Load all TOML files
	requestHandler, err := LoadPresetFile(preset, "request")
	if err != nil {
		return "", fmt.Errorf("failed to load request: %v", err)
	}

	headersHandler, err := LoadPresetFile(preset, "headers")
	if err != nil {
		return "", fmt.Errorf("failed to load headers: %v", err)
	}

	queryHandler, err := LoadPresetFile(preset, "query")
	if err != nil {
		return "", fmt.Errorf("failed to load query: %v", err)
	}

	bodyHandler, err := LoadPresetFile(preset, "body")
	if err != nil {
		return "", fmt.Errorf("failed to load body: %v", err)
	}

	// Extract request data
	method := requestHandler.GetAsString("method")
	if method == "" {
		method = "GET" // Default method
	}
	baseURL := requestHandler.GetAsString("url")
	if baseURL == "" {
		return "", fmt.Errorf("preset has no URL configured")
	}

	// Start building curl command
	var curlParts []string
	curlParts = append(curlParts, "curl")

	// Add method (only if not GET, as it's curl's default)
	if strings.ToUpper(method) != "GET" {
		curlParts = append(curlParts, fmt.Sprintf("-X %s", strings.ToUpper(method)))
	}

	// Handle query parameters
	queryKeys := queryHandler.Keys()
	finalURL := baseURL
	if len(queryKeys) > 0 {
		if strings.ToUpper(method) == "GET" {
			// For GET: use -G with --data-urlencode (modern curl standard)
			curlParts = append(curlParts, "-G")
			for _, key := range queryKeys {
				value := queryHandler.GetAsString(key)
				escapedValue := strings.ReplaceAll(value, "'", "'\\''")
				curlParts = append(curlParts, fmt.Sprintf("--data-urlencode '%s=%s'", key, escapedValue))
			}
		} else {
			// For non-GET: append query params directly to URL
			queryParams := url.Values{}
			for _, key := range queryKeys {
				value := queryHandler.GetAsString(key)
				queryParams.Add(key, value)
			}
			if strings.Contains(finalURL, "?") {
				finalURL += "&" + queryParams.Encode()
			} else {
				finalURL += "?" + queryParams.Encode()
			}
		}
	}

	// Add URL (quote it for safety)
	curlParts = append(curlParts, fmt.Sprintf("'%s'", finalURL))

	// Add headers
	headerKeys := headersHandler.Keys()
	for _, key := range headerKeys {
		value := headersHandler.GetAsString(key)
		// Escape single quotes in header values
		escapedValue := strings.ReplaceAll(value, "'", "'\\''")
		curlParts = append(curlParts, fmt.Sprintf("-H '%s: %s'", key, escapedValue))
	}

	// Add body (if not empty)
	bodyKeys := bodyHandler.Keys()
	if len(bodyKeys) > 0 {
		// Use compact JSON for curl compatibility (not pretty-printed)
		jsonBody, err := bodyHandler.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to convert body to JSON: %v", err)
		}
		// Escape single quotes in JSON for shell safety
		escapedJSON := strings.ReplaceAll(string(jsonBody), "'", "'\\''")
		curlParts = append(curlParts, fmt.Sprintf("-d '%s'", escapedJSON))
	}

	// Join with line continuations for readability
	return formatMultilineCurl(curlParts), nil
}

// formatMultilineCurl formats curl command parts into multiline string with backslash continuation
func formatMultilineCurl(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		return parts[0]
	}

	var lines []string

	// All parts with 2-space indent and backslash continuation (except last)
	for i := 0; i < len(parts); i++ {
		if i == 0 {
			// First line (curl) with backslash
			lines = append(lines, parts[i]+" \\")
		} else if i < len(parts)-1 {
			// Middle lines with indent and backslash
			lines = append(lines, "  "+parts[i]+" \\")
		} else {
			// Last line with indent, no backslash
			lines = append(lines, "  "+parts[i])
		}
	}

	return strings.Join(lines, "\n")
}

// escapeShellValue escapes a value for use in shell single quotes
func escapeShellValue(value string) string {
	// In single quotes, only ' needs escaping: 'can'\''t' → can't
	return strings.ReplaceAll(value, "'", "'\\''")
}

// urlEncodeParam encodes a parameter for URL query strings
func urlEncodeParam(key, value string) string {
	return fmt.Sprintf("%s=%s", url.QueryEscape(key), url.QueryEscape(value))
}