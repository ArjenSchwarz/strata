#!/bin/bash
set -e

# Strata GitHub Action - Modular Version
# This is the main entry point that orchestrates the modular components

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source all module dependencies
source "$SCRIPT_DIR/lib/action/utils.sh"
source "$SCRIPT_DIR/lib/action/security.sh"
source "$SCRIPT_DIR/lib/action/files.sh"
source "$SCRIPT_DIR/lib/action/binary.sh"
source "$SCRIPT_DIR/lib/action/strata.sh"

# DEBUGGING: Show what strata.sh file we loaded
echo "##[warning]DEBUG: Sourced strata.sh from: $SCRIPT_DIR/lib/action/strata.sh"
if [ -f "$SCRIPT_DIR/lib/action/strata.sh" ]; then
    echo "##[warning]DEBUG: strata.sh file exists, checking for debug markers:"
    grep -n "NEW_DEBUG" "$SCRIPT_DIR/lib/action/strata.sh" || echo "No NEW_DEBUG marker found"
    grep -n "DEBUG.*command building" "$SCRIPT_DIR/lib/action/strata.sh" || echo "No command building debug marker found"
fi

# DEBUGGING: Verify function is loaded correctly
echo "##[warning]DEBUG: Checking if run_strata function is available:"
if declare -f run_strata > /dev/null; then
    echo "##[warning]DEBUG: run_strata function is defined"
    echo "##[warning]DEBUG: First few lines of function:"
    declare -f run_strata | head -10
else
    echo "##[warning]DEBUG: run_strata function NOT FOUND!"
fi

source "$SCRIPT_DIR/lib/action/github.sh"

# Initialize global variables and environment
main() {
  log "Strata GitHub Action starting" "Modular version"
  
  # Validate and sanitize inputs
  validate_and_sanitize_inputs
  
  # Create temporary directory for downloads
  TEMP_DIR=$(mktemp -d)
  
  # Set up cleanup trap
  trap 'secure_exit_cleanup' EXIT
  
  # Setup Strata binary
  setup_strata_binary
  
  # Run analysis and distribute outputs
  execute_strata_analysis
  
  # Final cleanup
  log "Strata GitHub Action completed successfully"
}

# Function to validate and sanitize all inputs
validate_and_sanitize_inputs() {
  log "Validating and sanitizing inputs"

  # Sanitize and validate required inputs
  INPUT_PLAN_FILE=$(sanitize_input_parameter "plan-file" "$INPUT_PLAN_FILE" "path")
  if [ -z "$INPUT_PLAN_FILE" ]; then
    error "Plan file is required. Please specify the 'plan-file' input parameter."
  fi

  # Validate plan file existence and security
  if ! validate_file_path "$INPUT_PLAN_FILE" "plan_file"; then
    error "Plan file path validation failed: $INPUT_PLAN_FILE. Path contains security violations."
  fi

  if [ ! -f "$INPUT_PLAN_FILE" ]; then
    error "Plan file does not exist: $INPUT_PLAN_FILE. Please check the path and ensure the file is generated before running this action."
  fi

  # Check if plan file is readable
  if [ ! -r "$INPUT_PLAN_FILE" ]; then
    error "Plan file is not readable: $INPUT_PLAN_FILE. Please check file permissions."
  fi

  # Sanitize and validate output format
  OUTPUT_FORMAT=$(sanitize_input_parameter "output-format" "${INPUT_OUTPUT_FORMAT:-markdown}" "string")
  case "$OUTPUT_FORMAT" in
    markdown|json|table)
      log "Using output format: $OUTPUT_FORMAT"
      ;;
    *)
      warning "Invalid output format: $OUTPUT_FORMAT. Valid options are: markdown, json, table. Defaulting to markdown."
      OUTPUT_FORMAT="markdown"
      ;;
  esac

  # Sanitize and validate boolean inputs
  SHOW_DETAILS=$(sanitize_input_parameter "show-details" "${INPUT_SHOW_DETAILS:-false}" "boolean")
  COMMENT_ON_PR=$(sanitize_input_parameter "comment-on-pr" "${INPUT_COMMENT_ON_PR:-true}" "boolean")
  UPDATE_COMMENT=$(sanitize_input_parameter "update-comment" "${INPUT_UPDATE_COMMENT:-true}" "boolean")

  # Sanitize and validate danger threshold if provided
  if [ -n "$INPUT_DANGER_THRESHOLD" ]; then
    INPUT_DANGER_THRESHOLD=$(sanitize_input_parameter "danger-threshold" "$INPUT_DANGER_THRESHOLD" "integer")
    if [ -z "$INPUT_DANGER_THRESHOLD" ] || [ "$INPUT_DANGER_THRESHOLD" = "0" ]; then
      log "Danger threshold validation" "Using default value (empty)"
      INPUT_DANGER_THRESHOLD=""
    fi
  fi

  # Sanitize and validate config file if provided
  if [ -n "$INPUT_CONFIG_FILE" ]; then
    INPUT_CONFIG_FILE=$(sanitize_input_parameter "config-file" "$INPUT_CONFIG_FILE" "path")
    if [ -n "$INPUT_CONFIG_FILE" ]; then
      if ! validate_file_path "$INPUT_CONFIG_FILE" "config_file"; then
        warning "Config file path validation failed: $INPUT_CONFIG_FILE. The default configuration will be used."
        INPUT_CONFIG_FILE=""
      elif [ ! -f "$INPUT_CONFIG_FILE" ]; then
        warning "Config file does not exist: $INPUT_CONFIG_FILE. The default configuration will be used."
        INPUT_CONFIG_FILE=""
      elif [ ! -r "$INPUT_CONFIG_FILE" ]; then
        warning "Config file is not readable: $INPUT_CONFIG_FILE. The default configuration will be used."
        INPUT_CONFIG_FILE=""
      fi
    fi
  fi

  # Sanitize comment header
  COMMENT_HEADER=$(sanitize_input_parameter "comment-header" "${INPUT_COMMENT_HEADER:-🏗️ Terraform Plan Summary}" "string")
  
  log "Input validation and sanitization completed"
}

# Function to execute Strata analysis and handle outputs
execute_strata_analysis() {
  log "Starting Strata analysis with dual output system"
  
  # Execute with comprehensive error handling
  log "Executing main Strata analysis" "Plan file: $INPUT_PLAN_FILE, Show details: $SHOW_DETAILS"
  
  # DEBUGGING: Add marker before calling run_strata
  echo "##[warning]DEBUG: About to call run_strata from action.sh"
  
  # Call run_strata directly to allow real-time output display
  # The function will set STRATA_OUTPUT and STRATA_EXIT_CODE as global variables
  run_strata "table" "$INPUT_PLAN_FILE" "$SHOW_DETAILS"
  STRATA_EXIT_CODE=$?

  # Log the results of dual output execution
  if [ $STRATA_EXIT_CODE -eq 0 ]; then
    log "Dual output execution successful" "Exit code: $STRATA_EXIT_CODE"
  else
    log "Dual output execution failed" "Exit code: $STRATA_EXIT_CODE"
    warning "Strata analysis failed with exit code $STRATA_EXIT_CODE"
    warning "Error output: $STRATA_OUTPUT"
  fi

  # Handle execution failure with structured error content
  if [ $STRATA_EXIT_CODE -ne 0 ]; then
    warning "Strata analysis failed with exit code $STRATA_EXIT_CODE"
    
    # Write structured error to step summary using markdown content
    if [ -n "$GITHUB_STEP_SUMMARY" ]; then
      if [ -n "$MARKDOWN_CONTENT" ]; then
        echo "$MARKDOWN_CONTENT" >> "$GITHUB_STEP_SUMMARY"
        log "Wrote structured error content to step summary" "Size: ${#MARKDOWN_CONTENT} chars"
      else
        fallback_error_content="## ❌ Error Encountered

**Strata execution failed with exit code $STRATA_EXIT_CODE**

Please check the action logs for more details."
        echo "$fallback_error_content" >> "$GITHUB_STEP_SUMMARY"
        log "Wrote fallback error content to step summary"
      fi
    fi
    
    # Ensure we have fallback output if needed
    if [ -z "$STRATA_OUTPUT" ]; then
      STRATA_OUTPUT="Error: Failed to analyze Terraform plan. Please check the logs for details."
      log "Set fallback output message"
    fi
  fi

  # Parse statistics from the table output
  log "Parsing statistics from table output" "Source: table format"
  
  # Extract statistics from table output using simple parsing
  # Set reasonable defaults
  HAS_CHANGES="false"
  HAS_DANGERS="false"
  CHANGE_COUNT="0"
  DANGER_COUNT="0"
  ADD_COUNT="0"
  CHANGE_COUNT_DETAIL="0"
  DESTROY_COUNT="0"
  REPLACE_COUNT="0"
  
  # Parse from table output if available
  if [ -n "$STRATA_OUTPUT" ]; then
    # Look for the summary table in the output
    if echo "$STRATA_OUTPUT" | grep -q "TOTAL.*ADDED.*REMOVED"; then
      # Extract the data row from the table
      local table_data
      table_data=$(echo "$STRATA_OUTPUT" | grep -A1 "TOTAL.*ADDED.*REMOVED" | tail -1)
      
      if [ -n "$table_data" ]; then
        # Split the table data into fields
        local total added removed modified replacements conditionals high_risk
        read -r total added removed modified replacements conditionals high_risk <<< "$table_data"
        
        # Clean up the values (remove any non-numeric characters)
        total=$(echo "$total" | tr -d '[:alpha:][:space:]')
        added=$(echo "$added" | tr -d '[:alpha:][:space:]')
        removed=$(echo "$removed" | tr -d '[:alpha:][:space:]')
        modified=$(echo "$modified" | tr -d '[:alpha:][:space:]')
        replacements=$(echo "$replacements" | tr -d '[:alpha:][:space:]')
        high_risk=$(echo "$high_risk" | tr -d '[:alpha:][:space:]')
        
        # Set the parsed values
        ADD_COUNT="${added:-0}"
        DESTROY_COUNT="${removed:-0}"
        CHANGE_COUNT_DETAIL="${modified:-0}"
        REPLACE_COUNT="${replacements:-0}"
        DANGER_COUNT="${high_risk:-0}"
        
        # Calculate total changes
        CHANGE_COUNT=$((ADD_COUNT + DESTROY_COUNT + CHANGE_COUNT_DETAIL + REPLACE_COUNT))
        
        # Determine if there are changes or dangers
        if [ "$CHANGE_COUNT" -gt 0 ]; then
          HAS_CHANGES="true"
        fi
        
        if [ "$DANGER_COUNT" -gt 0 ]; then
          HAS_DANGERS="true"
        fi
        
        log "Statistics parsed successfully" "Changes: $CHANGE_COUNT, Dangers: $DANGER_COUNT"
      fi
    fi
  fi

  # Distribute outputs to GitHub contexts
  log "Initiating output distribution phase" "Distributing content to GitHub contexts"
  distribute_output "$STRATA_OUTPUT" "$MARKDOWN_CONTENT"
  log "Output distribution phase completed" "All content distributed successfully"

  # Return the Strata exit code
  return $STRATA_EXIT_CODE
}

# Export variables that are used by modules but defined here
export TEMP_DIR COMMENT_ON_PR UPDATE_COMMENT COMMENT_HEADER
export HAS_CHANGES HAS_DANGERS CHANGE_COUNT DANGER_COUNT
export ADD_COUNT CHANGE_COUNT_DETAIL DESTROY_COUNT REPLACE_COUNT

# Export functions that might be needed by modules
export -f log warning error set_output write_summary

# Run main function
main "$@"
exit_code=$?

# Exit with the final exit code
exit $exit_code