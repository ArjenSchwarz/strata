#!/bin/bash
#
# Tests for GitHub integration features
# Tests Step Summary, PR comment functionality, environment detection
#
# Requirements tested:
# 8.1: Write Step Summary with display output
# 8.2: Check if in PR context using GITHUB_EVENT_NAME
# 8.3: Extract PR number from GITHUB_EVENT_PATH
# 8.4: Generate unique marker with workflow and job names
# 8.5: Search for existing comment with marker using GET API
# 8.6: Update existing comment with PATCH or create new with POST
# 8.7: Handle PR detection using standard GITHUB_EVENT_NAME checks
# 8.8: Gracefully skip PR features when not in PR context
# 8.10: Include environment/job information in the comment marker
# 8.11: Fall back to creating new comment if update fails

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

assert_file_exists() {
    local file="$1"
    local message="$2"

    if [[ -f "$file" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected file to exist: $file"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

assert_exit_code() {
    local expected_code="$1"
    local actual_code="$2"
    local message="$3"

    if [[ "$expected_code" -eq "$actual_code" ]]; then
        echo -e "${GREEN}  ✓${NC} $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}  ✗${NC} $message"
        echo "      Expected exit code: $expected_code"
        echo "      Actual exit code:   $actual_code"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Source the script to test (without executing main)
source_script() {
    if [[ -z "${SCRIPT_SOURCED:-}" ]]; then
        # Set up test TEMP_DIR if not already set
        if [[ -z "${TEST_TEMP_DIR:-}" ]]; then
            export TEST_TEMP_DIR="$TEST_OUTPUT_DIR/test_temp_$(date +%s)"
            mkdir -p "$TEST_TEMP_DIR"
        fi

        # Create a temporary copy that doesn't execute main or set readonly TEMP_DIR
        local temp_script="$TEST_OUTPUT_DIR/action_temp.sh"

        # Copy everything except the main execution guard and readonly TEMP_DIR
        sed -e '/^if \[\[ "${BASH_SOURCE\[0\]}" == "${0}" \]\]; then$/,/^fi$/d' \
            -e 's/^readonly TEMP_DIR$/# readonly TEMP_DIR/' \
            -e '/^TEMP_DIR=$(mktemp -d)$/d' \
            action_simplified.sh > "$temp_script"

        # Add a test-safe TEMP_DIR assignment
        echo "TEMP_DIR=\${TEST_TEMP_DIR:-\$(mktemp -d)}" >> "$temp_script"

        # Source the modified script
        source "$temp_script"
        export SCRIPT_SOURCED=1
    fi
}

# Create mock GitHub event file for PR context
create_mock_pr_event() {
    local pr_number="${1:-123}"
    local event_file="$TEST_OUTPUT_DIR/github_event.json"

    cat > "$event_file" <<EOF
{
  "number": $pr_number,
  "pull_request": {
    "number": $pr_number,
    "id": 456789123,
    "title": "Test PR",
    "state": "open"
  },
  "repository": {
    "name": "test-repo",
    "full_name": "owner/test-repo"
  }
}
EOF

    echo "$event_file"
}

# Mock curl function for API testing
mock_curl() {
    local args=("$@")
    local method="GET"
    local url=""
    local data=""
    local response_file=""

    # Parse curl arguments
    for ((i=0; i<${#args[@]}; i++)); do
        case "${args[i]}" in
            -X)
                method="${args[i+1]}"
                i=$((i+1))
                ;;
            -d)
                data="${args[i+1]}"
                i=$((i+1))
                ;;
            */comments)
                url="comments"
                ;;
            */comments/*)
                url="comment_update"
                ;;
        esac
    done

    # Generate mock responses based on request
    case "$method-$url" in
        "GET-comments")
            # Mock existing comments response
            if [[ -f "$TEST_OUTPUT_DIR/mock_comments.json" ]]; then
                cat "$TEST_OUTPUT_DIR/mock_comments.json"
            else
                echo "[]"
            fi
            echo -e "\n200"
            ;;
        "POST-comments")
            # Mock comment creation success
            echo '{"id": 987654321}'
            echo -e "\n201"
            ;;
        "PATCH-comment_update")
            # Mock comment update success
            echo '{"id": 123456789}'
            echo -e "\n200"
            ;;
        *)
            # Default failure
            echo '{"error": "Not found"}'
            echo -e "\n404"
            ;;
    esac
}

# ============================================================================
# Test Cases: Step Summary Generation
# ============================================================================

test_step_summary_generation() {
    log_section "Step Summary Generation"

    log_test "Test Step Summary writing to GITHUB_STEP_SUMMARY (Requirement 8.1)"

    source_script

    # Set up test environment
    local step_summary_file="$TEST_OUTPUT_DIR/step_summary.txt"
    export GITHUB_STEP_SUMMARY="$step_summary_file"

    # Create test display output
    local test_output="Test Plan Summary
No changes detected"

    # Call extract_outputs function with test data
    local json_file="$TEST_OUTPUT_DIR/test_metadata.json"
    cat > "$json_file" <<EOF
{
  "statistics": {
    "total_changes": 0,
    "dangerous_changes": 0
  }
}
EOF

    # Test extract_outputs function
    extract_outputs "$json_file" "$test_output"

    # Verify Step Summary was written
    assert_file_exists "$step_summary_file" "Step Summary file should be created"

    if [[ -f "$step_summary_file" ]]; then
        local summary_content=$(cat "$step_summary_file")
        assert_contains "$summary_content" "Test Plan Summary" "Step Summary should contain display output"
        assert_contains "$summary_content" "No changes detected" "Step Summary should contain complete output"
    fi

    unset GITHUB_STEP_SUMMARY
}

test_step_summary_multiline_content() {
    log_test "Test Step Summary handling of multi-line content (Requirement 5.6)"

    source_script

    local step_summary_file="$TEST_OUTPUT_DIR/step_summary_multiline.txt"
    export GITHUB_STEP_SUMMARY="$step_summary_file"

    # Create multi-line test output with special characters
    local multiline_output="# Terraform Plan Summary

## Changes
- Resource A will be created
- Resource B will be updated

## Warnings
⚠️ Some resources will be replaced

**Total**: 2 changes"

    local json_file="$TEST_OUTPUT_DIR/test_metadata.json"
    cat > "$json_file" <<EOF
{
  "statistics": {
    "total_changes": 2,
    "dangerous_changes": 1
  }
}
EOF

    extract_outputs "$json_file" "$multiline_output"

    if [[ -f "$step_summary_file" ]]; then
        local summary_content=$(cat "$step_summary_file")
        assert_contains "$summary_content" "# Terraform Plan Summary" "Should preserve markdown formatting"
        assert_contains "$summary_content" "## Changes" "Should preserve multi-line structure"
        assert_contains "$summary_content" "⚠️ Some resources" "Should preserve emoji characters"
        assert_contains "$summary_content" "**Total**: 2 changes" "Should preserve markdown bold"
    fi

    unset GITHUB_STEP_SUMMARY
}

# ============================================================================
# Test Cases: PR Context Detection
# ============================================================================

test_pr_context_detection() {
    log_section "PR Context Detection"

    log_test "Test PR context detection via GITHUB_EVENT_NAME (Requirement 8.2, 8.7)"

    source_script

    # Test PR context detection
    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"

    # Create PR event file
    local event_file=$(create_mock_pr_event 123)
    export GITHUB_EVENT_PATH="$event_file"
    export GITHUB_REPOSITORY="owner/test-repo"
    export GITHUB_WORKFLOW="test-workflow"
    export GITHUB_JOB="test-job"
    export GITHUB_TOKEN="fake-token"

    # Override curl to use mock
    curl() { mock_curl "$@"; }

    # Test update_pr_comment function in PR context
    local test_content="Test PR comment content"
    local comment_header="Test Header"

    # This should attempt to make API calls
    local exit_code=0
    update_pr_comment "$test_content" "$comment_header" "false" || exit_code=$?

    assert_equals "0" "$exit_code" "Should succeed in PR context"

    # Clean up exports
    unset GITHUB_EVENT_NAME GITHUB_EVENT_PATH GITHUB_REPOSITORY
    unset GITHUB_WORKFLOW GITHUB_JOB GITHUB_TOKEN INPUT_COMMENT_ON_PR
}

test_non_pr_context_skip() {
    log_test "Test graceful skip when not in PR context (Requirement 8.8)"

    source_script

    # Test non-PR context (push event)
    export GITHUB_EVENT_NAME="push"
    export INPUT_COMMENT_ON_PR="true"

    # Test update_pr_comment function in non-PR context
    local test_content="Test content"
    local comment_header="Test Header"

    local exit_code=0
    update_pr_comment "$test_content" "$comment_header" "false" || exit_code=$?

    assert_equals "0" "$exit_code" "Should skip gracefully in non-PR context"

    # Test when comment_on_pr is disabled
    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="false"

    exit_code=0
    update_pr_comment "$test_content" "$comment_header" "false" || exit_code=$?

    assert_equals "0" "$exit_code" "Should skip when comment_on_pr is disabled"

    unset GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR
}

test_pr_number_extraction() {
    log_test "Test PR number extraction from GITHUB_EVENT_PATH (Requirement 8.3)"

    source_script

    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"

    # Test with valid PR event
    local event_file=$(create_mock_pr_event 456)
    export GITHUB_EVENT_PATH="$event_file"
    export GITHUB_REPOSITORY="owner/test-repo"
    export GITHUB_WORKFLOW="test-workflow"
    export GITHUB_JOB="test-job"
    export GITHUB_TOKEN="fake-token"

    # Override curl to capture the API call
    curl() {
        # Check if the URL contains the correct PR number
        if [[ "$*" == *"/issues/456/comments"* ]]; then
            echo "[]"
            echo -e "\n200"
        else
            echo '{"error": "Wrong PR number"}'
            echo -e "\n400"
        fi
    }

    local test_content="Test content"
    local comment_header="Test Header"

    local exit_code=0
    update_pr_comment "$test_content" "$comment_header" "false" || exit_code=$?

    assert_equals "0" "$exit_code" "Should extract correct PR number from event"

    # Test with invalid/missing event path
    export GITHUB_EVENT_PATH="/nonexistent/file"

    exit_code=0
    update_pr_comment "$test_content" "$comment_header" "false" || exit_code=$?

    # Should handle gracefully (return 0 but skip processing)
    assert_equals "0" "$exit_code" "Should handle missing event file gracefully"

    unset GITHUB_EVENT_PATH GITHUB_REPOSITORY GITHUB_WORKFLOW GITHUB_JOB GITHUB_TOKEN GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR
}

# ============================================================================
# Test Cases: Comment Marker Generation
# ============================================================================

test_comment_marker_generation() {
    log_section "Comment Marker Generation"

    log_test "Test unique marker generation with workflow and job names (Requirement 8.4, 8.10)"

    source_script

    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"
    export GITHUB_WORKFLOW="ci-workflow"
    export GITHUB_JOB="test-job"
    export GITHUB_TOKEN="fake-token"
    export GITHUB_REPOSITORY="owner/repo"

    # Create PR event file
    local event_file=$(create_mock_pr_event 789)
    export GITHUB_EVENT_PATH="$event_file"

    # Write the output to a file to check the marker
    local output_file="$TEST_OUTPUT_DIR/comment_body.txt"

    # Mock curl to capture the JSON payload
    curl() {
        local json_data=""
        local parsing_data=false

        # Parse arguments more carefully
        for arg in "$@"; do
            if [[ "$parsing_data" == true ]]; then
                json_data="$arg"
                break
            elif [[ "$arg" == "-d" ]]; then
                parsing_data=true
            fi
        done

        # Save the JSON data for inspection
        echo "$json_data" > "$output_file"

        echo "[]"
        echo -e "\n200"
    }

    local test_content="Test content"
    local comment_header="Test Header"

    update_pr_comment "$test_content" "$comment_header" "false"

    # Check if the output file contains the expected marker
    local found_marker=""
    if [[ -f "$output_file" ]]; then
        local comment_data
        comment_data=$(cat "$output_file")
        if [[ "$comment_data" == *"<!-- strata-ci-workflow-test-job -->"* ]]; then
            found_marker="strata-ci-workflow-test-job"
        fi
    fi

    assert_equals "strata-ci-workflow-test-job" "$found_marker" "Should generate marker with workflow and job names"

    # Test with missing workflow/job (should use defaults)
    unset GITHUB_WORKFLOW GITHUB_JOB

    rm -f "$output_file"

    update_pr_comment "$test_content" "$comment_header" "false"

    # Check for default marker
    found_marker=""
    if [[ -f "$output_file" ]]; then
        local comment_data=$(cat "$output_file")
        if [[ "$comment_data" == *"<!-- strata-workflow-job -->"* ]]; then
            found_marker="strata-workflow-job"
        fi
    fi

    assert_equals "strata-workflow-job" "$found_marker" "Should use default values when workflow/job not set"

    unset GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR GITHUB_TOKEN GITHUB_REPOSITORY GITHUB_EVENT_PATH
}

# ============================================================================
# Test Cases: Comment Creation and Updates
# ============================================================================

test_comment_creation() {
    log_section "Comment Creation"

    log_test "Test new comment creation with POST API (Requirement 8.6)"

    source_script

    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"
    export GITHUB_WORKFLOW="test-workflow"
    export GITHUB_JOB="test-job"
    export GITHUB_TOKEN="fake-token"
    export GITHUB_REPOSITORY="owner/repo"

    local event_file=$(create_mock_pr_event 101)
    export GITHUB_EVENT_PATH="$event_file"

    local post_called=false
    local post_data=""

    curl() {
        local method=""
        local data=""

        # Parse arguments
        for ((i=0; i<$#; i++)); do
            if [[ "${!i}" == "-X" ]]; then
                method="${@:$((i+2)):1}"
            elif [[ "${!i}" == "-d" ]]; then
                data="${@:$((i+2)):1}"
            fi
        done

        if [[ "$method" == "POST" ]] && [[ "$*" == *"/comments" ]]; then
            post_called=true
            post_data="$data"
            echo '{"id": 123456}'
            echo -e "\n201"
        else
            echo "[]"
            echo -e "\n200"
        fi
    }

    local test_content="New comment content"
    local comment_header="🏗️ Test Header"

    update_pr_comment "$test_content" "$comment_header" "false"

    if [[ "$post_called" == true ]]; then
        echo -e "${GREEN}  ✓${NC} POST API was called for comment creation"
        TESTS_PASSED=$((TESTS_PASSED + 1))

        # Check that the data contains our content
        if [[ "$post_data" == *"New comment content"* ]] && [[ "$post_data" == *"Test Header"* ]]; then
            echo -e "${GREEN}  ✓${NC} Comment data contains expected content and header"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}  ✗${NC} Comment data missing expected content"
            echo "      Data: $post_data"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        echo -e "${RED}  ✗${NC} POST API was not called"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    unset GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR GITHUB_WORKFLOW GITHUB_JOB GITHUB_TOKEN GITHUB_REPOSITORY GITHUB_EVENT_PATH
}

test_comment_update() {
    log_test "Test existing comment update with PATCH API (Requirement 8.5, 8.6)"

    source_script

    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"
    export GITHUB_WORKFLOW="update-workflow"
    export GITHUB_JOB="update-job"
    export GITHUB_TOKEN="fake-token"
    export GITHUB_REPOSITORY="owner/repo"

    local event_file=$(create_mock_pr_event 202)
    export GITHUB_EVENT_PATH="$event_file"

    # Create mock response with existing comment
    cat > "$TEST_OUTPUT_DIR/mock_comments.json" <<EOF
[
  {
    "id": 987654321,
    "body": "<!-- strata-update-workflow-update-job -->\\nOld comment content"
  }
]
EOF

    local get_called=false
    local patch_called=false
    local patch_comment_id=""

    curl() {
        local method="GET"
        local url=""

        # Parse arguments
        for ((i=0; i<$#; i++)); do
            if [[ "${!i}" == "-X" ]]; then
                method="${@:$((i+2)):1}"
            fi
        done

        # Determine URL type
        if [[ "$*" == *"/comments/"* ]] && [[ "$*" != *"/comments"* ]]; then
            url="comment_update"
            # Extract comment ID from URL
            patch_comment_id=$(echo "$*" | grep -o '/comments/[0-9]*' | grep -o '[0-9]*')
        elif [[ "$*" == *"/comments" ]]; then
            url="comments"
        fi

        case "$method-$url" in
            "GET-comments")
                get_called=true
                cat "$TEST_OUTPUT_DIR/mock_comments.json"
                echo -e "\n200"
                ;;
            "PATCH-comment_update")
                patch_called=true
                echo '{"id": 987654321}'
                echo -e "\n200"
                ;;
            *)
                echo "[]"
                echo -e "\n200"
                ;;
        esac
    }

    local test_content="Updated comment content"
    local comment_header="🏗️ Updated Header"

    update_pr_comment "$test_content" "$comment_header" "true"

    if [[ "$get_called" == true ]]; then
        echo -e "${GREEN}  ✓${NC} GET API was called to find existing comment"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}  ✗${NC} GET API was not called"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    if [[ "$patch_called" == true ]]; then
        echo -e "${GREEN}  ✓${NC} PATCH API was called to update comment"
        TESTS_PASSED=$((TESTS_PASSED + 1))

        if [[ "$patch_comment_id" == "987654321" ]]; then
            echo -e "${GREEN}  ✓${NC} Correct comment ID was used for update"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}  ✗${NC} Wrong comment ID used: $patch_comment_id"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        echo -e "${RED}  ✗${NC} PATCH API was not called"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    unset GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR GITHUB_WORKFLOW GITHUB_JOB GITHUB_TOKEN GITHUB_REPOSITORY GITHUB_EVENT_PATH
    rm -f "$TEST_OUTPUT_DIR/mock_comments.json"
}

test_comment_update_fallback() {
    log_test "Test fallback to new comment creation when update fails (Requirement 8.11)"

    source_script

    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"
    export GITHUB_WORKFLOW="fallback-workflow"
    export GITHUB_JOB="fallback-job"
    export GITHUB_TOKEN="fake-token"
    export GITHUB_REPOSITORY="owner/repo"

    local event_file=$(create_mock_pr_event 303)
    export GITHUB_EVENT_PATH="$event_file"

    # Create mock response with existing comment
    cat > "$TEST_OUTPUT_DIR/mock_comments.json" <<EOF
[
  {
    "id": 555555555,
    "body": "<!-- strata-fallback-workflow-fallback-job -->\\nExisting comment"
  }
]
EOF

    local get_called=false
    local patch_called=false
    local post_called=false

    curl() {
        local method="GET"
        local url=""

        # Parse arguments
        for ((i=0; i<$#; i++)); do
            if [[ "${!i}" == "-X" ]]; then
                method="${@:$((i+2)):1}"
            fi
        done

        # Determine URL type
        if [[ "$*" == *"/comments/"* ]] && [[ "$*" != *"/comments"* ]]; then
            url="comment_update"
        elif [[ "$*" == *"/comments" ]]; then
            url="comments"
        fi

        case "$method-$url" in
            "GET-comments")
                get_called=true
                cat "$TEST_OUTPUT_DIR/mock_comments.json"
                echo -e "\n200"
                ;;
            "PATCH-comment_update")
                patch_called=true
                # Simulate PATCH failure
                echo '{"error": "Update failed"}'
                echo -e "\n500"
                ;;
            "POST-comments")
                post_called=true
                echo '{"id": 666666666}'
                echo -e "\n201"
                ;;
            *)
                echo "[]"
                echo -e "\n200"
                ;;
        esac
    }

    local test_content="Fallback comment content"
    local comment_header="🏗️ Fallback Header"

    update_pr_comment "$test_content" "$comment_header" "true"

    if [[ "$get_called" == true ]] && [[ "$patch_called" == true ]] && [[ "$post_called" == true ]]; then
        echo -e "${GREEN}  ✓${NC} Attempted update, then fell back to creating new comment"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}  ✗${NC} Fallback sequence not executed correctly"
        echo "      GET called: $get_called, PATCH called: $patch_called, POST called: $post_called"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    unset GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR GITHUB_WORKFLOW GITHUB_JOB GITHUB_TOKEN GITHUB_REPOSITORY GITHUB_EVENT_PATH
    rm -f "$TEST_OUTPUT_DIR/mock_comments.json"
}

# ============================================================================
# Test Cases: Error Handling
# ============================================================================

test_github_error_handling() {
    log_section "GitHub Integration Error Handling"

    log_test "Test handling of GitHub API failures"

    source_script

    export GITHUB_EVENT_NAME="pull_request"
    export INPUT_COMMENT_ON_PR="true"
    export GITHUB_WORKFLOW="error-workflow"
    export GITHUB_JOB="error-job"
    export GITHUB_TOKEN="fake-token"
    export GITHUB_REPOSITORY="owner/repo"

    local event_file=$(create_mock_pr_event 404)
    export GITHUB_EVENT_PATH="$event_file"

    # Mock curl to simulate API failures
    curl() {
        echo '{"error": "API Error"}'
        echo -e "\n500"
    }

    local test_content="Error test content"
    local comment_header="🏗️ Error Header"

    # This should handle the error gracefully and not crash
    local exit_code=0
    update_pr_comment "$test_content" "$comment_header" "false" || exit_code=$?

    # The function should exit with error code 5 (GitHub integration failure)
    assert_equals "5" "$exit_code" "Should exit with GitHub integration failure code on API error"

    unset GITHUB_EVENT_NAME INPUT_COMMENT_ON_PR GITHUB_WORKFLOW GITHUB_JOB GITHUB_TOKEN GITHUB_REPOSITORY GITHUB_EVENT_PATH
}

# ============================================================================
# Main Test Runner
# ============================================================================

run_tests() {
    echo ""
    echo "================================================"
    echo "GitHub Features Integration Tests"
    echo "================================================"

    # Check if action_simplified.sh exists
    if [[ ! -f "action_simplified.sh" ]]; then
        echo -e "${RED}Error: action_simplified.sh not found${NC}"
        echo "Please run this test from the project root directory"
        exit 1
    fi

    # Verify required tools
    if ! command -v jq >/dev/null 2>&1; then
        echo -e "${RED}Error: jq is required but not installed${NC}"
        exit 1
    fi

    # Run all test suites
    test_step_summary_generation
    test_step_summary_multiline_content
    test_pr_context_detection
    test_non_pr_context_skip
    test_pr_number_extraction
    test_comment_marker_generation
    test_comment_creation
    test_comment_update
    test_comment_update_fallback
    test_github_error_handling

    # Print summary
    echo ""
    echo "================================================"
    echo "Test Summary"
    echo "================================================"
    echo -e "Tests run:     ${TESTS_RUN}"
    echo -e "Tests passed:  ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests failed:  ${RED}${TESTS_FAILED}${NC}"
    echo ""

    # Exit with appropriate code
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}TESTS FAILED${NC}"
        exit 1
    else
        echo -e "${GREEN}ALL TESTS PASSED${NC}"
        exit 0
    fi
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    run_tests
fi