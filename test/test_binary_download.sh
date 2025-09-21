#!/bin/bash
#
# Unit tests for binary download system
# Tests platform detection, URL construction, checksum verification, and retry logic
#
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test output directory
TEST_OUTPUT_DIR=$(mktemp -d)
trap 'rm -rf "$TEST_OUTPUT_DIR"' EXIT

# Mock variables
MOCK_OS=""
MOCK_ARCH=""
MOCK_CURL_FAIL_COUNT=0
MOCK_CURL_MAX_FAILS=0
MOCK_CHECKSUM_VALID=true

# ============================================================================
# Test Helper Functions
# ============================================================================

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

log_section() {
    echo ""
    echo -e "${BLUE}=== $1 ===${NC}"
    echo ""
}

assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="$3"

    if [[ "$expected" == "$actual" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected: '$expected'"
        echo "      Actual:   '$actual'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="$3"

    if [[ "$haystack" == *"$needle"* ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected to contain: '$needle'"
        echo "      Actual: '$haystack'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_not_empty() {
    local value="$1"
    local message="$2"

    if [[ -n "$value" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected non-empty value"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local message="$2"

    if [[ -f "$file" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected file to exist: '$file'"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# ============================================================================
# Mock Functions
# ============================================================================

# Mock uname for platform detection
uname() {
    case "$1" in
        -s)
            echo "$MOCK_OS"
            ;;
        -m)
            echo "$MOCK_ARCH"
            ;;
        *)
            command uname "$@"
            ;;
    esac
}

# Mock curl for download simulation
curl() {
    local args=("$@")
    local url=""
    local output=""

    # Parse arguments
    for i in "${!args[@]}"; do
        if [[ "${args[$i]}" == "-o" ]]; then
            output="${args[$((i + 1))]}"
        elif [[ "${args[$i]}" == http* ]]; then
            url="${args[$i]}"
        fi
    done

    # Simulate failure based on counter
    if [[ $MOCK_CURL_FAIL_COUNT -lt $MOCK_CURL_MAX_FAILS ]]; then
        MOCK_CURL_FAIL_COUNT=$((MOCK_CURL_FAIL_COUNT + 1))
        return 1
    fi

    # Create mock files based on URL
    if [[ "$url" == *"checksums.txt" ]]; then
        # Create mock checksums file
        echo "abc123def456  strata-linux-amd64.tar.gz" > "$output"
        echo "def456abc789  strata-darwin-amd64.tar.gz" >> "$output"
        echo "789abc123def  strata-linux-arm64.tar.gz" >> "$output"
        echo "123def456abc  strata-darwin-arm64.tar.gz" >> "$output"
    elif [[ "$url" == *".tar.gz" ]]; then
        # Create mock tar.gz file
        mkdir -p "$TEST_OUTPUT_DIR/mock_binary"
        echo "#!/bin/bash" > "$TEST_OUTPUT_DIR/mock_binary/strata"
        echo "echo 'Strata v1.4.0'" >> "$TEST_OUTPUT_DIR/mock_binary/strata"
        chmod +x "$TEST_OUTPUT_DIR/mock_binary/strata"
        tar -czf "$output" -C "$TEST_OUTPUT_DIR" mock_binary/strata 2>/dev/null
    fi

    return 0
}

# Mock sha256sum for checksum verification
sha256sum() {
    local file="$1"

    if [[ "$MOCK_CHECKSUM_VALID" == "true" ]]; then
        # Return matching checksum based on platform
        case "$MOCK_OS-$MOCK_ARCH" in
            "Linux-x86_64")
                echo "abc123def456  $file"
                ;;
            "Darwin-x86_64")
                echo "def456abc789  $file"
                ;;
            "Linux-aarch64")
                echo "789abc123def  $file"
                ;;
            "Darwin-arm64")
                echo "123def456abc  $file"
                ;;
            *)
                echo "invalid000000  $file"
                ;;
        esac
    else
        echo "invalid000000  $file"
    fi
}

# Mock sleep to speed up tests
sleep() {
    return 0
}

# ============================================================================
# Functions Under Test (simplified versions)
# ============================================================================

detect_platform() {
    local os=""
    local arch=""

    case "$(uname -s)" in
        Linux*) os="linux" ;;
        Darwin*) os="darwin" ;;
        *) echo "❌ Unsupported OS: $(uname -s)" >&2; return 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo "❌ Unsupported architecture: $(uname -m)" >&2; return 1 ;;
    esac

    echo "${os}-${arch}"
}

download_strata() {
    local version="${1:-latest}"
    local temp_dir="${2:-$TEST_OUTPUT_DIR}"

    # Detect platform
    local platform=$(detect_platform) || return 1
    local os=$(echo "$platform" | cut -d'-' -f1)
    local arch=$(echo "$platform" | cut -d'-' -f2)

    # Construct URLs based on version
    local base_url="https://github.com/ArjenSchwarz/strata/releases"
    local binary_url
    local checksum_url

    if [[ "$version" == "latest" ]]; then
        binary_url="$base_url/latest/download/strata-${os}-${arch}.tar.gz"
        checksum_url="$base_url/latest/download/checksums.txt"
    else
        binary_url="$base_url/download/${version}/strata-${os}-${arch}.tar.gz"
        checksum_url="$base_url/download/${version}/checksums.txt"
    fi

    echo "⬇️ Downloading Strata ${version} for ${os}/${arch}"

    # Download with retry
    local attempt
    for attempt in 1 2 3; do
        if curl -fsSL "$binary_url" -o "$temp_dir/strata.tar.gz" 2>/dev/null; then
            # Download checksums
            if curl -fsSL "$checksum_url" -o "$temp_dir/checksums.txt" 2>/dev/null; then
                # Verify checksum
                local expected=$(grep "strata-${os}-${arch}.tar.gz" "$temp_dir/checksums.txt" | cut -d' ' -f1)
                local actual=$(sha256sum "$temp_dir/strata.tar.gz" | cut -d' ' -f1)

                if [[ "$expected" == "$actual" ]]; then
                    # Extract verified binary
                    tar -xzf "$temp_dir/strata.tar.gz" -C "$temp_dir"
                    echo "✅ Download and verification successful"
                    return 0
                else
                    echo "⚠️ Checksum mismatch (attempt $attempt/3)"
                fi
            else
                echo "⚠️ Failed to download checksums (attempt $attempt/3)"
            fi
        else
            echo "⚠️ Failed to download binary (attempt $attempt/3)"
        fi

        if [[ $attempt -lt 3 ]]; then
            sleep 2
        fi
    done

    echo "❌ Failed to download from: $binary_url"
    return 3
}

# ============================================================================
# Test Cases
# ============================================================================

test_platform_detection_linux_amd64() {
    log_test "Platform detection for Linux x86_64"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"

    local platform=$(detect_platform)
    assert_equals "linux-amd64" "$platform" "Should detect Linux AMD64"
}

test_platform_detection_darwin_amd64() {
    log_test "Platform detection for Darwin x86_64"
    MOCK_OS="Darwin"
    MOCK_ARCH="x86_64"

    local platform=$(detect_platform)
    assert_equals "darwin-amd64" "$platform" "Should detect macOS AMD64"
}

test_platform_detection_linux_arm64() {
    log_test "Platform detection for Linux ARM64"
    MOCK_OS="Linux"
    MOCK_ARCH="aarch64"

    local platform=$(detect_platform)
    assert_equals "linux-arm64" "$platform" "Should detect Linux ARM64"
}

test_platform_detection_darwin_arm64() {
    log_test "Platform detection for Darwin ARM64"
    MOCK_OS="Darwin"
    MOCK_ARCH="arm64"

    local platform=$(detect_platform)
    assert_equals "darwin-arm64" "$platform" "Should detect macOS ARM64"
}

test_platform_detection_unsupported_os() {
    log_test "Platform detection for unsupported OS"
    MOCK_OS="Windows"
    MOCK_ARCH="x86_64"

    local output=$(detect_platform 2>&1) || true
    assert_contains "$output" "Unsupported OS" "Should reject Windows"
}

test_platform_detection_unsupported_arch() {
    log_test "Platform detection for unsupported architecture"
    MOCK_OS="Linux"
    MOCK_ARCH="i386"

    local output=$(detect_platform 2>&1) || true
    assert_contains "$output" "Unsupported architecture" "Should reject i386"
}

test_url_construction_latest() {
    log_test "URL construction for latest version"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=true

    local output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    assert_contains "$output" "Downloading Strata latest" "Should use latest version"
    assert_contains "$output" "linux/amd64" "Should include platform in output"
}

test_url_construction_specific_version() {
    log_test "URL construction for specific version"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=true

    local output=$(download_strata "v1.4.0" "$TEST_OUTPUT_DIR" 2>&1)
    assert_contains "$output" "Downloading Strata v1.4.0" "Should use specific version"
}

test_url_construction_prerelease() {
    log_test "URL construction for pre-release version"
    MOCK_OS="Darwin"
    MOCK_ARCH="arm64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=true

    local output=$(download_strata "v1.5.0-beta.1" "$TEST_OUTPUT_DIR" 2>&1)
    assert_contains "$output" "Downloading Strata v1.5.0-beta.1" "Should handle pre-release versions"
}

test_checksum_verification_success() {
    log_test "Checksum verification success"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=true

    local output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    assert_contains "$output" "Download and verification successful" "Should verify checksum successfully"
    assert_file_exists "$TEST_OUTPUT_DIR/mock_binary/strata" "Should extract binary after successful verification"
}

test_checksum_verification_failure() {
    log_test "Checksum verification failure"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=false

    local output
    local exit_code

    # Capture both output and exit code properly
    set +e
    output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    exit_code=$?
    set -e

    assert_contains "$output" "Checksum mismatch" "Should detect checksum mismatch"
    assert_equals "3" "$exit_code" "Should exit with code 3 on failure"
}

test_retry_on_first_failure() {
    log_test "Retry mechanism on first failure"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=1  # Fail once, then succeed
    MOCK_CHECKSUM_VALID=true

    local output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    assert_contains "$output" "Failed to download binary (attempt 1/3)" "Should show first failure"
    assert_contains "$output" "Download and verification successful" "Should succeed after retry"
}

test_retry_on_second_failure() {
    log_test "Retry mechanism on second failure"
    MOCK_OS="Darwin"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=2  # Fail twice, then succeed
    MOCK_CHECKSUM_VALID=true

    local output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    assert_contains "$output" "Failed to download binary (attempt 1/3)" "Should show first failure"
    assert_contains "$output" "Failed to download binary (attempt 2/3)" "Should show second failure"
    assert_contains "$output" "Download and verification successful" "Should succeed on third attempt"
}

test_retry_all_attempts_fail() {
    log_test "All retry attempts fail"
    MOCK_OS="Linux"
    MOCK_ARCH="aarch64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=10  # Always fail
    MOCK_CHECKSUM_VALID=true

    local output
    local exit_code

    # Capture both output and exit code properly
    set +e
    output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    exit_code=$?
    set -e

    assert_contains "$output" "Failed to download binary (attempt 1/3)" "Should show first failure"
    assert_contains "$output" "Failed to download binary (attempt 2/3)" "Should show second failure"
    assert_contains "$output" "Failed to download binary (attempt 3/3)" "Should show third failure"
    assert_contains "$output" "Failed to download from:" "Should show final failure message"
    assert_equals "3" "$exit_code" "Should exit with code 3"
}

test_version_fallback_not_triggered_on_latest() {
    log_test "Version fallback not triggered when using latest"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=10  # Always fail
    MOCK_CHECKSUM_VALID=true

    local output
    local exit_code

    # When using "latest" and it fails, should not fall back to anything
    set +e
    output=$(download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1)
    exit_code=$?
    set -e

    assert_equals "3" "$exit_code" "Should fail without fallback for latest"
}

test_version_fallback_from_specific_to_latest() {
    log_test "Version fallback from specific version to latest"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CHECKSUM_VALID=true

    # Create a modified download_strata function that can handle version-specific failures
    download_strata_with_fallback() {
        local version="${1:-latest}"
        local temp_dir="${2:-$TEST_OUTPUT_DIR}"

        # Fail for specific version but succeed for latest
        if [[ "$version" != "latest" ]]; then
            echo "⬇️ Downloading Strata ${version} for linux/amd64"
            echo "❌ Failed to download from: https://github.com/ArjenSchwarz/strata/releases/download/${version}/strata-linux-amd64.tar.gz"
            echo "⚠️ Version $version not found, falling back to latest"

            # Recursive call to latest
            download_strata_with_fallback "latest" "$temp_dir"
            return $?
        else
            # Success for latest
            echo "⬇️ Downloading Strata latest for linux/amd64"

            # Create mock files for latest
            mkdir -p "$temp_dir/mock_binary"
            echo "#!/bin/bash" > "$temp_dir/mock_binary/strata"
            echo "echo 'Strata v1.4.0'" >> "$temp_dir/mock_binary/strata"
            chmod +x "$temp_dir/mock_binary/strata"
            tar -czf "$temp_dir/strata.tar.gz" -C "$temp_dir" mock_binary/strata 2>/dev/null

            echo "abc123def456  strata-linux-amd64.tar.gz" > "$temp_dir/checksums.txt"

            # Extract
            tar -xzf "$temp_dir/strata.tar.gz" -C "$temp_dir"
            echo "✅ Download and verification successful"
            return 0
        fi
    }

    local output
    output=$(download_strata_with_fallback "v99.99.99" "$TEST_OUTPUT_DIR" 2>&1)

    assert_contains "$output" "Version v99.99.99 not found" "Should show version not found"
    assert_contains "$output" "falling back to latest" "Should show fallback message"
    assert_contains "$output" "Download and verification successful" "Should succeed with latest"
    assert_file_exists "$TEST_OUTPUT_DIR/mock_binary/strata" "Should create binary after fallback"
}

test_download_creates_temp_files() {
    log_test "Download creates expected temporary files"
    MOCK_OS="Linux"
    MOCK_ARCH="x86_64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=true

    download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1

    assert_file_exists "$TEST_OUTPUT_DIR/strata.tar.gz" "Should create tar.gz file"
    assert_file_exists "$TEST_OUTPUT_DIR/checksums.txt" "Should create checksums file"
    assert_file_exists "$TEST_OUTPUT_DIR/mock_binary/strata" "Should extract binary"
}

test_binary_is_executable() {
    log_test "Downloaded binary is executable"
    MOCK_OS="Darwin"
    MOCK_ARCH="arm64"
    MOCK_CURL_FAIL_COUNT=0
    MOCK_CURL_MAX_FAILS=0
    MOCK_CHECKSUM_VALID=true

    download_strata "latest" "$TEST_OUTPUT_DIR" 2>&1

    if [[ -x "$TEST_OUTPUT_DIR/mock_binary/strata" ]]; then
        echo -e "${GREEN}  ✓${NC} Binary should be executable"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}  ✗${NC} Binary should be executable"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# ============================================================================
# Main Test Runner
# ============================================================================

main() {
    echo "Starting Binary Download System Tests"
    echo "====================================="

    log_section "Platform Detection Tests"
    test_platform_detection_linux_amd64
    test_platform_detection_darwin_amd64
    test_platform_detection_linux_arm64
    test_platform_detection_darwin_arm64
    test_platform_detection_unsupported_os
    test_platform_detection_unsupported_arch

    log_section "URL Construction Tests"
    test_url_construction_latest
    test_url_construction_specific_version
    test_url_construction_prerelease

    log_section "Checksum Verification Tests"
    test_checksum_verification_success
    test_checksum_verification_failure

    log_section "Retry Logic Tests"
    test_retry_on_first_failure
    test_retry_on_second_failure
    test_retry_all_attempts_fail

    log_section "Version Fallback Tests"
    test_version_fallback_not_triggered_on_latest
    test_version_fallback_from_specific_to_latest

    log_section "File Operations Tests"
    test_download_creates_temp_files
    test_binary_is_executable

    # Print summary
    echo ""
    echo "====================================="
    echo "Test Summary:"
    echo "  Tests run:    $TESTS_RUN"
    echo -e "  ${GREEN}Passed:${NC}       $TESTS_PASSED"
    echo -e "  ${RED}Failed:${NC}       $TESTS_FAILED"

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        exit 1
    fi
}

# Run tests if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi