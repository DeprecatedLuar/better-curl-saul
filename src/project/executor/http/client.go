package http

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/toml"
)

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