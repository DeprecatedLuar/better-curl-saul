# Better-Curl (Saul) - Action Plan

## Project Overview
Comprehensive implementation plan for Better-Curl (Saul) - a workspace-based HTTP client that eliminates complex curl command pain through TOML-based configuration.

## Current State Analysis

### ✅ **Implemented**
- **Phase 1 Complete**: Foundation & TOML Integration
  - Modular Go structure following conventions
  - Command parsing system with global and preset commands  
  - Directory management with lazy file creation
  - TOML file operations integrated
- **Phase 2 Complete**: Core TOML Operations & Variable System
  - 5-file structure (Unix philosophy): body, headers, query, request, variables
  - Special request syntax: `set url/method/timeout` (no = sign)
  - Variable system: `@` for hard variables, `?` for soft variables *(needs syntax update)*
  - Target normalization and validation
  - Comprehensive test suite validation
- **Phase 3 Complete**: HTTP Execution Engine
  - `saul call preset` command fully functional
  - Variable prompting system (`@` hard variables, `?` soft variables) *(updated to braced syntax)*
  - TOML file merging *(replaced with separate handlers)*
  - HTTP client integration using go-resty
  - Support for all major HTTP methods
  - JSON body conversion and pretty-printed responses
  - Smart Variable Deduplication feature
- **Phase 3.5 Complete**: Architecture & Variable Syntax Fix
  - ✅ Separate handler implementation (no field misclassification)
  - ✅ Braced variable syntax `{@name}` and `{?name}` (no URL conflicts)
  - ✅ Real-world URL support: `https://api.github.com/@username` works correctly
  - ✅ Complex URLs with mixed literal and variable symbols supported
  - ✅ All existing functionality preserved with new syntax

### ❌ **Missing Core Components**
- **Response history system**: Storage, management, and access commands
- **Interactive mode**: Command shell for preset management
- **Advanced command system**: Enhanced help and management
- **Production readiness**: Cross-platform compatibility, error handling polish

### 🔧 **Technical Debt**
- No response history for debugging API interactions
- No interactive mode for workflow efficiency
- Container-level editing (Phase 4A.2) not yet implemented

## Implementation Phases

### **Phase 1: Foundation & TOML Integration** ✅ **COMPLETED**
*All functionality implemented and tested.*

### **Phase 2: Core TOML Operations & Variable System** ✅ **COMPLETED**  
*All functionality implemented and tested.*

### **Phase 3: HTTP Execution Engine** ✅ **COMPLETED**
*All functionality implemented and tested.*

---

### **Phase 3.5: HTTP Architecture & Variable Syntax Fix** ✅ **COMPLETED**
*Goal: Fix TOML merging logic AND variable syntax conflicts to enable real-world URL usage*

### **Phase 3.6: Variable System Critical Fix** ✅ **COMPLETED**
*Goal: Fix variable substitution lookup to enable proper prompting and eliminate URL corruption*

#### 3.6.1 Critical Bug Analysis ✅ **IDENTIFIED & RESOLVED**
**Problem: Variable Substitution Lookup Mismatch**
- Variable `{@pokemon}` in URL → stored as `url.pokemon = "pikachu"`
- Substitution tried to find: `substitutions["url"]` ← WRONG KEY
- Should look for: `substitutions["url.pokemon"]` ← CORRECT KEY
- Result: No substitution found → control characters `\x16\x18` in URL

**Root Cause:** Line 243 in `variables.go` - incorrect key lookup for full string variables

#### 3.6.2 Surgical Fix Implementation ✅ **COMPLETED**
- **Single Line Fix**: Modified variable key construction in `SubstituteVariables()`
- **Zero Collateral Damage**: No changes to storage format or detection logic
- **Result**: Perfect variable prompting and clean URL substitution

#### 3.6.3 Success Criteria ✅ **ALL ACHIEVED**
- [x] ✅ Variable prompting works correctly (no more silence during `call`)
- [x] ✅ Smart variable deduplication works as specified in vision.md
- [x] ✅ No control characters in URLs (`\x16\x18` eliminated)
- [x] ✅ Clean HTTP requests with proper variable substitution
- [x] ✅ All existing functionality preserved

#### 3.5.1 Root Cause Analysis ✅ **IDENTIFIED & RESOLVED**
**Problem 1 - TOML Merging Bug:**
- `MergePresetFiles()` loses file context, causing URL variables to be classified as headers
- Lines 158-172 in `http.go` assume "strings = headers" which breaks with URL variables
- Impact: Any string value in non-header files gets misclassified (URL vars, query params, string body fields)

**Problem 2 - Variable Syntax Conflicts:**
- Current `@name`/`?name` syntax conflicts with real URLs containing @ and ? characters
- Examples that break: `https://api.github.com/@username`, `https://api.com/search?q=test`
- Cannot distinguish between actual URL characters and variable placeholders

**Combined Impact:**
- Real-world APIs with @ and ? in URLs cannot be tested properly
- Users must avoid common API patterns or get incorrect behavior

#### 3.5.2 Combined Implementation Strategy

**Fix 1: Separate Handler Implementation** ✅ **IMPLEMENTED**
- [x] **Rewrite HTTP Execution Flow** in `src/project/executor/http.go`:
  - ✅ Removed `MergePresetFiles()` function entirely - no longer exists
  - ✅ Created `LoadPresetFile(preset, filename)` helper function (lines 98-112)
  - ✅ Load each TOML file as separate handler: request, headers, body, query (lines 46-49)
  - ✅ Apply variable substitution to each handler separately (lines 52-67)

- [x] **Update BuildHTTPRequest() Logic** in `BuildHTTPRequestFromHandlers()`:
  - ✅ Extract URL/method/timeout explicitly from request handler (lines 136-155)
  - ✅ Extract headers explicitly from headers handler (lines 158-163)
  - ✅ Extract body explicitly from body handler (lines 174-187)
  - ✅ Extract query parameters explicitly from query handler (lines 166-171)
  - ✅ Eliminated all guessing/heuristic logic - clean separation

**Fix 2: Variable Syntax Update** ✅ **IMPLEMENTED**
- [x] **Update Variable Detection** in `src/project/executor/variables.go`:
  - ✅ Changed `DetectVariableType()` to recognize `{@name}` and `{?name}` (lines 36-43)
  - ✅ Updated regex patterns to `^\{@(\w*)\}$` and `^\{\?(\w*)\}$` for proper detection
  - ✅ Updated `SubstituteVariables()` to handle braced format in all handlers (lines 233-261)

- [x] **Update Variable Processing**:
  - ✅ Modified variable prompting to display braced syntax correctly (lines 69-116)
  - ✅ URL parsing works correctly with braced variables (lines 184-229)
  - ✅ Complex URLs with multiple variables and URL-native @ and ? work correctly

#### 3.5.3 Implementation Architecture
```go
// NEW: Combined fix - separate handlers + braced variable syntax
func ExecuteCallCommand(cmd parser.Command) error {
    // Variable prompting now works with braced syntax {@ and {?
    substitutions, err := PromptForVariables(cmd.Preset, persist)
    if err != nil {
        return fmt.Errorf("variable prompting failed: %v", err)
    }
    
    // Load each file as separate handler - no merging
    requestHandler := LoadPresetFile(cmd.Preset, "request")
    headersHandler := LoadPresetFile(cmd.Preset, "headers")  
    bodyHandler := LoadPresetFile(cmd.Preset, "body")
    queryHandler := LoadPresetFile(cmd.Preset, "query")

    // Apply variable substitutions to each separately (now handles {@ and {?)
    SubstituteVariables(requestHandler, substitutions)
    SubstituteVariables(headersHandler, substitutions)
    SubstituteVariables(bodyHandler, substitutions)
    SubstituteVariables(queryHandler, substitutions)

    // Build HTTP request components explicitly - no guessing
    request := &HTTPRequestConfig{
        Method:  requestHandler.GetAsString("method"),
        URL:     requestHandler.GetAsString("url"), // Can now handle URLs with native @ and ?
        Timeout: parseTimeout(requestHandler.GetAsString("timeout")),
        Headers: headersHandler.ToMap(),  // Only from headers.toml - never misclassified
        Body:    bodyHandler.ToJSON(),    // Only from body.toml - preserves structure
        Query:   queryHandler.ToMap(),    // Only from query.toml - no confusion
    }
    
    // ... rest stays same ...
}

// Helper function for clean file loading
func LoadPresetFile(preset, filename string) *toml.TomlHandler {
    presetPath, _ := presets.GetPresetPath(preset)
    filePath := filepath.Join(presetPath, filename+".toml")
    handler, err := toml.NewTomlHandler(filePath)
    if err != nil {
        // Return empty handler if file doesn't exist
        return createEmptyHandler()
    }
    return handler
}

// Updated variable detection (in variables.go)
func DetectVariableType(value string) (VariableType, string) {
    // OLD: @name and ?name (conflicts with URLs)
    // NEW: {@ name} and {?name} (no conflicts)
    if matched := regexp.MustCompile(`\{@(\w+)\}`).FindStringSubmatch(value); matched != nil {
        return HardVariable, matched[1]
    }
    if matched := regexp.MustCompile(`\{\?(\w*)\}`).FindStringSubmatch(value); matched != nil {
        return SoftVariable, matched[1]
    }
    return NoVariable, ""
}

// Example usage - now works with real URLs:
// saul api set url https://api.github.com/{@username}/repos?type=public
// saul api set url https://search.api.com/@mentions?q={?term}
```

#### 3.5.4 Test Coverage Enhancement ✅ **COMPLETED**
- [x] **Add Regression Tests**: Created comprehensive tests for both bugs and fixes
  ```bash
  # Test 1: TOML merging fix - URL variables stay in request.toml
  saul testapi set url https://api.github.com/@octocat/repos  
  saul testapi set header Authorization=Bearer{@token}
  # ✅ FIXED: @octocat stays literal in URL, {token} variable in header
  
  # Test 2: Variable syntax fix - real URLs work correctly
  saul testapi set url https://api.github.com/@octocat/repos?type=public
  # ✅ FIXED: @octocat treated as literal URL part, no variable detection
  
  # Test 3: Complex real-world scenario works perfectly
  saul testapi set url https://api.twitter.com/@user/posts?search=@mentions&filter=recent
  # ✅ FIXED: All @ symbols literal, no variable detection chaos
  ```

- [x] **Validate Combined Fix**: Both fixes work together seamlessly
  ```bash
  # After fix - works correctly:
  saul testapi set url https://api.github.com/{@username}/repos?type=public  
  saul testapi set header Authorization=Bearer{@token}
  saul testapi set body search.query={?searchterm}
  # ✅ No misclassification, real URLs work, variables are braced
  ```

- [x] **Integration Testing**: All existing functionality works with new syntax
- [x] **Real-World URL Testing**: Comprehensive testing with actual API URLs completed
- [x] **Update Test Suite**: Added Phase 3.5 test section to `test_suite_fixed.sh`

**Phase 3.5 Success Criteria:** ✅ **ALL ACHIEVED**
**Fix 1 - TOML Merging:**
- [x] ✅ No more field misclassification (URL variables stay in request context)
- [x] ✅ Headers only come from `headers.toml` - never from other files
- [x] ✅ Body only comes from `body.toml` - complex structures preserved
- [x] ✅ Query only comes from `query.toml` - no string confusion
- [x] ✅ Architecture respects Unix philosophy (each file = one clear purpose)

**Fix 2 - Variable Syntax:**
- [x] ✅ All variable syntax migrated to braced format `{@name}`/`{?name}`
- [x] ✅ No URL parsing conflicts with variable syntax
- [x] ✅ Real-world URLs work correctly: `https://api.github.com/@username` (literal @)
- [x] ✅ Complex URLs work: `https://api.com/{@user}/posts?search=@mentions&token={@auth}`
- [x] ✅ Variable detection is unambiguous and predictable

**Combined Integration:**
- [x] ✅ All existing Phase 1-3 tests continue passing with new syntax
- [x] ✅ Real-world API URLs can be tested immediately
- [x] ✅ No workarounds needed for common URL patterns

**Benefits:**
- ✅ **Eliminates Two Bug Classes**: No guessing logic + no syntax conflicts
- ✅ **Predictable Behavior**: File source + braced syntax = always clear
- ✅ **Real-World Ready**: Works with actual API URLs immediately
- ✅ **KISS Compliance**: Simpler, more explicit code flow
- ✅ **Future-Proof**: Solid foundation for Phase 4+ features

---

### **Phase 4A: Edit Command System** ✅ **COMPLETED**
*Goal: Interactive field editing and quick variable syntax changes*

#### 4A.1 Field-Level Edit Implementation ✅ **COMPLETED**

**Dependency Decision:** ✅ Use `github.com/chzyer/readline v1.5.1` for pre-filled terminal editing
- Lightweight pure-Go library (~50KB compiled)
- Standard choice for Go CLI tools (23k+ projects use it)
- Provides true terminal editing experience with cursor movement, backspace, etc.

**Implementation Completed:**

- [x] **Add Dependency** (`go.mod`): ✅ **COMPLETED**
  ```go
  require github.com/chzyer/readline v1.5.1
  ```

- [x] **Add Command Recognition** (`parser/command.go`): ✅ **COMPLETED**
  ```go
  // Edit command handling added with same syntax as check command
  ```

- [x] **Add Command Routing** (`cmd/main.go`): ✅ **COMPLETED**
  ```go
  case "edit":
      return executor.ExecuteEditCommand(cmd)
  ```

- [x] **Implement ExecuteEditCommand** (`executor/commands.go`): ✅ **COMPLETED**
  ```go
  func ExecuteEditCommand(cmd parser.Command) error {
      // 1. Load current value using existing patterns
      handler, _ := presets.LoadPresetFile(cmd.Preset, normalizeTarget(cmd.Target))
      currentValue := handler.GetAsString(cmd.Key)

      // 2. Pre-filled interactive editing with readline
      rl, _ := readline.New(fmt.Sprintf("%s: ", cmd.Key))
      rl.WriteStdin([]byte(currentValue))
      newValue, err := rl.Readline()

      // 3. Save using existing validation and patterns
      handler.Set(cmd.Key, newValue)
      return presets.SavePresetFile(cmd.Preset, cmd.Target, handler)
  }
  ```

**Implementation Scope (KISS - Start Simple):**
- ✅ Field-level editing only: `saul api edit url`, `saul api edit body pokemon.name`
- ✅ String values only (handles 90% of use cases)
- ✅ Uses existing validation, normalization, and TOML patterns
- ✅ Same syntax as check command for consistency
- ❌ Variable editing (`edit @name`) - defer to Phase 4A.2
- ❌ Container-level editing (`edit body`) - defer to Phase 4A.2

**Field Existence Handling:**
- Non-existent fields → Show empty string for editing
- Use existing `normalizeTarget()` validation
- Reuse `validateRequestField()` for request fields

#### 4A.2 Container-Level Edit Implementation
- [ ] **Editor Integration**:
  - Handle container-level editing: `edit body`, `edit header`, `edit query`
  - Detect default editor from `$EDITOR` environment variable
  - Implement cross-platform editor detection and launching
  - Handle editor exit codes and provide user feedback

- [ ] **Command Routing Integration**:
  - Add edit command routing to main command parser
  - Distinguish between field-level and container-level editing based on arguments
  - Integrate with existing target normalization system
  - Add edit command help and usage examples

**Phase 4A.1 Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ `saul api edit url` shows pre-filled readline prompt with current value
- [x] ✅ User can backspace, edit characters, move cursor in terminal
- [x] ✅ `saul api edit body pokemon.name` prompts for nested field with current value
- [x] ✅ Non-existent fields show empty string for editing (create new)
- [x] ✅ Uses existing validation (URL format, method validation, etc.)
- [x] ✅ All existing Phase 1-3.5 functionality unchanged
- [x] ✅ Zero regression - purely additive feature

**Current Status:** Field-level edit command fully functional with readline integration

**Phase 4A.2 Success Criteria (Future):**
- [ ] Variable editing: `saul api edit @pokename`
- [ ] Container editing: `saul api edit body` (opens in $EDITOR)
- [ ] Field creation safety prompts
- [ ] Cross-platform editor integration

**Phase 4A Testing:**
```bash
#!/bin/bash
# Phase 4A Edit Command Tests

echo "4A.1 Testing field-level editing..."
saul testapi set url https://example.com
echo "https://newurl.com" | saul testapi edit url
saul testapi check url | grep -q "newurl.com"

echo "4A.2 Testing variable editing..."
saul testapi set body name={@pokename}
echo "pikachu" | saul testapi edit @pokename
grep -q 'pokename.*=.*pikachu' ~/.config/saul/presets/testapi/variables.toml

echo "4A.3 Testing container-level editing (if EDITOR set)..."
# This test requires manual verification with editor

echo "4A.4 Testing field creation safety..."
echo "n" | saul testapi edit body nonexistent.field | grep -q "doesn't exist"

echo "✓ Phase 4A Edit Command System: PASSED"
```

---

### **Phase 4B: Response Formatting System**
*Goal: Smart JSON→TOML response display for optimal readability*

#### 4B.1 JSON to TOML Conversion Engine
- [ ] **Add FromJSON() Method to TomlHandler**:
  - Implement `NewTomlHandlerFromJSON(jsonData []byte)` in `toml/handler.go`
  - Create JSON → Go map → TOML tree conversion pipeline
  - Handle nested objects, arrays, and primitive types correctly
  - Add error handling for invalid JSON with graceful fallback

- [ ] **Smart Response Formatting Logic**:
  - Modify `DisplayResponse()` in `executor/http.go` to detect content types
  - JSON responses → Convert to TOML for readable display
  - Non-JSON responses → Display raw content as-is
  - Add response metadata header (status, timing, size, content-type)
  - Implement graceful fallback to raw display if conversion fails

#### 4A.2 Content-Type Detection & Display
- [ ] **Enhanced Response Display**:
  - Format response header: `Status: 200 OK (324ms, 2.1KB)`
  - Add content-type detection from response headers
  - Smart TOML formatting for JSON responses with metadata
  - Preserve raw display for HTML, XML, plain text, and other formats
  - Handle edge cases: empty responses, malformed JSON, large responses

- [ ] **Comprehensive API Testing**:
  - **JSONPlaceholder** (`jsonplaceholder.typicode.com`) - Simple JSON testing
  - **PokéAPI** (`pokeapi.co`) - Complex nested structures, arrays
  - **HTTPBin** (`httpbin.org`) - Multiple content types, edge cases
  - **GitHub API** (`api.github.com`) - Real-world complexity, large responses
  - Validate formatting across all API types and response patterns

**Phase 4A Success Criteria:**
- [ ] `saul call pokeapi` displays JSON responses in readable TOML format
- [ ] Response metadata shows clearly: status, timing, size, content-type
- [ ] Non-JSON responses display raw content unchanged
- [ ] Invalid JSON gracefully falls back to raw display
- [ ] All 4 test APIs (JSONPlaceholder, Pokémon, HTTPBin, GitHub) format correctly
- [ ] Existing Phase 1-3.5 functionality unchanged

**Phase 4A Testing:**
```bash
#!/bin/bash
# Phase 4A Edit Command Tests

echo "4A.1 Testing field-level editing..."
saul testapi set url https://example.com
echo "https://newurl.com" | saul testapi edit url
saul testapi check url | grep -q "newurl.com"

echo "4A.2 Testing variable editing..."
saul testapi set body name={@pokename}
echo "pikachu" | saul testapi edit @pokename
grep -q 'pokename.*=.*pikachu' ~/.config/saul/presets/testapi/variables.toml

echo "4A.3 Testing container-level editing (if EDITOR set)..."
# This test requires manual verification with editor

echo "4A.4 Testing field creation safety..."
echo "n" | saul testapi edit body nonexistent.field | grep -q "doesn't exist"

echo "✓ Phase 4A Edit Command System: PASSED"
```

---

### **Phase 4C: Response History Storage**
*Goal: Add response storage and management for debugging*

#### 4C.1 History Storage Management
- [ ] **History Storage Integration**:
  - Modify `ExecuteCallCommand` to store responses when history enabled
  - Implement `CreateHistoryDirectory(preset string)` in presets package
  - Add history rotation logic (keep last N, delete oldest)
  - Create response file naming: `response-001.json`, `response-002.json`, etc.
  - Include request metadata (method, URL, timestamp) with raw response
  - Store raw JSON responses but display with Phase 4A formatting

- [ ] **History Configuration**:
  - Extend `request.toml` structure to include `[settings]` section
  - Add `history_count = N` setting (0 = disabled)
  - Implement `set history N` command to configure per preset
  - Update `ExecuteSetCommand` to handle history configuration

#### 4C.2 History Access Commands
- [ ] **Check History Command**:
  - Implement `ExecuteCheckHistoryCommand` for history access
  - Add interactive menu: list all stored responses with metadata
  - Support direct access: `check history N` for specific response
  - Add `check history last` alias for most recent response
  - Display stored responses using Phase 4A smart formatting

- [ ] **History Management**:
  - Implement `rm history` command with confirmation prompt
  - Add "Delete all history for 'preset'? (y/N):" confirmation
  - Support selective deletion: `rm history N` (future enhancement)
  - Handle cases where history doesn't exist (silent success)

#### 4C.3 Enhanced Command Routing
- [ ] **Extended Check Command**:
  - Add history routing to existing `ExecuteCheckCommand`
  - Handle `check history` variations (no args = menu, N = direct, last = recent)
  - Maintain existing check functionality for TOML inspection

- [ ] **Extended Set Command**:
  - Add history configuration to `ExecuteSetCommand`
  - Validate history count values (non-negative integers)
  - Handle `set history 0` to disable without deleting existing history

**Phase 4C Success Criteria:**
- [ ] `saul api set history 5` enables history collection
- [ ] `saul call api` automatically stores responses when history enabled
- [ ] `saul api check history` shows interactive menu of stored responses
- [ ] `saul api check history 1` displays most recent response with Phase 4A formatting
- [ ] `saul api rm history` deletes all history with confirmation prompt
- [ ] History rotation works correctly (keeps last N, deletes oldest)
- [ ] Stored responses use Phase 4A smart formatting when displayed

**Phase 4C Testing:**
```bash
#!/bin/bash
# Phase 4C History Storage Tests

echo "4C.1 Testing history configuration..."
saul testapi set history 3
grep -q 'history_count = 3' ~/.config/saul/presets/testapi/request.toml

echo "4C.2 Testing history storage..."
saul call testapi >/dev/null  # Should store response
[ -d ~/.config/saul/presets/testapi/history ]
[ -f ~/.config/saul/presets/testapi/history/response-001.json ]

echo "4C.3 Testing history access with formatting..."
saul testapi check history | grep -q "1." # Should show menu
saul testapi check history 1 | grep -q "Status:" # Should show formatted response

echo "4C.4 Testing history management..."
echo "y" | saul testapi rm history
[ ! -d ~/.config/saul/presets/testapi/history ]

echo "✓ Phase 4C Response History Storage: PASSED"
```

---

### **Phase 5: Interactive Mode**
*Goal: Working interactive shell for preset management*

#### 5.1 Interactive Shell Implementation
- [ ] **Shell Mode Detection**:
  - Detect when `saul preset` called without additional commands
  - Implement `EnterInteractiveMode(preset string)` function
  - Create command loop with `[preset]> ` prompt showing current preset
  - Handle shell-specific commands: `exit`, `quit`, `help`

- [ ] **Command Processing in Interactive Mode**:
  - Reuse existing command parsing but strip preset name
  - Route commands through same executors as single-line mode
  - Maintain command history within session
  - Handle multi-word commands and proper argument parsing

- [ ] **Interactive User Experience**:
  - Show welcome message: "Entered interactive mode for 'preset'"
  - Display help reminder: "Type 'help' for commands or 'exit' to leave"
  - Handle Ctrl+C gracefully (exit interactive mode, return to shell)
  - Clear error handling without exiting interactive session

#### 5.2 Interactive Command Integration
- [ ] **All Existing Commands Work**:
  - `set url/method/timeout` commands work identically
  - `set body/headers/query` commands work identically  
  - `call` command works with variable prompting
  - `check` commands work including history access
  - `rm` commands work with confirmations

- [ ] **Interactive-Specific Enhancements**:
  - Command abbreviation support (optional): `c` for `call`, `s` for `set`
  - Tab completion for commands and targets (optional)
  - Show current configuration summary on demand
  - Context-aware help based on current preset state

#### 5.3 Advanced Interactive Features
- [ ] **Session Management**:
  - Track commands executed in session for debugging
  - Provide session summary on exit
  - Handle long-running sessions gracefully
  - Memory management for extended usage

**Phase 5 Success Criteria:**
- [ ] `saul myapi` enters interactive mode successfully
- [ ] All commands work identically to single-line mode
- [ ] `exit` and Ctrl+C handling works properly
- [ ] Interactive session maintains state correctly
- [ ] Help system works in interactive context
- [ ] User experience feels natural and responsive

**Phase 5 Testing:**
```bash
# Interactive mode testing (manual)
echo "Testing interactive mode..."
echo -e "set url https://httpbin.org/get\nset method GET\ncall\nexit" | saul testapi
echo "✓ Interactive mode basic functionality works"
```

---

### **Phase 6: Advanced Features & Polish**
*Goal: Complete feature set with editing and production readiness*

#### 6.1 File Editing Integration  
- [ ] **Editor Command Implementation**:
  - Implement `edit header/body/query/request/variables` commands
  - Detect default editor from `$EDITOR` environment variable
  - Fallback editor detection (nano, vim, emacs, notepad on Windows)
  - Handle editor exit codes and provide feedback

- [ ] **Cross-platform Compatibility**:
  - Windows editor integration (`notepad.exe`, VS Code, etc.)
  - macOS editor integration (TextEdit, VS Code, etc.)
  - Linux/Unix editor integration (nano, vim, emacs, etc.)
  - Handle file locking and concurrent editing scenarios

#### 6.2 Advanced Variable Features
- [ ] **Enhanced Variable Management**:
  - Support custom variable names: `pokemon.name={?pokename}`
  - Variable validation and type hints during prompting
  - Variable reuse across multiple requests in same session
  - Variable templating: common variable sets for API families

- [ ] **Variable Import/Export**:
  - Export variable sets: `saul myapi export variables > vars.json`
  - Import variable sets: `saul myapi import variables < vars.json`
  - Share variable configurations between presets
  - Variable set versioning and backup

#### 6.3 Production Readiness
- [ ] **Comprehensive Error Handling**:
  - Network timeout handling with retry logic
  - DNS resolution error handling  
  - SSL/TLS certificate error handling
  - HTTP error status code explanations
  - File permission and disk space error handling

- [ ] **Performance Optimization**:
  - TOML file caching for large configurations
  - Lazy loading of presets and history
  - Memory usage optimization for long-running sessions
  - Response streaming for large API responses

- [ ] **Cross-platform Features**:
  - Windows path handling and directory creation
  - macOS keychain integration for credentials (future)
  - Linux desktop integration (future)
  - Consistent behavior across all platforms

- [ ] **Build and Distribution**:
  - GitHub Actions build pipeline for multiple platforms
  - Binary distribution for Windows, macOS, Linux
  - Package manager integration (Homebrew, apt, etc.)
  - Version management and update checking

**Phase 6 Success Criteria:**
- [ ] `saul myapi edit body` opens body.toml in default editor
- [ ] All edge cases handled gracefully with helpful error messages
- [ ] Performance is acceptable for typical usage (< 100ms command response)
- [ ] Cross-platform compatibility verified on Windows, macOS, Linux
- [ ] Ready for end-user distribution with installation documentation

## Comprehensive Testing Strategy

### **Expandable Test Suite: `other/testing/test_suite.sh`**

The existing test suite will be expanded to include Phase 4+ functionality:

```bash
#!/bin/bash
# test_suite.sh - Comprehensive test suite for all phases

# Existing Phase 1-3 tests continue to work...

# ✅ IMPLEMENTED: Phase 3.5 tests (Architecture & Syntax Fix)
echo "===== PHASE 3.5 TESTS: Architecture & Variable Syntax Fix ====="

echo "3.5.1 Testing separate handlers (no field misclassification)..."
saul testapi set url https://api.github.com/{@username}/repos?type=public
saul testapi set header Authorization=Bearer{@token}
saul testapi set body search.query={?term}
# ✅ VERIFIED: URL variables stay in request.toml, not misclassified

echo "3.5.2 Testing braced variable syntax..."
echo -e "octocat\ntoken123\nrepos" | saul call testapi >/dev/null
# ✅ VERIFIED: Works with real URLs containing literal @ and ?

echo "3.5.3 Testing real-world URL patterns..."
saul testapi set url https://api.twitter.com/@mentions?search={?query}
# ✅ VERIFIED: Only {?query} prompts, @mentions stays literal

echo "✓ Phase 3.5: Architecture & Variable Syntax Fix - PASSED"

# NEW: Phase 4A tests (Response Formatting)
echo "===== PHASE 4A TESTS: Response Formatting System ====="

echo "4A.1 Testing JSON→TOML conversion..."
saul pokeapi set url https://pokeapi.co/api/v2/pokemon/1
saul call pokeapi | grep -q "name = " # Should show TOML format

echo "4A.2 Testing complex nested JSON..."
saul ghapi set url https://api.github.com/repos/octocat/Hello-World
saul call ghapi | grep -q "\[" # Should show TOML sections

echo "4A.3 Testing non-JSON responses..."
saul httpbin set url https://httpbin.org/html
saul call httpbin | grep -q "<html>" # Should show raw HTML

echo "✓ Phase 4A: Response Formatting System - PASSED"

# NEW: Phase 4B tests (History Storage)
echo "===== PHASE 4B TESTS: Response History Storage ====="

echo "4B.1 Testing history configuration..."
saul testapi set history 3
grep -q 'history_count = 3' ~/.config/saul/presets/testapi/request.toml

echo "4C.2 Testing history storage..."
echo -e "testuser\n123" | saul call testapi >/dev/null
[ -f ~/.config/saul/presets/testapi/history/response-001.json ]

echo "4C.3 Testing history access with formatting..."
saul testapi check history | grep -q "1\."
saul testapi check history 1 | grep -q "Status:"

echo "✓ Phase 4B: Response History Storage - PASSED"

# Future phases will add similar test sections...
```

### **Testing Philosophy**
- **Foundation First**: Phase 3.5 fixes core architecture before adding features
- **Real-World Validation**: Test with actual API URLs containing @ and ? characters
- **Backward Compatibility**: New phases must not break existing functionality
- **Migration Testing**: Verify smooth transition from old to new variable syntax in Phase 3.5
- **Integration Testing**: Each system integrates seamlessly with existing commands  
- **Edge Case Coverage**: URL edge cases, large responses, network failures
- **Cross-platform Testing**: Verify functionality on multiple operating systems

## Development Guidelines

### **KISS Principles**
- **Simple**: Each function has one clear responsibility
- **Clean**: Self-documenting code with minimal comments
- **Intelligent**: Smart type detection and error handling
- **Resilient**: Graceful handling of edge cases and network issues

### **Breaking Change Management**
- **Phase 3.5 Migration**: Variable syntax change is breaking but necessary for real-world usage
- **Combined Fix Strategy**: Fix both architecture and syntax together for comprehensive solution
- **User Communication**: Clear migration guide and examples in documentation
- **Backward Compatibility**: Consider supporting both syntaxes briefly during transition
- **Testing**: Comprehensive testing to ensure no regression in core functionality

### **Go Best Practices**
- Follow standard Go project layout
- Use Go modules properly  
- Error handling at every boundary
- Clear package separation of concerns
- Minimal external dependencies

## Risk Mitigation

### **Phase 3.5 Specific Risks**
- **Breaking Change Impact**: Variable syntax change affects all existing users
- **URL Parsing Complexity**: Braced variables in URLs require careful parsing
- **Dual Architecture Change**: Fixing both merging and syntax simultaneously increases complexity
- **Real-World URL Edge Cases**: Many API patterns use @ and ? that must be handled correctly

### **Phase 4A Specific Risks**
- **User Input Validation**: Handling malformed input in pre-filled prompts
- **Editor Integration Complexity**: Cross-platform editor detection and launching
- **Variable Reference Validation**: Ensuring variable editing safety and error handling

### **Phase 4B Specific Risks**
- **JSON Parsing Edge Cases**: Malformed JSON, extremely large responses, deeply nested structures
- **TOML Conversion Complexity**: JSON arrays and complex objects may not translate cleanly to TOML
- **Performance Impact**: JSON→TOML conversion could slow response display for large payloads

### **Phase 4C Specific Risks**
- **History Storage Size**: Large API responses could consume significant disk space
- **File System Edge Cases**: History directory creation and rotation edge cases
- **Storage Performance**: History access could become slow with many stored responses

### **Mitigation Strategies**
- **Migration Testing**: Comprehensive test coverage for syntax change
- **Documentation**: Clear examples of new variable syntax in all documentation
- **Storage Limits**: Implement response size limits and compression options
- **Graceful Degradation**: History system fails gracefully if disk space insufficient

## Success Metrics

### **Phase 3.5 Completion Criteria**
- All existing functionality works with new variable syntax
- No field misclassification (separate handlers work correctly)
- No URL parsing conflicts with variable syntax  
- Real-world API URLs work without workarounds
- Migration from old to new syntax is seamless
- Test suite passes completely including new Phase 3.5 tests

### **Phase 4A Completion Criteria**
- Field-level editing works with pre-filled prompts showing current values
- Variable editing works with stored hard variable values
- Container-level editing opens files in default editor correctly
- Field creation safety prompts work for non-existent fields
- Variable editing safety prevents editing non-existent variables
- Cross-platform editor integration works on major platforms
- All existing Phase 1-3.5 functionality unchanged

### **Phase 4B Completion Criteria**
- JSON responses display in readable TOML format with metadata header
- Content-type detection works correctly (JSON vs non-JSON)
- Graceful fallback to raw display for invalid JSON or non-JSON content
- All 4 test APIs (JSONPlaceholder, Pokémon, HTTPBin, GitHub) format correctly
- No performance degradation for typical API response sizes
- All existing Phase 1-3.5 and Phase 4A functionality unchanged

### **Phase 4C Completion Criteria**
- History system stores and retrieves responses correctly
- History configuration and rotation work properly
- History access commands provide useful debugging workflow using Phase 4B formatting
- All existing Phase 1-3.5, Phase 4A, and Phase 4B functionality unchanged

### **Final Project Success**
- All commands from vision.md work correctly
- Variable syntax handles all URL edge cases without conflicts (Phase 3.5)
- No field misclassification bugs in HTTP execution (Phase 3.5)
- Edit command system provides quick field and variable editing workflow (Phase 4A)
- Smart response formatting provides readable output for API development (Phase 4B)
- History system provides valuable debugging workflow (Phase 4C)
- Interactive mode enables efficient preset management (Phase 5)
- Ready for production distribution with advanced features (Phase 6)
- Maintains KISS principles while adding powerful features throughout

---

*This action plan prioritizes edit command implementation (Phase 4A) as the immediate next step after fixing critical architecture issues (Phase 3.5). This approach provides immediate workflow improvements with zero dependencies, followed by response formatting (Phase 4B) for readable API output, and finally history storage (Phase 4C). The strategic sequence allows for incremental implementation with minimal risk while maximizing user value at each step.*