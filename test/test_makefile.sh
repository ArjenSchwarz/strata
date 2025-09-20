#!/bin/bash
# Test script for Makefile targets

# Set up colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to run a test and report result
run_test() {
    local test_name=$1
    local command=$2
    
    echo -n "Testing $test_name... "
    
    # Run the command and capture output and exit code
    output=$(eval $command 2>&1)
    exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}PASS${NC}"
    else
        echo -e "${RED}FAIL${NC}"
        echo "Command: $command"
        echo "Output: $output"
        echo "Exit code: $exit_code"
        failures=$((failures + 1))
    fi
}

# Initialize failure counter
failures=0

# Test build target
run_test "build target" "make build"

# Test test target
run_test "test target" "make test"

# Test run-sample target with missing parameter
run_test "run-sample missing parameter" "make run-sample 2>&1 | grep -q 'SAMPLE parameter is required'"

# Test run-sample target with non-existent file
run_test "run-sample non-existent file" "make run-sample SAMPLE=nonexistent.json 2>&1 | grep -q 'not found'"

# Test run-sample target with valid file
run_test "run-sample with valid file" "make run-sample SAMPLE=danger-sample.json"

# Test run-sample-details target with valid file
run_test "run-sample-details with valid file" "make run-sample-details SAMPLE=k8ssample.json"

# Report summary
echo ""
if [ $failures -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}$failures test(s) failed.${NC}"
    exit 1
fi