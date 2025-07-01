#!/bin/bash
set -e

# Function to log messages
log() {
  echo "::group::$1" >&2
  shift
  echo "$@" >&2
  echo "::endgroup::" >&2
}

# Function to log errors
error() {
  echo "::error::$1" >&2
  
  # Write error to step summary if available
  if [ -n "$GITHUB_STEP_SUMMARY" ]; then
    {
      echo "## ❌ Error Encountered"
      echo ""
      echo "**$1**"
      echo ""
      echo "Please check the action logs for more details."
    } >> "$GITHUB_STEP_SUMMARY"
  fi
  
  exit 1
}

# Function to log warnings
warning() {
  echo "::warning::$1" >&2
}

# Function to set GitHub Actions outputs
set_output() {
  echo "$1=$2" >> "$GITHUB_OUTPUT"
}

# Function to write to GitHub Step Summary
write_summary() {
  echo "$1" >> "$GITHUB_STEP_SUMMARY"
}

# Function to compile from source
compile_from_source() {
  log "Compiling Strata from source"
  
  # Check if Go is installed
  if ! command -v go >/dev/null 2>&1; then
    log "Installing Go"
    GO_VERSION="1.24.1"
    GO_URL="https://golang.org/dl/go${GO_VERSION}.${PLATFORM}-${ARCH}.tar.gz"
    
    # Download and install Go
    curl -L -s -o "$TEMP_DIR/go.tar.gz" "$GO_URL"
    tar -C "$TEMP_DIR" -xzf "$TEMP_DIR/go.tar.gz"
    export PATH="$TEMP_DIR/go/bin:$PATH"
    
    # Verify Go installation
    if ! command -v go >/dev/null 2>&1; then
      error "Failed to install Go, cannot compile from source"
      return 1
    fi
  fi
  
  # Create directory for source code
  SRC_DIR="$TEMP_DIR/src"
  mkdir -p "$SRC_DIR"
  cd "$SRC_DIR"
  
  # Clone repository
  log "Cloning Strata repository"
  if ! git clone --depth 1 https://github.com/ArjenSchwarz/strata.git .; then
    error "Failed to clone repository"
    return 1
  fi
  
  # If version is specified, checkout that version
  if [ -n "$1" ]; then
    log "Checking out version $1"
    if ! git checkout "$1" 2>/dev/null; then
      warning "Failed to checkout version $1, using default branch"
    fi
  fi
  
  # Build binary
  log "Building Strata binary"
  if ! go build -o "$TEMP_DIR/$BINARY_NAME" .; then
    error "Failed to build Strata binary"
    return 1
  fi
  
  # Verify binary works
  if ! "$TEMP_DIR/$BINARY_NAME" --version; then
    error "Compiled binary verification failed"
    return 1
  fi
  
  log "Successfully compiled Strata from source"
  return 0
}

# Function to download with retry
download_with_retry() {
  local url=$1
  local output=$2
  local max_attempts=3
  local attempt=1
  local timeout=10
  
  while [ $attempt -le $max_attempts ]; do
    log "Download attempt $attempt/$max_attempts" "URL: $url"
    if curl -L -s --connect-timeout $timeout -o "$output" "$url"; then
      return 0
    fi
    
    attempt=$((attempt + 1))
    if [ $attempt -le $max_attempts ]; then
      sleep_time=$((2 ** (attempt - 1)))
      log "Retrying download in ${sleep_time}s"
      sleep $sleep_time
    fi
  done
  
  return 1
}

# Function to verify checksum
verify_checksum() {
  local file=$1
  local expected_checksum=$2
  local algorithm=$3
  
  if [ -z "$expected_checksum" ]; then
    warning "No checksum provided for verification, skipping"
    return 0
  fi
  
  local actual_checksum
  case $algorithm in
    md5)
      if command -v md5sum >/dev/null 2>&1; then
        actual_checksum=$(md5sum "$file" | cut -d' ' -f1)
      elif command -v md5 >/dev/null 2>&1; then
        actual_checksum=$(md5 -q "$file")
      else
        warning "Neither md5sum nor md5 found, skipping checksum verification"
        return 0
      fi
      ;;
    sha256)
      if command -v sha256sum >/dev/null 2>&1; then
        actual_checksum=$(sha256sum "$file" | cut -d' ' -f1)
      elif command -v shasum >/dev/null 2>&1; then
        actual_checksum=$(shasum -a 256 "$file" | cut -d' ' -f1)
      else
        warning "Neither sha256sum nor shasum found, skipping checksum verification"
        return 0
      fi
      ;;
    sha512)
      if command -v sha512sum >/dev/null 2>&1; then
        actual_checksum=$(sha512sum "$file" | cut -d' ' -f1)
      elif command -v shasum >/dev/null 2>&1; then
        actual_checksum=$(shasum -a 512 "$file" | cut -d' ' -f1)
      else
        warning "Neither sha512sum nor shasum found, skipping checksum verification"
        return 0
      fi
      ;;
    *)
      warning "Unsupported checksum algorithm: $algorithm"
      return 0
      ;;
  esac
  
  if [ "$actual_checksum" = "$expected_checksum" ]; then
    log "Checksum verification passed" "$algorithm: $actual_checksum"
    return 0
  else
    warning "Checksum verification failed. Expected: $expected_checksum, Got: $actual_checksum"
    return 1
  fi
}

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
  if [ "${file_path}" != "${file_path%$'\0'*}" ]; then
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

# Validate and sanitize inputs
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

# Determine platform for binary download
PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

log "Detected platform" "$PLATFORM-$ARCH"

# Set binary name based on platform
if [ "$PLATFORM" = "windows" ]; then
  BINARY_NAME="strata.exe"
else
  BINARY_NAME="strata"
fi

# Create temporary directory for downloads
TEMP_DIR=$(mktemp -d)

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

trap 'secure_exit_cleanup' EXIT

# Setup cache directories
CACHE_DIR="$HOME/.cache/strata"
mkdir -p "$CACHE_DIR"

# Function to check if binary is cached
check_cache() {
  local version=$1
  local platform=$2
  local arch=$3
  local cache_path="$CACHE_DIR/strata_${version}_${platform}_${arch}"
  
  if [ -f "$cache_path" ] && [ -x "$cache_path" ]; then
    log "Found cached binary" "$cache_path"
    # Verify cached binary works
    if "$cache_path" --version >/dev/null 2>&1; then
      log "Cached binary verification successful"
      cp "$cache_path" "$TEMP_DIR/$BINARY_NAME"
      return 0
    else
      log "Cached binary verification failed, will download fresh copy"
      return 1
    fi
  fi
  
  return 1
}

# Function to save binary to cache
save_to_cache() {
  local binary_path=$1
  local version=$2
  local platform=$3
  local arch=$4
  local cache_path="$CACHE_DIR/strata_${version}_${platform}_${arch}"
  
  # Only cache if binary works
  if "$binary_path" --version >/dev/null 2>&1; then
    log "Saving binary to cache" "$cache_path"
    cp "$binary_path" "$cache_path"
    chmod +x "$cache_path"
    return 0
  else
    log "Binary verification failed, not caching"
    return 1
  fi
}

# Download latest release information
log "Determining latest release"
LATEST_RELEASE_URL="https://api.github.com/repos/ArjenSchwarz/strata/releases/latest"
LATEST_RELEASE=$(curl -s -H "Accept: application/vnd.github.v3+json" $LATEST_RELEASE_URL)
LATEST_VERSION=$(echo "$LATEST_RELEASE" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)

if [ -z "$LATEST_VERSION" ]; then
  warning "Could not determine latest version, falling back to compilation"
  if ! compile_from_source; then
    error "Failed to compile from source after failing to determine latest version"
  fi
else
  log "Latest version" "$LATEST_VERSION"
  
  # Check if binary is in cache
  if check_cache "$LATEST_VERSION" "$PLATFORM" "$ARCH"; then
    log "Using cached binary"
  else
    # Determine file extension based on platform
    if [ "$PLATFORM" = "windows" ]; then
      FILE_EXT="zip"
    else
      FILE_EXT="tar.gz"
    fi
    
    # Construct download URLs with correct filename format
    ARCHIVE_NAME="strata-${LATEST_VERSION}-${PLATFORM}-${ARCH}.${FILE_EXT}"
    DOWNLOAD_URL="https://github.com/ArjenSchwarz/strata/releases/download/${LATEST_VERSION}/${ARCHIVE_NAME}"
    CHECKSUM_URL="https://github.com/ArjenSchwarz/strata/releases/download/${LATEST_VERSION}/${ARCHIVE_NAME}.md5"
    
    # Download archive with retry
    log "Downloading Strata archive" "$DOWNLOAD_URL"
    ARCHIVE_PATH="$TEMP_DIR/$ARCHIVE_NAME"
    if ! download_with_retry "$DOWNLOAD_URL" "$ARCHIVE_PATH"; then
      warning "Failed to download archive after multiple attempts, falling back to compilation"
      if ! compile_from_source "$LATEST_VERSION"; then
        error "Failed to compile from source after download failure"
      fi
    else
      # Extract binary from archive
      log "Extracting binary from archive"
      if [ "$PLATFORM" = "windows" ]; then
        # Extract from zip
        if command -v unzip >/dev/null 2>&1; then
          unzip -q "$ARCHIVE_PATH" -d "$TEMP_DIR/extract/"
          cp "$TEMP_DIR/extract/$BINARY_NAME" "$TEMP_DIR/$BINARY_NAME"
        else
          error "unzip command not found, cannot extract Windows archive"
        fi
      else
        # Extract from tar.gz
        tar -xzf "$ARCHIVE_PATH" -C "$TEMP_DIR/"
        # Find and copy the binary (it might be in a subdirectory)
        if [ -f "$TEMP_DIR/$BINARY_NAME" ]; then
          # Binary is in the root of the archive
          log "Binary found in archive root"
        elif [ -f "$TEMP_DIR/strata/$BINARY_NAME" ]; then
          # Binary is in a strata subdirectory
          cp "$TEMP_DIR/strata/$BINARY_NAME" "$TEMP_DIR/$BINARY_NAME"
          log "Binary found in strata subdirectory"
        else
          # Try to find the binary anywhere in the extracted files
          FOUND_BINARY=$(find "$TEMP_DIR" -name "$BINARY_NAME" -type f | head -1)
          if [ -n "$FOUND_BINARY" ]; then
            cp "$FOUND_BINARY" "$TEMP_DIR/$BINARY_NAME"
            log "Binary found at: $FOUND_BINARY"
          else
            error "Could not find binary $BINARY_NAME in extracted archive"
          fi
        fi
      fi
      
      # Ensure binary is executable
      chmod +x "$TEMP_DIR/$BINARY_NAME"
      
      # Download and verify checksum if available
      log "Downloading checksum" "$CHECKSUM_URL"
      if download_with_retry "$CHECKSUM_URL" "$TEMP_DIR/checksum.md5"; then
        # Extract expected checksum for our archive
        EXPECTED_CHECKSUM=$(cut -d' ' -f1 < "$TEMP_DIR/checksum.md5")
        
        if [ -n "$EXPECTED_CHECKSUM" ]; then
          # Verify checksum of the archive
          if ! verify_checksum "$ARCHIVE_PATH" "$EXPECTED_CHECKSUM" "md5"; then
            warning "Checksum verification failed, falling back to compilation"
            if ! compile_from_source "$LATEST_VERSION"; then
              error "Failed to compile from source after checksum verification failure"
            fi
          fi
        else
          warning "Could not extract checksum from checksum file"
        fi
      else
        warning "Failed to download checksum, skipping verification"
      fi
      
      # Verify binary works
      log "Verifying binary"
      if ! "$TEMP_DIR/$BINARY_NAME" --version; then
        warning "Downloaded binary verification failed, falling back to compilation"
        if ! compile_from_source "$LATEST_VERSION"; then
          error "Failed to compile from source after binary verification failure"
        fi
      else
        log "Binary verification successful"
        
        # Save to cache
        save_to_cache "$TEMP_DIR/$BINARY_NAME" "$LATEST_VERSION" "$PLATFORM" "$ARCH"
      fi
    fi
  fi
fi

# Function to run Strata with specified parameters
run_strata() {
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

# Global variable to store markdown content from dual output
MARKDOWN_CONTENT=""

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
  
  # Verify file is not readable by group/others
  if [ -r "$temp_file" ]; then
    # Check if others can read (this is a basic check)
    local test_result
    test_result=$(stat -c "%a" "$temp_file" 2>/dev/null || stat -f "%A" "$temp_file" 2>/dev/null | cut -c8-10)
    if [ "$test_result" != "---" ]; then
      warning "File permissions may allow group/other access: $temp_file"
      log "Permission security check" "Group/Other permissions: $test_result"
    fi
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

# Function to cleanup tracked temporary files
cleanup_temp_files() {
  for temp_file in "${TEMP_FILES[@]}"; do
    if [ -f "$temp_file" ]; then
      rm -f "$temp_file"
      log "Cleaned up temporary file" "$temp_file"
    fi
  done
  TEMP_FILES=()
}

# Function to process markdown content for different GitHub contexts
process_markdown_for_context() {
  local context=$1  # "step-summary" or "pr-comment"
  local content="$2"
  
  case "$context" in
    "step-summary")
      # Add step summary specific enhancements
      echo "# $COMMENT_HEADER"
      echo ""
      echo "$content"
      echo ""
      add_workflow_info
      ;;
    "pr-comment")
      # Add PR comment specific enhancements
      echo "## $COMMENT_HEADER"
      echo ""
      echo "<!-- strata-comment-id: $GITHUB_WORKFLOW-$GITHUB_JOB -->"
      echo ""
      echo "$content"
      echo ""
      add_pr_footer
      ;;
    *)
      echo "$content"
      ;;
  esac
}

# Function to add workflow information for step summaries
add_workflow_info() {
  echo "<details>"
  echo "<summary>ℹ️ Workflow Information</summary>"
  echo ""
  echo "- **Repository:** $GITHUB_REPOSITORY"
  echo "- **Workflow:** $GITHUB_WORKFLOW"
  echo "- **Run ID:** $GITHUB_RUN_ID"
  echo "- **Strata Version:** $("$TEMP_DIR"/$BINARY_NAME --version | head -n 1)"
  echo ""
  echo "</details>"
  echo ""
  echo "---"
  echo "*Generated by [Strata](https://github.com/ArjenSchwarz/strata)*"
}

# Function to add PR comment footer
add_pr_footer() {
  echo "---"
  echo "*Generated by [Strata](https://github.com/ArjenSchwarz/strata) in [workflow run](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID})*"
}

# Function to optimize content for specific contexts
optimize_content_for_context() {
  local context=$1
  local content="$2"
  
  case "$context" in
    "step-summary")
      # Add collapsible sections for better organization
      echo "$content" | add_collapsible_sections
      ;;
    "pr-comment")
      # Limit content size for PR comments
      echo "$content" | limit_content_size 65000
      ;;
    *)
      echo "$content"
      ;;
  esac
}

# Function to add collapsible sections to large content blocks
add_collapsible_sections() {
  local content
  content=$(cat)
  
  # If content is longer than 2000 characters, wrap sections in collapsible details
  if [ ${#content} -gt 2000 ]; then
    # Look for section headers (lines starting with ##) and wrap content
    echo "$content" | awk '
      BEGIN { in_section = 0; section_content = ""; section_header = "" }
      /^## / {
        if (in_section && section_content != "") {
          print "<details>"
          print "<summary>" section_header "</summary>"
          print ""
          print section_content
          print ""
          print "</details>"
          print ""
        }
        section_header = $0
        section_content = ""
        in_section = 1
        next
      }
      {
        if (in_section) {
          section_content = section_content $0 "\n"
        } else {
          print $0
        }
      }
      END {
        if (in_section && section_content != "") {
          print "<details>"
          print "<summary>" section_header "</summary>"
          print ""
          print section_content
          print ""
          print "</details>"
        }
      }
    '
  else
    echo "$content"
  fi
}

# Function to limit content size for PR comments
limit_content_size() {
  local max_size=$1
  local content
  content=$(cat)
  
  if [ ${#content} -gt "$max_size" ]; then
    # Calculate truncation point (try to break at a reasonable place)
    local truncate_point=$((max_size - 200))
    local truncated_content="${content:0:$truncate_point}"
    
    # Try to find a good break point (end of line)
    local last_newline
    last_newline=$(echo "$truncated_content" | grep -n $'\n' | tail -1 | cut -d: -f1)
    
    if [ -n "$last_newline" ] && [ "$last_newline" -gt $((truncate_point - 500)) ]; then
      truncated_content=$(echo "$content" | head -n "$last_newline")
    fi
    
    echo "$truncated_content"
    echo ""
    echo "<details>"
    echo "<summary>⚠️ Content truncated due to size limits</summary>"
    echo ""
    echo "The full output was too large for a GitHub comment. "
    echo "Please check the [workflow run](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}) for complete details."
    echo ""
    echo "</details>"
  else
    echo "$content"
  fi
}

# Function to sanitize content for GitHub output to prevent script injection
sanitize_github_content() {
  local content="$1"
  
  log "Sanitizing content for GitHub output" "Size: ${#content} chars"
  
  # Remove potential script tags and other dangerous content
  local sanitized_content
  sanitized_content=$(echo "$content" | \
    sed 's/<script[^>]*>.*<\/script>//gi' | \
    sed 's/<iframe[^>]*>.*<\/iframe>//gi' | \
    sed 's/<object[^>]*>.*<\/object>//gi' | \
    sed 's/<embed[^>]*>.*<\/embed>//gi' | \
    sed 's/<applet[^>]*>.*<\/applet>//gi' | \
    sed 's/<form[^>]*>.*<\/form>//gi' | \
    sed 's/<input[^>]*>//gi' | \
    sed 's/<button[^>]*>.*<\/button>//gi' | \
    sed 's/javascript:[^"'\'']*//gi' | \
    sed 's/vbscript:[^"'\'']*//gi' | \
    sed 's/data:[^"'\'']*//gi' | \
    sed 's/on[a-zA-Z]*=[^"'\'']*//gi' | \
    sed 's/<link[^>]*>//gi' | \
    sed 's/<meta[^>]*>//gi' | \
    sed 's/<style[^>]*>.*<\/style>//gi')
  
  # Remove potential HTML entities that could be used for injection
  sanitized_content=$(echo "$sanitized_content" | \
    sed 's/&lt;script/\&amp;lt;script/gi' | \
    sed 's/&gt;/\&amp;gt;/g' | \
    sed 's/&quot;/\&amp;quot;/g')
  
  # Remove null bytes and control characters
  sanitized_content=$(echo "$sanitized_content" | tr -d '\000-\010\013\014\016-\037')
  
  log "Content sanitization completed" "Original: ${#content} chars, Sanitized: ${#sanitized_content} chars"
  
  echo "$sanitized_content"
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

# Function to provide clear feedback when file output is disabled or fails
log_file_output_status() {
  local status=$1
  local reason="$2"
  local details="$3"
  
  case "$status" in
    "disabled")
      log "File output status" "DISABLED - $reason"
      log "Dual output mode" "Single output mode active (stdout only)"
      if [ -n "$details" ]; then
        log "File output details" "$details"
      fi
      log "Impact assessment" "GitHub features will use stdout content as fallback"
      ;;
    "failed")
      log "File output status" "FAILED - $reason"
      log "Dual output mode" "Fallback mode active (using stdout for all outputs)"
      if [ -n "$details" ]; then
        log "Failure details" "$details"
      fi
      log "Impact assessment" "GitHub features will use fallback content"
      ;;
    "partial")
      log "File output status" "PARTIAL - $reason"
      log "Dual output mode" "Degraded mode active (some file operations failed)"
      if [ -n "$details" ]; then
        log "Partial failure details" "$details"
      fi
      log "Impact assessment" "Some GitHub features may use fallback content"
      ;;
    "success")
      log "File output status" "SUCCESS - $reason"
      log "Dual output mode" "Full dual output active (all formats available)"
      if [ -n "$details" ]; then
        log "Success details" "$details"
      fi
      log "Impact assessment" "All GitHub features have optimized content"
      ;;
    *)
      log "File output status" "UNKNOWN - $reason"
      log "Dual output mode" "Status unclear"
      ;;
  esac
}

# Function to handle file operation errors with fallback mechanisms
handle_file_operation_error() {
  local operation=$1
  local file_path=$2
  local error_message="$3"
  local fallback_content="$4"
  
  warning "File operation failed: $operation on $file_path - $error_message"
  log "Implementing fallback mechanism" "Operation: $operation"
  
  # Provide clear feedback about the failure
  log_file_output_status "failed" "file operation error" "Operation: $operation, Error: $error_message"
  
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
    "validate_temp_file")
      warning "Temporary file validation failed, using fallback content"
      log "File validation failed" "Path: $file_path"
      
      if [ -n "$fallback_content" ]; then
        MARKDOWN_CONTENT="$fallback_content"
      else
        MARKDOWN_CONTENT="## ⚠️ File Validation Error

Temporary file validation failed: $file_path

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


# Enhanced function to run Strata with dual output
run_strata_dual_output() {
  local stdout_format=$1
  local plan_file=$2
  local show_details=$3
  
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
    log_file_output_status "disabled" "temporary file creation failed" "mktemp command failed or returned invalid path"
    
    local fallback_output
    fallback_output=$(run_strata "$stdout_format" "$plan_file" "$show_details")
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
    fallback_output=$(run_strata "$stdout_format" "$plan_file" "$show_details")
    echo "$fallback_output"
    return $?
  fi
  
  # Set up trap for cleanup (will be called on script exit, not function exit)
  # Note: We rely on the main script's EXIT trap to clean up temp files
  
  local cmd="$TEMP_DIR/$BINARY_NAME plan summary"
  
  # Add optional arguments
  if [ -n "$INPUT_CONFIG_FILE" ]; then
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
  
  # Try dual output first, fall back to single output if not supported
  cmd="$cmd --output $stdout_format"
  
  # Check if dual output is supported by testing the --file flag
  if "$TEMP_DIR/$BINARY_NAME" plan summary --help 2>&1 | grep -q -- "--file"; then
    log "Dual output supported" "Using --file flag for markdown output"
    cmd="$cmd --file $temp_markdown_file --file-format markdown"
  else
    log "Dual output not supported" "Falling back to single output mode"
    log_file_output_status "disabled" "Strata version does not support dual output" "Binary version: $("$TEMP_DIR/$BINARY_NAME" --version 2>/dev/null || echo 'unknown')"
  fi
  
  # Add plan file
  cmd="$cmd $plan_file"
  
  log "Executing Strata with dual output" "Command: $cmd"
  log "Output formats" "Stdout: $stdout_format, File: markdown"
  log "Temporary file path" "$temp_markdown_file"
  
  # Execute command and capture stdout with enhanced error handling
  log "Executing Strata command" "Starting dual output execution"
  local stdout_output
  stdout_output=$(eval "$cmd" 2>&1)
  local exit_code=$?
  
  # Log execution results
  log "Strata execution completed" "Exit code: $exit_code, Output size: ${#stdout_output} chars"
  
  # Handle execution errors with structured error content
  if [ $exit_code -ne 0 ]; then
    warning "Strata execution failed with exit code $exit_code"
    warning "Error output: $stdout_output"
    log "Dual output execution failed" "Both stdout and file output affected"
    
    # Use structured error content creation
    MARKDOWN_CONTENT=$(create_structured_error_content "strata_execution_failed" "$stdout_output" "$exit_code" "Command: $cmd")
    
    log "Generated structured error content for execution failure" "Exit code: $exit_code"
    
    # Ensure cleanup happens in error scenarios
    cleanup_temp_files
    
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
          log "Format conversion status" "SUCCESS - stdout ($stdout_format) and file (markdown) formats generated"
          log_file_output_status "success" "dual output generation completed" "Stdout: $stdout_format format, File: markdown format"
        else
          # Handle read failure
          log "Format conversion status" "PARTIAL FAILURE - markdown file read failed, using stdout fallback"
          handle_file_operation_error "read_temp_file" "$temp_markdown_file" "cat command failed (exit code: $read_result)" "$stdout_output"
        fi
      else
        # Handle size check or empty file
        if [ $size_check_result -ne 0 ]; then
          log "Format conversion status" "FAILURE - file size check failed"
          handle_file_operation_error "read_temp_file" "$temp_markdown_file" "wc command failed (exit code: $size_check_result)" "$stdout_output"
        else
          log "Format conversion status" "FAILURE - markdown file is empty"
          handle_file_operation_error "read_temp_file" "$temp_markdown_file" "file is empty (size: $file_size)" "$stdout_output"
        fi
      fi
    else
      # Handle unreadable file
      log "Format conversion status" "FAILURE - markdown file is not readable"
      handle_file_operation_error "read_temp_file" "$temp_markdown_file" "file is not readable" "$stdout_output"
    fi
  else
    # Handle missing file - this is expected when dual output is not supported
    log "Format conversion status" "SINGLE OUTPUT MODE - using stdout content for markdown"
    log_file_output_status "disabled" "dual output not supported by Strata version" "Using stdout content as fallback"
    
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
    log "Format conversion status" "FALLBACK - using stdout content for markdown output"
    MARKDOWN_CONTENT="$stdout_output"
  fi
  
  # Final status logging with comprehensive feedback
  local final_status
  if [ -n "$MARKDOWN_CONTENT" ] && [ "$MARKDOWN_CONTENT" != "$stdout_output" ]; then
    final_status="success"
    log_file_output_status "success" "dual output execution completed successfully" "Different content generated for stdout and file outputs"
  elif [ -n "$MARKDOWN_CONTENT" ]; then
    final_status="partial"
    log_file_output_status "partial" "dual output execution completed with fallback" "Same content used for both stdout and file outputs"
  else
    final_status="failed"
    log_file_output_status "failed" "dual output execution failed" "No markdown content available"
  fi
  
  log "Dual output execution completed" "Status: $final_status"
  log "Output formats summary" "Stdout: $stdout_format (${#stdout_output} chars), Markdown: ${#MARKDOWN_CONTENT} chars"
  
  echo "$stdout_output"
  return $exit_code
}

# Run Strata with dual output for optimal formatting with enhanced error handling
log "Starting Strata analysis with dual output system"
log "Output format configuration" "Terminal display: table format, GitHub features: markdown format"
log "Dual output system status" "INITIALIZING - preparing to generate multiple output formats"

# Execute with comprehensive error handling
log "Executing main Strata analysis" "Plan file: $INPUT_PLAN_FILE, Show details: $SHOW_DETAILS"
STRATA_OUTPUT=$(run_strata_dual_output "table" "$INPUT_PLAN_FILE" "$SHOW_DETAILS")
STRATA_EXIT_CODE=$?

# Log the results of dual output execution
if [ $STRATA_EXIT_CODE -eq 0 ]; then
  log "Dual output execution successful" "Exit code: $STRATA_EXIT_CODE"
  log "Output content generated" "Stdout: ${#STRATA_OUTPUT} chars, Markdown: ${#MARKDOWN_CONTENT} chars"
  log "Dual output system status" "SUCCESS - both formats available for distribution"
else
  log "Dual output execution failed" "Exit code: $STRATA_EXIT_CODE"
  log "Dual output system status" "FAILED - error handling activated"
  
  # Use enhanced error handling
  handle_dual_output_error $STRATA_EXIT_CODE "$STRATA_OUTPUT" "main_execution"
fi

# Handle execution failure with structured error content
if [ $STRATA_EXIT_CODE -ne 0 ]; then
  warning "Strata analysis failed with exit code $STRATA_EXIT_CODE"
  
  # Write structured error to step summary using markdown content
  if [ -n "$GITHUB_STEP_SUMMARY" ]; then
    if [ -n "$MARKDOWN_CONTENT" ]; then
      # Use the structured error content from handle_dual_output_error
      echo "$MARKDOWN_CONTENT" >> "$GITHUB_STEP_SUMMARY"
      log "Wrote structured error content to step summary" "Size: ${#MARKDOWN_CONTENT} chars"
    else
      # Fallback error content if markdown content is not available
      fallback_error_content=$(create_structured_error_content "strata_execution_failed" "$STRATA_OUTPUT" "$STRATA_EXIT_CODE" "Main execution phase")
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

# Always get JSON output for parsing, regardless of primary output format
log "Generating JSON output for statistics parsing" "Format: json, Purpose: statistics extraction"
JSON_OUTPUT=$(run_strata "json" "$INPUT_PLAN_FILE" "false")
JSON_EXIT_CODE=$?

if [ $JSON_EXIT_CODE -ne 0 ]; then
  warning "Failed to get JSON output for parsing, some features may not work correctly"
  log "JSON parsing error" "Exit code: $JSON_EXIT_CODE"
  log "Format conversion status" "JSON generation failed - using default statistics"
  
  # Log the JSON parsing failure but continue with defaults
  log "JSON parsing failed, using default values" "JSON generation failed with exit code $JSON_EXIT_CODE"
  
  # Set default values for parsing to ensure action continues
  HAS_CHANGES="false"
  HAS_DANGERS="false"
  CHANGE_COUNT="0"
  DANGER_COUNT="0"
  ADD_COUNT="0"
  CHANGE_COUNT_DETAIL="0"
  DESTROY_COUNT="0"
  REPLACE_COUNT="0"
  
  # Log JSON error for debugging
  log "JSON parsing error content created" "For debugging purposes"
else
  log "JSON output retrieved successfully" "Size: ${#JSON_OUTPUT} chars"
  log "Format conversion status" "JSON generation successful - statistics parsing enabled"
fi

# Function to extract value from JSON
extract_json_value() {
  local json=$1
  local key=$2
  local default=$3
  
  local value
  if command -v jq >/dev/null 2>&1; then
    # Use jq if available for more reliable parsing
    value=$(echo "$json" | jq -r ".$key" 2>/dev/null)
    if [ "$value" = "null" ] || [ -z "$value" ]; then
      value="$default"
    fi
  else
    # Fallback to grep for basic parsing
    value=$(echo "$json" | grep -o "\"$key\":[^,}]*" | cut -d':' -f2 | tr -d '"{}[],' 2>/dev/null)
    if [ -z "$value" ]; then
      value="$default"
    fi
  fi
  
  echo "$value"
}

# Parse JSON output for statistics
HAS_CHANGES=$(extract_json_value "$JSON_OUTPUT" "hasChanges" "false")
HAS_DANGERS=$(extract_json_value "$JSON_OUTPUT" "hasDangers" "false")
CHANGE_COUNT=$(extract_json_value "$JSON_OUTPUT" "totalChanges" "0")
DANGER_COUNT=$(extract_json_value "$JSON_OUTPUT" "dangerCount" "0")

# Extract additional statistics if available
ADD_COUNT=$(extract_json_value "$JSON_OUTPUT" "addCount" "0")
CHANGE_COUNT_DETAIL=$(extract_json_value "$JSON_OUTPUT" "changeCount" "0")
DESTROY_COUNT=$(extract_json_value "$JSON_OUTPUT" "destroyCount" "0")
REPLACE_COUNT=$(extract_json_value "$JSON_OUTPUT" "replaceCount" "0")

# Prepare step summary content using markdown processing
STEP_SUMMARY_BASE_CONTENT=""

# Add plan status indicators
if [ "$HAS_CHANGES" = "true" ]; then
  if [ "$HAS_DANGERS" = "true" ]; then
    STEP_SUMMARY_BASE_CONTENT="⚠️ **Plan contains changes with potential risks**

"
  else
    STEP_SUMMARY_BASE_CONTENT="✅ **Plan contains changes**

"
  fi
else
  STEP_SUMMARY_BASE_CONTENT="ℹ️ **Plan contains no changes**

"
fi

# Add statistics summary table
STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}## Statistics Summary
| TO ADD | TO CHANGE | TO DESTROY | REPLACEMENTS | HIGH RISK |
|--------|-----------|------------|--------------|----------|
| $ADD_COUNT | $CHANGE_COUNT_DETAIL | $DESTROY_COUNT | $REPLACE_COUNT | $DANGER_COUNT |

"

# Add main output using markdown content for better formatting
STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}## Resource Changes
"
if [ -n "$MARKDOWN_CONTENT" ]; then
  STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}${MARKDOWN_CONTENT}"
else
  STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}${STRATA_OUTPUT}"
fi
STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}

"

# Add detailed information in collapsible section if available
if [ "$SHOW_DETAILS" = "true" ]; then
  # Get detailed output using dual output for consistent formatting
  DETAILED_STDOUT=$(run_strata_dual_output "table" "$INPUT_PLAN_FILE" "true")
  
  STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}<details>
<summary>📋 Detailed Changes</summary>

"
  # Use markdown content if available, otherwise use stdout
  if [ -n "$MARKDOWN_CONTENT" ]; then
    STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}${MARKDOWN_CONTENT}"
  else
    STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}${DETAILED_STDOUT}"
  fi
  STEP_SUMMARY_BASE_CONTENT="${STEP_SUMMARY_BASE_CONTENT}

</details>

"
fi

# Function to distribute content to different GitHub contexts
distribute_output() {
  local stdout_output="$1"
  local markdown_content="$2"
  
  log "Starting output distribution to GitHub contexts" "Stdout size: ${#stdout_output} chars, Markdown size: ${#markdown_content} chars"
  log "Distribution targets" "Step Summary: $([ -n "$GITHUB_STEP_SUMMARY" ] && echo "ENABLED" || echo "DISABLED"), PR Comments: $([ "$COMMENT_ON_PR" = "true" ] && echo "ENABLED" || echo "DISABLED")"
  
  # Prepare base content with statistics for both contexts
  local base_content=""
  
  # Add plan status indicators
  if [ "$HAS_CHANGES" = "true" ]; then
    if [ "$HAS_DANGERS" = "true" ]; then
      base_content="⚠️ **Plan contains changes with potential risks**

"
    else
      base_content="✅ **Plan contains changes**

"
    fi
  else
    base_content="ℹ️ **Plan contains no changes**

"
  fi
  
  # Add statistics summary table for step summary
  local step_summary_stats="## Statistics Summary
| TO ADD | TO CHANGE | TO DESTROY | REPLACEMENTS | HIGH RISK |
|--------|-----------|------------|--------------|----------|
| $ADD_COUNT | $CHANGE_COUNT_DETAIL | $DESTROY_COUNT | $REPLACE_COUNT | $DANGER_COUNT |

"
  
  # Add PR comment statistics (more compact format)
  local pr_comment_stats="**Statistics:**
- 📝 **Changes**: $CHANGE_COUNT
- ⚠️ **Dangerous**: $DANGER_COUNT
- 🔄 **Replacements**: $REPLACE_COUNT

"
  
  # Prepare main content section
  local main_content_section="## Resource Changes
"
  if [ -n "$markdown_content" ]; then
    main_content_section="${main_content_section}${markdown_content}"
  else
    main_content_section="${main_content_section}${stdout_output}"
  fi
  main_content_section="${main_content_section}

"
  
  # Add detailed information in collapsible section if available
  local details_section=""
  if [ "$SHOW_DETAILS" = "true" ]; then
    # Get detailed output using dual output for consistent formatting
    local detailed_stdout
    detailed_stdout=$(run_strata_dual_output "table" "$INPUT_PLAN_FILE" "true")
    
    details_section="<details>
<summary>📋 Detailed Changes</summary>

"
    # Use markdown content if available, otherwise use stdout
    if [ -n "$MARKDOWN_CONTENT" ]; then
      details_section="${details_section}${MARKDOWN_CONTENT}"
    else
      details_section="${details_section}${detailed_stdout}"
    fi
    details_section="${details_section}

</details>

"
  fi
  
  # Write to GitHub Step Summary using processed markdown content
  if [ -n "$GITHUB_STEP_SUMMARY" ]; then
    log "Processing content for GitHub Step Summary" "Target: $GITHUB_STEP_SUMMARY"
    
    # Prepare step summary content
    local step_summary_content="${base_content}${step_summary_stats}${main_content_section}${details_section}"
    
    # Process content for step summary context
    local processed_step_summary
    processed_step_summary=$(process_markdown_for_context "step-summary" "$step_summary_content")
    
    # Optimize and sanitize content
    local optimized_step_summary
    optimized_step_summary=$(echo "$processed_step_summary" | optimize_content_for_context "step-summary")
    local sanitized_step_summary
    sanitized_step_summary=$(sanitize_github_content "$optimized_step_summary")
    
    # Write to step summary
    echo "$sanitized_step_summary" >> "$GITHUB_STEP_SUMMARY"
    log "Step summary written successfully" "Size: ${#sanitized_step_summary} chars"
    log "Step summary distribution completed" "Format: markdown, Context: step-summary"
  else
    log "GitHub Step Summary not available" "GITHUB_STEP_SUMMARY environment variable not set"
    log "Step summary distribution skipped" "Reason: GitHub Step Summary disabled"
  fi
  
  # Handle PR comments if enabled
  if [ "$COMMENT_ON_PR" = "true" ] && [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
    log "Processing content for PR comment" "Event: $GITHUB_EVENT_NAME"
    
    # Extract PR number from GitHub event
    local pr_number
    pr_number=$(jq -r .pull_request.number "$GITHUB_EVENT_PATH" 2>/dev/null)
    
    if [ -n "$pr_number" ] && [ "$pr_number" != "null" ]; then
      # Prepare PR comment content (more compact than step summary)
      local pr_comment_content="${base_content}${pr_comment_stats}**Plan Summary:**
"
      
      # Use markdown content if available, otherwise use stdout
      if [ -n "$markdown_content" ]; then
        pr_comment_content="${pr_comment_content}${markdown_content}"
      else
        pr_comment_content="${pr_comment_content}${stdout_output}"
      fi
      
      pr_comment_content="${pr_comment_content}

${details_section}"
      
      # Process content for PR comment context
      local processed_pr_comment
      processed_pr_comment=$(process_markdown_for_context "pr-comment" "$pr_comment_content")
      
      # Optimize and sanitize content
      local optimized_pr_comment
      optimized_pr_comment=$(echo "$processed_pr_comment" | optimize_content_for_context "pr-comment")
      local sanitized_pr_comment
      sanitized_pr_comment=$(sanitize_github_content "$optimized_pr_comment")
      
      # Post the PR comment
      log "Posting PR comment" "PR number: $pr_number, Content size: ${#sanitized_pr_comment} chars"
      post_pr_comment "$pr_number" "$sanitized_pr_comment"
      log "PR comment distribution completed" "Format: markdown, Context: pr-comment"
    else
      warning "Could not determine PR number, skipping PR comment"
      log "PR comment distribution skipped" "Reason: PR number not available"
    fi
  else
    if [ "$COMMENT_ON_PR" != "true" ]; then
      log "PR commenting disabled" "COMMENT_ON_PR setting: $COMMENT_ON_PR"
    else
      log "Not in pull request context" "GITHUB_EVENT_NAME: $GITHUB_EVENT_NAME"
    fi
    log "PR comment distribution skipped" "Reason: disabled or wrong context"
  fi
  
  # Set action outputs with appropriate content for each output type
  log "Setting GitHub Action outputs" "Preparing outputs for workflow consumption"
  
  set_output "summary" "$stdout_output"
  log "Action output set" "summary: ${#stdout_output} chars (table format)"
  
  set_output "has-changes" "$HAS_CHANGES"
  log "Action output set" "has-changes: $HAS_CHANGES"
  
  set_output "has-dangers" "$HAS_DANGERS"
  log "Action output set" "has-dangers: $HAS_DANGERS"
  
  # Keep json-summary as JSON for backward compatibility and programmatic access
  set_output "json-summary" "$JSON_OUTPUT"
  log "Action output set" "json-summary: ${#JSON_OUTPUT} chars (JSON format)"
  
  set_output "change-count" "$CHANGE_COUNT"
  log "Action output set" "change-count: $CHANGE_COUNT"
  
  set_output "danger-count" "$DANGER_COUNT"
  log "Action output set" "danger-count: $DANGER_COUNT"
  
  # Set markdown content as additional output for GitHub features
  if [ -n "$MARKDOWN_CONTENT" ]; then
    set_output "markdown-summary" "$MARKDOWN_CONTENT"
    log "Action output set" "markdown-summary: ${#MARKDOWN_CONTENT} chars (markdown format)"
  else
    log "Markdown summary output skipped" "No markdown content available"
  fi
  
  log "GitHub Action outputs completed" "All outputs set successfully"
  log "Output distribution completed successfully" "All targets processed"
}

# Function to post PR comment with proper error handling
post_pr_comment() {
  local pr_number=$1
  local comment_body="$2"
  
  log "Posting PR comment" "PR: $pr_number, Size: ${#comment_body} chars"
  
  # Check if we should update existing comment
  if [ "$UPDATE_COMMENT" = "true" ]; then
    # Function to make GitHub API request with retry
    github_api_request() {
      local method=$1
      local url=$2
      local data=$3
      local max_attempts=3
      local attempt=1
      local timeout=10
      local response
      
      while [ $attempt -le $max_attempts ]; do
        log "API request attempt $attempt/$max_attempts" "$method $url"
        
        # Check for rate limiting
        local rate_limit_check
        rate_limit_check=$(curl -s -I -H "Authorization: token $GITHUB_TOKEN" "${GITHUB_API_URL}/rate_limit")
        local remaining
        remaining=$(echo "$rate_limit_check" | grep -i "x-ratelimit-remaining" | cut -d':' -f2 | tr -d ' \r\n')
        
        if [ -n "$remaining" ] && [ "$remaining" -le 5 ]; then
          local reset_time
          reset_time=$(echo "$rate_limit_check" | grep -i "x-ratelimit-reset" | cut -d':' -f2 | tr -d ' \r\n')
          local current_time
          current_time=$(date +%s)
          local wait_time
          wait_time=$((reset_time - current_time + 5))
          
          if [ "$wait_time" -gt 0 ]; then
            warning "Rate limit almost reached ($remaining remaining). Waiting for $wait_time seconds before retry."
            sleep "$wait_time"
          fi
        fi
        
        # Make the request
        if [ -z "$data" ]; then
          response=$(curl -s -w "%{http_code}" -X "$method" -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" "$url")
        else
          response=$(curl -s -w "%{http_code}" -X "$method" -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" -d "$data" "$url")
        fi
        
        local http_status
        http_status=${response: -3}
        
        # Check for success
        if [ "$http_status" -lt 400 ]; then
          echo "$response"
          return 0
        fi
        
        # Check for rate limiting
        if [ "$http_status" -eq 403 ] && echo "${response%???}" | grep -q "rate limit"; then
          local reset_time
          reset_time=$(echo "${response%???}" | jq -r '.rate.reset' 2>/dev/null || echo 0)
          local current_time
          current_time=$(date +%s)
          local wait_time
          wait_time=$((reset_time - current_time + 5))
          
          if [ "$wait_time" -gt 0 ]; then
            warning "Rate limit reached. Waiting for $wait_time seconds before retry."
            sleep "$wait_time"
            attempt=$((attempt))
            continue
          fi
        fi
        
        # For other errors, retry with backoff
        attempt=$((attempt + 1))
        if [ $attempt -le $max_attempts ]; then
          sleep_time=$((2 ** (attempt - 1)))
          warning "API request failed with status $http_status. Retrying in ${sleep_time}s"
          sleep $sleep_time
        else
          warning "API request failed after $max_attempts attempts with status $http_status"
          echo "$response"
          return 1
        fi
      done
      
      return 1
    }
    
    # Search for existing comment
    local comments_url="${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments"
    log "Searching for existing comments" "URL: $comments_url"
    
    local comments_response
    comments_response=$(github_api_request "GET" "$comments_url")
    local http_status=${comments_response: -3}
    
    if [ "$http_status" -ge 400 ]; then
      warning "Failed to fetch existing comments, creating new comment instead"
      local comment_id=""
    else
      local comments=${comments_response%???}
      local comment_id
      comment_id=$(echo "$comments" | jq -r ".[] | select(.body | contains(\"strata-comment-id: $GITHUB_WORKFLOW-$GITHUB_JOB\")) | .id" 2>/dev/null)
    fi
    
    if [ -n "$comment_id" ]; then
      # Update existing comment
      log "Updating existing comment" "ID: $comment_id"
      local comment_url="${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}"
      local response
      response=$(github_api_request "PATCH" "$comment_url" "{\"body\": $(echo "$comment_body" | jq -R -s .)}")
      local update_status=${response: -3}
      
      if [ "$update_status" -ge 400 ]; then
        warning "Failed to update comment (HTTP $update_status), creating new comment instead"
        local error_details="HTTP $update_status: ${response%???}"
        log "Comment update error details" "$error_details"
        
        # Try to create new comment as fallback
        local fallback_response
        fallback_response=$(github_api_request "POST" "$comments_url" "{\"body\": $(echo "$comment_body" | jq -R -s .)}")
        local fallback_status=${fallback_response: -3}
        
        if [ "$fallback_status" -ge 400 ]; then
          # Both update and create failed
          warning "Both comment update and creation failed"
          log "GitHub API operations failed" "Update and create both failed"
        else
          log "Successfully created new comment after update failure"
        fi
      else
        log "Successfully updated existing comment"
      fi
    else
      # Create new comment
      log "Creating new comment"
      local response
      response=$(github_api_request "POST" "$comments_url" "{\"body\": $(echo "$comment_body" | jq -R -s .)}")
      local create_status=${response: -3}
      
      if [ "$create_status" -ge 400 ]; then
        local error_details="HTTP $create_status: ${response%???}"
        warning "Failed to create comment: $error_details"
        
        # Log GitHub API failure
        log "GitHub API comment creation failed" "HTTP status: $create_status"
      else
        log "Successfully created new comment"
      fi
    fi
  else
    # Always create new comment
    log "Creating new comment"
    local comments_url="${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments"
    local response
    response=$(github_api_request "POST" "$comments_url" "{\"body\": $(echo "$comment_body" | jq -R -s .)}")
    local create_status=${response: -3}
    
    if [ "$create_status" -ge 400 ]; then
      warning "Failed to create comment (HTTP $create_status)"
    fi
  fi
}

# Use the new output distribution system with error handling
log "Initiating output distribution phase" "Distributing content to GitHub contexts"
distribute_output "$STRATA_OUTPUT" "$MARKDOWN_CONTENT"
log "Output distribution phase completed" "All content distributed successfully"

# Perform final cleanup before exit with secure cleanup for sensitive files
log "Initiating final cleanup phase" "Cleaning up temporary files and resources securely"
cleanup_temp_files "true"
cleanup_result=$?

if [ $cleanup_result -eq 0 ]; then
  log "Final secure cleanup completed successfully" "All temporary resources cleaned securely"
else
  log "Final secure cleanup completed with warnings" "Some cleanup operations failed"
fi

# Log comprehensive final status
log "Action execution summary" "Strata exit code: $STRATA_EXIT_CODE, Cleanup result: $cleanup_result"

if [ $STRATA_EXIT_CODE -eq 0 ]; then
  log "GitHub Action completed successfully" "Exit code: $STRATA_EXIT_CODE"
  log "Dual output system final status" "SUCCESS - all formats generated and distributed"
else
  log "GitHub Action completed with errors" "Exit code: $STRATA_EXIT_CODE"
  log "Dual output system final status" "PARTIAL - error handling provided fallback content"
fi

# Provide clear feedback about what formats were used
if [ -n "$MARKDOWN_CONTENT" ] && [ "$MARKDOWN_CONTENT" != "$STRATA_OUTPUT" ]; then
  log "Format usage summary" "Stdout: table format, GitHub features: markdown format (dual output successful)"
else
  log "Format usage summary" "Stdout: table format, GitHub features: fallback content (dual output partial)"
fi

# Exit with Strata's exit code
exit $STRATA_EXIT_CODE