# Better-Curl (Saul) - Project Vision

## Core Problem
Eliminate the pain of cramming complex JSON payloads and HTTP configurations into single curl command lines. No more escaping hell or unreadable one-liners.

## Solution Approach
A workspace-based HTTP client that builds requests incrementally using separate files for each component (headers, body, query parameters).

## Key Concepts

### Presets (Workspaces)
- Each preset is a folder containing TOML files that define an HTTP request
- Stored in `~/.config/saul/presets/[preset-name]/`
- Contains separate files for different request components (Unix philosophy - one purpose per file):
  - `headers.toml` - HTTP headers
  - `body.toml` - Request body/payload (converts to JSON)
  - `query.toml` - Query/search payload data (NOT URL parameters)
  - `request.toml` - HTTP method, URL, and request settings
  - `variables.toml` - Hard variables only (soft variables never stored)
  - `filters.toml` - Response filtering configuration (optional)
  - `history/` - Response history storage (optional, per-preset)

### TOML File Structure
Uses proper TOML sections (not flat keys) for clean manual editing that auto-converts to JSON:

**body.toml example:**
```toml
[pokemon]
name = "{?name}"
level = 25

[pokemon.stats]
hp = 100
attack = "{@attack}"

[pokemon.abilities]
primary = "static"
secondary = "lightning-rod"
```

**Converts to JSON payload:**
```json
{
  "pokemon": {
    "name": "{?name}",
    "level": 25,
    "stats": {
      "hp": 100,
      "attack": "{@attack}"
    },
    "abilities": {
      "primary": "static",
      "secondary": "lightning-rod"
    }
  }
}
```

**request.toml example:**
```toml
method = "POST"
url = "https://pokeapi.co/api/v2/pokemon"
timeout = 30

[settings]
history_count = 5  # 0 = disabled, N = keep last N responses
```

**variables.toml example:**
```toml
# Only hard variables stored (soft variables always prompt fresh)
"pokemon.stats.attack" = "80"
"trainer.id" = "ash123"
```

### Variable System

**Soft Variables (Always Prompt):**
- Syntax: `field={?}` or `field={?customname}`
- Example: `set body pokemon.name={?}` → prompts `pokemon.name:` on call (uses full field path)
- Example: `set body pokemon.name={?pokename}` → prompts `pokename:` on call (uses custom name)

**Hard Variables (Persistent):**
- Syntax: `field={@}` or `field={@customname}`  
- Example: `set body pokemon.age={@}` → prompts `pokemon.age:` when using persistence flag
- Example: `set body pokemon.age={@age}` → prompts `age:` when using persistence flag
- Values stored in `variables.toml` as flat key-value pairs
- Prompting shows current value: `attack: 80_` (delete to change, Enter to keep)
- **Storage design**: Only hard variables stored (soft variables never stored - always prompt fresh)
- **Naming**: Bare `{@}` uses full field path for prompting (no conflicts), named `{@customname}` uses custom name

**Variable Usage:**
- Variables can be used anywhere: URL, headers, body, query parameters
- Example URL with variables: `https://api.example.com/{@version}/users/{?pokename}` (braced syntax prevents conflicts)
- **Future Enhancement**: Variables in request commands: `saul testapi set url https://api.com/{@endpoint}` and `saul testapi set method {@method}`

**Smart Variable Deduplication:**
Variables with the same name are prompted only once, allowing consistent values across multiple locations:

```bash
saul api set url https://httpbin.org/{@method}
saul api set method {@method}
saul call api
method: post_    # Single prompt fills both URL and method
```

This eliminates redundancy and enforces consistency - perfect for REST APIs where the HTTP method often matches the URL path segment.

### Variable Resolution System
- **Timing**: Variables resolve at `call` time (not pre-call)
- **Storage**: Keep resolved data in memory during execution
- **Process**: TOML files → variable resolution → JSON conversion → HTTP execution

### Response Filtering System
Keep API responses readable and terminal-friendly by filtering large JSON responses to only essential fields.

**Core Concept:**
- **Problem**: APIs like PokéAPI return 100+ fields that flood terminal displays
- **Solution**: Whitelist filtering shows only specified fields
- **Philosophy**: Terminal-friendly responses without losing raw data storage

**Command Syntax:**
```bash
# Set response filter (space-separated field paths)
saul pokeapi set filters field1=name field2=stats.0.base_stat field3=types.0.type.name

# Edit filters interactively
saul pokeapi edit filters

# Check current filter settings
saul pokeapi check filterss

# Clear filters (show all fields)
saul pokeapi set filters field1=""

# Filters apply automatically during calls
saul pokeapi call
```

**Field Path Syntax (Industry Standard):**
- **Basic**: `name`, `id`, `stats` (top-level fields or entire objects)
- **Nested**: `types[0].type.name`, `pokemon.stats.hp` (dot notation)
- **Arrays**: `stats[0]`, `moves[5].move.name` (bracket notation)

**filters.toml Storage:**
```toml
fields = [
    "name",
    "stats[0]", 
    "stats[1]",
    "types[0].type.name"
]
```

**Execution Flow:**
```
HTTP Response → Filter Extraction → Smart TOML Conversion → Display
```

**Key Features:**
- **Silent Error Handling**: Missing fields are ignored, no execution breakage
- **Real-world Tested**: Validated against PokéAPI, GitHub API, JSONPlaceholder
- **Terminal Optimized**: Large responses become readable and manageable
- **Integration**: Works seamlessly with existing smart JSON→TOML response formatting

### Response History System
- **Storage**: `~/.config/saul/presets/[preset]/history/response-001.json` (numbered, latest first)
- **Configuration**: Per-preset history count in `request.toml` under `[settings]`
- **Rotation**: Automatic cleanup when limit exceeded (delete oldest, keep newest N)

#### Response Display & Formatting
- **Storage Format**: Always preserve exact raw response with metadata (timestamp, request details, status, headers)
- **Display Format**: Smart content-type based formatting for optimal readability
  - **JSON responses** → Convert to TOML for clean, readable display (innovative approach)
  - **HTML/XML/Text/Other** → Display raw content as-is (future: syntax highlighting)
  - **Error fallback** → Always display raw content if conversion fails

#### Display Options
- **Default**: `saul api check history 1` - Smart formatting in terminal
- **Raw mode**: `saul api check history 1 --raw` - Exact server response in terminal  
- **Editor mode**: `saul api check history 1 -e` - Open formatted content in read-only editor
- **Raw editor**: `saul api check history 1 -e --raw` - Open raw response in read-only editor

#### Example: JSON Response Formatting
**Original JSON Response:**
```json
{"name":"pikachu","id":25,"types":[{"slot":1,"type":{"name":"electric"}}]}
```

**Smart TOML Display:**
```toml
Status: 200 OK (324ms, 2.1KB)
Content-Type: application/json

name = "pikachu"
id = 25

[[types]]
slot = 1

[types.type]
name = "electric"
```

This dual approach optimizes for both debugging fidelity (raw storage) and human readability (smart display).

### File Management Strategy
- **Approach**: Parse-merge-write (not append-only)
- **Process**: Read existing TOML → Parse → Modify → Write back
- **Benefits**: Reliable, handles conflicts, maintains data integrity
- **Tool**: Repurposed MinseokOh/toml-cli source code for TOML manipulation

### Dot Notation Support
Dot notation creates proper TOML sections:
```bash
saul pokeapi set body pokemon.stats.hp=100
```
Creates:
```toml
[pokemon]
[pokemon.stats]
hp = 100
```

### Array Handling
Use TOML native array syntax:
```bash
set body tags=red,blue,green              # Auto-detects array
```
Creates:
```toml
tags = ["red", "blue", "green"]
```

### Dual Command Modes

**Single-line Mode (Primary):**
```bash
saul pokeapi set header Content-Type=application/json
saul pokeapi set body pokemon.name=pikachu
saul call pokeapi
```

**Interactive Mode (Secondary):**
```bash
saul pokeapi          # Enter preset mode
> set header Content-Type=application/json
> set body pokemon.name=pikachu
> call                # Execute request
> exit                # Exit preset mode
```

### Core Commands

**Special Request Configuration (No = Syntax):**
- `set url https://api.example.com` - Set endpoint URL
- `set method POST` - Set HTTP method (GET, POST, PUT, DELETE, etc.)
- `set timeout 30` - Set request timeout in seconds
- `set history N` - Set response history count (0 = disabled)
- `set filters field1=path1 field2=path2 field3=path3` - Set response filtering (space-separated field paths)

**Regular TOML Configuration (With = Syntax):**
- `set header key=value` - Add HTTP header
- `set body object.field=value` - Set body parameters using dot notation
- `set query param=value` - Add query parameters
- `set variables varname=value` - Set hard variable values directly

**Inspection Commands:**
- `check url` - Display current URL (smart routing to request.toml)
- `check method` - Display current HTTP method  
- `check filters` - Display current response filter settings
- `check body pokemon.name` - Display specific field with formatting
- `check headers` - Display all headers (full file view)
- `check history` - Interactive history menu or show available responses
- `check history N` - Show specific response (1 = most recent, 2 = second most recent)
- `check history last` - Show most recent response

**History Management:**
- `rm history` - Delete all stored responses (with confirmation: "Delete all history for 'preset'? (y/N):")

**Execution:**
- `call preset` - Execute HTTP request (prompts for soft variables only)
- `call preset --persist` - Execute with prompting for both soft and hard variables

**Variable Prompting Flow:**
```bash
> call pokeapi --persist
name: ____                    # Soft variable (always empty)
attack: 80_                   # Hard variable (shows current value)
trainer_id: ash123_           # Hard variable (shows current value)
```

**Editing Commands:**
- `edit url` - Pre-filled prompt for quick URL edits (ideal for variable syntax changes)
- `edit filters` - Pre-filled prompt for response filter editing
- `edit body pokemon.name` - Pre-filled prompt for specific field editing
- `edit header Authorization` - Pre-filled prompt for specific header editing
- `edit @pokename` - Pre-filled prompt for editing stored hard variable values
- `edit body` - Opens entire body.toml in default editor (complex editing)
- `edit header` - Opens entire headers.toml in default editor (complex editing)
- `edit query` - Opens entire query.toml in default editor (complex editing)

**Edit Command Behavior:**
- **Field-level editing** (e.g., `edit url`, `edit body pokemon.name`) → Pre-filled interactive prompt for quick field content tweaks
- **Variable editing** (e.g., `edit @pokename`) → Pre-filled prompt for editing stored hard variable values
- **Container-level editing** (e.g., `edit body`, `edit header`) → Opens entire file in default editor
- **Field creation safety**: Non-existent fields prompt "Field 'path' doesn't exist. Create? (y/N)"
- **Variable editing safety**: Non-existent variables show error "Variable '@name' not found. Create variables by using them in fields first."
- **Variable cleanup**: Orphaned variables (no longer used in fields) are kept in variables.toml - no automatic cleanup
- **Primary use case**: Quick variable syntax changes (`{@var}` ↔ `{?var}`) and hard variable value updates without retyping

**Management:**
- `version` (alias: `v`) - Show version
- `remove` (alias: `rm`) - Remove configurations
- `list` - Show all presets
- `rm presetname` - Delete preset (with confirmation)

### Workspace Navigation

**System Command Delegation:**
Saul delegates familiar system commands to operate within your preset workspace, combining the power of existing tools with workspace-aware context.

**Cross-Platform Listing Commands:**
```bash
# Linux/Mac users - use familiar ls with all native flags
saul ls -la --color=always        # Long format with colors
saul ls -t | head -5              # 5 most recent presets
saul ls *.api                     # Glob pattern filtering

# Windows users - use native dir command
saul dir /w /p                    # Wide format with pagination

# Power users with modern tools
saul exa --tree --git --icons     # Tree view with git status
saul lsd --tree --depth 2         # LSDeluxe with limited depth
saul tree -C -L 2                 # Colored tree, 2 levels deep
```

**Command Composition & Pipelines:**
```bash
# Unix pipeline integration works seamlessly
saul ls -1 | grep api             # Filter presets containing "api"
saul exa --long | sort -k5        # Sort presets by file size
saul tree | grep -E '\.toml$'     # Find all TOML files in preset tree
```

**Supported Commands:**
- `ls` - Linux/Mac directory listing (all flags supported)
- `dir` - Windows directory listing (all flags supported)
- `exa` - Modern ls replacement with git integration
- `lsd` - LSDeluxe with icons and colors
- `tree` - Directory tree visualization

**Security & Safety:**
- **Whitelist-only**: Only safe, read-only commands are delegated
- **No destructive operations**: `rm`, `del`, `mv` commands are never delegated
- **Native tool behavior**: All arguments pass through to system commands unchanged

**Design Philosophy:**
Instead of forcing users to learn new listing syntax, Saul embraces Unix composition - use the tools you already know, but automatically operate in your preset workspace context. This eliminates the need to `cd ~/.config/saul/presets` before listing presets.

**Future Enhancement:**
User-configurable command preferences and platform-aware defaults in `~/.config/saul/config.toml`. See [COMMAND_DELEGATION.md](./COMMAND_DELEGATION.md) for detailed architecture documentation.

### Command Structure
```
saul [global] [preset] [command] [target] [field=value]

Examples:
saul pokeapi set header Authorization=Bearer123
saul pokeapi set body pokemon.name={?}
saul pokeapi set filters field1=name field2=stats.0.base_stat field3=types.0.type.name
saul call pokeapi
saul pokeapi check filters
saul pokeapi check history
saul pokeapi rm history
```

### File Editing
- `saul preset edit header` - Opens headers.toml in default editor
- `saul preset edit body` - Opens body.toml in default editor
- `saul preset edit config` - Opens config.toml in default editor

## Technical Implementation

### Architecture Stack
- **Language:** Go for fast, single-binary distribution
- **TOML Library:** pelletier/go-toml for TOML manipulation
- **HTTP Client:** go-resty/resty for clean HTTP requests
- **File Storage:** `~/.config/saul/` following Linux/Unix conventions
- **TOML Manipulation:** Repurposed code from MinseokOh/toml-cli

### Data Pipeline
```
TOML files → Parse-merge-write → Variable resolution → JSON conversion → HTTP execution → Response filtering → Response history storage
```

### Implementation Priority Order
1. **TOML manipulation system** (parse-merge-write approach) ✅ **COMPLETED**
2. **Variable substitution system** (`{?}/{@}` variable handling) ✅ **COMPLETED** 
3. **JSON conversion** (TOML → Go structs → JSON) ✅ **COMPLETED**
4. **HTTP execution engine** (using go-resty) ✅ **COMPLETED**
5. **Single-line commands** (primary interface) ✅ **COMPLETED**
6. **Response filtering system** (whitelist field extraction) ✅ **Phase 4C - COMPLETED**
7. **Response history system** ⏳ **Phase 4D - PENDING**
8. **Interactive mode** (secondary interface built on single-line) ⏳ **Phase 5 - PENDING**

### Libraries and Dependencies
- `github.com/pelletier/go-toml/v1` - TOML parsing and manipulation
- `github.com/go-resty/resty/v2` - HTTP client library
- `github.com/tidwall/gjson` - JSON path extraction for response filtering
- `github.com/DeprecatedLuar/toml-vars-letsgooo` - Existing tomv integration
- Standard library `os`, `filepath` - File operations

## User Experience Goals
- **Simple:** Intuitive commands that feel natural
- **Clean:** No JSON escaping or single-line nightmares
- **Flexible:** Both interactive and scriptable modes
- **Reusable:** Save and reuse complex request configurations
- **Interactive:** Smart prompting for variable values
- **Readable:** Pretty-formatted response display for easy analysis
- **Terminal-Friendly:** Response filtering keeps complex APIs manageable
- **Debuggable:** Response history for API development and troubleshooting
- **Productive:** Comma-separated syntax for batch operations (future enhancement)

## Target Users
- Developers testing APIs
- DevOps engineers automating HTTP requests
- Anyone frustrated with curl's complexity for structured data

## Development Philosophy
- **KISS Principles:** Simple, intelligent, self-maintained, resilient code
- **AI-Assisted Development:** Leverage AI for rapid iteration and learning
- **Parse-merge-write:** Reliable over fast for file operations
- **Single-line first:** Build interactive mode on proven single-line foundation
- **Unix Philosophy:** Each file has one purpose, commands are composable

## Final Polish & Easter Eggs

### Better Call Saul Easter Egg
- **Command:** `saul call saul`
- **Behavior:** Opens browser to random Better Call Saul video
- **URLs:**
  - https://www.youtube.com/watch?v=gDjMZvYWUdo
  - https://www.youtube.com/watch?v=zj2IhcuS5iM
  - https://www.youtube.com/watch?v=SH_mdu8W0bc
  - https://www.youtube.com/watch?v=z9_OX1WVXXU
  - https://www.youtube.com/Watch?v=XfQQ7CIOEoM
  - https://www.youtube.com/watch?v=pL4fke8vkFE
- **Implementation:** Random URL selection + cross-platform browser opening with graceful fallback
- **Priority:** Implement after all core functionality is complete and tested