# Better-Curl (Saul) - AI Development Context

## Project Context
- Current Phase: Feature-complete (Phases 0-6A), focus on code quality improvements
- Framework: `modules/` = reusable infrastructure, `project/` = app-specific business logic
- Purpose: Workspace-based HTTP client eliminating complex curl commands with JSON payloads

## Key Commands
```bash
go run cmd/main.go [command]          # Run from project root
go run cmd/main.go pokeapi            # Test with preset
go run cmd/main.go pokeapi set body pokemon.name=pikachu
go run cmd/main.go pokeapi call       # Execute HTTP request
```

## Project-Specific Patterns
- **5-File TOML Structure**: body.toml, headers.toml, query.toml, request.toml, variables.toml
- **Variable System**: `{@name}` (hard/stored), `{?name}` (soft/prompt)
- **Special Request Syntax**: `set url/method/timeout` (no = sign)
- **Space-Separated Bulk**: `saul rm preset1 preset2 preset3`
- **Session Memory**: Terminal-scoped preset persistence via TTY-based sessions

## Architecture Flow
```
User Input → core.ParseCommand() → handlers/commands/ → toml/ operations → HTTP execution
```

## Context Discovery
Check `.purpose.md` files in folders for framework context before working. README.md is source of truth for project vision and requirements.

## Testing