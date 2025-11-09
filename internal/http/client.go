package http

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
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
// Now uses variant-aware workspace.LoadPresetFile() for proper variant support
func LoadPresetFile(preset, filename string) *workspace.TomlHandler {
	handler, err := workspace.LoadPresetFile(preset, filename)
	if err != nil {
		// Return empty handler if file doesn't exist or can't be loaded
		return createEmptyHandler()
	}
	return handler
}

// createEmptyHandler creates an empty TOML handler for missing files
func createEmptyHandler() *workspace.TomlHandler {
	// Create a temporary file to initialize empty handler
	tempFile, err := os.CreateTemp("", "empty*.toml")
	if err != nil {
		return nil
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	handler, _ := workspace.NewTomlHandler(tempFile.Name())
	return handler
}

// BuildHTTPRequestFromHandlers builds HTTP request from separate handlers - no guessing
func BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler *workspace.TomlHandler) (*HTTPRequestConfig, error) {
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
		return nil, fmt.Errorf(display.ErrMissingURL)
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
			return nil, fmt.Errorf(display.ErrRequestBuildFailed)
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
		return nil, fmt.Errorf(display.ErrUnsupportedMethod, config.Method)
	}
}

// ValidateRequestField validates HTTP request field values
func ValidateRequestField(key, value string) error {
	switch strings.ToLower(key) {
	case "method":
		return validateHTTPMethod(value)
	case "url":
		return validateURL(value)
	case "timeout":
		return validateTimeout(value)
	case "history", "history_count":
		return validateHistoryCount(value)
	default:
		return nil
	}
}

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

	return fmt.Errorf(display.ErrInvalidMethod, method)
}

func validateURL(url string) error {
	if url == "" {
		return fmt.Errorf(display.ErrMissingURL)
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf(display.ErrInvalidURL)
	}
	return nil
}

func validateTimeout(timeout string) error {
	if _, err := strconv.Atoi(timeout); err != nil {
		return fmt.Errorf(display.ErrInvalidTimeout)
	}
	return nil
}

func validateHistoryCount(count string) error {
	historyCount, err := strconv.Atoi(count)
	if err != nil {
		return fmt.Errorf("invalid history count: %s (must be a number)", count)
	}
	if historyCount < 0 {
		return fmt.Errorf("history count cannot be negative: %d", historyCount)
	}
	if historyCount > 100 {
		return fmt.Errorf("history count cannot exceed 100: %d", historyCount)
	}
	return nil
}