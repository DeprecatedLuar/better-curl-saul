package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/better-curl-saul/internal"
	"github.com/DeprecatedLuar/better-curl-saul/internal/commands/parser"
	"github.com/DeprecatedLuar/better-curl-saul/internal/workspace"
	"github.com/DeprecatedLuar/better-curl-saul/pkg/display"
)

// normalizeVariableSyntax strips leading @ from patterns like @{?name} or @{@name}
// This handles curl's file upload syntax that users might use in JSON
func normalizeVariableSyntax(value string) string {
	// Strip leading @ from @{?...} or @{@...} patterns
	if strings.HasPrefix(value, "@{") && (strings.Contains(value, "{?") || strings.Contains(value, "{@")) {
		return value[1:] // Remove the leading @
	}
	return value
}

// Import handles importing curl commands into preset TOML files
func Import(cmd parser.Command) error {
	if cmd.Preset == "" {
		return fmt.Errorf(display.ErrPresetNameRequired)
	}

	// Create a temporary file with instructions
	tmpFile, err := os.CreateTemp("", "saul-import-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	tmpFile.Close()

	// Detect and launch editor
	editor := detectEditor()
	if editor == "" {
		return fmt.Errorf(display.ErrEditorNotFound)
	}

	// Launch editor
	editorCmd := exec.Command(editor, tmpPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf(display.ErrEditorFailed, err)
	}

	// Read the edited content
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited content: %v", err)
	}

	// Extract curl command
	var curlLines []string
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			curlLines = append(curlLines, line)
		}
	}

	if len(curlLines) == 0 {
		return fmt.Errorf("no curl command found in editor")
	}

	curlCmd := strings.Join(curlLines, " ")

	// Save raw curl command as backup
	presetPath, _ := workspace.GetPresetPath(cmd.Preset)
	if presetPath == "" {
		workspace.CreatePresetDirectory(cmd.Preset)
		presetPath, _ = workspace.GetPresetPath(cmd.Preset)
	}
	backupPath := filepath.Join(presetPath, ".curl-backup")
	if err := os.WriteFile(backupPath, []byte(curlCmd), 0644); err != nil {
		display.Warning(fmt.Sprintf("Could not save curl backup: %v", err))
	}

	// Parse the curl command using existing parser
	parsed, err := internal.ImportCurlString(curlCmd)
	if err != nil {
		return fmt.Errorf("failed to parse curl command: %v\nRaw curl saved to: %s", err, backupPath)
	}

	// Show what was parsed for validation
	display.Info("Parsed curl command:")
	if parsed.Method != "" {
		fmt.Printf("  Method: %s\n", parsed.Method)
	}
	if parsed.BaseURL != "" {
		fmt.Printf("  URL: %s\n", parsed.BaseURL)
	}
	if len(parsed.Headers) > 0 {
		fmt.Printf("  Headers: %d\n", len(parsed.Headers))
	}
	if len(parsed.Query) > 0 {
		fmt.Printf("  Query params: %d\n", len(parsed.Query))
	}
	if parsed.Body != "" {
		bodyLines := strings.Split(parsed.Body, "\n")
		if len(bodyLines) > 1 {
			fmt.Printf("  Body: %d fields\n", len(bodyLines))
		} else {
			fmt.Printf("  Body: %d bytes\n", len(parsed.Body))
		}
	}

	// Save parsed data to TOML files
	if err := saveParsedCurl(cmd.Preset, parsed); err != nil {
		return fmt.Errorf("failed to save parsed curl: %v\nRaw curl saved to: %s", err, backupPath)
	}

	display.Success(fmt.Sprintf("Curl command imported to preset '%s'", cmd.Preset))
	return nil
}

// saveParsedCurl saves a parsed curl request to preset TOML files
func saveParsedCurl(preset string, req *internal.CurlRequest) error {
	// Ensure preset directory exists
	presetPath, err := workspace.GetPresetPath(preset)
	if err != nil {
		// Create the preset directory if it doesn't exist
		if err := workspace.CreatePresetDirectory(preset); err != nil {
			return fmt.Errorf("failed to create preset: %v", err)
		}
		presetPath, _ = workspace.GetPresetPath(preset)
	}

	// Clear existing TOML files to implement overwrite behavior on re-import
	targets := []string{"request", "headers", "query", "body", "variables"}
	for _, target := range targets {
		targetPath, _ := workspace.GetVariantPath(preset, target)
		if targetPath != "" {
			os.Remove(targetPath) // Ignore errors - file might not exist
		}
	}

	// Save request.toml (method and URL)
	if req.Method != "" || req.BaseURL != "" {
		handler, _ := workspace.LoadPresetFile(preset, "request")
		if req.Method != "" {
			handler.Set("method", strings.ToUpper(req.Method))
		}
		if req.BaseURL != "" {
			handler.Set("url", req.BaseURL)
		}
		if err := workspace.SavePresetFile(preset, "request", handler); err != nil {
			return err
		}
	}

	// Save headers.toml
	if len(req.Headers) > 0 {
		handler, _ := workspace.LoadPresetFile(preset, "headers")
		for key, value := range req.Headers {
			// Normalize variable syntax
			normalizedVal := normalizeVariableSyntax(value)

			// Detect if value is a variable
			isVar, varType, varName := workspace.DetectVariableType(normalizedVal)
			if isVar {
				err := workspace.StoreVariableInfo(preset, "headers."+key, varType, varName)
				if err != nil {
					display.Warning(fmt.Sprintf("Failed to store variable info for %s: %v", key, err))
				}
			}
			handler.Set(key, normalizedVal)
		}
		if err := workspace.SavePresetFile(preset, "headers", handler); err != nil {
			return err
		}
	}

	// Save query.toml
	if len(req.Query) > 0 {
		handler, _ := workspace.LoadPresetFile(preset, "query")
		for key, value := range req.Query {
			// Normalize variable syntax
			normalizedVal := normalizeVariableSyntax(value)

			// Detect if value is a variable
			isVar, varType, varName := workspace.DetectVariableType(normalizedVal)
			if isVar {
				err := workspace.StoreVariableInfo(preset, "query."+key, varType, varName)
				if err != nil {
					display.Warning(fmt.Sprintf("Failed to store variable info for %s: %v", key, err))
				}
			}
			handler.Set(key, normalizedVal)
		}
		if err := workspace.SavePresetFile(preset, "query", handler); err != nil {
			return err
		}
	}

	// Save body.toml
	if req.Body != "" {
		bodyPath := filepath.Join(presetPath, "body.toml")
		handler, _ := workspace.LoadPresetFile(preset, "body")

		// Try to parse as JSON first
		var bodyMap map[string]interface{}
		if err := json.Unmarshal([]byte(req.Body), &bodyMap); err == nil {
			// Valid JSON - convert to TOML and save
			for key, value := range bodyMap {
				// Check if value is a string variable
				if strVal, ok := value.(string); ok {
					// Normalize variable syntax: strip leading @ from @{?...} or @{@...}
					normalizedVal := normalizeVariableSyntax(strVal)

					isVar, varType, varName := workspace.DetectVariableType(normalizedVal)
					if isVar {
						// Store variable info in config.toml
						err := workspace.StoreVariableInfo(preset, "body."+key, varType, varName)
						if err != nil {
							display.Warning(fmt.Sprintf("Failed to store variable info for %s: %v", key, err))
						}
					}

					// Store the normalized value
					handler.Set(key, normalizedVal)
				} else {
					// Non-string value, store as-is
					handler.Set(key, value)
				}
			}
			if err := workspace.SavePresetFile(preset, "body", handler); err != nil {
				return err
			}
		} else {
			// Not JSON - check if it's form data (key=value format)
			lines := strings.Split(req.Body, "\n")
			isFormData := true
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// Form data format: "key=value" or "file=@path"
				if !strings.Contains(line, "=") {
					isFormData = false
					break
				}
			}

			if isFormData {
				// Parse and save form data as TOML key-value pairs
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						// Normalize variable syntax
						normalizedVal := normalizeVariableSyntax(value)

						// Detect if value is a variable
						isVar, varType, varName := workspace.DetectVariableType(normalizedVal)
						if isVar {
							// Store variable info in config.toml
							err := workspace.StoreVariableInfo(preset, "body."+key, varType, varName)
							if err != nil {
								display.Warning(fmt.Sprintf("Failed to store variable info for %s: %v", key, err))
							}
						}

						handler.Set(key, normalizedVal)
					}
				}
				if err := workspace.SavePresetFile(preset, "body", handler); err != nil {
					return err
				}
			} else {
				// Raw body - save as comment
				rawBodyComment := fmt.Sprintf("# Raw body (non-JSON/form):\n# %s\n", strings.ReplaceAll(req.Body, "\n", "\n# "))
				if err := os.WriteFile(bodyPath, []byte(rawBodyComment), 0644); err != nil {
					return fmt.Errorf("failed to save body: %v", err)
				}
			}
		}
	}

	return nil
}
