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

# Test run_analysis argument handling for safety and spaces
test_run_analysis_argument_safety() {
    log_test "run_analysis argument safety"

    local run_dir="$TEST_DIR/run_analysis_safety"
    local temp_dir="$run_dir/temp"
    local harness="$run_dir/run_analysis_harness.sh"
    mkdir -p "$temp_dir"

    cat > "$harness" <<'EOF'
#!/bin/bash
set -euo pipefail

ACTION_PATH="$1"
TEST_TEMP_DIR="$2"
PLAN_INPUT="$3"
CONFIG_INPUT="$4"
EXPECTED_PLAN="$5"
EXPECTED_CONFIG="$6"

cat > "$TEST_TEMP_DIR/strata" <<'MOCK'
#!/bin/bash
set -euo pipefail

json_file=""
config_file=""
plan_file=""
saw_separator="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    plan|summary) shift ;;
    --output) shift 2 ;;
    --file) json_file="$2"; shift 2 ;;
    --file-format) shift 2 ;;
    --details|--details=*|--expand-all|--expand-all=*) shift ;;
    --highlight-dangers|--highlight-dangers=*) shift ;;
    --config) config_file="$2"; shift 2 ;;
    --) saw_separator="true"; shift; plan_file="$1"; shift ;;
    *) echo "unexpected-arg:$1" >&2; exit 99 ;;
  esac
done

[[ "$saw_separator" == "true" ]] || { echo "missing -- separator" >&2; exit 98; }
[[ "$plan_file" == "$EXPECTED_PLAN" ]] || { echo "plan mismatch: $plan_file" >&2; exit 97; }
if [[ -n "$EXPECTED_CONFIG" ]]; then
  [[ "$config_file" == "$EXPECTED_CONFIG" ]] || { echo "config mismatch: $config_file" >&2; exit 96; }
fi

cat > "$json_file" <<'JSON'
{"statistics":{"total_changes":1,"dangerous_changes":0}}
JSON
echo "ok"
MOCK
chmod +x "$TEST_TEMP_DIR/strata"

export EXPECTED_PLAN EXPECTED_CONFIG
TEMP_DIR="$TEST_TEMP_DIR"
STRATA_BIN="$TEST_TEMP_DIR/strata"
INPUT_PLAN_FILE="$PLAN_INPUT"
INPUT_OUTPUT_FORMAT="markdown"
INPUT_SHOW_DETAILS="false"
INPUT_EXPAND_ALL="false"
INPUT_CONFIG_FILE="$CONFIG_INPUT"

extract_outputs() { :; }
set_default_outputs() { :; }

eval "$(sed -n '/^run_analysis()/,/^}/p' "$ACTION_PATH")"
run_analysis
EOF
    chmod +x "$harness"

    local plan_with_spaces="$run_dir/plan with spaces.tfplan"
    local config_with_spaces="$run_dir/config with spaces.yaml"
    mkdir -p "$run_dir"
    echo "plan" > "$plan_with_spaces"
    echo "config" > "$config_with_spaces"

    if bash "$harness" "action.sh" "$temp_dir" "$plan_with_spaces" "$config_with_spaces" "$plan_with_spaces" "$config_with_spaces" >/dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC} run_analysis should handle plan/config paths containing spaces"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} run_analysis should handle plan/config paths containing spaces"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    local dash_prefixed_plan="$run_dir/-dash-prefixed.tfplan"
    echo "plan" > "$dash_prefixed_plan"

    if bash "$harness" "action.sh" "$temp_dir" "$dash_prefixed_plan" "" "$dash_prefixed_plan" "" >/dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC} run_analysis should handle plan paths that start with '-'"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} run_analysis should handle plan paths that start with '-'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Regression test for T-1202: the documented `highlight-dangers` action input
# must reach strata. run_analysis must forward --highlight-dangers=true|false
# based on INPUT_HIGHLIGHT_DANGERS, defaulting to true when the input is unset
# so danger highlighting stays on by default.
test_run_analysis_highlight_dangers_wiring() {
    log_test "run_analysis highlight-dangers wiring (T-1202)"

    local run_dir="$TEST_DIR/run_analysis_highlight_dangers"
    local temp_dir="$run_dir/temp"
    local harness="$run_dir/highlight_dangers_harness.sh"
    mkdir -p "$temp_dir"

    # The harness sources run_analysis from action.sh with a mock strata that
    # records the --highlight-dangers value it received, then asserts it matches
    # the expected value derived from INPUT_HIGHLIGHT_DANGERS.
    cat > "$harness" <<'EOF'
#!/bin/bash
set -euo pipefail

ACTION_PATH="$1"
TEST_TEMP_DIR="$2"
INPUT_VALUE="$3"
EXPECTED_FLAG="$4"

cat > "$TEST_TEMP_DIR/strata" <<'MOCK'
#!/bin/bash
set -euo pipefail

json_file=""
highlight_flag="<unset>"

while [[ $# -gt 0 ]]; do
  case "$1" in
    plan|summary) shift ;;
    --output) shift 2 ;;
    --file) json_file="$2"; shift 2 ;;
    --file-format) shift 2 ;;
    --details|--details=*|--expand-all|--expand-all=*) shift ;;
    --highlight-dangers=*) highlight_flag="${1#--highlight-dangers=}"; shift ;;
    --highlight-dangers) highlight_flag="$2"; shift 2 ;;
    --config) shift 2 ;;
    --) shift; shift ;;
    *) echo "unexpected-arg:$1" >&2; exit 99 ;;
  esac
done

[[ "$highlight_flag" == "$EXPECTED_FLAG" ]] || {
  echo "highlight-dangers mismatch: got '$highlight_flag', expected '$EXPECTED_FLAG'" >&2
  exit 95
}

cat > "$json_file" <<'JSON'
{"statistics":{"total_changes":1,"dangerous_changes":0}}
JSON
echo "ok"
MOCK
chmod +x "$TEST_TEMP_DIR/strata"

export EXPECTED_FLAG
TEMP_DIR="$TEST_TEMP_DIR"
STRATA_BIN="$TEST_TEMP_DIR/strata"
INPUT_PLAN_FILE="$TEST_TEMP_DIR/plan.tfplan"
echo "plan" > "$INPUT_PLAN_FILE"
INPUT_OUTPUT_FORMAT="markdown"
INPUT_SHOW_DETAILS="false"
INPUT_EXPAND_ALL="false"
INPUT_CONFIG_FILE=""
if [[ "$INPUT_VALUE" != "<unset>" ]]; then
  INPUT_HIGHLIGHT_DANGERS="$INPUT_VALUE"
fi

extract_outputs() { :; }
set_default_outputs() { :; }

eval "$(sed -n '/^run_analysis()/,/^}/p' "$ACTION_PATH")"
run_analysis
EOF
    chmod +x "$harness"

    # Default (input unset) must keep highlighting on.
    if bash "$harness" "action.sh" "$temp_dir" "<unset>" "true" >/dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC} run_analysis defaults highlight-dangers to true"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} run_analysis defaults highlight-dangers to true"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Explicit true must forward --highlight-dangers=true.
    if bash "$harness" "action.sh" "$temp_dir" "true" "true" >/dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC} run_analysis forwards highlight-dangers=true"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} run_analysis forwards highlight-dangers=true"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Explicit false must forward --highlight-dangers=false so danger
    # highlighting can be disabled via the documented input.
    if bash "$harness" "action.sh" "$temp_dir" "false" "false" >/dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC} run_analysis forwards highlight-dangers=false"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} run_analysis forwards highlight-dangers=false"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Regression test for T-1096: a "latest" run must not reuse a stale cached
# binary. download_strata must resolve the current latest release tag and only
# reuse the cached binary when its --version matches that tag. If the cached
# version is older, the binary must be re-downloaded rather than silently used.
test_latest_cache_version_validation() {
    log_test "latest cache validation (T-1096)"

    local run_dir="$TEST_DIR/latest_cache_validation"
    local harness="$run_dir/latest_cache_harness.sh"
    mkdir -p "$run_dir"

    # The harness sources download_strata + get_version_tag from action.sh,
    # then overrides get_version_tag, detect_platform and curl so the network is
    # never touched. CACHED_VERSION and LATEST_TAG are supplied per scenario.
    cat > "$harness" <<'EOF'
#!/bin/bash
set -uo pipefail

ACTION_PATH="$1"
CACHE_DIR="$2"
CACHED_VERSION="$3"
LATEST_TAG="$4"

mkdir -p "$CACHE_DIR"
STRATA_BIN="$CACHE_DIR/strata"

# Mock cached binary reporting CACHED_VERSION.
cat > "$STRATA_BIN" <<MOCK
#!/bin/bash
echo "strata version $CACHED_VERSION"
MOCK
chmod +x "$STRATA_BIN"

INPUT_STRATA_VERSION="latest"
GITHUB_API_URL="https://example.invalid"

# Override platform detection to avoid depending on the host.
detect_platform() { OS="linux"; ARCH="amd64"; }

# Override version resolution to return the controlled latest tag.
get_version_tag() {
  if [[ -n "$LATEST_TAG" ]]; then
    echo "$LATEST_TAG"
    return 0
  fi
  return 1
}

# Stub curl so any re-download attempt fails fast (no network).
curl() { return 1; }

# Stub sleep so retry loops do not slow the test down.
sleep() { :; }

export OS ARCH STRATA_BIN CACHE_DIR INPUT_STRATA_VERSION GITHUB_API_URL

eval "$(sed -n '/^download_strata()/,/^}/p' "$ACTION_PATH")"

download_strata
EOF
    chmod +x "$harness"

    # Scenario 1: cached binary is stale relative to the resolved latest tag.
    # The function must NOT reuse it. With curl stubbed to fail, a re-download
    # attempt means the function exits non-zero rather than returning 0 with
    # "Using cached Strata binary".
    local stale_cache="$run_dir/stale"
    local stale_output
    # download_strata exits non-zero once the (stubbed) re-download fails; that
    # non-zero exit is itself proof the stale cache was rejected. Guard the
    # assignment so the surrounding `set -e` does not abort the test run.
    stale_output=$(bash "$harness" "action.sh" "$stale_cache" "v0.0.1" "v1.5.0" 2>&1) || true
    if echo "$stale_output" | grep -q "Using cached Strata binary"; then
        echo -e "${RED}[FAIL]${NC} latest run must not reuse a stale cached binary"
        echo "  Output: $stale_output"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo -e "${GREEN}[PASS]${NC} latest run must not reuse a stale cached binary"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi

    # Scenario 2: cached binary matches the resolved latest tag. The function
    # must reuse it without attempting a download.
    local match_cache="$run_dir/match"
    if bash "$harness" "action.sh" "$match_cache" "v1.5.0" "v1.5.0" 2>&1 | grep -q "Using cached Strata binary"; then
        echo -e "${GREEN}[PASS]${NC} latest run reuses cached binary matching the latest tag"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} latest run reuses cached binary matching the latest tag"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Scenario 3: latest tag cannot be resolved (offline / API failure). The
    # function should fall back to reusing the cached binary as a best effort
    # rather than failing the whole action.
    local offline_cache="$run_dir/offline"
    if bash "$harness" "action.sh" "$offline_cache" "v0.0.1" "" 2>&1 | grep -qi "cached Strata binary"; then
        echo -e "${GREEN}[PASS]${NC} latest run reuses cached binary when latest tag cannot be resolved"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} latest run reuses cached binary when latest tag cannot be resolved"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Regression test for T-1118: extract_outputs must read the go-output document
# that `strata plan summary --file ... --file-format json` actually produces.
# That file is a rendered table-document ARRAY (a "Summary Statistics" section),
# NOT the internal PlanSummary schema with a top-level `.statistics` object.
# The old code parsed `.statistics.total`, which fails on an array and silently
# falls back to 0 — so a plan WITH changes wrongly reported change-count=0.
test_extract_outputs_go_output_array() {
    log_test "extract_outputs parses go-output document (T-1118)"

    local run_dir="$TEST_DIR/extract_outputs_go_output"
    local harness="$run_dir/harness.sh"
    mkdir -p "$run_dir"

    # The harness sources the real extract_outputs + set_default_outputs from
    # action.sh, then runs extract_outputs against a JSON file matching the
    # actual go-output shape passed as $2. It echoes the resulting GITHUB_OUTPUT.
    cat > "$harness" <<'EOF'
#!/bin/bash
set -uo pipefail

ACTION_PATH="$1"
JSON_FILE="$2"

GITHUB_OUTPUT="$(mktemp)"
export GITHUB_OUTPUT
DISPLAY_OUTPUT="display"
OUTPUTS_WRITTEN=false

eval "$(sed -n '/^set_default_outputs()/,/^}/p' "$ACTION_PATH")"
eval "$(sed -n '/^extract_outputs()/,/^}/p' "$ACTION_PATH")"

extract_outputs "$JSON_FILE"
grep -E '^(has-changes|has-dangers|change-count|danger-count)=' "$GITHUB_OUTPUT"
rm -f "$GITHUB_OUTPUT"
EOF
    chmod +x "$harness"

    # Case 1: plan WITH changes, horizontal (default) statistics layout.
    # This is the real go-output array shape; old code reported 0.
    local horizontal_json="$run_dir/horizontal.json"
    cat > "$horizontal_json" <<'EOF'
[
  {
    "title": "Plan Information",
    "data": [{"Plan File": "samples/web-sample.json", "Version": "1.6.2"}]
  },
  {
    "title": "Summary Statistics",
    "data": [{"Total Changes": 7, "Added": 2, "Removed": 1, "Modified": 2, "Replacements": 2, "High Risk": 2, "Unmodified": 0}]
  },
  {
    "title": "Resource Changes",
    "data": []
  }
]
EOF
    local result
    result=$(bash "$harness" "action.sh" "$horizontal_json")
    assert_equals "true" "$(echo "$result" | grep '^has-changes=' | cut -d= -f2)" "horizontal: has-changes=true for changed plan"
    assert_equals "7" "$(echo "$result" | grep '^change-count=' | cut -d= -f2)" "horizontal: change-count=7"
    assert_equals "true" "$(echo "$result" | grep '^has-dangers=' | cut -d= -f2)" "horizontal: has-dangers=true"
    assert_equals "2" "$(echo "$result" | grep '^danger-count=' | cut -d= -f2)" "horizontal: danger-count=2"

    # Case 2: vertical statistics layout ({Metric, Value} rows).
    local vertical_json="$run_dir/vertical.json"
    cat > "$vertical_json" <<'EOF'
[
  {
    "title": "Summary Statistics",
    "data": [
      {"Metric": "Total Changes", "Value": 4},
      {"Metric": "Added", "Value": 1},
      {"Metric": "High Risk", "Value": 3},
      {"Metric": "Unmodified", "Value": 0}
    ]
  }
]
EOF
    result=$(bash "$harness" "action.sh" "$vertical_json")
    assert_equals "4" "$(echo "$result" | grep '^change-count=' | cut -d= -f2)" "vertical: change-count=4"
    assert_equals "3" "$(echo "$result" | grep '^danger-count=' | cut -d= -f2)" "vertical: danger-count=3"

    # Case 3: no-changes document is a single text object, not an array.
    local nochange_json="$run_dir/nochange.json"
    cat > "$nochange_json" <<'EOF'
{"content": "No changes detected", "type": "text"}
EOF
    result=$(bash "$harness" "action.sh" "$nochange_json")
    assert_equals "false" "$(echo "$result" | grep '^has-changes=' | cut -d= -f2)" "no-changes: has-changes=false"
    assert_equals "0" "$(echo "$result" | grep '^change-count=' | cut -d= -f2)" "no-changes: change-count=0"
    assert_equals "0" "$(echo "$result" | grep '^danger-count=' | cut -d= -f2)" "no-changes: danger-count=0"

    # Case 4: --details=true embeds multi-line, escaped "details" strings in the
    # Resource Changes section. extract_outputs must still find the statistics
    # counts and not choke on that content (it reads the file directly rather
    # than round-tripping the payload through a shell variable).
    local details_json="$run_dir/details.json"
    cat > "$details_json" <<'EOF'
[
  {
    "title": "Summary Statistics",
    "data": [{"Total Changes": 4, "High Risk": 2, "Added": 1, "Removed": 0, "Modified": 2, "Replacements": 1, "Unmodified": 0}]
  },
  {
    "title": "Resource Changes",
    "data": [
      {"Address": "aws_db_instance.main", "details": "  ~ tags {\n    ~ Environment = \"staging\" -> \"production\"\n  }\n  ~ username = \"a\" -> \"b\""}
    ]
  }
]
EOF
    result=$(bash "$harness" "action.sh" "$details_json")
    assert_equals "true" "$(echo "$result" | grep '^has-changes=' | cut -d= -f2)" "details: has-changes=true despite multi-line details"
    assert_equals "4" "$(echo "$result" | grep '^change-count=' | cut -d= -f2)" "details: change-count=4"
    assert_equals "2" "$(echo "$result" | grep '^danger-count=' | cut -d= -f2)" "details: danger-count=2"
}

# Regression test for T-1226: test scripts must start with a valid shebang.
# A corrupted shebang (e.g. "#\!/bin/bash" with an escaped bang) makes the
# kernel reject direct execution with "exec format error" on Linux, even
# though "bash <script>" still works and hides the problem.
test_script_shebangs() {
    log_test "test scripts have a valid shebang"

    local test_scripts_dir
    test_scripts_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

    local script
    local bad_scripts=()
    for script in "$test_scripts_dir"/*.sh; do
        [ -f "$script" ] || continue
        # First two bytes must be the "#!" magic. Reject anything else,
        # including the escaped-bang corruption "#\".
        local magic
        magic=$(head -c 2 "$script")
        if [ "$magic" != "#!" ]; then
            bad_scripts+=("$(basename "$script"): '$magic'")
        fi
    done

    if [ ${#bad_scripts[@]} -eq 0 ]; then
        echo -e "${GREEN}[PASS]${NC} all test scripts start with a valid '#!' shebang"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}[FAIL]${NC} test scripts with invalid shebang found:"
        printf '  %s\n' "${bad_scripts[@]}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
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
test_run_analysis_argument_safety
test_run_analysis_highlight_dangers_wiring
test_latest_cache_version_validation
test_extract_outputs_go_output_array
test_script_shebangs

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
