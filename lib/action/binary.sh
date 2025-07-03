#!/bin/bash

# Binary management functions for Strata GitHub Action
# This module handles binary downloading, caching, and compilation

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
  cd "$SRC_DIR" || return 1
  
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

# Function to download and setup binary
setup_strata_binary() {
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

  # Setup cache directories
  CACHE_DIR="$HOME/.cache/strata"
  mkdir -p "$CACHE_DIR"

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
}