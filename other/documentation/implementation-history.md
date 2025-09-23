# Better-Curl (Saul) - Implementation History

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