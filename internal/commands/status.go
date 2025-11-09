package commands

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// Status displays the current preset configuration summary
func Status(cmd parser.Command, sessionManager *workspace.SessionManager) error {
	if !sessionManager.HasCurrentPreset() {
		return fmt.Errorf(display.ErrNoActivePreset)
	}

	preset := sessionManager.GetCurrentPreset()

	// Load request.toml
	requestHandler, err := workspace.LoadPresetFile(preset, "request")
	if err != nil {
		return fmt.Errorf(display.ErrRequestConfigFailed, err)
	}

	// Extract request details
	url := requestHandler.GetAsString("url")
	method := requestHandler.GetAsString("method")
	timeout := requestHandler.GetAsString("timeout")

	// Default values
	if url == "" {
		url = "(not set)"
	}
	if method == "" {
		method = "GET"
	}
	if timeout == "" {
		timeout = "30s"
	}

	// Count variables
	varCount := 0
	if varHandler, err := workspace.LoadPresetFile(preset, "variables"); err == nil {
		varCount = len(varHandler.Keys())
	}

	// Count headers
	headerCount := 0
	if headerHandler, err := workspace.LoadPresetFile(preset, "headers"); err == nil {
		headerCount = len(headerHandler.Keys())
	}

	// Count query params
	queryCount := 0
	if queryHandler, err := workspace.LoadPresetFile(preset, "query"); err == nil {
		queryCount = len(queryHandler.Keys())
	}

	// Count body keys
	bodyCount := 0
	if bodyHandler, err := workspace.LoadPresetFile(preset, "body"); err == nil {
		bodyCount = len(bodyHandler.Keys())
	}

	// Count history responses
	historyCount := 0
	historyPath, err := workspace.GetHistoryPath(preset)
	if err == nil {
		if entries, err := os.ReadDir(historyPath); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					historyCount++
				}
			}
		}
	}

	// Display status
	display.Info(fmt.Sprintf("Status: %s", preset))
	display.Plain("")
	display.Plain(fmt.Sprintf("  URL:     %s", url))
	display.Plain(fmt.Sprintf("  Method:  %s", method))
	display.Plain(fmt.Sprintf("  Timeout: %s", timeout))
	display.Plain("")
	display.Plain(fmt.Sprintf("  Variables: %d", varCount))
	display.Plain(fmt.Sprintf("  Headers:   %d", headerCount))
	display.Plain(fmt.Sprintf("  Query:     %d", queryCount))
	display.Plain(fmt.Sprintf("  Body keys: %d", bodyCount))
	display.Plain(fmt.Sprintf("  History:   %d", historyCount))

	return nil
}
