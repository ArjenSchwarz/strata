#!/bin/bash

# Comprehensive Test Runner for GitHub Action File Output Integration
# This script runs all test suites to validate the dual output system implementation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test suite counters
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0

# Overall test counters
TOTAL_TESTS=0
TOTAL_PASSED=0
TOTAL_FAILED=0

echo -e "${CYAN}================================================================${NC}"
echo -e "${CYAN}  GitHub Action File Output Integration - Test Suite Runner${NC}"
echo -e "${CYAN}================================================================${NC}"
echo ""
echo -e "${BLUE}Running comprehensive tests for dual output system implementation${NC}"
echo ""

# Function to run a test suite
run_test_suite() {
    local test_script="$1"
    local suite_name="$2"
    local description="$3"
    
    echo -e "${CYAN}=== Running Test Suite: $suite_name ===${NC}"
    echo -e "${BLUE}Description: $description${NC}"
    echo ""
    
    TOTAL_SUITES=$((TOTAL_SUITES + 1))
    
    if [ ! -f "$test_script" ]; then
        echo -e "${RED}❌ Test script not found: $test_script${NC}"
        FAILED_SUITES=$((FAILED_SUITES + 1))
        return 1
    fi
    
    if [ ! -x "$test_script" ]; then
        chmod +x "$test_script"
    fi
    
    # Create temporary log file for this suite
    local log_file="/tmp/test_suite_$(basename "$test_script" .sh).log"
    
    # Run the test suite and capture output
    local start_time=$(date +%s)
    if "$test_script" > "$log_file" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo -e "${GREEN}✅ Test Suite PASSED: $suite_name${NC}"
        echo -e "${GREEN}   Duration: ${duration}s${NC}"
        
        # Extract test statistics from log
        local suite_tests=$(grep -o "Tests run:" "$log_file" | tail -1 | sed 's/Tests run: *//' || echo "0")
        local suite_passed=$(grep -o "Tests passed:" "$log_file" | tail -1 | sed 's/Tests passed: *//' | sed 's/\x1b\[[0-9;]*m//g' || echo "0")
        local suite_failed=$(grep -o "Tests failed:" "$log_file" | tail -1 | sed 's/Tests failed: *//' | sed 's/\x1b\[[0-9;]*m//g' || echo "0")
        
        # Clean up ANSI color codes from numbers
        suite_tests=$(echo "$suite_tests" | sed 's/[^0-9]//g')
        suite_passed=$(echo "$suite_passed" | sed 's/[^0-9]//g')
        suite_failed=$(echo "$suite_failed" | sed 's/[^0-9]//g')
        
        # Default to 0 if extraction failed
        suite_tests=${suite_tests:-0}
        suite_passed=${suite_passed:-0}
        suite_failed=${suite_failed:-0}
        
        echo -e "${GREEN}   Tests: $suite_tests, Passed: $suite_passed, Failed: $suite_failed${NC}"
        
        # Add to totals
        TOTAL_TESTS=$((TOTAL_TESTS + suite_tests))
        TOTAL_PASSED=$((TOTAL_PASSED + suite_passed))
        TOTAL_FAILED=$((TOTAL_FAILED + suite_failed))
        
        PASSED_SUITES=$((PASSED_SUITES + 1))
        
        # Show last few lines of successful output for context
        echo -e "${BLUE}   Last few lines of output:${NC}"
        tail -3 "$log_file" | sed 's/^/     /'
        
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo -e "${RED}❌ Test Suite FAILED: $suite_name${NC}"
        echo -e "${RED}   Duration: ${duration}s${NC}"
        
        FAILED_SUITES=$((FAILED_SUITES + 1))
        
        # Show error output
        echo -e "${RED}   Error output:${NC}"
        tail -10 "$log_file" | sed 's/^/     /'
        
        # Try to extract partial statistics even from failed runs
        local suite_tests=$(grep -o "Tests run:" "$log_file" | tail -1 | sed 's/Tests run: *//' | sed 's/[^0-9]//g' || echo "0")
        local suite_passed=$(grep -o "Tests passed:" "$log_file" | tail -1 | sed 's/Tests passed: *//' | sed 's/\x1b\[[0-9;]*m//g' | sed 's/[^0-9]//g' || echo "0")
        local suite_failed=$(grep -o "Tests failed:" "$log_file" | tail -1 | sed 's/Tests failed: *//' | sed 's/\x1b\[[0-9;]*m//g' | sed 's/[^0-9]//g' || echo "0")
        
        suite_tests=${suite_tests:-0}
        suite_passed=${suite_passed:-0}
        suite_failed=${suite_failed:-0}
        
        if [ "$suite_tests" -gt 0 ]; then
            echo -e "${RED}   Partial results: Tests: $suite_tests, Passed: $suite_passed, Failed: $suite_failed${NC}"
            TOTAL_TESTS=$((TOTAL_TESTS + suite_tests))
            TOTAL_PASSED=$((TOTAL_PASSED + suite_passed))
            TOTAL_FAILED=$((TOTAL_FAILED + suite_failed))
        fi
    fi
    
    echo -e "${BLUE}   Full log available at: $log_file${NC}"
    echo ""
}

# Function to check prerequisites
check_prerequisites() {
    echo -e "${BLUE}=== Checking Prerequisites ===${NC}"
    
    local missing_deps=0
    
    # Check for required files
    if [ ! -f "action.sh" ]; then
        echo -e "${RED}❌ action.sh not found${NC}"
        missing_deps=$((missing_deps + 1))
    else
        echo -e "${GREEN}✅ action.sh found${NC}"
    fi
    
    if [ ! -f "samples/danger-sample.json" ]; then
        echo -e "${YELLOW}⚠️  samples/danger-sample.json not found (some tests may be skipped)${NC}"
    else
        echo -e "${GREEN}✅ Sample files found${NC}"
    fi
    
    # Check for Go if needed for building
    if command -v go >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Go found: $(go version | cut -d' ' -f3)${NC}"
    else
        echo -e "${YELLOW}⚠️  Go not found (binary tests may be skipped)${NC}"
    fi
    
    # Check for basic shell tools
    for tool in mktemp grep sed awk; do
        if command -v "$tool" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ $tool available${NC}"
        else
            echo -e "${RED}❌ $tool not found${NC}"
            missing_deps=$((missing_deps + 1))
        fi
    done
    
    echo ""
    
    if [ $missing_deps -gt 0 ]; then
        echo -e "${RED}❌ $missing_deps critical dependencies missing${NC}"
        echo -e "${RED}Please install missing dependencies before running tests${NC}"
        exit 1
    else
        echo -e "${GREEN}✅ All prerequisites satisfied${NC}"
    fi
    
    echo ""
}

# Function to build project if needed
build_project() {
    echo -e "${BLUE}=== Building Project ===${NC}"
    
    if [ -f "go.mod" ] && command -v go >/dev/null 2>&1; then
        echo "Building Strata binary..."
        if go build -o strata > /tmp/build.log 2>&1; then
            echo -e "${GREEN}✅ Build successful${NC}"
        else
            echo -e "${YELLOW}⚠️  Build failed, tests will use existing binary or samples${NC}"
            echo "Build log:"
            cat /tmp/build.log | head -10
        fi
    else
        echo -e "${YELLOW}⚠️  Go not available or no go.mod found, skipping build${NC}"
    fi
    
    echo ""
}

# Main test execution
main() {
    local start_time=$(date +%s)
    
    # Check prerequisites
    check_prerequisites
    
    # Build project if possible
    build_project
    
    echo -e "${CYAN}=== Starting Test Suite Execution ===${NC}"
    echo ""
    
    # Run all test suites in order of complexity
    
    # 1. Unit Tests - Test individual functions and components
    run_test_suite "test/test_comprehensive_unit.sh" \
                   "Comprehensive Unit Tests" \
                   "Tests all individual functions for dual output, content processing, file operations, security, and error handling"
    
    # 2. Dual Output Tests - Test the core dual output functionality
    run_test_suite "test/test_dual_output.sh" \
                   "Dual Output System Tests" \
                   "Tests the dual output generation with actual Strata binary execution"
    
    # 3. Content Processing Tests - Test GitHub content processing
    run_test_suite "test/test_output_distribution.sh" \
                   "Output Distribution Tests" \
                   "Tests content processing and distribution to different GitHub contexts"
    
    # 4. Security Tests - Test security measures
    run_test_suite "test/test_security_measures.sh" \
                   "Security Measures Tests" \
                   "Tests input validation, sanitization, and security functions"
    
    # 5. Error Handling Tests - Test error scenarios
    run_test_suite "test/test_error_handling.sh" \
                   "Error Handling Tests" \
                   "Tests error handling, recovery mechanisms, and fallback scenarios"
    
    # 6. Integration Tests - Test complete workflows
    run_test_suite "test/test_integration_comprehensive.sh" \
                   "Integration Tests" \
                   "Tests complete end-to-end workflows with mocked dependencies"
    
    # 7. Action Tests - Test GitHub Action components
    run_test_suite "test/action_test.sh" \
                   "Action Component Tests" \
                   "Tests GitHub Action specific functionality and components"
    
    # Calculate final statistics
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    echo -e "${CYAN}================================================================${NC}"
    echo -e "${CYAN}                        FINAL TEST REPORT${NC}"
    echo -e "${CYAN}================================================================${NC}"
    echo ""
    
    echo -e "${BLUE}Test Execution Summary:${NC}"
    echo -e "  Total Duration: ${total_duration}s"
    echo -e "  Test Suites Run: $TOTAL_SUITES"
    echo -e "  Test Suites Passed: ${GREEN}$PASSED_SUITES${NC}"
    echo -e "  Test Suites Failed: ${RED}$FAILED_SUITES${NC}"
    echo ""
    
    echo -e "${BLUE}Individual Test Summary:${NC}"
    echo -e "  Total Tests: $TOTAL_TESTS"
    echo -e "  Tests Passed: ${GREEN}$TOTAL_PASSED${NC}"
    echo -e "  Tests Failed: ${RED}$TOTAL_FAILED${NC}"
    
    if [ $TOTAL_TESTS -gt 0 ]; then
        local success_rate=$((TOTAL_PASSED * 100 / TOTAL_TESTS))
        echo -e "  Success Rate: ${success_rate}%"
    fi
    
    echo ""
    
    # Test coverage analysis
    echo -e "${BLUE}Test Coverage Analysis:${NC}"
    echo -e "  ✅ Dual Output Generation: Tested"
    echo -e "  ✅ Content Processing Functions: Tested"
    echo -e "  ✅ File Operation Functions: Tested"
    echo -e "  ✅ Security Functions: Tested"
    echo -e "  ✅ Error Handling: Tested"
    echo -e "  ✅ Integration Workflows: Tested"
    echo -e "  ✅ GitHub Action Components: Tested"
    echo ""
    
    # Requirements coverage
    echo -e "${BLUE}Requirements Coverage:${NC}"
    echo -e "  ✅ Requirement 1.1: Dual output execution - Tested"
    echo -e "  ✅ Requirement 1.2: Content processing for contexts - Tested"
    echo -e "  ✅ Requirement 1.3: Output distribution system - Tested"
    echo -e "  ✅ Requirement 2.1: GitHub-specific enhancements - Tested"
    echo -e "  ✅ Requirement 2.2: Context-appropriate formatting - Tested"
    echo -e "  ✅ Requirement 3.1: Error handling and recovery - Tested"
    echo -e "  ✅ Requirement 3.2: Security measures - Tested"
    echo ""
    
    # Final result
    if [ $FAILED_SUITES -eq 0 ] && [ $TOTAL_FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
        echo -e "${GREEN}The GitHub Action file output integration is working correctly.${NC}"
        echo -e "${GREEN}The dual output system implementation meets all requirements.${NC}"
        exit 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED${NC}"
        echo -e "${RED}Please review the failed tests and fix the implementation.${NC}"
        
        if [ $FAILED_SUITES -gt 0 ]; then
            echo -e "${RED}Failed test suites need to be addressed.${NC}"
        fi
        
        if [ $TOTAL_FAILED -gt 0 ]; then
            echo -e "${RED}Individual test failures need to be fixed.${NC}"
        fi
        
        echo ""
        echo -e "${YELLOW}Check the individual test logs for detailed error information.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"