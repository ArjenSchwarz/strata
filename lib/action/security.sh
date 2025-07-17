#!/bin/bash

# Security functions for Strata GitHub Action
# This module provides input validation, sanitization, and security checks

# Function to validate file paths for security and prevent path traversal attacks
validate_file_path() {
  local file_path="$1"
  local context="${2:-general}"
  
  log "Validating file path for security" "Path: $file_path, Context: $context"
  
  # Check for empty path
  if [ -z "$file_path" ]; then
    warning "Empty file path provided"
    log "Path validation failed" "Reason: empty path"
    return 1
  fi
  
  # Check for path traversal attempts (multiple patterns)
  if [[ "$file_path" == *".."* ]]; then
    warning "Path traversal detected in file path: $file_path"
    log "Security violation" "Path traversal attempt detected"
    return 1
  fi
  
  # Check for additional path traversal patterns
  if [[ "$file_path" == *"/./"* ]] || [[ "$file_path" == *"/../"* ]] || [[ "$file_path" == *"/."* ]]; then
    warning "Suspicious path pattern detected: $file_path"
    log "Security violation" "Suspicious path traversal pattern"
    return 1
  fi
  
  # Check for null bytes (security concern)
  if printf '%s' "$file_path" | grep -q '\x00'; then
    warning "Null byte detected in file path: $file_path"
    log "Security violation" "Null byte injection attempt"
    return 1
  fi
  
  # Check for control characters
  if printf '%s' "$file_path" | LC_ALL=C grep -q '[[:cntrl:]]'; then
    warning "Control characters detected in file path: $file_path"
    log "Security violation" "Control character injection attempt"
    return 1
  fi
  
  # Check for excessively long paths (potential buffer overflow)
  if [ ${#file_path} -gt 4096 ]; then
    warning "File path too long: ${#file_path} characters (max: 4096)"
    log "Security violation" "Path length exceeds safe limits"
    return 1
  fi
  
  # Check for dangerous characters in filename
  local basename_file
  basename_file=$(basename "$file_path")
  if [[ "$basename_file" == *";"* ]] || [[ "$basename_file" == *"|"* ]] || [[ "$basename_file" == *"&"* ]] || [[ "$basename_file" == *'$'* ]] || [[ "$basename_file" == *'`'* ]]; then
    warning "Dangerous characters in filename: $basename_file"
    log "Security violation" "Shell metacharacters in filename"
    return 1
  fi
  
  # Resolve path if possible for additional validation
  local resolved_path
  if command -v realpath >/dev/null 2>&1; then
    if ! resolved_path=$(realpath "$file_path" 2>/dev/null); then
      # Path doesn't exist yet, try to resolve parent directory
      local parent_dir
      parent_dir=$(dirname "$file_path")
      if [ -d "$parent_dir" ]; then
        local resolved_parent
        if resolved_parent=$(realpath "$parent_dir" 2>/dev/null); then
          resolved_path="$resolved_parent/$(basename "$file_path")"
        else
          warning "Cannot resolve parent directory: $parent_dir"
          log "Path validation warning" "Parent directory resolution failed"
          resolved_path="$file_path"
        fi
      else
        warning "Parent directory does not exist: $parent_dir"
        log "Path validation warning" "Parent directory missing"
        resolved_path="$file_path"
      fi
    fi
  else
    # Fallback validation without realpath
    resolved_path="$file_path"
    log "Path resolution" "realpath not available, using basic validation"
  fi
  
  # Context-specific validation with enhanced security checks
  case "$context" in
    "temp_file")
      # Ensure temp files are in appropriate directories with strict validation
      local valid_temp_dir=false
      
      # Check against known safe temporary directories
      if [[ "$resolved_path" == /tmp/* ]] || \
         [[ "$resolved_path" == "$TEMP_DIR"/* ]] || \
         [[ "$resolved_path" == "$TMPDIR"/* ]] || \
         [[ "$resolved_path" == /var/folders/* ]] || \
         [[ "$resolved_path" == /private/var/folders/* ]] || \
         [[ "$resolved_path" == /private/tmp/* ]]; then
        valid_temp_dir=true
      fi
      
      if [ "$valid_temp_dir" = false ]; then
        warning "Temporary file not in safe directory: $resolved_path"
        log "Security violation" "Temporary file outside safe directories"
        return 1
      fi
      
      # Additional check: ensure temp file has secure naming pattern
      local temp_basename
      temp_basename=$(basename "$resolved_path")
      if [[ "$temp_basename" != tmp.* ]] && [[ "$temp_basename" != strata_secure.* ]]; then
        log "Temporary file naming" "Non-standard naming pattern: $temp_basename"
      fi
      ;;
    "plan_file")
      # Ensure plan files have reasonable extensions and are not in system directories
      if [[ "$file_path" != *.tfplan ]] && [[ "$file_path" != *.json ]] && [[ "$file_path" != *.plan ]]; then
        warning "Plan file has unexpected extension: $file_path"
        log "File validation warning" "Unexpected plan file extension"
      fi
      
      # Prevent access to system directories
      if [[ "$resolved_path" == /etc/* ]] || [[ "$resolved_path" == /bin/* ]] || [[ "$resolved_path" == /sbin/* ]] || [[ "$resolved_path" == /usr/bin/* ]]; then
        warning "Plan file in system directory: $resolved_path"
        log "Security violation" "Plan file in restricted system directory"
        return 1
      fi
      ;;
    "config_file")
      # Ensure config files have reasonable extensions and are not in system directories
      if [[ "$file_path" != *.yaml ]] && [[ "$file_path" != *.yml ]] && [[ "$file_path" != *.json ]]; then
        warning "Config file has unexpected extension: $file_path"
        log "File validation warning" "Unexpected config file extension"
      fi
      
      # Prevent access to system directories
      if [[ "$resolved_path" == /etc/* ]] || [[ "$resolved_path" == /bin/* ]] || [[ "$resolved_path" == /sbin/* ]] || [[ "$resolved_path" == /usr/bin/* ]]; then
        warning "Config file in system directory: $resolved_path"
        log "Security violation" "Config file in restricted system directory"
        return 1
      fi
      ;;
    "general")
      # General file validation - prevent access to sensitive system files
      if [[ "$resolved_path" == /etc/passwd ]] || [[ "$resolved_path" == /etc/shadow ]] || [[ "$resolved_path" == /etc/hosts ]] || [[ "$resolved_path" == /proc/* ]]; then
        warning "Access to sensitive system file blocked: $resolved_path"
        log "Security violation" "Attempt to access sensitive system file"
        return 1
      fi
      ;;
  esac
  
  log "File path validation successful" "Resolved: $resolved_path, Context: $context"
  return 0
}

# Function to sanitize and validate input parameters
sanitize_input_parameter() {
  local param_name="$1"
  local param_value="$2"
  local param_type="${3:-string}"
  
  log "Sanitizing input parameter" "Name: $param_name, Type: $param_type"
  
  # Check for empty parameter
  if [ -z "$param_value" ]; then
    log "Parameter validation" "$param_name is empty"
    echo ""
    return 0
  fi
  
  # Remove null bytes and control characters
  local sanitized_value
  sanitized_value=$(printf '%s' "$param_value" | tr -d '\000-\010\013\014\016-\037\177')
  
  # Type-specific validation and sanitization
  case "$param_type" in
    "boolean")
      # Validate boolean values
      case "$sanitized_value" in
        "true"|"false")
          echo "$sanitized_value"
          return 0
          ;;
        *)
          warning "Invalid boolean value for $param_name: $sanitized_value"
          echo "false"
          return 1
          ;;
      esac
      ;;
    "integer")
      # Validate integer values
      if [[ "$sanitized_value" =~ ^[0-9]+$ ]]; then
        echo "$sanitized_value"
        return 0
      else
        warning "Invalid integer value for $param_name: $sanitized_value"
        echo "0"
        return 1
      fi
      ;;
    "path")
      # Validate and sanitize file paths
      if validate_file_path "$sanitized_value" "general"; then
        echo "$sanitized_value"
        return 0
      else
        warning "Invalid path for $param_name: $sanitized_value"
        echo ""
        return 1
      fi
      ;;
    "string")
      # Basic string sanitization
      # Remove shell metacharacters that could be dangerous
      sanitized_value=$(printf '%s' "$sanitized_value" | sed 's/[;&|`$(){}[\]\\]//g')
      
      # Limit string length to prevent buffer overflow
      if [ ${#sanitized_value} -gt 1024 ]; then
        warning "Parameter $param_name too long, truncating"
        sanitized_value="${sanitized_value:0:1024}"
      fi
      
      echo "$sanitized_value"
      return 0
      ;;
    *)
      warning "Unknown parameter type: $param_type"
      echo "$sanitized_value"
      return 1
      ;;
  esac
}

# Function to sanitize content for GitHub output to prevent script injection
sanitize_github_content() {
  local content="$1"
  
  log "Sanitizing content for GitHub output" "Size: ${#content} chars"
  log "DEBUG: Content before sanitization" "$content"
  
  # Remove potential script tags and other dangerous content
  local sanitized_content
  sanitized_content="$content"
  
  log "DEBUG: Starting HTML tag removal" "Initial size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<script[^>]*>.*<\/script>//gi')
  log "DEBUG: After removing script tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<iframe[^>]*>.*<\/iframe>//gi')
  log "DEBUG: After removing iframe tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<object[^>]*>.*<\/object>//gi')
  log "DEBUG: After removing object tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<embed[^>]*>.*<\/embed>//gi')
  log "DEBUG: After removing embed tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<applet[^>]*>.*<\/applet>//gi')
  log "DEBUG: After removing applet tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<form[^>]*>.*<\/form>//gi')
  log "DEBUG: After removing form tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<input[^>]*>//gi')
  log "DEBUG: After removing input tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<button[^>]*>.*<\/button>//gi')
  log "DEBUG: After removing button tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/javascript:[^"'\'']*//gi')
  log "DEBUG: After removing javascript URLs" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/vbscript:[^"'\'']*//gi')
  log "DEBUG: After removing vbscript URLs" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/data:[^"'\'']*//gi')
  log "DEBUG: After removing data URLs" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/on[a-zA-Z]*=[^"'\'']*//gi')
  log "DEBUG: After removing event handlers" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<link[^>]*>//gi')
  log "DEBUG: After removing link tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<meta[^>]*>//gi')
  log "DEBUG: After removing meta tags" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/<style[^>]*>.*<\/style>//gi')
  log "DEBUG: After removing style tags" "Size: ${#sanitized_content} chars"
  
  # Remove potential HTML entities that could be used for injection
  log "DEBUG: Starting HTML entity processing" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/&lt;script/\&amp;lt;script/gi')
  log "DEBUG: After escaping lt script entities" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/&gt;/\&amp;gt;/g')
  log "DEBUG: After escaping gt entities" "Size: ${#sanitized_content} chars"
  
  sanitized_content=$(echo "$sanitized_content" | sed 's/&quot;/\&amp;quot;/g')
  log "DEBUG: After escaping quot entities" "Size: ${#sanitized_content} chars"
  
  # Remove null bytes and control characters
  log "DEBUG: Before removing control characters" "Size: ${#sanitized_content} chars"
  sanitized_content=$(echo "$sanitized_content" | tr -d '\000-\010\013\014\016-\037')
  log "DEBUG: After removing control characters" "Size: ${#sanitized_content} chars"
  
  log "Content sanitization completed" "Original: ${#content} chars, Sanitized: ${#sanitized_content} chars"
  log "DEBUG: Content after sanitization" "$sanitized_content"
  
  echo "$sanitized_content"
}