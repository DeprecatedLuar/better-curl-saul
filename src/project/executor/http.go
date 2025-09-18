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

// PresetHandlers holds separate TOML handlers for each file type
type PresetHandlers struct {
	Request   *toml.TomlHandler
	Headers   *toml.TomlHandler
	Query     *toml.TomlHandler
	Body      *toml.TomlHandler
	Variables *toml.TomlHandler
}

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

	// Load all TOML files into separate handlers
	handlers, err := LoadPresetFiles(cmd.Preset)
	if err != nil {
		return fmt.Errorf("failed to load preset files: %v", err)
	}

	// Apply variable substitutions across all handlers
	err = SubstituteVariablesInHandlers(handlers, substitutions)
	if err != nil {
		return fmt.Errorf("variable substitution failed: %v", err)
	}

	// Extract HTTP request components using file-based classification
	request, err := BuildHTTPRequest(handlers)
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

// LoadPresetFiles loads all TOML files in a preset into separate handlers
func LoadPresetFiles(preset string) (*PresetHandlers, error) {
	presetPath, err := presets.GetPresetPath(preset)
	if err != nil {
		return nil, fmt.Errorf("failed to get preset path: %v", err)
	}

	handlers := &PresetHandlers{}

	// Map of file names to handler pointers
	files := map[string]**toml.TomlHandler{
		"request.toml":   &handlers.Request,
		"headers.toml":   &handlers.Headers,
		"query.toml":     &handlers.Query,
		"body.toml":      &handlers.Body,
		"variables.toml": &handlers.Variables,
	}

	// Load each file that exists (lazy creation - missing files are ok)
	for filename, handlerPtr := range files {
		filePath := filepath.Join(presetPath, filename)
		if handler, err := toml.NewTomlHandler(filePath); err == nil {
			*handlerPtr = handler
		}
		// Skip files that don't exist or can't be loaded
	}

	return handlers, nil
}

// SubstituteVariablesInHandlers replaces variables across all handlers
func SubstituteVariablesInHandlers(handlers *PresetHandlers, substitutions map[string]string) error {
	handlerList := []*toml.TomlHandler{
		handlers.Request,
		handlers.Headers,
		handlers.Query,
		handlers.Body,
	}

	for _, handler := range handlerList {
		if handler != nil {
			err := SubstituteVariables(handler, substitutions)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// BuildHTTPRequest converts separate TOML handlers to HTTP request configuration using file-based classification
func BuildHTTPRequest(handlers *PresetHandlers) (*HTTPRequestConfig, error) {
	config := &HTTPRequestConfig{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	// Extract request settings from request.toml only
	if handlers.Request != nil {
		if method := handlers.Request.GetAsString("method"); method != "" {
			config.Method = strings.ToUpper(method)
		} else {
			config.Method = "GET" // Default method
		}

		config.URL = handlers.Request.GetAsString("url")
		if config.URL == "" {
			return nil, fmt.Errorf("URL is required")
		}

		// Parse timeout
		if timeoutStr := handlers.Request.GetAsString("timeout"); timeoutStr != "" {
			if timeout, err := strconv.Atoi(timeoutStr); err == nil {
				config.Timeout = timeout
			}
		}
	} else {
		return nil, fmt.Errorf("URL is required")
	}

	if config.Timeout == 0 {
		config.Timeout = 30 // Default timeout
	}

	// Extract headers from headers.toml only
	if handlers.Headers != nil {
		for _, key := range handlers.Headers.Keys() {
			config.Headers[key] = handlers.Headers.GetAsString(key)
		}
	}

	// Extract query parameters from query.toml only
	if handlers.Query != nil {
		for _, key := range handlers.Query.Keys() {
			config.Query[key] = handlers.Query.GetAsString(key)
		}
	}

	// Convert body.toml to JSON (and only body.toml)
	if handlers.Body != nil {
		jsonData, err := handlers.Body.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to convert body to JSON: %v", err)
		}

		// Only set body if we have actual content
		if string(jsonData) != "{}" {
			config.Body = jsonData

			// Set Content-Type if not already set
			if _, exists := config.Headers["Content-Type"]; !exists {
				config.Headers["Content-Type"] = "application/json"
			}
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