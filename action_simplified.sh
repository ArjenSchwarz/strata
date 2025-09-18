#!/bin/bash
#
# Strata GitHub Action - Simplified Version
# A single-file implementation for reliability and clarity
#
# Exit codes:
# 0: Success
# 1: General failure
# 2: Invalid input
# 3: Download failure
# 4: Analysis failure
# 5: GitHub integration failure

set -euo pipefail

# ============================================================================
# Constants and Environment Setup
# ============================================================================

readonly GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"

# Create temporary directory for all operations
readonly TEMP_DIR
TEMP_DIR=$(mktemp -d)

# ============================================================================
# Error Handling and Cleanup
# ============================================================================

# Cleanup function for exit trap
cleanup() {
  local exit_code=$?

  # Set default outputs on failure
  if [[ $exit_code -ne 0 ]] && [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    {
      echo "has-changes=false"
      echo "has-dangers=false"
      echo "change-count=0"
      echo "danger-count=0"
      echo "summary=Analysis failed"
      echo "json-summary={}"
    } >> "$GITHUB_OUTPUT"
  fi

  # Remove temporary directory
  [[ -d "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"

  # Exit with the original exit code
  exit $exit_code
}

# Set up trap for cleanup on exit
trap cleanup EXIT

# ============================================================================
# Logging Functions
# ============================================================================

# Log informational message
log_info() {
  echo "ℹ️  $*" >&2
}

# Log success message
log_success() {
  echo "✅ $*" >&2
}

# Log error message and exit
log_error() {
  local message="$1"
  local exit_code="${2:-1}"
  echo "❌ $message" >&2
  exit "$exit_code"
}

# Log warning message
log_warning() {
  echo "⚠️  $*" >&2
}

# Log action start
log_start() {
  echo "🚀 $*" >&2
}

# Log download operation
log_download() {
  echo "⬇️  $*" >&2
}

# Log analysis operation
log_analyze() {
  echo "🔍 $*" >&2
}

# Log configuration operation
log_config() {
  echo "⚙️  $*" >&2
}

# ============================================================================
# Input Validation Functions
# ============================================================================

# Validate required inputs exist
validate_required_inputs() {
  if [[ -z "${INPUT_PLAN_FILE:-}" ]]; then
    log_error "Plan file is required. Please specify the 'plan-file' input parameter." 2
  fi
}

# Validate file exists and is readable
validate_file_exists() {
  local file="$1"
  local file_type="${2:-File}"

  if [[ ! -f "$file" ]]; then
    log_error "$file_type not found: $file" 2
  fi

  if [[ ! -r "$file" ]]; then
    log_error "$file_type not readable: $file" 2
  fi
}

# Validate output format
validate_output_format() {
  local format="$1"

  if [[ ! "$format" =~ ^(table|json|markdown|html)$ ]]; then
    log_error "Invalid output format: $format. Must be one of: table, json, markdown, html" 2
  fi
}

# Basic path security validation
validate_path_security() {
  local path="$1"
  local path_type="${2:-Path}"

  # Check for path traversal
  if [[ "$path" =~ \.\. ]]; then
    log_error "$path_type contains path traversal: $path" 2
  fi

  # Check for excessive length
  if [[ ${#path} -gt 4096 ]]; then
    log_error "$path_type too long (max 4096 characters): ${#path}" 2
  fi
}

# ============================================================================
# Platform Detection Functions
# ============================================================================

# Detect operating system
detect_os() {
  case "$(uname -s)" in
    Linux*) echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)
      log_error "Unsupported operating system: $(uname -s)" 1
      ;;
  esac
}

# Detect CPU architecture
detect_arch() {
  case "$(uname -m)" in
    x86_64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *)
      log_error "Unsupported architecture: $(uname -m)" 1
      ;;
  esac
}

# ============================================================================
# Binary Download Functions
# ============================================================================

# Download Strata binary with retry and checksum verification
download_strata() {
  local version="${1:-latest}"
  local os
  local arch
  os=$(detect_os)
  arch=$(detect_arch)

  # Construct download URLs
  local base_url="https://github.com/ArjenSchwarz/strata/releases"
  local binary_url
  local checksum_url

  if [[ "$version" == "latest" ]]; then
    binary_url="$base_url/latest/download/strata-${os}-${arch}.tar.gz"
    checksum_url="$base_url/latest/download/checksums.txt"
    log_info "Using latest Strata release"
  else
    binary_url="$base_url/download/${version}/strata-${os}-${arch}.tar.gz"
    checksum_url="$base_url/download/${version}/checksums.txt"
    log_info "Using Strata version: $version"
  fi

  log_download "Downloading Strata ${version} for ${os}/${arch}"

  # Download with retry (max 3 attempts)
  local attempt
  for attempt in 1 2 3; do
    # Download binary
    if curl -fsSL "$binary_url" -o "$TEMP_DIR/strata.tar.gz" 2>/dev/null; then
      # Download checksums
      if curl -fsSL "$checksum_url" -o "$TEMP_DIR/checksums.txt" 2>/dev/null; then
        # Verify checksum
        local expected
        local actual
        expected=$(grep "strata-${os}-${arch}.tar.gz" "$TEMP_DIR/checksums.txt" | cut -d' ' -f1)
        actual=$(sha256sum "$TEMP_DIR/strata.tar.gz" | cut -d' ' -f1)

        if [[ "$expected" == "$actual" ]]; then
          # Extract verified binary
          tar -xz -C "$TEMP_DIR" -f "$TEMP_DIR/strata.tar.gz"
          chmod +x "$TEMP_DIR/strata"
          log_success "Download and verification successful"

          # Verify version
          "$TEMP_DIR/strata" --version
          return 0
        else
          log_warning "Checksum mismatch on attempt $attempt/3"
        fi
      else
        log_warning "Failed to download checksums on attempt $attempt/3"
      fi
    else
      log_warning "Failed to download binary on attempt $attempt/3"
    fi

    # Wait before retry (except on last attempt)
    [[ $attempt -lt 3 ]] && sleep 2
  done

  # If specific version failed and we're not already using latest, try fallback
  if [[ "$version" != "latest" ]]; then
    log_warning "Version $version not found, falling back to latest"
    download_strata "latest"
    return $?
  fi

  log_error "Failed to download Strata from: $binary_url" 3
}

# ============================================================================
# Strata Execution Functions
# ============================================================================

# Run Strata analysis with dual output
run_analysis() {
  local plan_file="$1"
  local output_format="$2"
  local show_details="$3"
  local expand_all="$4"
  local config_file="$5"
  local json_file="$TEMP_DIR/metadata.json"

  # Build command
  local cmd="$TEMP_DIR/strata plan summary"
  cmd="$cmd --output $output_format"
  cmd="$cmd --file $json_file --file-format json"

  [[ "$show_details" == "true" ]] && cmd="$cmd --show-details"
  [[ "$expand_all" == "true" ]] && cmd="$cmd --expand-all"
  [[ -n "$config_file" ]] && cmd="$cmd --config $config_file"

  cmd="$cmd $plan_file"

  log_analyze "Analyzing Terraform plan"
  log_config "Running: $cmd"

  # Execute and capture display output
  local display_output
  if display_output=$($cmd 2>&1); then
    log_success "Analysis complete"

    # Store display output for later use
    echo "$display_output" > "$TEMP_DIR/display_output.txt"

    # Parse JSON for GitHub Action outputs
    if [[ -f "$json_file" ]]; then
      extract_outputs "$json_file" "$display_output"
    fi

    return 0
  else
    log_error "Analysis failed: $display_output" 4
  fi
}

# Extract and set GitHub Action outputs
extract_outputs() {
  local json_file="$1"
  local display_output="$2"

  if [[ ! -f "$json_file" ]]; then
    log_warning "JSON metadata file not found"
    return 1
  fi

  local json
  json=$(cat "$json_file")

  # Extract statistics using jq (pre-installed on GitHub runners)
  local total
  local dangers
  total=$(echo "$json" | jq -r '.statistics.total_changes // 0')
  dangers=$(echo "$json" | jq -r '.statistics.dangerous_changes // 0')

  # Set GitHub Action outputs
  if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    {
      echo "has-changes=$([[ $total -gt 0 ]] && echo true || echo false)"
      echo "has-dangers=$([[ $dangers -gt 0 ]] && echo true || echo false)"
      echo "change-count=$total"
      echo "danger-count=$dangers"
      echo "summary<<EOF"
      echo "$display_output"
      echo "EOF"
      echo "json-summary<<EOF"
      cat "$json_file"
      echo "EOF"
    } >> "$GITHUB_OUTPUT"
  fi

  # Write to GitHub Step Summary
  if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
    echo "$display_output" >> "$GITHUB_STEP_SUMMARY"
  fi

  return 0
}

# ============================================================================
# GitHub Integration Functions
# ============================================================================

# Post or update PR comment
update_pr_comment() {
  local content="$1"
  local comment_header="$2"
  local update_comment="$3"

  # Skip if not in PR context
  if [[ "${GITHUB_EVENT_NAME:-}" != "pull_request" ]] || [[ "${INPUT_COMMENT_ON_PR:-true}" != "true" ]]; then
    return 0
  fi

  # Get PR number
  local pr_number
  pr_number=$(jq -r .pull_request.number "${GITHUB_EVENT_PATH:-}")
  if [[ -z "$pr_number" ]] || [[ "$pr_number" == "null" ]]; then
    log_warning "Unable to determine PR number"
    return 0
  fi

  # Create unique marker for this workflow/job
  local marker="<!-- strata-${GITHUB_WORKFLOW:-workflow}-${GITHUB_JOB:-job} -->"
  local body="${marker}
${comment_header}

${content}"

  if [[ "$update_comment" == "true" ]]; then
    # Try to find and update existing comment
    log_info "Looking for existing comment to update"

    local comments
    comments=$(curl -s -H "Authorization: token ${GITHUB_TOKEN}" \
      "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments")

    local comment_id
    comment_id=$(echo "$comments" | \
      jq -r ".[] | select(.body | contains(\"$marker\")) | .id" | head -1)

    if [[ -n "$comment_id" ]] && [[ "$comment_id" != "null" ]]; then
      log_info "Updating existing comment #$comment_id"

      local response
      response=$(curl -s -w "\n%{http_code}" -X PATCH \
        -H "Authorization: token ${GITHUB_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"body\": $(echo "$body" | jq -R -s .)}" \
        "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}")

      local http_code="${response##*$'\n'}"

      if [[ "$http_code" -ge 200 ]] && [[ "$http_code" -lt 300 ]]; then
        log_success "Comment updated successfully"
        return 0
      else
        log_warning "Failed to update comment (HTTP $http_code), will create new"
      fi
    fi
  fi

  # Create new comment
  log_info "Creating new PR comment"

  local response
  response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"body\": $(echo "$body" | jq -R -s .)}" \
    "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments")

  local http_code="${response##*$'\n'}"

  if [[ "$http_code" -ge 200 ]] && [[ "$http_code" -lt 300 ]]; then
    log_success "Comment posted successfully"
  else
    log_error "Failed to post comment (HTTP $http_code)" 5
  fi
}

# ============================================================================
# Main Function
# ============================================================================

main() {
  log_start "Strata GitHub Action starting (simplified version)"

  # Process and validate inputs
  local plan_file="${INPUT_PLAN_FILE}"
  local output_format="${INPUT_OUTPUT_FORMAT:-markdown}"
  local show_details="${INPUT_SHOW_DETAILS:-false}"
  local expand_all="${INPUT_EXPAND_ALL:-false}"
  local config_file="${INPUT_CONFIG_FILE:-}"
  local strata_version="${INPUT_STRATA_VERSION:-latest}"
  local update_comment="${INPUT_UPDATE_COMMENT:-true}"
  local comment_header="${INPUT_COMMENT_HEADER:-🏗️ Terraform Plan Summary}"

  # Validate required inputs
  validate_required_inputs

  # Security validation
  validate_path_security "$plan_file" "Plan file path"

  # Validate plan file exists
  validate_file_exists "$plan_file" "Plan file"

  # Validate output format
  validate_output_format "$output_format"

  # Validate optional config file if provided
  if [[ -n "$config_file" ]]; then
    validate_path_security "$config_file" "Config file path"
    validate_file_exists "$config_file" "Config file"
  fi

  log_success "Input validation complete"

  # Download Strata binary
  download_strata "$strata_version"

  # Run analysis
  run_analysis "$plan_file" "$output_format" "$show_details" "$expand_all" "$config_file"

  # Update PR comment if applicable
  if [[ -f "$TEMP_DIR/display_output.txt" ]]; then
    local display_output
    display_output=$(cat "$TEMP_DIR/display_output.txt")
    update_pr_comment "$display_output" "$comment_header" "$update_comment"
  fi

  log_success "Strata GitHub Action completed successfully"
}

# ============================================================================
# Script Entry Point
# ============================================================================

# Only run main if this script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi