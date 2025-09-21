package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
)


// DisplayResponse formats and displays the HTTP response with optional filtering
func DisplayResponse(response *resty.Response, rawMode bool, preset string) {
	// Format response size
	size := formatBytes(len(response.Body()))
	
	// Get content type for metadata
	contentType := response.Header().Get("Content-Type")
	
	// Prepare response content
	body := response.String()
	var content string
	
	if body != "" {
		// Check if content appears to be JSON
		if isJSONContent(contentType, response.Body()) {
			// Apply filtering if filters are configured
			filteredBody := applyFiltering(response.Body(), preset)
			
			// If raw mode requested, show pretty JSON
			if rawMode {
				var jsonObj interface{}
				if err := json.Unmarshal(filteredBody, &jsonObj); err == nil {
					if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
						content = string(prettyJSON)
					}
				}
			} else {
				// Check if response is too large for TOML conversion
				if len(filteredBody) > 10000 {
					content = fmt.Sprintf("Response too large for TOML (%d bytes) - showing JSON:\n", len(filteredBody))
					var jsonObj interface{}
					if err := json.Unmarshal(filteredBody, &jsonObj); err == nil {
						if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
							content += string(prettyJSON)
						}
					}
				} else {
					// Default: Try TOML formatting for JSON responses
					if tomlFormatted := formatAsToml(filteredBody); tomlFormatted != "" {
						content = tomlFormatted
					} else {
						// Fallback to pretty JSON if TOML conversion fails
						var jsonObj interface{}
						if err := json.Unmarshal(filteredBody, &jsonObj); err == nil {
							if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
								content = string(prettyJSON)
							}
						}
					}
				}
			}
		}
		
		// If no content generated yet, use raw body
		if content == "" {
			content = body
		}
	} else {
		content = "(empty response)"
	}
	
	if rawMode {
		// Raw mode: output only the response body (Unix style)
		fmt.Print(response.String())
	} else {
		// Normal mode: formatted display with headers and metadata
		formatted := display.FormatResponse(
			response.Status(),
			contentType,
			response.Time().String(),
			size,
			content,
		)
		
		display.Plain(formatted)
	}
}

// formatBytes converts byte count to human-readable format
func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d bytes", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	}
}

// isJSONContent determines if the response content is JSON based on Content-Type and content
func isJSONContent(contentType string, body []byte) bool {
	// Check Content-Type header first
	if strings.Contains(strings.ToLower(contentType), "application/json") ||
		strings.Contains(strings.ToLower(contentType), "text/json") {
		return true
	}

	// If no clear Content-Type, try to parse as JSON
	var jsonObj interface{}
	return json.Unmarshal(body, &jsonObj) == nil
}

// formatAsToml converts JSON response to TOML format for readability
func formatAsToml(jsonData []byte) string {
	// Use our new TomlHandler FromJSON capability
	handler, err := toml.NewTomlHandlerFromJSON(jsonData)
	if err != nil {
		return "" // Fallback to other formatting
	}

	// Convert to TOML string
	tomlBytes, err := handler.ToBytes()
	if err != nil {
		return "" // Fallback to other formatting
	}

	return string(tomlBytes)
}

// applyFiltering applies JSON filtering if filters are configured for the preset
func applyFiltering(jsonData []byte, preset string) []byte {
	// Load filters configuration
	filtersHandler, err := presets.LoadPresetFile(preset, "filters")
	if err != nil {
		// No filters configured, return original data
		return jsonData
	}

	// Get fields array from filters.toml
	fieldsValue := filtersHandler.Get("fields")
	if fieldsValue == nil {
		// No fields configured, return original data
		return jsonData
	}

	// Convert to string slice (TOML array becomes []interface{})
	var fields []string
	switch v := fieldsValue.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				fields = append(fields, str)
			}
		}
	case []string:
		fields = v
	default:
		// Fallback: try as string for backward compatibility
		if str, ok := fieldsValue.(string); ok && str != "" {
			fields = strings.Split(str, ",")
		}
	}

	if len(fields) == 0 {
		return jsonData
	}

	// Apply filtering using gjson
	filtered := make(map[string]interface{})
	jsonStr := string(jsonData)

	for _, field := range fields {
		if field == "" {
			continue
		}

		// Use gjson to extract the field value
		result := gjson.Get(jsonStr, field)
		if result.Exists() {
			// Store the value in our filtered map
			// Use the original field path as the key for clarity
			filtered[field] = result.Value()
		}
	}

	// Convert filtered map back to JSON
	filteredJSON, err := json.Marshal(filtered)
	if err != nil {
		// If filtering fails, return original data
		return jsonData
	}

	// Warn if no fields matched
	if len(filtered) == 0 {
		display.Warning(fmt.Sprintf("No fields matched filters %v - check filter syntax", fields))
	}

	return filteredJSON
}

// FormatResponseContent applies same filtering/formatting as DisplayResponse
func FormatResponseContent(jsonData []byte, preset string, rawMode bool) string {
	filteredBody := applyFiltering(jsonData, preset)

	if rawMode {
		var jsonObj interface{}
		if json.Unmarshal(filteredBody, &jsonObj) == nil {
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				return string(prettyJSON)
			}
		}
		return string(filteredBody)
	}

	if tomlFormatted := formatAsToml(filteredBody); tomlFormatted != "" {
		return tomlFormatted
	}

	var jsonObj interface{}
	if json.Unmarshal(filteredBody, &jsonObj) == nil {
		if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
			return string(prettyJSON)
		}
	}

	return string(filteredBody)
}