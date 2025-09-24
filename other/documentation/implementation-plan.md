# Implementation Plan

<!-- WORKFLOW IMPLEMENTATION GUIDE:
- This file contains active phases for implementation (completed phases moved to implementation-history.md)
- Each phase = one focused session, follow top-to-bottom order
- Focus on actionable steps: "Update file X, add function Y"
- Avoid verbose explanations - just implement what's specified and valuable
- Success criteria must be testable
- Make sure to test implementation after conclusion of phase
- Stop implementation and call out ideas if you find better approaches during implementation
-->

# Better-Curl (Saul) - Action Plan

## Current Project Status

Better-Curl (Saul) is **feature-complete** for HTTP client functionality with all core phases (0-6A) implemented. The project has:
- Full HTTP execution engine with variable system
- Response formatting and filtering capabilities
- Terminal session memory and response history
- Production-ready command interface with Unix integration

**Current Focus**: Code quality improvements and minor enhancements as identified in code reviews.

## Active Development Phases

### Phase 5B: Display System Migration & Check Command Enhancement
**Status**: Medium Priority
**Objective**: Clean up remaining display system inconsistencies and optimize check command behavior

#### 5B.1 Check Command Behavior Update
**Current Issue**: Check commands show entire file content when viewing individual fields
**Proposed Change**: Return just the field value for better Unix composition
```bash
# Current behavior:
saul api check body | grep pokemon    # Shows entire TOML file

# Proposed behavior:
saul api check body pokemon.name      # Shows just "pikachu"
```

**Implementation Steps**:
1. Update `src/project/handlers/commands/check.go`
2. Add field-specific logic in `ExecuteCheckCommand()`
3. Preserve full file display when no specific field requested
4. Update test suite with new check command tests

**Success Criteria**:
- Field-specific checks return raw values only
- Full file checks remain unchanged
- All existing functionality preserved
- Unix composition improved

#### 5B.2 Display System Audit
**Objective**: Identify and consolidate any remaining display inconsistencies
**Tasks**:
1. Audit all remaining `fmt.Print*` usage (currently 20 legitimate cases)
2. Verify error message consistency with `src/modules/errors/messages.go`
3. Ensure raw mode philosophy correctly implemented across all commands

### Phase 6: Advanced Features & Polish
**Status**: Future Enhancement
**Priority**: Low (quality of life improvements)

#### 6.1 File Editing Integration
**Concept**: Direct TOML file editing with `$EDITOR`
```bash
saul api edit --file body              # Opens body.toml in editor
saul api edit --file                   # Opens entire preset directory
```

**Implementation Considerations**:
- File locking for concurrent editing scenarios
- Validation after editor close
- Integration with existing validation system
- Handle editor exit codes and cancellation

#### 6.2 Advanced Variable Features
**Concept**: Enhanced variable management
- Variable listing: `saul api check variables`
- Variable clearing: `saul api unset @variable_name`
- Environment variable integration: `{$ENV_VAR}` syntax
- Variable templates for common patterns

#### 6.3 Production Readiness
**Distribution Preparation**:
- Cross-platform binary builds (Linux, macOS, Windows)
- Package manager integration (brew, apt, etc.)
- Installation script with configuration setup
- Man page documentation
- Shell completion scripts (bash, zsh, fish)

**Security Enhancements**:
- Credential management integration
- API key storage best practices
- Secure variable storage options
- File permission validation

## Testing Strategy

### Current Test Coverage
All implemented phases have comprehensive test coverage in `other/testing/test_suite.sh`:
- Foundation & TOML operations (Phase 1-2)
- HTTP execution engine (Phase 3)
- Edit commands (Phase 4A)
- Response formatting (Phase 4B)
- Syntax enhancements (Phase 4B-Post)
- Filtering system (Phase 4C)
- History system (Phase 4E)
- Flag system (Phase 5A)

### Test Enhancement for Phase 5B
```bash
# New Phase 5B tests to add:
echo "Testing Phase 5B: Display System Migration"

# Test field-specific check behavior
saul test set body pokemon.name=pikachu
result=$(saul test check body pokemon.name)
if [ "$result" = "pikachu" ]; then
    echo "✓ Field-specific check returns raw value"
else
    echo "✗ Field-specific check failed"
fi

# Test full file check behavior (unchanged)
result=$(saul test check body | wc -l)
if [ "$result" -gt 1 ]; then
    echo "✓ Full file check shows complete TOML"
else
    echo "✗ Full file check failed"
fi
```

## Development Guidelines

### Code Quality Standards
- **KISS Principles**: Prioritize simple, clean solutions over complex architectures
- **Zero Regression**: All changes must preserve existing functionality
- **Unix Philosophy**: Small, composable functions with clear single responsibilities
- **Go Best Practices**: Follow established patterns from completed phases

### Change Management
- **Test First**: Add test coverage before implementing new features
- **Incremental**: Small, focused changes with clear validation
- **Documentation**: Update CLAUDE.md and README.md for significant changes
- **Backward Compatibility**: Maintain compatibility with existing configurations

## Success Metrics

### Phase 5B Completion Criteria
- [ ] Field-specific check commands return raw values only
- [ ] Full file check behavior preserved and unchanged
- [ ] All existing Phase 1-5A functionality unchanged
- [ ] Test suite expanded with Phase 5B validation
- [ ] Unix composition improved for check commands

### Future Phase Success
- [ ] Direct file editing integration working with common editors
- [ ] Advanced variable features enhance workflow efficiency
- [ ] Production distribution ready for public release
- [ ] Security best practices implemented for credential management

## Priority Assessment

**High Priority**: None (project is feature-complete)
**Medium Priority**: Phase 5B display system cleanup
**Low Priority**: Phase 6 advanced features and production polish

The project has achieved its core objectives and is ready for production use. Remaining work focuses on quality of life improvements and distribution preparation.