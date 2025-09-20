#!/bin/bash
set -euo pipefail

# Strata GitHub Action - Simplified Version
# Single file implementation with improved reliability and clearer logging

# Constants and configuration
readonly GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"
readonly TEMP_DIR=$(mktemp -d)
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Cleanup on exit (success or failure)
trap cleanup EXIT

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

  # Clean temp directory
  [[ -d "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"
}

# Platform detection
detect_platform() {
  case "$(uname -s)" in
    Linux*) OS="linux" ;;
    Darwin*) OS="darwin" ;;
    *) echo "❌ Unsupported OS: $(uname -s)"; exit 1 ;;
  esac

  case "$(uname -m)" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "❌ Unsupported architecture: $(uname -m)"; exit 1 ;;
  esac
}

# Binary download with checksum verification
download_strata() {
  local version="${INPUT_STRATA_VERSION:-latest}"

  detect_platform
  echo "🚀 Starting Strata GitHub Action (${OS}/${ARCH})"

  # Construct URLs based on version
  local base_url="https://github.com/ArjenSchwarz/strata/releases"
  local binary_url checksum_url

  if [[ "$version" == "latest" ]]; then
    binary_url="$base_url/latest/download/strata-${OS}-${ARCH}.tar.gz"
    checksum_url="$base_url/latest/download/checksums.txt"
    echo "📦 Using latest Strata release"
  else
    binary_url="$base_url/download/${version}/strata-${OS}-${ARCH}.tar.gz"
    checksum_url="$base_url/download/${version}/checksums.txt"
    echo "📦 Using Strata version: $version"
  fi

  echo "⬇️ Downloading Strata binary"

  # Download with retry
  for attempt in 1 2 3; do
    if curl -fsSL "$binary_url" -o "$TEMP_DIR/strata.tar.gz" 2>/dev/null; then
      # Download checksums
      if curl -fsSL "$checksum_url" -o "$TEMP_DIR/checksums.txt" 2>/dev/null; then
        # Verify checksum
        local expected=$(grep "strata-${OS}-${ARCH}.tar.gz" "$TEMP_DIR/checksums.txt" | cut -d' ' -f1)
        local actual=$(sha256sum "$TEMP_DIR/strata.tar.gz" | cut -d' ' -f1)

        if [[ "$expected" == "$actual" ]]; then
          # Extract verified binary
          tar -xz -C "$TEMP_DIR" -f "$TEMP_DIR/strata.tar.gz"
          chmod +x "$TEMP_DIR/strata"
          echo "✅ Download and verification successful"

          # Log version for debugging
          echo "🔍 Strata version: $("$TEMP_DIR/strata" --version 2>/dev/null || echo "unknown")"
          return 0
        else
          echo "⚠️ Checksum mismatch, retrying..."
        fi
      else
        echo "⚠️ Failed to download checksums, retrying..."
      fi
    fi
    echo "⚠️ Attempt $attempt/3 failed"
    [[ $attempt -lt 3 ]] && sleep 2
  done

  # Fallback to latest if specific version fails
  if [[ "$version" != "latest" ]]; then
    echo "⚠️ Version $version not found, trying latest"
    local fallback_url="$base_url/latest/download/strata-${OS}-${ARCH}.tar.gz"
    if curl -fsSL "$fallback_url" | tar -xz -C "$TEMP_DIR" 2>/dev/null; then
      chmod +x "$TEMP_DIR/strata"
      echo "✅ Fallback to latest successful"
      return 0
    fi
  fi

  echo "❌ Failed to download from: $binary_url"
  exit 3
}

# Input validation with essential security checks
validate_inputs() {
  echo "🔍 Validating inputs"

  # Required plan file validation
  if [[ -z "${INPUT_PLAN_FILE:-}" ]]; then
    echo "❌ Plan file is required"
    exit 2
  fi

  # Path traversal protection
  if [[ "$INPUT_PLAN_FILE" =~ \.\./ ]]; then
    echo "❌ Path traversal detected in plan file path"
    exit 2
  fi

  # Input length validation
  if [[ ${#INPUT_PLAN_FILE} -gt 4096 ]]; then
    echo "❌ Path too long (max 4096 characters)"
    exit 2
  fi

  # File existence and readability
  if [[ ! -f "$INPUT_PLAN_FILE" ]]; then
    echo "❌ Plan file not found: $INPUT_PLAN_FILE"
    exit 2
  fi

  if [[ ! -r "$INPUT_PLAN_FILE" ]]; then
    echo "❌ Plan file is not readable: $INPUT_PLAN_FILE"
    exit 2
  fi

  # Output format validation
  local format="${INPUT_OUTPUT_FORMAT:-markdown}"
  if [[ ! "$format" =~ ^(table|json|markdown|html)$ ]]; then
    echo "❌ Invalid output format: $format"
    exit 2
  fi

  # Config file validation (if provided)
  if [[ -n "${INPUT_CONFIG_FILE:-}" ]]; then
    if [[ "$INPUT_CONFIG_FILE" =~ \.\./ ]]; then
      echo "❌ Path traversal detected in config file path"
      exit 2
    fi
    if [[ ! -f "$INPUT_CONFIG_FILE" ]]; then
      echo "⚠️ Config file not found: $INPUT_CONFIG_FILE (using defaults)"
      INPUT_CONFIG_FILE=""
    elif [[ ! -r "$INPUT_CONFIG_FILE" ]]; then
      echo "⚠️ Config file not readable: $INPUT_CONFIG_FILE (using defaults)"
      INPUT_CONFIG_FILE=""
    fi
  fi

  echo "✅ Input validation completed"
}

# Execute Strata analysis with dual output
run_analysis() {
  local json_file="$TEMP_DIR/metadata.json"

  # Build command with all flags
  local cmd="$TEMP_DIR/strata plan summary"
  cmd="$cmd --output ${INPUT_OUTPUT_FORMAT:-markdown}"
  cmd="$cmd --file $json_file --file-format json"

  [[ "${INPUT_SHOW_DETAILS:-false}" == "true" ]] && cmd="$cmd --show-details"
  [[ "${INPUT_EXPAND_ALL:-false}" == "true" ]] && cmd="$cmd --expand-all"
  [[ -n "${INPUT_CONFIG_FILE:-}" ]] && cmd="$cmd --config $INPUT_CONFIG_FILE"

  cmd="$cmd $INPUT_PLAN_FILE"

  echo "🔍 Analyzing Terraform plan"
  echo "⚙️ Running: $cmd"

  # Execute and capture display output
  if display_output=$($cmd 2>&1); then
    echo "✅ Analysis complete"

    # Parse JSON for GitHub Action outputs
    if [[ -f "$json_file" ]]; then
      extract_outputs "$json_file"
    else
      echo "⚠️ JSON metadata file not found, setting default outputs"
      set_default_outputs
    fi

    # Store display output for GitHub integration
    DISPLAY_OUTPUT="$display_output"
  else
    echo "❌ Analysis failed: $display_output"

    # Set failure outputs
    {
      echo "has-changes=false"
      echo "has-dangers=false"
      echo "change-count=0"
      echo "danger-count=0"
      echo "summary=Analysis failed: $display_output"
      echo "json-summary={\"error\": \"Analysis failed\"}"
    } >> "$GITHUB_OUTPUT"

    exit 4
  fi
}

# Extract outputs from JSON metadata
extract_outputs() {
  local json_file="$1"

  if ! command -v jq >/dev/null 2>&1; then
    echo "⚠️ jq not available, setting default outputs"
    set_default_outputs
    return
  fi

  local json
  if ! json=$(cat "$json_file" 2>/dev/null); then
    echo "⚠️ Failed to read JSON file, setting default outputs"
    set_default_outputs
    return
  fi

  # Extract statistics safely
  local total_changes danger_changes
  total_changes=$(echo "$json" | jq -r '.statistics.total_changes // 0' 2>/dev/null || echo "0")
  danger_changes=$(echo "$json" | jq -r '.statistics.dangerous_changes // 0' 2>/dev/null || echo "0")

  # Set GitHub Action outputs
  {
    echo "has-changes=$([[ $total_changes -gt 0 ]] && echo true || echo false)"
    echo "has-dangers=$([[ $danger_changes -gt 0 ]] && echo true || echo false)"
    echo "change-count=$total_changes"
    echo "danger-count=$danger_changes"
    echo "summary<<EOF"
    echo "$DISPLAY_OUTPUT"
    echo "EOF"
    echo "json-summary<<EOF"
    cat "$json_file"
    echo "EOF"
  } >> "$GITHUB_OUTPUT"
}

# Set default outputs when JSON parsing fails
set_default_outputs() {
  {
    echo "has-changes=false"
    echo "has-dangers=false"
    echo "change-count=0"
    echo "danger-count=0"
    echo "summary<<EOF"
    echo "${DISPLAY_OUTPUT:-No output available}"
    echo "EOF"
    echo "json-summary={}"
  } >> "$GITHUB_OUTPUT"
}

# GitHub Step Summary
write_step_summary() {
  if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
    echo "$DISPLAY_OUTPUT" >> "$GITHUB_STEP_SUMMARY"
    echo "📝 Step summary updated"
  fi
}

# PR comment functionality
update_pr_comment() {
  # Skip if not in PR context
  if [[ "${GITHUB_EVENT_NAME:-}" != "pull_request" ]] || [[ "${INPUT_COMMENT_ON_PR:-true}" != "true" ]]; then
    return 0
  fi

  local pr_number
  if ! pr_number=$(jq -r '.pull_request.number // empty' "$GITHUB_EVENT_PATH" 2>/dev/null); then
    echo "⚠️ Could not determine PR number"
    return 0
  fi

  if [[ -z "$pr_number" ]] || [[ "$pr_number" == "null" ]]; then
    echo "⚠️ No PR number found"
    return 0
  fi

  local marker="<!-- strata-${GITHUB_WORKFLOW:-default}-${GITHUB_JOB:-default} -->"
  local header="${INPUT_COMMENT_HEADER:-🏗️ Terraform Plan Summary}"
  local body="${marker}
${header}

${DISPLAY_OUTPUT}"

  echo "📝 Processing PR comment for PR #$pr_number"

  if [[ "${INPUT_UPDATE_COMMENT:-true}" == "true" ]]; then
    # Try to update existing comment
    local comments
    if comments=$(curl -s -H "Authorization: token ${GITHUB_TOKEN:-}" \
      "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/$pr_number/comments" 2>/dev/null); then

      local comment_id
      if comment_id=$(echo "$comments" | jq -r ".[] | select(.body | contains(\"$marker\")) | .id" 2>/dev/null | head -1); then
        if [[ -n "$comment_id" ]] && [[ "$comment_id" != "null" ]]; then
          echo "📝 Updating existing comment #$comment_id"
          if update_comment "$comment_id" "$body"; then
            return 0
          fi
        fi
      fi
    fi
  fi

  # Create new comment
  create_comment "$pr_number" "$body"
}

# Update existing comment
update_comment() {
  local comment_id="$1"
  local body="$2"

  local response http_code
  response=$(curl -s -w "\n%{http_code}" -X PATCH \
    -H "Authorization: token ${GITHUB_TOKEN:-}" \
    -H "Content-Type: application/json" \
    -d "$(echo "$body" | jq -R -s '{"body": .}')" \
    "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/comments/$comment_id" 2>/dev/null)

  http_code="${response##*$'\n'}"

  if [[ "$http_code" -ge 200 ]] && [[ "$http_code" -lt 300 ]]; then
    echo "✅ Comment updated successfully"
    return 0
  else
    echo "⚠️ Update failed (HTTP $http_code), will create new comment"
    return 1
  fi
}

# Create new comment
create_comment() {
  local pr_number="$1"
  local body="$2"

  echo "📝 Creating new PR comment"

  local response http_code
  response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: token ${GITHUB_TOKEN:-}" \
    -H "Content-Type: application/json" \
    -d "$(echo "$body" | jq -R -s '{"body": .}')" \
    "$GITHUB_API_URL/repos/$GITHUB_REPOSITORY/issues/$pr_number/comments" 2>/dev/null)

  http_code="${response##*$'\n'}"

  if [[ "$http_code" -ge 200 ]] && [[ "$http_code" -lt 300 ]]; then
    echo "✅ Comment posted successfully"
  else
    echo "❌ Failed to post comment (HTTP $http_code)"
  fi
}

# Main execution flow
main() {
  # Step 1: Validate inputs
  validate_inputs

  # Step 2: Download and setup Strata binary
  download_strata

  # Step 3: Run analysis
  run_analysis

  # Step 4: Write GitHub Step Summary
  write_step_summary

  # Step 5: Handle PR comments
  update_pr_comment

  echo "🎉 Strata GitHub Action completed successfully"
}

# Run main function
main "$@"