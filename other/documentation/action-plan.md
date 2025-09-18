# Better-Curl (Saul) - Action Plan

## Project Overview
Comprehensive implementation plan for Better-Curl (Saul) - a workspace-based HTTP client that eliminates complex curl command pain through TOML-based configuration.

## Current State Analysis

### ✅ **Implemented**
- **Command Parsing**: Basic Command struct + ParseCommand function in `src/project/parser/command.go`
- **TOML Handler**: Complete repurposed TomlHandler with dot notation, merge, JSON conversion in `src/project/handler(repurposed).go`
- **Project Structure**: Modular Go structure following conventions
- **Constants**: Basic command constants and aliases in `src/project/config/constants.go`

### ❌ **Missing Core Components**
- Directory structure management (`~/.config/saul/presets/`)
- Variable substitution system (`?/$` variables)
- HTTP execution engine (go-resty integration)
- Preset/workspace management
- TOML file operations integration
- Single-line command execution
- Interactive mode
- Configuration editing system

### 🔧 **Technical Debt**
- Dependency mismatch: Using `BurntSushi/toml` instead of required `pelletier/go-toml/v1`
- Handler not integrated with command system
- Missing `go-resty/resty/v2` for HTTP client
- No tests or validation system

## Implementation Phases

### **Phase 1: Foundation & TOML Integration** ✅ **COMPLETED**
*Goal: Solid base with working TOML operations and directory management*

#### 1.1 Dependencies & Structure
- [x] Update `go.mod` with correct dependencies:
  - `github.com/pelletier/go-toml v1.9.5` (BurntSushi is an indirect import from toml vars)
  - `github.com/go-resty/resty/v2 v2.7.0` for HTTP client
- [x] Fix TomlHandler package declaration (currently `package main`)
- [x] Move TomlHandler to `src/project/toml/handler.go`
- [x] Create directory management utilities in `src/project/presets/manager.go`

**AI Execution Notes - Phase 1:**
```go
// Required exact function signatures for Phase 1:

// In src/project/presets/manager.go:
func CreatePresetDirectory(name string) error
func ListPresets() ([]string, error)
func DeletePreset(name string) error
func GetPresetPath(name string) string  // Reads from settings.toml, returns full path
func GetConfigDir() string              // Reads from settings.toml, returns full path
func LoadSettings() (*Settings, error)  // Loads src/settings/settings.toml

type Settings struct {
    Directories struct {
        ConfigDir  string `toml:"config_dir"`
        AppDir     string `toml:"app_dir"`
        PresetsDir string `toml:"presets_dir"`
    } `toml:"directories"`
}

// In src/project/toml/handler.go (moved from handler(repurposed).go):
package toml  // Change from "package main"

// In src/project/presets/manager.go:
func LoadPresetFile(preset, fileType string) (*toml.TomlHandler, error)
func SavePresetFile(preset, fileType string, handler *toml.TomlHandler) error

// fileType values: "headers", "body", "query", "request", "variables"
```

**Expected Command Flow - Phase 1:**
```bash
Input: `saul list`
→ main.go calls parser.ParseCommand(["list"])
→ Returns Command{Global: "list"}
→ main.go calls manager.ListPresets()
→ Output: Lists all directories in ~/.config/saul/presets/

Input: `saul myapi` (preset creation)
→ Command{Preset: "myapi"}
→ manager.CreatePresetDirectory("myapi")
→ Creates ~/.config/saul/presets/myapi/{headers,body,query,config}.toml
```

**Integration Handoff - Phase 1 → Phase 2:**
- **Phase 1 Outputs**: Working directory structure, TOML file loading/saving
- **Phase 2 Inputs**: Expects `LoadPresetFile()` and `SavePresetFile()` functions
- **Critical**: Phase 2 will call `LoadPresetFile("myapi", "body")` expecting working TomlHandler

#### 1.2 Directory Management System
- [x] Implement `CreatePresetDirectory(name string)` function
- [x] Implement `ListPresets()` function
- [x] Implement `DeletePreset(name string)` function
- [x] Create default TOML files (headers.toml, body.toml, query.toml, request.toml, variables.toml)
- [x] Handle `~/.config/saul/` directory creation and permissions

#### 1.3 TOML File Operations
- [x] Integrate TomlHandler with preset directory structure
- [x] Implement `LoadPresetFile(preset, fileType string)` function
- [x] Implement `SavePresetFile(preset, fileType string, handler *TomlHandler)` function
- [x] Add error handling for missing files/directories

**Phase 1 Success Criteria:** ✅ **ALL PASSED**
- [x] `saul list` shows all presets
- [x] `saul myapi` creates preset directory with empty TOML files
- [x] `saul rm myapi` removes preset with confirmation
- [x] All file operations work reliably with proper error handling

**Phase 1 Testing:**
```bash
# Test directory operations
go run cmd/main.go list
go run cmd/main.go mytest  # Should create preset directory
ls ~/.config/saul/presets/mytest/  # Should show 4 TOML files
go run cmd/main.go rm mytest  # Should remove preset
```

---

### **Phase 2: Core TOML Operations & Variable System** ✅ **COMPLETED**
*Goal: Working set/get operations with variable substitution*

**AI Execution Notes - Phase 2:**
```go
// Required exact function signatures for Phase 2:

// In src/project/executor/commands.go (new file):
func ExecuteSetCommand(cmd parser.Command) error
func ExecuteGetCommand(cmd parser.Command) (interface{}, error)  // For debugging

// Expected Command struct usage:
// cmd.Preset = "myapi"
// cmd.Command = "set"
// cmd.Target = "body"  // "headers", "body", "query", "config"
// cmd.Key = "pokemon.stats.hp"
// cmd.Value = "100"

// Variable detection:
func DetectVariableType(value string) (isVariable bool, varType string, varName string)
// Returns: (true, "soft", "name") for "?name"
// Returns: (true, "hard", "attack") for "$attack"
// Returns: (false, "", "") for "pikachu"
```

**Expected Command Flow - Phase 2:**
```bash
Input: `saul myapi set body pokemon.stats.hp=100`
→ ParseCommand returns Command{Preset:"myapi", Command:"set", Target:"body", Key:"pokemon.stats.hp", Value:"100"}
→ ExecuteSetCommand(cmd):
  1. LoadPresetFile("myapi", "body") → gets TomlHandler
  2. handler.Set("pokemon.stats.hp", 100)  // Auto-converts string "100" to int
  3. SavePresetFile("myapi", "body", handler)
→ Result: ~/.config/saul/presets/myapi/body.toml contains:
```toml
[pokemon]
[pokemon.stats]
hp = 100
```

**Integration Handoff - Phase 2 → Phase 3:**
- **Phase 2 Outputs**: Working set/get commands, variable detection system
- **Phase 3 Inputs**: Expects working TOML manipulation, variable list for prompting
- **Critical**: Phase 3 will merge all TOML files and resolve variables during `fire`

#### 2.1 Command Integration
- [ ] Connect Command struct to TOML operations
- [ ] Implement `ExecuteSetCommand(cmd Command)` function
- [ ] Implement `ExecuteGetCommand(cmd Command)` function (for debugging)
- [ ] Add validation for command structure and arguments

#### 2.2 TOML Set Operations
- [ ] Implement dot notation parsing: `body.pokemon.stats.hp=100`
- [ ] Handle different value types (string, int, bool, array)
- [ ] Array detection and parsing: `tags=red,blue,green`
- [ ] Type inference without explicit declarations

#### 2.3 Variable System Foundation
- [x] Create variable detection logic (`?` and `@` prefixes) - changed from $ to @ to avoid shell conflicts
- [x] Implement variable storage in `variables.toml` (hard variables only - soft variables never stored)
- [x] Implement check command for TOML inspection
- [x] Add smart target routing and aliases
- [ ] Implement variable prompting system (basic version) - moved to Phase 3

**Phase 2 Success Criteria:** ✅ **ALL PASSED**
- [x] `saul myapi set body pokemon.name=pikachu` works correctly
- [x] `saul myapi set header Content-Type=application/json` works correctly  
- [x] `saul myapi set body pokemon.stats.hp=100` creates proper nested structure
- [x] `saul myapi set body tags=red,blue,green` creates TOML array
- [x] Variable syntax `pokemon.name=?` and `pokemon.level=@` are detected and stored
- [x] Special request syntax `saul myapi set url/method/timeout` works correctly
- [x] Check command `saul myapi check url` works for inspection
- [x] 5-file lazy creation system works correctly

**Phase 2 Testing:** ✅ **AUTOMATED**
```bash
# Run comprehensive test suite
./test_suite.sh

# Tests all Phase 2 functionality:
# - Special request syntax (url/method/timeout)
# - Regular TOML operations (body/headers/query)
# - Variable detection (@ and ? syntax)
# - Check command functionality
# - 5-file lazy creation system
# - Target aliases and validation
# - Error handling and edge cases
```

---

### **Phase 3: HTTP Execution Engine** ✅ **COMPLETED**
*Goal: Working call command that executes HTTP requests*

**AI Execution Notes - Phase 3:**
```go
// Required exact function signatures for Phase 3:

// In src/project/executor/http.go (new file):
func ExecuteFireCommand(cmd parser.Command) error
func BuildHTTPRequest(preset string) (*resty.Request, error)
func PromptForVariables(variables []Variable) (map[string]string, error)
func MergePresetFiles(preset string) (*toml.TomlHandler, error)

// Variable struct for prompting:
type Variable struct {
    Name     string  // "name", "attack"
    Type     string  // "soft", "hard"
    Current  string  // Current value for hard variables (from config.toml)
}

// HTTP execution flow:
func ExecuteHTTPRequest(req *resty.Request, method, url string) ([]byte, error)
```

**Expected Command Flow - Phase 3:** ✅ **IMPLEMENTED**
```bash
Input: `saul call myapi`
→ ExecuteCallCommand(Command{Global:"call", Preset:"myapi"}):
  1. Check preset exists (prevents calling non-existent presets)
  2. MergePresetFiles("myapi") → merges request.toml + headers.toml + body.toml + query.toml
  3. Extract variables: finds "?name" and "@attack" in merged data
  4. PromptForVariables() → prompts user:
     name: ____                    # Soft variable (always empty)
     attack: 80_                   # Hard variable (shows current value)
  5. Replace variables in merged TOML with user input
  6. Convert to JSON for body, extract headers/query separately
  7. BuildHTTPRequest() → creates go-resty request with all components
  8. ExecuteHTTPRequest() → sends HTTP request
  9. DisplayResponse() → shows formatted response with status, headers, pretty JSON

Input: `saul call myapi --persist`
→ Same flow but prompts for hard variables too and saves new values to variables.toml
```

**Integration Handoff - Phase 3 → Phase 4:** ✅ **COMPLETED**
- **Phase 3 Outputs**: Complete HTTP execution engine, variable resolution system, TOML merging
- **Phase 4 Inputs**: All core functionality ready, needs enhanced command routing and remaining features
- **Critical**: Phase 4 will add remaining management commands and advanced features

#### 3.1 HTTP Request Builder ✅ **COMPLETED**
- [x] Implement `BuildHTTPRequest(preset string)` function
- [x] Merge all TOML files (request.toml + headers.toml + body.toml + query.toml + variables.toml)
- [x] Convert merged TOML to go-resty request structure
- [x] Handle different HTTP methods (GET, POST, PUT, DELETE, etc.)

#### 3.2 Variable Resolution System ✅ **COMPLETED**
- [x] Implement variable prompting during `call` command
- [x] Handle soft variables (`?`) - always prompt with empty input
- [x] Handle hard variables (`@`) - prompt with current value shown
- [x] Implement `--persist` flag for hard variable updates (basic version)
- [x] Store resolved variables in memory for request execution
- [x] Smart Variable Deduplication feature

#### 3.3 HTTP Execution ✅ **COMPLETED**
- [x] Integrate go-resty for HTTP requests
- [x] Implement clean response formatting with pretty JSON
- [x] Add comprehensive error handling and validation
- [x] Handle HTTP errors gracefully with status display
- [x] Add timeout configuration support

**Phase 3 Success Criteria:** ✅ **ALL PASSED**
- [x] `saul myapi set url https://jsonplaceholder.typicode.com/posts/1` and `saul myapi set method GET` sets endpoint (special syntax)
- [x] `saul call myapi` executes HTTP request successfully
- [x] Variable prompting works for both `?` and `@` variables
- [x] `saul call myapi --persist` framework ready for hard variable updates
- [x] Response is displayed cleanly and readable with pretty formatting
- [x] All HTTP methods work correctly (GET, POST, PUT, DELETE)
- [x] TOML file merging works across all 5 files
- [x] Comprehensive test suite validates all functionality

**Phase 3 Testing:** ✅ **COMPREHENSIVE SUITE IMPLEMENTED**
```bash
# All tests automated in other/testing/test_suite.sh
# Phase 3 tests include:
# - Basic call command with GET requests
# - POST requests with JSON body conversion
# - Variable prompting system validation
# - All HTTP methods (GET, POST, PUT, DELETE)
# - Headers and complex request handling
# - Error handling for missing URLs and non-existent presets
# - TOML file merging validation
# - Test isolation with backup/restore functionality
```

---

### **Phase 4: Complete Command System**
*Goal: All single-line commands working as specified in vision*

#### 4.1 Command Execution Router
- [ ] Implement `ExecuteCommand(cmd Command)` router function
- [ ] Handle global commands: `version`, `list`, `rm`
- [ ] Handle preset commands: `set`, `fire`, `edit` (basic)
- [ ] Add comprehensive error messages and help text

#### 4.2 Configuration Commands
- [ ] Implement `set url METHOD https://...` command
- [ ] Implement `set header key=value` command
- [ ] Implement `set body object.field=value` command
- [ ] Implement `set query param=value` command
- [ ] Add validation for each command type

#### 4.3 Management Commands
- [ ] Implement `version` command with proper version display
- [ ] Implement `list` command with formatted preset listing
- [ ] Implement `rm preset` command with confirmation prompt
- [ ] Add `help` command with usage examples

**Phase 4 Success Criteria:**
- All commands from vision.md work correctly
- Error messages are helpful and specific
- `saul help` shows comprehensive usage
- `saul version` shows correct version info
- All preset operations work reliably

**Phase 4 Testing:**
```bash
# Test all commands
go run cmd/main.go version
go run cmd/main.go help
go run cmd/main.go list
go run cmd/main.go testapi set url POST https://httpbin.org/post
go run cmd/main.go testapi set header Content-Type=application/json
go run cmd/main.go testapi set body message=hello
go run cmd/main.go testapi fire
```

---

### **Phase 5: Interactive Mode**
*Goal: Working interactive shell for preset management*

#### 5.1 Interactive Shell
- [ ] Implement `EnterInteractiveMode(preset string)` function
- [ ] Create command loop with `> ` prompt
- [ ] Handle `exit` command to leave interactive mode
- [ ] Add command history and basic editing

#### 5.2 Interactive Commands
- [ ] All `set` commands work in interactive mode
- [ ] `fire` command works in interactive mode
- [ ] Tab completion for commands (optional)
- [ ] Clear error handling within interactive session

#### 5.3 User Experience Enhancements
- [ ] Show current preset in prompt: `[myapi]> `
- [ ] Add context-aware help in interactive mode
- [ ] Handle Ctrl+C gracefully
- [ ] Add command validation before execution

**Phase 5 Success Criteria:**
- `saul myapi` enters interactive mode successfully
- All commands work identically in interactive vs single-line mode
- Exit/Ctrl+C handling works properly
- User experience feels natural and responsive

**Phase 5 Testing:**
```bash
# Test interactive mode
go run cmd/main.go myapi
> set header Authorization=Bearer123
> set body pokemon.name=pikachu
> fire
> exit
```

---

### **Phase 6: Advanced Features & Polish**
*Goal: Complete feature set with editing and advanced management*

#### 6.1 File Editing Integration
- [ ] Implement `edit header` command to open headers.toml
- [ ] Implement `edit body` command to open body.toml
- [ ] Implement `edit config` command to open config.toml
- [ ] Detect default editor from environment ($EDITOR)

#### 6.2 Advanced Variable Features
- [ ] Custom variable names: `pokemon.name=?pokename`
- [ ] Variable validation and type hints
- [ ] Variable reuse across multiple requests
- [ ] Export/import variable sets

#### 6.3 Production Readiness
- [ ] Comprehensive error handling for all edge cases
- [ ] Add proper logging system
- [ ] Performance optimization for large TOML files
- [ ] Cross-platform path handling
- [ ] Build system for binary distribution

**Phase 6 Success Criteria:**
- `saul myapi edit body` opens body.toml in default editor
- All edge cases handled gracefully
- Performance is acceptable for typical usage
- Ready for end-user distribution

## Comprehensive Testing Strategy

### **Single Test File: `test_suite.sh`**
Expandable comprehensive test script that validates all implemented functionality across phases:

```bash
#!/bin/bash
# test_suite.sh - Expandable comprehensive test suite

# Run current tests
./test_suite.sh

# Key features:
# - Phase-organized test sections  
# - Expandable as new phases are implemented
# - Validates all functionality comprehensively
# - Clear pass/fail reporting
# - Automated cleanup and setup
```

### **Testing Philosophy**
- **Phase-organized**: Tests grouped by implementation phases
- **Expandable**: Easy to add new test sections as features are implemented
- **Comprehensive**: One test file covers entire project systematically
- **Practical**: Tests real usage scenarios from vision.md
- **Fast**: Quick feedback loop for development with clear reporting
- **Automated**: Self-contained with setup and cleanup

## Development Guidelines

### **KISS Principles**
- **Simple**: Each function has one clear responsibility
- **Clean**: Self-documenting code with minimal comments
- **Intelligent**: Smart type detection and error handling
- **Resilient**: Graceful handling of edge cases

### **Go Best Practices**
- Follow standard Go project layout
- Use Go modules properly
- Error handling at every boundary
- Clear package separation of concerns
- Minimal external dependencies

### **Learning Focus**
- Understand each component before moving to next phase
- Ask questions about architectural decisions
- Review code together before committing
- Focus on comprehension over speed

## Risk Mitigation

### **Potential Issues**
- **File Permission Problems**: Handle `~/.config/saul/` creation carefully
- **TOML Parsing Errors**: Robust error handling for malformed files
- **Variable Substitution Edge Cases**: Test with special characters
- **HTTP Request Failures**: Timeout and retry logic
- **Cross-platform Compatibility**: Path handling differences

### **Mitigation Strategies**
- Test on clean environment frequently
- Add comprehensive error messages
- Validate input at every boundary
- Test with edge cases early
- Keep backups of working TOML files

## Success Metrics

### **Phase Completion Criteria**
Each phase must pass its testing section completely before proceeding to the next phase.

### **Final Project Success**
- All commands from vision.md work correctly
- `test_saul.sh` passes completely
- Ready for real-world usage
- Code is clean, understandable, and maintainable
- Performance is acceptable for typical use cases

---

*This action plan prioritizes incremental development with continuous validation, ensuring each phase builds a solid foundation for the next while maintaining code quality and user experience standards.*