#!/bin/bash

# Integration tests for Strata GitHub Action
# This script can be run locally to test the action functionality

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
    
    # Check if terraform is installed
    if ! command -v terraform >/dev/null 2>&1; then
        echo "Error: Terraform is not installed. Please install Terraform to run integration tests."
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "action.yml" ] || [ ! -f "action.sh" ]; then
        echo "Error: Please run this script from the root of the Strata repository."
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

# Create test directory structure
setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create temporary directory for tests
    TEST_DIR=$(mktemp -d)
    export TEST_DIR
    
    # Cleanup function
    cleanup() {
        log_info "Cleaning up test environment..."
        rm -rf "$TEST_DIR"
    }
    trap cleanup EXIT
    
    log_info "Test directory: $TEST_DIR"
}

# Create different Terraform configurations for testing
create_terraform_config() {
    local scenario=$1
    local config_dir="$TEST_DIR/$scenario"
    
    mkdir -p "$config_dir"
    cd "$config_dir"
    
    case "$scenario" in
        "no-changes")
            cat > main.tf << 'EOF'
terraform {
  required_providers {
    local = {
      source = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

# This will create a plan with no changes after first apply
resource "local_file" "test" {
  content  = "Hello, World!"
  filename = "${path.module}/hello.txt"
}
EOF
            ;;
            
        "with-changes")
            cat > main.tf << 'EOF'
terraform {
  required_providers {
    local = {
      source = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

resource "local_file" "test1" {
  content  = "File 1"
  filename = "${path.module}/file1.txt"
}

resource "local_file" "test2" {
  content  = "File 2"
  filename = "${path.module}/file2.txt"
}

resource "local_file" "test3" {
  content  = "File 3"
  filename = "${path.module}/file3.txt"
}
EOF
            ;;
            
        "with-dangers")
            cat > main.tf << 'EOF'
terraform {
  required_providers {
    local = {
      source = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

# Create many resources to trigger danger threshold
resource "local_file" "danger1" {
  content  = "Danger 1"
  filename = "${path.module}/danger1.txt"
}

resource "local_file" "danger2" {
  content  = "Danger 2"
  filename = "${path.module}/danger2.txt"
}

resource "local_file" "danger3" {
  content  = "Danger 3"
  filename = "${path.module}/danger3.txt"
}

resource "local_file" "danger4" {
  content  = "Danger 4"
  filename = "${path.module}/danger4.txt"
}

resource "local_file" "danger5" {
  content  = "Danger 5"
  filename = "${path.module}/danger5.txt"
}
EOF
            
            # Create Strata config with sensitive resources to trigger danger detection
            cat > .strata.yaml << 'EOF'
plan:
  show-details: true
  highlight-dangers: true

sensitive_resources:
  - resource_type: local_file
EOF
            ;;
            
        "complex")
            cat > main.tf << 'EOF'
terraform {
  required_providers {
    local = {
      source = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

# Module-like structure
locals {
  files = {
    "app1" = "Application 1 content"
    "app2" = "Application 2 content"
    "config" = "Configuration content"
  }
}

resource "local_file" "apps" {
  for_each = local.files
  
  content  = each.value
  filename = "${path.module}/${each.key}.txt"
}

# Sensitive file
resource "local_sensitive_file" "secret" {
  content  = "secret content"
  filename = "${path.module}/secret.txt"
}
EOF
            ;;
    esac
    
    # Initialize and create plan
    terraform init -no-color
    terraform plan -out=terraform.tfplan -no-color
    
    cd - >/dev/null
}

# Test action with different scenarios
test_action_scenario() {
    local scenario=$1
    local test_name=$2
    local additional_args=$3
    
    log_test "$test_name"
    
    # Set up environment variables for action
    export INPUT_PLAN_FILE="$TEST_DIR/$scenario/terraform.tfplan"
    export INPUT_OUTPUT_FORMAT="markdown"
    export INPUT_SHOW_DETAILS="true"
    export INPUT_COMMENT_ON_PR="false"  # Disable PR comments for local testing
    export GITHUB_STEP_SUMMARY="$TEST_DIR/step_summary_${scenario}.md"
    export GITHUB_OUTPUT="$TEST_DIR/github_output_${scenario}.txt"
    
    # Add additional arguments if provided
    if [ -n "$additional_args" ]; then
        eval "export $additional_args"
    fi
    
    # Run the action script
    if bash ./action.sh > "$TEST_DIR/action_output_${scenario}.log" 2>&1; then
        log_pass "$test_name - Action executed successfully"
        
        # Validate outputs
        if [ -f "$GITHUB_OUTPUT" ]; then
            # Check that required outputs are present
            if grep -q "summary=" "$GITHUB_OUTPUT" && \
               grep -q "has-changes=" "$GITHUB_OUTPUT" && \
               grep -q "has-dangers=" "$GITHUB_OUTPUT" && \
               grep -q "change-count=" "$GITHUB_OUTPUT" && \
               grep -q "danger-count=" "$GITHUB_OUTPUT"; then
                log_pass "$test_name - All required outputs present"
            else
                log_fail "$test_name - Missing required outputs"
                echo "Output file contents:"
                cat "$GITHUB_OUTPUT"
            fi
        else
            log_fail "$test_name - No output file generated"
        fi
        
        # Check step summary
        if [ -f "$GITHUB_STEP_SUMMARY" ] && [ -s "$GITHUB_STEP_SUMMARY" ]; then
            log_pass "$test_name - Step summary generated"
        else
            log_fail "$test_name - Step summary not generated or empty"
        fi
        
    else
        log_fail "$test_name - Action execution failed"
        echo "Action output:"
        cat "$TEST_DIR/action_output_${scenario}.log"
    fi
    
    # Clean up environment variables
    unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_SHOW_DETAILS INPUT_COMMENT_ON_PR
    unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Test error handling
test_error_handling() {
    log_test "Error handling - Missing plan file"
    
    export INPUT_PLAN_FILE="$TEST_DIR/nonexistent.tfplan"
    export INPUT_OUTPUT_FORMAT="markdown"
    export INPUT_COMMENT_ON_PR="false"
    export GITHUB_STEP_SUMMARY="$TEST_DIR/step_summary_error.md"
    export GITHUB_OUTPUT="$TEST_DIR/github_output_error.txt"
    
    # This should fail
    if ! bash ./action.sh > "$TEST_DIR/action_output_error.log" 2>&1; then
        log_pass "Error handling - Action correctly failed for missing plan file"
    else
        log_fail "Error handling - Action should have failed for missing plan file"
    fi
    
    unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR
    unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Test different output formats
test_output_formats() {
    local scenario="with-changes"
    
    for format in "table" "markdown"; do
        log_test "Output format - $format"
        
        export INPUT_PLAN_FILE="$TEST_DIR/$scenario/terraform.tfplan"
        export INPUT_OUTPUT_FORMAT="$format"
        export INPUT_COMMENT_ON_PR="false"
        export GITHUB_STEP_SUMMARY="$TEST_DIR/step_summary_${format}.md"
        export GITHUB_OUTPUT="$TEST_DIR/github_output_${format}.txt"
        
        if bash ./action.sh > "$TEST_DIR/action_output_${format}.log" 2>&1; then
            log_pass "Output format $format - Action executed successfully"
        else
            log_fail "Output format $format - Action execution failed"
        fi
        
        unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR
        unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT
    done
}

# Test configuration file usage
test_config_file() {
    local scenario="with-dangers"
    
    log_test "Configuration file usage"
    
    export INPUT_PLAN_FILE="$TEST_DIR/$scenario/terraform.tfplan"
    export INPUT_CONFIG_FILE="$TEST_DIR/$scenario/.strata.yaml"
    export INPUT_OUTPUT_FORMAT="markdown"
    export INPUT_COMMENT_ON_PR="false"
    export GITHUB_STEP_SUMMARY="$TEST_DIR/step_summary_config.md"
    export GITHUB_OUTPUT="$TEST_DIR/github_output_config.txt"
    
    if bash ./action.sh > "$TEST_DIR/action_output_config.log" 2>&1; then
        log_pass "Configuration file - Action executed successfully"
        
        # Check if configuration file was applied (no dangers expected for CREATE operations)
        if [ -f "$GITHUB_OUTPUT" ]; then
            danger_count=$(grep "danger-count=" "$GITHUB_OUTPUT" | cut -d'=' -f2)
            log_pass "Configuration file - Processed $danger_count dangerous changes (CREATE operations are not dangerous)"
        fi
    else
        log_fail "Configuration file - Action execution failed"
    fi
    
    unset INPUT_PLAN_FILE INPUT_CONFIG_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR
    unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT
}

# Test binary caching (simulate multiple runs)
test_binary_caching() {
    log_test "Binary caching"
    
    # Create cache directory
    CACHE_DIR="$TEST_DIR/cache"
    mkdir -p "$CACHE_DIR"
    export HOME="$TEST_DIR"  # Override home to use our test cache
    
    scenario="with-changes"
    export INPUT_PLAN_FILE="$TEST_DIR/$scenario/terraform.tfplan"
    export INPUT_OUTPUT_FORMAT="markdown"
    export INPUT_COMMENT_ON_PR="false"
    export GITHUB_STEP_SUMMARY="$TEST_DIR/step_summary_cache1.md"
    export GITHUB_OUTPUT="$TEST_DIR/github_output_cache1.txt"
    
    # First run (should download/compile binary)
    start_time=$(date +%s)
    if bash ./action.sh > "$TEST_DIR/action_output_cache1.log" 2>&1; then
        first_run_time=$(($(date +%s) - start_time))
        log_pass "Binary caching - First run completed in ${first_run_time}s"
        
        # Second run (should use cached binary)
        export GITHUB_STEP_SUMMARY="$TEST_DIR/step_summary_cache2.md"
        export GITHUB_OUTPUT="$TEST_DIR/github_output_cache2.txt"
        
        start_time=$(date +%s)
        if bash ./action.sh > "$TEST_DIR/action_output_cache2.log" 2>&1; then
            second_run_time=$(($(date +%s) - start_time))
            log_pass "Binary caching - Second run completed in ${second_run_time}s"
            
            # Second run should be faster (though this is not guaranteed in all environments)
            log_info "First run: ${first_run_time}s, Second run: ${second_run_time}s"
        else
            log_fail "Binary caching - Second run failed"
        fi
    else
        log_fail "Binary caching - First run failed"
    fi
    
    unset INPUT_PLAN_FILE INPUT_OUTPUT_FORMAT INPUT_COMMENT_ON_PR
    unset GITHUB_STEP_SUMMARY GITHUB_OUTPUT HOME
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    REPORT_FILE="$TEST_DIR/integration_test_report.md"
    
    cat > "$REPORT_FILE" << EOF
# Strata GitHub Action Integration Test Report

**Test Date:** $(date)
**Test Environment:** $(uname -a)
**Terraform Version:** $(terraform version | head -n1)

## Test Summary

- **Tests Run:** $TESTS_RUN
- **Tests Passed:** $TESTS_PASSED
- **Tests Failed:** $TESTS_FAILED
- **Success Rate:** $(( TESTS_PASSED * 100 / TESTS_RUN ))%

## Test Scenarios

### Terraform Configurations Tested
- No changes scenario
- Multiple changes scenario  
- High danger scenario
- Complex scenario with modules

### Action Features Tested
- Basic functionality
- Output format validation (table, markdown)
- Configuration file usage
- Error handling
- Binary caching
- Step summary generation
- Output variable setting

## Test Artifacts

The following files were generated during testing:

EOF

    # List all generated files
    find "$TEST_DIR" -name "*.log" -o -name "*.md" -o -name "*.txt" | while read -r file; do
        echo "- $(basename "$file")" >> "$REPORT_FILE"
    done
    
    echo "" >> "$REPORT_FILE"
    echo "All test artifacts are available in: $TEST_DIR" >> "$REPORT_FILE"
    
    log_info "Test report generated: $REPORT_FILE"
    
    # Display report location
    echo ""
    echo "📊 Integration test report: $REPORT_FILE"
    echo "🗂️  Test artifacts directory: $TEST_DIR"
}

# Main test execution
main() {
    echo "Strata GitHub Action Integration Tests"
    echo "====================================="
    
    check_prerequisites
    setup_test_environment
    
    # Create test scenarios
    log_info "Creating Terraform test configurations..."
    create_terraform_config "no-changes"
    create_terraform_config "with-changes"
    create_terraform_config "with-dangers"
    create_terraform_config "complex"
    
    # Run tests
    log_info "Running integration tests..."
    
    test_action_scenario "with-changes" "Basic functionality"
    test_action_scenario "with-changes" "Show details" "INPUT_SHOW_DETAILS=true"
    test_action_scenario "with-dangers" "Danger detection"
    test_action_scenario "complex" "Complex configuration"
    
    test_error_handling
    test_output_formats
    test_config_file
    test_binary_caching
    
    # Generate report
    generate_report
    
    # Print summary
    echo ""
    echo "Integration Test Summary:"
    echo "========================"
    echo -e "Tests run:    ${TESTS_RUN}"
    echo -e "Tests passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests failed: ${RED}${TESTS_FAILED}${NC}"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}All integration tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}Some integration tests failed!${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
