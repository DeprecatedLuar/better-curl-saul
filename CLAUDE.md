# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Vision and Approach

This project is **Better-Curl (Saul)** - a workspace-based HTTP client designed to eliminate the pain of complex curl commands with JSON payloads. 

**README.md is the source of truth for AI-assisted development** - it documents all core ideas, concepts, and foundational vision that guide development decisions. Always reference this file for project scope, feature requirements, and architectural direction. The action-plan.md handles specific technical implementation details.

**Collaborative Development Philosophy:**
- Follow KISS principles: clean, intelligent, self-maintained, resilient code above all else
- Avoid over-engineering at all costs - prioritize simple and clean solutions
- Always engage in strategic discussion before implementation
- Break down complex tasks into understandable components
- Explain architectural decisions and reasoning during development

## Development Commands

**Build and Run:**
```bash
go run cmd/main.go [command]      # Run from project root
go mod tidy                       # Manage dependencies
go build -o saul cmd/main.go      # Build binary

# Example commands to test:
go run cmd/main.go version
go run cmd/main.go pokeapi
go run cmd/main.go pokeapi set body pokemon.name=pikachu
```

## Current Architecture State

**✅ PHASE 0 INFRASTRUCTURE CLEANUP COMPLETED (2025-09-22)**
- Global state eliminated with proper SessionManager implementation
- Module imports validated and cleaned
- Code follows Go conventions and compiles successfully
- 6% compliance improvement in code review metrics

**✅ PACKAGE REORGANIZATION COMPLETED (2025-09-22)**
- Consolidated single-file packages (`delegation/`, `session/`, `parser/`) → `core/`
- Renamed `executor/` → `handlers/` for clearer intent
- Eliminated over-engineering with tiny packages while maintaining 250-line constraint
- Updated all imports and function references across codebase
- All builds passing with improved Go convention compliance

**✅ CONSOLE OUTPUT STANDARDIZATION COMPLETED (2025-09-22)**
- Fixed main.go display violations: 6 direct `fmt.Print*` → proper `display.*` functions
- Architectural refactor: `modules/display/history.go` → `project/handlers/commands/history.go`
- Proper layer separation: Business logic moved from generic display to project layer
- Preserved Unix raw output philosophy: 20 legitimate `fmt.Print*` kept for `--raw` piping
- Import cleanup and build validation: All refactoring tested and working

**Project Structure:**
```
better-curl-saul/
├── go.mod                        # Go module (github.com/DeprecatedLuar/better-curl-saul)
├── README.md                     # Complete project specification (moved from other/documentation/vision.md)
├── other/documentation/action-plan.md # Development action plan
├── cmd/
│   └── main.go                  # Clean entry point - dependency injection pattern
├── src/
│   ├── modules/
│   │   ├── errors/              # Centralized error handling system
│   │   │   └── messages.go      # All error/warning constants with casual tone
│   │   └── display/             # Universal printing system
│   │       ├── printer.go       # Error, Success, Warning, Info, Tip, Plain functions
│   │       └── sections.go      # Section formatting (temporary)
│   └── project/
│       ├── core/                # ✅ REORGANIZED: CLI fundamentals consolidated
│       │   ├── parser.go        # Command struct + ParseCommand function
│       │   ├── session.go       # SessionManager with encapsulated state
│       │   └── delegation.go    # System command delegation
│       ├── handlers/            # ✅ RENAMED: Command & HTTP handling logic (was executor/)
│       │   ├── commands/        # Command execution subfolder
│       │   │   ├── set.go       # Set command handler
│       │   │   ├── get.go       # Get command handler
│       │   │   ├── check.go     # Check command handler
│       │   │   ├── edit.go      # Edit command handler
│       │   │   └── utils.go     # Command utilities
│       │   ├── variables/       # Variable processing subfolder
│       │   │   ├── detection.go # Variable detection logic
│       │   │   ├── prompting.go # Variable prompting logic
│       │   │   └── storage.go   # Variable storage logic
│       │   ├── http/            # HTTP execution subfolder
│       │   │   ├── client.go    # HTTP client setup and execution
│       │   │   └── response.go  # Response formatting and display
│       │   ├── http.go          # Main HTTP execution logic
│       │   ├── variables.go     # Variable re-export layer
│       │   └── validation.go    # Input validation logic
│       ├── presets/             # Workspace management
│       │   ├── manager.go       # Preset directory management
│       │   ├── files.go         # TOML file operations
│       │   └── history.go       # Response history management
│       ├── toml/                # TOML manipulation
│       │   ├── handler.go       # Core TOML manipulation
│       │   ├── json.go          # JSON conversion logic
│       │   └── io.go            # File I/O operations
│       └── config/              # Configuration
│           ├── constants.go     # Constants and command aliases
│           └── settings.go      # Configuration loading
```

**Core Architecture Concepts (from README.md):**
- **Presets**: Folders in `~/.config/saul/presets/[preset-name]/` containing TOML files
- **5-File Structure**: headers.toml, body.toml, query.toml, request.toml, variables.toml (Unix philosophy)
- **Variable System**: Soft variables (`{?name}`) always prompt, hard variables (`{@name}`) persist in variables.toml
- **Special Syntax**: Request commands use no = syntax: `set url https://...`, `set method POST`
- **Command Modes**: Both interactive mode and single-line commands
- **Data Flow**: TOML → Variable resolution → JSON conversion → HTTP execution

**Current Implementation Status:**
- ✅ **Phase 1 Complete**: Foundation & TOML Integration
  - Modular Go structure following conventions
  - Command parsing system with global and preset commands  
  - Directory management with lazy file creation
  - TOML file operations integrated
- ✅ **Phase 2 Complete**: Core TOML Operations & Variable System
  - 5-file structure (Unix philosophy): body, headers, query, request, variables
  - Special request syntax: `set url/method/timeout` (no = sign)
  - Variable system: `{@}` for hard variables, `{?}` for soft variables
  - Target normalization and validation
  - Comprehensive test suite validation
- ✅ **Phase 3 Complete**: HTTP Execution Engine
  - `saul preset call` command fully functional
  - Variable prompting and substitution system
  - HTTP client integration using go-resty
  - Support for all major HTTP methods
  - JSON body conversion and pretty-printed responses
- ✅ **Phase 3.5 Complete**: Architecture & Variable Syntax Fix
  - Separate handler implementation (no field misclassification)
  - Braced variable syntax `{@name}` and `{?name}` (no URL conflicts)
  - Real-world URL support: `https://api.github.com/@username` works correctly
  - All existing functionality preserved with new syntax
- ✅ **Phase 3.7 Complete**: Variable Detection System Simplification
  - Replaced complex TOML structure parsing with simple regex-based detection
  - Fixed nested TOML variable detection: `[pokemon] name = "{@pokename}"` now works
  - Reduced ~100 lines of complex code to ~20 lines of regex
  - Zero breaking changes, same user experience, much more reliable
- ✅ **Phase 4A Complete**: Edit Command System
  - Field-level editing with pre-filled readline prompts
  - Interactive terminal editing experience with cursor movement
  - Uses existing validation and TOML patterns
  - Zero regression - purely additive feature
- ✅ **Phase 4B Complete**: Response Formatting System
  - Smart JSON→TOML conversion for optimal readability
  - Intelligent content-type detection with graceful fallback
  - HTTP subfolder refactoring for clean architecture
  - Real-world tested with JSONPlaceholder, PokéAPI, HTTPBin, GitHub APIs
- ✅ **Phase 4B-Post Complete**: Comma-Separated Syntax Enhancement
  - Unix-like parsing approach with unified KeyValuePairs array architecture
  - Multiple key=value support: `Auth=token,Accept=json` (50%+ fewer commands)
  - Quoted values with commas: `Type="application/json,charset=utf-8"`
  - Explicit array syntax: `Tags=[red,blue,green]` with intelligent bracket detection
  - Perfect backward compatibility, zero regression, no shell escaping needed
- ✅ **Bulk Operations System Complete**: Universal Space-Separated Pattern
  - Bulk removal: `saul rm preset1 preset2 preset3` (space-separated arguments)
  - Continue + warn approach: delete existing presets, warn about non-existent
  - Parser enhancement: `Targets []string` field for multiple targets
  - Command execution: graceful error handling with warnings to stderr
  - Foundation for universal space-separated bulk operations across all commands
- ✅ **Phase 4B-Post-2 Complete**: Space-Separated Key-Value Migration
  - Universal space-separated pattern: `saul api set body name=val1 type=val2`
  - Code simplification: Eliminated ~100 lines of complex parsing logic
  - Perfect Unix consistency: Same approach for all bulk operations
  - Zero regression: All existing functionality preserved with cleaner syntax
- ✅ **Phase 4C Complete**: Response Filtering System
  - Terminal overflow solved: 257KB APIs → filtered fields display
  - Pure UNIX design: Uses existing KeyValuePairs system (zero special parsing)
  - Clean syntax: `saul api set filters field1=name field2=stats.0.base_stat field3=types.0.type.name`
  - TOML array storage: `fields = ["name", "stats.0.base_stat", "types.0.type.name"]`
  - Real-world tested: PokéAPI, JSONPlaceholder complex filtering works perfectly
  - Silent error handling: Missing fields ignored gracefully
- ✅ **Phase 4D Complete**: Terminal Session Memory System
  - Terminal-scoped preset memory: `saul api set body name=val` → `saul check body` (no preset needed)
  - TTY-based session isolation: Each terminal maintains independent current preset
  - Automatic preset injection: Action commands (`set`, `check`, `get`, `edit`) use current preset
  - Clean preset switching: Any explicit preset command updates current session
  - Session files: `~/.config/saul/.session_[tty]` (terminal-specific, auto-cleanup on startup)
  - Zero overhead: ~50 lines of code, pure stdlib implementation with automatic stale session cleanup
- ✅ **Phase 6A Complete**: System Command Delegation - Unix philosophy implementation
  - Replaced custom `saul list` with system command delegation (`saul ls`)
  - Whitelist-based security: only safe commands (ls, exa, lsd, tree, dir) allowed
  - Working directory automatically set to presets folder for all delegated commands
  - Perfect workspace visibility: see actual TOML files and directory structure
- ✅ **Phase 4E Complete**: Response History System with Split Command Architecture
  - Unix list-then-select pattern: `saul check history` (list) + `saul check response N` (fetch)
  - Sequential file naming: `001.json`, `002.json`, `003.json` (CLI research-backed standard)
  - Hidden directory storage: `~/.config/saul/presets/[preset]/.history/` (dot-prefixed)
  - Metadata-in-content: timestamp, method, URL, status, duration stored inside JSON files (no filename clutter)
  - Simple configuration: `saul set history N` (just the number, Unix-style)
  - Consistent filtering: History displays same filtered TOML view as live responses
  - Minimal implementation: Extracted `FormatResponseContent()` function for code reuse
  - Zero code duplication: Same filtering + TOML conversion pipeline for live and historical responses
  - Raw mode support: `saul check history --raw`, `saul check response 1 --raw` for automation
  - Full data preservation: Stores complete responses, applies filtering at display time

## Codebase Architecture Flow

**Command Flow (Understanding the Complete Request Lifecycle):**

```
User Input → Command Parsing → Command Routing → Command Execution → TOML Operations
```

### 1. Entry Point: `cmd/main.go`
- **Purpose**: Clean entry point following Go conventions
- **Flow**: `os.Args[1:]` → `core.ParseCommand()` → `executeCommand()` → Route to handlers
- **Routing**: Global commands (`list`, `rm`) vs Preset commands (`set`, `check`, `get`, `call`)

### 2. Command Parsing: `src/project/core/parser.go`
- **Input**: Raw command line arguments
- **Output**: `Command` struct with structured fields
- **Special Logic**:
  - Special request syntax: `saul api set url https://...` (no = sign)
  - Regular TOML syntax: `saul api set body name=value` (with = sign)
  - Check command routing: `saul api check url` → auto-maps to request target
  - **Bulk operations**: `Targets []string` field for space-separated arguments
  - **Universal pattern**: `saul rm preset1 preset2 preset3` (spaces for all bulk operations)

### 3. Command Execution: `src/project/handlers/commands/`
- **Current Commands**: `Set()`, `Check()`, `Get()`, `Edit()` functions
- **TOML Integration**: Uses `presets.LoadPresetFile()` → TOML handler operations → `presets.SavePresetFile()`
- **Validation**: Target normalization, request field validation, variable detection
- **Unix Philosophy**: Silent success on completion

### 4. TOML Operations: `src/project/toml/handler.go`
- **Core Methods**: `.Get()`, `.Set()`, `.Has()`, `.Delete()`, `.Write()`
- **Conversion**: `.ToJSON()` for HTTP requests, `.ToJSONPretty()` for display
- **Advanced**: `.Merge()`, `.Clone()`, dot notation for nested fields

### 5. Preset Management: `src/project/presets/`
- **File Structure**: `~/.config/saul/presets/[preset]/[file].toml`
- **6-File System**: body.toml, headers.toml, query.toml, request.toml, variables.toml, filters.toml
- **Operations**: `LoadPresetFile()`, `SavePresetFile()`, `CreatePresetDirectory()`

### 6. HTTP Execution: `src/project/handlers/http/` (Phase 4B Refactored)
- **client.go**: HTTP client setup, request execution, error handling
- **response.go**: Smart response formatting with filtering (JSON→Filter→TOML conversion)
- **request.go**: HTTP request building from TOML handlers
- **Variable Resolution**: Load variables.toml → Prompt for missing → Substitute in all files
- **Request Building**: Separate handlers per file → Extract components → Build HTTP request
- **Filtering Pipeline**: Load filters.toml → Apply gjson filtering → Smart TOML display
- **Execution**: go-resty HTTP client → Filter → Smart-formatted response display

### 7. Variable System: `src/project/handlers/variables/`
- **Detection**: `{@name}` (hard - stored) vs `{?name}` (soft - always prompt)
- **Resolution**: Prompt user → Store hard variables → Substitute in TOML before HTTP
- **Integration**: Works seamlessly with URL variables, no conflicts

**Key Architecture Principles:**
- **Clean Separation**: Each file/package has single responsibility
- **TOML-First**: All configuration stored in human-readable TOML files
- **Variable Flexibility**: Soft vs hard variables for different workflow needs
- **Unix Philosophy**: Small, composable functions that do one thing well
- **Zero Dependencies**: Edit commands use existing TOML manipulation, no new complexity
- **Centralized Error Handling**: All error messages use constants from `src/modules/errors/messages.go` with consistent casual tone
- **Raw-First Display Philosophy**: File operations output raw content for Unix composition, only HTTP responses get pretty formatting

**Edit Command Integration Points:**
- **Parser**: Add "edit" recognition in `ParseCommand()`
- **Router**: Add case in `executePresetCommand()` switch
- **Executor**: Add `ExecuteEditCommand()` using existing patterns:
  - Load preset file with `presets.LoadPresetFile()`
  - Get current value with `handler.Get()`
  - Prompt user for new value
  - Set new value with `handler.Set()`
  - Save with `presets.SavePresetFile()`

## TOML Manipulation System

**Core Library**: Repurposed TomlHandler from toml-cli project
- **Location**: `src/project/toml/handler.go`
- **Purpose**: Dot notation TOML manipulation for Saul commands
- **Key methods**: `.Set()`, `.Get()`, `.ToJSON()` for HTTP conversion

**Integration Pattern:**
- Regular: `saul pokeapi set body pokemon.stats.hp=100`
- Special: `saul pokeapi set url https://api.com` (no = sign)
- Variables: `saul pokeapi set body name={@pokename}` (hard) or `name={?}` (soft)
- Flow: Parse command → TomlHandler.Set("pokemon.stats.hp", 100) → Write to appropriate .toml file

**Variable Substitution**: Variables stored in variables.toml (hard only), resolved during preset `call` command

## Response History System Architecture

**Split Command Pattern (Unix Philosophy):**
- **LIST**: `saul check history` → Show metadata (method, URL, status, timestamp) for all responses
- **FETCH**: `saul check response N` → Show specific response content with formatting
- **DEFAULT**: `saul check response` → Most recent response (no number needed for 80% use case)

**File Organization:**
- **Location**: `~/.config/saul/presets/[preset]/.history/`
- **Naming**: `001.json`, `002.json`, `003.json` (sequential, CLI standard)
- **Content**: JSON with embedded metadata + raw response data
- **Rotation**: Automatic when limit exceeded (keeps newest N, removes oldest)

**JSON Structure:**
```json
{
  "metadata": {
    "timestamp": "2025-01-15T14:32:45Z",
    "method": "POST",
    "endpoint": "/api/users",
    "status": 201,
    "duration": "0.234s",
    "size": "1.2KB"
  },
  "response": { /* raw response data */ }
}
```

**Configuration:**
- **Syntax**: `saul set history N` (just the number, Unix-style)
- **Storage**: `history_count` in `request.toml` alongside other settings
- **Range**: 0-100 (0 = disabled)

**Benefits:**
- **Discoverable**: List-then-select workflow shows what's available
- **Research-backed**: Sequential naming follows universal CLI patterns
- **Clean**: Metadata in content, not cluttered filenames
- **Efficient**: List command fast, fetch command loads full content only when needed

## Development Approach

**Key Technical Components Remaining:**
1. ✅ Command parsing and validation system
2. ✅ TOML file operations and directory structure management  
3. ✅ HTTP request execution engine (preset `call` command)
4. ✅ Variable substitution system during request execution
5. ✅ TOML-to-JSON conversion with variable resolution
6. ⏳ Interactive command mode with state management

**Architecture Principles:**
- Single binary distribution (Go's strength)
- File-based configuration using TOML for human readability
- Clean separation between CLI parsing, file operations, and HTTP execution
- Intelligent type detection without verbose declarations

**Target User Experience:**
- Intuitive commands with universal space-separated bulk operations:
  - **Bulk removal**: `saul rm preset1 preset2 preset3` (space-separated)
  - **Special syntax**: `saul pokeapi set url https://api.com` (no = sign)
  - **Single key-value**: `saul pokeapi set body pokemon.name={?}` (single field)
  - **⏳ Planned migration**: `saul pokeapi set header Auth=token Accept=json` (space-separated)
  - **Current**: `saul pokeapi set header Auth=token,Accept=json` (comma-separated)
- Clean configuration files that are manually editable
- Smart prompting for variable values during execution
- Both scriptable and interactive usage patterns
- Universal space-separated pattern for all bulk operations (Unix consistency)

**Command Structure (Current & Planned):** 
- **Bulk removal**: `saul rm [preset1] [preset2] [preset3]` ✅ **IMPLEMENTED**
- **Special syntax**: `saul [preset] set url/method/timeout [value]` ✅ **IMPLEMENTED** 
- **Single field**: `saul [preset] set [target] [field=value]` ✅ **IMPLEMENTED**
- **⏳ Planned**: `saul [preset] set [target] [field1=value1] [field2=value2]` (space-separated)
- **Current**: `saul [preset] set [target] [field1=value1,field2=value2]` (comma-separated)

**Examples (Current & Planned):**
```bash
# ✅ IMPLEMENTED: Bulk removal with spaces
saul rm preset1 preset2 preset3

# ✅ IMPLEMENTED: Special syntax and single fields  
saul pokeapi set url https://api.com
saul pokeapi set method POST
saul pokeapi set header Authorization=Bearer123

# ⏳ PLANNED: Universal space-separated pattern
saul pokeapi set header Auth=token Accept=json
saul pokeapi set body pokemon.name={@pokename} pokemon.level=25

# 📝 CURRENT: Comma-separated (to be migrated)
saul pokeapi set header Auth=token,Accept=json,Type="app/json,utf-8"
saul pokeapi set body pokemon.name={@pokename},pokemon.level=25
```

## Testing

**Comprehensive Test Suite**: `other/testing/test_suite.sh`
- Phase-organized testing structure that expands with each implementation phase
- Validates all implemented functionality end-to-end from first step
- **Critical Development Practice**: ALWAYS add new features to test suite immediately upon implementation
- Automated setup and cleanup with clear pass/fail reporting
- Prevents regressions and ensures complete feature coverage

**Testing Philosophy**: The test suite is the single source of truth for feature validation. Every new capability must be added to the corresponding phase section in test_suite.sh to maintain comprehensive coverage.

**Current Status**:
- ✅ Phase 1 & 2: Fully tested and validated
- ✅ Phase 3: HTTP execution engine complete with comprehensive testing
- ✅ Phase 3.5: Critical architecture fix (TOML merging + variable syntax) - **COMPLETED**
  - Separate handler implementation eliminates field misclassification
  - Braced variable syntax prevents URL conflicts
  - Test suite refactored with reliable automation
- ⏳ Phase 4: Response history system - ready for implementation

## Important Notes

- **Phase 1, 2, 3 & 3.5 Complete**: Solid foundation with HTTP execution engine and architecture fixes
- **Core Functionality Ready**: Variable system, TOML operations, and HTTP execution fully implemented
- **Architecture Fixed**: Separate handlers eliminate field misclassification, braced variables prevent URL conflicts
- Focus on incremental development with full understanding of each component
- Prioritize clean, readable code over complex features
- Always validate against the README.md requirements during development
- Use `other/testing/test_suite_fixed.sh` for reliable automated testing

## Phase 3, 3.5, 4A & 4B Implementation Summary

**✅ HTTP Execution Engine Complete (Phase 3):**
- `saul preset call` command fully functional
- Variable prompting system with `{@}` hard variables, `{?}` soft variables
- HTTP client integration using go-resty
- Support for all major HTTP methods (GET, POST, PUT, DELETE, etc.)
- JSON body conversion and pretty-printed responses
- Comprehensive error handling and validation
- Smart Variable Deduplication feature documented and working

**✅ Architecture & Variable Syntax Fixes Complete (Phase 3.5):**
- ✅ Separate handler implementation eliminates field misclassification
- ✅ Braced variable syntax `{@name}` and `{?name}` prevents URL conflicts
- ✅ Real-world APIs work correctly: `https://api.github.com/@username`
- ✅ Complex URLs supported: `https://api.com/{@user}/posts?search=@mentions&token={@auth}`
- ✅ All existing functionality preserved with new syntax

**✅ Edit Command System Complete (Phase 4A):**
- ✅ Field-level editing with pre-filled readline prompts
- ✅ Interactive terminal editing experience with cursor movement and backspace
- ✅ Uses existing validation and TOML patterns - zero new complexity
- ✅ Commands: `saul api edit url`, `saul api edit body pokemon.name`
- ✅ Zero regression - purely additive feature

**✅ Response Formatting System Complete (Phase 4B):**
- ✅ Smart JSON→TOML conversion for dramatically improved readability
- ✅ Intelligent content-type detection with graceful fallback to raw display
- ✅ HTTP subfolder refactoring: `client.go`, `display.go`, `request.go`
- ✅ Real-world tested with JSONPlaceholder, PokéAPI, HTTPBin, GitHub APIs
- ✅ Response metadata headers show status, timing, size, content-type
- ✅ All existing functionality preserved with enhanced output formatting

**✅ Raw-First Display Architecture (Phase 4C):**
- ✅ File operations (`check`, `get`) output raw content for Unix composition
- ✅ HTTP responses retain pretty formatting with metadata (status, timing, size)
- ✅ Clean header separator without footer clutter
- ✅ Perfect for piping: `saul api check body | grep pokemon`
- ✅ Maintains scriptability while keeping HTTP responses readable

**Architecture Improvements:**
- Clean file separation: commands.go, variables.go, validation.go, http.go
- Separate TOML handlers for each file type (no merging conflicts)
- Robust test isolation with backup/restore functionality
- Reliable testing using JSONPlaceholder API and refactored test suite
- All tests passing with comprehensive coverage