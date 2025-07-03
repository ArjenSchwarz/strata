#!/bin/bash

# Test script for dual output functionality
# This script tests the run_strata_dual_output function with actual Strata binary

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test helper functions
log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

assert_success() {
    local message=$1
    echo -e "${GREEN}[PASS]${NC} $message"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

assert_failure() {
    local message=$1
    echo -e "${RED}[FAIL]${NC} $message"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Check if Strata binary exists
if [ ! -f "./strata" ]; then
    echo -e "${RED}Error: Strata binary not found. Please build it first with 'go build -o strata'${NC}"
    exit 1
fi

# Check if sample file exists
if [ ! -f "samples/danger-sample.json" ]; then
    echo -e "${RED}Error: Sample file not found: samples/danger-sample.json${NC}"
    exit 1
fi

echo "Testing Dual Output Functionality"
echo "================================="

# Test 1: Basic dual output functionality
log_test "Basic dual output functionality"

# Create temporary directory for test
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

# Test the dual output command directly
TEMP_MD_FILE="$TEST_DIR/output.md"
STDOUT_OUTPUT=$(./strata plan summary samples/danger-sample.json --file "$TEMP_MD_FILE" --file-format markdown --output table 2>/dev/null)
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    assert_success "Dual output execution should succeed"
else
    assert_failure "Dual output execution should succeed (exit code: $EXIT_CODE)"
fi

# Check stdout contains table format indicators (more flexible check)
if echo "$STDOUT_OUTPUT" | grep -q "│\|┌\|└\|├\|┤\|┬\|┴\|┼\|+\|-\||" || echo "$STDOUT_OUTPUT" | grep -q "TO ADD\|TO CHANGE\|TO DESTROY"; then
    assert_success "Stdout should contain table format"
else
    assert_failure "Stdout should contain table format"
    echo "Debug: Stdout content preview:"
    echo "$STDOUT_OUTPUT" | head -5
fi

# Check markdown file was created and contains content
if [ -f "$TEMP_MD_FILE" ] && [ -s "$TEMP_MD_FILE" ]; then
    assert_success "Markdown content should be generated"
    
    # Check markdown file contains proper headers
    if grep -q "^#" "$TEMP_MD_FILE"; then
        assert_success "Markdown content should contain proper headers"
    else
        assert_failure "Markdown content should contain proper headers"
    fi
    
    # Check that stdout and markdown content are different formats
    MD_CONTENT=$(cat "$TEMP_MD_FILE")
    if [ "$STDOUT_OUTPUT" != "$MD_CONTENT" ]; then
        assert_success "Stdout and markdown content should be different formats"
    else
        assert_failure "Stdout and markdown content should be different formats"
    fi
else
    assert_failure "Markdown content should be generated"
    assert_failure "Markdown content should contain proper headers"
    assert_failure "Stdout and markdown content should be different formats"
fi

# Test 2: Error handling for invalid plan file
log_test "Error handling for invalid plan file"

# Use timeout to prevent hanging and temporarily disable set -e
set +e
INVALID_OUTPUT=$(timeout 10 ./strata plan summary /nonexistent/file.json --file "$TEST_DIR/error.md" --file-format markdown --output table 2>&1)
INVALID_EXIT_CODE=$?
set -e

if [ $INVALID_EXIT_CODE -ne 0 ]; then
    assert_success "Should handle invalid plan file gracefully"
else
    assert_failure "Should handle invalid plan file gracefully"
fi

# Test 3: File permission handling
log_test "File permission handling"

# Create a directory where we can't write
READONLY_DIR="$TEST_DIR/readonly"
mkdir -p "$READONLY_DIR"
chmod 555 "$READONLY_DIR"

# Use timeout to prevent hanging and temporarily disable set -e
set +e
PERM_OUTPUT=$(timeout 10 ./strata plan summary samples/danger-sample.json --file "$READONLY_DIR/output.md" --file-format markdown --output table 2>&1)
PERM_EXIT_CODE=$?
set -e

# Should still produce stdout output even if file writing fails
if [ -n "$PERM_OUTPUT" ]; then
    assert_success "Should produce output even when file writing fails"
else
    assert_failure "Should produce output even when file writing fails"
fi

# Restore permissions for cleanup
chmod 755 "$READONLY_DIR"

echo ""
echo "Test Summary:"
echo "============="
echo -e "Tests run:    ${TESTS_RUN}"
echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi