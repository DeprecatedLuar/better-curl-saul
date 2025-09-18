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

### TOML File Structure
Uses proper TOML sections (not flat keys) for clean manual editing that auto-converts to JSON:

**body.toml example:**
```toml
[pokemon]
name = "?name"
level = 25

[pokemon.stats]
hp = 100
attack = "$attack"

[pokemon.abilities]
primary = "static"
secondary = "lightning-rod"
```

**Converts to JSON payload:**
```json
{
  "pokemon": {
    "name": "?name",
    "level": 25,
    "stats": {
      "hp": 100,
      "attack": "$attack"
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
```

**variables.toml example:**
```toml
# Only hard variables stored (soft variables always prompt fresh)
"pokemon.stats.attack" = "80"
"trainer.id" = "ash123"
```

### Variable System

**Soft Variables (Always Prompt):**
- Syntax: `field=?` or `field=?customname`
- Example: `set body pokemon.name=?` → prompts `pokemon.name:` on fire (uses full field path)
- Example: `set body pokemon.name=?pokename` → prompts `pokename:` on fire (uses custom name)

**Hard Variables (Persistent):**
- Syntax: `field=@` or `field=@customname`  
- Example: `set body pokemon.age=@` → prompts `pokemon.age:` when using persistence flag
- Example: `set body pokemon.age=@age` → prompts `age:` when using persistence flag
- Values stored in `variables.toml` as flat key-value pairs
- Prompting shows current value: `attack: 80_` (delete to change, Enter to keep)
- **Storage design**: Only hard variables stored (soft variables never stored - always prompt fresh)
- **Naming**: Bare `@` uses full field path for prompting (no conflicts), named `@customname` uses custom name

**Variable Usage:**
- Variables can be used anywhere: URL, headers, body, query parameters
- Example URL with variables: `https://api.example.com/@version/users/?pokename` (@ and ? in URLs)
- **Future Enhancement**: Variables in request commands: `saul testapi set url https://api.com/@endpoint` and `saul testapi set method @method`

**Smart Variable Deduplication:**
Variables with the same name are prompted only once, allowing consistent values across multiple locations:

```bash
saul api set url https://httpbin.org/@method
saul api set method @method
saul call api
method: post_    # Single prompt fills both URL and method
```

This eliminates redundancy and enforces consistency - perfect for REST APIs where the HTTP method often matches the URL path segment.

### Variable Resolution System
- **Timing**: Variables resolve at `fire` time (not pre-fire)
- **Storage**: Keep resolved data in memory during execution
- **Process**: TOML files → variable resolution → JSON conversion → HTTP execution

### File Management Strategy
- **Approach**: Parse-merge-write (not append-only)
- **Process**: Read existing TOML → Parse → Modify → Write back
- **Benefits**: Reliable, handles conflicts, maintains data integrity
- **Tool**: Repurpose MinseokOh/toml-cli source code for TOML manipulation

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

**Regular TOML Configuration (With = Syntax):**
- `set header key=value` - Add HTTP header
- `set body object.field=value` - Set body parameters using dot notation
- `set query param=value` - Add query parameters
- `set variables varname=value` - Set hard variable values directly

**Inspection Commands:**
- `check url` - Display current URL (smart routing to request.toml)
- `check method` - Display current HTTP method  
- `check body pokemon.name` - Display specific field with formatting
- `check headers` - Display all headers (full file view)

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

**Management:**
- `version` (alias: `v`) - Show version
- `remove` (alias: `rm`) - Remove configurations
- `edit` (alias: `ed`) - Edit configurations
- `list` - Show all presets
- `rm presetname` - Delete preset (with confirmation)

### Command Structure
```
saul [global] [preset] [command] [target] [field=value]

Examples:
saul pokeapi set header Authorization=Bearer123
saul pokeapi set body pokemon.name=?
saul call pokeapi
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
TOML files → Parse-merge-write → Variable resolution → JSON conversion → HTTP execution
```

### Implementation Priority Order
1. **TOML manipulation system** (parse-merge-write approach) ✅ **COMPLETED**
2. **Variable substitution system** (`?/@` variable handling) ✅ **COMPLETED**
3. **JSON conversion** (TOML → Go structs → JSON) ✅ **COMPLETED**
4. **HTTP execution engine** (using go-resty) ✅ **COMPLETED**
5. **Single-line commands** (primary interface) ✅ **COMPLETED**
6. **Interactive mode** (secondary interface built on single-line) ⏳ **PENDING**

### Libraries and Dependencies
- `github.com/pelletier/go-toml/v1` - TOML parsing and manipulation
- `github.com/go-resty/resty/v2` - HTTP client library
- `github.com/DeprecatedLuar/toml-vars-letsgooo` - Existing tomv integration
- Standard library `os`, `filepath` - File operations

## User Experience Goals
- **Simple:** Intuitive commands that feel natural
- **Clean:** No JSON escaping or single-line nightmares
- **Flexible:** Both interactive and scriptable modes
- **Reusable:** Save and reuse complex request configurations
- **Interactive:** Smart prompting for variable values
- **Readable:** Pretty-formatted response display for easy analysis

## Target Users
- Developers testing APIs
- DevOps engineers automating HTTP requests
- Anyone frustrated with curl's complexity for structured data

## Development Philosophy
- **KISS Principles:** Simple, intelligent, self-maintained, resilient code
- **AI-Assisted Development:** Leverage AI for rapid iteration and learning
- **Parse-merge-write:** Reliable over fast for file operations
- **Single-line first:** Build interactive mode on proven single-line foundation

## Final Polish & Easter Eggs

### Better Call Saul Easter Egg
- **Command:** `saul call saul`
- **Behavior:** Opens browser to random Better Call Saul video
- **URLs:**
  - https://www.youtube.com/watch?v=gDjMZvYWUdo
  - https://www.youtube.com/watch?v=zj2IhcuS5iM
  - https://www.youtube.com/watch?v=SH_mdu8W0bc
  - https://www.youtube.com/watch?v=z9_OX1WVXXU
  - https://www.youtube.com/watch?v=XfQQ7CIOEoM
  - https://www.youtube.com/watch?v=pL4fke8vkFE
- **Implementation:** Random URL selection + cross-platform browser opening with graceful fallback
- **Priority:** Implement after all core functionality is complete and tested
