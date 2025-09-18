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
- **TOML Structure**: Separate files for headers.toml, body.toml, query.toml, config.toml
- **Variable System**: Soft variables (`?name`) always prompt, hard variables (`$name`) persist in config
- **Command Modes**: Both interactive mode and single-line commands
- **Data Flow**: TOML → JSON conversion → HTTP execution

**Current Implementation Status:**
- ✅ **Modular Go structure** following Go conventions
- ✅ **Command parsing system** handles both global and preset commands
- ✅ **Clean separation of concerns** (main.go, parser, config)
- ✅ **Dual command support** (global commands like `rm`, `list` + preset commands)
- ⏳ **Next**: TOML file operations and directory structure management

## TOML Manipulation System

**Core Library**: Repurposed TomlHandler from toml-cli project
- **Location**: `src/project/toml/handler.go`
- **Purpose**: Dot notation TOML manipulation for Saul commands
- **Key methods**: `.Set()`, `.Get()`, `.ToJSON()` for HTTP conversion

**Integration Pattern:**
- Command: `saul pokeapi set body pokemon.stats.hp=100`
- Flow: Parse command → TomlHandler.Set("pokemon.stats.hp", 100) → Write to body.toml

**Variable Substitution**: Integrate with existing tomv library after TOML operations

## Development Approach

**Key Technical Components to Implement:**
1. Command parsing and validation system
2. TOML file operations and directory structure management
3. Variable substitution system (soft/hard variables)
4. TOML-to-JSON conversion with dot notation support
5. HTTP request execution engine
6. Interactive command mode with state management

**Architecture Principles:**
- Single binary distribution (Go's strength)
- File-based configuration using TOML for human readability
- Clean separation between CLI parsing, file operations, and HTTP execution
- Intelligent type detection without verbose declarations

**Target User Experience:**
- Intuitive commands: `saul pokeapi set body pokemon.name=?` (no "preset" keyword)
- Clean configuration files that are manually editable
- Smart prompting for variable values during execution
- Both scriptable and interactive usage patterns

**Command Structure:** `saul [preset] [command] [target] [field=value]`
- Example: `saul pokeapi set header Authorization=Bearer123`
- Example: `saul pokeapi set body pokemon.name=?`
- Example: `saul pokeapi fire`

## Important Notes

- The codebase is in early development - most core functionality needs implementation
- Focus on incremental development with full understanding of each component
- Prioritize clean, readable code over complex features
- Always validate against the vision.md requirements during development