# HTTP Client CLI Formatting Specification

## Overview
This document defines the visual formatting system for Better-Curl (Saul) - our Go-based HTTP client CLI tool. The system uses a consistent "visual sandwich" approach with headers, content, and footers to create clear content boundaries and professional terminal output.

## Core Design Principles

### 1. Visual Sandwich Pattern
All formatted output uses a three-part structure:
- **Header**: Metadata about the content (status, file type, etc.)
- **Content**: The actual data (TOML format)
- **Footer**: Closing separator (same width as header)

### 2. Responsive Width
Separator lines adapt to terminal size:
- Use **80 characters** as the target width
- If terminal width < 100 chars, use **80% of terminal width**
- Never exceed terminal boundaries

### 3. Saul-Specific Integration
- Integrates with existing `src/modules/display/` system
- Uses current TOML display patterns from Phase 4B
- Maintains consistency with established error/success messaging
- Leverages existing response filtering from Phase 4C

## Formatting Patterns

### HTTP Response Display (Primary Use Case)
```
Response: 200 OK • 1.2KB • application/json
────────────────────────────────────────────────────────────────────────────────
[response]
message = "Success"
timestamp = "2024-01-15T10:30:00Z"
data = { users = 42, active = true }
────────────────────────────────────────────────────────────────────────────────
```

**Header format**: `Response: {status_code} {status_text} • {size} • {content_type}`

### File Display (Request Bodies, Headers, Filters)
```
Headers • 0.5KB • 3 entries
────────────────────────────────────────────────────────────────────────────────
[headers]
authorization = "Bearer token"
content-type = "application/json"  
user-agent = "saul/1.0"
────────────────────────────────────────────────────────────────────────────────
```

**Header format**: `{file_type} • {size} • {entry_count} entries`

### Check Command Integration (Saul-Specific)
When user runs `saul api check {field}` (e.g., `check url`, `check method`), always display the entire containing file with full formatting, not just the requested field.

**Example**:
```bash
$ saul pokeapi check url
Request • 0.1KB • 2 entries
────────────────────────────────────────────────────────────────────────────────
method = "GET"
url = "https://pokeapi.co/api/v2/pokemon/{@pokename}"
────────────────────────────────────────────────────────────────────────────────
```

### Filter Display Integration (Phase 4C)
```
Filtered Response: 200 OK • 0.3KB (filtered from 257KB) • application/json
────────────────────────────────────────────────────────────────────────────────
[filtered_fields]
name = "pikachu"
stats = { 0 = { base_stat = 35 } }
types = { 0 = { type = { name = "electric" } } }
────────────────────────────────────────────────────────────────────────────────
```

## Implementation Details for Saul

### Separator Width Logic
```go
func getSeparatorWidth() int {
    termWidth := getTerminalWidth()
    maxWidth := int(float64(termWidth) * 0.8)
    
    // Use 80 chars unless terminal is too narrow
    if maxWidth < 80 {
        return maxWidth
    }
    return 80
}

func createSeparator(width int) string {
    return strings.Repeat("─", width) // Unicode U+2500
}
```

### Integration Points in Saul Architecture

#### 1. Update `src/project/executor/http/display.go` (Phase 4B)
- Add formatting wrapper around existing JSON→TOML conversion
- Enhance response display with sandwich formatting
- Integrate with current content-type detection

#### 2. Update `src/project/executor/commands.go` (Check Commands)
- Wrap existing TOML display in check commands with formatting
- Maintain current functionality while adding visual enhancement

#### 3. Add to `src/modules/display/` System
- Create new `FormatSection()` function for reusable formatting
- Integrate with existing `printer.go` system
- Maintain consistency with current error/success patterns

### File Types and Metadata (Saul-Specific)

#### Request Files (`request.toml`)
- **Display Name**: "Request"
- **Count Logic**: Count of configuration properties (method, url, timeout, etc.)
- **Example**: `Request • 0.1KB • 3 properties`

#### Headers Files (`headers.toml`)
- **Display Name**: "Headers"
- **Count Logic**: Number of HTTP header fields
- **Example**: `Headers • 0.3KB • 4 headers`

#### Body Files (`body.toml`)
- **Display Name**: "Body"
- **Count Logic**: Number of top-level body fields
- **Example**: `Body • 0.5KB • 6 fields`

#### Query Files (`query.toml`)
- **Display Name**: "Query Parameters"
- **Count Logic**: Number of query parameter fields
- **Example**: `Query Parameters • 0.2KB • 3 parameters`

#### Filter Files (`filters.toml`)
- **Display Name**: "Response Filters"
- **Count Logic**: Number of configured filter fields
- **Example**: `Response Filters • 0.1KB • 5 filters`

#### Variables Files (`variables.toml`)
- **Display Name**: "Variables"
- **Count Logic**: Number of stored variable values
- **Example**: `Variables • 0.1KB • 3 variables`

### Response Display Enhancement
- **Size Calculation**: Show response size and filtered size when applicable
- **Status Integration**: Use existing HTTP status handling from Phase 4B
- **Content-Type**: Leverage existing content-type detection
- **Filtering Info**: Show filtering information when filters applied

## Required Go Libraries (Saul-Compatible)
- `golang.org/x/term` - for terminal width detection (preferred over deprecated crypto/ssh/terminal)
- `strings` - for separator generation (already used)
- File size calculation utilities (can reuse from existing HTTP display)

## Implementation Strategy

### Phase 1: Core Formatting Functions
1. Create `FormatSection(title, content, metadata string) string` in display package
2. Add terminal width detection utilities
3. Create separator generation functions

### Phase 2: Response Display Integration
1. Update `DisplayResponse()` in `http/display.go`
2. Integrate with existing JSON→TOML conversion pipeline
3. Add response metadata formatting

### Phase 3: Check Command Integration
1. Update `ExecuteCheckCommand()` in `commands.go`
2. Wrap existing TOML display with formatting
3. Add file-specific metadata calculation

### Phase 4: Universal Application
1. Apply formatting to all TOML displays consistently
2. Update error/warning displays if needed
3. Ensure visual consistency across all commands

## Examples by File Type (Saul Context)

### Request File Display
```
Request • 0.1KB • 3 properties
────────────────────────────────────────────────────────────────────────────────
method = "GET"
url = "https://pokeapi.co/api/v2/pokemon/{@pokename}"
timeout = "30s"
────────────────────────────────────────────────────────────────────────────────
```

### Headers File Display
```
Headers • 0.3KB • 4 headers
────────────────────────────────────────────────────────────────────────────────
authorization = "Bearer xyz123"
content-type = "application/json"
user-agent = "saul/1.0"
accept = "application/json"
────────────────────────────────────────────────────────────────────────────────
```

### Filtered Response Display
```
Filtered Response: 200 OK • 0.8KB (from 257KB) • application/json
────────────────────────────────────────────────────────────────────────────────
[pokemon]
name = "pikachu"
stats = { 0 = { base_stat = 35, stat = { name = "hp" } } }
types = { 0 = { type = { name = "electric" } } }
────────────────────────────────────────────────────────────────────────────────
```

## Error Cases and Saul-Specific Handling
- If terminal width cannot be determined, fallback to 80 characters
- If file is empty, show "0 entries/fields/properties" in header
- If file read fails, use existing error display system from `src/modules/errors/`
- Maintain consistent error formatting with current patterns

## Consistency Rules for Saul
1. Always use the same separator character: `─` (U+2500)
2. Always include file size in KB format (0.1KB, 1.2KB, etc.)
3. Always use bullet separator `•` in headers
4. Always include entry/field/property counts when applicable
5. Always use opening and closing separators (sandwich pattern)
6. Integrate seamlessly with existing `display.Error()`, `display.Success()` functions
7. Maintain current TOML structure from Phase 4B JSON→TOML conversion
8. Preserve all existing functionality while adding visual enhancement

## Benefits for Saul Users
- **Professional Appearance**: Clean, organized terminal output suitable for development workflows
- **Clear Content Boundaries**: Sandwich formatting eliminates visual confusion
- **Responsive Design**: Works well on different terminal sizes
- **Consistent UX**: Same visual patterns across all commands reduce cognitive load
- **Enhanced Readability**: Metadata headers provide context for displayed content
- **Integration Ready**: Builds on existing Phase 4B/4C systems without disruption