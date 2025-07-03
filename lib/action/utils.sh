#!/bin/bash

# Core utility functions for Strata GitHub Action
# This module provides logging, error handling, and basic utility functions

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