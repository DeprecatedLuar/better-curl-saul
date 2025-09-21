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
- **Phase 3.7 Complete**: Variable Detection System Simplification
  - Replaced complex TOML structure parsing with simple regex-based detection
  - Fixed nested TOML variable detection: `[pokemon] name = "{@pokename}"` now works
  - Reduced ~100 lines of complex code to ~20 lines of regex
  - Zero breaking changes, same user experience, much more reliable
- **Phase 4A Complete**: Edit Command System
  - Field-level editing with pre-filled readline prompts
  - Interactive terminal editing experience with cursor movement
  - Uses existing validation and TOML patterns
  - Zero regression - purely additive feature
- **Phase 4B Complete**: Response Formatting System
  - Smart JSON→TOML conversion for optimal readability
  - Intelligent content-type detection with graceful fallback
  - HTTP subfolder refactoring for clean architecture
  - Real-world tested with multiple API types
- **Phase 4B-Post Complete**: Comma-Separated Syntax Enhancement
  - Unix-like parsing approach: right tool for each job
  - Unified KeyValuePairs array system for clean architecture
  - Multiple key=value pairs: `Auth=token,Accept=json` (50%+ fewer commands)
  - Quoted values with commas: `Type="application/json,charset=utf-8"`
  - Explicit array syntax: `Tags=[red,blue,green]` with bracket notation
  - Zero regression, perfect backward compatibility, no shell escaping needed
- **Bulk Operations System Complete**: Space-Separated Universal Bulk Pattern
  - Universal bulk detection: `saul rm preset1 preset2 preset3` (space-separated)
  - Continue + warn approach: delete existing presets, warn about non-existent
  - Parser enhancement: `Targets []string` field for multiple space-separated arguments
  - Command execution: iterate over all targets with graceful error handling
  - Consistent Unix pattern: same space-separated approach for all bulk operations
- ✅ **Phase 4B-Post-2 Complete**: Space-Separated Key-Value Migration for Universal Consistency
  - Universal space-separated pattern: `saul api set body name=val1 type=val2` (space-separated)
  - Code simplification: Eliminated ~100 lines of complex comma/quote parsing logic
  - Parser enhancement: `args[3:]` with simple iteration replaces complex regex patterns
  - Perfect Unix consistency: Same space-separated approach for all bulk operations (rm, set, etc.)
  - Zero regression: All existing functionality preserved with cleaner, more intuitive syntax
- ✅ **Phase 4C Complete**: Response Filtering System
  - Terminal overflow solved: 257KB APIs → filtered fields display
- ✅ **Phase 4D Complete**: Terminal Session Memory System *(emergent feature - implemented without prior planning)*
  - Terminal-scoped preset memory enables shorthand commands: `saul set body name=val` (no preset needed)
  - TTY-based session isolation with automatic preset injection for improved workflow efficiency
  - Pure UNIX design: Zero special parsing, uses existing KeyValuePairs system
  - Clean syntax: `saul api set filters field1=name field2=stats.0.base_stat field3=types.0.type.name`
  - TOML array storage: `fields = ["name", "stats.0.base_stat", "types.0.type.name"]`
  - Real-world tested: PokéAPI, JSONPlaceholder complex filtering works perfectly
  - Silent error handling: Missing fields ignored gracefully
- ✅ **Phase 5A Complete**: Universal Flag System
  - Flag parsing foundation with extensible architecture for future flags
  - `--raw` flag implemented across all commands: check, call, list
  - Check commands: raw TOML file contents (cat behavior) for scripting
  - Call commands: raw response body only (no headers/metadata) for automation
  - List commands: space-separated preset names for shell scripting
  - Perfect Unix philosophy: crude, scriptable output when `--raw` specified
  - Zero regression: all existing formatted output remains default behavior
- ✅ **Phase 6A Complete**: System Command Delegation
  - Unix philosophy implementation: leverage existing tools instead of rebuilding
  - Replaced custom `saul list` with system command delegation (`saul ls`)
  - Whitelist-based security: only safe commands (ls, exa, lsd, tree, dir) allowed
  - Working directory automatically set to presets folder for all delegated commands
  - Cross-platform support with user's preferred tools (exa, lsd, etc.)
  - Perfect workspace visibility: see actual TOML files and directory structure

### ⏳ **Next Priority Phases**

### ❌ **Missing Core Components**
- **Response history system**: Storage, management, and access commands
- **Interactive mode**: Command shell for preset management
- **Advanced command system**: Enhanced help and management
- **Production readiness**: Cross-platform compatibility, error handling polish

### 🔧 **Technical Debt**
- No response history for debugging API interactions
- No interactive mode for workflow efficiency

### ✅ **Major Systems Complete**
- **System Command Delegation**: Unix philosophy - leverage existing tools (ls, exa, tree)
- **Flag System**: `--raw` flag with extensible architecture for future flags
- **Response Filtering**: Terminal-friendly filtering for large API responses
- **Visual Formatting**: Professional display system with consistent formatting
- **Variable System**: Braced syntax with hard/soft variable support
- **HTTP Execution**: Full HTTP method support with smart response formatting
- **TOML Operations**: Complete TOML manipulation with Unix philosophy

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

---

### **Phase 3.7: Variable Detection System Simplification** ✅ **COMPLETED**
*Goal: Replace complex TOML structure parsing with simple regex-based variable detection*

#### 3.7.1 Problem Analysis ✅ **IDENTIFIED & RESOLVED**
**Problem: Complex Variable Detection Fails on Nested TOML**
- Current system uses recursive TOML object parsing to find variables
- Fails on nested structures like `[pokemon] name = "{@pokename}"`
- Over-engineered approach: parsing → navigation → extraction
- Fragile and hard to debug when TOML structure changes

**Root Cause:** `scanHandlerForVariables()` only handles flat structures, skips nested objects entirely

#### 3.7.2 KISS Simplification Implementation ✅ **COMPLETED**
**Replaced Complex System with Simple Regex Approach:**
- ✅ **Replace `findAllVariables()`**: Now reads files as plain text, uses regex to find `{@}` and `{?}` patterns
- ✅ **Remove Complex Functions**: Deleted `scanHandlerForVariables()`, `scanNestedMap()`, `extractPartialVariables()`
- ✅ **Simplify Substitution**: `SubstituteVariables()` now uses simple regex replacement
- ✅ **Zero Breaking Changes**: Same API signatures, same behavior, same user experience

**New Architecture (Much Simpler):**
```go
// OLD: Complex TOML parsing
func findAllVariables(preset string) ([]VariableInfo, error) {
    // Load TOML handlers, parse structure, navigate objects...
    targetVars := scanHandlerForVariables(handler, "") // ~100 lines of complexity
}

// NEW: Simple file scanning
func findAllVariables(preset string) ([]VariableInfo, error) {
    content, _ := os.ReadFile(filePath)
    regex := regexp.MustCompile(`\{([@?])(\w*)\}`)
    matches := regex.FindAllStringSubmatch(string(content), -1) // ~20 lines total
}
```

#### 3.7.3 Benefits Achieved ✅ **ALL REALIZED**
- ✅ **Works Everywhere**: Detects variables regardless of TOML nesting depth
- ✅ **Much Simpler**: Reduced ~100 lines of complex code to ~20 lines of regex
- ✅ **More Reliable**: Regex is battle-tested, doesn't break on TOML structure changes
- ✅ **Faster Performance**: Text search vs recursive object traversal
- ✅ **Easier Debug**: Simple regex vs complex recursive logic
- ✅ **Zero Breaking Changes**: Perfect interface compatibility

#### 3.7.4 Success Criteria ✅ **ALL ACHIEVED**
- [x] ✅ Nested TOML variables now work: `[pokemon] name = "{@pokename}"` prompts correctly
- [x] ✅ All existing functionality preserved (URL variables, body variables, etc.)
- [x] ✅ Same user experience and command syntax
- [x] ✅ Simplified codebase with much less complexity
- [x] ✅ Better maintainability and debuggability

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

### **Phase 4B: Response Formatting System** ✅ **COMPLETED**
*Goal: Smart JSON→TOML response display for optimal readability*

#### 4B.1 JSON to TOML Conversion Engine ✅ **COMPLETED**
- [x] **Add FromJSON() Method to TomlHandler**: ✅ **IMPLEMENTED**
  - ✅ Implemented `NewTomlHandlerFromJSON(jsonData []byte)` in `toml/handler.go`
  - ✅ Created JSON → Go map → TOML tree conversion pipeline
  - ✅ Handles nested objects, arrays, and primitive types correctly
  - ✅ Added error handling for invalid JSON with graceful fallback

- [x] **Smart Response Formatting Logic**: ✅ **IMPLEMENTED**
  - ✅ Modified `DisplayResponse()` in `executor/http.go` to detect content types
  - ✅ JSON responses → Convert to TOML for readable display
  - ✅ Non-JSON responses → Display raw content as-is
  - ✅ Added response metadata header (status, timing, size, content-type)
  - ✅ Implemented graceful fallback to raw display if conversion fails

#### 4B.2 Content-Type Detection & Display ✅ **COMPLETED**
- [x] **Enhanced Response Display**: ✅ **IMPLEMENTED**
  - ✅ Format response header: `Status: 200 OK (324ms, 2.1KB)`
  - ✅ Added content-type detection from response headers
  - ✅ Smart TOML formatting for JSON responses with metadata
  - ✅ Preserve raw display for HTML, XML, plain text, and other formats
  - ✅ Handle edge cases: empty responses, malformed JSON, large responses

- [x] **Comprehensive API Testing**: ✅ **VALIDATED**
  - ✅ **JSONPlaceholder** (`jsonplaceholder.typicode.com`) - Simple JSON testing
  - ✅ **PokéAPI** (`pokeapi.co`) - Complex nested structures, arrays
  - ✅ **HTTPBin** (`httpbin.org`) - Multiple content types, edge cases
  - ✅ **GitHub API** (`api.github.com`) - Real-world complexity, large responses
  - ✅ Validated formatting across all API types and response patterns

#### 4B.3 HTTP Subfolder Refactoring ✅ **COMPLETED**
- [x] **Clean Architecture Organization**: ✅ **IMPLEMENTED**
  - ✅ Moved HTTP execution files to `src/project/executor/http/` subfolder
  - ✅ Organized: `client.go`, `display.go`, `request.go` for clean separation
  - ✅ Updated all import paths throughout codebase
  - ✅ Maintained backward compatibility and functionality

**Phase 4B Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ `saul call pokeapi` displays JSON responses in readable TOML format
- [x] ✅ Response metadata shows clearly: status, timing, size, content-type
- [x] ✅ Non-JSON responses display raw content unchanged
- [x] ✅ Invalid JSON gracefully falls back to raw display
- [x] ✅ All 4 test APIs (JSONPlaceholder, Pokémon, HTTPBin, GitHub) format correctly
- [x] ✅ Existing Phase 1-3.7 functionality unchanged
- [x] ✅ Smart content-type detection works flawlessly
- [x] ✅ Clean HTTP subfolder organization completed

**Benefits Achieved:**
- ✅ **Dramatically Improved Readability**: JSON APIs now display in clean TOML format
- ✅ **Smart Defaults**: Automatic JSON→TOML conversion with intelligent fallback
- ✅ **Real-World Tested**: Works perfectly with JSONPlaceholder, PokéAPI, HTTPBin, GitHub
- ✅ **Clean Architecture**: HTTP code organized in logical subfolder structure
- ✅ **Zero Regressions**: All existing functionality preserved perfectly

---

### **Phase 4B-Visual: Visual Formatting Enhancement** ✅ **COMPLETED**
*Goal: Professional visual organization for terminal-friendly response display*

#### 4B-Visual.1 ASCII Art Sandwich Formatting ✅ **COMPLETED**
- [x] ✅ **Visual Headers**: Implemented `┌─ Response ─┐` style headers for section identification
- [x] ✅ **Visual Footers**: Added matching `──────────────────────` separator lines for clean closure
- [x] ✅ **Sandwich Format**: Perfect visual containment with matching top and bottom separators
- [x] ✅ **Consistent Styling**: Same visual approach for both API responses and check commands

#### 4B-Visual.2 Minimal Headers Approach ✅ **COMPLETED**
- [x] ✅ **Essential Headers Only**: Display only status line + content-type (eliminates header noise)
- [x] ✅ **Removed Header Dump**: No more overwhelming 15+ line header displays from CDN/cache systems
- [x] ✅ **Clean Focus**: Emphasizes actual response content over infrastructure metadata
- [x] ✅ **Planned Raw Mode**: Documented support for `--raw` flag to show full headers when needed

#### 4B-Visual.3 Universal Visual Consistency ✅ **COMPLETED**
- [x] ✅ **Check Commands**: All check commands use same sandwich formatting with appropriate headers
- [x] ✅ **API Responses**: HTTP responses use consistent visual structure
- [x] ✅ **Dynamic Headers**: Section headers adapt to content type ("Response", "Body", "Headers", etc.)
- [x] ✅ **Professional Appearance**: Clean, organized terminal output that scales from simple to complex content

**Phase 4B-Visual Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ Visual sandwich formatting provides clear content separation
- [x] ✅ Minimal headers eliminate noise while preserving essential information
- [x] ✅ Consistent visual approach across all command types
- [x] ✅ Professional terminal appearance suitable for development workflows
- [x] ✅ Foundation ready for future raw flag implementation

**Benefits Achieved:**
- ✅ **Professional Visual Design**: Clean ASCII art formatting creates organized, scannable output
- ✅ **Noise Reduction**: Minimal headers approach eliminates CDN/cache header clutter
- ✅ **Consistent UX**: Same visual patterns across all commands reduce cognitive load
- ✅ **Terminal Optimized**: Formatting scales well from simple checks to complex API responses
- ✅ **Future Ready**: Architecture supports planned raw mode for verbose output when needed

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

### **Phase 4B-Post: Comma-Separated Syntax Enhancement** ✅ **COMPLETED**
*Goal: Enable batch operations for dramatically improved testing and configuration efficiency*

#### 4B-Post.1 Parser Enhancement for Comma Detection ✅ **COMPLETED**
- [x] ✅ **Command Detection Logic**: 
  - ✅ Modified `ParseCommand()` with unified KeyValuePairs array approach
  - ✅ Implemented Unix-like parsing: right tool for each job (simple split vs regex)
  - ✅ Special fields remain single-value only (no comma support)
  - ✅ Regular fields support comma-separated key=value pairs

- [x] ✅ **Value Splitting Logic**:
  - ✅ Implemented simple Unix approach: `parseSinglePair()` for most cases, regex for multiple pairs
  - ✅ Handle edge cases: quoted values with commas, array syntax `[item1,item2]`
  - ✅ Perfect backward compatibility: single values work unchanged
  - ✅ Full validation using existing logic

#### 4B-Post.2 Executor Enhancement for Batch Processing ✅ **COMPLETED**
- [x] ✅ **ExecuteSetCommand Modification**:
  - ✅ Enhanced `Set()` function to handle KeyValuePairs array
  - ✅ Loops through all pairs using existing TOML set logic
  - ✅ Single transaction: load TOML → multiple sets → save once (atomic operation)
  - ✅ Reuses all existing validation, normalization, and error handling

- [x] ✅ **Implementation Strategy**:
  ```go
  // Final Implementation: Clean Unix approach
  // Step 1: Unified KeyValuePairs array in Command struct
  // Step 2: Smart parsing - simple split for single, regex for multiple
  // Step 3: Enhanced Set() loops through all pairs
  for _, kvp := range cmd.KeyValuePairs {
      // Validate, process variables, infer types, set value
      handler.Set(kvp.Key, inferredValue)
  }
  // Step 4: Single atomic save operation
  ```

#### 4B-Post.3 Testing & Validation ✅ **COMPLETED**
- [x] ✅ **Comprehensive Test Suite**:
  - ✅ Validated comma-separated headers: `Auth=token,Accept=json` ✅ Works
  - ✅ Validated quoted values with commas: `Test="value,with,commas"` ✅ Works  
  - ✅ Validated array syntax: `Colors=[red,blue,green]` ✅ Works
  - ✅ Validated error handling for malformed syntax

- [x] ✅ **Real-World Usage Testing**:
  - ✅ Complex configurations work: multiple headers, body fields, arrays
  - ✅ Massive productivity improvement: 50%+ fewer commands for complex setups
  - ✅ Zero regression: all existing single-value functionality works unchanged
  - ✅ Edge cases handled: quotes, commas in values, array syntax, no shell escaping needed

#### 4B-Post.4 Command Scope Definition ✅ **COMPLETED**
**✅ Supported Commands (Comma Syntax):**
- ✅ `saul api set header Auth=token,Accept=json` - Multiple headers in one command
- ✅ `saul api set body name=pikachu,level=25,type=electric` - Multiple body fields 
- ✅ `saul api set query type=electric,generation=1,limit=10` - Multiple query params
- ✅ `saul api set variables pokename=pikachu,trainerId=ash123` - Multiple variables

**✅ Special Syntax Support:**
- ✅ `saul api set header Type="application/json,charset=utf-8"` - Quoted values with commas
- ✅ `saul api set body Tags=[red,blue,green]` - Explicit array syntax with brackets
- ✅ `saul api set url https://api.com` - Special fields remain single-value (correct)

**Phase 4B-Post Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ `saul api set header Auth=Bearer123,Accept=json` sets both headers in one command
- [x] ✅ `saul api set body name=pikachu,level=25` sets both body fields in one command  
- [x] ✅ All existing single-value commands continue working unchanged
- [x] ✅ Dramatically improved testing efficiency (50%+ fewer commands for complex setups)
- [x] ✅ Error handling works correctly for malformed comma syntax  
- [x] ✅ All existing Phase 1-4B functionality unchanged (zero regression)
- [x] ✅ Bonus: Array syntax `[item1,item2]` and quoted comma values work perfectly

**Benefits Achieved:** ✅ **ALL DELIVERED**
- ✅ **Immediate Productivity**: 50%+ fewer commands for complex API configurations
- ✅ **Enhanced Testing**: Much faster iteration, ready for filtering system development
- ✅ **KISS Compliance**: Clean Unix approach - right tool for each job
- ✅ **Zero Risk**: Purely additive feature with perfect backward compatibility  
- ✅ **Robust Foundation**: Perfect base for efficient filter system testing in Phase 4C
- ✅ **No Shell Escaping**: Works without single quotes for most cases

---

### **Phase 4B-Post-2: Space-Separated Key-Value Migration** ✅ **COMPLETED**
*Goal: Migrate from comma-separated to space-separated key-value syntax for universal consistency*

#### 4B-Post-2.1 Parser Migration Analysis ✅ **COMPLETED**
- [x] ✅ **Current System Analysis**:
  - Current: `args[3]` as single comma-separated string: `"name=val1,type=val2"`
  - Proposed: `args[3:]` as multiple space-separated strings: `["name=val1", "type=val2"]`
  - Implementation: Very easy - change from single string parsing to multiple string iteration

- [x] ✅ **Code Simplification Benefits**:
  - Removes complex comma/quote parsing logic entirely
  - Simplifies to basic `key=value` parsing per argument
  - Eliminates quote handling, escaping, and comma conflicts
  - Results in much cleaner, more maintainable code

#### 4B-Post-2.2 Implementation Strategy ✅ **COMPLETED**
- [x] ✅ **Parser Modification** (`parser/command.go`):
  ```go
  // OLD: Single comma-separated string
  if len(args) > 3 {
      keyValueInput := args[3]
      pairs, err := parseCommaSeparatedKeyValues(keyValueInput)
  }

  // NEW: Multiple space-separated strings
  if len(args) > 3 {
      keyValueArgs := args[3:]  // ["name=val1", "type=val2", ...]
      pairs, err := parseSpaceSeparatedKeyValues(keyValueArgs)
  }
  ```

- [x] ✅ **New Function Implementation**:
  ```go
  func parseSpaceSeparatedKeyValues(args []string) ([]KeyValuePair, error) {
      var pairs []KeyValuePair
      for _, arg := range args {
          parts := strings.SplitN(arg, "=", 2)
          if len(parts) != 2 {
              return nil, fmt.Errorf("invalid key=value format: %s", arg)
          }
          pairs = append(pairs, KeyValuePair{
              Key:   strings.TrimSpace(parts[0]),
              Value: strings.TrimSpace(parts[1]),
          })
      }
      return pairs, nil
  }
  ```

- [x] ✅ **Remove Complex Parsing**: Deleted `parseCommaSeparatedKeyValues()` and all comma logic (~100 lines reduced to ~20 lines)

#### 4B-Post-2.3 Migration Benefits ✅ **ACHIEVED**
**Universal Unix Consistency:**
- ✅ Bulk rm: `saul rm preset1 preset2 preset3` (spaces)
- ✅ Bulk set: `saul api set body name=val1 type=val2` (spaces)
- ✅ All bulk operations: Same intuitive space-separated pattern

**Simplified Architecture:**
- ✅ **Much Simpler Code**: Removed ~100 lines of complex comma/quote parsing, reduced to ~20 lines
- ✅ **No Special Syntax**: No quotes, escaping, or comma conflicts to remember
- ✅ **Shell-Friendly**: Works perfectly with tab completion and history
- ✅ **More Maintainable**: Simple iteration vs complex regex patterns

**Enhanced User Experience:**
- ✅ **Cognitive Consistency**: One pattern for all bulk operations
- ✅ **Natural Language**: Matches how people think ("set this AND set that")
- ✅ **Easier Learning**: No special syntax to remember or get wrong

#### 4B-Post-2.4 Usage Examples ✅ **IMPLEMENTED**
```bash
# OLD (comma-separated):
saul api set body name=pikachu,type=electric,level=25
saul api set header Auth=token,Accept=json

# NEW (space-separated):
saul api set body name=pikachu type=electric level=25
saul api set header Auth=token Accept=json

# Consistency with bulk rm:
saul rm preset1 preset2 preset3           # Same pattern
saul api set body name=val1 type=val2     # Same pattern

# Real examples that now work:
saul testapi set header Authorization=Bearer123 Content-Type=application/json
saul testapi set body pokemon.name=pikachu pokemon.level=25 pokemon.type=electric
saul testapi set query type=electric generation=1 limit=10
saul testapi set variables pokename=pikachu trainerId=ash123 region=kanto
```

**Phase 4B-Post-2 Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ All key-value commands use space-separated syntax
- [x] ✅ Much simpler parsing code (removed complex comma logic entirely)
- [x] ✅ Universal space-separated pattern for all bulk operations
- [x] ✅ Perfect shell integration (tab completion, history, etc.)
- [x] ✅ All existing functionality preserved with new syntax
- [x] ✅ Zero regression - all tests pass with space-separated syntax

**Benefits Realized:**
- ✅ **Code Simplification**: Eliminated ~100 lines of complex parsing, removed regexp dependency
- ✅ **Unix Philosophy**: Perfect consistency with bulk rm command pattern
- ✅ **User Experience**: Natural, intuitive syntax that matches shell expectations
- ✅ **Zero Breaking Changes**: All special syntax (URL, method, timeout) works unchanged
- ✅ **Perfect Backward Compatibility**: Single values work identically to before

---

### **Phase 4C: Response Filtering System** ✅ **COMPLETED**
*Goal: Terminal-friendly response filtering to solve API response overflow*

#### 4C.1 Core Filtering Implementation ✅ **COMPLETED**
- [x] ✅ **Dependency Integration**:
  - ✅ Added `github.com/tidwall/gjson` to go.mod for robust JSON path extraction
  - ✅ Integrated gjson into existing HTTP execution pipeline in `response.go`
  - ✅ Zero breaking changes to current functionality

- [x] ✅ **Filter Storage System**:
  - ✅ Created filters.toml handling as 6th file in preset structure
  - ✅ Implemented clean TOML array format for optimal readability:
    ```toml
    fields = ["name", "stats.0.base_stat", "types.0.type.name"]
    ```
  - ✅ Uses existing preset file management patterns seamlessly

- [x] ✅ **Filter Execution Pipeline**:
  - ✅ Integrated filtering into HTTP execution: `HTTP Response → Filter Extraction → Smart TOML Conversion → Display`
  - ✅ Applied filtering before existing Phase 4B response formatting in `src/project/executor/http/response.go`
  - ✅ Perfect Unix philosophy: filtering does one job, TOML conversion does another
  - ✅ Silent error handling: missing fields ignored, no execution breakage

#### 4C.2 Filter Command System ✅ **COMPLETED**
- [x] ✅ **Command Integration**:
  - ✅ Added "filters" as valid target in preset file management
  - ✅ Implemented filter commands using existing space-separated patterns:
    - ✅ `saul api set filters field1=name field2=stats.0.base_stat field3=types.0.type.name`
    - ✅ `saul api check filters` - displays clean TOML array
    - ✅ `saul api edit filters` - full editor support
  - ✅ Routes through existing command executor architecture (zero special parsing)

- [x] ✅ **Field Path Syntax (Industry Standard)**:
  - ✅ Basic fields: `name`, `id`, `stats`
  - ✅ Nested access: `types.0.type.name`, `stats.0.base_stat`
  - ✅ Array indexing: `stats.0`, `moves.5.move.name`
  - ✅ Real-world validated: PokéAPI, JSONPlaceholder field paths work perfectly

#### 4C.3 Testing & Real-World Validation ✅ **COMPLETED**
- [x] ✅ **Real-World API Testing**:
  - ✅ **JSONPlaceholder**: Simple filtering (title, body, id) works perfectly
  - ✅ **PokéAPI**: Complex nested filtering (257KB → 3 fields) works beautifully
  - ✅ Field path extraction accuracy validated with real API structures
  - ✅ Silent error handling tested - missing fields ignored gracefully

- [x] ✅ **Integration with Space-Separated System**:
  - ✅ Enhanced testing using existing space-separated syntax:
    ```bash
    saul api set filters field1=name field2=stats.0.base_stat field3=types.0.type.name
    saul api set url https://pokeapi.co/api/v2/pokemon/1
    saul call api  # Shows only filtered fields in clean TOML
    ```

#### 4C.4 Implementation Architecture ✅ **PERFECT UNIX DESIGN**
- [x] ✅ **Zero Special Parsing**: Uses existing KeyValuePairs system completely
- [x] ✅ **Intelligent Storage**: Special handling in Set command stores values as TOML array
- [x] ✅ **Clean Integration**: Filtering function reads array format with backward compatibility
- [x] ✅ **Consistent UX**: Same space-separated syntax as all other commands
- [x] ✅ **Minimal Code**: Reuses 95% of existing architecture, adds only essential filtering logic

**Phase 4C Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ Large PokéAPI responses (257KB) display only specified fields in terminal
- [x] ✅ Filter commands integrate seamlessly with existing patterns (zero special cases)
- [x] ✅ Field path extraction works perfectly with real-world API structures
- [x] ✅ Silent error handling prevents execution breakage (tested with missing fields)
- [x] ✅ Perfect integration with Phase 4B smart TOML conversion
- [x] ✅ All existing Phase 1-4B-Post functionality unchanged (zero regression)

**Benefits Achieved:**
- ✅ **Terminal Overflow Solved**: 257KB Pokémon response → 3 clean fields
- ✅ **Pure UNIX Philosophy**: One tool (existing parser) handles everything
- ✅ **Incredible Simplicity**: Minimal special cases, maximum code reuse
- ✅ **Production Ready**: Real-world tested with complex APIs
- ✅ **Perfect UX**: Consistent space-separated syntax across all commands

---

### **Phase 4D: Professional Visual Formatting System** ✅ **COMPLETED**
*Goal: Professional visual organization with responsive terminal-friendly display*

#### 4D.1 Core Formatting Engine Implementation ✅ **COMPLETED**
- [x] ✅ **Create Universal Formatting System**:
  - ✅ Created new `src/modules/display/formatter.go` for visual formatting logic
  - ✅ Kept existing `src/modules/display/printer.go` for output mechanics (Error, Success, Warning, etc.)
  - ✅ Added `FormatSection(title, content, metadata string) string` function to formatter.go
  - ✅ Implemented terminal width detection using `golang.org/x/term`
  - ✅ Created responsive separator generation with 80-character target, 80% fallback
  - ✅ Replaced temporary `sections.go` with permanent formatting functions

- [x] ✅ **Clean Separation Architecture**:
  - ✅ Content Generation: Commands produce TOML content using existing handlers
  - ✅ Visual Formatting: `formatter.go` wraps content with clean headers/footers
  - ✅ Output Delivery: `printer.go` handles actual printing (use existing `Plain()` function)
  - ✅ Integration Pattern: `display.Plain(display.FormatSection("Title", content, "metadata"))`

- [x] ✅ **Clean Visual Pattern Implementation**:
  - ✅ Implemented clean three-part structure: Header → Content → Footer
  - ✅ Use Unicode separator `─` (U+2500) for consistent visual boundaries
  - ✅ Clean metadata headers with bullet separators: `Response: 200 OK • 1.2KB • application/json`
  - ✅ Consistent footer width with proper terminal spacing

#### 4D.2 Response Display Enhancement ✅ **COMPLETED**
- [x] ✅ **HTTP Response Integration** (`src/project/executor/http/response.go`):
  - ✅ Wrapped existing Phase 4B JSON→TOML conversion with clean formatting
  - ✅ Added response metadata: status, size, content-type
  - ✅ Integrated with Phase 4C filtering seamlessly
  - ✅ Maintained existing content-type detection and graceful fallback
  - ✅ Added proper file size formatting with `formatBytes()` helper

- [x] ✅ **Enhanced Response Headers**:
  - ✅ Standard responses: `Response: 200 OK • 1.2KB • application/json`
  - ✅ Clean, professional appearance with consistent bullet separators
  - ✅ Human-readable file sizes (bytes, KB, MB)
  - ✅ Preserved existing HTTP execution pipeline

#### 4D.3 Check Command Visual Enhancement ✅ **COMPLETED**
- [x] ✅ **File Display Integration** (`src/project/executor/commands/check.go`):
  - ✅ Wrapped all check command outputs with consistent formatting
  - ✅ File-specific headers: `Headers: 0.5KB • 3 entries`, `Request: 0.1KB • 2 entries`
  - ✅ Smart entry counting with `calculateEntryCount()` function
  - ✅ Maintained current check command functionality (show entire file, not just field)

- [x] ✅ **Universal TOML Display**:
  - ✅ Applied formatting to all TOML file displays consistently
  - ✅ Intelligent entry counting for each file type
  - ✅ File size calculation and display in human-readable format with `formatFileSize()`
  - ✅ Full integration with existing preset file management

#### 4D.4 Terminal Responsiveness ✅ **COMPLETED**
- [x] ✅ **Dynamic Width Management**:
  - ✅ Terminal width detection with graceful fallback to 80 characters
  - ✅ Responsive separator width: 80% of terminal width if < 100 chars, otherwise 80 chars
  - ✅ Consistent separator generation across all display contexts with `calculateSeparatorWidth()`
  - ✅ Cross-platform terminal compatibility using `golang.org/x/term`

- [x] ✅ **Visual Consistency Rules**:
  - ✅ Same separator character `─` throughout application
  - ✅ Consistent bullet separator `•` in all metadata headers
  - ✅ File size in human-readable format (bytes, KB, MB)
  - ✅ Clean opening and closing separators for all formatted content
  - ✅ Added proper spacing from terminal prompt with initial line break

#### 4D.5 Help and List Command Enhancement ✅ **COMPLETED**
- [x] ✅ **Updated Help System** (`cmd/main.go`):
  - ✅ Converted help sections to use new formatter (`FormatSimpleSection`)
  - ✅ Clean, professional help display with consistent visual boundaries
  - ✅ Maintained all existing help content with enhanced readability

- [x] ✅ **Updated List Command**:
  - ✅ Converted preset listing to use new formatter
  - ✅ Clean "No Presets Found" and "Available Presets" displays
  - ✅ Consistent visual presentation across all global commands

**Phase 4D Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ `saul call api` displays responses with professional clean formatting
- [x] ✅ `saul api check url` shows entire request file with consistent visual boundaries
- [x] ✅ All TOML displays use same visual formatting pattern
- [x] ✅ Responsive width works correctly on different terminal sizes
- [x] ✅ Integration with Phase 4B (JSON→TOML) and Phase 4C (filtering) seamless
- [x] ✅ All existing Phase 1-4C functionality unchanged (zero regression)
- [x] ✅ Clean spacing from terminal prompt with proper line breaks

**Benefits Achieved:**
- ✅ **Immediate Professional Appeal**: Every command looks organized and polished
- ✅ **Enhanced Readability**: Clear content boundaries eliminate visual confusion
- ✅ **Perfect Terminal Integration**: Proper spacing and responsive width detection
- ✅ **Universal Consistency**: Same clean formatting across all commands
- ✅ **Zero Breaking Changes**: Pure visual enhancement of existing functionality

**Phase 4D Testing:**
```bash
#!/bin/bash
# Phase 4D Professional Visual Formatting Tests

echo "4D.1 Testing response formatting..."
saul pokeapi call | grep -q "Response:" # Should show formatted header
saul pokeapi call | grep -q "─────" # Should show separators

echo "4D.2 Testing check command formatting..."
saul pokeapi check url | grep -q "Request •" # Should show file type header
saul pokeapi check headers | grep -q "Headers •" # Should show headers header

echo "4D.3 Testing filtered response formatting..."
saul pokeapi set filters field1=name field2=stats.0.base_stat
saul pokeapi call | grep -q "Filtered Response:" # Should show filtered header

echo "4D.4 Testing width responsiveness..."
# Manual test: resize terminal and verify separator width adapts

echo "✓ Phase 4D Professional Visual Formatting: PASSED"
```

**Benefits:**
- **Immediate Professional Appeal**: Every command looks organized and polished
- **Enhanced Readability**: Clear content boundaries eliminate visual confusion
- **Foundation for History**: Professional formatting ready for Phase 4E history display
- **Terminal Optimized**: Responsive design works on all terminal sizes
- **Zero Breaking Changes**: Pure visual enhancement of existing functionality

---

### **Phase 5A: Universal Flag System** ✅ **COMPLETED**
*Goal: Implement --raw flag and establish foundation for all future flags*

#### 5A.1 Flag Parsing Foundation ✅ **COMPLETED**
- [x] ✅ **Parser Enhancement** (`parser/command.go`):
  - ✅ Added `RawOutput bool` field to Command struct
  - ✅ Implemented flag detection logic: arguments starting with `--`
  - ✅ Parse `--raw` flag and set `cmd.RawOutput = true`
  - ✅ Maintained backward compatibility with existing argument parsing
  - ✅ Support combined flag usage: `saul api check url --raw`

- [x] ✅ **Flag Architecture**:
  - ✅ Clean separation: flag parsing vs command parsing via `parseFlags()` function
  - ✅ Forward compatibility: extensible for future flags (`--verbose`, `--format=json`, etc.)
  - ✅ Error handling: unknown flags return clear error messages
  - ✅ Foundation ready for `--help` flag support

#### 5A.2 Check Command Raw Implementation ✅ **COMPLETED**
- [x] ✅ **Conditional Output Logic** (`commands/check.go`):
  - ✅ Special fields (url/method/timeout): `if cmd.RawOutput { fmt.Print(value) } else { display.FormatSection(...) }`
  - ✅ File structures (body/headers/query): `if cmd.RawOutput { fmt.Print(fileContent) } else { display.FormatFileDisplay(...) }`
  - ✅ Proper newlines in raw mode for terminal compatibility
  - ✅ Preserved all existing formatted display as default

- [x] ✅ **Real Usage Examples Working**:
  ```bash
  # Raw for scripting
  saul api check url --raw                    # https://jsonplaceholder.typicode.com/posts/1
  saul api check body --raw                   # Raw TOML file contents (cat behavior)
  
  # Formatted for humans (default)
  saul api check url                          # Shows entire request.toml with context  
  saul api check body                         # Shows body.toml with metadata
  ```

#### 5A.3 Call Command Raw Implementation ✅ **COMPLETED**
- [x] ✅ **Response Raw Mode** (`http/response.go`):
  - ✅ `if cmd.RawOutput { fmt.Print(response.String()) } else { /* existing Phase 4B formatting */ }`
  - ✅ No filtering, no TOML conversion, no metadata headers in raw mode
  - ✅ Pure response body output for automation and scripting
  - ✅ Maintained all existing smart formatting as default

#### 5A.4 List Command Raw Implementation ✅ **COMPLETED**
- [x] ✅ **List Raw Mode** (`cmd/main.go`):
  - ✅ Space-separated preset names: `github httpbin jsonplaceholder pokeapi posttest`
  - ✅ Perfect for shell scripting: `for preset in $(saul list --raw); do saul call $preset --raw; done`
  - ✅ Silent on empty preset list (Unix-friendly)
  - ✅ Maintained formatted display as default

#### 5A.5 Display System Integration ✅ **COMPLETED**
- [x] ✅ **Universal Pattern**: All output-producing commands check `cmd.RawOutput`
- [x] ✅ **Future-Proof**: Established pattern for additional flags (`--verbose`, `--format`, etc.)
- [x] ✅ **Testing**: Comprehensive real-world testing with multiple presets and APIs

**Phase 5A Success Criteria:** ✅ **ALL ACHIEVED**
- [x] ✅ `saul api check url --raw` outputs bare URL value for scripting
- [x] ✅ `saul api check body --raw` outputs raw TOML file contents (cat behavior)
- [x] ✅ `saul call api --raw` outputs raw JSON response without formatting
- [x] ✅ `saul list --raw` outputs space-separated preset names for shell loops
- [x] ✅ All existing formatted output remains default behavior
- [x] ✅ Flag parsing foundation ready for future flag additions
- [x] ✅ Zero regression in existing functionality

**Benefits Achieved:**
- ✅ **Perfect Unix Integration**: Raw mode enables shell scripting and automation
- ✅ **Extensible Architecture**: Clean foundation for future flags (`--verbose`, `--help`, `--format`)
- ✅ **Zero Breaking Changes**: All existing commands work identically by default
- ✅ **Real-World Tested**: Working with JSONPlaceholder, HTTPBin, GitHub APIs

**Development Environment Enhanced:**
- ✅ **Additional Test Presets**: Added `jsonplaceholder`, `httpbin`, `github`, `posttest` for comprehensive testing
- ✅ **Shared Configuration**: Symlinked tenshi user to luar's saul config for unified development
- ✅ **Complete Test Coverage**: All flag functionality validated with real APIs

---

### **Phase 5B: Display System Migration & Check Command Enhancement** ⏳ **MEDIUM PRIORITY**
*Goal: Complete display system migration and improve check command consistency*

#### 5B.1 Check Command Behavior Update ✅ **PLANNED**
- [ ] **Remove Special Case Logic** (`commands/check.go` lines 40-48):
  - Remove bare value printing for URL/method/timeout fields
  - Let all check commands fall through to standard file display
  - Show entire request.toml with context for URL/method/timeout checks
  - Maintain raw flag functionality for bare values when needed

#### 5B.2 Display System Audit ✅ **PLANNED**
- [ ] **Find Remaining fmt.Printf Usage**:
  - Audit codebase for any direct printing not using display system
  - Migrate list command if not already using display.FormatSection
  - Update help system to use display formatting
  - Ensure all user-facing output uses display.Plain(), display.Error(), etc.

**Phase 5B Success Criteria:**
- [ ] `saul api check url` shows entire request.toml with clean formatting
- [ ] All commands use consistent display system formatting
- [ ] No direct fmt.Printf for user-facing output (except raw mode)
- [ ] Visual consistency across all command outputs

---

### **Phase 5C: Response History Storage** ⏳ **MEDIUM PRIORITY**
*Goal: Add response storage and management for debugging workflow*

#### 4E.1 History Storage Management
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

#### 4E.2 History Access Commands
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

#### 4E.3 Enhanced Command Routing
- [ ] **Extended Check Command**:
  - Add history routing to existing `ExecuteCheckCommand`
  - Handle `check history` variations (no args = menu, N = direct, last = recent)
  - Maintain existing check functionality for TOML inspection

- [ ] **Extended Set Command**:
  - Add history configuration to `ExecuteSetCommand`
  - Validate history count values (non-negative integers)
  - Handle `set history 0` to disable without deleting existing history

**Phase 4E Success Criteria:**
- [ ] `saul api set history 5` enables history collection
- [ ] `saul call api` automatically stores responses when history enabled
- [ ] `saul api check history` shows interactive menu of stored responses
- [ ] `saul api check history 1` displays most recent response with Phase 4A formatting
- [ ] `saul api rm history` deletes all history with confirmation prompt
- [ ] History rotation works correctly (keeps last N, deletes oldest)
- [ ] Stored responses use Phase 4A smart formatting when displayed

**Phase 4E Testing:**
```bash
#!/bin/bash
# Phase 4E History Storage Tests

echo "4E.1 Testing history configuration..."
saul testapi set history 3
grep -q 'history_count = 3' ~/.config/saul/presets/testapi/request.toml

echo "4E.2 Testing history storage..."
saul call testapi >/dev/null  # Should store response
[ -d ~/.config/saul/presets/testapi/history ]
[ -f ~/.config/saul/presets/testapi/history/response-001.json ]

echo "4E.3 Testing history access with formatting..."
saul testapi check history | grep -q "1." # Should show menu
saul testapi check history 1 | grep -q "Status:" # Should show formatted response

echo "4E.4 Testing history management..."
echo "y" | saul testapi rm history
[ ! -d ~/.config/saul/presets/testapi/history ]

echo "✓ Phase 4E Response History Storage: PASSED"
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

# NEW: Phase 4B-Post tests (Comma-Separated Syntax)
echo "===== PHASE 4B-POST TESTS: Comma-Separated Syntax Enhancement ====="

echo "4B-Post.1 Testing comma-separated headers..."
saul testapi set header Authorization=Bearer123,Content-Type=application/json
saul testapi check header | grep -q "Authorization.*Bearer123"
saul testapi check header | grep -q "Content-Type.*application/json"

echo "4B-Post.2 Testing comma-separated body fields..."
saul testapi set body pokemon.name=pikachu,pokemon.level=25
saul testapi check body | grep -q "name.*pikachu"
saul testapi check body | grep -q "level.*25"

echo "4B-Post.3 Testing single-value backward compatibility..."
saul testapi set url https://example.com
saul testapi check url | grep -q "example.com"

echo "✓ Phase 4B-Post: Comma-Separated Syntax Enhancement - PASSED"

# NEW: Phase 4C tests (Response Filtering)
echo "===== PHASE 4C TESTS: Response Filtering System ====="

echo "4C.1 Testing filter configuration..."
saul pokeapi set filter name,stats[0],types[0].type.name
saul pokeapi check filter | grep -q "name.*stats\[0\].*types\[0\]"

echo "4C.2 Testing filtered response display..."
saul pokeapi set url https://pokeapi.co/api/v2/pokemon/1
saul call pokeapi | grep -q "name = " # Should show only filtered fields
saul call pokeapi | grep -v "abilities\|moves" # Should NOT show unfiltered fields

echo "4C.3 Testing filter integration with comma syntax..."
saul testapi set filter name,id,types[0].type.name
saul testapi set header Authorization=Bearer123,Content-Type=application/json

echo "✓ Phase 4C: Response Filtering System - PASSED"

# NEW: Phase 4D tests (History Storage)
echo "===== PHASE 4D TESTS: Response History Storage ====="

echo "4D.1 Testing history configuration..."
saul testapi set history 3
grep -q 'history_count = 3' ~/.config/saul/presets/testapi/request.toml

echo "4E.2 Testing history storage..."
echo -e "testuser\n123" | saul call testapi >/dev/null
[ -f ~/.config/saul/presets/testapi/history/response-001.json ]

echo "4D.3 Testing history access with formatting..."
saul testapi check history | grep -q "1\."
saul testapi check history 1 | grep -q "Status:"

echo "✓ Phase 4D: Response History Storage - PASSED"

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

### **Phase 4B-Post Completion Criteria**
- Comma-separated syntax works for header, body, query, and variables commands
- Single-value commands continue working unchanged (backward compatibility)
- Dramatically improved testing efficiency (50% fewer commands for complex setups)
- Error handling works correctly for malformed comma syntax
- All existing Phase 1-4B functionality unchanged (zero regression)

### **Phase 4C Completion Criteria**
- Response filtering system works with real-world APIs (PokéAPI, GitHub, etc.)
- Filter commands integrate seamlessly with existing command patterns
- Field path extraction works with nested JSON and array indexing
- Silent error handling prevents execution breakage for missing fields
- Integration with Phase 4B smart TOML conversion pipeline
- All existing Phase 1-4B-Post functionality unchanged

### **Phase 4D Completion Criteria**
- History system stores and retrieves responses correctly
- History configuration and rotation work properly
- History access commands provide useful debugging workflow using Phase 4B formatting
- All existing Phase 1-4C functionality unchanged

### **Final Project Success**
- All commands from vision.md work correctly
- Variable syntax handles all URL edge cases without conflicts (Phase 3.5)
- No field misclassification bugs in HTTP execution (Phase 3.5)
- Edit command system provides quick field and variable editing workflow (Phase 4A)
- Smart response formatting provides readable output for API development (Phase 4B)
- Comma-separated syntax dramatically improves configuration efficiency (Phase 4B-Post)
- Response filtering solves terminal overflow for large APIs (Phase 4C)
- History system provides valuable debugging workflow (Phase 4D)
- Interactive mode enables efficient preset management (Phase 5)
- Ready for production distribution with advanced features (Phase 6)
- Maintains KISS principles while adding powerful features throughout

---

*This action plan prioritizes comma-separated syntax enhancement (Phase 4B-Post) as the immediate next step for productivity gains, followed by response filtering (Phase 4C) for terminal-friendly API responses, and finally history storage (Phase 4D). This strategic sequence maximizes immediate user value with simple implementations first, building toward more complex features on a proven foundation. The comma-first approach enables efficient testing of filtering systems while maintaining KISS principles throughout.*