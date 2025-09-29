package workspace

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/core"
)

// ImportCurlString converts a curl command string into TOML preset files
func ImportCurlString(preset string, curlCmd string) error {
	// Parse curl command
	result, err := core.ParseCurl(curlCmd)
	if err != nil {
		return fmt.Errorf("failed to parse curl command: %v", err)
	}

	// Validate URL exists
	if result.URL == "" {
		return fmt.Errorf("no URL found in curl command")
	}

	// Ensure preset directory exists
	err = CreatePresetDirectory(preset)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create preset directory: %v", err)
	}

	// Convert body (JSON → TOML)
	if result.Body != "" {
		bodyHandler, err := NewTomlHandlerFromJSON([]byte(result.Body))
		if err != nil {
			return fmt.Errorf("invalid JSON body: %v", err)
		}
		err = SavePresetFile(preset, "body", bodyHandler)
		if err != nil {
			return fmt.Errorf("failed to save body: %v", err)
		}
	}

	// Convert headers
	if len(result.Headers) > 0 {
		headersHandler, err := LoadPresetFile(preset, "headers")
		if err != nil {
			return fmt.Errorf("failed to load headers file: %v", err)
		}
		for key, val := range result.Headers {
			headersHandler.Set(key, val)
		}
		err = SavePresetFile(preset, "headers", headersHandler)
		if err != nil {
			return fmt.Errorf("failed to save headers: %v", err)
		}
	}

	// Convert query params
	if len(result.Query) > 0 {
		queryHandler, err := LoadPresetFile(preset, "query")
		if err != nil {
			return fmt.Errorf("failed to load query file: %v", err)
		}
		for key, val := range result.Query {
			queryHandler.Set(key, val)
		}
		err = SavePresetFile(preset, "query", queryHandler)
		if err != nil {
			return fmt.Errorf("failed to save query: %v", err)
		}
	}

	// Convert request (method, baseURL without query params)
	requestHandler, err := LoadPresetFile(preset, "request")
	if err != nil {
		return fmt.Errorf("failed to load request file: %v", err)
	}
	requestHandler.Set("method", result.Method)
	requestHandler.Set("url", result.BaseURL)
	err = SavePresetFile(preset, "request", requestHandler)
	if err != nil {
		return fmt.Errorf("failed to save request: %v", err)
	}

	return nil
}