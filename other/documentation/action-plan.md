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
  - Variable prompting system (`@` hard variables, `?` soft variables) *(needs syntax update)*
  - TOML file merging (request + headers + body + query + variables)
  - HTTP client integration using go-resty
  - Support for all major HTTP methods
  - JSON body conversion and pretty-printed responses
  - Smart Variable Deduplication feature

### ❌ **Missing Core Components**
- **Variable syntax update**: Change from bare `@`/`?` to braced `{@}`/`{?}` format
- **Response history system**: Storage, management, and access commands
- **Interactive mode**: Command shell for preset management
- **Advanced command system**: Enhanced help, editing, and management
- **Production readiness**: Cross-platform compatibility, error handling polish

### 🔧 **Technical Debt**
- Variable syntax conflicts with URLs (@ and ? are valid URL characters)
- No response history for debugging API interactions
- Limited command system compared to vision
- No interactive mode for workflow efficiency

## Implementation Phases

### **Phase 1: Foundation & TOML Integration** ✅ **COMPLETED**
*All functionality implemented and tested.*

### **Phase 2: Core TOML Operations & Variable System** ✅ **COMPLETED**  
*All functionality implemented and tested.*

### **Phase 3: HTTP Execution Engine** ✅ **COMPLETED**
*All functionality implemented and tested.*

---

### **Phase 3.5: HTTP Architecture & Variable Syntax Fix** 🔧 **CRITICAL BUG FIX**
*Goal: Fix TOML merging logic AND variable syntax conflicts to enable real-world URL usage*

#### 3.5.1 Root Cause Analysis ✅ **IDENTIFIED**
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

**Fix 1: Separate Handler Implementation**
- [ ] **Rewrite HTTP Execution Flow** in `src/project/executor/http.go`:
  - Remove `MergePresetFiles()` function entirely (lines 85-123)
  - Create `LoadPresetFile(preset, filename)` helper function
  - Load each TOML file as separate handler: request, headers, body, query
  - Apply variable substitution to each handler separately

- [ ] **Update BuildHTTPRequest() Logic** (lines 125-192):
  - Extract URL/method/timeout explicitly from request handler
  - Extract headers explicitly from headers handler  
  - Extract body explicitly from body handler
  - Extract query parameters explicitly from query handler
  - Eliminate all guessing/heuristic logic (lines 158-174)

**Fix 2: Variable Syntax Update**
- [ ] **Update Variable Detection** in `src/project/executor/variables.go`:
  - Change `DetectVariableType()` to recognize `{@name}` and `{?name}` instead of `@name`/`?name`
  - Update regex patterns to `{@\w+}` and `{?\w*}` for proper detection
  - Update `SubstituteVariables()` to handle braced format in all handlers

- [ ] **Update Variable Processing**:
  - Modify variable prompting to display braced syntax correctly
  - Ensure URL parsing works correctly with braced variables
  - Test complex URLs with multiple variables and URL-native @ and ?

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

#### 3.5.4 Test Coverage Enhancement
- [ ] **Add Regression Tests**: Create tests that demonstrate both current bugs before fix
  ```bash
  # Test 1: TOML merging bug (URL variables misclassified as headers)
  saul testapi set url https://api.com/@username/posts  
  saul testapi set header Authorization=Bearer123
  # BUG: @username from URL appears in headers
  
  # Test 2: Variable syntax conflicts with real URLs  
  saul testapi set url https://api.github.com/@octocat/repos?type=public
  # BUG: @octocat treated as variable instead of literal URL part
  
  # Test 3: Complex real-world scenario
  saul testapi set url https://api.twitter.com/@user/posts?search=@mentions&filter=recent
  # BUG: Multiple @ symbols cause variable detection chaos
  ```

- [ ] **Validate Combined Fix**: Test both fixes work together
  ```bash
  # After fix - should work correctly:
  saul testapi set url https://api.github.com/{@username}/repos?type=public  
  saul testapi set header Authorization=Bearer{@token}
  saul testapi set body search.query={?searchterm}
  # No misclassification, real URLs work, variables are braced
  ```

- [ ] **Integration Testing**: All existing functionality unchanged with new syntax
- [ ] **Real-World URL Testing**: Comprehensive testing with actual API URLs
- [ ] **Update Test Suite**: Add Phase 3.5 test section to `test_suite.sh`

**Phase 3.5 Success Criteria:**
**Fix 1 - TOML Merging:**
- [ ] No more field misclassification (URL variables stay in request context)
- [ ] Headers only come from `headers.toml` - never from other files
- [ ] Body only comes from `body.toml` - complex structures preserved
- [ ] Query only comes from `query.toml` - no string confusion
- [ ] Architecture respects Unix philosophy (each file = one clear purpose)

**Fix 2 - Variable Syntax:**
- [ ] All variable syntax migrated to braced format `{@name}`/`{?name}`
- [ ] No URL parsing conflicts with variable syntax
- [ ] Real-world URLs work correctly: `https://api.github.com/@username` (literal @)
- [ ] Complex URLs work: `https://api.com/{@user}/posts?search=@mentions&token={@auth}`
- [ ] Variable detection is unambiguous and predictable

**Combined Integration:**
- [ ] All existing Phase 1-3 tests continue passing with new syntax
- [ ] Real-world API URLs can be tested immediately
- [ ] No workarounds needed for common URL patterns

**Benefits:**
- ✅ **Eliminates Two Bug Classes**: No guessing logic + no syntax conflicts
- ✅ **Predictable Behavior**: File source + braced syntax = always clear
- ✅ **Real-World Ready**: Works with actual API URLs immediately
- ✅ **KISS Compliance**: Simpler, more explicit code flow
- ✅ **Future-Proof**: Solid foundation for Phase 4+ features

---

### **Phase 4: Response History System**
*Goal: Add response history for debugging API interactions*

#### 4.1 History Storage Management
- [ ] **History Storage Management**:
  - Implement `CreateHistoryDirectory(preset string)` in presets package
  - Add history rotation logic (keep last N, delete oldest)
  - Create response file naming: `response-001.json`, `response-002.json`, etc.
  - Add history metadata (timestamp, status, method, URL, size)

- [ ] **History Configuration**:
  - Extend `request.toml` structure to include `[settings]` section
  - Add `history_count = N` setting (0 = disabled)
  - Implement `set history N` command to configure per preset
  - Update `ExecuteSetCommand` to handle history configuration

- [ ] **History Storage Integration**:
  - Modify `ExecuteCallCommand` to store responses when history enabled
  - Add response storage after successful HTTP execution
  - Include request metadata (method, URL, timestamp) with response
  - Handle response size limits and truncation for large responses

#### 4.2 History Access Commands
- [ ] **Check History Command**:
  - Implement `ExecuteCheckHistoryCommand` for history access
  - Add interactive menu: list all stored responses with metadata
  - Support direct access: `check history N` for specific response
  - Add `check history last` alias for most recent response
  - Display response with same formatting as live responses

- [ ] **History Management**:
  - Implement `rm history` command with confirmation prompt
  - Add "Delete all history for 'preset'? (y/N):" confirmation
  - Support selective deletion: `rm history N` (future enhancement)
  - Handle cases where history doesn't exist (silent success)

#### 4.3 Enhanced Command Routing
- [ ] **Extended Check Command**:
  - Add history routing to existing `ExecuteCheckCommand`
  - Handle `check history` variations (no args = menu, N = direct, last = recent)
  - Maintain existing check functionality for TOML inspection
  
- [ ] **Extended Set Command**:  
  - Add history configuration to `ExecuteSetCommand`
  - Validate history count values (non-negative integers)
  - Handle `set history 0` to disable without deleting existing history

**Phase 4 Success Criteria:**
- [ ] `saul api set history 5` enables history collection
- [ ] `saul call api` automatically stores responses when history enabled
- [ ] `saul api check history` shows interactive menu of stored responses
- [ ] `saul api check history 1` displays most recent response
- [ ] `saul api rm history` deletes all history with confirmation prompt
- [ ] History rotation works correctly (keeps last N, deletes oldest)
- [ ] All existing Phase 1-3.5 functionality unchanged

**Phase 4 Testing:**
```bash
#!/bin/bash
# Phase 4 test additions

echo "4.1 Testing history configuration..."
saul testapi set history 3
grep -q 'history_count = 3' ~/.config/saul/presets/testapi/request.toml

echo "4.2 Testing history storage..."
saul call testapi >/dev/null  # Should store response
[ -d ~/.config/saul/presets/testapi/history ]
[ -f ~/.config/saul/presets/testapi/history/response-001.json ]

echo "4.3 Testing history access..."
saul testapi check history | grep -q "1." # Should show menu
saul testapi check history 1 | grep -q "Status:" # Should show response

echo "4.4 Testing history management..."
echo "y" | saul testapi rm history
[ ! -d ~/.config/saul/presets/testapi/history ]

echo "✓ Phase 4 Response History System: PASSED"
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

# NEW: Phase 3.5 tests (Architecture & Syntax Fix)
echo "===== PHASE 3.5 TESTS: Architecture & Variable Syntax Fix ====="

echo "3.5.1 Testing separate handlers (no field misclassification)..."
saul testapi set url https://api.github.com/{@username}/repos?type=public
saul testapi set header Authorization=Bearer{@token}
saul testapi set body search.query={?term}
# Verify URL variables don't appear in headers

echo "3.5.2 Testing braced variable syntax..."
echo -e "octocat\ntoken123\nrepos" | saul call testapi >/dev/null
# Should work with real URLs containing literal @ and ?

echo "3.5.3 Testing real-world URL patterns..."
saul testapi set url https://api.twitter.com/@mentions?search={?query}
# Only {?query} should prompt, @mentions should be literal

echo "✓ Phase 3.5: Architecture & Variable Syntax Fix - PASSED"

# NEW: Phase 4 tests (History System)
echo "===== PHASE 4 TESTS: Response History System ====="

echo "4.1 Testing history configuration..."
saul testapi set history 3
grep -q 'history_count = 3' ~/.config/saul/presets/testapi/request.toml

echo "4.2 Testing history storage..."
echo -e "testuser\n123" | saul call testapi >/dev/null
[ -f ~/.config/saul/presets/testapi/history/response-001.json ]

echo "4.3 Testing history access..."
saul testapi check history | grep -q "1\."
saul testapi check history 1 | grep -q "Status:"

echo "✓ Phase 4: Response History System - PASSED"

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

### **Phase 4 Specific Risks**
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

### **Phase 4 Completion Criteria**
- History system stores and retrieves responses correctly
- History configuration and rotation work properly
- History access commands provide useful debugging workflow
- All existing Phase 1-3.5 functionality unchanged

### **Final Project Success**
- All commands from vision.md work correctly
- Variable syntax handles all URL edge cases without conflicts (Phase 3.5)
- No field misclassification bugs in HTTP execution (Phase 3.5)
- History system provides valuable debugging workflow (Phase 4)
- Interactive mode enables efficient preset management (Phase 5)
- Ready for production distribution with advanced features (Phase 6)
- Maintains KISS principles while adding powerful features throughout

---

*This action plan prioritizes fixing critical architecture issues (Phase 3.5) before adding new features, ensuring the project has a solid foundation for real-world API development workflows. The combined fix addresses both TOML merging bugs and variable syntax conflicts to enable immediate testing with actual API URLs.*