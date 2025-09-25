# Implementation History

<!-- Completed phases moved from implementation-plan.md in chronological order.
Use this to understand the evolution and building blocks of this project. -->

## Architecture Evolution Summary

This document archives the technical evolution of Better-Curl (Saul) from initial infrastructure to a fully functional HTTP client. All phases listed here are **COMPLETED** and serve as context for future development.

## Major Architectural Decisions

### Phase 0: Infrastructure Foundation (2025-09-22)
**Why**: Eliminated global state violations and established proper Go conventions
- **Global State Elimination**: Replaced `var currentPreset string` with SessionManager pattern using dependency injection
- **Module Import Cleanup**: Fixed go.mod references to match repository structure
- **File Operations**: Implemented atomic writing operations to prevent corruption during concurrent access

### Phase 1-2: TOML-First Architecture
**Why**: Human-readable configuration files enable both manual editing and programmatic manipulation
- **5-File Unix Philosophy**: Separate concerns into body.toml, headers.toml, query.toml, request.toml, variables.toml
- **Directory Structure**: `~/.config/saul/presets/[preset-name]/` workspace pattern
- **Special Request Syntax**: `set url/method/timeout [value]` (no = sign) vs regular `set target field=value`

### Phase 3: HTTP Execution Engine
**Why**: Variable system enables reusable configurations across different environments
- **Variable Syntax Evolution**: `@name` → `{@name}` (hard variables), `{?name}` (soft variables) to prevent URL conflicts
- **Separate Handler Pattern**: Individual TOML handlers per file type eliminates field misclassification bugs
- **Variable Detection Simplification**: Replaced complex TOML parsing with regex detection (~100 lines → ~20 lines)

### Phase 4A-B: User Experience Enhancement
**Why**: Interactive editing and readable responses are essential for API development workflow
- **Edit Command System**: Field-level editing with pre-filled readline prompts using existing TOML patterns
- **Response Formatting**: Smart JSON→TOML conversion with intelligent content-type detection
- **HTTP Subfolder Refactoring**: Clean separation into client.go, response.go for maintainability

### Phase 4B-Post: Syntax Optimization
**Why**: Reduce command verbosity while maintaining Unix composability
- **Space-Separated Pattern**: Universal approach for all bulk operations (`saul rm preset1 preset2 preset3`)
- **Key-Value Migration**: `saul api set body name=val1 type=val2` (space-separated) eliminates complex parsing
- **Perfect Backward Compatibility**: Zero regression with cleaner, more intuitive syntax

### Phase 4C-E: Production Features
**Why**: Handle real-world API complexity and debugging workflows
- **Response Filtering**: Field selection for large APIs using existing KeyValuePairs system
- **Terminal Session Memory**: TTY-based current preset tracking with automatic cleanup
- **Response History**: Split command architecture (`check history` + `check response N`) with metadata storage

### Phase 5-6: Polish & Standards
**Why**: Production-ready distribution with extensible architecture
- **Universal Flag System**: `--raw` flag with extensible foundation for future flags
- **System Command Delegation**: Whitelist-based security for `saul ls`, `saul tree` etc.
- **Package Documentation**: Go-style documentation for all 8 core packages

### Phase 6A-Post: Command Naming Refinement (2025-09-24)
**Why**: Improve semantic clarity and align with industry standards
- **`check` → `get` Rename**: Changed primary display command to match standard CLI conventions (kubectl get, git config --get)
- **Dead Code Removal**: Eliminated incomplete `get.go` (debugging function never documented/tested)
- **Zero Breaking Changes**: Original `get` command was internal-only, never exposed to users
- **Improved UX**: `saul get url` more intuitive than `saul check url` for retrieval operations

### Phase 5B.1: Get Command Field-Specific Behavior (2025-09-25)
**Why**: Improve Unix composability by returning raw values for specific field queries
- **Field-Specific Returns**: `saul api get body pokemon.name` returns just "pikachu" instead of entire TOML file
- **Raw Value Output**: Individual field queries output raw values using `fmt.Println(value)` (lines 63-85 in get.go)
- **Preserved Functionality**: Full file display unchanged when no specific field requested
- **Unix Philosophy**: Enhanced command composition and piping capabilities

### Phase 6.1: File Editing Integration (2025-09-25)
**Why**: Direct TOML file editing eliminates intermediate command sequences
- **Container-Level Editing**: `saul api edit body` opens body.toml directly in $EDITOR
- **Editor Detection**: Automatic detection of $EDITOR with fallback to nano/vim/vi/emacs (edit.go:140-155)
- **File Creation**: Automatic creation of empty TOML files if they don't exist
- **Dual Edit Modes**: Field-level editing (existing functionality) vs container-level editing (new)

### Phase 6.3: Production Distribution Readiness (2025-09-25)
**Why**: Enable wide distribution and easy installation across platforms
- **Cross-Platform Install Script**: Auto-detects OS/architecture and downloads appropriate binaries
- **GitHub Release Automation**: Automated release workflows with cross-platform binary builds
- **Fallback Build System**: Local source build when binaries unavailable
- **Installation Pipeline**: Complete end-to-end installation from curl one-liner to working binary

### Phase 7A: Response Field Extraction Feature (2025-09-25)
**Why**: Enable granular inspection of stored HTTP response history for debugging and analysis workflows
- **Objective**: Implement field extraction from HTTP response history (`saul get response1 body`, `saul get response headers`)
- **Strategic Decision**: Response Field Extraction chosen over Flag System (90% existing infrastructure reuse, minimal risk, quick user value)
- **Single-Line Format**: Standardized on compact format (`response1`, `response2`) eliminating dual-format confusion
- **Exact Filtering Logic**: Uses identical data pipeline as live API calls - stored response body converted to bytes then filtered/formatted
- **Architecture Improvements**: Removed space-separated format support, unified parsing logic with number-first fallback-to-field detection
- **Key Features Delivered**:
  - `saul get response1 body` - Extract body from specific response with filtering applied
  - `saul get response headers` - Extract headers from most recent response (no filtering)
  - `saul get response1` - Show whole response (single-line support)
  - `saul get response status/url/method/duration` - Simple field extraction for metadata
- **Critical Fix**: Body filtering now uses exact same `http.FormatResponseContent()` as live API (stored string → bytes → applyFiltering → TOML)
- **Zero Breaking Changes**: All existing functionality preserved, error messages remain format-agnostic

### Phase 7A: Variable Prompting UX Enhancement (2025-09-25)
**Why**: Eliminate user frustration from having to retype entire variable values during editing
- **Problem Identified**: Variable prompting used `bufio.Scanner` forcing users to retype complete values instead of editing existing ones
- **Solution**: Upgraded to `readline` library with pre-filled prompt functionality
- **Technical Implementation**:
  - **File**: `src/project/handlers/variables/prompting.go`
  - **Pattern Reuse**: Leveraged exact same logic as edit command (`github.com/chzyer/readline` + `WriteStdin()`)
  - **Behavior Preservation**: Hard variables (`{@name}`) pre-fill with stored values, soft variables (`{?name}`) remain empty
  - **Error Handling**: Maintained existing error patterns using display messages
- **User Experience Improvement**:
  - **Before**: `token [abc123]: _` (empty input, must retype everything)
  - **After**: `token: abc123_` (can edit existing value directly)
- **Zero Breaking Changes**: All variable prompting behavior preserved, only improved editing experience
- **Implementation Quality**: Clean import changes, removed unused `bufio` and `os` imports, proper readline lifecycle management

### Phase 8: Advanced Flag System (2025-09-25)
**Why**: Complete HTTP client flag ecosystem for advanced workflow optimization
- **Strategic Decision**: Full flag infrastructure implementation providing comprehensive request manipulation and response filtering capabilities
- **Phase 8A - Flag Infrastructure**: Extended Command struct with new flag fields (VariableFlags, ResponseFormat, DryRun) and comprehensive flag parser supporting both long (`--dry-run`, `--headers-only`) and short (`-v`) flags
- **Phase 8B - Dry-Run Feature**: Added request preview functionality showing complete HTTP request details without execution for workflow validation
- **Phase 8C - Variable Management**: Implemented selective variable prompting with `-v` flag supporting both specific variable lists (`-v token username`) and all-variables mode (`-v`)
- **Phase 8D - Response Formatting**: Complete response format system with `--headers-only`, `--body-only`, `--status-only` flags providing targeted output for scripting and debugging
- **Critical Enhancement**: Fixed `--raw` consistency ensuring truly raw output (no filtering, no pretty printing) across all response formats
- **Technical Implementation**:
  - **File**: `src/project/core/parser.go` - Extended flag parsing with robust error handling
  - **File**: `src/project/handlers/http.go` - Integrated dry-run logic and variable flag handling
  - **File**: `src/project/handlers/http/response.go` - Enhanced response formatting with consistent raw mode behavior
  - **File**: `src/project/handlers/variables/prompting.go` - Added `PromptForSpecificVariables()` function
  - **File**: `src/project/handlers/variables.go` - Exported new function for handler integration
- **Key Features Delivered**:
  - `saul call --dry-run` - Request preview without execution
  - `saul call -v token` - Prompt for specific variables only
  - `saul call -v` - Prompt for all variables
  - `saul call --body-only` - Filtered body with TOML formatting
  - `saul call --body-only --raw` - Unfiltered body with raw JSON
  - `saul call --headers-only` - Raw HTTP headers
  - `saul call --status-only` - Status code only
- **Consistency Achievement**: `--raw` now means "completely raw" across all output modes - no filtering, no pretty printing, perfect for Unix toolchain integration and scripting workflows
- **Zero Breaking Changes**: All existing functionality preserved while adding comprehensive flag ecosystem

## Key Technical Patterns Established

### TOML Manipulation Engine
- **Location**: `src/project/toml/handler.go`
- **Pattern**: Dot notation operations (`.Set()`, `.Get()`, `.ToJSON()`)
- **Integration**: Repurposed from toml-cli project for HTTP client needs

### Variable Resolution System
- **Hard Variables** (`{@name}`): Stored in variables.toml, persist across sessions
- **Soft Variables** (`{?name}`): Always prompt, never stored
- **Substitution**: During `call` command before HTTP execution

### File Structure Evolution
```
Phase 1: Single handlers, merged files
Phase 3.5: Separate handlers per file type (eliminates misclassification)
Phase 4B: HTTP subfolder (client.go, response.go)
Phase 4B-Post: Command subfolder (set.go, check.go, edit.go)
```

### Command Architecture
```
Entry Point (main.go) → Parser (core/parser.go) → Router → Handler Functions
- Global Commands: rm, ls (system delegation)
- Preset Commands: set, check, get, edit, call
- Special Syntax: Request fields (url, method, timeout) use no = sign
- Universal Pattern: Space-separated arguments for all bulk operations
```

## Critical Lessons Learned

### Architecture Decisions
1. **Separate TOML Handlers**: Prevents field misclassification between different file types
2. **Braced Variable Syntax**: `{@name}` prevents conflicts with literal @ symbols in URLs
3. **Regex-Based Detection**: Much more reliable than recursive TOML structure parsing
4. **Space-Separated Universality**: Consistent pattern across all bulk operations reduces cognitive load

### Implementation Approach
1. **Test-First**: Each phase added comprehensive test coverage before moving forward
2. **Zero Regression**: All new features preserved existing functionality completely
3. **KISS Principles**: Simple solutions preferred over complex architectures
4. **Unix Philosophy**: Small, composable functions with clear single responsibilities

### User Experience Insights
1. **Terminal Session Memory**: Automatically tracking current preset eliminates repetitive preset specification
2. **Split Command History**: `list then select` pattern more discoverable than complex single commands
3. **Response Filtering**: Essential for large APIs, pure UNIX design using existing systems
4. **Raw Mode Philosophy**: File operations output raw content, HTTP responses get pretty formatting

## Package Structure (Final)
```
src/
├── modules/
│   ├── errors/messages.go        # Centralized error constants
│   └── display/                  # Universal printing system
│       ├── printer.go           # Error, Success, Warning, Info, Plain functions
│       └── sections.go          # Section formatting
└── project/
    ├── core/                    # CLI fundamentals (parser, session, delegation)
    ├── handlers/                # Command & HTTP handling logic
    │   ├── commands/           # set.go, get.go, check.go, edit.go
    │   ├── variables/          # Variable detection, prompting, storage
    │   └── http/               # HTTP client execution, response formatting
    ├── presets/                # Workspace management
    ├── toml/                   # TOML manipulation engine
    └── config/                 # Configuration constants
```

## Completion Timeline
- **Phase 0** (Infrastructure): September 22, 2025
- **Phase 1-2** (Foundation): Earlier completion
- **Phase 3** (HTTP Engine): Earlier completion
- **Phase 4A** (Edit Commands): User experience enhancement
- **Phase 4B** (Response Formatting): API development workflow improvement
- **Phase 4B-Post** (Syntax Optimization): Command efficiency improvement
- **Phase 4C** (Response Filtering): Large API handling
- **Phase 4D** (Session Memory): Workflow optimization
- **Phase 4E** (Response History): Debugging workflow
- **Phase 5A** (Flag System): Production polish
- **Phase 6A** (System Commands): Unix integration
- **Phase 3.5** (Architecture Fix): Critical debugging phase
All phases achieved their success criteria with zero regression and established a solid foundation for future enhancements.