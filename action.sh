#!/bin/bash
set -e

# Function to log messages
log() {
  echo "::group::$1"
  shift
  echo "$@"
  echo "::endgroup::"
}

# Function to log errors
error() {
  echo "::error::$1"
  
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
  echo "::warning::$1"
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

# Validate inputs
log "Validating inputs"

# Check required inputs
if [ -z "$INPUT_PLAN_FILE" ]; then
  error "Plan file is required. Please specify the 'plan-file' input parameter."
fi

# Validate plan file
if [ ! -f "$INPUT_PLAN_FILE" ]; then
  error "Plan file does not exist: $INPUT_PLAN_FILE. Please check the path and ensure the file is generated before running this action."
fi

# Check if plan file is readable
if [ ! -r "$INPUT_PLAN_FILE" ]; then
  error "Plan file is not readable: $INPUT_PLAN_FILE. Please check file permissions."
fi

# Validate output format
OUTPUT_FORMAT="${INPUT_OUTPUT_FORMAT:-markdown}"
case "$OUTPUT_FORMAT" in
  markdown|json|table)
    log "Using output format: $OUTPUT_FORMAT"
    ;;
  *)
    warning "Invalid output format: $OUTPUT_FORMAT. Valid options are: markdown, json, table. Defaulting to markdown."
    OUTPUT_FORMAT="markdown"
    ;;
esac

# Validate boolean inputs
SHOW_DETAILS="${INPUT_SHOW_DETAILS:-false}"
if [ "$SHOW_DETAILS" != "true" ] && [ "$SHOW_DETAILS" != "false" ]; then
  warning "Invalid value for show-details: $SHOW_DETAILS. Expected 'true' or 'false'. Defaulting to 'false'."
  SHOW_DETAILS="false"
fi

COMMENT_ON_PR="${INPUT_COMMENT_ON_PR:-true}"
if [ "$COMMENT_ON_PR" != "true" ] && [ "$COMMENT_ON_PR" != "false" ]; then
  warning "Invalid value for comment-on-pr: $COMMENT_ON_PR. Expected 'true' or 'false'. Defaulting to 'true'."
  COMMENT_ON_PR="true"
fi

UPDATE_COMMENT="${INPUT_UPDATE_COMMENT:-true}"
if [ "$UPDATE_COMMENT" != "true" ] && [ "$UPDATE_COMMENT" != "false" ]; then
  warning "Invalid value for update-comment: $UPDATE_COMMENT. Expected 'true' or 'false'. Defaulting to 'true'."
  UPDATE_COMMENT="true"
fi

# Validate danger threshold if provided
if [ -n "$INPUT_DANGER_THRESHOLD" ]; then
  if ! [[ "$INPUT_DANGER_THRESHOLD" =~ ^[0-9]+$ ]]; then
    warning "Invalid danger threshold: $INPUT_DANGER_THRESHOLD. Expected a positive integer. The default value will be used."
    INPUT_DANGER_THRESHOLD=""
  fi
fi

# Validate config file if provided
if [ -n "$INPUT_CONFIG_FILE" ] && [ ! -f "$INPUT_CONFIG_FILE" ]; then
  warning "Config file does not exist: $INPUT_CONFIG_FILE. The default configuration will be used."
  INPUT_CONFIG_FILE=""
fi

# Set other default values
COMMENT_HEADER="${INPUT_COMMENT_HEADER:-🏗️ Terraform Plan Summary}"

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
trap 'rm -rf "$TEMP_DIR"' EXIT

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
        EXPECTED_CHECKSUM=$(cat "$TEMP_DIR/checksum.md5" | cut -d' ' -f1)
        
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
    cmd="$cmd --show-details"
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

# Run Strata for primary output format
log "Running Strata analysis"
STRATA_OUTPUT=$(run_strata "$OUTPUT_FORMAT" "$INPUT_PLAN_FILE" "$SHOW_DETAILS")
STRATA_EXIT_CODE=$?

if [ $STRATA_EXIT_CODE -ne 0 ]; then
  warning "Strata analysis failed with exit code $STRATA_EXIT_CODE"
  
  # Write error to step summary
  if [ -n "$GITHUB_STEP_SUMMARY" ]; then
    {
      echo "## ⚠️ Strata Analysis Warning"
      echo ""
      echo "Strata encountered an issue while analyzing the Terraform plan."
      echo ""
      echo "**Possible causes:**"
      echo "- Invalid plan file format"
      echo "- Unsupported Terraform version"
      echo "- Plan file corruption"
      echo ""
      echo "Please check the action logs for more details."
    } >> "$GITHUB_STEP_SUMMARY"
  fi
  
  # Set fallback output
  STRATA_OUTPUT="Error: Failed to analyze Terraform plan. Please check the logs for details."
fi

# Always get JSON output for parsing, regardless of primary output format
log "Getting JSON output for parsing"
JSON_OUTPUT=$(run_strata "json" "$INPUT_PLAN_FILE" "false")
JSON_EXIT_CODE=$?

if [ $JSON_EXIT_CODE -ne 0 ]; then
  warning "Failed to get JSON output for parsing, some features may not work correctly"
  # Set default values for parsing
  HAS_CHANGES="false"
  HAS_DANGERS="false"
  CHANGE_COUNT="0"
  DANGER_COUNT="0"
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

# Write to GitHub Step Summary
write_summary "# $COMMENT_HEADER"
write_summary ""

# Add plan status indicators
if [ "$HAS_CHANGES" = "true" ]; then
  if [ "$HAS_DANGERS" = "true" ]; then
    write_summary "⚠️ **Plan contains changes with potential risks**"
  else
    write_summary "✅ **Plan contains changes**"
  fi
else
  write_summary "ℹ️ **Plan contains no changes**"
fi
write_summary ""

# Add statistics summary table
write_summary "## Statistics Summary"
write_summary "| TO ADD | TO CHANGE | TO DESTROY | REPLACEMENTS | HIGH RISK |"
write_summary "|--------|-----------|------------|--------------|----------|"
write_summary "| $ADD_COUNT | $CHANGE_COUNT_DETAIL | $DESTROY_COUNT | $REPLACE_COUNT | $DANGER_COUNT |"
write_summary ""

# Add main output
write_summary "## Resource Changes"
write_summary "$STRATA_OUTPUT"
write_summary ""

# Add detailed information in collapsible section if available
if [ "$SHOW_DETAILS" = "true" ]; then
  DETAILED_OUTPUT=$(run_strata "$OUTPUT_FORMAT" "$INPUT_PLAN_FILE" "true")
  
  write_summary "<details>"
  write_summary "<summary>📋 Detailed Changes</summary>"
  write_summary ""
  write_summary "$DETAILED_OUTPUT"
  write_summary ""
  write_summary "</details>"
  write_summary ""
fi

# Add workflow information
write_summary "<details>"
write_summary "<summary>ℹ️ Workflow Information</summary>"
write_summary ""
write_summary "- **Repository:** $GITHUB_REPOSITORY"
write_summary "- **Workflow:** $GITHUB_WORKFLOW"
write_summary "- **Run ID:** $GITHUB_RUN_ID"
write_summary "- **Strata Version:** $("$TEMP_DIR"/$BINARY_NAME --version | head -n 1)"
write_summary ""
write_summary "</details>"
write_summary ""

write_summary "---"
write_summary "*Generated by [Strata](https://github.com/ArjenSchwarz/strata)*"

# Set outputs
set_output "summary" "$STRATA_OUTPUT"
set_output "has-changes" "$HAS_CHANGES"
set_output "has-dangers" "$HAS_DANGERS"
set_output "json-summary" "$JSON_OUTPUT"
set_output "change-count" "$CHANGE_COUNT"
set_output "danger-count" "$DANGER_COUNT"

# Check if we should comment on PR
if [ "$COMMENT_ON_PR" = "true" ] && [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
  log "Adding comment to PR"
  
  # Extract PR number from GitHub event
  PR_NUMBER=$(jq -r .pull_request.number "$GITHUB_EVENT_PATH" 2>/dev/null)
  
  if [ -z "$PR_NUMBER" ] || [ "$PR_NUMBER" = "null" ]; then
    warning "Could not determine PR number, skipping comment"
  else
    # Prepare status indicator
    if [ "$HAS_CHANGES" = "true" ]; then
      if [ "$HAS_DANGERS" = "true" ]; then
        STATUS_INDICATOR="⚠️ **Plan contains changes with potential risks**"
      else
        STATUS_INDICATOR="✅ **Plan contains changes**"
      fi
    else
      STATUS_INDICATOR="ℹ️ **Plan contains no changes**"
    fi
    
    # Prepare statistics section
    STATS_SECTION="**Statistics:**
- 📝 **Changes**: $CHANGE_COUNT
- ⚠️ **Dangerous**: $DANGER_COUNT
- 🔄 **Replacements**: $REPLACE_COUNT"
    
    # Prepare detailed changes section
    if [ "$SHOW_DETAILS" = "true" ]; then
      DETAILED_OUTPUT=$(run_strata "$OUTPUT_FORMAT" "$INPUT_PLAN_FILE" "true")
      DETAILS_SECTION="<details>
<summary>📋 Detailed Changes</summary>

$DETAILED_OUTPUT

</details>"
    else
      DETAILS_SECTION=""
    fi
    
    # Prepare comment body
    COMMENT_BODY="## $COMMENT_HEADER

<!-- strata-comment-id: $GITHUB_WORKFLOW-$GITHUB_JOB -->

$STATUS_INDICATOR

$STATS_SECTION

**Plan Summary:**
$STRATA_OUTPUT

$DETAILS_SECTION

---
*Generated by [Strata](https://github.com/ArjenSchwarz/strata) in [workflow run](${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID})*"
    
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
      COMMENTS_URL="${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments"
      log "Searching for existing comments" "URL: $COMMENTS_URL"
      
      COMMENTS_RESPONSE=$(github_api_request "GET" "$COMMENTS_URL")
      HTTP_STATUS=${COMMENTS_RESPONSE: -3}
      
      if [ "$HTTP_STATUS" -ge 400 ]; then
        warning "Failed to fetch existing comments, creating new comment instead"
        COMMENT_ID=""
      else
        COMMENTS=${COMMENTS_RESPONSE%???}
        COMMENT_ID=$(echo "$COMMENTS" | jq -r ".[] | select(.body | contains(\"strata-comment-id: $GITHUB_WORKFLOW-$GITHUB_JOB\")) | .id" 2>/dev/null)
      fi
      
      if [ -n "$COMMENT_ID" ]; then
        # Update existing comment
        log "Updating existing comment" "ID: $COMMENT_ID"
        COMMENT_URL="${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/comments/${COMMENT_ID}"
        RESPONSE=$(github_api_request "PATCH" "$COMMENT_URL" "{\"body\": $(echo "$COMMENT_BODY" | jq -R -s .)}")
        HTTP_STATUS=${RESPONSE: -3}
        
        if [ "$HTTP_STATUS" -ge 400 ]; then
          warning "Failed to update comment (HTTP $HTTP_STATUS), creating new comment instead"
          github_api_request "POST" "$COMMENTS_URL" "{\"body\": $(echo "$COMMENT_BODY" | jq -R -s .)}"
        fi
      else
        # Create new comment
        log "Creating new comment"
        RESPONSE=$(github_api_request "POST" "$COMMENTS_URL" "{\"body\": $(echo "$COMMENT_BODY" | jq -R -s .)}")
        HTTP_STATUS=${RESPONSE: -3}
        
        if [ "$HTTP_STATUS" -ge 400 ]; then
          warning "Failed to create comment (HTTP $HTTP_STATUS)"
        fi
      fi
    else
      # Always create new comment
      log "Creating new comment"
      RESPONSE=$(github_api_request "POST" "$COMMENTS_URL" "{\"body\": $(echo "$COMMENT_BODY" | jq -R -s .)}")
      HTTP_STATUS=${RESPONSE: -3}
      
      if [ "$HTTP_STATUS" -ge 400 ]; then
        warning "Failed to create comment (HTTP $HTTP_STATUS)"
      fi
    fi
  fi
fi

# Exit with Strata's exit code
exit $STRATA_EXIT_CODE