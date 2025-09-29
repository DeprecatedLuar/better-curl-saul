# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Better-Curl (Saul) is a feature-complete workspace-based HTTP client that eliminates complex curl commands with JSON payloads. The project is in maintenance mode, having completed all core functionality (Phases 0-6A).

**Core Purpose**: Transform complex curl commands into simple, organized workspace configurations with smart variable systems.

## Common Development Commands

### Building and Running
```bash
# Build the binary
go build -o saul cmd/main.go

# Run directly from source
go run cmd/main.go [command]

# Run with test preset
go run cmd/main.go pokeapi
go run cmd/main.go pokeapi set body pokemon.name=pikachu
go run cmd/main.go pokeapi call
```

### Testing
```bash
# Run integration tests
cd src/project && go test -v

# Test installation script locally
./install.sh
```

### Development Workflow
```bash
# Create preset and test HTTP functionality
./saul demo set url https://jsonplaceholder.typicode.com/posts/1
./saul demo call

# Build for distribution
go build -ldflags "-s -w" -o saul cmd/main.go
```

## Code Architecture

### High-Level Structure
```
src/
├── modules/          # Reusable infrastructure (framework-level)
│   ├── display/      # Output formatting and user messages
│   └── logging/      # Logging utilities
├── project/          # Application-specific business logic
│   ├── commands/     # Command implementations (set, get, edit, history)
│   ├── config/       # Configuration constants and paths
│   ├── core/         # Command parsing, curl parsing, and session management
│   ├── http/         # HTTP client, execution, and response handling
│   ├── utils/        # Project utilities, types, and version management
│   ├── variables/    # Variable detection, prompting, and storage
│   └── workspace/    # Preset management and TOML file operations
└── settings/         # Global settings configuration
```

### Architecture Flow
```
User Input → core.ParseCommand() → commands/ → workspace/ TOML operations → http/ execution
```

### Key Design Patterns

**5-File TOML Structure**: Each preset workspace contains:
- `body.toml` - HTTP request body (JSON)
- `headers.toml` - HTTP headers
- `query.toml` - Query/search parameters
- `request.toml` - HTTP method, URL, timeout
- `variables.toml` - Hard variables storage

**Variable System**:
- `{@name}` - Hard variables (stored, prompt once)
- `{?name}` - Soft variables (prompt every time)

**Special Request Syntax**:
- `set url/method/timeout` (no = sign for request fields)
- Regular TOML syntax with = sign for structured data

**Session Memory**: Terminal-scoped preset persistence via TTY-based sessions in `core.SessionManager`

### Module Separation
- **modules/**: Framework-level, reusable across projects
- **project/**: Application-specific business logic
- All source files must follow 250-line limits and single responsibility principle

### Command Processing
1. **cmd/main.go**: Entry point, session initialization, command injection for current presets
2. **core.ParseCommand()**: Command parsing and validation
3. **commands/**: Command execution (set, get, edit, history)
4. **workspace/**: TOML file operations for configuration persistence

### HTTP Execution
- Built on `go-resty/resty/v2` for HTTP client functionality
- Response formatting and filtering via `modules/display`
- Variable substitution before request execution
- Response history management in preset directories

## Project-Specific Conventions

### File Organization
- Check `.purpose.md` files in directories for context
- `implementation-plan.md` contains implementation documentation,focus on value avoid verbosity
- `implementation-history.md` contains concluded phases
- `README-draft.md` is the source of truth for project vision

### Go Code Standards
- 250-line file limit enforced
- Single responsibility principle for functions and modules
- Clean separation between framework (modules/) and application (project/)
- No external testing framework - uses Go's built-in testing

### TOML Configuration
- Dot notation for nested objects: `obj.field=value`
- Space-separated bulk operations: `saul rm preset1 preset2 preset3`
- Variable substitution during HTTP execution, not storage

### Command Patterns
```bash
# Preset creation/switching
saul [preset]

# Configuration commands
saul [preset] set [target] [key=value]
saul [preset] get [target] [key]  # Raw value output
saul [preset] check [target]     # Formatted display

# HTTP execution
saul [preset] call
saul call  # Uses current preset from session
```

## Dependencies and Requirements

**Core Dependencies**:
- `github.com/go-resty/resty/v2` - HTTP client
- `github.com/pelletier/go-toml` - TOML parsing
- `github.com/chzyer/readline` - Interactive input
- `github.com/tidwall/gjson` - JSON response filtering
- `golang.org/x/term` - Terminal utilities

**Go Version**: 1.24.6+

## Current Project Status

**Phase**: Feature-complete, maintenance mode
**Focus**: Code quality improvements and bug fixes
**Testing**: Integration tests in `src/project/integration_test.go`

### Development Guidelines
- **KISS Principles**: Simple, clean solutions over complex architectures
- **Zero Regression**: All changes must preserve existing functionality
- **Unix Philosophy**: Small, composable functions with clear single responsibilities
- **Backward Compatibility**: Maintain compatibility with existing configurations

## Installation and Distribution

**Installation Script**: `install.sh` provides one-liner installation with auto-detection
**Build Process**: Standard Go build produces single binary
**Target Platforms**: Linux, macOS, Windows

The project includes version management utilities in `project/utils/` and supports self-updating via the `saul update` command.