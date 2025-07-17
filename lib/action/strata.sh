#!/bin/bash

# Strata execution functions for GitHub Action
# This module handles running Strata commands and processing output

# Global variable to store markdown content from dual output
MARKDOWN_CONTENT=""


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



# Function to run Strata with dual output
run_strata() {
  echo "##[warning]DEBUG: IMMEDIATE FUNCTION ENTRY - run_strata called"
  log "[warning]DEBUG: FUNCTION ENTRY" "called"
  local stdout_format=$1
  local plan_file=$2
  local show_details=$3
  
  # DEBUGGING: Add very visible marker to confirm function is called
  echo "##[warning]DEBUG: run_strata function ENTRY POINT"
  
  # Comprehensive logging for dual output initialization
  log "Initializing dual output execution" "Display format: $stdout_format, File format: markdown, Show details: $show_details"
  log "Dual output configuration" "Stdout: $stdout_format (for terminal display), File: markdown (for GitHub features)"
  log "Plan file path" "$plan_file"
  
  # Validate plan file path
  if ! validate_file_path "$plan_file" "plan_file"; then
    local error_msg="Invalid plan file path: $plan_file"
    warning "$error_msg"
    error "$error_msg"
    return 1
  fi
  
  # Convert plan file to absolute path if it's relative
  # Use GITHUB_WORKSPACE as the base directory in GitHub Actions environment
  if [[ "$plan_file" != /* ]]; then
    if [ -n "$GITHUB_WORKSPACE" ]; then
      plan_file="$GITHUB_WORKSPACE/$plan_file"
    else
      plan_file="$(pwd)/$plan_file"
    fi
  fi
  
  # Use a predictable markdown file in the workspace instead of temp file
  local markdown_file
  if [ -n "$GITHUB_WORKSPACE" ]; then
    markdown_file="$GITHUB_WORKSPACE/strata-run-output.md"
  else
    markdown_file="$(pwd)/strata-run-output.md"
  fi
  
  log "Using workspace markdown file" "Path: $markdown_file"
  
  # Validate markdown file path
  if ! validate_file_path "$markdown_file" "workspace_file"; then
    error "Invalid markdown file path: $markdown_file"
    return 1
  fi
  
  # DEBUGGING: Confirm we reach command building phase
  log "DEBUG: Starting command building phase" "in run_strata function"
  
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
  log "DEBUG: Testing for --file flag support" "Checking Strata binary capabilities"
  if "$TEMP_DIR/$BINARY_NAME" --help 2>&1 | grep -q -- "--file"; then
    log "Dual output supported" "Using --file flag for markdown output"
    # Add global file output flags
    cmd="$cmd --file $markdown_file --file-format markdown"
    log "DEBUG: Added file output flags" "--file $markdown_file --file-format markdown"
  else
    log "Dual output not supported" "Falling back to single output mode"
    log "DEBUG: Help output check" "$(\"$TEMP_DIR/$BINARY_NAME\" --help 2>&1 | head -10)"
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
  log "DEBUG: Final command breakdown" "Binary: $TEMP_DIR/$BINARY_NAME, Markdown file: $markdown_file"
  
  # DEBUGGING: Add marker before command execution
  log "DEBUG: About to execute command" "in run_strata function"
  
  # Display the full command for debugging
  echo "::group::Strata Command"
  echo "Executing: $cmd"
  echo "::endgroup::"
  
  # Execute command and capture stdout with enhanced error handling
  log "Executing Strata command" "Starting dual output execution"
  
  # DEBUGGING: Add marker before real-time output
  log "DEBUG: Starting real-time output capture" "about to execute Strata command"
  
  # Execute command and capture output with error handling
  echo "::group::Strata Real-time Output"
  local stdout_output
  stdout_output=$(eval "$cmd" 2>&1)
  local exit_code=$?
  
  # Show the actual output immediately
  if [ -n "$stdout_output" ]; then
    echo "$stdout_output"
    log "DEBUG: Strata produced output" "Size: ${#stdout_output} chars"
  else
    echo "No output produced by command"
    log "DEBUG: No output from Strata" "Command may have failed silently"
  fi
  echo "::endgroup::"
  
  # Log execution results
  log "Strata execution completed" "Exit code: $exit_code, Output size: ${#stdout_output} chars"
  
  # Handle execution errors with structured error content
  if [ $exit_code -ne 0 ]; then
    warning "Strata execution failed with exit code $exit_code"
    warning "Error output: $stdout_output"
    log "Dual output execution failed" "Both stdout and file output affected"
    
    warning "Strata execution failed with exit code $exit_code"
    warning "Error output: $stdout_output"
    
    echo "$stdout_output"
    return $exit_code
  fi
  
  # Handle file operations with comprehensive error checking
  log "DEBUG: Checking for markdown file" "Path: $markdown_file"
  if [ -f "$markdown_file" ]; then
    log "DEBUG: Markdown file exists" "Checking file contents"
    # Check if file is readable
    if [ -r "$markdown_file" ]; then
      # Read markdown content with size validation and error handling
      local file_size
      file_size=$(wc -c < "$markdown_file" 2>/dev/null)
      local size_check_result=$?
      
      log "DEBUG: File size check" "Size: $file_size bytes, Check result: $size_check_result"
      if [ $size_check_result -eq 0 ] && [ "$file_size" -gt 0 ]; then
        # Attempt to read file content
        MARKDOWN_CONTENT=$(cat "$markdown_file" 2>/dev/null)
        local read_result=$?
        
        log "DEBUG: File read attempt" "Read result: $read_result, Content length: ${#MARKDOWN_CONTENT}"
        if [ $read_result -eq 0 ] && [ -n "$MARKDOWN_CONTENT" ]; then
          log "Successfully generated markdown content" "Size: $file_size bytes"
        else
          # Handle read failure - use stdout as fallback
          log "File read failed, using stdout content" "Path: $markdown_file"
          MARKDOWN_CONTENT="$stdout_output"
        fi
      else
        # Handle size check or empty file - use stdout as fallback
        log "File is empty or size check failed, using stdout content" "Path: $markdown_file, Size: $file_size"
        MARKDOWN_CONTENT="$stdout_output"
      fi
    else
      # Handle unreadable file - use stdout as fallback
      log "File is not readable, using stdout content" "Path: $markdown_file"
      MARKDOWN_CONTENT="$stdout_output"
    fi
  else
    # Handle missing file - this is expected when dual output is not supported
    log "DEBUG: Markdown file missing" "File does not exist: $markdown_file"
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
  
  # Clean up the markdown file from workspace (optional)
  if [ -f "$markdown_file" ]; then
    log "DEBUG: Cleaning up workspace markdown file" "Path: $markdown_file"
    rm -f "$markdown_file" 2>/dev/null || true
  fi
  
  echo "$stdout_output"
  return $exit_code
}