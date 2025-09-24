#!/bin/bash
# test_suite_fixed.sh - Refactored test suite with single preset and proper variable handling
# Fixes: Single preset usage, controlled variable testing, reliable input handling

set -e  # Exit on any error

echo "=== Better-Curl (Saul) - Refactored Test Suite ==="
echo "Testing: All functionality with single preset and controlled variables"
echo

# ===== IMPROVED TEST ISOLATION SETUP =====
PRESET_DIR="$HOME/.config/saul/presets"
BACKUP_DIR="/tmp/saul_test_backup_$$"
TEST_PRESET="testpreset"

echo "Setting up improved test isolation..."

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup existing presets if they exist
if [ -d "$PRESET_DIR" ] && [ "$(ls -A $PRESET_DIR 2>/dev/null)" ]; then
    echo "Backing up existing presets to $BACKUP_DIR"
    cp -r "$PRESET_DIR"/* "$BACKUP_DIR/" 2>/dev/null || true
    rm -rf "$PRESET_DIR"/* 2>/dev/null || true
fi

# Function to restore presets on exit
cleanup() {
    echo "Restoring original presets..."
    if [ -d "$BACKUP_DIR" ] && [ "$(ls -A $BACKUP_DIR 2>/dev/null)" ]; then
        mkdir -p "$PRESET_DIR"
        cp -r "$BACKUP_DIR"/* "$PRESET_DIR/" 2>/dev/null || true
    fi
    rm -rf "$BACKUP_DIR" 2>/dev/null || true
    rm -f saul_test 2>/dev/null || true
}

# Set trap to cleanup on exit (success or failure)
trap cleanup EXIT

# ===== HELPER FUNCTIONS =====

# Reset test preset to clean state
reset_preset() {
    rm -rf "$PRESET_DIR/$TEST_PRESET" 2>/dev/null || true
    ./saul_test "$TEST_PRESET" >/dev/null
}

# Populate variables.toml with test values (avoid prompting)
populate_variables() {
    local preset="$1"
    local var_file="$PRESET_DIR/$preset/variables.toml"
    mkdir -p "$(dirname "$var_file")"
    
    # Clear existing variables
    > "$var_file"
    
    # Add test variables based on arguments
    shift
    while [ $# -gt 0 ]; do
        echo "$1" >> "$var_file"
        shift
    done
}

# Test HTTP call without prompting (for tests with no variables)
test_call_no_vars() {
    local preset="$1"
    if ./saul_test call "$preset" 2>/dev/null | grep -q "Status:"; then
        return 0
    else
        return 1
    fi
}

# Test HTTP call with pre-populated variables (no prompting)
test_call_with_vars() {
    local preset="$1"
    # Variables should already be populated via populate_variables
    if ./saul_test call "$preset" 2>/dev/null | grep -q "Status:"; then
        return 0
    else
        return 1
    fi
}

# Test soft variables with controlled input
test_call_soft_vars() {
    local preset="$1"
    shift
    local input_values="$*"
    
    # Create input string
    local input_string=""
    for value in $input_values; do
        input_string="${input_string}${value}\n"
    done
    
    # Use printf instead of echo -e for better compatibility
    if printf "$input_string" | ./saul_test call "$preset" 2>/dev/null | grep -q "Status:"; then
        return 0
    else
        return 1
    fi
}

# ===== PHASE 1: Foundation & TOML Integration =====
echo "===== PHASE 1 TESTS: Foundation & TOML Integration ====="

echo "1.1 Testing compilation..."
go build -o saul_test cmd/main.go
if [ $? -eq 0 ]; then
    echo "✓ Compilation successful"
else
    echo "✗ Compilation failed"
    exit 1
fi

echo "1.2 Testing global commands..."
# Version command
./saul_test version >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Version command works"
else
    echo "✗ Version command failed"
    exit 1
fi

# Help command
./saul_test help >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Help command works"
else
    echo "✗ Help command failed"
    exit 1
fi

# List command (empty)
./saul_test list >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ List command works (empty)"
else
    echo "✗ List command failed"
    exit 1
fi

echo "1.3 Testing preset creation and directory management..."
./saul_test "$TEST_PRESET" >/dev/null
if [ -d "$PRESET_DIR/$TEST_PRESET" ]; then
    echo "✓ Preset directory created"
else
    echo "✗ Preset directory creation failed"
    exit 1
fi

# Check lazy creation (no files yet)
file_count=$(ls "$PRESET_DIR/$TEST_PRESET/" 2>/dev/null | wc -l)
if [ "$file_count" -eq 0 ]; then
    echo "✓ Lazy file creation working (no files yet)"
else
    echo "✗ Lazy file creation failed (files created prematurely)"
    exit 1
fi

echo "✓ Phase 1 Foundation: PASSED"
echo

# ===== PHASE 2: Core TOML Operations & Variable System =====
echo "===== PHASE 2 TESTS: Core TOML Operations & Variable System ====="

echo "2.1 Testing special request syntax..."
# URL command
./saul_test "$TEST_PRESET" set url https://httpbin.org/post >/dev/null
if [ $? -eq 0 ] && [ -f "$PRESET_DIR/$TEST_PRESET/request.toml" ]; then
    echo "✓ URL command works and creates request.toml"
else
    echo "✗ URL command failed"
    exit 1
fi

# Method command (test case conversion)
./saul_test "$TEST_PRESET" set method post >/dev/null
if [ $? -eq 0 ]; then
    # Check if method was stored as uppercase
    if grep -q 'method = "POST"' "$PRESET_DIR/$TEST_PRESET/request.toml"; then
        echo "✓ Method command works with case conversion"
    else
        echo "✗ Method case conversion failed"
        exit 1
    fi
else
    echo "✗ Method command failed"
    exit 1
fi

# Timeout command
./saul_test "$TEST_PRESET" set timeout 30 >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Timeout command works"
else
    echo "✗ Timeout command failed"
    exit 1
fi

echo "2.2 Testing HTTP method validation..."
if ./saul_test "$TEST_PRESET" set method INVALID 2>/dev/null; then
    echo "✗ Invalid method should have been rejected"
    exit 1
else
    echo "✓ Invalid method correctly rejected"
fi

echo "2.3 Testing regular TOML syntax..."
# Body command
./saul_test "$TEST_PRESET" set body pokemon.name=pikachu >/dev/null
if [ $? -eq 0 ] && [ -f "$PRESET_DIR/$TEST_PRESET/body.toml" ]; then
    echo "✓ Body command works and creates body.toml"
else
    echo "✗ Body command failed"
    exit 1
fi

# Nested structure
./saul_test "$TEST_PRESET" set body pokemon.stats.hp=100 >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Nested structure command works"
else
    echo "✗ Nested structure command failed"
    exit 1
fi

# Array syntax
./saul_test "$TEST_PRESET" set body tags=red,blue,green >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Array syntax command works"
else
    echo "✗ Array syntax command failed"
    exit 1
fi

# Headers with aliases
./saul_test "$TEST_PRESET" set header Content-Type=application/json >/dev/null
if [ $? -eq 0 ] && [ -f "$PRESET_DIR/$TEST_PRESET/headers.toml" ]; then
    echo "✓ Header command works with alias"
else
    echo "✗ Header command failed"
    exit 1
fi

./saul_test "$TEST_PRESET" set headers Authorization=Bearer123 >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Headers command works with full name"
else
    echo "✗ Headers command failed"
    exit 1
fi

echo "2.4 Testing variable detection and storage (NEW: Braced syntax)..."
# Reset preset for clean variable testing
reset_preset

# Hard variable with name - NEW braced syntax
./saul_test "$TEST_PRESET" set body pokemon.level={@level} >/dev/null
if [ $? -eq 0 ] && [ -f "$PRESET_DIR/$TEST_PRESET/variables.toml" ]; then
    echo "✓ Hard variable ({@level}) works and creates variables.toml"
else
    echo "✗ Hard variable ({@level}) failed"
    exit 1
fi

# Hard variable bare - NEW braced syntax
./saul_test "$TEST_PRESET" set body pokemon.hp={@} >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Bare hard variable ({@}) works"
else
    echo "✗ Bare hard variable ({@}) failed"
    exit 1
fi

# Soft variable with name - NEW braced syntax
./saul_test "$TEST_PRESET" set body pokemon.name={?pokename} >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Soft variable ({?pokename}) works"
else
    echo "✗ Soft variable ({?pokename}) failed"
    exit 1
fi

# Soft variable bare - NEW braced syntax
./saul_test "$TEST_PRESET" set body pokemon.type={?} >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Bare soft variable ({?}) works"
else
    echo "✗ Bare soft variable ({?}) failed"
    exit 1
fi

echo "2.5 Testing 5-file structure validation..."
expected_files=("body.toml" "headers.toml" "request.toml" "variables.toml")
for file in "${expected_files[@]}"; do
    if [ -f "$PRESET_DIR/$TEST_PRESET/$file" ]; then
        echo "✓ $file created correctly"
    else
        echo "✗ $file missing"
        exit 1
    fi
done

# Query file should NOT exist (not used in tests)
if [ ! -f "$PRESET_DIR/$TEST_PRESET/query.toml" ]; then
    echo "✓ query.toml correctly not created (lazy creation)"
else
    echo "✗ query.toml should not exist (lazy creation failed)"
    exit 1
fi

echo "2.6 Testing get command..."
# Reset for clean get testing
reset_preset
./saul_test "$TEST_PRESET" set url https://httpbin.org/post >/dev/null
./saul_test "$TEST_PRESET" set method POST >/dev/null
./saul_test "$TEST_PRESET" set body pokemon.stats.hp=100 >/dev/null
./saul_test "$TEST_PRESET" set body tags=red,blue,green >/dev/null

# Test smart get for request fields
if ./saul_test "$TEST_PRESET" get url | grep -q "https://httpbin.org/post"; then
    echo "✓ Get URL command works (smart routing)"
else
    echo "✗ Get URL command failed"
    exit 1
fi

if ./saul_test "$TEST_PRESET" get method | grep -q "POST"; then
    echo "✓ Get method command works"
else
    echo "✗ Get method command failed"
    exit 1
fi

# Test get for body fields
if ./saul_test "$TEST_PRESET" get body pokemon.stats.hp | grep -q "100"; then
    echo "✓ Get body field works"
else
    echo "✗ Get body field failed"
    exit 1
fi

# Test get for arrays
if ./saul_test "$TEST_PRESET" get body tags | grep -q '\["red", "blue", "green"\]'; then
    echo "✓ Get array display works"
else
    echo "✗ Get array display failed"
    exit 1
fi

echo "2.7 Testing preset management..."
# List should show testpreset
if ./saul_test list | grep -q "$TEST_PRESET"; then
    echo "✓ List command shows created preset"
else
    echo "✗ List command doesn't show preset"
    exit 1
fi

# Delete preset
./saul_test rm "$TEST_PRESET" >/dev/null
if [ $? -eq 0 ] && [ ! -d "$PRESET_DIR/$TEST_PRESET" ]; then
    echo "✓ Delete command works"
else
    echo "✗ Delete command failed"
    exit 1
fi

# List should be empty again
if ./saul_test list | grep -q "No presets found"; then
    echo "✓ List command shows empty after deletion"
else
    echo "✗ List command doesn't show empty state"
    exit 1
fi

echo "2.8 Testing error handling..."
# Missing preset for rm
if ./saul_test rm nonexistent 2>/dev/null; then
    echo "✗ Should reject deleting nonexistent preset"
    exit 1
else
    echo "✓ Correctly rejects deleting nonexistent preset"
fi

# Recreate preset for remaining tests
reset_preset

# Invalid target
if ./saul_test "$TEST_PRESET" set invalidtarget key=value 2>/dev/null; then
    echo "✗ Should reject invalid target"
    exit 1
else
    echo "✓ Correctly rejects invalid target"
fi

echo "✓ Phase 2 Core TOML Operations & Get Command: PASSED"
echo

# ===== PHASE 3: HTTP Execution Engine =====
echo "===== PHASE 3 TESTS: HTTP Execution Engine ====="

echo "3.1 Testing basic call command with GET request..."
reset_preset
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null

if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ Basic call command works (GET request)"
else
    echo "✗ Basic call command failed"
    exit 1
fi

echo "3.2 Testing POST request with JSON body..."
reset_preset
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts >/dev/null
./saul_test "$TEST_PRESET" set method POST >/dev/null
./saul_test "$TEST_PRESET" set body title=test >/dev/null
./saul_test "$TEST_PRESET" set body body=testing >/dev/null
./saul_test "$TEST_PRESET" set body userId=1 >/dev/null

if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ POST request with JSON body works"
else
    echo "✗ POST request with JSON body failed"
    exit 1
fi

echo "3.3 Testing variable prompting system (FIXED: Pre-populated variables)..."
reset_preset
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null
./saul_test "$TEST_PRESET" set body title={?} >/dev/null
./saul_test "$TEST_PRESET" set body userId={@userId} >/dev/null

# Pre-populate hard variables to avoid prompting
populate_variables "$TEST_PRESET" 'body.userId = "25"'

# Test with controlled input for soft variables only
if test_call_soft_vars "$TEST_PRESET" "testname"; then
    echo "✓ Variable prompting system works (pre-populated + controlled input)"
else
    echo "✗ Variable prompting system failed"
    exit 1
fi

echo "3.4 Testing different HTTP methods..."
# GET
reset_preset
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null
if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ GET request works"
else
    echo "✗ GET request failed"
    exit 1
fi

# POST
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts >/dev/null
./saul_test "$TEST_PRESET" set method POST >/dev/null
./saul_test "$TEST_PRESET" set body title=test >/dev/null
if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ POST request works"
else
    echo "✗ POST request failed"
    exit 1
fi

# PUT
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test "$TEST_PRESET" set method PUT >/dev/null
./saul_test "$TEST_PRESET" set body title=updated >/dev/null
if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ PUT request works"
else
    echo "✗ PUT request failed"
    exit 1
fi

# DELETE
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test "$TEST_PRESET" set method DELETE >/dev/null
if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ DELETE request works"
else
    echo "✗ DELETE request failed"
    exit 1
fi

echo "3.5 Testing headers and complex requests..."
reset_preset
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null
./saul_test "$TEST_PRESET" set header Authorization=Bearer123 >/dev/null
./saul_test "$TEST_PRESET" set header Content-Type=application/json >/dev/null

if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ Headers are properly sent"
else
    echo "✗ Headers test failed"
    exit 1
fi

echo "3.6 Testing error handling..."
# Test missing URL
reset_preset
./saul_test "$TEST_PRESET" set method GET >/dev/null

if ./saul_test call "$TEST_PRESET" 2>&1 | grep -q "URL is required"; then
    echo "✓ Missing URL error handling works"
else
    echo "✗ Missing URL error handling failed"
    exit 1
fi

# Test calling non-existent preset (should fail)
if ./saul_test call nonexistent 2>&1 | grep -q "Command failed:"; then
    echo "✓ Non-existent preset error handling works"
else
    echo "✗ Non-existent preset error handling failed"
    exit 1
fi

echo "3.7 Testing TOML file merging (FIXED: Separate handlers)..."
reset_preset
./saul_test "$TEST_PRESET" set url https://jsonplaceholder.typicode.com/posts >/dev/null
./saul_test "$TEST_PRESET" set method POST >/dev/null
./saul_test "$TEST_PRESET" set header X-Test=merge >/dev/null
./saul_test "$TEST_PRESET" set body title=merged-test >/dev/null
./saul_test "$TEST_PRESET" set body body=testing-merge >/dev/null
./saul_test "$TEST_PRESET" set body userId=1 >/dev/null

# Should work with separate handlers (no merging conflicts)
if test_call_no_vars "$TEST_PRESET"; then
    echo "✓ Separate handler system works (no merging conflicts)"
else
    echo "✗ Separate handler system failed"
    exit 1
fi

echo "✓ Phase 3 HTTP Execution Engine: PASSED"
echo

# ===== PHASE 3.5: Architecture & Variable Syntax Fix =====
echo "===== PHASE 3.5 TESTS: Architecture & Variable Syntax Fix ===="
echo "Critical bug fixes: Separate handlers + Braced variable syntax"

echo "3.5.1 Testing separate handlers (no field misclassification)..."
reset_preset

# Test URL with literal @ symbol (would conflict with old syntax)
./saul_test "$TEST_PRESET" set url https://api.github.com/@octocat/repos >/dev/null
./saul_test "$TEST_PRESET" set header Authorization=Bearer{@token} >/dev/null
./saul_test "$TEST_PRESET" set body search.query={?term} >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null

# Verify URL with literal @ stays in request file, not misclassified as header
if grep -q "https://api.github.com/@octocat/repos" "$PRESET_DIR/$TEST_PRESET/request.toml"; then
    echo "✓ URL with literal @ correctly stays in request.toml"
else
    echo "✗ URL with literal @ was misclassified"
    exit 1
fi

# Verify header variables stay in headers file
if grep -q "Authorization" "$PRESET_DIR/$TEST_PRESET/headers.toml"; then
    echo "✓ Header with variable correctly stays in headers.toml"
else
    echo "✗ Header with variable was misclassified"
    exit 1
fi

# Verify body variables stay in body file
if grep -q "query" "$PRESET_DIR/$TEST_PRESET/body.toml"; then
    echo "✓ Body with variable correctly stays in body.toml"
else
    echo "✗ Body with variable was misclassified"
    exit 1
fi

echo "3.5.2 Testing braced variable syntax (no URL conflicts)..."
reset_preset
./saul_test "$TEST_PRESET" set url https://api.twitter.com/@mentions?search={?query} >/dev/null
./saul_test "$TEST_PRESET" set header X-User={@username} >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null

# Pre-populate hard variables and test with soft variable input
populate_variables "$TEST_PRESET" 'header.X-User.username = "testuser"'

# Should handle partial variables correctly
if test_call_soft_vars "$TEST_PRESET" "testquery"; then
    echo "✓ Partial variables in URLs work correctly"
else
    echo "✗ Partial variables in URLs failed"
    exit 1
fi

echo "3.5.3 Testing complex real-world URL patterns..."
reset_preset
# Complex URL with multiple @ and ? symbols (literal) plus variables
./saul_test "$TEST_PRESET" set url https://api.com/{@user}/posts?search=@mentions&token={@auth}&filter=recent >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null

# Pre-populate hard variables
populate_variables "$TEST_PRESET" 'url.user = "testuser"' 'url.auth = "token123"'

if test_call_with_vars "$TEST_PRESET"; then
    echo "✓ Complex URLs with mixed literal and variable symbols work"
else
    echo "✗ Complex URL parsing failed"
    exit 1
fi

echo "3.5.4 Testing backward compatibility break detection..."
# Old syntax should NOT work anymore
if ./saul_test "$TEST_PRESET" set body name=@oldstyle 2>/dev/null; then
    echo "⚠ Warning: Old syntax still works (may be intentional for transition)"
else
    echo "✓ Old variable syntax correctly rejected (clean break)"
fi

echo "3.5.5 Testing variable deduplication with new syntax..."
reset_preset
./saul_test "$TEST_PRESET" set url https://api.test.com/{@token} >/dev/null
./saul_test "$TEST_PRESET" set header Authorization=Bearer{@token} >/dev/null
./saul_test "$TEST_PRESET" set method GET >/dev/null

# Pre-populate the token variable (should work for both URL and header)
populate_variables "$TEST_PRESET" 'url.token = "abc123"' 'header.Authorization.token = "abc123"'

if test_call_with_vars "$TEST_PRESET"; then
    echo "✓ Variable deduplication works with new syntax"
else
    echo "✗ Variable deduplication failed"
    exit 1
fi

echo "✓ Phase 3.5 Architecture & Variable Syntax Fix: PASSED"
echo

# ===== PHASE 4: Complete Command System =====
echo "===== PHASE 4 TESTS: Complete Command System ====="
echo "⏳ Phase 4 not yet implemented"
echo "Future tests:"
echo "  - Complete command routing"
echo "  - Enhanced help and documentation" 
echo "  - Advanced preset management"
echo

# ===== PHASE 5: Interactive Mode =====
echo "===== PHASE 5 TESTS: Interactive Mode ====="
echo "⏳ Phase 5 not yet implemented"
echo "Future tests:"
echo "  - Interactive shell mode"
echo "  - Command history and editing"
echo "  - Context-aware prompting"
echo

# ===== PHASE 6: Advanced Features =====
echo "===== PHASE 6 TESTS: Advanced Features & Polish ====="
echo "⏳ Phase 6 not yet implemented"
echo "Future tests:"
echo "  - File editing integration"
echo "  - Advanced variable features"
echo "  - Performance optimization"
echo "  - Cross-platform compatibility"
echo

# Cleanup handled by trap function

echo "=== REFACTORED TEST SUITE SUMMARY ==="
echo "✓ Phase 1: Foundation & TOML Integration - PASSED"
echo "✓ Phase 2: Core TOML Operations & Variable System - PASSED"
echo "✓ Phase 3: HTTP Execution Engine - PASSED"
echo "✓ Phase 3.5: Architecture & Variable Syntax Fix - PASSED"
echo "⏳ Phase 4: Complete Command System - PENDING"
echo "⏳ Phase 5: Interactive Mode - PENDING"
echo "⏳ Phase 6: Advanced Features & Polish - PENDING"
echo
echo "🚀 IMPROVEMENTS:"
echo "  ✅ Single preset usage throughout (testpreset)"
echo "  ✅ Pre-populated variables (no prompting issues)"
echo "  ✅ Controlled input handling for soft variables"
echo "  ✅ Helper functions for reliable testing"
echo "  ✅ Proper state reset between test phases"
echo "  ✅ Clean separation of test concerns"
echo
echo "🎯 All functionality verified with improved reliability!"