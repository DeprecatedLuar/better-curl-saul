package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DeprecatedLuar/better-curl-saul/src/project/commands"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/core"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/variables"
	"github.com/DeprecatedLuar/better-curl-saul/src/project/workspace"
)

func setupTestPreset(t *testing.T, name string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "saul-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	os.Setenv("SAUL_CONFIG_DIR_PATH", tempDir)
	os.Setenv("SAUL_APP_DIR_NAME", "saul")
	os.Setenv("SAUL_PRESETS_DIR_NAME", "presets")

	err = workspace.CreatePresetDirectory(name)
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

			handler, err := workspace.LoadPresetFile(preset, tt.setCmd.Target)
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
			err := commands.ValidateRequestField("method", tt.method)
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

	presetPath, _ := workspace.GetPresetPath(preset)

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
			result := commands.InferValueType(tt.value)

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

	handler, err := workspace.LoadPresetFile(preset, "body")
	if err != nil {
		t.Fatalf("LoadPresetFile failed: %v", err)
	}

	handler.Set("user.name", "alice")
	handler.Set("user.age", int64(30))
	handler.Set("user.active", true)
	handler.Set("tags", []string{"admin", "developer"})

	if err := workspace.SavePresetFile(preset, "body", handler); err != nil {
		t.Fatalf("SavePresetFile failed: %v", err)
	}

	loadedHandler, err := workspace.LoadPresetFile(preset, "body")
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

func TestImportCurlString(t *testing.T) {
	preset, cleanup := setupTestPreset(t, "curltest")
	defer cleanup()

	tests := []struct {
		name     string
		curlCmd  string
		validate func(t *testing.T, preset string)
	}{
		{
			name:    "basic POST with JSON body",
			curlCmd: `curl -X POST 'https://api.example.com/users' -d '{"name":"test","age":25}'`,
			validate: func(t *testing.T, preset string) {
				// Check request file
				reqHandler, err := workspace.LoadPresetFile(preset, "request")
				if err != nil {
					t.Fatalf("LoadPresetFile request failed: %v", err)
				}
				if method := reqHandler.Get("method"); method != "POST" {
					t.Errorf("method = %v, want POST", method)
				}
				if url := reqHandler.Get("url"); url != "https://api.example.com/users" {
					t.Errorf("url = %v, want https://api.example.com/users", url)
				}

				// Check body file
				bodyHandler, err := workspace.LoadPresetFile(preset, "body")
				if err != nil {
					t.Fatalf("LoadPresetFile body failed: %v", err)
				}
				if name := bodyHandler.Get("name"); name != "test" {
					t.Errorf("body.name = %v, want test", name)
				}
				if age := bodyHandler.Get("age"); age != float64(25) {
					t.Errorf("body.age = %v, want 25", age)
				}
			},
		},
		{
			name:    "GET with query params and headers",
			curlCmd: `curl -X GET 'https://api.example.com/search?q=golang&limit=10' -H 'Authorization: Bearer token123' -H 'Accept: application/json'`,
			validate: func(t *testing.T, preset string) {
				// Check request file
				reqHandler, err := workspace.LoadPresetFile(preset, "request")
				if err != nil {
					t.Fatalf("LoadPresetFile request failed: %v", err)
				}
				if method := reqHandler.Get("method"); method != "GET" {
					t.Errorf("method = %v, want GET", method)
				}
				if url := reqHandler.Get("url"); url != "https://api.example.com/search" {
					t.Errorf("url = %v, want https://api.example.com/search (without query)", url)
				}

				// Check query params file
				queryHandler, err := workspace.LoadPresetFile(preset, "query")
				if err != nil {
					t.Fatalf("LoadPresetFile query failed: %v", err)
				}
				if q := queryHandler.Get("q"); q != "golang" {
					t.Errorf("query.q = %v, want golang", q)
				}
				if limit := queryHandler.Get("limit"); limit != "10" {
					t.Errorf("query.limit = %v, want 10", limit)
				}

				// Check headers file
				headersHandler, err := workspace.LoadPresetFile(preset, "headers")
				if err != nil {
					t.Fatalf("LoadPresetFile headers failed: %v", err)
				}
				if auth := headersHandler.Get("Authorization"); auth != "Bearer token123" {
					t.Errorf("headers.Authorization = %v, want Bearer token123", auth)
				}
				if accept := headersHandler.Get("Accept"); accept != "application/json" {
					t.Errorf("headers.Accept = %v, want application/json", accept)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := workspace.ImportCurlString(preset, tt.curlCmd)
			if err != nil {
				t.Fatalf("ImportCurlString failed: %v", err)
			}

			tt.validate(t, preset)
		})
	}
}