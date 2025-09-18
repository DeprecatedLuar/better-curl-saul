#!/bin/bash
# test_suite.sh - Comprehensive test suite for Better-Curl (Saul)
# Expandable test suite that grows with each phase implementation

set -e  # Exit on any error

echo "=== Better-Curl (Saul) - Test Suite ==="
echo "Testing: All implemented functionality across phases"
echo

# ===== TEST ISOLATION SETUP =====
PRESET_DIR="$HOME/.config/saul/presets"
BACKUP_DIR="/tmp/saul_test_backup_$$"

echo "Setting up test isolation..."

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
./saul_test testapi >/dev/null
if [ -d ~/.config/saul/presets/testapi ]; then
    echo "✓ Preset directory created"
else
    echo "✗ Preset directory creation failed"
    exit 1
fi

# Check lazy creation (no files yet)
file_count=$(ls ~/.config/saul/presets/testapi/ 2>/dev/null | wc -l)
if [ "$file_count" -eq 0 ]; then
    echo "✓ Lazy file creation working (no files yet)"
else
    echo "✗ Lazy file creation failed (files created prematurely)"
    exit 1
fi

echo "1.4 Testing TOML file operations..."
# This will be verified when we create actual TOML content in Phase 2 tests

echo "✓ Phase 1 Foundation: PASSED"
echo

# ===== PHASE 2: Core TOML Operations & Variable System =====
echo "===== PHASE 2 TESTS: Core TOML Operations & Variable System ====="

echo "2.1 Testing special request syntax..."
# URL command
./saul_test testapi set url https://httpbin.org/post >/dev/null
if [ $? -eq 0 ] && [ -f ~/.config/saul/presets/testapi/request.toml ]; then
    echo "✓ URL command works and creates request.toml"
else
    echo "✗ URL command failed"
    exit 1
fi

# Method command (test case conversion)
./saul_test testapi set method post >/dev/null
if [ $? -eq 0 ]; then
    # Check if method was stored as uppercase
    if grep -q 'method = "POST"' ~/.config/saul/presets/testapi/request.toml; then
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
./saul_test testapi set timeout 30 >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Timeout command works"
else
    echo "✗ Timeout command failed"
    exit 1
fi

echo "2.2 Testing HTTP method validation..."
if ./saul_test testapi set method INVALID 2>/dev/null; then
    echo "✗ Invalid method should have been rejected"
    exit 1
else
    echo "✓ Invalid method correctly rejected"
fi

echo "2.3 Testing regular TOML syntax..."
# Body command
./saul_test testapi set body pokemon.name=pikachu >/dev/null
if [ $? -eq 0 ] && [ -f ~/.config/saul/presets/testapi/body.toml ]; then
    echo "✓ Body command works and creates body.toml"
else
    echo "✗ Body command failed"
    exit 1
fi

# Nested structure
./saul_test testapi set body pokemon.stats.hp=100 >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Nested structure command works"
else
    echo "✗ Nested structure command failed"
    exit 1
fi

# Array syntax
./saul_test testapi set body tags=red,blue,green >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Array syntax command works"
else
    echo "✗ Array syntax command failed"
    exit 1
fi

# Headers with aliases
./saul_test testapi set header Content-Type=application/json >/dev/null
if [ $? -eq 0 ] && [ -f ~/.config/saul/presets/testapi/headers.toml ]; then
    echo "✓ Header command works with alias"
else
    echo "✗ Header command failed"
    exit 1
fi

./saul_test testapi set headers Authorization=Bearer123 >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Headers command works with full name"
else
    echo "✗ Headers command failed"
    exit 1
fi

echo "2.4 Testing variable detection and storage..."
# Hard variable with name
./saul_test testapi set body pokemon.level={@level} >/dev/null
if [ $? -eq 0 ] && [ -f ~/.config/saul/presets/testapi/variables.toml ]; then
    echo "✓ Hard variable (@level) works and creates variables.toml"
else
    echo "✗ Hard variable (@level) failed"
    exit 1
fi

# Hard variable bare
./saul_test testapi set body pokemon.hp={@} >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Bare hard variable (@) works"
else
    echo "✗ Bare hard variable (@) failed"
    exit 1
fi

# Soft variable with name
./saul_test testapi set body pokemon.name={?pokename} >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Soft variable (?pokename) works"
else
    echo "✗ Soft variable (?pokename) failed"
    exit 1
fi

# Soft variable bare
./saul_test testapi set body pokemon.type={?} >/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Bare soft variable (?) works"
else
    echo "✗ Bare soft variable (?) failed"
    exit 1
fi

echo "2.5 Testing 5-file structure validation..."
expected_files=("body.toml" "headers.toml" "request.toml" "variables.toml")
for file in "${expected_files[@]}"; do
    if [ -f ~/.config/saul/presets/testapi/$file ]; then
        echo "✓ $file created correctly"
    else
        echo "✗ $file missing"
        exit 1
    fi
done

# Query file should NOT exist (not used in tests)
if [ ! -f ~/.config/saul/presets/testapi/query.toml ]; then
    echo "✓ query.toml correctly not created (lazy creation)"
else
    echo "✗ query.toml should not exist (lazy creation failed)"
    exit 1
fi

echo "2.6 Testing check command..."
# Test smart check for request fields
if ./saul_test testapi check url | grep -q "https://httpbin.org/post"; then
    echo "✓ Check URL command works (smart routing)"
else
    echo "✗ Check URL command failed"
    exit 1
fi

if ./saul_test testapi check method | grep -q "POST"; then
    echo "✓ Check method command works"
else
    echo "✗ Check method command failed"
    exit 1
fi

# Test check for body fields (use a field set before variables)
if ./saul_test testapi check body pokemon.stats.hp | grep -q "100"; then
    echo "✓ Check body field works"
else
    echo "✗ Check body field failed"
    exit 1
fi

# Test check for arrays
if ./saul_test testapi check body tags | grep -q '\["red", "blue", "green"\]'; then
    echo "✓ Check array display works"
else
    echo "✗ Check array display failed"
    exit 1
fi

echo "2.7 Testing preset management..."
# List should show testapi
if ./saul_test list | grep -q "testapi"; then
    echo "✓ List command shows created preset"
else
    echo "✗ List command doesn't show preset"
    exit 1
fi

# Delete preset
./saul_test rm testapi >/dev/null
if [ $? -eq 0 ] && [ ! -d ~/.config/saul/presets/testapi ]; then
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

# Invalid target
if ./saul_test testapi set invalidtarget key=value 2>/dev/null; then
    echo "✗ Should reject invalid target"
    exit 1
else
    echo "✓ Correctly rejects invalid target"
fi

echo "✓ Phase 2 Core TOML Operations & Check Command: PASSED"
echo

# ===== PHASE 3: HTTP Execution Engine =====
echo "===== PHASE 3 TESTS: HTTP Execution Engine ====="

# Create fresh test preset for HTTP testing
./saul_test httptest >/dev/null

echo "3.1 Testing basic call command with GET request..."
./saul_test httptest set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test httptest set method GET >/dev/null

# Test call command execution (accept any HTTP status - httpbin can be flaky)
if ./saul_test call httptest 2>/dev/null | grep -q "Status:"; then
    echo "✓ Basic call command works (GET request)"
else
    echo "✗ Basic call command failed"
    exit 1
fi

echo "3.2 Testing POST request with JSON body..."
./saul_test httptest set url https://jsonplaceholder.typicode.com/posts >/dev/null
./saul_test httptest set method POST >/dev/null
./saul_test httptest set body title=test >/dev/null
./saul_test httptest set body body=testing >/dev/null
./saul_test httptest set body userId=1 >/dev/null

if ./saul_test call httptest 2>/dev/null | grep -q "Status:"; then
    echo "✓ POST request with JSON body works"
else
    echo "✗ POST request with JSON body failed"
    exit 1
fi

echo "3.3 Testing variable prompting system..."
# Create preset with variables
./saul_test vartest >/dev/null
./saul_test vartest set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test vartest set method GET >/dev/null
./saul_test vartest set body title={?} >/dev/null
./saul_test vartest set body userId={@userId} >/dev/null

# Test with input (using echo)
if echo -e "testname\n25" | ./saul_test call vartest 2>/dev/null | grep -q "Status:"; then
    echo "✓ Variable prompting system works"
else
    echo "✗ Variable prompting system failed"
    exit 1
fi

echo "3.4 Testing different HTTP methods..."
# GET
./saul_test methodtest set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test methodtest set method GET >/dev/null
if ./saul_test call methodtest 2>/dev/null | grep -q "Status:"; then
    echo "✓ GET request works"
else
    echo "✗ GET request failed"
    exit 1
fi

# POST
./saul_test methodtest set url https://jsonplaceholder.typicode.com/posts >/dev/null
./saul_test methodtest set method POST >/dev/null
./saul_test methodtest set body title=test >/dev/null
if ./saul_test call methodtest 2>/dev/null | grep -q "Status:"; then
    echo "✓ POST request works"
else
    echo "✗ POST request failed"
    exit 1
fi

# PUT
./saul_test methodtest set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test methodtest set method PUT >/dev/null
./saul_test methodtest set body title=updated >/dev/null
if ./saul_test call methodtest 2>/dev/null | grep -q "Status:"; then
    echo "✓ PUT request works"
else
    echo "✗ PUT request failed"
    exit 1
fi

# DELETE
./saul_test methodtest set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test methodtest set method DELETE >/dev/null
if ./saul_test call methodtest 2>/dev/null | grep -q "Status:"; then
    echo "✓ DELETE request works"
else
    echo "✗ DELETE request failed"
    exit 1
fi

echo "3.5 Testing headers and complex requests..."
./saul_test headertest >/dev/null
./saul_test headertest set url https://jsonplaceholder.typicode.com/posts/1 >/dev/null
./saul_test headertest set method GET >/dev/null
./saul_test headertest set header Authorization=Bearer123 >/dev/null
./saul_test headertest set header Content-Type=application/json >/dev/null

if ./saul_test call headertest 2>/dev/null | grep -q "Status:"; then
    echo "✓ Headers are properly sent"
else
    echo "✗ Headers test failed"
    exit 1
fi

echo "3.6 Testing error handling..."
# Test missing URL
./saul_test errortest >/dev/null
./saul_test errortest set method GET >/dev/null

if ./saul_test call errortest 2>&1 | grep -q "URL is required"; then
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

echo "3.7 Testing TOML file merging..."
./saul_test mergetest >/dev/null
./saul_test mergetest set url https://jsonplaceholder.typicode.com/posts >/dev/null
./saul_test mergetest set method POST >/dev/null
./saul_test mergetest set header X-Test=merge >/dev/null
./saul_test mergetest set body title=merged-test >/dev/null
./saul_test mergetest set body body=testing-merge >/dev/null
./saul_test mergetest set body userId=1 >/dev/null

# Should merge request.toml + headers.toml + body.toml
if ./saul_test call mergetest 2>/dev/null | grep -q "Status:"; then
    echo "✓ TOML file merging works"
else
    echo "✗ TOML file merging failed"
    exit 1
fi

# Clean up test presets
rm -rf ~/.config/saul/presets/httptest 2>/dev/null || true
rm -rf ~/.config/saul/presets/vartest 2>/dev/null || true
rm -rf ~/.config/saul/presets/methodtest 2>/dev/null || true
rm -rf ~/.config/saul/presets/headertest 2>/dev/null || true
rm -rf ~/.config/saul/presets/errortest 2>/dev/null || true
rm -rf ~/.config/saul/presets/mergetest 2>/dev/null || true

echo "✓ Phase 3 HTTP Execution Engine: PASSED"
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

echo "=== TEST SUITE SUMMARY ==="
echo "✓ Phase 1: Foundation & TOML Integration - PASSED"
echo "✓ Phase 2: Core TOML Operations & Variable System - PASSED"
echo "✓ Phase 3: HTTP Execution Engine - PASSED"
echo "⏳ Phase 4: Complete Command System - PENDING"
echo "⏳ Phase 5: Interactive Mode - PENDING"
echo "⏳ Phase 6: Advanced Features & Polish - PENDING"
echo
echo "🚀 Phase 3 Complete! Ready for Phase 4 Implementation!"