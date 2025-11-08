// Package internal provides format parsing and conversion utilities for Better-Curl-Saul.
// This file contains ALL format parsing logic: curl, JSON, TOML, HTTP.
package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	lib "github.com/pelletier/go-toml"
)

// ===== CURL PARSING =====

// CurlRequest represents a parsed curl command
type CurlRequest struct {
	Method  string
	URL     string
	BaseURL string
	Query   map[string]string
	Headers map[string]string
	Body    string
}

// ParseCurl parses a curl command string into a structured request
func ParseCurl(curlCmd string) (*CurlRequest, error) {
	curlCmd = strings.TrimSpace(curlCmd)

	if !strings.HasPrefix(curlCmd, "curl") {
		return nil, fmt.Errorf("command must start with 'curl'")
	}

	req := &CurlRequest{
		Method:  "GET",
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	methodRegex := regexp.MustCompile(`(?:-X|--request)\s+([A-Z]+)`)
	if match := methodRegex.FindStringSubmatch(curlCmd); len(match) > 1 {
		req.Method = match[1]
	}

	urlRegex := regexp.MustCompile(`(?:'(https?://[^']+)'|"(https?://[^"]+)"|(https?://\S+))`)
	if match := urlRegex.FindStringSubmatch(curlCmd); len(match) > 1 {
		if match[1] != "" {
			req.URL = match[1]
		} else if match[2] != "" {
			req.URL = match[2]
		} else if match[3] != "" {
			req.URL = match[3]
		}
	}

	if req.URL != "" {
		parsedURL, err := url.Parse(req.URL)
		if err == nil {
			req.BaseURL = parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
			for key, values := range parsedURL.Query() {
				if len(values) > 0 {
					req.Query[key] = values[0]
				}
			}
		}
	}

	headerRegex := regexp.MustCompile(`(?:-H|--header)\s+(?:'([^']+)'|"([^"]+)")`)
	for _, match := range headerRegex.FindAllStringSubmatch(curlCmd, -1) {
		headerStr := match[1]
		if headerStr == "" {
			headerStr = match[2]
		}
		parts := strings.SplitN(headerStr, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Headers[key] = value
		}
	}

	bodyRegex := regexp.MustCompile(`(?:-d|--data|--data-raw)\s+(?:'([^']*)'|"([^"]*)"|(\S+))`)
	if match := bodyRegex.FindStringSubmatch(curlCmd); len(match) > 1 {
		if match[1] != "" {
			req.Body = match[1]
		} else if match[2] != "" {
			req.Body = match[2]
		} else if match[3] != "" {
			req.Body = match[3]
		}
	}

	return req, nil
}

// ===== JSON/TOML CONVERSION =====

// TomlHandler reference type for conversions (actual type defined in workspace)
type TomlHandler struct {
	tree *lib.Tree
}

// NewTomlHandlerFromJSON creates a TOML handler from JSON bytes
func NewTomlHandlerFromJSON(jsonData []byte) (*TomlHandler, error) {
	var goMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &goMap); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	tree, err := lib.TreeFromMap(goMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create TOML tree: %v", err)
	}

	return &TomlHandler{tree: tree}, nil
}

// FormatAsToml converts JSON response to TOML format for readability
func FormatAsToml(jsonData []byte) string {
	handler, err := NewTomlHandlerFromJSON(jsonData)
	if err != nil {
		return ""
	}

	tomlBytes, err := handler.tree.Marshal()
	if err != nil {
		return ""
	}

	return string(tomlBytes)
}

// FormatResponseContent applies filtering and formatting to response content
func FormatResponseContent(jsonData []byte, preset string, rawMode bool) string {
	if rawMode {
		return string(jsonData)
	}

	filteredBody := ApplyFiltering(jsonData, preset)

	if tomlFormatted := FormatAsToml(filteredBody); tomlFormatted != "" {
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

// ===== JSON FILTERING =====

// ApplyFiltering applies JSON filtering if filters are configured for the preset
func ApplyFiltering(jsonData []byte, preset string) []byte {
	// This will be implemented to load filters from workspace
	// For now, just return jsonData as-is
	// TODO: Load filters from preset and apply gjson filtering
	return jsonData
}

// ===== CONTENT TYPE DETECTION =====

// IsJSONContent checks if content type or body indicates JSON
func IsJSONContent(contentType string, body []byte) bool {
	if strings.Contains(strings.ToLower(contentType), "application/json") ||
		strings.Contains(strings.ToLower(contentType), "text/json") {
		return true
	}

	var jsonObj interface{}
	return json.Unmarshal(body, &jsonObj) == nil
}

// ===== CURL IMPORT/EXPORT (High-level operations) =====

// ImportCurlString parses a curl command and returns structured data
// Caller is responsible for saving to workspace files
func ImportCurlString(curlCmd string) (*CurlRequest, error) {
	return ParseCurl(curlCmd)
}

// ExportToCurl generates a curl command from structured data
func ExportToCurl(method, baseURL string, headers, query map[string]string, body []byte) (string, error) {
	var curlParts []string
	curlParts = append(curlParts, "curl")

	if strings.ToUpper(method) != "GET" {
		curlParts = append(curlParts, fmt.Sprintf("-X %s", strings.ToUpper(method)))
	}

	finalURL := baseURL
	if len(query) > 0 {
		if strings.ToUpper(method) == "GET" {
			curlParts = append(curlParts, "-G")
			for key, value := range query {
				escapedValue := strings.ReplaceAll(value, "'", "'\\''")
				curlParts = append(curlParts, fmt.Sprintf("--data-urlencode '%s=%s'", key, escapedValue))
			}
		} else {
			queryParams := url.Values{}
			for key, value := range query {
				queryParams.Add(key, value)
			}
			if strings.Contains(finalURL, "?") {
				finalURL += "&" + queryParams.Encode()
			} else {
				finalURL += "?" + queryParams.Encode()
			}
		}
	}

	curlParts = append(curlParts, fmt.Sprintf("'%s'", finalURL))

	for key, value := range headers {
		escapedValue := strings.ReplaceAll(value, "'", "'\\''")
		curlParts = append(curlParts, fmt.Sprintf("-H '%s: %s'", key, escapedValue))
	}

	if len(body) > 0 {
		escapedJSON := strings.ReplaceAll(string(body), "'", "'\\''")
		curlParts = append(curlParts, fmt.Sprintf("-d '%s'", escapedJSON))
	}

	return formatMultilineCurl(curlParts), nil
}

// formatMultilineCurl formats curl command parts into multiline string
func formatMultilineCurl(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	var lines []string
	for i := 0; i < len(parts); i++ {
		if i == 0 {
			lines = append(lines, parts[i]+" \\")
		} else if i < len(parts)-1 {
			lines = append(lines, "  "+parts[i]+" \\")
		} else {
			lines = append(lines, "  "+parts[i])
		}
	}
	return strings.Join(lines, "\n")
}
