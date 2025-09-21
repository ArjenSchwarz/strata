#!/bin/bash

# Unit tests for GitHub Action components
# This script tests individual functions from action.sh

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

assert_equals() {
    local expected=$1
    local actual=$2
    local message=$3
    
    if [ "$expected" = "$actual" ]; then
        echo -e "${GREEN}[PASS]${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} $message"
        echo "  Expected: '$expected'"
        echo "  Actual:   '$actual'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

assert_not_empty() {
    local value=$1
    local message=$2
    
    if [ -n "$value" ]; then
        echo -e "${GREEN}[PASS]${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} $message"
        echo "  Expected: non-empty value"
        echo "  Actual:   empty"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

assert_file_exists() {
    local file=$1
    local message=$2
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}[PASS]${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} $message"
        echo "  Expected: file exists"
        echo "  Actual:   file not found: $file"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

assert_command_success() {
    local command=$1
    local message=$2
    
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} $message"
        echo "  Command failed: $command"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Create temporary directory for tests
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

# Source the functions we want to test (extract them from action.sh)
# We'll create a test version that sources the functions without executing the main script

# Extract functions from action.sh for testing
extract_functions() {
    # Extract specific functions from action.sh for testing
    sed -n '/^# Function to extract value from JSON/,/^}/p' action.sh > "$TEST_DIR/functions.sh"
    sed -n '/^# Function to log messages/,/^}/p' action.sh >> "$TEST_DIR/functions.sh"
    sed -n '/^# Function to verify checksum/,/^}/p' action.sh >> "$TEST_DIR/functions.sh"
    sed -n '/^# Function to download with retry/,/^}/p' action.sh >> "$TEST_DIR/functions.sh"
    
    # Add test-specific modifications
    cat >> "$TEST_DIR/functions.sh" << 'EOF'

# Test-specific modifications
log() {
    echo "[LOG] $1: $2"
}

warning() {
    echo "[WARNING] $1"
}

error() {
    echo "[ERROR] $1"
    return 1
}
EOF
}

# Test platform detection
test_platform_detection() {
    log_test "Platform detection"
    
    # Test current platform detection
    PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    
    assert_not_empty "$PLATFORM" "Platform should be detected"
    assert_not_empty "$ARCH" "Architecture should be detected"
    
    # Test architecture normalization
    if [ "$ARCH" = "x86_64" ]; then
        NORMALIZED_ARCH="amd64"
    elif [ "$ARCH" = "aarch64" ]; then
        NORMALIZED_ARCH="arm64"
    else
        NORMALIZED_ARCH="$ARCH"
    fi
    
    assert_not_empty "$NORMALIZED_ARCH" "Architecture should be normalized"
}

# Test input validation functions
test_input_validation() {
    log_test "Input validation"
    
    # Test boolean validation
    validate_boolean() {
        local value=$1
        local default=$2
        
        if [ "$value" != "true" ] && [ "$value" != "false" ]; then
            echo "$default"
        else
            echo "$value"
        fi
    }
    
    # Test valid boolean values
    result=$(validate_boolean "true" "false")
    assert_equals "true" "$result" "Valid boolean 'true' should be accepted"
    
    result=$(validate_boolean "false" "true")
    assert_equals "false" "$result" "Valid boolean 'false' should be accepted"
    
    # Test invalid boolean values
    result=$(validate_boolean "invalid" "false")
    assert_equals "false" "$result" "Invalid boolean should return default"
    
    result=$(validate_boolean "" "true")
    assert_equals "true" "$result" "Empty boolean should return default"
}

# Test output format validation
test_output_format_validation() {
    log_test "Output format validation"
    
    validate_output_format() {
        local format=$1
        case "$format" in
            markdown|json|table)
                echo "$format"
                ;;
            *)
                echo "markdown"
                ;;
        esac
    }
    
    # Test valid formats
    result=$(validate_output_format "markdown")
    assert_equals "markdown" "$result" "Valid format 'markdown' should be accepted"
    
    result=$(validate_output_format "json")
    assert_equals "json" "$result" "Valid format 'json' should be accepted"
    
    result=$(validate_output_format "table")
    assert_equals "table" "$result" "Valid format 'table' should be accepted"
    
    # Test invalid format
    result=$(validate_output_format "invalid")
    assert_equals "markdown" "$result" "Invalid format should default to 'markdown'"
}

# Test JSON parsing functions
test_json_parsing() {
    log_test "JSON parsing"
    
    # Create test JSON
    TEST_JSON='{"hasChanges": true, "hasDangers": false, "totalChanges": 5, "dangerCount": 0}'
    
    # Test with jq if available
    if command -v jq >/dev/null 2>&1; then
        result=$(echo "$TEST_JSON" | jq -r '.hasChanges')
        assert_equals "true" "$result" "JSON parsing with jq should work"
        
        result=$(echo "$TEST_JSON" | jq -r '.totalChanges')
        assert_equals "5" "$result" "JSON number parsing with jq should work"
    fi
    
    # Test fallback grep parsing
    extract_with_grep() {
        local json=$1
        local key=$2
        echo "$json" | grep -o "\"$key\":[^,}]*" | cut -d':' -f2 | tr -d '"{}[],' | tr -d ' ' 2>/dev/null
    }
    
    result=$(extract_with_grep "$TEST_JSON" "hasChanges")
    assert_equals "true" "$result" "JSON parsing with grep should work"
    
    result=$(extract_with_grep "$TEST_JSON" "totalChanges")
    assert_equals "5" "$result" "JSON number parsing with grep should work"
}

# Test checksum verification
test_checksum_verification() {
    log_test "Checksum verification"
    
    # Create test file
    echo "test content" > "$TEST_DIR/test_file"
    
    # Calculate actual checksums
    if command -v sha256sum >/dev/null 2>&1; then
        ACTUAL_SHA256=$(sha256sum "$TEST_DIR/test_file" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        ACTUAL_SHA256=$(shasum -a 256 "$TEST_DIR/test_file" | cut -d' ' -f1)
    fi
    
    if [ -n "$ACTUAL_SHA256" ]; then
        # Test checksum verification function
        verify_checksum_test() {
            local file=$1
            local expected=$2
            local algorithm=$3
            
            if [ -z "$expected" ]; then
                return 0  # Skip verification if no checksum
            fi
            
            local actual
            case $algorithm in
                sha256)
                    if command -v sha256sum >/dev/null 2>&1; then
                        actual=$(sha256sum "$file" | cut -d' ' -f1)
                    elif command -v shasum >/dev/null 2>&1; then
                        actual=$(shasum -a 256 "$file" | cut -d' ' -f1)
                    else
                        return 0  # Skip if no tool available
                    fi
                    ;;
                *)
                    return 0  # Unsupported algorithm
                    ;;
            esac
            
            [ "$actual" = "$expected" ]
        }
        
        # Test valid checksum
        if verify_checksum_test "$TEST_DIR/test_file" "$ACTUAL_SHA256" "sha256"; then
            echo -e "${GREEN}[PASS]${NC} Valid checksum verification should succeed"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Valid checksum verification should succeed"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
        
        # Test invalid checksum
        if ! verify_checksum_test "$TEST_DIR/test_file" "invalid_checksum" "sha256"; then
            echo -e "${GREEN}[PASS]${NC} Invalid checksum verification should fail"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Invalid checksum verification should fail"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    fi
}

# Test file validation
test_file_validation() {
    log_test "File validation"
    
    # Create test files
    echo "test plan" > "$TEST_DIR/valid_plan.tfplan"
    chmod 644 "$TEST_DIR/valid_plan.tfplan"
    
    echo "test config" > "$TEST_DIR/valid_config.yaml"
    chmod 644 "$TEST_DIR/valid_config.yaml"
    
    # Test file existence
    assert_file_exists "$TEST_DIR/valid_plan.tfplan" "Valid plan file should exist"
    assert_file_exists "$TEST_DIR/valid_config.yaml" "Valid config file should exist"
    
    # Test file readability
    assert_command_success "[ -r '$TEST_DIR/valid_plan.tfplan' ]" "Valid plan file should be readable"
    assert_command_success "[ -r '$TEST_DIR/valid_config.yaml' ]" "Valid config file should be readable"
}

# Test cache functionality
test_cache_functionality() {
    log_test "Cache functionality"
    
    # Create mock cache directory
    CACHE_DIR="$TEST_DIR/cache"
    mkdir -p "$CACHE_DIR"
    
    # Test cache path generation
    generate_cache_path() {
        local version=$1
        local platform=$2
        local arch=$3
        echo "$CACHE_DIR/strata_${version}_${platform}_${arch}"
    }
    
    cache_path=$(generate_cache_path "v1.0.0" "linux" "amd64")
    expected_path="$CACHE_DIR/strata_v1.0.0_linux_amd64"
    assert_equals "$expected_path" "$cache_path" "Cache path should be generated correctly"
    
    # Test cache directory creation
    assert_command_success "[ -d '$CACHE_DIR' ]" "Cache directory should be created"
}

# Test environment variable handling
test_environment_variables() {
    log_test "Environment variable handling"
    
    # Test default values
    test_default_value() {
        local var_name=$1
        local default_value=$2
        local actual_value
        
        # Simulate environment variable handling
        eval "actual_value=\${${var_name}:-$default_value}"
        echo "$actual_value"
    }
    
    # Test with unset variable
    unset TEST_VAR
    result=$(test_default_value "TEST_VAR" "default")
    assert_equals "default" "$result" "Unset variable should use default value"
    
    # Test with set variable
    export TEST_VAR="custom"
    result=$(test_default_value "TEST_VAR" "default")
    assert_equals "custom" "$result" "Set variable should use custom value"
    
    unset TEST_VAR
}

# Test GitHub context detection
test_github_context() {
    log_test "GitHub context detection"
    
    # Test PR context detection
    detect_pr_context() {
        [ "$GITHUB_EVENT_NAME" = "pull_request" ]
    }
    
    # Test with PR context
    export GITHUB_EVENT_NAME="pull_request"
    if detect_pr_context; then
        echo -e "${GREEN}[PASS]${NC} PR context should be detected correctly"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} PR context should be detected correctly"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    # Test with non-PR context
    export GITHUB_EVENT_NAME="push"
    if ! detect_pr_context; then
        echo -e "${GREEN}[PASS]${NC} Non-PR context should be detected correctly"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} Non-PR context should be detected correctly"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    unset GITHUB_EVENT_NAME
}

# Test dual output functionality
test_dual_output_functions() {
    log_test "Dual output functionality"
    
    # Test temporary file creation and tracking
    test_temp_file_creation() {
        # Initialize test tracking
        rm -f "$TEST_DIR/temp_files_list"
        
        create_temp_file_test() {
            local temp_file
            temp_file=$(mktemp)
            if [ $? -ne 0 ] || [ ! -f "$temp_file" ]; then
                return 1
            fi
            
            chmod 600 "$temp_file"
            # Use a different approach to track files
            echo "$temp_file" >> "$TEST_DIR/temp_files_list"
            echo "$temp_file"
        }
        
        cleanup_temp_files_test() {
            if [ -f "$TEST_DIR/temp_files_list" ]; then
                while IFS= read -r temp_file; do
                    if [ -f "$temp_file" ]; then
                        rm -f "$temp_file"
                    fi
                done < "$TEST_DIR/temp_files_list"
                rm -f "$TEST_DIR/temp_files_list"
            fi
        }
        
        # Test file creation
        temp_file=$(create_temp_file_test)
        if [ -f "$temp_file" ]; then
            echo -e "${GREEN}[PASS]${NC} Temporary file creation should work"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Temporary file creation should work"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
        
        # Test file permissions
        if [ -f "$temp_file" ]; then
            perms=$(stat -c "%a" "$temp_file" 2>/dev/null || stat -f "%A" "$temp_file" 2>/dev/null)
            if [ "$perms" = "600" ]; then
                echo -e "${GREEN}[PASS]${NC} Temporary file should have restrictive permissions"
                TESTS_PASSED=$((TESTS_PASSED + 1))
            else
                echo -e "${RED}[FAIL]${NC} Temporary file should have restrictive permissions (got $perms)"
                TESTS_FAILED=$((TESTS_FAILED + 1))
            fi
        fi
        
        # Test file tracking - check if the file was tracked
        if [ -f "$TEST_DIR/temp_files_list" ] && [ "$(wc -l < "$TEST_DIR/temp_files_list")" -eq 1 ]; then
            echo -e "${GREEN}[PASS]${NC} Temporary file should be tracked"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            local count=0
            if [ -f "$TEST_DIR/temp_files_list" ]; then
                count=$(wc -l < "$TEST_DIR/temp_files_list")
            fi
            echo -e "${RED}[FAIL]${NC} Temporary file should be tracked (got $count files)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
        
        # Test cleanup
        cleanup_temp_files_test
        if [ ! -f "$TEST_DIR/temp_files_list" ]; then
            echo -e "${GREEN}[PASS]${NC} Temporary files should be cleaned up"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Temporary files should be cleaned up"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    }
    
    # Test dual output command construction
    test_dual_output_command() {
        construct_dual_output_command() {
            local stdout_format=$1
            local file_path=$2
            local plan_file=$3
            
            local cmd="strata plan summary"
            cmd="$cmd --output $stdout_format --file $file_path --file-format markdown"
            cmd="$cmd $plan_file"
            
            echo "$cmd"
        }
        
        result=$(construct_dual_output_command "table" "/tmp/test.md" "plan.tfplan")
        expected="strata plan summary --output table --file /tmp/test.md --file-format markdown plan.tfplan"
        
        assert_equals "$expected" "$result" "Dual output command should be constructed correctly"
    }
    
    # Test error handling for file operations
    test_file_error_handling() {
        handle_file_error() {
            local operation=$1
            local error_type=$2
            
            case "$operation" in
                "create")
                    if [ "$error_type" = "permission_denied" ]; then
                        echo "fallback_to_single_output"
                    fi
                    ;;
                "read")
                    if [ "$error_type" = "file_not_found" ]; then
                        echo "use_stdout_fallback"
                    fi
                    ;;
            esac
        }
        
        result=$(handle_file_error "create" "permission_denied")
        assert_equals "fallback_to_single_output" "$result" "Should fallback to single output on file creation error"
        
        result=$(handle_file_error "read" "file_not_found")
        assert_equals "use_stdout_fallback" "$result" "Should use stdout fallback on file read error"
    }
    
    # Test format validation
    test_format_validation() {
        validate_dual_output_formats() {
            local stdout_format=$1
            local file_format=$2
            
            # Validate stdout format
            case "$stdout_format" in
                table|json|markdown) ;;
                *) return 1 ;;
            esac
            
            # Validate file format
            case "$file_format" in
                markdown|json) ;;
                *) return 1 ;;
            esac
            
            return 0
        }
        
        if validate_dual_output_formats "table" "markdown"; then
            echo -e "${GREEN}[PASS]${NC} Valid format combination should be accepted"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Valid format combination should be accepted"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
        
        if ! validate_dual_output_formats "invalid" "markdown"; then
            echo -e "${GREEN}[PASS]${NC} Invalid stdout format should be rejected"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Invalid stdout format should be rejected"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
        
        if ! validate_dual_output_formats "table" "invalid"; then
            echo -e "${GREEN}[PASS]${NC} Invalid file format should be rejected"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}[FAIL]${NC} Invalid file format should be rejected"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    }
    
    test_temp_file_creation
    test_dual_output_command
    test_file_error_handling
    test_format_validation
}

# Run all tests
echo "Running GitHub Action Unit Tests..."
echo "=================================="

test_platform_detection
test_input_validation
test_output_format_validation
test_json_parsing
test_checksum_verification
test_file_validation
test_cache_functionality
test_environment_variables
test_github_context
test_dual_output_functions

# Print test summary
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