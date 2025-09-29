# Implementation Plan

<!-- WORKFLOW IMPLEMENTATION GUIDE:
- This file contains active phases for implementation (completed phases moved to implementation-history.md)
- Each phase = one focused session until stable git commit, follow top-to-bottom order
- SPLIT complex phases into subphases for safety, testing, and incremental value
- Avoid verbose explanations - just implement what's specified and valuable
- Focus on actionable steps: "Update file X, add function Y"
- Success criteria must be testable
- Test after each (sub)phase completion
-->

# Current Status

Better-Curl (Saul) has achieved feature-complete status with comprehensive HTTP client functionality including:

- ✅ **Core HTTP Client**: Full request/response handling with variable substitution
- ✅ **Workspace Management**: TOML-based preset system with 5-file structure
- ✅ **Advanced Editing**: Field-level and container-level editing capabilities
- ✅ **Response Processing**: Filtering, formatting, and history management
- ✅ **Flag System**: Complete flag ecosystem (--raw, --dry-run, -v, --body-only, etc.)
- ✅ **Unix Integration**: System command delegation and composable output

## Future Development

New features should be implemented as new phases in this file following the established patterns:
- Single-focused sessions
- Testable success criteria
- Zero-regression requirement
- Documentation in implementation-history.md upon completion

---

# Phase 7: Curl Import Feature

## 7.1 - Curl Parser Implementation ✅
**Goal**: Build custom curl parser and validate it works

**Completed Tasks**:
1. Built custom parser in `src/project/core/curl_parser.go` (no external dependency)
2. Created comprehensive test suite in `src/project/core/curl_parser_test.go`
3. Validated parser with real-world examples (Instantly.ai API)
4. Refactored project structure:
   - Renamed `core/parser.go` → `core/command_parser.go`
   - Renamed `core/delegation.go` → `core/command_delegation.go`
   - Placed curl parser in `core/` (input parsing layer)

**Parser Features**:
- Extracts: method, URL, headers, body
- Separates query params from URL into structured map
- Handles multiline JSON bodies
- Supports multiple header flags
- Works with both quoted and unquoted URLs

**Success Criteria**: ✅
- All tests pass with complex real-world curl commands
- Query parameters correctly extracted and separated
- Parser handles edge cases (multiline bodies, multiple headers)

**Commit**: "feat: implement custom curl parser and refactor core structure"

---

## 7.2 - Core Import Function ✅
**Goal**: Create curl → TOML conversion logic (no editor yet)

**Completed Tasks**:
1. Created file: `src/project/workspace/curl_import.go` (placed in workspace/ for better architectural fit)
2. Implemented `ImportCurlString(preset string, curlCmd string) error`:
   - Parse curl using `core.ParseCurl()`
   - Validate URL exists
   - Convert body: `workspace.NewTomlHandlerFromJSON([]byte(result.Body))`
   - Convert headers: iterate `result.Headers`, use `Set(key, val)`
   - Convert query params: iterate `result.Query`, write to query.toml
   - Convert request: use `result.BaseURL` (not full URL with query string)
   - Save all files: `workspace.SavePresetFile(preset, target, handler)`
   - Auto-create preset directory if it doesn't exist
3. Added integration tests in `src/project/integration_test.go`:
   - Test: Basic POST with JSON body
   - Test: GET with query params and headers
   - All assertions validate correct TOML file creation

**Success Criteria**: ✅
- Test creates correct TOML files (body, headers, query, request)
- Body JSON → TOML conversion works
- Headers map → TOML works
- Query params extracted to separate file
- Request uses BaseURL (query params removed)

**Commit**: "feat: implement curl string to TOML conversion"

---

## 7.3 - Editor Integration
**Goal**: Add temp file + editor workflow

**Tasks**:
1. In `curl_import.go`, add `ImportCurlViaEditor(preset string) error`:
   - Create temp file: `os.CreateTemp("", fmt.Sprintf("saul-%s-*.txt", preset))`
   - Check `$EDITOR` env var, fallback to `nano`
   - Open editor: `exec.Command(editor, tempFile.Name()).Run()`
   - Read file: `os.ReadFile(tempFile.Name())`
   - Call `ImportCurlString(preset, content)`
   - Clean up: `defer os.Remove(tempFile.Name())`
2. Add error handling for:
   - Empty file
   - Invalid curl syntax
   - Editor not found

**Success Criteria**:
- Manual test: `saul test set --raw` opens editor
- Pasting valid curl → creates TOML files
- Invalid curl → shows error message
- Temp file cleaned up after

**Commit**: "feat: add editor workflow for curl import"

---

## 7.4 - Command Integration
**Goal**: Wire up `saul set --raw` command

**Tasks**:
1. Modify `src/project/handlers/commands/set.go`:
```go
func Set(cmd core.Command) error {
    if cmd.RawOutput {
        return ImportCurlViaEditor(cmd.Preset)
    }
    // ... existing logic
}
```
2. Verify flag parsing already handles `--raw` (it does in parser.go)

**Success Criteria**:
- `saul myapi set --raw` opens editor
- Complete workflow works end-to-end
- All TOML files created correctly

**Commit**: "feat: wire up saul set --raw command"

---

## 7.5 - Edge Cases & Polish
**Goal**: Handle real-world scenarios

**Tasks**:
1. Handle query parameters in URL (curl library should parse this)
2. Test with complex curl commands:
   - Multiple headers
   - Nested JSON body
   - URL with query params
3. Add helpful error messages
4. Update integration test suite with edge cases

**Success Criteria**:
- Works with real API examples (GitHub, Stripe, etc.)
- Error messages are clear and actionable
- No crashes on invalid input

**Commit**: "feat: handle curl import edge cases"

---

## Phase 7 Testing Checklist
Run after each subphase:
```bash
# Unit tests
cd src/project && go test -v

# Manual test
go run cmd/main.go test set --raw
# Paste curl, verify TOML files created correctly

# Verify existing commands still work
go run cmd/main.go test set body name=value
go run cmd/main.go test call
```

**Future Enhancement** (not in Phase 7):
- `saul edit --raw` for reverse (TOML → curl display)
- Add to roadmap after Phase 7 complete

