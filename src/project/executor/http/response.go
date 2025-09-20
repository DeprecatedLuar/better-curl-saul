package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
)

// DisplayResponse formats and displays the HTTP response
func DisplayResponse(response *resty.Response, rawMode bool) {
	// Display response metadata
	fmt.Printf("Status: %s (%v, %d bytes)\n", response.Status(), response.Time(), len(response.Body()))

	// Get content type for smart formatting
	contentType := response.Header().Get("Content-Type")
	fmt.Printf("Content-Type: %s\n", contentType)

	// Display headers
	if len(response.Header()) > 0 {
		fmt.Println("\nHeaders:")
		for key, values := range response.Header() {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	// Display body with smart formatting
	fmt.Println("\nResponse:")
	body := response.String()
	if body != "" {
		// Check if content appears to be JSON
		if isJSONContent(contentType, response.Body()) {
			// If raw mode requested, show pretty JSON
			if rawMode {
				var jsonObj interface{}
				if err := json.Unmarshal(response.Body(), &jsonObj); err == nil {
					if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
						fmt.Println(string(prettyJSON))
						return
					}
				}
			} else {
				// Check if response is too large for TOML conversion
				if len(response.Body()) > 10000 {
					fmt.Printf("Response too large for TOML (%d bytes) - showing JSON:\n", len(response.Body()))
					var jsonObj interface{}
					if err := json.Unmarshal(response.Body(), &jsonObj); err == nil {
						if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
							fmt.Println(string(prettyJSON))
							return
						}
					}
					// If JSON parsing fails, fall through to raw display
				} else {
					// Default: Try TOML formatting for JSON responses
					if tomlFormatted := formatAsToml(response.Body()); tomlFormatted != "" {
						fmt.Println(tomlFormatted)
						return
					}
				}
				// Fallback to pretty JSON if TOML conversion fails
				var jsonObj interface{}
				if err := json.Unmarshal(response.Body(), &jsonObj); err == nil {
					if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
						fmt.Println(string(prettyJSON))
						return
					}
				}
			}
		}
		// Fallback to raw body for non-JSON or failed conversions
		fmt.Println(body)
	} else {
		fmt.Println("(empty response)")
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