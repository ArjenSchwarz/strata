#!/bin/bash

# File operations for Strata GitHub Action
# This module handles temporary file management and cleanup

# Array to track temporary files for cleanup
TEMP_FILES=()

# Function to create secure temporary files with restrictive permissions
create_secure_temp_file() {
  local temp_file
  local mktemp_result
  local context="${1:-general}"
  
  log "Creating secure temporary file" "Context: $context, Using mktemp for secure file creation"
  
  # Create temporary file with secure template
  temp_file=$(mktemp -t "strata_secure.XXXXXXXXXX" 2>/dev/null)
  mktemp_result=$?
  
  # Check for creation failure
  if [ $mktemp_result -ne 0 ] || [ -z "$temp_file" ]; then
    warning "mktemp command failed" "Exit code: $mktemp_result"
    log "Secure temporary file creation failed" "mktemp returned: $mktemp_result, output: '$temp_file'"
    return 1
  fi
  
  # Verify file was actually created
  if [ ! -f "$temp_file" ]; then
    warning "Temporary file was not created despite mktemp success: $temp_file"
    log "File creation verification failed" "Path: $temp_file, File exists: false"
    return 1
  fi
  
  log "Temporary file created successfully" "Path: $temp_file"
  
  # Validate the file path for security
  if ! validate_file_path "$temp_file" "temp_file"; then
    warning "Temporary file path validation failed: $temp_file"
    log "Security validation failed" "Removing insecure temporary file: $temp_file"
    secure_file_cleanup "$temp_file"
    return 1
  fi
  
  # Set highly restrictive permissions (owner read/write only)
  if ! chmod 600 "$temp_file" 2>/dev/null; then
    warning "Failed to set secure permissions on temporary file: $temp_file"
    log "Permission setting failed" "Could not set 600 permissions on: $temp_file"
    secure_file_cleanup "$temp_file"
    return 1
  fi
  
  # Verify permissions were set correctly
  local file_perms
  file_perms=$(stat -c "%a" "$temp_file" 2>/dev/null || stat -f "%A" "$temp_file" 2>/dev/null)
  if [ "$file_perms" != "600" ]; then
    warning "Secure permissions verification failed: $temp_file (permissions: $file_perms)"
    log "Permission verification failed" "Expected: 600, Actual: $file_perms"
    secure_file_cleanup "$temp_file"
    return 1
  fi
  
  log "Security permissions applied and verified" "File: $temp_file, Permissions: $file_perms"
  
  # Verify file is writable by owner
  if [ ! -w "$temp_file" ]; then
    warning "Temporary file is not writable: $temp_file"
    log "Write permission check failed" "File: $temp_file"
    secure_file_cleanup "$temp_file"
    return 1
  fi
  
  # Track the file for cleanup
  TEMP_FILES+=("$temp_file")
  
  log "Secure temporary file ready for use" "Path: $temp_file, Tracked for cleanup: true, Context: $context"
  log "Secure temporary file creation completed" "Total tracked files: ${#TEMP_FILES[@]}"
  echo "$temp_file"
  return 0
}

# Function to create and track temporary files with enhanced error handling (backward compatibility)
create_temp_file() {
  create_secure_temp_file "$@"
}

# Function to securely clean up sensitive temporary files
secure_file_cleanup() {
  local file_path="$1"
  local overwrite_passes="${2:-3}"
  
  if [ -z "$file_path" ] || [ ! -f "$file_path" ]; then
    log "Secure cleanup skipped" "File does not exist: $file_path"
    return 0
  fi
  
  log "Starting secure file cleanup" "File: $file_path, Overwrite passes: $overwrite_passes"
  
  # Get file size for logging
  local file_size
  file_size=$(wc -c < "$file_path" 2>/dev/null || echo "unknown")
  
  # Overwrite file content multiple times with random data
  local pass=1
  while [ $pass -le "$overwrite_passes" ]; do
    log "Secure overwrite pass $pass/$overwrite_passes" "File: $file_path"
    
    # Try different methods to overwrite the file
    if command -v shred >/dev/null 2>&1; then
      # Use shred if available (Linux)
      if shred -vfz -n 1 "$file_path" 2>/dev/null; then
        log "Secure overwrite successful (shred)" "Pass: $pass"
      else
        log "Shred failed, trying alternative method" "Pass: $pass"
        # Fallback to dd with random data
        if dd if=/dev/urandom of="$file_path" bs="$file_size" count=1 2>/dev/null; then
          log "Secure overwrite successful (dd)" "Pass: $pass"
        else
          log "DD overwrite failed, using basic overwrite" "Pass: $pass"
          # Basic overwrite with zeros
          if true > "$file_path" 2>/dev/null; then
            log "Basic overwrite successful" "Pass: $pass"
          else
            warning "All overwrite methods failed" "Pass: $pass"
            break
          fi
        fi
      fi
    elif [ -c /dev/urandom ]; then
      # Use dd with urandom (macOS/BSD)
      if dd if=/dev/urandom of="$file_path" bs="$file_size" count=1 2>/dev/null; then
        log "Secure overwrite successful (dd/urandom)" "Pass: $pass"
      else
        log "DD/urandom failed, using basic overwrite" "Pass: $pass"
        if true > "$file_path" 2>/dev/null; then
          log "Basic overwrite successful" "Pass: $pass"
        else
          warning "Basic overwrite failed" "Pass: $pass"
          break
        fi
      fi
    else
      # Fallback to basic truncation
      log "No secure overwrite tools available, using truncation" "Pass: $pass"
      if true > "$file_path" 2>/dev/null; then
        log "File truncation successful" "Pass: $pass"
      else
        warning "File truncation failed" "Pass: $pass"
        break
      fi
    fi
    
    pass=$((pass + 1))
  done
  
  # Finally remove the file
  if rm -f "$file_path" 2>/dev/null; then
    log "Secure file cleanup completed" "File removed: $file_path"
    return 0
  else
    warning "Failed to remove file after secure overwrite: $file_path"
    log "Secure cleanup partial failure" "File overwritten but not removed"
    return 1
  fi
}

# Enhanced cleanup function with comprehensive error handling and secure cleanup
cleanup_temp_files() {
  local cleanup_errors=0
  local files_cleaned=0
  local files_skipped=0
  local secure_cleanup="${1:-false}"
  
  log "Starting comprehensive temporary file cleanup" "Files to clean: ${#TEMP_FILES[@]}, Secure mode: $secure_cleanup"
  
  if [ ${#TEMP_FILES[@]} -eq 0 ]; then
    log "No temporary files to clean up" "Cleanup completed immediately"
    return 0
  fi
  
  for temp_file in "${TEMP_FILES[@]}"; do
    if [ -n "$temp_file" ]; then
      log "Processing temporary file for cleanup" "$temp_file"
      
      if [ -f "$temp_file" ]; then
        # Get file size before cleanup for logging
        local file_size
        file_size=$(wc -c < "$temp_file" 2>/dev/null || echo "unknown")
        
        # Use secure cleanup for sensitive files
        if [ "$secure_cleanup" = "true" ]; then
          log "Applying secure cleanup" "File: $temp_file"
          if secure_file_cleanup "$temp_file"; then
            log "Successfully performed secure cleanup" "$temp_file (size: $file_size bytes)"
            files_cleaned=$((files_cleaned + 1))
          else
            warning "Secure cleanup failed for: $temp_file"
            log "Secure cleanup failure details" "File: $temp_file, Size: $file_size bytes"
            cleanup_errors=$((cleanup_errors + 1))
          fi
        else
          # Standard cleanup
          if rm -f "$temp_file" 2>/dev/null; then
            log "Successfully cleaned up temporary file" "$temp_file (size: $file_size bytes)"
            files_cleaned=$((files_cleaned + 1))
          else
            warning "Failed to remove temporary file: $temp_file"
            log "Cleanup failure details" "File: $temp_file, Size: $file_size bytes"
            cleanup_errors=$((cleanup_errors + 1))
            
            # Try alternative cleanup methods
            if [ -w "$temp_file" ]; then
              # Try to truncate the file if we can't remove it
              if true > "$temp_file" 2>/dev/null; then
                log "Truncated temporary file (could not remove)" "$temp_file"
                log "Alternative cleanup applied" "File truncated instead of removed"
              else
                warning "Could not truncate temporary file: $temp_file"
                log "All cleanup methods failed" "File: $temp_file"
              fi
            else
              log "File not writable for alternative cleanup" "$temp_file"
            fi
          fi
        fi
      else
        log "Temporary file already removed or never existed" "$temp_file"
        files_skipped=$((files_skipped + 1))
      fi
    else
      log "Empty temporary file path encountered" "Skipping cleanup"
      files_skipped=$((files_skipped + 1))
    fi
  done
  
  # Clear the array
  TEMP_FILES=()
  
  # Clean up any orphaned temporary files in our temp directory
  if [ -d "$TEMP_DIR" ]; then
    log "Checking for orphaned temporary files" "Directory: $TEMP_DIR"
    local orphaned_files
    orphaned_files=$(find "$TEMP_DIR" -name "tmp.*" -o -name "strata_secure.*" -type f 2>/dev/null | wc -l)
    if [ "$orphaned_files" -gt 0 ]; then
      log "Cleaning up orphaned temporary files" "Count: $orphaned_files"
      
      if [ "$secure_cleanup" = "true" ]; then
        # Secure cleanup for orphaned files
        find "$TEMP_DIR" \( -name "tmp.*" -o -name "strata_secure.*" \) -type f 2>/dev/null | while read -r orphaned_file; do
          if [ -f "$orphaned_file" ]; then
            log "Secure cleanup of orphaned file" "$orphaned_file"
            secure_file_cleanup "$orphaned_file" 1  # Single pass for orphaned files
          fi
        done
      else
        # Standard cleanup for orphaned files
        find "$TEMP_DIR" \( -name "tmp.*" -o -name "strata_secure.*" \) -type f -delete 2>/dev/null || true
      fi
      
      log "Orphaned file cleanup completed" "Files processed: $orphaned_files"
    else
      log "No orphaned temporary files found" "Directory clean"
    fi
  else
    log "Temporary directory not found" "Path: $TEMP_DIR"
  fi
  
  # Final cleanup summary
  log "Temporary file cleanup summary" "Cleaned: $files_cleaned, Skipped: $files_skipped, Errors: $cleanup_errors, Secure mode: $secure_cleanup"
  
  if [ $cleanup_errors -gt 0 ]; then
    warning "Encountered $cleanup_errors errors during cleanup"
    log "Cleanup completed with errors" "Some files may not have been removed properly"
    return 1
  else
    log "Temporary file cleanup completed successfully" "All tracked files processed"
    return 0
  fi
}

# Enhanced trap for secure cleanup on exit
secure_exit_cleanup() {
  log "Performing secure exit cleanup" "Cleaning up all temporary resources"
  
  # Perform secure cleanup of sensitive temporary files
  cleanup_temp_files "true"
  
  # Clean up main temporary directory
  if [ -d "$TEMP_DIR" ]; then
    log "Cleaning up main temporary directory" "$TEMP_DIR"
    rm -rf "$TEMP_DIR" 2>/dev/null || {
      warning "Failed to remove temporary directory: $TEMP_DIR"
      log "Manual cleanup may be required" "Directory: $TEMP_DIR"
    }
  fi
  
  log "Secure exit cleanup completed"
}