#!/bin/bash

# Strata execution functions for GitHub Action
# This module handles running Strata commands and processing output

# Global variable to store markdown content from dual output
MARKDOWN_CONTENT=""

# Function to run Strata with specified parameters (renamed to avoid conflict)
run_strata_ORIGINAL() {
  local output_format=$1
  local plan_file=$2
  local show_details=$3
  
  # Convert plan file to absolute path if it's relative
  if [[ "$plan_file" != /* ]]; then
    plan_file="$(pwd)/$plan_file"
  fi
  
  local cmd="$TEMP_DIR/$BINARY_NAME plan summary"
  
  # Add optional arguments
  if [ -n "$INPUT_CONFIG_FILE" ]; then
    # Convert config file to absolute path if it's relative
    local config_file="$INPUT_CONFIG_FILE"
    if [[ "$config_file" != /* ]]; then
      config_file="$(pwd)/$config_file"
    fi
    cmd="$cmd --config $config_file"
  fi
  
  if [ -n "$INPUT_DANGER_THRESHOLD" ]; then
    cmd="$cmd --danger-threshold $INPUT_DANGER_THRESHOLD"
  fi
  
  if [ "$show_details" = "true" ]; then
    cmd="$cmd --details"
  fi
  
  # Add output format
  cmd="$cmd --output $output_format"
  
  # Add plan file
  cmd="$cmd $plan_file"
  
  log "Running Strata with parameters" "Format=$output_format Details=$show_details"
  
  # Execute command and capture output
  local output
  output=$(eval "$cmd" 2>&1)
  local exit_code=$?
  
  if [ $exit_code -ne 0 ]; then
    warning "Strata execution failed with exit code $exit_code"
    warning "Error output: $output"
  fi
  
  echo "$output"
  return $exit_code
}

# Function to create structured error content for GitHub features
create_structured_error_content() {
  local error_type=$1
  local error_details="$2"
  local exit_code="${3:-1}"
  local additional_context="$4"
  
  log "Creating structured error content" "Type: $error_type, Exit code: $exit_code"
  
  local error_content=""
  
  case "$error_type" in
    "strata_execution_failed")
      error_content="## ⚠️ Strata Execution Failed

The Terraform plan analysis could not be completed.

**Exit Code:** $exit_code

**Error Output:**
\`\`\`
$error_details
\`\`\`

**Common Solutions:**
- Verify the plan file is valid and readable
- Check Terraform version compatibility
- Ensure sufficient disk space and permissions
- Review the plan file format and structure

$additional_context"
      ;;
    "binary_download_failed")
      error_content="## ⚠️ Binary Download Failed

Could not download or prepare the Strata binary.

**Error Details:**
\`\`\`
$error_details
\`\`\`

**Possible Solutions:**
- Check network connectivity
- Verify GitHub releases are accessible
- Try running the action again
- Check if the repository has the latest releases

$additional_context"
      ;;
    "file_operation_failed")
      error_content="## ⚠️ File Operation Failed

A file operation required for dual output processing failed.

**Error Details:**
\`\`\`
$error_details
\`\`\`

**Possible Solutions:**
- Check disk space availability
- Verify file system permissions
- Ensure temporary directory is writable
- Check for file system corruption

$additional_context"
      ;;
    "format_conversion_failed")
      error_content="## ⚠️ Format Conversion Failed

Could not convert output between different formats.

**Error Details:**
\`\`\`
$error_details
\`\`\`

**Possible Solutions:**
- Verify the source format is valid
- Check for corrupted input data
- Try with a different output format
- Review the plan file structure

$additional_context"
      ;;
    "github_api_failed")
      error_content="## ⚠️ GitHub API Operation Failed

Could not complete GitHub API operations (PR comments, etc.).

**Error Details:**
\`\`\`
$error_details
\`\`\`

**Possible Solutions:**
- Check GitHub token permissions
- Verify repository access rights
- Check for API rate limiting
- Ensure the PR/issue exists and is accessible

$additional_context"
      ;;
    *)
      error_content="## ⚠️ Unknown Error

An unexpected error occurred during plan analysis.

**Error Type:** $error_type  
**Exit Code:** $exit_code

**Error Details:**
\`\`\`
$error_details
\`\`\`

$additional_context"
      ;;
  esac
  
  # Add common footer with workflow information
  error_content="$error_content

---
**Workflow Information:**
- Repository: $GITHUB_REPOSITORY
- Workflow: $GITHUB_WORKFLOW  
- Run ID: $GITHUB_RUN_ID
- Timestamp: $(date -u '+%Y-%m-%d %H:%M:%S UTC')

For more details, check the [workflow run logs](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID})."
  
  echo "$error_content"
}

# Function to handle dual output errors gracefully
handle_dual_output_error() {
  local exit_code=$1
  local stdout_output="$2"
  local error_context="${3:-dual_output_execution}"
  
  log "Handling dual output error" "Exit code: $exit_code, Context: $error_context"
  
  # Create structured error content for GitHub features
  local error_content="## ⚠️ Strata Analysis Error

Strata encountered an issue while analyzing the Terraform plan.

**Error Context:** $error_context  
**Exit Code:** $exit_code

**Error Details:**
\`\`\`
$stdout_output
\`\`\`

**Possible causes:**
- Invalid plan file format
- Unsupported Terraform version
- Plan file corruption
- File permission issues
- Insufficient disk space
- Network connectivity issues (if downloading binary)

**Troubleshooting steps:**
1. Verify the plan file was generated successfully
2. Check file permissions and disk space
3. Ensure Terraform version compatibility
4. Review action logs for additional details

Please check the [workflow run](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}) for complete logs."
  
  # Set fallback markdown content
  MARKDOWN_CONTENT="$error_content"
  
  # Log error handling completion
  log "Error content generated" "Size: ${#error_content} chars"
  
  # Ensure cleanup happens even in error scenarios
  cleanup_temp_files
  
  return "$exit_code"
}

# Function to handle file operation errors with fallback mechanisms
handle_file_operation_error() {
  local operation=$1
  local file_path=$2
  local error_message="$3"
  local fallback_content="$4"
  
  warning "File operation failed: $operation on $file_path - $error_message"
  log "Implementing fallback mechanism" "Operation: $operation"
  
  case "$operation" in
    "create_temp_file")
      warning "Could not create temporary file, falling back to stdout-only mode"
      log "File output disabled" "Reason: temporary file creation failed"
      
      # Set fallback markdown content
      if [ -n "$fallback_content" ]; then
        MARKDOWN_CONTENT="## Terraform Plan Summary

> **Note:** Dual output mode unavailable - using single output format.

$fallback_content"
      else
        MARKDOWN_CONTENT="## ⚠️ File Operation Error

Could not create temporary file for dual output processing.

**Error:** $error_message

Continuing with stdout-only mode."
      fi
      return 1
      ;;
    "write_temp_file")
      warning "Could not write to temporary file, using fallback content"
      log "File write failed" "Path: $file_path"
      
      if [ -n "$fallback_content" ]; then
        MARKDOWN_CONTENT="## Terraform Plan Summary

> **Note:** File write operation failed - using fallback content.

$fallback_content"
      else
        MARKDOWN_CONTENT="## ⚠️ File Write Error

Could not write to temporary file: $file_path

**Error:** $error_message"
      fi
      return 1
      ;;
    "read_temp_file")
      warning "Could not read temporary file, using stdout content as fallback"
      log "File read failed" "Path: $file_path"
      
      if [ -n "$fallback_content" ]; then
        MARKDOWN_CONTENT="$fallback_content"
      else
        MARKDOWN_CONTENT="## ⚠️ File Read Error

Could not read temporary file: $file_path

**Error:** $error_message"
      fi
      return 1
      ;;
    *)
      warning "Unknown file operation error: $operation"
      if [ -n "$fallback_content" ]; then
        MARKDOWN_CONTENT="$fallback_content"
      else
        MARKDOWN_CONTENT="## ⚠️ Unknown File Operation Error

An unknown file operation error occurred.

**Operation:** $operation  
**File:** $file_path  
**Error:** $error_message"
      fi
      return 1
      ;;
  esac
}

# Enhanced function to run Strata with dual output (renamed to run_strata for debugging)
run_strata() {
  echo "##[warning]DEBUG: IMMEDIATE FUNCTION ENTRY - run_strata called (this is actually the dual output function!)"
  local stdout_format=$1
  local plan_file=$2
  local show_details=$3
  
  # DEBUGGING: Add very visible marker to confirm function is called
  echo "##[warning]DEBUG: run_strata_dual_output function ENTRY POINT"
  
  # Comprehensive logging for dual output initialization
  log "Initializing dual output execution" "Display format: $stdout_format, File format: markdown, Show details: $show_details"
  log "Dual output configuration" "Stdout: $stdout_format (for terminal display), File: markdown (for GitHub features)"
  log "Plan file path" "$plan_file"
  
  # Validate plan file path
  if ! validate_file_path "$plan_file" "plan_file"; then
    local error_msg="Invalid plan file path: $plan_file"
    warning "$error_msg"
    handle_dual_output_error 1 "$error_msg" "plan_file_validation"
    return 1
  fi
  
  # Convert plan file to absolute path if it's relative
  if [[ "$plan_file" != /* ]]; then
    plan_file="$(pwd)/$plan_file"
  fi
  
  # Create secure temporary file for markdown output with enhanced error handling
  log "Creating secure temporary file for markdown output" "Purpose: dual output file generation"
  local temp_markdown_file
  temp_markdown_file=$(create_secure_temp_file "markdown_output")
  local temp_file_result=$?
  
  if [ $temp_file_result -ne 0 ]; then
    # Fallback to single output mode with proper error handling
    local fallback_output
    fallback_output=$(run_strata_ORIGINAL "$stdout_format" "$plan_file" "$show_details")
    local fallback_exit_code=$?
    
    # Use file operation error handler
    handle_file_operation_error "create_temp_file" "N/A" "mktemp failed" "$fallback_output"
    
    log "Fallback to single output completed" "Format: $stdout_format, Exit code: $fallback_exit_code"
    echo "$fallback_output"
    return $fallback_exit_code
  fi
  
  # Validate temporary file path
  if ! validate_file_path "$temp_markdown_file" "temp_file"; then
    local error_msg="Invalid temporary file path: $temp_markdown_file"
    warning "$error_msg"
    handle_file_operation_error "validate_temp_file" "$temp_markdown_file" "$error_msg" ""
    
    # Fallback to single output
    local fallback_output
    fallback_output=$(run_strata_ORIGINAL "$stdout_format" "$plan_file" "$show_details")
    echo "$fallback_output"
    return $?
  fi
  
  # DEBUGGING: Confirm we reach command building phase
  echo "##[warning]DEBUG: Starting command building phase in run_strata_dual_output"
  
  # Start building command with binary name
  local cmd="$TEMP_DIR/$BINARY_NAME"
  
  # Add optional global config argument
  if [ -n "$INPUT_CONFIG_FILE" ]; then
    local config_file="$INPUT_CONFIG_FILE"
    if [[ "$config_file" != /* ]]; then
      config_file="$(pwd)/$config_file"
    fi
    cmd="$cmd --config $config_file"
  fi
  
  # Check if dual output is supported by testing the --file flag
  if "$TEMP_DIR/$BINARY_NAME" --help 2>&1 | grep -q -- "--file"; then
    log "Dual output supported" "Using --file flag for markdown output"
    # Add global file output flags
    cmd="$cmd --file $temp_markdown_file --file-format markdown"
  else
    log "Dual output not supported" "Falling back to single output mode"
  fi
  
  # Add the subcommand
  cmd="$cmd plan summary"
  
  # Add subcommand-specific arguments
  cmd="$cmd --output $stdout_format"
  
  if [ -n "$INPUT_DANGER_THRESHOLD" ]; then
    cmd="$cmd --danger-threshold $INPUT_DANGER_THRESHOLD"
  fi
  
  if [ "$show_details" = "true" ]; then
    cmd="$cmd --details"
  fi
  
  # Add plan file
  cmd="$cmd $plan_file"
  
  log "Executing Strata with dual output" "Command: $cmd"
  
  # DEBUGGING: Add marker before command execution
  echo "##[warning]DEBUG: About to execute command in run_strata_dual_output"
  
  # Display the full command for debugging
  echo "::group::Strata Command"
  echo "Executing: $cmd"
  echo "::endgroup::"
  
  # Execute command and capture stdout with enhanced error handling
  log "Executing Strata command" "Starting dual output execution"
  
  # DEBUGGING: Add marker before real-time output
  echo "##[warning]DEBUG: Starting real-time output capture"
  
  # Execute command and capture output with error handling
  echo "::group::Strata Real-time Output"
  local stdout_output
  stdout_output=$(eval "$cmd" 2>&1)
  local exit_code=$?
  
  # Show the actual output immediately
  if [ -n "$stdout_output" ]; then
    echo "$stdout_output"
  else
    echo "No output produced by command"
  fi
  echo "::endgroup::"
  
  # Log execution results
  log "Strata execution completed" "Exit code: $exit_code, Output size: ${#stdout_output} chars"
  
  # Handle execution errors with structured error content
  if [ $exit_code -ne 0 ]; then
    warning "Strata execution failed with exit code $exit_code"
    warning "Error output: $stdout_output"
    log "Dual output execution failed" "Both stdout and file output affected"
    
    handle_dual_output_error $exit_code "$stdout_output" "main_execution"
    
    echo "$stdout_output"
    return $exit_code
  fi
  
  # Handle file operations with comprehensive error checking
  if [ -f "$temp_markdown_file" ]; then
    # Check if file is readable
    if [ -r "$temp_markdown_file" ]; then
      # Read markdown content with size validation and error handling
      local file_size
      file_size=$(wc -c < "$temp_markdown_file" 2>/dev/null)
      local size_check_result=$?
      
      if [ $size_check_result -eq 0 ] && [ "$file_size" -gt 0 ]; then
        # Attempt to read file content
        MARKDOWN_CONTENT=$(cat "$temp_markdown_file" 2>/dev/null)
        local read_result=$?
        
        if [ $read_result -eq 0 ] && [ -n "$MARKDOWN_CONTENT" ]; then
          log "Successfully generated markdown content" "Size: $file_size bytes"
        else
          # Handle read failure
          handle_file_operation_error "read_temp_file" "$temp_markdown_file" "cat command failed (exit code: $read_result)" "$stdout_output"
        fi
      else
        # Handle size check or empty file
        if [ $size_check_result -ne 0 ]; then
          handle_file_operation_error "read_temp_file" "$temp_markdown_file" "wc command failed (exit code: $size_check_result)" "$stdout_output"
        else
          handle_file_operation_error "read_temp_file" "$temp_markdown_file" "file is empty (size: $file_size)" "$stdout_output"
        fi
      fi
    else
      # Handle unreadable file
      handle_file_operation_error "read_temp_file" "$temp_markdown_file" "file is not readable" "$stdout_output"
    fi
  else
    # Handle missing file - this is expected when dual output is not supported
    log "Format conversion status" "SINGLE OUTPUT MODE - using stdout content for markdown"
    
    # Convert stdout output to markdown format if needed
    if [ "$stdout_format" = "markdown" ]; then
      MARKDOWN_CONTENT="$stdout_output"
      log "Markdown content set from stdout" "Direct markdown output"
    else
      # Convert table/json output to basic markdown
      MARKDOWN_CONTENT="## Terraform Plan Summary

\`\`\`
$stdout_output
\`\`\`"
      log "Markdown content generated from stdout" "Wrapped in code block"
    fi
  fi
  
  # Final validation that we have content
  if [ -z "$MARKDOWN_CONTENT" ]; then
    warning "No markdown content available after all error handling, using stdout as final fallback"
    MARKDOWN_CONTENT="$stdout_output"
  fi
  
  log "Dual output execution completed"
  
  echo "$stdout_output"
  return $exit_code
}