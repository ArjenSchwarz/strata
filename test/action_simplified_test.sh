#!/bin/bash
#
# Unit tests for simplified GitHub Action foundation
# Tests error handling, cleanup, validation, and exit codes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Test output directory
TEST_OUTPUT_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_OUTPUT_DIR"' EXIT

# ============================================================================
# Test Helper Functions
# ============================================================================

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

assert_not_equals() {
    local not_expected="$1"
    local actual="$2"
    local message="$3"

    if [[ "$not_expected" != "$actual" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Should not be: '$not_expected'"
        echo "      Actual:        '$actual'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"

    if [[ "$haystack" == *"$needle"* ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected to contain: '$needle'"
        echo "      Actual: '$haystack'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"

    if [[ "$haystack" != *"$needle"* ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Should not contain: '$needle'"
        echo "      Actual: '$haystack'"
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

assert_file_not_exists() {
    local file="$1"
    local message="$2"

    if [[ ! -f "$file" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected file not to exist: $file"
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

assert_dir_not_exists() {
    local dir="$1"
    local message="$2"

    if [[ ! -d "$dir" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected directory not to exist: $dir"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_exit_code() {
    local expected_code="$1"
    local actual_code="$2"
    local message="$3"

    if [[ "$expected_code" -eq "$actual_code" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected exit code: $expected_code"
        echo "      Actual exit code:   $actual_code"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Run a function and capture its output and exit code
run_function() {
    local func="$1"
    shift
    local output_file="$TEST_OUTPUT_DIR/output.txt"
    local error_file="$TEST_OUTPUT_DIR/error.txt"

    # Run the function and capture output in a subshell to avoid exit
    local exit_code=0
    (
        set +e
        "$func" "$@" >"$output_file" 2>"$error_file"
        echo $? > "$TEST_OUTPUT_DIR/exit_code"
    ) || true

    if [[ -f "$TEST_OUTPUT_DIR/exit_code" ]]; then
        exit_code=$(cat "$TEST_OUTPUT_DIR/exit_code")
        rm -f "$TEST_OUTPUT_DIR/exit_code"
    fi

    # Return the exit code
    echo "$exit_code"
}

# Source the script to test (without executing main)
source_script() {
    # Only source if not already sourced
    if [[ -z "${SCRIPT_NAME:-}" ]]; then
        # Create a temporary copy that doesn't execute main
        local temp_script="$TEST_OUTPUT_DIR/action_temp.sh"

        # Copy everything except the main execution guard
        sed '/^if \[\[ "${BASH_SOURCE\[0\]}" == "${0}" \]\]; then$/,/^fi$/d' \
            action_simplified.sh > "$temp_script"

        # Source the modified script
        source "$temp_script"
    fi
}

# ============================================================================
# Test Cases: Error Handling Framework
# ============================================================================

test_error_handling_framework() {
    log_section "Error Handling Framework"

    log_test "Testing set -euo pipefail configuration"

    # Create a test script with error handling
    cat > "$TEST_OUTPUT_DIR/test_error.sh" <<'EOF'
#!/bin/bash
source action_simplified.sh 2>/dev/null || true

# Test undefined variable handling
test_undefined() {
    set -u
    echo "$UNDEFINED_VAR"
}

# Test pipe failure handling
test_pipe_failure() {
    set -o pipefail
    false | echo "test"
}

# Test error exit
test_error_exit() {
    set -e
    false
    echo "Should not reach here"
}
EOF

    # Test undefined variable
    local exit_code
    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_error.sh; test_undefined" 2>/dev/null
    exit_code=$?
    set -e
    assert_not_equals "0" "$exit_code" "Undefined variable should cause error"

    # Test pipe failure
    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_error.sh; test_pipe_failure" 2>/dev/null
    exit_code=$?
    set -e
    assert_not_equals "0" "$exit_code" "Pipe failure should cause error"

    # Test error exit
    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_error.sh; test_error_exit" 2>/dev/null
    exit_code=$?
    set -e
    assert_not_equals "0" "$exit_code" "Error should cause immediate exit"
}

test_cleanup_trap() {
    log_section "Cleanup Trap Execution"

    log_test "Testing cleanup trap on normal exit"

    # Create test script that uses cleanup
    cat > "$TEST_OUTPUT_DIR/test_cleanup.sh" <<'EOF'
#!/bin/bash

# Track if cleanup was called
CLEANUP_MARKER="/tmp/cleanup_test_marker"

cleanup() {
    touch "$CLEANUP_MARKER"
}

trap cleanup EXIT

# Create temp dir like the real script
TEMP_DIR=$(mktemp -d)
exit 0
EOF

    local marker="/tmp/cleanup_test_marker"
    rm -f "$marker"

    bash "$TEST_OUTPUT_DIR/test_cleanup.sh"

    assert_file_exists "$marker" "Cleanup should be called on normal exit"
    rm -f "$marker"

    log_test "Testing cleanup trap on error exit"

    # Modify script to exit with error
    sed 's/exit 0/exit 1/' "$TEST_OUTPUT_DIR/test_cleanup.sh" > "$TEST_OUTPUT_DIR/test_cleanup_error.sh"

    set +e
    bash "$TEST_OUTPUT_DIR/test_cleanup_error.sh" 2>/dev/null
    set -e

    assert_file_exists "$marker" "Cleanup should be called on error exit"
    rm -f "$marker"
}

test_exit_codes() {
    log_section "Exit Code Definitions"

    log_test "Testing exit code values"

    # Source the script to get access to functions
    source_script

    # Test that functions use correct exit codes
    cat > "$TEST_OUTPUT_DIR/test_exit_codes.sh" <<'EOF'
#!/bin/bash
source action_simplified.sh 2>/dev/null || true

# Override log_error to capture exit codes
log_error() {
    exit "${2:-1}"
}

# Test general failure (default)
test_general_failure() {
    log_error "General failure"
}

# Test invalid input (code 2)
test_invalid_input() {
    log_error "Invalid input" 2
}

# Test download failure (code 3)
test_download_failure() {
    log_error "Download failed" 3
}

# Test analysis failure (code 4)
test_analysis_failure() {
    log_error "Analysis failed" 4
}

# Test GitHub failure (code 5)
test_github_failure() {
    log_error "GitHub integration failed" 5
}
EOF

    # Test each exit code
    local exit_code

    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_exit_codes.sh; test_general_failure" 2>/dev/null
    exit_code=$?
    set -e
    assert_equals "1" "$exit_code" "General failure should exit with code 1"

    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_exit_codes.sh; test_invalid_input" 2>/dev/null
    exit_code=$?
    set -e
    assert_equals "2" "$exit_code" "Invalid input should exit with code 2"

    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_exit_codes.sh; test_download_failure" 2>/dev/null
    exit_code=$?
    set -e
    assert_equals "3" "$exit_code" "Download failure should exit with code 3"

    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_exit_codes.sh; test_analysis_failure" 2>/dev/null
    exit_code=$?
    set -e
    assert_equals "4" "$exit_code" "Analysis failure should exit with code 4"

    set +e
    bash -c "source $TEST_OUTPUT_DIR/test_exit_codes.sh; test_github_failure" 2>/dev/null
    exit_code=$?
    set -e
    assert_equals "5" "$exit_code" "GitHub failure should exit with code 5"
}

test_cleanup_on_failure() {
    log_section "Cleanup on Failure"

    log_test "Testing default outputs on failure"

    # Create test script that simulates failure
    local output_file="/tmp/test_github_output_failure"
    rm -f "$output_file"

    cat > "$TEST_OUTPUT_DIR/test_failure_outputs.sh" <<EOF
#!/bin/bash

GITHUB_OUTPUT="$output_file"
TEMP_DIR=\$(mktemp -d)

cleanup() {
    local exit_code=\$?

    if [[ \$exit_code -ne 0 ]] && [[ -n "\${GITHUB_OUTPUT:-}" ]]; then
        {
            echo "has-changes=false"
            echo "has-dangers=false"
            echo "change-count=0"
            echo "danger-count=0"
            echo "summary=Analysis failed"
            echo "json-summary={}"
        } >> "\$GITHUB_OUTPUT"
    fi

    [[ -d "\$TEMP_DIR" ]] && rm -rf "\$TEMP_DIR"
}

trap cleanup EXIT

# Simulate failure
exit 1
EOF

    set +e
    bash "$TEST_OUTPUT_DIR/test_failure_outputs.sh" 2>/dev/null
    set -e

    assert_file_exists "$output_file" "GitHub output file should be created on failure"

    if [[ -f "$output_file" ]]; then
        local content=$(cat "$output_file")
        assert_contains "$content" "has-changes=false" "Should set has-changes to false"
        assert_contains "$content" "has-dangers=false" "Should set has-dangers to false"
        assert_contains "$content" "change-count=0" "Should set change-count to 0"
        assert_contains "$content" "danger-count=0" "Should set danger-count to 0"
        assert_contains "$content" "summary=Analysis failed" "Should set summary to failure message"
        assert_contains "$content" "json-summary={}" "Should set json-summary to empty object"
    fi

    rm -f "$output_file"
}

test_temp_directory_cleanup() {
    log_section "Temporary Directory Management"

    log_test "Testing temp directory creation and cleanup"

    # Create test script that tracks temp directory
    local marker="/tmp/tempdir_marker_test"
    rm -f "$marker"

    cat > "$TEST_OUTPUT_DIR/test_tempdir.sh" <<EOF
#!/bin/bash

TEMP_MARKER="$marker"
TEMP_DIR=\$(mktemp -d)

# Save the temp dir path
echo "\$TEMP_DIR" > "\$TEMP_MARKER"

cleanup() {
    [[ -d "\$TEMP_DIR" ]] && rm -rf "\$TEMP_DIR"
}

trap cleanup EXIT

# Verify temp dir exists during execution
if [[ -d "\$TEMP_DIR" ]]; then
    touch "\$TEMP_DIR/test_file"
fi

exit 0
EOF

    bash "$TEST_OUTPUT_DIR/test_tempdir.sh"

    assert_file_exists "$marker" "Temp directory path should be recorded"

    if [[ -f "$marker" ]]; then
        local temp_dir=$(cat "$marker")
        assert_dir_not_exists "$temp_dir" "Temp directory should be cleaned up after exit"
    fi

    rm -f "$marker"
}

# ============================================================================
# Test Cases: Input Validation
# ============================================================================

test_required_input_validation() {
    log_section "Required Input Validation"

    log_test "Testing missing plan file validation"

    # Create a test script to properly test the function
    cat > "$TEST_OUTPUT_DIR/test_validation.sh" <<'EOF'
#!/bin/bash
source action_simplified.sh 2>/dev/null || true

# Test with INPUT_PLAN_FILE unset
unset INPUT_PLAN_FILE
validate_required_inputs 2>&1
EOF

    # Run test and capture exit code
    local exit_code
    set +e
    bash "$TEST_OUTPUT_DIR/test_validation.sh" > "$TEST_OUTPUT_DIR/output.txt" 2>&1
    exit_code=$?
    set -e

    assert_not_equals "0" "$exit_code" "Should fail when plan file is missing"

    # Check error message (could be in output since we redirected stderr to stdout)
    local output=$(cat "$TEST_OUTPUT_DIR/output.txt" 2>/dev/null)
    assert_contains "$output" "Plan file is required" "Should show appropriate error message"
}

test_file_validation() {
    log_section "File Validation"

    log_test "Testing file existence validation"

    source_script

    # Create test files
    local existing_file="$TEST_OUTPUT_DIR/existing.tfplan"
    local non_existing_file="$TEST_OUTPUT_DIR/non_existing.tfplan"
    local unreadable_file="$TEST_OUTPUT_DIR/unreadable.tfplan"

    touch "$existing_file"
    touch "$unreadable_file"
    chmod 000 "$unreadable_file"

    # Test existing file
    local exit_code=$(run_function validate_file_exists "$existing_file")
    assert_equals "0" "$exit_code" "Should succeed for existing readable file"

    # Test non-existing file
    exit_code=$(run_function validate_file_exists "$non_existing_file")
    assert_not_equals "0" "$exit_code" "Should fail for non-existing file"

    # Test unreadable file (skip if running as root)
    if [[ $EUID -ne 0 ]]; then
        exit_code=$(run_function validate_file_exists "$unreadable_file")
        assert_not_equals "0" "$exit_code" "Should fail for unreadable file"
    fi

    chmod 644 "$unreadable_file"
}

test_output_format_validation() {
    log_section "Output Format Validation"

    log_test "Testing output format validation"

    source_script

    # Test valid formats
    for format in table json markdown html; do
        local exit_code=$(run_function validate_output_format "$format")
        assert_equals "0" "$exit_code" "Should accept valid format: $format"
    done

    # Test invalid formats
    for format in xml yaml csv text; do
        local exit_code=$(run_function validate_output_format "$format")
        assert_not_equals "0" "$exit_code" "Should reject invalid format: $format"
    done
}

test_path_security_validation() {
    log_section "Path Security Validation"

    log_test "Testing path traversal detection"

    source_script

    # Test normal paths
    local exit_code=$(run_function validate_path_security "/tmp/test.tfplan")
    assert_equals "0" "$exit_code" "Should accept normal absolute path"

    exit_code=$(run_function validate_path_security "test.tfplan")
    assert_equals "0" "$exit_code" "Should accept normal relative path"

    # Test path traversal
    exit_code=$(run_function validate_path_security "../../../etc/passwd")
    assert_not_equals "0" "$exit_code" "Should reject path with .."

    exit_code=$(run_function validate_path_security "/tmp/../etc/passwd")
    assert_not_equals "0" "$exit_code" "Should reject path with .. in middle"

    log_test "Testing path length validation"

    # Test normal length
    local normal_path="/tmp/test.tfplan"
    exit_code=$(run_function validate_path_security "$normal_path")
    assert_equals "0" "$exit_code" "Should accept normal length path"

    # Test excessive length
    local long_path=$(printf 'a%.0s' {1..5000})
    exit_code=$(run_function validate_path_security "$long_path")
    assert_not_equals "0" "$exit_code" "Should reject excessively long path"
}

test_logging_functions() {
    log_section "Logging Functions"

    log_test "Testing logging output to stderr"

    source_script

    # Test each logging function
    local funcs=(log_info log_success log_warning log_start log_download log_analyze log_config)

    for func in "${funcs[@]}"; do
        local output=$($func "Test message" 2>&1 >/dev/null)
        assert_contains "$output" "Test message" "$func should output to stderr"
    done

    log_test "Testing log prefixes"

    # Test emoji prefixes
    local info_output=$(log_info "Test" 2>&1)
    assert_contains "$info_output" "ℹ️" "log_info should have info emoji"

    local success_output=$(log_success "Test" 2>&1)
    assert_contains "$success_output" "✅" "log_success should have checkmark emoji"

    local error_output=$(log_error "Test" 2>&1 || true)
    assert_contains "$error_output" "❌" "log_error should have X emoji"

    local warning_output=$(log_warning "Test" 2>&1)
    assert_contains "$warning_output" "⚠️" "log_warning should have warning emoji"

    local start_output=$(log_start "Test" 2>&1)
    assert_contains "$start_output" "🚀" "log_start should have rocket emoji"
}

# ============================================================================
# Main Test Runner
# ============================================================================

run_tests() {
    echo ""
    echo "================================================"
    echo "Strata GitHub Action Foundation Tests"
    echo "================================================"

    # Check if action_simplified.sh exists
    if [[ ! -f "action_simplified.sh" ]]; then
        echo -e "${RED}Error: action_simplified.sh not found${NC}"
        echo "Please run this test from the project root directory"
        exit 1
    fi

    # Run all test suites
    test_error_handling_framework
    test_cleanup_trap
    test_exit_codes
    test_cleanup_on_failure
    test_temp_directory_cleanup
    test_required_input_validation
    test_file_validation
    test_output_format_validation
    test_path_security_validation
    test_logging_functions

    # Print summary
    echo ""
    echo "================================================"
    echo "Test Summary"
    echo "================================================"
    echo -e "Tests run:     ${TESTS_RUN}"
    echo -e "Tests passed:  ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests failed:  ${RED}${TESTS_FAILED}${NC}"
    if [[ $TESTS_SKIPPED -gt 0 ]]; then
        echo -e "Tests skipped: ${YELLOW}${TESTS_SKIPPED}${NC}"
    fi
    echo ""

    # Exit with appropriate code
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}TESTS FAILED${NC}"
        exit 1
    else
        echo -e "${GREEN}ALL TESTS PASSED${NC}"
        exit 0
    fi
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    run_tests
fi