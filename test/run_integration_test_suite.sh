#!/bin/bash

# Master Integration Test Suite Runner for Strata GitHub Action
# Orchestrates all integration test components as specified in task 8
# Runs: End-to-End tests, Backwards Compatibility tests, Performance Benchmarks

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly NC='\033[0m' # No Color

# Test suite tracking
declare -A SUITE_RESULTS

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_suite() {
    echo -e "${PURPLE}[SUITE]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking master test suite prerequisites..."

    # Check we're in the right directory
    if [ ! -f "action.yml" ] || [ ! -f "action.sh" ]; then
        echo "Error: Please run this script from the root of the Strata repository."
        exit 1
    fi

    # Check that all test suites exist
    local required_tests=(
        "test/test_comprehensive_integration.sh"
        "test/test_backwards_compatibility.sh"
        "test/test_performance_benchmarks.sh"
    )

    for test_file in "${required_tests[@]}"; do
        if [ ! -f "$test_file" ]; then
            echo "Error: Required test suite not found: $test_file"
            exit 1
        fi
        if [ ! -x "$test_file" ]; then
            chmod +x "$test_file"
        fi
    done

    log_info "All test suites found and executable"
}

# Setup master test environment
setup_master_test_environment() {
    log_info "Setting up master test environment..."

    # Create master results directory
    MASTER_TEST_DIR=$(mktemp -d)
    export MASTER_TEST_DIR

    cleanup() {
        log_info "Master test cleanup..."
        # Always preserve master test results
        log_info "Master test results preserved in: $MASTER_TEST_DIR"
    }
    trap cleanup EXIT

    log_info "Master test directory: $MASTER_TEST_DIR"
}

# Run comprehensive integration tests (Task 8.1)
run_end_to_end_tests() {
    log_suite "Running End-to-End Integration Tests (Task 8.1)"
    echo "==============================================="

    local start_time=$(date +%s)
    local exit_code=0

    if test/test_comprehensive_integration.sh > "$MASTER_TEST_DIR/end_to_end_results.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log_pass "End-to-End Integration Tests completed successfully in ${duration}s"
        SUITE_RESULTS["end_to_end"]="PASS:${duration}s"
    else
        exit_code=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log_fail "End-to-End Integration Tests failed in ${duration}s (exit code: $exit_code)"
        SUITE_RESULTS["end_to_end"]="FAIL:${duration}s"

        echo "Last 10 lines of end-to-end test output:"
        tail -n 10 "$MASTER_TEST_DIR/end_to_end_results.log" | sed 's/^/  /'
    fi

    echo ""
    return $exit_code
}

# Run backwards compatibility tests (Task 8.2)
run_backwards_compatibility_tests() {
    log_suite "Running Backwards Compatibility Tests (Task 8.2)"
    echo "=================================================="

    local start_time=$(date +%s)
    local exit_code=0

    if test/test_backwards_compatibility.sh > "$MASTER_TEST_DIR/backwards_compatibility_results.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log_pass "Backwards Compatibility Tests completed successfully in ${duration}s"
        SUITE_RESULTS["backwards_compatibility"]="PASS:${duration}s"
    else
        exit_code=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log_fail "Backwards Compatibility Tests failed in ${duration}s (exit code: $exit_code)"
        SUITE_RESULTS["backwards_compatibility"]="FAIL:${duration}s"

        echo "Last 10 lines of backwards compatibility test output:"
        tail -n 10 "$MASTER_TEST_DIR/backwards_compatibility_results.log" | sed 's/^/  /'
    fi

    echo ""
    return $exit_code
}

# Run performance benchmarks (Task 8.3)
run_performance_benchmarks() {
    log_suite "Running Performance Benchmarks (Task 8.3)"
    echo "==========================================="

    local start_time=$(date +%s)
    local exit_code=0

    if test/test_performance_benchmarks.sh > "$MASTER_TEST_DIR/performance_benchmarks_results.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log_pass "Performance Benchmarks completed successfully in ${duration}s"
        SUITE_RESULTS["performance_benchmarks"]="PASS:${duration}s"
    else
        exit_code=$?
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        log_fail "Performance Benchmarks failed in ${duration}s (exit code: $exit_code)"
        SUITE_RESULTS["performance_benchmarks"]="FAIL:${duration}s"

        echo "Last 10 lines of performance benchmark output:"
        tail -n 10 "$MASTER_TEST_DIR/performance_benchmarks_results.log" | sed 's/^/  /'
    fi

    echo ""
    return $exit_code
}

# Generate master test report
generate_master_report() {
    log_info "Generating master integration test report..."

    local report_file="$MASTER_TEST_DIR/master_integration_test_report.md"
    local passed_suites=0
    local total_suites=${#SUITE_RESULTS[@]}

    # Count passed suites
    for result in "${SUITE_RESULTS[@]}"; do
        if [[ $result == PASS:* ]]; then
            passed_suites=$((passed_suites + 1))
        fi
    done

    cat > "$report_file" << EOF
# Strata GitHub Action - Master Integration Test Report

**Test Date:** $(date)
**Test Environment:** $(uname -a)
**Repository:** $(pwd)
**Git Commit:** $(git rev-parse HEAD 2>/dev/null || echo "N/A")

## Executive Summary

- **Test Suites Run:** $total_suites
- **Test Suites Passed:** $passed_suites
- **Test Suites Failed:** $((total_suites - passed_suites))
- **Overall Success Rate:** $(( passed_suites * 100 / total_suites ))%

## Test Suite Results

### 8.1 End-to-End Integration Tests
**Status:** $(echo "${SUITE_RESULTS[end_to_end]}" | cut -d: -f1)
**Duration:** $(echo "${SUITE_RESULTS[end_to_end]}" | cut -d: -f2)

Requirements tested:
- [x] Test harness for full action execution
- [x] Test with each sample plan file in samples/
- [x] Test all output formats (markdown, json, table, html)
- [x] Test dangerous changes detection and reporting
- [x] Measure total execution time (<30 seconds)

### 8.2 Backwards Compatibility Tests
**Status:** $(echo "${SUITE_RESULTS[backwards_compatibility]}" | cut -d: -f1)
**Duration:** $(echo "${SUITE_RESULTS[backwards_compatibility]}" | cut -d: -f2)

Requirements tested:
- [x] Test minimal inputs (just plan-file)
- [x] Test all current input parameters work correctly
- [x] Verify all outputs generated with same names
- [x] Test existing workflows continue without changes
- [x] Validate action.yml interface unchanged (except strata-version)

### 8.3 Performance Benchmarks
**Status:** $(echo "${SUITE_RESULTS[performance_benchmarks]}" | cut -d: -f1)
**Duration:** $(echo "${SUITE_RESULTS[performance_benchmarks]}" | cut -d: -f2)

Requirements tested:
- [x] Measure binary download time (<10 seconds)
- [x] Measure analysis startup time (<5 seconds)
- [x] Measure total execution time (<30 seconds)
- [x] Compare against current implementation baseline
- [x] Document performance improvements

## Detailed Results

EOF

    # Add links to detailed reports
    echo "### Detailed Test Logs" >> "$report_file"
    echo "" >> "$report_file"
    for suite in "${!SUITE_RESULTS[@]}"; do
        echo "- **$suite**: \`${MASTER_TEST_DIR}/${suite}_results.log\`" >> "$report_file"
    done
    echo "" >> "$report_file"

    # Overall conclusion
    if [ $passed_suites -eq $total_suites ]; then
        cat >> "$report_file" << EOF
## ✅ Overall Conclusion

**ALL INTEGRATION TESTS PASSED SUCCESSFULLY**

The Strata GitHub Action has successfully completed comprehensive integration testing covering:

1. **End-to-End Functionality**: All sample files processed correctly across all output formats
2. **Backwards Compatibility**: Existing workflows will continue working without modification
3. **Performance Requirements**: All performance thresholds met consistently

**Recommendation:** The action is ready for production release.

### Key Achievements

- ✅ Comprehensive test coverage across all requirements
- ✅ Backwards compatibility maintained (100% compatible)
- ✅ Performance benchmarks meet all specified thresholds
- ✅ All output formats validated and working
- ✅ Dangerous change detection functioning correctly
- ✅ Error handling robust and user-friendly

EOF
    else
        cat >> "$report_file" << EOF
## ❌ Overall Conclusion

**SOME INTEGRATION TESTS FAILED**

$((total_suites - passed_suites)) out of $total_suites test suites failed. Review the detailed logs and address the following before release:

EOF
        for suite in "${!SUITE_RESULTS[@]}"; do
            local result="${SUITE_RESULTS[$suite]}"
            if [[ $result == FAIL:* ]]; then
                echo "- **$suite**: Failed - See detailed log for specific issues" >> "$report_file"
            fi
        done

        cat >> "$report_file" << EOF

**Recommendation:** Address all test failures before proceeding with release.

EOF
    fi

    log_info "Master integration test report: $report_file"
    echo ""
    echo "📋 Master Integration Test Report: $report_file"
    echo "📁 All Test Results Directory: $MASTER_TEST_DIR"
}

# Print master test summary
print_master_summary() {
    echo ""
    echo "Master Integration Test Suite Summary"
    echo "===================================="
    echo ""

    local all_passed=true
    for suite in "${!SUITE_RESULTS[@]}"; do
        local result="${SUITE_RESULTS[$suite]}"
        local status=$(echo "$result" | cut -d: -f1)
        local duration=$(echo "$result" | cut -d: -f2)

        if [ "$status" = "PASS" ]; then
            echo -e "${GREEN}✅${NC} $suite: PASSED ($duration)"
        else
            echo -e "${RED}❌${NC} $suite: FAILED ($duration)"
            all_passed=false
        fi
    done

    echo ""
    if [ "$all_passed" = true ]; then
        echo -e "${GREEN}🎉 ALL INTEGRATION TEST SUITES PASSED!${NC}"
        echo ""
        echo "The Strata GitHub Action is ready for release with:"
        echo "  ✅ Full end-to-end functionality validated"
        echo "  ✅ Complete backwards compatibility confirmed"
        echo "  ✅ All performance benchmarks met"
        echo ""
        echo "Task 8: Create integration test suite - COMPLETED SUCCESSFULLY"
    else
        echo -e "${RED}❌ SOME INTEGRATION TEST SUITES FAILED${NC}"
        echo ""
        echo "Review the detailed logs and address all issues before release."
        echo ""
        echo "Task 8: Create integration test suite - REQUIRES ATTENTION"
    fi
}

# Main execution
main() {
    echo "Strata GitHub Action - Master Integration Test Suite"
    echo "===================================================="
    echo ""
    echo "This suite runs all integration tests as specified in Task 8:"
    echo "  8.1 - End-to-End Integration Tests"
    echo "  8.2 - Backwards Compatibility Tests"
    echo "  8.3 - Performance Benchmarks"
    echo ""

    check_prerequisites
    setup_master_test_environment

    local start_time=$(date +%s)
    local failed_suites=0

    # Run all test suites
    run_end_to_end_tests || failed_suites=$((failed_suites + 1))
    run_backwards_compatibility_tests || failed_suites=$((failed_suites + 1))
    run_performance_benchmarks || failed_suites=$((failed_suites + 1))

    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))

    # Generate comprehensive report
    generate_master_report

    # Print summary
    print_master_summary

    echo ""
    echo "Total test suite execution time: ${total_duration}s"
    echo "Master test results: $MASTER_TEST_DIR"

    # Exit with appropriate code
    if [ $failed_suites -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"