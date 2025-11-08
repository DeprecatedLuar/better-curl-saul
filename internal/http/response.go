package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/internal"
)


// DisplayResponse formats and displays the HTTP response with optional filtering
func DisplayResponse(response *resty.Response, rawMode bool, preset string, responseFormat string) {
	// Format response size
	size := formatBytes(len(response.Body()))

	// Get content type for metadata
	contentType := response.Header().Get("Content-Type")

	// Handle response format overrides
	if responseFormat != "" {
		displayFormattedResponse(response, responseFormat, rawMode, preset)
		return
	}

	// Prepare response content
	body := response.String()
	var content string

	if body != "" {
		// Check if content appears to be JSON
		if internal.IsJSONContent(contentType, response.Body()) {
			// Apply filtering if filters are configured
			filteredBody := internal.ApplyFiltering(response.Body(), preset)

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
					// Fallback to pretty JSON
					var jsonObj interface{}
					if err := json.Unmarshal(filteredBody, &jsonObj); err == nil {
						if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
							content += string(prettyJSON)
						}
					}
				} else {
					// Normal flow - try TOML first
					if tomlFormatted := internal.FormatAsToml(filteredBody); tomlFormatted != "" {
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
		} else {
			// Non-JSON content
			content = body
		}
	} else {
		content = "(empty body)"
	}

	// Display status, size, content type
	fmt.Printf("%s (%s, %s)\n\n", response.Status(), size, contentType)

	// Display formatted content
	fmt.Print(content)

	// Ensure final newline
	if !strings.HasSuffix(content, "\n") {
		fmt.Println()
	}
}

// displayFormattedResponse handles custom response format displays
func displayFormattedResponse(response *resty.Response, format string, rawMode bool, preset string) {
	switch format {
	case "headers-only":
		for key, values := range response.Header() {
			fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
		}
	case "status-only":
		fmt.Println(response.Status())
	case "body-only":
		fmt.Print(internal.FormatResponseContent(response.Body(), preset, rawMode))
	}
}

func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d bytes", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	}
}
