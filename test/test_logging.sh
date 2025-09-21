#!/bin/bash

# Test script to verify the comprehensive logging functionality
# This tests the new logging functions added for task 6

# Source the functions from action.sh (extract just the logging functions)
source <(grep -A 50 "log_file_output_status()" action.sh)

# Test the log_file_output_status function
echo "Testing log_file_output_status function..."

echo "=== Testing DISABLED status ==="
log_file_output_status "disabled" "temporary file creation failed" "mktemp command failed"

echo ""
echo "=== Testing FAILED status ==="
log_file_output_status "failed" "file read operation failed" "Permission denied on temporary file"

echo ""
echo "=== Testing PARTIAL status ==="
log_file_output_status "partial" "some file operations succeeded" "Markdown generated but optimization failed"

echo ""
echo "=== Testing SUCCESS status ==="
log_file_output_status "success" "dual output generation completed" "Both stdout and file formats available"

echo ""
echo "=== Testing UNKNOWN status ==="
log_file_output_status "unknown" "unexpected condition" "Status could not be determined"

echo ""
echo "Logging function tests completed successfully!"