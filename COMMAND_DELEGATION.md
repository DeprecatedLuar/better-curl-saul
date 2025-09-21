# Command Delegation Architecture

## Overview

Better-Curl (Saul) implements a **workspace-aware command delegation system** that allows users to execute system commands within the context of their preset workspace. This follows pure Unix composition philosophy where tools are combined to create more powerful workflows.

## Core Concept

Instead of forcing users to learn new syntax, Saul delegates familiar system commands to their native implementations while automatically operating in the correct workspace context.

**Traditional Approach (Rejected):**
```bash
saul list -la        # Forces new "list" command
saul list --tree     # Doesn't leverage existing tool knowledge
```

**Delegation Approach (Implemented):**
```bash
saul ls -la          # Delegates to system `ls` in preset workspace
saul exa --tree      # Delegates to system `exa` in preset workspace
saul dir /w          # Delegates to system `dir` on Windows
```

## Architecture Pattern

### Command Resolution Flow

```
User Input → Command Parsing → System Command Detection → Delegation or Preset Handling
```

1. **Parse Command**: `saul ls -la` → Command{Name: "ls", Args: ["-la"]}
2. **Check Whitelist**: Is "ls" in allowed system commands?
3. **Delegate**: If yes, execute `ls /preset/workspace -la`
4. **Fallback**: If no, handle as preset command

### Implementation Structure

```go
func ExecuteCommand(cmd Command) error {
    if isAllowedSystemCommand(cmd.Name) {
        return delegateToSystem(cmd.Name, cmd.Args)
    }
    return executePresetCommand(cmd)
}

func delegateToSystem(command string, args []string) error {
    presetDir := getPresetDirectory()
    fullArgs := append([]string{presetDir}, args...)

    execCmd := exec.Command(command, fullArgs...)
    execCmd.Stdout = os.Stdout
    execCmd.Stderr = os.Stderr

    return execCmd.Run()
}
```

## Supported System Commands

### Safe Read-Only Commands (Whitelisted)

| Command | Platform | Purpose | Example Usage |
|---------|----------|---------|---------------|
| `ls` | Linux/Mac | List directory contents | `saul ls -la --color` |
| `exa` | Cross-platform | Modern ls replacement | `saul exa --tree --git` |
| `lsd` | Cross-platform | LSDeluxe listing | `saul lsd --tree --icon` |
| `tree` | Cross-platform | Directory tree view | `saul tree -C -L 2` |
| `dir` | Windows | Windows directory listing | `saul dir /w /p` |

### Security Considerations

**Whitelist-Only Approach**: Only explicitly allowed commands can be delegated to prevent security risks.

**Excluded Commands** (Never allowed):
- Destructive: `rm`, `del`, `rmdir`
- File modification: `mv`, `cp`, `move`
- System access: `sudo`, `su`, `chmod`

## User Experience Examples

### Cross-Platform Listing

**Linux/Mac Users:**
```bash
saul ls -la                    # Long format with hidden files
saul ls -t | head -5          # 5 most recent presets
saul ls *.api                 # Glob pattern filtering
```

**Windows Users:**
```bash
saul dir /w                   # Wide format listing
saul dir /s /b               # Bare format with subdirectories
```

**Power Users with Modern Tools:**
```bash
saul exa --tree --git --icons # Tree view with git status and icons
saul lsd --tree --depth 2     # Limited depth tree with LSDeluxe
```

### Command Composition

**Unix Pipeline Integration:**
```bash
saul ls -1 | grep api        # Filter presets containing "api"
saul tree | grep -E '\.toml$' # Find all TOML files in tree
saul exa --long | sort -k5    # Sort by file size
```

## Configuration Options

### Future Enhancement: Command Preferences

**User Configuration** (`~/.config/saul/config.toml`):
```toml
[commands]
list = "exa"           # Use exa instead of ls by default
list_args = ["--git"] # Default arguments for list command

[aliases]
ll = "ls -la"         # Custom command aliases
tree = "exa --tree"   # Override system tree with exa tree
```

### Platform Auto-Detection

**Smart Defaults by Platform:**
- **Linux**: `ls` → `exa` (if available) → `ls`
- **Mac**: `ls` → `exa` (if available) → `ls`
- **Windows**: `dir` → `ls` (if Git Bash) → `dir`

## Implementation Benefits

### For Users
1. **Zero Learning Curve**: Know `ls`? You know `saul ls`
2. **Platform Native**: Each OS uses its optimal tools
3. **Tool Flexibility**: Works with any listing tool user prefers
4. **Pipeline Compatible**: Works seamlessly with Unix pipes and filters

### For Developers
1. **Code Simplicity**: Delegate complexity to proven system tools
2. **Maintenance Reduction**: No custom listing logic to maintain
3. **Feature Inheritance**: Get all tool features automatically
4. **Cross-Platform**: One pattern works everywhere

### For the Project
1. **Unix Philosophy**: "Do one thing well" - compose existing tools
2. **Extensibility**: Easy to add new system commands
3. **User Empowerment**: Users choose their preferred tools
4. **Backward Compatibility**: Preset commands unchanged

## Command Namespace Resolution

### Priority Order

1. **System Commands with Arguments**: `saul ls -la` → Delegate to system
2. **Exact Preset Match**: `saul pokeapi` → Handle as preset
3. **System Commands without Arguments**: `saul ls` → Delegate (show preset workspace)

### Avoiding Collisions

**Whitelist Strategy**: Only safe, explicitly allowed commands are delegated.

**Preset Namespace Protection**: Preset names take priority over system commands when unambiguous.

**Future Enhancement - Explicit Syntax**:
```bash
saul @ls -la         # @ prefix forces system command
saul ls              # Always treated as preset (if exists)
```

## Technical Implementation Notes

### Process Management

**Parent-Child Relationship**:
```
saul (parent) → ls (child) → output to terminal
```

**Stream Handling**:
- Child stdout → Parent stdout (direct terminal output)
- Child stderr → Parent stderr (error passthrough)
- Exit codes bubble up from child to parent

### Error Handling

**Command Not Found**:
```bash
$ saul invalidcommand
Error: Command 'invalidcommand' not found and no preset exists
```

**System Command Errors**:
```bash
$ saul ls --invalid-flag
ls: invalid option -- 'invalid-flag'
Try 'ls --help' for more information.
```

## Future Enhancements

### Phase 1: Basic Delegation (Current)
- ✅ Hardcoded whitelist of safe commands
- ✅ Simple argument passthrough
- ✅ Cross-platform command detection

### Phase 2: Configuration System
- ⏳ User-configurable command preferences
- ⏳ Custom aliases and default arguments
- ⏳ Platform-specific defaults

### Phase 3: Advanced Features
- ⏳ Command validation and suggestions
- ⏳ Integration with preset workspace state
- ⏳ Custom command plugins

## Related Documentation

- [README.md](./README.md) - Project overview and core concepts
- [CLAUDE.md](./CLAUDE.md) - Development guidance and architecture
- [action-plan.md](./other/documentation/action-plan.md) - Implementation roadmap

---

**Design Philosophy**: Embrace the Unix way - compose simple tools to create powerful workflows. Don't reinvent what already works perfectly.