#!/bin/bash

# Test script for enhanced error handling and recovery functionality
# This script tests the error handling functions implemented in task 4

set -e

# Source the action script functions (we'll need to extract them for testing)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Test configuration
TEST_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_DIR"' EXIT

# Mock GitHub environment variables for testing
export GITHUB_REPOSITORY="test/repo"
export GITHUB_WORKFLOW="test-workflow"
export GITHUB_RUN_ID="123456"
export GITHUB_SERVER_URL="https://github.com"
export GITHUB_JOB="test-job"

# Initialize test variables
MARKDOWN_CONTENT=""

echo "=== Enhanced Error Handling Tests ==="
echo "Test directory: $TEST_DIR"

# Source the modular action functions
LIB_DIR="$SCRIPT_DIR/../lib/action"

# Source all modules that contain the functions we need to test
source "$LIB_DIR/utils.sh"
source "$LIB_DIR/security.sh"
source "$LIB_DIR/files.sh"
source "$LIB_DIR/strata.sh"
source "$LIB_DIR/github.sh"

# Use the modular create_temp_file function, but ensure we're using the same TEMP_FILES array
# The create_secure_temp_file function from files.sh should work fine

# Test 1: Test handle_dual_output_error function
echo
echo "Test 1: Testing handle_dual_output_error function"
echo "================================================"

test_handle_dual_output_error() {
  local test_output="Test error output"
  local test_exit_code=1
  local test_context="test_context"
  
  echo "Testing handle_dual_output_error with exit code $test_exit_code"
  
  # Call the function (disable exit on error temporarily)
  set +e
  handle_dual_output_error $test_exit_code "$test_output" "$test_context"
  local result=$?
  set -e
  
  # Check that it returns the same exit code
  if [ $result -eq $test_exit_code ]; then
    echo "✓ Function returned correct exit code: $result"
  else
    echo "✗ Function returned wrong exit code: $result (expected: $test_exit_code)"
    return 1
  fi
  
  # Check that MARKDOWN_CONTENT was set
  if [ -n "$MARKDOWN_CONTENT" ]; then
    echo "✓ MARKDOWN_CONTENT was set"
    echo "  Content preview: $(echo "$MARKDOWN_CONTENT" | head -1)"
  else
    echo "✗ MARKDOWN_CONTENT was not set"
    return 1
  fi
  
  # Check that content contains expected elements
  if echo "$MARKDOWN_CONTENT" | grep -q "Strata Analysis Error"; then
    echo "✓ Error content contains expected header"
  else
    echo "✗ Error content missing expected header"
    return 1
  fi
  
  if echo "$MARKDOWN_CONTENT" | grep -q "$test_context"; then
    echo "✓ Error content contains context information"
  else
    echo "✗ Error content missing context information"
    return 1
  fi
  
  echo "✓ handle_dual_output_error test passed"
  return 0
}

test_handle_dual_output_error

# Test 2: Test handle_file_operation_error function
echo
echo "Test 2: Testing handle_file_operation_error function"
echo "==================================================="

test_handle_file_operation_error() {
  local test_operation="create_temp_file"
  local test_file_path="/tmp/test_file"
  local test_error_message="Test error message"
  local test_fallback_content="Test fallback content"
  
  echo "Testing handle_file_operation_error with operation: $test_operation"
  
  # Reset MARKDOWN_CONTENT
  MARKDOWN_CONTENT=""
  
  # Call the function (disable exit on error temporarily)
  set +e
  handle_file_operation_error "$test_operation" "$test_file_path" "$test_error_message" "$test_fallback_content"
  local result=$?
  set -e
  
  # Check that it returns error code
  if [ $result -eq 1 ]; then
    echo "✓ Function returned error code: $result"
  else
    echo "✗ Function returned unexpected code: $result (expected: 1)"
    return 1
  fi
  
  # Check that MARKDOWN_CONTENT was set with fallback
  if [ -n "$MARKDOWN_CONTENT" ]; then
    echo "✓ MARKDOWN_CONTENT was set"
  else
    echo "✗ MARKDOWN_CONTENT was not set"
    return 1
  fi
  
  # Check that content contains fallback content
  if echo "$MARKDOWN_CONTENT" | grep -q "$test_fallback_content"; then
    echo "✓ Error content contains fallback content"
  else
    echo "✗ Error content missing fallback content"
    return 1
  fi
  
  echo "✓ handle_file_operation_error test passed"
  return 0
}

test_handle_file_operation_error

# Test 3: Test create_structured_error_content function
echo
echo "Test 3: Testing create_structured_error_content function"
echo "======================================================="

test_create_structured_error_content() {
  local test_error_type="strata_execution_failed"
  local test_error_details="Test error details"
  local test_exit_code=2
  local test_additional_context="Test additional context"
  
  echo "Testing create_structured_error_content with type: $test_error_type"
  
  # Call the function
  local error_content
  error_content=$(create_structured_error_content "$test_error_type" "$test_error_details" "$test_exit_code" "$test_additional_context")
  
  # Check that content was generated
  if [ -n "$error_content" ]; then
    echo "✓ Error content was generated"
  else
    echo "✗ Error content was not generated"
    return 1
  fi
  
  # Check for expected elements
  if echo "$error_content" | grep -q "Strata Execution Failed"; then
    echo "✓ Error content contains expected title"
  else
    echo "✗ Error content missing expected title"
    return 1
  fi
  
  if echo "$error_content" | grep -q "$test_error_details"; then
    echo "✓ Error content contains error details"
  else
    echo "✗ Error content missing error details"
    return 1
  fi
  
  if echo "$error_content" | grep -q "$test_exit_code"; then
    echo "✓ Error content contains exit code"
  else
    echo "✗ Error content missing exit code"
    return 1
  fi
  
  if echo "$error_content" | grep -q "$test_additional_context"; then
    echo "✓ Error content contains additional context"
  else
    echo "✗ Error content missing additional context"
    return 1
  fi
  
  echo "✓ create_structured_error_content test passed"
  return 0
}

test_create_structured_error_content

# Test 4: Test cleanup_temp_files function
echo
echo "Test 4: Testing cleanup_temp_files function"
echo "==========================================="

test_cleanup_temp_files() {
  echo "Testing cleanup_temp_files function"
  
  # Create some test temporary files manually and add to array
  local temp_file1
  local temp_file2
  temp_file1=$(mktemp -t "strata_test.XXXXXXXXXX")
  temp_file2=$(mktemp -t "strata_test.XXXXXXXXXX")
  chmod 600 "$temp_file1" "$temp_file2"
  
  # Manually add to TEMP_FILES array to test cleanup
  TEMP_FILES+=("$temp_file1")
  TEMP_FILES+=("$temp_file2")
  
  if [ -z "$temp_file1" ] || [ -z "$temp_file2" ]; then
    echo "✗ Failed to create test temporary files"
    return 1
  fi
  
  echo "Created test files: $temp_file1, $temp_file2"
  
  # Verify files exist
  if [ -f "$temp_file1" ] && [ -f "$temp_file2" ]; then
    echo "✓ Test files created successfully"
  else
    echo "✗ Test files not created properly"
    return 1
  fi
  
  # Call cleanup function
  cleanup_temp_files
  local result=$?
  
  # Check return code
  if [ $result -eq 0 ]; then
    echo "✓ Cleanup function returned success"
  else
    echo "✗ Cleanup function returned error: $result"
    return 1
  fi
  
  # Verify files were removed
  if [ ! -f "$temp_file1" ] && [ ! -f "$temp_file2" ]; then
    echo "✓ Temporary files were cleaned up"
  else
    echo "✗ Temporary files were not cleaned up properly"
    return 1
  fi
  
  # Verify TEMP_FILES array was cleared
  if [ ${#TEMP_FILES[@]} -eq 0 ]; then
    echo "✓ TEMP_FILES array was cleared"
  else
    echo "✗ TEMP_FILES array was not cleared (size: ${#TEMP_FILES[@]})"
    return 1
  fi
  
  echo "✓ cleanup_temp_files test passed"
  return 0
}

test_cleanup_temp_files

# Test 5: Test validate_file_path function
echo
echo "Test 5: Testing validate_file_path function"
echo "==========================================="

test_validate_file_path() {
  echo "Testing validate_file_path function"
  
  # Test valid path
  local valid_path="/tmp/valid_file.txt"
  set +e
  validate_file_path "$valid_path" "temp_file"
  local result1=$?
  set -e
  if [ $result1 -eq 0 ]; then
    echo "✓ Valid path accepted: $valid_path"
  else
    echo "✗ Valid path rejected: $valid_path"
    return 1
  fi
  
  # Test path traversal attempt
  local invalid_path="/tmp/../etc/passwd"
  set +e
  validate_file_path "$invalid_path" "temp_file"
  local result2=$?
  set -e
  if [ $result2 -ne 0 ]; then
    echo "✓ Path traversal attempt rejected: $invalid_path"
  else
    echo "✗ Path traversal attempt accepted: $invalid_path"
    return 1
  fi
  
  # Test empty path
  set +e
  validate_file_path "" "temp_file"
  local result3=$?
  set -e
  if [ $result3 -ne 0 ]; then
    echo "✓ Empty path rejected"
  else
    echo "✗ Empty path accepted"
    return 1
  fi
  
  echo "✓ validate_file_path test passed"
  return 0
}

test_validate_file_path

# Test 6: Integration test - Error handling in realistic scenario
echo
echo "Test 6: Integration test - Error handling flow"
echo "=============================================="

test_error_handling_integration() {
  echo "Testing integrated error handling flow"
  
  # Simulate a scenario where temp file creation fails
  # We'll override the create_temp_file function to fail
  create_temp_file() {
    echo "Simulated temp file creation failure" >&2
    return 1
  }
  
  # Test the error handling flow
  local test_stdout="Test stdout output"
  
  # This should trigger the file operation error handling (disable exit on error temporarily)
  set +e
  handle_file_operation_error "create_temp_file" "N/A" "mktemp failed" "$test_stdout"
  local result=$?
  set -e
  
  if [ $result -eq 1 ]; then
    echo "✓ Error handling returned expected error code"
  else
    echo "✗ Error handling returned unexpected code: $result"
    return 1
  fi
  
  if [ -n "$MARKDOWN_CONTENT" ]; then
    echo "✓ Fallback content was generated"
    if echo "$MARKDOWN_CONTENT" | grep -q "$test_stdout"; then
      echo "✓ Fallback content includes original stdout"
    else
      echo "✗ Fallback content missing original stdout"
      return 1
    fi
  else
    echo "✗ No fallback content generated"
    return 1
  fi
  
  echo "✓ Integration test passed"
  return 0
}

test_error_handling_integration

echo
echo "=== All Error Handling Tests Completed ==="
echo "✓ All tests passed successfully!"
echo
echo "Enhanced error handling implementation verified:"
echo "- handle_dual_output_error function works correctly"
echo "- handle_file_operation_error provides proper fallbacks"
echo "- create_structured_error_content generates appropriate content"
echo "- cleanup_temp_files properly cleans up resources"
echo "- validate_file_path provides security validation"
echo "- Integration scenarios work as expected"