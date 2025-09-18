package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"main/src/project/parser"
	"main/src/project/presets"
	"main/src/project/toml"
)

// ExecuteCallCommand handles HTTP execution for call commands
func ExecuteCallCommand(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf("preset name required for call command")
	}

	// Check if preset exists first
	presetPath, err := presets.GetPresetPath(cmd.Preset)
	if err != nil {
		return fmt.Errorf("failed to get preset path: %v", err)
	}

	// Check if preset directory exists
	if _, err := os.Stat(presetPath); os.IsNotExist(err) {
		return fmt.Errorf("preset '%s' does not exist. Create it first with: saul %s", cmd.Preset, cmd.Preset)
	}

	// Check for --persist flag
	persist := false
	// TODO: Add proper flag parsing when needed

	// Prompt for variables and get substitution map
	substitutions, err := PromptForVariables(cmd.Preset, persist)
	if err != nil {
		return fmt.Errorf("variable prompting failed: %v", err)
	}

	// Merge all TOML files into one
	mergedHandler, err := MergePresetFiles(cmd.Preset)
	if err != nil {
		return fmt.Errorf("failed to merge preset files: %v", err)
	}

	// Apply variable substitutions
	err = SubstituteVariables(mergedHandler, substitutions)
	if err != nil {
		return fmt.Errorf("variable substitution failed: %v", err)
	}

	// Extract HTTP request components
	request, err := BuildHTTPRequest(mergedHandler)
	if err != nil {
		return fmt.Errorf("failed to build HTTP request: %v", err)
	}

	// Execute the HTTP request
	response, err := ExecuteHTTPRequest(request)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}

	// Display response
	DisplayResponse(response)

	return nil
}

// HTTPRequestConfig holds the components of an HTTP request
type HTTPRequestConfig struct {
	Method  string
	URL     string
	Timeout int
	Headers map[string]string
	Body    []byte
	Query   map[string]string
}

// MergePresetFiles merges all TOML files in a preset into one handler
func MergePresetFiles(preset string) (*toml.TomlHandler, error) {
	presetPath, err := presets.GetPresetPath(preset)
	if err != nil {
		return nil, fmt.Errorf("failed to get preset path: %v", err)
	}

	// List of files to merge in order
	files := []string{"request.toml", "headers.toml", "query.toml", "body.toml"}
	var handlers []*toml.TomlHandler

	// Load each file that exists
	for _, filename := range files {
		filePath := filepath.Join(presetPath, filename)
		handler, err := toml.NewTomlHandler(filePath)
		if err != nil {
			// Skip files that don't exist or can't be loaded
			continue
		}
		handlers = append(handlers, handler)
	}

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no TOML files found in preset %s", preset)
	}

	// Start with first handler as base
	base := handlers[0]

	// Merge remaining handlers
	if len(handlers) > 1 {
		err := base.MergeMultiple(handlers[1:]...)
		if err != nil {
			return nil, fmt.Errorf("failed to merge TOML files: %v", err)
		}
	}

	return base, nil
}

// BuildHTTPRequest converts merged TOML to HTTP request configuration
func BuildHTTPRequest(handler *toml.TomlHandler) (*HTTPRequestConfig, error) {
	config := &HTTPRequestConfig{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	// Extract request settings
	if method := handler.GetAsString("method"); method != "" {
		config.Method = strings.ToUpper(method)
	} else {
		config.Method = "GET" // Default method
	}

	config.URL = handler.GetAsString("url")
	if config.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	// Parse timeout
	if timeoutStr := handler.GetAsString("timeout"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			config.Timeout = timeout
		}
	}
	if config.Timeout == 0 {
		config.Timeout = 30 // Default timeout
	}

	// Extract headers - look for flat key-value pairs that might be headers
	for _, key := range handler.Keys() {
		value := handler.Get(key)
		if key != "method" && key != "url" && key != "timeout" {
			// Check if this is a simple string value (likely a header)
			if strValue, ok := value.(string); ok {
				config.Headers[key] = strValue
			}
		}
	}

	// Convert remaining data to JSON for body (excluding request metadata)
	bodyData := make(map[string]interface{})
	for _, key := range handler.Keys() {
		if key != "method" && key != "url" && key != "timeout" {
			value := handler.Get(key)
			// Skip simple string values (headers) and include complex structures
			if _, isString := value.(string); !isString {
				bodyData[key] = value
			}
		}
	}

	// Convert body to JSON if we have data
	if len(bodyData) > 0 {
		jsonData, err := json.Marshal(bodyData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert body to JSON: %v", err)
		}
		config.Body = jsonData

		// Set Content-Type if not already set
		if _, exists := config.Headers["Content-Type"]; !exists {
			config.Headers["Content-Type"] = "application/json"
		}
	}

	return config, nil
}

// ExecuteHTTPRequest performs the actual HTTP request using resty
func ExecuteHTTPRequest(config *HTTPRequestConfig) (*resty.Response, error) {
	client := resty.New()
	client.SetTimeout(time.Duration(config.Timeout) * time.Second)

	request := client.R()

	// Set headers
	for key, value := range config.Headers {
		request.SetHeader(key, value)
	}

	// Set query parameters
	for key, value := range config.Query {
		request.SetQueryParam(key, value)
	}

	// Set body if present
	if len(config.Body) > 0 {
		request.SetBody(config.Body)
	}

	// Execute request based on method
	switch config.Method {
	case "GET":
		return request.Get(config.URL)
	case "POST":
		return request.Post(config.URL)
	case "PUT":
		return request.Put(config.URL)
	case "DELETE":
		return request.Delete(config.URL)
	case "PATCH":
		return request.Patch(config.URL)
	case "HEAD":
		return request.Head(config.URL)
	case "OPTIONS":
		return request.Options(config.URL)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", config.Method)
	}
}

// DisplayResponse formats and displays the HTTP response
func DisplayResponse(response *resty.Response) {
	fmt.Printf("Status: %s\n", response.Status())
	fmt.Printf("Time: %v\n", response.Time())
	fmt.Printf("Size: %d bytes\n", len(response.Body()))

	// Display headers
	if len(response.Header()) > 0 {
		fmt.Println("\nHeaders:")
		for key, values := range response.Header() {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	// Display body
	fmt.Println("\nResponse:")
	body := response.String()
	if body != "" {
		// Try to pretty-print JSON
		var jsonObj interface{}
		if err := json.Unmarshal(response.Body(), &jsonObj); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
				fmt.Println(string(prettyJSON))
				return
			}
		}
		// Fallback to raw body
		fmt.Println(body)
	} else {
		fmt.Println("(empty response)")
	}
}