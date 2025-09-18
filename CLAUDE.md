# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Vision and Approach

This project is **Better-Curl (Saul)** - a workspace-based HTTP client designed to eliminate the pain of complex curl commands with JSON payloads. The complete project specification, command structure, and user experience goals are documented in `other/vision.md` - **always reference this file for implementation details and requirements validation**.

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
├── other/vision.md              # Complete project specification
├── cmd/
│   └── main.go                  # Clean entry point - program flow only
├── src/project/
│   ├── parser/
│   │   └── command.go           # Command struct + ParseCommand function
│   └── config/
│       └── constants.go         # Constants and command aliases
```

**Core Architecture Concepts (from vision.md):**
- **Presets**: Folders in `~/.config/saul/presets/[preset-name]/` containing TOML files
- **5-File Structure**: headers.toml, body.toml, query.toml, request.toml, variables.toml (Unix philosophy)
- **Variable System**: Soft variables (`?name`) always prompt, hard variables (`@name`) persist in variables.toml
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
  - Variable system: `@` for hard variables, `?` for soft variables
  - Target normalization and validation
  - Comprehensive test suite validation
- ⏳ **Next**: Phase 3 - HTTP Execution Engine with `fire` command

## TOML Manipulation System

**Core Library**: Repurposed TomlHandler from toml-cli project
- **Location**: `src/project/toml/handler.go`
- **Purpose**: Dot notation TOML manipulation for Saul commands
- **Key methods**: `.Set()`, `.Get()`, `.ToJSON()` for HTTP conversion

**Integration Pattern:**
- Regular: `saul pokeapi set body pokemon.stats.hp=100`
- Special: `saul pokeapi set url https://api.com` (no = sign)
- Variables: `saul pokeapi set body name=@pokename` (hard) or `name=?` (soft)
- Flow: Parse command → TomlHandler.Set("pokemon.stats.hp", 100) → Write to appropriate .toml file

**Variable Substitution**: Variables stored in variables.toml (hard only), resolved during `fire` command

## Development Approach

**Key Technical Components Remaining:**
1. ✅ Command parsing and validation system
2. ✅ TOML file operations and directory structure management  
3. ⏳ HTTP request execution engine (`fire` command)
4. ⏳ Variable substitution system during request execution
5. ⏳ TOML-to-JSON conversion with variable resolution
6. ⏳ Interactive command mode with state management

**Architecture Principles:**
- Single binary distribution (Go's strength)
- File-based configuration using TOML for human readability
- Clean separation between CLI parsing, file operations, and HTTP execution
- Intelligent type detection without verbose declarations

**Target User Experience:**
- Intuitive commands with dual syntax:
  - Special: `saul pokeapi set url https://api.com` (no = sign)
  - Regular: `saul pokeapi set body pokemon.name=?` (with = sign)
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
  - `saul pokeapi set body pokemon.name=@pokename`
  - `saul pokeapi fire`

## Testing

**Comprehensive Test Suite**: `./test_suite.sh`
- Phase-organized testing structure
- Validates all implemented functionality
- Expandable as new phases are completed
- Automated setup and cleanup
- Clear pass/fail reporting

**Current Status**: Phases 1 & 2 fully tested and validated

## Important Notes

- **Phase 1 & 2 Complete**: Solid foundation with comprehensive testing
- Focus on incremental development with full understanding of each component  
- Prioritize clean, readable code over complex features
- Always validate against the vision.md requirements during development
- Use `./test_suite.sh` to validate all functionality before proceeding