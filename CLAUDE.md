# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Vision and Approach

This project is **Better-Curl (Saul)** - a workspace-based HTTP client designed to eliminate the pain of complex curl commands with JSON payloads. The complete project specification, command structure, and user experience goals are documented in `other/documentation/vision.md` - **always reference this file for implementation details and requirements validation**.

**Collaborative Development Philosophy:**
- This is a learning-focused project where the user wants to understand every piece of code generated
- Follow KISS principles: clean, intelligent, self-maintained, resilient code above all else
- Avoid over-engineering at all costs - prioritize simple and clean solutions
- Always engage in strategic discussion before implementation
- Break down complex tasks into understandable components
- Explain architectural decisions and reasoning during development
- User wants to learn Go through AI-assisted development while maintaining deep code understanding

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

**Project Structure:**
```
better-curl-saul/
├── go.mod                        # Go module (module name: "main")
├── other/documentation/vision.md # Complete project specification
├── other/documentation/action-plan.md # Development action plan
├── cmd/
│   └── main.go                  # Clean entry point - program flow only
├── src/project/
│   ├── parser/
│   │   └── command.go           # Command struct + ParseCommand function
│   └── config/
│       └── constants.go         # Constants and command aliases
```

**Core Architecture Concepts (from other/documentation/vision.md):**
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
  - `saul call preset` command fully functional
  - Variable prompting and substitution system
  - HTTP client integration using go-resty
  - Support for all major HTTP methods
  - JSON body conversion and pretty-printed responses
- ✅ **Phase 3.5 Complete**: Architecture & Variable Syntax Fix
  - Separate handler implementation (no field misclassification)
  - Braced variable syntax `{@name}` and `{?name}` (no URL conflicts)
  - Real-world URL support: `https://api.github.com/@username` works correctly
  - All existing functionality preserved with new syntax
- ⏳ **Next**: Phase 4 - Response History System

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

**Variable Substitution**: Variables stored in variables.toml (hard only), resolved during `call` command

## Development Approach

**Key Technical Components Remaining:**
1. ✅ Command parsing and validation system
2. ✅ TOML file operations and directory structure management  
3. ✅ HTTP request execution engine (`call` command)
4. ✅ Variable substitution system during request execution
5. ✅ TOML-to-JSON conversion with variable resolution
6. ⏳ Interactive command mode with state management

**Architecture Principles:**
- Single binary distribution (Go's strength)
- File-based configuration using TOML for human readability
- Clean separation between CLI parsing, file operations, and HTTP execution
- Intelligent type detection without verbose declarations

**Target User Experience:**
- Intuitive commands with dual syntax:
  - Special: `saul pokeapi set url https://api.com` (no = sign)
  - Regular: `saul pokeapi set body pokemon.name={?}` (with = sign)
- Clean configuration files that are manually editable
- Smart prompting for variable values during execution
- Both scriptable and interactive usage patterns

**Command Structure:** 
- Special: `saul [preset] set url/method/timeout [value]`
- Regular: `saul [preset] set [target] [field=value]`
- Examples:
  - `saul pokeapi set url https://api.com`
  - `saul pokeapi set method POST`
  - `saul pokeapi set header Authorization=Bearer123`
  - `saul pokeapi set body pokemon.name={@pokename}`
  - `saul pokeapi call`

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
- Always validate against the other/documentation/vision.md requirements during development
- Use `other/testing/test_suite_fixed.sh` for reliable automated testing

## Phase 3 & 3.5 Implementation Summary

**✅ HTTP Execution Engine Complete (Phase 3):**
- `saul call preset` command fully functional
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

**Architecture Improvements:**
- Clean file separation: commands.go, variables.go, validation.go, http.go
- Separate TOML handlers for each file type (no merging conflicts)
- Robust test isolation with backup/restore functionality
- Reliable testing using JSONPlaceholder API and refactored test suite
- All tests passing with comprehensive coverage