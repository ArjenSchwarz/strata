#!/bin/bash
#
# Simple test for file operations (Task 7.1)
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

TEST_OUTPUT_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_OUTPUT_DIR"' EXIT

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

log_section() {
    echo ""
    echo -e "${BLUE}=== $1 ===${NC}"
    echo ""
}

assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="$3"

    if [[ "$expected" == "$actual" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected: '$expected'"
        echo "      Actual:   '$actual'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local message="$2"

    if [[ -f "$file" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected file to exist: $file"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_dir_exists() {
    local dir="$1"
    local message="$2"

    if [[ -d "$dir" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected directory to exist: $dir"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

test_mktemp_functionality() {
    log_section "mktemp Functionality Test"

    log_test "Testing mktemp -d creates temp directories"

    local temp1 temp2
    temp1=$(mktemp -d)
    temp2=$(mktemp -d)

    assert_dir_exists "$temp1" "First temp directory should exist"
    assert_dir_exists "$temp2" "Second temp directory should exist"

    if [[ "$temp1" != "$temp2" ]]; then
        echo -e "${GREEN}  ✓${NC} Temp directories are unique"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}  ✗${NC} Temp directories should be unique"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_RUN=$((TESTS_RUN + 1))

    # Clean up
    rm -rf "$temp1" "$temp2"
}

test_trap_cleanup() {
    log_section "Trap Cleanup Test"

    log_test "Testing cleanup on normal exit"

    # Create test script
    cat > "$TEST_OUTPUT_DIR/test_cleanup.sh" <<'EOF'
#!/bin/bash
set -euo pipefail

TEMP_DIR=$(mktemp -d)
CLEANUP_MARKER="/tmp/cleanup_test_marker"

cleanup() {
    touch "$CLEANUP_MARKER"
    [[ -d "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"
}

trap cleanup EXIT

echo "test content" > "$TEMP_DIR/test_file"
exit 0
EOF

    local marker="/tmp/cleanup_test_marker"
    rm -f "$marker"

    bash "$TEST_OUTPUT_DIR/test_cleanup.sh"

    assert_file_exists "$marker" "Cleanup should run on normal exit"
    rm -f "$marker"

    log_test "Testing cleanup on error exit"

    # Modify script to exit with error
    sed 's/exit 0/exit 1/' "$TEST_OUTPUT_DIR/test_cleanup.sh" > "$TEST_OUTPUT_DIR/test_cleanup_error.sh"

    set +e
    bash "$TEST_OUTPUT_DIR/test_cleanup_error.sh" 2>/dev/null
    local exit_code=$?
    set -e

    assert_file_exists "$marker" "Cleanup should run on error exit"

    if [[ $exit_code -ne 0 ]]; then
        echo -e "${GREEN}  ✓${NC} Script exited with error code as expected"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}  ✗${NC} Script should have exited with error code"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_RUN=$((TESTS_RUN + 1))

    rm -f "$marker"
}

test_performance() {
    log_section "Performance Test (< 50ms)"

    log_test "Testing file operations performance"

    local start_time end_time duration
    start_time=$(date +%s%N)

    local temp_dir
    temp_dir=$(mktemp -d)

    # Standard file operations
    echo "test content" > "$temp_dir/test1.txt"
    cat "$temp_dir/test1.txt" > "$temp_dir/test2.txt"
    cp "$temp_dir/test1.txt" "$temp_dir/test3.txt"
    rm "$temp_dir/test1.txt"

    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))  # Convert to milliseconds

    if [[ $duration -lt 50 ]]; then
        echo -e "${GREEN}  ✓${NC} File operations completed in ${duration}ms (< 50ms)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}  ✗${NC} File operations too slow: ${duration}ms (should be < 50ms)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    TESTS_RUN=$((TESTS_RUN + 1))

    rm -rf "$temp_dir"
}

test_simple_structure() {
    log_section "Simple Directory Structure Test"

    log_test "Testing no complex directory structures"

    local temp_dir
    temp_dir=$(mktemp -d)

    # Simulate action operations
    echo "binary content" > "$temp_dir/strata.tar.gz"
    echo "checksum content" > "$temp_dir/checksums.txt"
    echo "json content" > "$temp_dir/metadata.json"
    echo "output content" > "$temp_dir/display_output.txt"

    # Count directories (should only be the temp_dir itself)
    local dir_count
    dir_count=$(find "$temp_dir" -type d | wc -l | tr -d ' ')

    assert_equals "1" "$dir_count" "Should only have single temp directory, no subdirectories"

    # Test that all expected files exist
    assert_file_exists "$temp_dir/strata.tar.gz" "Binary file should exist"
    assert_file_exists "$temp_dir/checksums.txt" "Checksum file should exist"
    assert_file_exists "$temp_dir/metadata.json" "JSON file should exist"
    assert_file_exists "$temp_dir/display_output.txt" "Output file should exist"

    rm -rf "$temp_dir"
}

# Run tests
echo "================================================"
echo "File Operations Tests (Task 7.1)"
echo "================================================"

test_mktemp_functionality
test_trap_cleanup
test_performance
test_simple_structure

# Print summary
echo ""
echo "================================================"
echo "Test Summary"
echo "================================================"
echo -e "Tests run:     ${TESTS_RUN}"
echo -e "Tests passed:  ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Tests failed:  ${RED}${TESTS_FAILED}${NC}"
echo ""

if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}SOME TESTS FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
    exit 0
fi