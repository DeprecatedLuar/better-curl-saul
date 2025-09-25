# Implementation Plan

<!-- WORKFLOW IMPLEMENTATION GUIDE:
- This file contains active phases for implementation (completed phases moved to implementation-history.md)
- Each phase = one focused session until stable git commit, follow top-to-bottom order
- SPLIT complex phases (>5 steps) into subphases for safety, testing, and incremental value
- Avoid verbose explanations - just implement what's specified and valuable
- Focus on actionable steps: "Update file X, add function Y"
- Success criteria must be testable
- Test after each (sub)phase completion
-->


# Phase 8A: Flag Infrastructure

**Dependencies**: Phase 7A
**Value**: Foundation for all flag features

---

## Step 1: Extend Command Structure
**File**: `src/project/core/parser.go`  
**Location**: Lines 13-23 (Command struct)

**Current**:
```go
type Command struct {
    // ... existing fields ...
    RawOutput     bool          // For --raw flag
}
```

**Update To**:
```go
type Command struct {
    // ... existing fields ...
    RawOutput        bool     // For --raw flag
    VariableFlags    []string // -v var1 var2 var3 (space-separated variables to prompt)
    ResponseFormat   string   // --headers-only, --body-only, --status-only
    DryRun          bool     // --dry-run
}
```

---

## Step 2: Extend Flag Parser
**File**: `src/project/core/parser.go`  
**Function**: `parseFlags()` (lines 200-219)

**Current**:
```go
switch arg {
case "--raw":
    cmd.RawOutput = true
default:
    return nil, fmt.Errorf("unknown flag: %s", arg)
}
```

**Update To**:
```go
switch arg {
case "--raw":
    cmd.RawOutput = true
case "--headers-only":
    cmd.ResponseFormat = "headers-only"
case "--body-only":
    cmd.ResponseFormat = "body-only"
case "--status-only":
    cmd.ResponseFormat = "status-only"
case "--dry-run":
    cmd.DryRun = true
default:
    return nil, fmt.Errorf("unknown flag: %s", arg)
}
```

**Add Short Flag Parsing** (after existing switch, before filteredArgs):
```go
} else if strings.HasPrefix(arg, "-") && len(arg) > 1 && !strings.HasPrefix(arg, "--") {
    // Handle short flags
    flagPart := arg[1:] // Remove leading -
    if flagPart == "v" {
        // -v flag: collect all following non-flag arguments as variable names
        for j := i + 1; j < len(args); j++ {
            if strings.HasPrefix(args[j], "-") {
                break // Stop at next flag
            }
            cmd.VariableFlags = append(cmd.VariableFlags, args[j])
            i = j // Skip these args in main loop
        }
        // If no variables specified, empty slice signals "all variables"
        if len(cmd.VariableFlags) == 0 {
            cmd.VariableFlags = []string{}
        }
    } else {
        return nil, fmt.Errorf("unknown flag: %s", arg)
    }
```

---

# Phase 8B: Dry-Run Feature

**Dependencies**: Phase 8A
**Value**: Preview requests without execution

---

## Step 3: Implement Dry Run Display Function
**File**: `src/project/handlers/http.go`  
**Location**: After `storeResponseHistory()` function (around line 160)

**Add Function**:
```go
// displayDryRunRequest shows request details without executing
func displayDryRunRequest(request *http.HTTPRequestConfig) error {
    fmt.Printf("%s %s\n", request.Method, request.URL)

    if len(request.Headers) > 0 {
        fmt.Println("Headers:")
        for key, value := range request.Headers {
            fmt.Printf("  %s: %s\n", key, value)
        }
    }

    if request.Body != nil && len(request.Body) > 0 {
        fmt.Println("Body:")
        fmt.Println("  " + strings.Replace(string(request.Body), "\n", "\n  ", -1))
    }

    if len(request.Query) > 0 {
        fmt.Println("Query Parameters:")
        for key, value := range request.Query {
            fmt.Printf("  %s: %s\n", key, value)
        }
    }

    fmt.Println("\n(Request not sent - dry run mode)")
    return nil
}
```

---

## Step 4: Add Dry Run Logic to Call Command Handler  
**File**: `src/project/handlers/http.go`  
**Function**: `ExecuteCallCommand()` (after request building, around line 67)

**Current**:
```go
// Build HTTP request components explicitly - no guessing
request, err := http.BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
if err != nil {
    return fmt.Errorf(display.ErrRequestBuildFailed)
}

// Execute the HTTP request
response, err := http.ExecuteHTTPRequest(request)
```

**Update To**:
```go
// Build HTTP request components explicitly - no guessing
request, err := http.BuildHTTPRequestFromHandlers(requestHandler, headersHandler, bodyHandler, queryHandler)
if err != nil {
    return fmt.Errorf(display.ErrRequestBuildFailed)
}

// Handle dry-run mode
if cmd.DryRun {
    return displayDryRunRequest(request)
}

// Execute the HTTP request (only if not dry-run)
response, err := http.ExecuteHTTPRequest(request)
```

---

## Step 5: Add Missing Import
**File**: `src/project/handlers/http.go`  
**Location**: Top of file with other imports

**Add**:
```go
import (
    "fmt"
    "os"
    "strconv"
    "strings"  // Add this import

    "github.com/go-resty/resty/v2"
    // ... other imports
)
```

---

# Phase 8C: Variable Management

**Dependencies**: Phase 7A (better prompting UX) + Phase 8A (flag parsing)
**Value**: Selective variable prompting with -v flag

---

## Step 6: Variable Management Integration
**File**: `src/project/handlers/variables/prompting.go`  
**Function**: Add new function after `PromptForVariables()` (around line 93)

**Add Function**:
```go
// PromptForSpecificVariables prompts only for specified variables
func PromptForSpecificVariables(preset string, variableNames []string, persist bool) (map[string]string, error) {
    scanner := bufio.NewScanner(os.Stdin)
    substitutions := make(map[string]string)

    // Load variables.toml to get hard variables
    variablesHandler, err := presets.LoadPresetFile(preset, "variables")
    if err != nil {
        return nil, fmt.Errorf(display.ErrVariableLoadFailed)
    }

    // Find all variables across all TOML files
    allVariables, err := FindAllVariables(preset)
    if err != nil {
        return nil, fmt.Errorf(display.ErrVariableLoadFailed)
    }

    // Filter to only requested variables
    var targetVariables []VariableInfo
    for _, variable := range allVariables {
        for _, requestedName := range variableNames {
            if variable.Key == requestedName || variable.Name == requestedName {
                targetVariables = append(targetVariables, variable)
                break
            }
        }
    }

    // Use same prompting logic as PromptForVariables but on filtered set
    for _, variable := range targetVariables {
        var prompt string
        var currentValue string

        if variable.Type == "hard" {
            // Hard variables: use stored value if exists, show for editing
            currentValue = variablesHandler.GetAsString(variable.Key)
            if variable.Name != "" {
                prompt = variable.Name + " [" + currentValue + "]: "
            } else {
                prompt = variable.Key + " [" + currentValue + "]: "
            }
        } else {
            // Soft variables: always prompt with empty input
            if variable.Name != "" {
                prompt = variable.Name + ": "
            } else {
                prompt = variable.Key + ": "
            }
        }

        fmt.Print(prompt)
        if scanner.Scan() {
            userInput := strings.TrimSpace(scanner.Text())

            if variable.Type == "hard" && userInput == "" && currentValue != "" {
                // Keep existing value for hard variables if user presses Enter
                substitutions[variable.Key] = currentValue
            } else if userInput != "" {
                substitutions[variable.Key] = userInput

                // Save hard variables to variables.toml
                if variable.Type == "hard" {
                    variablesHandler.Set(variable.Key, userInput)
                    err := presets.SavePresetFile(preset, "variables", variablesHandler)
                    if err != nil {
                        return nil, fmt.Errorf(display.ErrVariableSaveFailed)
                    }
                }
            }
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf(display.ErrVariableLoadFailed)
    }

    return substitutions, nil
}
```

---

## Step 7: Update Call Command Variable Logic
**File**: `src/project/handlers/http.go`  
**Function**: `ExecuteCallCommand()` (around line 37)

**Current**:
```go
substitutions, err := PromptForVariables(cmd.Preset, persist)
```

**Update To**:
```go
var substitutions map[string]string
var err error

if len(cmd.VariableFlags) > 0 {
    // Specific variables requested via -v flag
    substitutions, err = variables.PromptForSpecificVariables(cmd.Preset, cmd.VariableFlags, persist)
} else {
    // Normal variable prompting
    substitutions, err = PromptForVariables(cmd.Preset, persist)
}
```

---

## Step 8: Update Variables Handler Export
**File**: `src/project/handlers/variables.go`  
**Location**: Around line 13

**Current**:
```go
var PromptForVariables = variables.PromptForVariables
```

**Update To**:
```go
var PromptForVariables = variables.PromptForVariables
var PromptForSpecificVariables = variables.PromptForSpecificVariables
```

---

# Phase 8D: Response Formatting

**Dependencies**: Phase 8A
**Value**: Clean output filtering with --headers-only, --body-only, --status-only

---

## Step 9: Response Format Configuration System
**File**: `src/project/handlers/http/response.go`  
**Function**: `DisplayResponse()` (lines 17-92)

**Current**:
```go
func DisplayResponse(response *resty.Response, rawMode bool, preset string) {
```

**Update To**:
```go
func DisplayResponse(response *resty.Response, rawMode bool, preset string, responseFormat string) {
```

**Add Format Logic** (after line 25, before content preparation):
```go
// Handle response format overrides
if responseFormat != "" {
    displayFormattedResponse(response, responseFormat)
    return
}
```

**Add Format Display Function** (after `DisplayResponse()`):
```go
// displayFormattedResponse handles specific response format requests
func displayFormattedResponse(response *resty.Response, format string) {
    switch format {
    case "headers-only":
        for key, values := range response.Header() {
            if len(values) > 0 {
                fmt.Printf("%s: %s\n", key, values[0])
            }
        }
    case "body-only":
        fmt.Print(response.String())
    case "status-only":
        fmt.Println(response.Status())
    }
}
```

---

## Step 10: Update Call Site in HTTP Handler
**File**: `src/project/handlers/http.go`  
**Location**: Line 86

**Current**:
```go
http.DisplayResponse(response, rawMode, cmd.Preset)
```

**Update To**:
```go
http.DisplayResponse(response, rawMode, cmd.Preset, cmd.ResponseFormat)
```

---

## Step 11: Add Missing Import to Response Handler
**File**: `src/project/handlers/http/response.go`  
**Location**: Top of file with other imports

**Add**:
```go
import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/go-resty/resty/v2"
    // ... other existing imports
)
```

---

## Testing Commands

**Test dry-run:**
```bash
go run cmd/main.go testapi call --dry-run
```

**Test variable flags:**
```bash
go run cmd/main.go testapi call -v
go run cmd/main.go testapi call -v username token
```

**Test response formatting:**
```bash
go run cmd/main.go testapi call --headers-only
go run cmd/main.go testapi call --body-only
go run cmd/main.go testapi call --status-only
```

**Test flag combinations:**
```bash
go run cmd/main.go testapi call -v username --dry-run
go run cmd/main.go testapi call --body-only --raw
```