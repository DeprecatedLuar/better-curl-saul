# Interactive Mode - Future Enhancement

This file contains the extracted interactive mode implementation plan for Better-Curl (Saul). This feature has been deferred to focus on core CLI functionality first.

## Overview

Interactive mode would provide a command shell interface for preset management, allowing users to work within a preset context without repeating the preset name for each command.

## Missing Core Component (Extracted from Action Plan)

**Interactive mode**: Command shell for preset management

## Technical Debt Note

No interactive mode for workflow efficiency

## Implementation Plan

### **Phase 5: Interactive Mode**
*Goal: Working interactive shell for preset management*

#### 5.1 Interactive Shell Implementation
- [ ] **Shell Mode Detection**:
  - Detect when `saul preset` called without additional commands
  - Implement `EnterInteractiveMode(preset string)` function
  - Create command loop with `[preset]> ` prompt showing current preset
  - Handle shell-specific commands: `exit`, `quit`, `help`

- [ ] **Command Processing in Interactive Mode**:
  - Reuse existing command parsing but strip preset name
  - Route commands through same executors as single-line mode
  - Maintain command history within session
  - Handle multi-word commands and proper argument parsing

- [ ] **Interactive User Experience**:
  - Show welcome message: "Entered interactive mode for 'preset'"
  - Display help reminder: "Type 'help' for commands or 'exit' to leave"
  - Handle Ctrl+C gracefully (exit interactive mode, return to shell)
  - Clear error handling without exiting interactive session

#### 5.2 Interactive Command Integration
- [ ] **All Existing Commands Work**:
  - `set url/method/timeout` commands work identically
  - `set body/headers/query` commands work identically
  - `call` command works with variable prompting
  - `check` commands work including history access
  - `rm` commands work with confirmations

- [ ] **Interactive-Specific Enhancements**:
  - Command abbreviation support (optional): `c` for `call`, `s` for `set`
  - Tab completion for commands and targets (optional)
  - Show current configuration summary on demand
  - Context-aware help based on current preset state

#### 5.3 Advanced Interactive Features
- [ ] **Session Management**:
  - Track commands executed in session for debugging
  - Provide session summary on exit
  - Handle long-running sessions gracefully
  - Memory management for extended usage

**Phase 5 Success Criteria:**
- [ ] `saul myapi` enters interactive mode successfully
- [ ] All commands work identically to single-line mode
- [ ] `exit` and Ctrl+C handling works properly
- [ ] Interactive session maintains state correctly
- [ ] Help system works in interactive context
- [ ] User experience feels natural and responsive

**Phase 5 Testing:**
```bash
# Interactive mode testing (manual)
echo "Testing interactive mode..."
echo -e "set url https://httpbin.org/get\nset method GET\ncall\nexit" | saul testapi
echo "✓ Interactive mode basic functionality works"
```

## Implementation Notes

### Integration with Existing Architecture
- Reuse all existing command parsing and execution logic
- Shell mode would be a wrapper around current CLI commands
- No changes needed to core TOML operations or HTTP execution
- Variable prompting system works identically in interactive mode

### User Experience Considerations
- Interactive mode should feel natural for preset-focused workflows
- Command abbreviations could improve efficiency for power users
- Session state management important for long editing sessions
- Graceful exit handling crucial for good user experience

### Technical Considerations
- Command loop implementation using existing parser
- Session state management (current preset context)
- Input handling for multi-line commands and editing
- Cross-platform terminal interaction requirements

## Future Benefits

When implemented, interactive mode would provide:
- **Workflow Efficiency**: Eliminate preset name repetition for focused work
- **Natural Context**: Stay "inside" a preset while configuring and testing
- **Enhanced UX**: Shell-like experience for power users familiar with CLI tools
- **Session Continuity**: Maintain context during extended configuration sessions

## Implementation Priority

This feature has been deferred because:
1. **Core Value First**: Single-line commands already provide excellent functionality
2. **Clean Separation**: Interactive mode is architecturally independent
3. **Learning Focus**: Better to master core HTTP client features first
4. **Solid Foundation**: Want proven filtering, history, and formatting before shell mode

Interactive mode can be implemented later without affecting any existing functionality, as it would be a pure addition to the current command interface.