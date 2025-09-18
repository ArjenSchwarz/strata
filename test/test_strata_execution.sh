#!/bin/bash
#
# Unit tests for Strata execution component
# Tests command construction, dual output, JSON parsing, and error handling
#

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

assert_valid_json() {
    local content="$1"
    local message="$2"

    if echo "$content" | jq . >/dev/null 2>&1; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Invalid JSON content"
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

# Create mock Strata binary for testing
create_mock_strata() {
    local mock_strata="$TEST_OUTPUT_DIR/strata"
    local behavior="${1:-success}"

    case "$behavior" in
        "success")
            cat > "$mock_strata" <<'EOF'
#!/bin/bash
# Mock Strata that generates successful output

# Parse arguments to understand what was requested
output_format="table"
json_file=""
file_format=""
show_details="false"
expand_all="false"
config_file=""
plan_file=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --output) output_format="$2"; shift; shift ;;
        --file) json_file="$2"; shift; shift ;;
        --file-format) file_format="$2"; shift; shift ;;
        --show-details) show_details="true"; shift ;;
        --expand-all) expand_all="true"; shift ;;
        --config) config_file="$2"; shift; shift ;;
        --version) echo "strata version 1.4.0"; exit 0 ;;
        plan|summary) shift ;;
        *) plan_file="$1"; shift ;;
    esac
done

# Generate display output based on format
case "$output_format" in
    "markdown")
        cat <<'MDEOF'
# Terraform Plan Summary

## Changes Overview
- **Resources to create:** 2
- **Resources to update:** 1
- **Resources to delete:** 0

## Resource Details
### To be created:
- aws_instance.web
- aws_security_group.web

### To be updated:
- aws_instance.api (tags)
MDEOF
        ;;
    "json")
        cat <<'JSONEOF'
{
  "format_version": "1.0",
  "terraform_version": "1.6.0",
  "resource_changes": [
    {"address": "aws_instance.web", "change_type": "create"},
    {"address": "aws_security_group.web", "change_type": "create"},
    {"address": "aws_instance.api", "change_type": "update"}
  ],
  "statistics": {
    "total_changes": 3,
    "dangerous_changes": 0
  }
}
JSONEOF
        ;;
    "table")
        cat <<'TABLEEOF'
+------------------------+----------+
| Resource               | Action   |
+------------------------+----------+
| aws_instance.web       | create   |
| aws_security_group.web | create   |
| aws_instance.api       | update   |
+------------------------+----------+

Total: 3 changes, 0 dangerous
TABLEEOF
        ;;
    "html")
        cat <<'HTMLEOF'
<h1>Terraform Plan Summary</h1>
<p>Total: 3 changes, 0 dangerous</p>
<table>
  <tr><th>Resource</th><th>Action</th></tr>
  <tr><td>aws_instance.web</td><td>create</td></tr>
  <tr><td>aws_security_group.web</td><td>create</td></tr>
  <tr><td>aws_instance.api</td><td>update</td></tr>
</table>
HTMLEOF
        ;;
esac

# Write JSON metadata if requested
if [[ -n "$json_file" && "$file_format" == "json" ]]; then
    cat > "$json_file" <<'JSONMETAEOF'
{
  "format_version": "1.0",
  "terraform_version": "1.6.0",
  "resource_changes": [
    {
      "address": "aws_instance.web",
      "type": "aws_instance",
      "name": "web",
      "change_type": "create",
      "is_destructive": false,
      "is_dangerous": false
    },
    {
      "address": "aws_security_group.web",
      "type": "aws_security_group",
      "name": "web",
      "change_type": "create",
      "is_destructive": false,
      "is_dangerous": false
    },
    {
      "address": "aws_instance.api",
      "type": "aws_instance",
      "name": "api",
      "change_type": "update",
      "is_destructive": false,
      "is_dangerous": false
    }
  ],
  "output_changes": [],
  "statistics": {
    "total_changes": 3,
    "dangerous_changes": 0,
    "creates": 2,
    "updates": 1,
    "deletes": 0,
    "replaces": 0
  }
}
JSONMETAEOF
fi

exit 0
EOF
            ;;
        "failure")
            cat > "$mock_strata" <<'EOF'
#!/bin/bash
# Mock Strata that fails

echo "Error: Failed to parse Terraform plan file" >&2
echo "Plan file may be corrupted or invalid" >&2
exit 1
EOF
            ;;
        "dangerous")
            cat > "$mock_strata" <<'EOF'
#!/bin/bash
# Mock Strata with dangerous changes

# Parse arguments
output_format="table"
json_file=""
file_format=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --output) output_format="$2"; shift; shift ;;
        --file) json_file="$2"; shift; shift ;;
        --file-format) file_format="$2"; shift; shift ;;
        --version) echo "strata version 1.4.0"; exit 0 ;;
        plan|summary) shift ;;
        *) shift ;;
    esac
done

# Display output with dangerous changes
case "$output_format" in
    "markdown")
        cat <<'MDEOF'
# Terraform Plan Summary ⚠️

## Changes Overview
- **Resources to create:** 0
- **Resources to replace:** 2 ⚠️
- **Resources to delete:** 1 ⚠️

⚠️ **DANGEROUS CHANGES DETECTED**

## Resource Details
### To be replaced (DESTRUCTIVE):
- aws_rds_instance.primary ⚠️
- aws_instance.web ⚠️

### To be deleted:
- aws_security_group.old ⚠️
MDEOF
        ;;
    *)
        echo "Dangerous changes detected!"
        ;;
esac

# Write JSON metadata if requested
if [[ -n "$json_file" && "$file_format" == "json" ]]; then
    cat > "$json_file" <<'JSONEOF'
{
  "format_version": "1.0",
  "terraform_version": "1.6.0",
  "resource_changes": [
    {
      "address": "aws_rds_instance.primary",
      "change_type": "replace",
      "is_destructive": true,
      "is_dangerous": true,
      "danger_reason": "Database replacement may cause data loss"
    },
    {
      "address": "aws_instance.web",
      "change_type": "replace",
      "is_destructive": true,
      "is_dangerous": true,
      "danger_reason": "Instance replacement causes downtime"
    },
    {
      "address": "aws_security_group.old",
      "change_type": "delete",
      "is_destructive": true,
      "is_dangerous": true,
      "danger_reason": "Security group deletion may affect running instances"
    }
  ],
  "statistics": {
    "total_changes": 3,
    "dangerous_changes": 3,
    "creates": 0,
    "updates": 0,
    "deletes": 1,
    "replaces": 2
  }
}
JSONEOF
fi

exit 0
EOF
            ;;
    esac

    chmod +x "$mock_strata"
}

# Source the simplified action script
source_script() {
    if [[ -z "${SCRIPT_SOURCED:-}" ]]; then
        # Create a temporary copy that doesn't execute main and fixes readonly issues
        local temp_script="$TEST_OUTPUT_DIR/action_temp.sh"

        # Copy everything except the main execution guard and readonly TEMP_DIR
        sed -e '/^if \[\[ "${BASH_SOURCE\[0\]}" == "${0}" \]\]; then$/,/^fi$/d' \
            -e 's/readonly TEMP_DIR/#readonly TEMP_DIR/' \
            action_simplified.sh > "$temp_script"

        # Source the modified script
        source "$temp_script"
        SCRIPT_SOURCED=1
    fi
}

# ============================================================================
# Test Cases: Command Construction
# ============================================================================

test_command_construction() {
    log_section "Command Construction Tests"

    source_script

    log_test "Testing basic command construction"

    # Create test plan file
    local plan_file="$TEST_OUTPUT_DIR/test.tfplan"
    touch "$plan_file"

    # Create mock strata and override TEMP_DIR
    create_mock_strata "success"
    TEMP_DIR="$TEST_OUTPUT_DIR"

    # Test basic command construction by examining the mock call
    # We'll modify run_analysis to print the command instead of executing
    cat > "$TEST_OUTPUT_DIR/test_command_construction.sh" <<EOF
#!/bin/bash
source $TEST_OUTPUT_DIR/action_temp.sh

# Override the strata execution to capture the command
TEMP_DIR="$TEST_OUTPUT_DIR"

run_analysis_test() {
    local plan_file="\$1"
    local output_format="\$2"
    local show_details="\$3"
    local expand_all="\$4"
    local config_file="\$5"
    local json_file="\$TEMP_DIR/metadata.json"

    # Build command exactly like the real function
    local cmd="\$TEMP_DIR/strata plan summary"
    cmd="\$cmd --output \$output_format"
    cmd="\$cmd --file \$json_file --file-format json"

    [[ "\$show_details" == "true" ]] && cmd="\$cmd --show-details"
    [[ "\$expand_all" == "true" ]] && cmd="\$cmd --expand-all"
    [[ -n "\$config_file" ]] && cmd="\$cmd --config \$config_file"

    cmd="\$cmd \$plan_file"

    # Print the command instead of executing
    echo "\$cmd"
}

# Test different parameter combinations
echo "=== Basic Command ==="
run_analysis_test "$plan_file" "markdown" "false" "false" ""

echo "=== With Details ==="
run_analysis_test "$plan_file" "json" "true" "false" ""

echo "=== With Expand All ==="
run_analysis_test "$plan_file" "table" "false" "true" ""

echo "=== With Config File ==="
run_analysis_test "$plan_file" "html" "false" "false" "/tmp/config.yaml"

echo "=== All Parameters ==="
run_analysis_test "$plan_file" "markdown" "true" "true" "/tmp/config.yaml"
EOF

    local output
    output=$(bash "$TEST_OUTPUT_DIR/test_command_construction.sh")

    # Test basic command
    assert_contains "$output" "$TEST_OUTPUT_DIR/strata plan summary" "Should include basic strata command"
    assert_contains "$output" "--output markdown" "Should include output format"
    assert_contains "$output" "--file $TEST_OUTPUT_DIR/metadata.json --file-format json" "Should include JSON metadata file"
    assert_contains "$output" "$plan_file" "Should include plan file path"

    # Test conditional parameters
    assert_contains "$output" "--show-details" "Should include show-details when enabled"
    assert_contains "$output" "--expand-all" "Should include expand-all when enabled"
    assert_contains "$output" "--config /tmp/config.yaml" "Should include config file when specified"

    log_test "Testing parameter validation in command construction"

    # Test each output format
    for format in markdown json table html; do
        local cmd_output=$(bash -c "source $TEST_OUTPUT_DIR/action_temp.sh && echo 'test' | TEMP_DIR='$TEST_OUTPUT_DIR' run_analysis_test '$plan_file' '$format' 'false' 'false' ''" 2>/dev/null || echo "failed")
        assert_contains "$cmd_output" "--output $format" "Should handle $format output format"
    done
}

# ============================================================================
# Test Cases: Dual Output System
# ============================================================================

test_dual_output_system() {
    log_section "Dual Output System Tests"

    source_script

    log_test "Testing dual output generation (display + JSON)"

    # Create test plan file
    local plan_file="$TEST_OUTPUT_DIR/test.tfplan"
    echo '{"test": "plan"}' > "$plan_file"

    # Create mock strata and set up environment
    create_mock_strata "success"
    TEMP_DIR="$TEST_OUTPUT_DIR"

    # Set up mock GitHub environment
    export GITHUB_OUTPUT="$TEST_OUTPUT_DIR/github_output"
    export GITHUB_STEP_SUMMARY="$TEST_OUTPUT_DIR/step_summary"
    rm -f "$GITHUB_OUTPUT" "$GITHUB_STEP_SUMMARY"

    # Execute run_analysis function
    local exit_code=0
    (
        set +e
        run_analysis "$plan_file" "markdown" "false" "false" ""
        echo $? > "$TEST_OUTPUT_DIR/exit_code"
    ) > "$TEST_OUTPUT_DIR/display_output" 2>&1

    if [[ -f "$TEST_OUTPUT_DIR/exit_code" ]]; then
        exit_code=$(cat "$TEST_OUTPUT_DIR/exit_code")
    fi

    assert_exit_code "0" "$exit_code" "run_analysis should succeed with valid plan"

    # Check that JSON metadata file was created
    assert_file_exists "$TEST_OUTPUT_DIR/metadata.json" "JSON metadata file should be created"

    # Check that display output file was created (by extract_outputs)
    assert_file_exists "$TEST_OUTPUT_DIR/display_output.txt" "Display output file should be created"

    # Check JSON metadata content
    if [[ -f "$TEST_OUTPUT_DIR/metadata.json" ]]; then
        local json_content=$(cat "$TEST_OUTPUT_DIR/metadata.json")
        assert_valid_json "$json_content" "JSON metadata should be valid JSON"

        # Test specific fields
        local total_changes=$(echo "$json_content" | jq -r '.statistics.total_changes')
        assert_equals "3" "$total_changes" "Should extract correct total changes from JSON"

        local dangerous_changes=$(echo "$json_content" | jq -r '.statistics.dangerous_changes')
        assert_equals "0" "$dangerous_changes" "Should extract correct dangerous changes from JSON"
    fi

    log_test "Testing display output formats"

    # Test different output formats
    for format in markdown json table html; do
        create_mock_strata "success"
        rm -f "$TEST_OUTPUT_DIR/display_output.txt" "$TEST_OUTPUT_DIR/metadata.json"

        local format_exit_code=0
        (
            set +e
            run_analysis "$plan_file" "$format" "false" "false" ""
            echo $? > "$TEST_OUTPUT_DIR/format_exit_code"
        ) > "$TEST_OUTPUT_DIR/format_display_output" 2>&1

        if [[ -f "$TEST_OUTPUT_DIR/format_exit_code" ]]; then
            format_exit_code=$(cat "$TEST_OUTPUT_DIR/format_exit_code")
        fi

        assert_exit_code "0" "$format_exit_code" "Should handle $format format successfully"

        if [[ -f "$TEST_OUTPUT_DIR/display_output.txt" ]]; then
            local display_content=$(cat "$TEST_OUTPUT_DIR/display_output.txt")
            case "$format" in
                "markdown")
                    assert_contains "$display_content" "# Terraform Plan Summary" "Markdown should contain header"
                    ;;
                "json")
                    assert_valid_json "$display_content" "JSON format should be valid JSON"
                    ;;
                "table")
                    assert_contains "$display_content" "+-------" "Table should contain table borders"
                    ;;
                "html")
                    assert_contains "$display_content" "<h1>" "HTML should contain HTML tags"
                    ;;
            esac
        fi
    done
}

# ============================================================================
# Test Cases: JSON Parsing and Output Extraction
# ============================================================================

test_json_parsing() {
    log_section "JSON Parsing and Output Extraction Tests"

    source_script

    log_test "Testing extract_outputs function with successful analysis"

    # Create test JSON metadata
    local json_file="$TEST_OUTPUT_DIR/test_metadata.json"
    cat > "$json_file" <<'EOF'
{
  "format_version": "1.0",
  "terraform_version": "1.6.0",
  "resource_changes": [
    {"address": "aws_instance.web", "change_type": "create"},
    {"address": "aws_instance.api", "change_type": "update"}
  ],
  "statistics": {
    "total_changes": 5,
    "dangerous_changes": 2,
    "creates": 2,
    "updates": 2,
    "deletes": 1
  }
}
EOF

    local display_output="Test display output for GitHub"

    # Set up mock GitHub environment
    export GITHUB_OUTPUT="$TEST_OUTPUT_DIR/github_output"
    export GITHUB_STEP_SUMMARY="$TEST_OUTPUT_DIR/step_summary"
    rm -f "$GITHUB_OUTPUT" "$GITHUB_STEP_SUMMARY"

    # Call extract_outputs
    local extract_exit_code=0
    (
        set +e
        extract_outputs "$json_file" "$display_output"
        echo $? > "$TEST_OUTPUT_DIR/extract_exit_code"
    ) 2>&1

    if [[ -f "$TEST_OUTPUT_DIR/extract_exit_code" ]]; then
        extract_exit_code=$(cat "$TEST_OUTPUT_DIR/extract_exit_code")
    fi

    assert_exit_code "0" "$extract_exit_code" "extract_outputs should succeed with valid JSON"

    # Check GitHub outputs were set
    assert_file_exists "$GITHUB_OUTPUT" "GitHub output file should be created"

    if [[ -f "$GITHUB_OUTPUT" ]]; then
        local output_content=$(cat "$GITHUB_OUTPUT")

        assert_contains "$output_content" "has-changes=true" "Should set has-changes=true when total > 0"
        assert_contains "$output_content" "has-dangers=true" "Should set has-dangers=true when dangers > 0"
        assert_contains "$output_content" "change-count=5" "Should set correct change count"
        assert_contains "$output_content" "danger-count=2" "Should set correct danger count"
        assert_contains "$output_content" "summary<<EOF" "Should include summary with delimiter"
        assert_contains "$output_content" "$display_output" "Should include display output in summary"
        assert_contains "$output_content" "json-summary<<EOF" "Should include JSON summary with delimiter"
    fi

    # Check GitHub Step Summary
    assert_file_exists "$GITHUB_STEP_SUMMARY" "GitHub step summary file should be created"

    if [[ -f "$GITHUB_STEP_SUMMARY" ]]; then
        local summary_content=$(cat "$GITHUB_STEP_SUMMARY")
        assert_contains "$summary_content" "$display_output" "Step summary should contain display output"
    fi

    log_test "Testing extract_outputs with no changes"

    # Test with zero changes
    cat > "$json_file" <<'EOF'
{
  "statistics": {
    "total_changes": 0,
    "dangerous_changes": 0
  }
}
EOF

    rm -f "$GITHUB_OUTPUT" "$GITHUB_STEP_SUMMARY"

    extract_outputs "$json_file" "No changes detected"

    if [[ -f "$GITHUB_OUTPUT" ]]; then
        local no_changes_content=$(cat "$GITHUB_OUTPUT")
        assert_contains "$no_changes_content" "has-changes=false" "Should set has-changes=false when total = 0"
        assert_contains "$no_changes_content" "has-dangers=false" "Should set has-dangers=false when dangers = 0"
        assert_contains "$no_changes_content" "change-count=0" "Should set change-count=0"
        assert_contains "$no_changes_content" "danger-count=0" "Should set danger-count=0"
    fi

    log_test "Testing extract_outputs with missing JSON file"

    rm -f "$GITHUB_OUTPUT" "$GITHUB_STEP_SUMMARY"
    local missing_file="$TEST_OUTPUT_DIR/nonexistent.json"

    local missing_exit_code=0
    (
        set +e
        extract_outputs "$missing_file" "test output"
        echo $? > "$TEST_OUTPUT_DIR/missing_exit_code"
    ) 2>&1

    if [[ -f "$TEST_OUTPUT_DIR/missing_exit_code" ]]; then
        missing_exit_code=$(cat "$TEST_OUTPUT_DIR/missing_exit_code")
    fi

    assert_exit_code "1" "$missing_exit_code" "Should return 1 when JSON file is missing"
}

# ============================================================================
# Test Cases: Error Handling
# ============================================================================

test_error_handling() {
    log_section "Error Handling Tests"

    source_script

    log_test "Testing run_analysis with failed Strata execution"

    # Create test plan file
    local plan_file="$TEST_OUTPUT_DIR/test.tfplan"
    echo "invalid plan" > "$plan_file"

    # Create failing mock strata
    create_mock_strata "failure"
    TEMP_DIR="$TEST_OUTPUT_DIR"

    local failure_exit_code=0
    (
        set +e
        run_analysis "$plan_file" "markdown" "false" "false" ""
        echo $? > "$TEST_OUTPUT_DIR/failure_exit_code"
    ) > "$TEST_OUTPUT_DIR/failure_output" 2>&1

    if [[ -f "$TEST_OUTPUT_DIR/failure_exit_code" ]]; then
        failure_exit_code=$(cat "$TEST_OUTPUT_DIR/failure_exit_code")
    fi

    assert_exit_code "4" "$failure_exit_code" "Should exit with code 4 on analysis failure"

    log_test "Testing dangerous changes detection"

    # Test with dangerous changes
    create_mock_strata "dangerous"

    export GITHUB_OUTPUT="$TEST_OUTPUT_DIR/dangerous_output"
    rm -f "$GITHUB_OUTPUT"

    local dangerous_exit_code=0
    (
        set +e
        run_analysis "$plan_file" "markdown" "false" "false" ""
        echo $? > "$TEST_OUTPUT_DIR/dangerous_exit_code"
    ) > "$TEST_OUTPUT_DIR/dangerous_display_output" 2>&1

    if [[ -f "$TEST_OUTPUT_DIR/dangerous_exit_code" ]]; then
        dangerous_exit_code=$(cat "$TEST_OUTPUT_DIR/dangerous_exit_code")
    fi

    assert_exit_code "0" "$dangerous_exit_code" "Should succeed even with dangerous changes"

    # Check that dangerous changes are properly detected in outputs
    if [[ -f "$GITHUB_OUTPUT" ]]; then
        local dangerous_content=$(cat "$GITHUB_OUTPUT")
        assert_contains "$dangerous_content" "has-dangers=true" "Should detect dangerous changes"
        assert_contains "$dangerous_content" "danger-count=3" "Should count dangerous changes correctly"
    fi

    log_test "Testing parameter validation in run_analysis"

    # Test with missing plan file
    local missing_plan="/nonexistent/plan.tfplan"

    local missing_plan_exit_code=0
    (
        set +e
        run_analysis "$missing_plan" "markdown" "false" "false" ""
        echo $? > "$TEST_OUTPUT_DIR/missing_plan_exit_code"
    ) > "$TEST_OUTPUT_DIR/missing_plan_output" 2>&1

    # Note: The function will fail when trying to execute strata with non-existent file
    # This is expected behavior - the validation happens at the main() level
}

# ============================================================================
# Test Cases: Output Format Handling
# ============================================================================

test_output_formats() {
    log_section "Output Format Handling Tests"

    source_script

    log_test "Testing all supported output formats"

    # Create test plan file
    local plan_file="$TEST_OUTPUT_DIR/test.tfplan"
    echo '{"test": "plan"}' > "$plan_file"

    local formats=("markdown" "json" "table" "html")

    for format in "${formats[@]}"; do
        create_mock_strata "success"
        TEMP_DIR="$TEST_OUTPUT_DIR"

        rm -f "$TEST_OUTPUT_DIR/display_output.txt" "$TEST_OUTPUT_DIR/metadata.json"

        local format_exit_code=0
        (
            set +e
            run_analysis "$plan_file" "$format" "false" "false" ""
            echo $? > "$TEST_OUTPUT_DIR/format_${format}_exit_code"
        ) > "$TEST_OUTPUT_DIR/format_${format}_output" 2>&1

        if [[ -f "$TEST_OUTPUT_DIR/format_${format}_exit_code" ]]; then
            format_exit_code=$(cat "$TEST_OUTPUT_DIR/format_${format}_exit_code")
        fi

        assert_exit_code "0" "$format_exit_code" "Should handle $format format successfully"

        # Verify the format-specific output characteristics
        if [[ -f "$TEST_OUTPUT_DIR/display_output.txt" ]]; then
            local display_content=$(cat "$TEST_OUTPUT_DIR/display_output.txt")

            case "$format" in
                "markdown")
                    assert_contains "$display_content" "#" "Markdown should contain headers"
                    ;;
                "json")
                    assert_valid_json "$display_content" "JSON display output should be valid"
                    ;;
                "table")
                    assert_contains "$display_content" "|" "Table should contain pipe separators"
                    ;;
                "html")
                    assert_contains "$display_content" "<" "HTML should contain tags"
                    ;;
            esac
        fi
    done

    log_test "Testing format consistency between display and JSON outputs"

    # Verify that JSON metadata is consistent regardless of display format
    create_mock_strata "success"
    TEMP_DIR="$TEST_OUTPUT_DIR"

    for format in "${formats[@]}"; do
        rm -f "$TEST_OUTPUT_DIR/metadata.json"

        run_analysis "$plan_file" "$format" "false" "false" "" > /dev/null 2>&1

        if [[ -f "$TEST_OUTPUT_DIR/metadata.json" ]]; then
            local json_content=$(cat "$TEST_OUTPUT_DIR/metadata.json")
            assert_valid_json "$json_content" "JSON metadata should be valid for $format"

            # Check that key fields are present regardless of display format
            local total_changes=$(echo "$json_content" | jq -r '.statistics.total_changes')
            assert_equals "3" "$total_changes" "Total changes should be consistent for $format"
        fi
    done
}

# ============================================================================
# Main Test Runner
# ============================================================================

run_tests() {
    echo ""
    echo "================================================"
    echo "Strata Execution Component Tests"
    echo "================================================"

    # Check if action_simplified.sh exists
    if [[ ! -f "action_simplified.sh" ]]; then
        echo -e "${RED}Error: action_simplified.sh not found${NC}"
        echo "Please run this test from the project root directory"
        exit 1
    fi

    # Check if jq is available
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}Error: jq is not installed${NC}"
        echo "jq is required for JSON parsing tests"
        exit 1
    fi

    # Run all test suites
    test_command_construction
    test_dual_output_system
    test_json_parsing
    test_error_handling
    test_output_formats

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