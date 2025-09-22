# Whitelist Filtering Feature Specification (Updated)

## Overview

This document outlines the complete specification for implementing whitelist filtering in Better-Curl (Saul) to solve the problem of bloated API responses that overflow terminal displays, particularly from complex APIs like PokéAPI.

## Problem Statement

**Core Issue**: APIs like PokéAPI return massive JSON responses that completely fill the terminal, making it impossible to see the entire response or scroll back to the beginning. Users typically only care about 3-5 specific fields but get overwhelmed with 100+ fields that consume the entire terminal view.

**Target Use Case**: PokéAPI returns 100+ fields including sprites, abilities, moves, game versions, etc. User only wants basic stats like `name`, `stats[0]`, `stats[1]`, `types[0].type.name` that fit comfortably in their terminal.

## Solution: Whitelist Filtering System

### Design Philosophy
- **Whitelist-first approach**: Explicitly specify what to include (vs blacklist what to exclude)
- **Minimal verbosity**: Follow Saul's existing simple command patterns
- **Manual editability**: Human-readable TOML configuration files
- **Scriptable**: Single command to set up filters
- **Maintains Unix philosophy**: Dedicated filters.toml file for single responsibility
- **Terminal-friendly**: Keep responses readable within terminal bounds

### Command Syntax

Following Saul's established special syntax pattern (no `=` for special fields):

```bash
# Set whitelist filter (comma-separated fields)
saul pokeapi set filter name,stats[0],stats[1],types[0].type.name

# Edit filter interactively (opens filters.toml in editor)
saul pokeapi edit filter

# Check current filter settings (displays filters.toml content)
saul pokeapi check filter

# Clear filters (show all fields)
saul pokeapi set filter ""

# Apply filter during response (automatic)
saul pokeapi call
```

### Field Path Syntax (Industry Standard)

**Basic Field Access:**
```bash
name                    # Top-level field
stats                   # Entire stats object/array
```

**Nested Field Access:**
```bash
types[0]                # First item in types array
types[0].type.name      # Nested field in array item
pokemon.stats.hp        # Nested object fields
```

**Array Access:**
```bash
stats[0]                # First stat
stats[1]                # Second stat  
moves[5].move.name      # Sixth move's name
```

**Rationale for Standard Syntax:**
- **Industry standard**: Matches jq, JSONPath, JavaScript conventions
- **Unambiguous**: Clear distinction between field access (.) and array indexing ([0])
- **Familiar**: Every developer knows obj.field and array[index] patterns
- **Composable**: Natural chaining like types[0].type.name

### Storage Format

**File**: `~/.config/saul/presets/[preset-name]/filters.toml`

**Structure** (TOML array for readability):
```toml
fields = [
    "name",
    "stats[0]",
    "stats[1]", 
    "types[0].type.name"
]
```

**Rationale for TOML Array**:
- Easy manual editing (add/remove lines)
- Git-friendly diffs (one field change = one line)
- Copy-paste friendly for sharing configurations
- Supports comments for field documentation
- No escaping issues with field names

### Implementation Details

#### Command Parsing
- **Input**: `saul pokeapi set filter name,stats[0],stats[1]`
- **Parser**: Recognize "filter" as special field (like "url", "method")
- **Processing**: Split comma-separated string into array
- **Storage**: Convert to TOML array format

#### Filter Application
- **Trigger**: During `saul call` command execution
- **Timing**: After HTTP response received, before TOML conversion
- **Process**:
  1. Load filters.toml if it exists
  2. If no filter defined, proceed with full response
  3. If filter defined, extract only whitelisted fields from JSON
  4. Continue with TOML conversion on filtered data

#### Field Path Resolution
- **Library**: Use `github.com/tidwall/gjson` for robust JSON path extraction
- **Syntax**: Industry standard notation for all access patterns
- **Arrays**: Bracket notation (`stats[0]`, `moves[5]`)
- **Nested**: Chain notation (`types[0].type.name`)

#### Error Handling Strategy (Silent Ignoring)

**Philosophy**: Silent ignoring of missing/invalid paths for maximum robustness.

**Behavior Examples**:
```bash
# Filter specifies these fields
saul pokeapi set filter name,stats[0],stats[999],nonexistent.field

# API response only contains name and stats[0]
# Result: Returns only name and stats[0], silently ignores missing fields
# No errors, no warnings, no broken execution
```

**Benefits**:
- ✅ **Robust**: API changes don't break filtering
- ✅ **User-friendly**: No error spam in output
- ✅ **Practical**: Missing fields are simply not needed
- ✅ **Simple**: No complex validation required

**Implementation**: `gjson.Get()` returns empty values for missing paths without errors.

### Integration Points

#### File System Integration
- **Location**: Part of existing preset structure
- **File Creation**: Lazy creation (only when first filter set)
- **File Management**: Uses existing `LoadPresetFile`/`SavePresetFile` patterns

#### Command System Integration
- **Parser**: Add "filter" recognition to command parsing
- **Router**: Add filter commands to preset command router
- **Executor**: New `ExecuteFilterCommand()` following existing patterns

#### HTTP Execution Integration
- **Location**: Filter application in HTTP execution pipeline
- **Timing**: Post-response, pre-TOML conversion (integrates seamlessly with existing Phase 4B response formatting)
- **Process**: HTTP Response → Filter Extraction (gjson) → Smart TOML Conversion → Display
- **Integration with Phase 4B**: Filtering occurs before existing smart JSON→TOML conversion in `src/project/executor/http/display.go`
- **Clean Architecture**: Maintains Unix philosophy - filtering does one job, TOML conversion does another
- **Error Handling**: Silent ignoring of missing fields, no execution breakage

### Future Extensibility

#### Advanced Path Features (Future)
```bash
# Current: Standard indexing
stats[0].base_stat

# Future: Named lookups (if needed)
stats[name=hp].base_stat
stats[type=primary].value
```

#### Blacklist Support (Future)
```bash
# Future syntax for blacklist mode
saul api set blacklist password,secrets,internal_data
```

**Storage format** (same file, different field):
```toml
# Whitelist mode (current implementation)
fields = ["name", "stats[0]"]

# Blacklist mode (future)
# blacklist = ["password", "secrets"] 
```

## Implementation Plan

### Phase 1: Core Whitelist Implementation
1. **Dependency**: Add `github.com/tidwall/gjson` to go.mod
2. **Command Parsing**: Add filter recognition to parser
3. **TOML Storage**: Implement filters.toml handling
4. **Filter Execution**: JSON field extraction using gjson during HTTP calls
5. **Integration**: Connect all components end-to-end

### Phase 2: Command Completeness
1. **Edit Command**: Interactive filter editing (opens file in editor)
2. **Check Command**: Display current filter settings (cat file contents)
3. **Integration Testing**: Validate with real APIs

### Phase 3: Testing & Polish
1. **Test Suite**: Add comprehensive filter testing to test_suite.sh
2. **Real-world Testing**: Validate with PokéAPI, GitHub API, JSONPlaceholder
   - ✅ **PokéAPI Validation Complete**: Tested field paths `name`, `stats[0]`, `stats[1]`, `types[0].type.name` against real API responses - all work correctly
3. **Documentation**: Update README.md with filter examples

### Phase 4: Future Features (If Needed)
1. **Advanced Path Syntax**: Named array lookups
2. **Blacklist Support**: Implement exclusion filtering
3. **Broader Comma-Separated System**: Extend comma syntax to regular fields for batch operations
   - `saul api set body name=pikachu,level=25,type=electric` (multiple key=value pairs)
   - `saul api set header Authorization=Bearer123,Content-Type=application/json` (multiple headers)
   - Maintains single transaction: load TOML → multiple sets → save once
   - Significant productivity enhancement for complex configurations

## Success Criteria

### Primary Goals
- **Problem Solved**: Large API responses fit comfortably in terminal
- **TOML Conversion**: Filtered responses convert cleanly to readable TOML
- **User Experience**: Simple, intuitive commands following Saul patterns
- **Performance**: No significant impact on HTTP execution speed
- **Robustness**: API changes don't break filtering (silent ignoring)

### Quality Metrics
- **Simplicity**: Single command to set up filtering
- **Consistency**: Follows existing Saul command patterns
- **Maintainability**: Clean code integration with existing architecture
- **Reliability**: Silent error handling prevents execution failures

## Technical Implementation Details

### Dependencies
- **Add gjson**: `github.com/tidwall/gjson` for robust JSON path extraction
- **Existing patterns**: Reuse command parsing and file management code
- **Go standard library**: Leverage existing JSON and TOML libraries

### Field Extraction Implementation
```go
import "github.com/tidwall/gjson"

func extractFields(jsonData []byte, fields []string) ([]byte, error) {
    result := make(map[string]interface{})
    
    for _, field := range fields {
        value := gjson.GetBytes(jsonData, field)
        if value.Exists() {
            // Only include fields that actually exist
            setNestedValue(result, field, value.Value())
        }
        // Silent ignore if field doesn't exist
    }
    
    return json.Marshal(result)
}
```

### Error Handling Philosophy
- **No validation errors**: Invalid paths are silently ignored
- **No execution failures**: Filtering never breaks HTTP execution
- **User-friendly**: Clean output without error noise
- **Robust**: Works with any API response structure

### Performance Considerations
- **Lazy loading**: Only load filters.toml when needed
- **Efficient extraction**: gjson provides fast JSON path access
- **Memory efficient**: Process filtering in single pass
- **Minimal overhead**: Simple field extraction without complex processing

## Conclusion

The whitelist filtering system addresses the specific problem of terminal overflow from large API responses while maintaining Saul's core philosophy of simplicity and reliability. The simple dot notation syntax and silent error handling provide a robust, user-friendly solution that integrates seamlessly with existing patterns.

This feature will significantly improve the user experience when working with complex APIs, making Saul the ideal tool for developers who need clean, readable, terminal-friendly API responses.
