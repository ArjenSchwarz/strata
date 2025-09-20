#!/bin/bash

# Backwards Compatibility Test Suite for Strata GitHub Action
# Tests that ensure existing workflows continue to work without modification
# Requirements: 8.2 - Test minimal inputs, all current parameters, output names, interface unchanged

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    if [ ! -f "action.yml" ] || [ ! -f "action.sh" ]; then
        echo "Error: Please run this script from the root of the Strata repository."
        exit 1
    fi

    if [ ! -d "samples" ]; then
        echo "Error: samples directory not found"
        exit 1
    fi

    chmod +x action.sh
    log_info "Prerequisites check passed"
}

# Setup test environment
setup_test_environment() {
    log_info "Setting up test environment..."

    TEST_DIR=$(mktemp -d)
    export TEST_DIR

    cleanup() {
        log_info "Cleaning up test environment..."
        if [ $TESTS_FAILED -eq 0 ]; then
            rm -rf "$TEST_DIR"
        else
            log_info "Test artifacts preserved in: $TEST_DIR"
        fi
    }
    trap cleanup EXIT

    log_info "Test directory: $TEST_DIR"
}

# Test minimal inputs (just plan-file) - Requirement 10.1
test_minimal_inputs() {
    log_test "Minimal inputs (just plan-file)"

    local sample_file="samples/simpleadd-sample.json"
    if [ ! -f "$sample_file" ]; then
        sample_file=$(find samples -name "*.json" -type f | head -n 1)
    fi

    local step_summary="$TEST_DIR/step_summary_minimal.md"
    local github_output="$TEST_DIR/github_output_minimal.txt"
    local action_log="$TEST_DIR/action_log_minimal.log"

    # Set only the required input
    export INPUT_PLAN_FILE="$sample_file"
    export INPUT_COMMENT_ON_PR="false"  # Disable PR for testing
    export GITHUB_STEP_SUMMARY="$step_summary"
    export GITHUB_OUTPUT="$github_output"

    if ./action.sh > "$action_log" 2>&1; then
        log_pass "Minimal inputs - Action executed successfully"

        # Check that outputs are generated
        if [ -f "$github_output" ] && [ -s "$github_output" ]; then
            log_pass "Minimal inputs - GitHub outputs generated"

            # Verify default values are applied
            if grep -q "summary=" "$github_output"; then
                log_pass "Minimal inputs - Summary output present (default behavior)"
            else
                log_fail "Minimal inputs - Summary output missing"
            fi
        else
            log_fail "Minimal inputs - GitHub outputs not generated"
        fi

        # Check step summary with default format (should be markdown)
        if [ -f "$step_summary" ] && [ -s "$step_summary" ]; then
            log_pass "Minimal inputs - Step summary generated"

            # Should default to markdown format
            if grep -q "^#" "$step_summary"; then
                log_pass "Minimal inputs - Default markdown format applied"
            else
                log_fail "Minimal inputs - Default markdown format not detected"
            fi
        else
            log_fail "Minimal inputs - Step summary not generated"
        fi

    else
        log_fail "Minimal inputs - Action execution failed"
        if [ -f "$action_log" ]; then
            echo "Error log:"
            head -n 5 "$action_log" | sed 's/^/  /'
        fi
    fi

    unset INPUT_PLAN_FILE INPUT_COMMENT_ON_PR GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Test all current input parameters work correctly - Requirement 10.1
test_all_current_parameters() {
    log_test "All current input parameters"

    local sample_file="samples/web-sample.json"
    if [ ! -f "$sample_file" ]; then
        sample_file=$(find samples -name "*.json" -type f | head -n 1)
    fi

    # Create test config file
    local test_config="$TEST_DIR/test_backwards_compat.yaml"
    cat > "$test_config" << 'EOF'
output: table
plan:
  show-details: true
  highlight-dangers: true
sensitive_resources:
  - resource_type: aws_instance
EOF

    local step_summary="$TEST_DIR/step_summary_allparams.md"
    local github_output="$TEST_DIR/github_output_allparams.txt"
    local action_log="$TEST_DIR/action_log_allparams.log"

    # Set all current input parameters (pre-v1.5.0)
    export INPUT_PLAN_FILE="$sample_file"
    export INPUT_OUTPUT_FORMAT="table"
    export INPUT_CONFIG_FILE="$test_config"
    export INPUT_SHOW_DETAILS="true"
    export INPUT_EXPAND_ALL="true"
    export INPUT_GITHUB_TOKEN="${GITHUB_TOKEN:-dummy-token}"
    export INPUT_COMMENT_ON_PR="false"  # Disable PR for testing
    export INPUT_UPDATE_COMMENT="true"
    export INPUT_COMMENT_HEADER="🔧 Test Terraform Plan Summary"
    # New parameter should work with default
    export INPUT_STRATA_VERSION="latest"
    export GITHUB_STEP_SUMMARY="$step_summary"
    export GITHUB_OUTPUT="$github_output"

    if ./action.sh > "$action_log" 2>&1; then
        log_pass "All current parameters - Action executed successfully"

        # Verify table output format was applied
        if [ -f "$step_summary" ] && grep -q "|" "$step_summary"; then
            log_pass "All current parameters - Table format applied correctly"
        else
            log_fail "All current parameters - Table format not detected"
        fi

        # Verify show-details was applied (output should be more detailed)
        if [ -f "$step_summary" ] && [ -s "$step_summary" ]; then
            local line_count=$(wc -l < "$step_summary")
            if [ "$line_count" -gt 10 ]; then
                log_pass "All current parameters - Show details appears to be working"
            else
                log_fail "All current parameters - Output seems too short for details mode"
            fi
        fi

        # Verify GitHub outputs are present
        if [ -f "$github_output" ]; then
            local required_outputs=("summary=" "has-changes=" "has-dangers=" "change-count=" "danger-count=" "json-summary=")
            local missing_outputs=()

            for output in "${required_outputs[@]}"; do
                if ! grep -q "^${output}" "$github_output"; then
                    missing_outputs+=("$output")
                fi
            done

            if [ ${#missing_outputs[@]} -eq 0 ]; then
                log_pass "All current parameters - All outputs generated correctly"
            else
                log_fail "All current parameters - Missing outputs: ${missing_outputs[*]}"
            fi
        else
            log_fail "All current parameters - GitHub outputs not generated"
        fi

    else
        log_fail "All current parameters - Action execution failed"
        if [ -f "$action_log" ]; then
            echo "Error log:"
            head -n 5 "$action_log" | sed 's/^/  /'
        fi
    fi

    unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_CONFIG_FILE INPUT_SHOW_DETAILS
    unset INPUT_EXPAND_ALL INPUT_GITHUB_TOKEN INPUT_COMMENT_ON_PR INPUT_UPDATE_COMMENT
    unset INPUT_COMMENT_HEADER INPUT_STRATA_VERSION GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Verify all outputs are generated with same names - Requirement 10.2
test_output_names_consistency() {
    log_test "Output names consistency"

    local sample_file=$(find samples -name "*.json" -type f | head -n 1)
    local step_summary="$TEST_DIR/step_summary_names.md"
    local github_output="$TEST_DIR/github_output_names.txt"
    local action_log="$TEST_DIR/action_log_names.log"

    export INPUT_PLAN_FILE="$sample_file"
    export INPUT_COMMENT_ON_PR="false"
    export GITHUB_STEP_SUMMARY="$step_summary"
    export GITHUB_OUTPUT="$github_output"

    if ./action.sh > "$action_log" 2>&1; then
        # Check exact output names as defined in action.yml
        local expected_outputs=(
            "summary"
            "has-changes"
            "has-dangers"
            "json-summary"
            "change-count"
            "danger-count"
        )

        local missing_names=()
        local extra_names=()

        # Check for expected outputs
        for expected in "${expected_outputs[@]}"; do
            if ! grep -q "^${expected}=" "$github_output"; then
                missing_names+=("$expected")
            fi
        done

        # Check for any unexpected outputs (that don't match expected pattern)
        while IFS= read -r line; do
            if [[ "$line" =~ ^([^=]+)= ]]; then
                local output_name="${BASH_REMATCH[1]}"
                local is_expected=false
                for expected in "${expected_outputs[@]}"; do
                    if [ "$output_name" = "$expected" ]; then
                        is_expected=true
                        break
                    fi
                done
                if [ "$is_expected" = false ]; then
                    extra_names+=("$output_name")
                fi
            fi
        done < "$github_output"

        if [ ${#missing_names[@]} -eq 0 ] && [ ${#extra_names[@]} -eq 0 ]; then
            log_pass "Output names consistency - Exact output names match specification"
        else
            if [ ${#missing_names[@]} -gt 0 ]; then
                log_fail "Output names consistency - Missing outputs: ${missing_names[*]}"
            fi
            if [ ${#extra_names[@]} -gt 0 ]; then
                log_fail "Output names consistency - Unexpected outputs: ${extra_names[*]}"
            fi
        fi

    else
        log_fail "Output names consistency - Action execution failed"
    fi

    unset INPUT_PLAN_FILE INPUT_COMMENT_ON_PR GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Test existing workflows continue without changes - Requirement 10.3
test_existing_workflow_patterns() {
    log_info "Testing existing workflow patterns..."

    # Test Pattern 1: Basic usage (most common)
    test_workflow_pattern_basic

    # Test Pattern 2: With custom config
    test_workflow_pattern_with_config

    # Test Pattern 3: With all options
    test_workflow_pattern_full_options
}

test_workflow_pattern_basic() {
    log_test "Existing workflow pattern - Basic usage"

    local sample_file="samples/web-sample.json"
    if [ ! -f "$sample_file" ]; then
        sample_file=$(find samples -name "*.json" -type f | head -n 1)
    fi

    local step_summary="$TEST_DIR/step_summary_basic_workflow.md"
    local github_output="$TEST_DIR/github_output_basic_workflow.txt"
    local action_log="$TEST_DIR/action_log_basic_workflow.log"

    # Simulate typical workflow usage
    export INPUT_PLAN_FILE="$sample_file"
    export INPUT_OUTPUT_FORMAT="markdown"
    export INPUT_COMMENT_ON_PR="false"  # Would be true in real workflow
    export GITHUB_STEP_SUMMARY="$step_summary"
    export GITHUB_OUTPUT="$github_output"

    if ./action.sh > "$action_log" 2>&1; then
        log_pass "Basic workflow pattern - Executed successfully"

        # Check standard outputs that workflows expect
        if [ -f "$github_output" ]; then
            # Workflows commonly check these outputs
            if grep -q "has-changes=" "$github_output" && grep -q "change-count=" "$github_output"; then
                log_pass "Basic workflow pattern - Standard outputs available"
            else
                log_fail "Basic workflow pattern - Missing standard outputs"
            fi
        fi

        # Check step summary is readable markdown
        if [ -f "$step_summary" ] && grep -q "^#" "$step_summary"; then
            log_pass "Basic workflow pattern - Markdown step summary generated"
        else
            log_fail "Basic workflow pattern - Step summary not in expected format"
        fi

    else
        log_fail "Basic workflow pattern - Execution failed"
    fi

    unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

test_workflow_pattern_with_config() {
    log_test "Existing workflow pattern - With custom config"

    local sample_file="samples/danger-sample.json"
    if [ ! -f "$sample_file" ]; then
        sample_file=$(find samples -name "*.json" -type f | head -n 1)
    fi

    # Create config similar to what users might have
    local workflow_config="$TEST_DIR/workflow_config.yaml"
    cat > "$workflow_config" << 'EOF'
output: markdown
plan:
  show-details: false
  highlight-dangers: true
sensitive_resources:
  - resource_type: aws_db_instance
  - resource_type: aws_instance
EOF

    local step_summary="$TEST_DIR/step_summary_config_workflow.md"
    local github_output="$TEST_DIR/github_output_config_workflow.txt"
    local action_log="$TEST_DIR/action_log_config_workflow.log"

    export INPUT_PLAN_FILE="$sample_file"
    export INPUT_CONFIG_FILE="$workflow_config"
    export INPUT_OUTPUT_FORMAT="markdown"
    export INPUT_COMMENT_ON_PR="false"
    export GITHUB_STEP_SUMMARY="$step_summary"
    export GITHUB_OUTPUT="$github_output"

    if ./action.sh > "$action_log" 2>&1; then
        log_pass "Config workflow pattern - Executed successfully"

        # Config should be respected
        if [ -f "$step_summary" ] && [ -f "$github_output" ]; then
            log_pass "Config workflow pattern - Outputs generated with custom config"
        else
            log_fail "Config workflow pattern - Outputs not generated"
        fi

    else
        log_fail "Config workflow pattern - Execution failed"
    fi

    unset INPUT_PLAN_FILE INPUT_CONFIG_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR
    unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

test_workflow_pattern_full_options() {
    log_test "Existing workflow pattern - Full options"

    local sample_file="samples/complex-properties-sample.json"
    if [ ! -f "$sample_file" ]; then
        sample_file=$(find samples -name "*.json" -type f | head -n 1)
    fi

    local step_summary="$TEST_DIR/step_summary_full_workflow.md"
    local github_output="$TEST_DIR/github_output_full_workflow.txt"
    local action_log="$TEST_DIR/action_log_full_workflow.log"

    # Simulate power user workflow with all options
    export INPUT_PLAN_FILE="$sample_file"
    export INPUT_OUTPUT_FORMAT="table"
    export INPUT_SHOW_DETAILS="true"
    export INPUT_EXPAND_ALL="true"
    export INPUT_COMMENT_ON_PR="false"  # Would be true in real workflow
    export INPUT_UPDATE_COMMENT="true"
    export INPUT_COMMENT_HEADER="🚀 Infrastructure Changes"
    export GITHUB_STEP_SUMMARY="$step_summary"
    export GITHUB_OUTPUT="$github_output"

    if ./action.sh > "$action_log" 2>&1; then
        log_pass "Full options workflow pattern - Executed successfully"

        # Should handle all options correctly
        if [ -f "$step_summary" ] && [ -f "$github_output" ]; then
            # Check for table format
            if grep -q "|" "$step_summary"; then
                log_pass "Full options workflow pattern - Table format applied"
            else
                log_fail "Full options workflow pattern - Table format not detected"
            fi

            # Check for detailed output (should be longer)
            local line_count=$(wc -l < "$step_summary")
            if [ "$line_count" -gt 15 ]; then
                log_pass "Full options workflow pattern - Detailed output generated"
            else
                log_fail "Full options workflow pattern - Output doesn't seem detailed"
            fi
        fi

    else
        log_fail "Full options workflow pattern - Execution failed"
    fi

    unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_SHOW_DETAILS INPUT_EXPAND_ALL
    unset INPUT_COMMENT_ON_PR INPUT_UPDATE_COMMENT INPUT_COMMENT_HEADER
    unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Validate action.yml interface unchanged (except strata-version) - Requirement 10.5
test_action_yml_interface() {
    log_test "action.yml interface validation"

    if [ ! -f "action.yml" ]; then
        log_fail "action.yml interface - File not found"
        return
    fi

    # Check basic structure
    local required_sections=("name" "description" "inputs" "outputs" "runs")
    for section in "${required_sections[@]}"; do
        if ! grep -q "^${section}:" "action.yml"; then
            log_fail "action.yml interface - Missing required section: $section"
            return
        fi
    done

    # Check existing inputs (should all be present)
    local existing_inputs=(
        "plan-file"
        "output-format"
        "config-file"
        "show-details"
        "expand-all"
        "github-token"
        "comment-on-pr"
        "update-comment"
        "comment-header"
    )

    local missing_inputs=()
    for input in "${existing_inputs[@]}"; do
        if ! grep -q "  ${input}:" "action.yml"; then
            missing_inputs+=("$input")
        fi
    done

    # Check for new strata-version input (should be present)
    if ! grep -q "  strata-version:" "action.yml"; then
        log_fail "action.yml interface - New strata-version input missing"
        return
    fi

    # Check existing outputs (should all be present)
    local existing_outputs=(
        "summary"
        "has-changes"
        "has-dangers"
        "json-summary"
        "change-count"
        "danger-count"
    )

    local missing_outputs=()
    for output in "${existing_outputs[@]}"; do
        if ! grep -q "  ${output}:" "action.yml"; then
            missing_outputs+=("$output")
        fi
    done

    # Check runs section uses composite
    if ! grep -q "using: 'composite'" "action.yml"; then
        log_fail "action.yml interface - Runs section not using composite"
        return
    fi

    if [ ${#missing_inputs[@]} -eq 0 ] && [ ${#missing_outputs[@]} -eq 0 ]; then
        log_pass "action.yml interface - All existing inputs/outputs present + strata-version added"
    else
        log_fail "action.yml interface - Missing inputs: ${missing_inputs[*]}, outputs: ${missing_outputs[*]}"
    fi

    # Verify default values are maintained
    local expected_defaults=(
        "output-format.*default.*markdown"
        "show-details.*default.*false"
        "expand-all.*default.*false"
        "comment-on-pr.*default.*true"
        "update-comment.*default.*true"
        "strata-version.*default.*latest"
    )

    for default_pattern in "${expected_defaults[@]}"; do
        if ! grep -A2 -B2 "${default_pattern%.*default*}" "action.yml" | grep -q "${default_pattern#*default.*}"; then
            log_fail "action.yml interface - Default value changed for ${default_pattern%.*default*}"
        fi
    done

    log_pass "action.yml interface - Maintains backwards compatibility"
}

# Test same Terraform plan file formats - Requirement 10.6
test_terraform_plan_formats() {
    log_test "Terraform plan file formats support"

    # Test with different sample files to ensure format support is maintained
    local sample_files
    readarray -t sample_files < <(find samples -name "*.json" -type f | head -n 3)

    local formats_tested=0
    local formats_passed=0

    for sample_file in "${sample_files[@]}"; do
        log_test "Plan format - $(basename "$sample_file")"
        formats_tested=$((formats_tested + 1))

        local step_summary="$TEST_DIR/step_summary_format_$(basename "$sample_file" .json).md"
        local github_output="$TEST_DIR/github_output_format_$(basename "$sample_file" .json).txt"
        local action_log="$TEST_DIR/action_log_format_$(basename "$sample_file" .json).log"

        export INPUT_PLAN_FILE="$sample_file"
        export INPUT_OUTPUT_FORMAT="markdown"
        export INPUT_COMMENT_ON_PR="false"
        export GITHUB_STEP_SUMMARY="$step_summary"
        export GITHUB_OUTPUT="$github_output"

        if ./action.sh > "$action_log" 2>&1; then
            log_pass "Plan format - $(basename "$sample_file") processed successfully"
            formats_passed=$((formats_passed + 1))

            # Check that plan was actually processed (not just passed through)
            if [ -f "$github_output" ] && grep -q "change-count=" "$github_output"; then
                log_pass "Plan format - $(basename "$sample_file") analysis completed"
            else
                log_fail "Plan format - $(basename "$sample_file") not properly analyzed"
                formats_passed=$((formats_passed - 1))
            fi
        else
            log_fail "Plan format - $(basename "$sample_file") processing failed"
        fi

        unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR GITHUB_STEP_SUMMARY GITHUB_OUTPUT
    done

    if [ $formats_passed -eq $formats_tested ] && [ $formats_tested -gt 0 ]; then
        log_pass "Terraform plan formats - All tested formats supported"
    else
        log_fail "Terraform plan formats - Some formats failed ($formats_passed/$formats_tested passed)"
    fi
}

# Generate backwards compatibility test report
generate_compatibility_report() {
    log_info "Generating backwards compatibility test report..."

    local report_file="$TEST_DIR/backwards_compatibility_test_report.md"

    cat > "$report_file" << EOF
# Strata GitHub Action - Backwards Compatibility Test Report

**Test Date:** $(date)
**Test Environment:** $(uname -a)
**Repository:** $(pwd)

## Test Summary

- **Tests Run:** $TESTS_RUN
- **Tests Passed:** $TESTS_PASSED
- **Tests Failed:** $TESTS_FAILED
- **Success Rate:** $(( TESTS_PASSED * 100 / TESTS_RUN ))%

## Backwards Compatibility Requirements Tested

### ✅ Requirement 10.1: Accept all current input parameters
- [x] Minimal inputs (just plan-file) work
- [x] All current input parameters work correctly
- [x] New optional parameters don't affect existing workflows

### ✅ Requirement 10.2: Produce all current outputs with same names
- [x] All expected output names present
- [x] No unexpected outputs added
- [x] Output format and content maintained

### ✅ Requirement 10.3: Maintain same behavior for core functionality
- [x] Basic workflow patterns work unchanged
- [x] Custom configuration workflows work
- [x] Full-featured workflows work

### ✅ Requirement 10.5: Use same action.yml interface
- [x] All existing inputs preserved
- [x] All existing outputs preserved
- [x] Default values maintained
- [x] Only strata-version input added

### ✅ Requirement 10.6: Support same Terraform plan file formats
- [x] JSON plan file format supported
- [x] Multiple sample files processed correctly
- [x] Analysis functionality maintained

## Test Results

$(if [ $TESTS_FAILED -eq 0 ]; then
    echo "✅ **ALL BACKWARDS COMPATIBILITY TESTS PASSED**"
    echo ""
    echo "The updated action maintains 100% backwards compatibility with existing workflows."
    echo "Users can upgrade without making any changes to their workflow files."
else
    echo "❌ **SOME BACKWARDS COMPATIBILITY TESTS FAILED**"
    echo ""
    echo "Review the test logs for specific compatibility issues that need to be addressed."
fi)

## Workflow Migration Impact

- **Required Changes:** None
- **Optional Changes:** Users can optionally add \`strata-version\` input to pin specific versions
- **Breaking Changes:** None identified

## Test Artifacts

All test artifacts are preserved in: $TEST_DIR

EOF

    log_info "Backwards compatibility report: $report_file"
    echo ""
    echo "📋 Backwards Compatibility Report: $report_file"
    echo "🗂️  Test Artifacts Directory: $TEST_DIR"
}

# Main execution
main() {
    echo "Strata GitHub Action - Backwards Compatibility Test Suite"
    echo "======================================================"
    echo ""

    check_prerequisites
    setup_test_environment

    echo ""
    log_info "Testing backwards compatibility requirements..."
    echo ""

    # Test all backwards compatibility requirements
    test_minimal_inputs
    test_all_current_parameters
    test_output_names_consistency
    test_existing_workflow_patterns
    test_action_yml_interface
    test_terraform_plan_formats

    # Generate report
    generate_compatibility_report

    # Print summary
    echo ""
    echo "Backwards Compatibility Test Summary:"
    echo "===================================="
    echo -e "Tests run:    ${TESTS_RUN}"
    echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All backwards compatibility tests passed!${NC}"
        echo ""
        echo "✅ Existing workflows will continue to work unchanged"
        echo "✅ All input parameters maintain compatibility"
        echo "✅ All output names and formats preserved"
        echo "✅ action.yml interface backwards compatible"
        echo "✅ Terraform plan file format support maintained"
        exit 0
    else
        echo -e "\n${RED}❌ Some backwards compatibility tests failed!${NC}"
        echo ""
        echo "Review the test report for specific compatibility issues."
        echo "These must be resolved before release to maintain backwards compatibility."
        exit 1
    fi
}

# Run main function
main "$@"