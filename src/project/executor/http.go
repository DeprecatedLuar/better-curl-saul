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

	// Load each file as separate handler - no merging
	requestHandler := LoadPresetFile(cmd.Preset, "request")
	headersHandler := LoadPresetFile(cmd.Preset, "headers")
	bodyHandler := LoadPresetFile(cmd.Preset, "body")
	queryHandler := LoadPresetFile(cmd.Preset, "query")

	// Apply variable substitutions to each separately
	err = SubstituteVariables(requestHandler, substitutions)
	if err != nil {
		return fmt.Errorf("request variable substitution failed: %v", err)
	}
	err = SubstituteVariables(headersHandler, substitutions)
	if err != nil {
		return fmt.Errorf("headers variable substitution failed: %v", err)
	}
	err = SubstituteVariables(bodyHandler, substitutions)
	if err != nil {
		return fmt.Errorf("body variable substitution failed: %v", err)
	}
	err = SubstituteVariables(queryHandler, substitutions)
	if err != nil {
		return fmt.Errorf("query variable substitution failed: %v", err)
	}

	// Build HTTP request components explicitly - no guessing
	request, err := BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
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

// LoadPresetFile loads a single TOML file as a handler, returns empty handler if file doesn't exist
func LoadPresetFile(preset, filename string) *toml.TomlHandler {
	presetPath, err := presets.GetPresetPath(preset)
	if err != nil {
		// Return empty handler if preset path fails
		return createEmptyHandler()
	}

	filePath := filepath.Join(presetPath, filename+".toml")
	handler, err := toml.NewTomlHandler(filePath)
	if err != nil {
		// Return empty handler if file doesn't exist or can't be loaded
		return createEmptyHandler()
	}
	return handler
}

// createEmptyHandler creates an empty TOML handler for missing files
func createEmptyHandler() *toml.TomlHandler {
	// Create a temporary file to initialize empty handler
	tempFile, err := os.CreateTemp("", "empty*.toml")
	if err != nil {
		return nil
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	handler, _ := toml.NewTomlHandler(tempFile.Name())
	return handler
}

// BuildHTTPRequestFromHandlers builds HTTP request from separate handlers - no guessing
func BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler *toml.TomlHandler) (*HTTPRequestConfig, error) {
	config := &HTTPRequestConfig{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	// Extract request settings ONLY from request handler
	if method := requestHandler.GetAsString("method"); method != "" {
		config.Method = strings.ToUpper(method)
	} else {
		config.Method = "GET" // Default method
	}

	config.URL = requestHandler.GetAsString("url")
	if config.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	// Parse timeout from request handler
	if timeoutStr := requestHandler.GetAsString("timeout"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			config.Timeout = timeout
		}
	}
	if config.Timeout == 0 {
		config.Timeout = 30 // Default timeout
	}

	// Extract headers ONLY from headers handler
	for _, key := range headersHandler.Keys() {
		value := headersHandler.GetAsString(key)
		if value != "" {
			config.Headers[key] = value
		}
	}

	// Extract query parameters ONLY from query handler
	for _, key := range queryHandler.Keys() {
		value := queryHandler.GetAsString(key)
		if value != "" {
			config.Query[key] = value
		}
	}

	// Convert body ONLY from body handler to JSON
	bodyKeys := bodyHandler.Keys()
	if len(bodyKeys) > 0 {
		// Convert entire body handler to JSON
		bodyJSON, err := bodyHandler.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to convert body to JSON: %v", err)
		}
		config.Body = []byte(bodyJSON)

		// Set Content-Type if not already set in headers
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