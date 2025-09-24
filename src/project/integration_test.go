package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/core"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/handlers"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/handlers/commands"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/handlers/variables"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)

func setupTestPreset(t *testing.T, name string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "saul-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	os.Setenv("SAUL_CONFIG_DIR_PATH", tempDir)
	os.Setenv("SAUL_APP_DIR_NAME", "saul")
	os.Setenv("SAUL_PRESETS_DIR_NAME", "presets")

	err = presets.CreatePresetDirectory(name)
	if err != nil {
		t.Fatalf("failed to create preset: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
		os.Unsetenv("SAUL_CONFIG_DIR_PATH")
		os.Unsetenv("SAUL_APP_DIR_NAME")
		os.Unsetenv("SAUL_PRESETS_DIR_NAME")
	}

	return name, cleanup
}

func TestSetAndGetFlow(t *testing.T) {
	preset, cleanup := setupTestPreset(t, "test")
	defer cleanup()

	tests := []struct {
		name      string
		setCmd    core.Command
		getKey    string
		wantValue string
	}{
		{
			name: "set and get URL",
			setCmd: core.Command{
				Preset: preset,
				Target: "request",
				KeyValuePairs: []core.KeyValuePair{
					{Key: "url", Value: "https://api.example.com"},
				},
			},
			getKey:    "url",
			wantValue: "https://api.example.com",
		},
		{
			name: "set and get method (uppercase)",
			setCmd: core.Command{
				Preset: preset,
				Target: "request",
				KeyValuePairs: []core.KeyValuePair{
					{Key: "method", Value: "post"},
				},
			},
			getKey:    "method",
			wantValue: "POST",
		},
		{
			name: "set and get body field",
			setCmd: core.Command{
				Preset: preset,
				Target: "body",
				KeyValuePairs: []core.KeyValuePair{
					{Key: "user.name", Value: "testuser"},
				},
			},
			getKey:    "user.name",
			wantValue: "testuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := commands.Set(tt.setCmd); err != nil {
				t.Fatalf("Set failed: %v", err)
			}

			handler, err := presets.LoadPresetFile(preset, tt.setCmd.Target)
			if err != nil {
				t.Fatalf("LoadPresetFile failed: %v", err)
			}

			value := handler.Get(tt.getKey)

			if value != tt.wantValue {
				t.Errorf("got %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestHardVariableDetection(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantVar  bool
		wantType string
		wantName string
	}{
		{
			name:     "hard variable with name",
			value:    "{@token}",
			wantVar:  true,
			wantType: "hard",
			wantName: "token",
		},
		{
			name:     "bare hard variable",
			value:    "{@}",
			wantVar:  true,
			wantType: "hard",
			wantName: "",
		},
		{
			name:     "soft variable with name",
			value:    "{?username}",
			wantVar:  true,
			wantType: "soft",
			wantName: "username",
		},
		{
			name:     "bare soft variable",
			value:    "{?}",
			wantVar:  true,
			wantType: "soft",
			wantName: "",
		},
		{
			name:     "URL with @ (not a variable)",
			value:    "https://api.github.com/@octocat",
			wantVar:  false,
			wantType: "",
			wantName: "",
		},
		{
			name:     "regular string",
			value:    "hello world",
			wantVar:  false,
			wantType: "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isVar, varType, varName := variables.DetectVariableType(tt.value)

			if isVar != tt.wantVar {
				t.Errorf("isVariable = %v, want %v", isVar, tt.wantVar)
			}
			if varType != tt.wantType {
				t.Errorf("varType = %v, want %v", varType, tt.wantType)
			}
			if varName != tt.wantName {
				t.Errorf("varName = %v, want %v", varName, tt.wantName)
			}
		})
	}
}

func TestMethodValidation(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{"valid GET", "GET", false},
		{"valid POST", "POST", false},
		{"valid lowercase post", "post", false},
		{"valid PUT", "PUT", false},
		{"valid DELETE", "DELETE", false},
		{"invalid method", "INVALID", true},
		{"invalid method YEET", "YEET", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateRequestField("method", tt.method)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequestField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVariableStorageAndRetrieval(t *testing.T) {
	preset, cleanup := setupTestPreset(t, "vartest")
	defer cleanup()

	setCmd := core.Command{
		Preset: preset,
		Target: "body",
		KeyValuePairs: []core.KeyValuePair{
			{Key: "api.token", Value: "{@token}"},
			{Key: "api.key", Value: "{@apikey}"},
		},
	}

	if err := commands.Set(setCmd); err != nil {
		t.Fatalf("Set with variables failed: %v", err)
	}

	foundVars, err := variables.FindAllVariables(preset)
	if err != nil {
		t.Fatalf("FindAllVariables failed: %v", err)
	}

	if len(foundVars) != 2 {
		t.Errorf("expected 2 variables, got %d", len(foundVars))
	}

	hasToken := false
	hasApikey := false
	for _, v := range foundVars {
		if v.Name == "token" && v.Type == "hard" {
			hasToken = true
		}
		if v.Name == "apikey" && v.Type == "hard" {
			hasApikey = true
		}
	}

	if !hasToken {
		t.Error("token variable not found")
	}
	if !hasApikey {
		t.Error("apikey variable not found")
	}
}

func TestLazyFileCreation(t *testing.T) {
	preset, cleanup := setupTestPreset(t, "lazy")
	defer cleanup()

	presetPath, _ := presets.GetPresetPath(preset)

	files := []string{"body.toml", "headers.toml", "query.toml", "request.toml", "variables.toml"}
	for _, file := range files {
		filePath := filepath.Join(presetPath, file)
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("file %s should not exist yet (lazy creation)", file)
		}
	}

	setCmd := core.Command{
		Preset: preset,
		Target: "body",
		KeyValuePairs: []core.KeyValuePair{
			{Key: "test", Value: "value"},
		},
	}
	commands.Set(setCmd)

	bodyPath := filepath.Join(presetPath, "body.toml")
	if _, err := os.Stat(bodyPath); os.IsNotExist(err) {
		t.Error("body.toml should exist after Set operation")
	}

	headersPath := filepath.Join(presetPath, "headers.toml")
	if _, err := os.Stat(headersPath); !os.IsNotExist(err) {
		t.Error("headers.toml should not exist (not used)")
	}
}

func TestTargetAliasNormalization(t *testing.T) {
	tests := []struct {
		alias      string
		normalized string
	}{
		{"body", "body"},
		{"header", "headers"},
		{"headers", "headers"},
		{"query", "query"},
		{"queries", "query"},
		{"request", "request"},
		{"req", "request"},
		{"url", "request"},
		{"variables", "variables"},
		{"vars", "variables"},
		{"var", "variables"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			normalized := commands.NormalizeTarget(tt.alias)
			if normalized != tt.normalized {
				t.Errorf("NormalizeTarget(%s) = %s, want %s", tt.alias, normalized, tt.normalized)
			}
		})
	}
}

func TestArrayInference(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  interface{}
	}{
		{
			name:  "bracketed array",
			value: "[red,blue,green]",
			want:  []string{"red", "blue", "green"},
		},
		{
			name:  "single value",
			value: "single",
			want:  "single",
		},
		{
			name:  "string number",
			value: "42",
			want:  "42",
		},
		{
			name:  "boolean true",
			value: "true",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handlers.InferValueType(tt.value)

			switch expected := tt.want.(type) {
			case []string:
				resultSlice, ok := result.([]string)
				if !ok {
					t.Errorf("expected []string, got %T", result)
					return
				}
				if len(resultSlice) != len(expected) {
					t.Errorf("slice length = %d, want %d", len(resultSlice), len(expected))
				}
			case bool:
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			case string:
				if result != expected {
					t.Errorf("got %v, want %v", result, expected)
				}
			}
		})
	}
}

func TestGetCommandParsing(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantTarget string
		wantKey    string
	}{
		{
			name:       "get url (special request field)",
			args:       []string{"test", "get", "url"},
			wantTarget: "request",
			wantKey:    "url",
		},
		{
			name:       "get body field",
			args:       []string{"test", "get", "body", "user.name"},
			wantTarget: "body",
			wantKey:    "user.name",
		},
		{
			name:       "get history",
			args:       []string{"test", "get", "history"},
			wantTarget: "history",
			wantKey:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := core.ParseCommand(tt.args)
			if err != nil {
				t.Fatalf("ParseCommand failed: %v", err)
			}

			if cmd.Target != tt.wantTarget {
				t.Errorf("target = %s, want %s", cmd.Target, tt.wantTarget)
			}

			if len(cmd.KeyValuePairs) > 0 {
				if cmd.KeyValuePairs[0].Key != tt.wantKey {
					t.Errorf("key = %s, want %s", cmd.KeyValuePairs[0].Key, tt.wantKey)
				}
			} else if tt.wantKey != "" {
				t.Errorf("expected key %s, got none", tt.wantKey)
			}
		})
	}
}

func TestTOMLHandlerBasics(t *testing.T) {
	preset, cleanup := setupTestPreset(t, "tomltest")
	defer cleanup()

	handler, err := presets.LoadPresetFile(preset, "body")
	if err != nil {
		t.Fatalf("LoadPresetFile failed: %v", err)
	}

	handler.Set("user.name", "alice")
	handler.Set("user.age", int64(30))
	handler.Set("user.active", true)
	handler.Set("tags", []string{"admin", "developer"})

	if err := presets.SavePresetFile(preset, "body", handler); err != nil {
		t.Fatalf("SavePresetFile failed: %v", err)
	}

	loadedHandler, err := presets.LoadPresetFile(preset, "body")
	if err != nil {
		t.Fatalf("LoadPresetFile after save failed: %v", err)
	}

	if val := loadedHandler.Get("user.name"); val != "alice" {
		t.Errorf("user.name = %v, want alice", val)
	}

	if val := loadedHandler.Get("user.age"); val != int64(30) {
		t.Errorf("user.age = %v, want 30", val)
	}

	if val := loadedHandler.Get("user.active"); val != true {
		t.Errorf("user.active = %v, want true", val)
	}
}