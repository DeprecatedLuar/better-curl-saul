# Better-Curl (Saul) - Project Vision

## Core Problem
Eliminate the pain of cramming complex JSON payloads and HTTP configurations into single curl command lines. No more escaping hell or unreadable one-liners.

## Solution Approach
A workspace-based HTTP client that builds requests incrementally using separate files for each component (headers, body, query parameters).

## Key Concepts

### Presets (Workspaces)
- Each preset is a folder containing TOML files that define an HTTP request
- Stored in `~/.config/saul/presets/[preset-name]/`
- Contains separate files for different request components:
  - `headers.toml` - HTTP headers
  - `body.toml` - Request body/payload (converts to JSON)
  - `query.toml` - Query parameters
  - `config.toml` - URL, method, and persistent hard variables

### TOML File Structure
Uses sectioned TOML for clean manual editing that auto-converts to JSON:

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

**config.toml example:**
```toml
url = "https://pokeapi.co/api/v2/pokemon"
method = "GET"

[hard_variables]
attack = 80
trainer_id = "ash123"
```

**File Editing:**
- `saul preset edit header` - Opens headers.toml in default editor
- `saul preset edit body` - Opens body.toml in default editor
- `saul preset edit config` - Opens config.toml in default editor

### Dual Command Modes

**Interactive Mode:**
```bash
saul pokeapi          # Enter preset mode
> set header Content-Type=application/json
> set body pokemon.name=pikachu
> fire                # Execute request
> exit                # Exit preset mode
```

**Single-line Mode:**
```bash
saul pokeapi set header Content-Type=application/json
saul pokeapi set body pokemon.name=pikachu
saul pokeapi fire
```

### Variable System

**Soft Variables (Always Prompt):**
- Syntax: `field=?` or `field=?customname`
- Example: `set body pokemon.name=?` → prompts `name:` on fire
- Example: `set body pokemon.name=?pokename` → prompts `pokename:` on fire

**Hard Variables (Persistent with Flag):**
- Syntax: `field=$` or `field=$customname`
- Example: `set body pokemon.age=$` → prompts `age:` when using persistence flag
- Values stored in `config.toml` and remembered between sessions
- Prompting shows current value: `attack: 80_` (delete to change, Enter to keep)

### Core Commands

**Configuration:**
- `set url [METHOD] https://api.example.com` - Set endpoint and method
- `set header key=value` - Add HTTP header
- `set body object.field=value` - Set body parameters using dot notation
- `set query param=value` - Add query parameters

**Execution:**
- `fire` - Execute HTTP request (prompts for soft variables only)
- `fire --persist` - Execute with prompting for both soft and hard variables

**Variable Prompting Flow:**
```bash
> fire --persist
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

**Preset Management:**
```bash
saul list                    # Show all presets
saul rm pokeapi             # Delete preset (prompts for confirmation)
```

### Command Structure
```
saul [preset] [command] [target] [field=value]

Examples:
saul pokeapi set header Authorization=Bearer123
saul pokeapi set body pokemon.name=?
saul pokeapi fire
```

### Intelligent Type Detection
Automatic detection based on syntax - no verbose type declarations needed:

**Arrays (Comma Detection):**
```bash
set body tags=red,blue,green              # Auto-detects array
set body pokemon.abilities=static,lightning-rod
```

**Objects (Dot Notation Detection):**
```bash
set body pokemon.name=pikachu            # Auto-detects nested object
set body user.profile.settings.theme=dark
```

**Simple Variables:**
```bash
set body level=25                        # Auto-detects simple key-value
set header Authorization=Bearer123
```

### Dot Notation Support
Build complex nested objects naturally:
```bash
set body user.profile.name=john
set body user.profile.age=25
set body user.settings.notifications=true
```

## Technical Approach
- **Language:** Go for fast, single-binary distribution
- **Config Format:** TOML files for human-readable configuration
- **TOML Library:** `github.com/pelletier/go-toml` for automatic TOML-to-JSON conversion
- **Architecture:** Command parser → TOML file management → JSON conversion → HTTP execution
- **File Storage:** `~/.config/saul/` following Linux/Unix conventions

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