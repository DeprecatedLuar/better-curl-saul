# Implementation Plan

<!-- WORKFLOW IMPLEMENTATION GUIDE:
- This file contains active phases for implementation (completed phases moved to implementation-history.md)
- Each phase = one focused session until stable git commit (can be broken down into subphases), follow top-to-bottom order
- Focus on actionable steps: "Update file X, add function Y"
- Avoid verbose explanations - just implement what's specified and valuable
- Success criteria must be testable
- Make sure to test implementation after conclusion of phase
- Stop implementation and call out ideas if you find better approaches during implementation
-->

## Phase 7A: Response Field Extraction Feature

**Objective**: Implement field extraction from HTTP response history (e.g., `saul get response1 body`, `saul get response headers`)

**Strategic Decision**: Response Field Extraction chosen over Flag System due to:
- Lower implementation complexity (90% existing infrastructure reuse)
- Minimal risk (zero breaking changes)
- Quick user value delivery
- Foundation building for future features

### Implementation Steps

#### Step 0: Add Error Constants to Central Messages
**File**: `src/modules/display/messages.go`
**Location**: After line 32 (before warning constants)

**Add Error Constants**:
```go
	ErrFieldNameRequired   = "Listen, counselor - need to specify what field you want! Options are: body, headers, status, url, method, duration"
	ErrUnknownResponseField = "That field '%s'? Not in my case files! Stick to the evidence: body, headers, status, url, method, duration"
	ErrResponseProcessFailed = "Response processing went sideways - technical difficulties in the evidence room: %v"
```

#### Step 1: Extend Get Command Parser Logic
**File**: `src/project/handlers/commands/get.go`
**Function**: `Get()` around line 26

**Current Logic**:
```go
if strings.ToLower(cmd.Target) == "response" {
    return getResponse(cmd)
}
```

**Update To**:
```go
if strings.ToLower(cmd.Target) == "response" {
    return getResponse(cmd)
} else if strings.HasPrefix(strings.ToLower(cmd.Target), "response") && len(cmd.Target) > 8 {
    // Handle response1, response2, etc. with field extraction
    return getResponseWithField(cmd)
}
```

#### Step 2: Create Response Field Extraction Function
**File**: `src/project/handlers/commands/get.go`
**Location**: After `getResponse()` function (around line 122)

**Add Function**:
```go
// getResponseWithField handles response field extraction (e.g., response1 body, response2 headers)
func getResponseWithField(cmd core.Command) error {
    // Extract response number from target (response1 -> 1)
    numberStr := cmd.Target[8:] // Remove "response" prefix
    number, err := ParseResponseNumber(numberStr, cmd.Preset)
    if err != nil {
        return err
    }

    // Check if field is specified
    if len(cmd.KeyValuePairs) == 0 || cmd.KeyValuePairs[0].Key == "" {
        return fmt.Errorf(display.ErrFieldNameRequired)
    }

    fieldName := strings.ToLower(cmd.KeyValuePairs[0].Key)

    // Load the response
    response, err := presets.LoadHistoryResponse(cmd.Preset, number)
    if err != nil {
        return err
    }

    // Extract and display the requested field
    return displayResponseField(response, fieldName, cmd.RawOutput)
}
```

#### Step 3: Create Field Display Function
**File**: `src/project/handlers/commands/get.go`
**Location**: After `getResponseWithField()` function

**Add Function**:
```go
// displayResponseField extracts and displays a specific field from response
func displayResponseField(response *presets.HistoryResponse, fieldName string, rawOutput bool) error {
    switch fieldName {
    case "body":
        // Convert body to TOML using existing formatAsToml function
        if response.Body == nil {
            fmt.Println("(empty body)")
            return nil
        }

        // Marshal body to JSON bytes for formatAsToml
        bodyJSON, err := json.Marshal(response.Body)
        if err != nil {
            return fmt.Errorf(display.ErrResponseProcessFailed, err)
        }

        // Use existing formatAsToml function from http package
        if tomlFormatted := http.FormatAsToml(bodyJSON); tomlFormatted != "" {
            fmt.Print(tomlFormatted)
        } else {
            // Fallback to JSON if TOML conversion fails
            if prettyJSON, err := json.MarshalIndent(response.Body, "", "  "); err == nil {
                fmt.Print(string(prettyJSON))
            } else {
                fmt.Print(response.Body)
            }
        }

    case "headers":
        // Convert headers to TOML
        if response.Headers == nil {
            fmt.Println("(no headers)")
            return nil
        }

        headersJSON, err := json.Marshal(response.Headers)
        if err != nil {
            return fmt.Errorf(display.ErrResponseProcessFailed, err)
        }

        if tomlFormatted := http.FormatAsToml(headersJSON); tomlFormatted != "" {
            fmt.Print(tomlFormatted)
        } else {
            if prettyJSON, err := json.MarshalIndent(response.Headers, "", "  "); err == nil {
                fmt.Print(string(prettyJSON))
            } else {
                fmt.Print(response.Headers)
            }
        }

    case "status":
        fmt.Println(response.Status)

    case "url":
        fmt.Println(response.URL)

    case "method":
        fmt.Println(response.Method)

    case "duration":
        fmt.Println(response.Duration)

    default:
        return fmt.Errorf(display.ErrUnknownResponseField, fieldName)
    }

    return nil
}
```

#### Step 4: Make FormatAsToml Function Accessible
**File**: `src/project/handlers/http/response.go`
**Function**: `formatAsToml()` around line 118

**Current**: `func formatAsToml(jsonData []byte) string` (private)
**Change To**: `func FormatAsToml(jsonData []byte) string` (public)

**Update Function Name**: Change `formatAsToml` to `FormatAsToml` on line 118

**Update Calls**: Update all internal calls from `formatAsToml` to `FormatAsToml`:
- Line 54: `if tomlFormatted := FormatAsToml(filteredBody); tomlFormatted != "" {`
- Line 220: `if tomlFormatted := FormatAsToml(filteredBody); tomlFormatted != "" {`

#### Step 5: Add Required Imports (Proper Grouping)
**File**: `src/project/handlers/commands/get.go`
**Location**: Import section (around line 3)

**Update Imports** (stdlib first, then local packages):
```go
import (
    "encoding/json"  // Add this line (stdlib group)
    "fmt"
    "strings"

    "github.com/DeprecatedLuar/better-curl-saul/src/modules/display"
    "github.com/DeprecatedLuar/better-curl-saul/src/project/core"
    "github.com/DeprecatedLuar/better-curl-saul/src/project/handlers/http" // Add this line
    "github.com/DeprecatedLuar/better-curl-saul/src/project/presets"
)
```

### Success Criteria

#### Functional Tests
1. **Basic Field Extraction**:
   ```bash
   # Setup test data first
   go run cmd/main.go pokeapi set url https://pokeapi.co/api/v2/pokemon/pikachu
   go run cmd/main.go pokeapi call

   # Test field extraction
   go run cmd/main.go pokeapi get response1 body      # Should show TOML formatted body
   go run cmd/main.go pokeapi get response1 headers   # Should show TOML formatted headers
   go run cmd/main.go pokeapi get response1 status    # Should show "200 OK"
   go run cmd/main.go pokeapi get response1 url       # Should show full URL
   go run cmd/main.go pokeapi get response1 method    # Should show "GET"
   go run cmd/main.go pokeapi get response1 duration  # Should show timing like "0.234s"
   ```

2. **Most Recent Response Support**:
   ```bash
   go run cmd/main.go pokeapi get response body       # Should work same as response1
   ```

3. **Error Handling**:
   ```bash
   go run cmd/main.go pokeapi get response1 invalid   # Should show "Saul Goodman" style unknown field error
   go run cmd/main.go pokeapi get response1           # Should show "Saul Goodman" style field required error
   go run cmd/main.go pokeapi get response99 body     # Should show response not found error
   ```

4. **Zero Breaking Changes**:
   ```bash
   go run cmd/main.go pokeapi get response            # Should work exactly as before
   go run cmd/main.go pokeapi get response 1          # Should work exactly as before
   go run cmd/main.go pokeapi get history             # Should work exactly as before
   ```

#### Code Quality Checks
1. **No Regressions**: All existing `get` commands must work identically
2. **Error Messages**: Use centralized error constants with "Saul Goodman" personality
3. **Import Grouping**: Follow stdlib-first, then local packages pattern
4. **Output Format**: TOML for complex fields (body, headers), plain text for simple fields
5. **Function Length**: New functions should follow 250-line limit per codebase standards

### Completion Verification
- [ ] All functional tests pass
- [ ] No breaking changes to existing functionality
- [ ] Code follows existing patterns and conventions
- [ ] Import statements follow proper grouping (stdlib first, then local)
- [ ] Error handling uses centralized display module constants
- [ ] Error messages maintain "Saul Goodman" personality

### Notes
- Leverages existing `LoadHistoryResponse()`, `ParseResponseNumber()`, `GetMostRecentResponseNumber()` functions
- Reuses proven `formatAsToml()` conversion logic from response display
- No parser changes needed - existing logic already handles compound targets
- Implementation focused on single responsibility: field extraction from stored responses