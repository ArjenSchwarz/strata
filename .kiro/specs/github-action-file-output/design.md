# GitHub Action File Output Integration Design Document

## Overview

This design document outlines the enhancement of the existing Strata GitHub Action to leverage the new dual-output file system. The enhancement will optimize the action to use table format for terminal display while generating markdown-formatted content for GitHub-specific features like PR comments and step summaries.

The design focuses on integrating the new `--file` and `--file-format` flags into the GitHub Action workflow to provide optimal formatting for different output contexts without breaking existing functionality.

## Architecture

The enhanced GitHub Action will follow a dual-output architecture that leverages Strata's new file output capabilities:

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Action Layer                      │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Action Script (action.sh)                 │ │
│  │  • Input validation and processing                     │ │
│  │  • Strata execution with dual output                   │ │
│  │  • Output processing and distribution                  │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Strata Execution Layer                   │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │  Primary Output │  │  File Output    │                  │
│  │  (stdout)       │  │  (temp file)    │                  │
│  │  Format: table  │  │  Format: markdown│                 │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Output Distribution Layer                   │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │  Step Summary   │  │  PR Comments    │                  │
│  │  (markdown)     │  │  (markdown)     │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

### Integration Strategy

The action will use Strata's dual-output system by:

1. **Primary Execution**: Run Strata with table format for stdout (terminal display)
2. **File Generation**: Simultaneously generate markdown output to a temporary file
3. **Content Distribution**: Use the markdown file content for GitHub-specific features
4. **Cleanup**: Remove temporary files after processing

## Components and Interfaces

### 1. Enhanced Strata Execution Component

**Location**: `action.sh` (modified `run_strata` function)

```bash
# Enhanced function to run Strata with dual output
run_strata_dual_output() {
    local stdout_format=$1
    local plan_file=$2
    local show_details=$3
    
    # Create temporary file for markdown output
    local temp_markdown_file=$(mktemp)
    trap "rm -f $temp_markdown_file" EXIT
    
    # Convert plan file to absolute path if it's relative
    if [[ "$plan_file" != /* ]]; then
        plan_file="$(pwd)/$plan_file"
    fi
    
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
        cmd="$cmd --show-details"
    fi
    
    # Add dual output: stdout format for display, markdown for file
    cmd="$cmd --output $stdout_format --file $temp_markdown_file --file-format markdown"
    
    # Add plan file
    cmd="$cmd $plan_file"
    
    log "Running Strata with dual output" "Display=$stdout_format File=markdown"
    
    # Execute command and capture stdout
    local stdout_output
    stdout_output=$(eval "$cmd" 2>&1)
    local exit_code=$?
    
    if [ $exit_code -ne 0 ]; then
        warning "Strata execution failed with exit code $exit_code"
        warning "Error output: $stdout_output"
        echo "$stdout_output"
        return $exit_code
    fi
    
    # Store markdown content in global variable for later use
    if [ -f "$temp_markdown_file" ]; then
        MARKDOWN_CONTENT=$(cat "$temp_markdown_file")
        log "Successfully generated markdown content" "Size: $(wc -c < "$temp_markdown_file") bytes"
    else
        warning "Markdown file was not generated, falling back to stdout content"
        MARKDOWN_CONTENT="$stdout_output"
    fi
    
    echo "$stdout_output"
    return $exit_code
}
```

### 2. Content Processing Component

**Location**: `action.sh` (new functions)

```bash
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

# Function to add workflow information
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
```

### 3. Output Distribution Component

**Location**: `action.sh` (modified output generation)

```bash
# Function to distribute content to different GitHub contexts
distribute_output() {
    local stdout_output="$1"
    local markdown_content="$2"
    
    # Write to GitHub Step Summary using processed markdown
    if [ -n "$GITHUB_STEP_SUMMARY" ]; then
        log "Writing to GitHub Step Summary"
        local step_summary_content
        step_summary_content=$(process_markdown_for_context "step-summary" "$markdown_content")
        echo "$step_summary_content" >> "$GITHUB_STEP_SUMMARY"
    fi
    
    # Handle PR comments if enabled
    if [ "$COMMENT_ON_PR" = "true" ] && [ "$GITHUB_EVENT_NAME" = "pull_request" ]; then
        log "Preparing PR comment"
        local pr_comment_content
        pr_comment_content=$(process_markdown_for_context "pr-comment" "$markdown_content")
        
        # Extract PR number and post comment
        local pr_number
        pr_number=$(jq -r .pull_request.number "$GITHUB_EVENT_PATH" 2>/dev/null)
        
        if [ -n "$pr_number" ] && [ "$pr_number" != "null" ]; then
            post_pr_comment "$pr_number" "$pr_comment_content"
        else
            warning "Could not determine PR number, skipping comment"
        fi
    fi
    
    # Set action outputs (using stdout content for summary, markdown for json-summary)
    set_output "summary" "$stdout_output"
    set_output "json-summary" "$markdown_content"
}
```

### 4. Error Handling Component

**Location**: `action.sh` (enhanced error handling)

```bash
# Enhanced error handling for dual output
handle_dual_output_error() {
    local exit_code=$1
    local stdout_output="$2"
    
    if [ $exit_code -ne 0 ]; then
        warning "Strata analysis failed with exit code $exit_code"
        
        # Create fallback content for GitHub features
        local error_content="## ⚠️ Strata Analysis Error

Strata encountered an issue while analyzing the Terraform plan.

**Error Details:**
\`\`\`
$stdout_output
\`\`\`

**Possible causes:**
- Invalid plan file format
- Unsupported Terraform version
- Plan file corruption

Please check the action logs for more details."
        
        # Use error content as markdown fallback
        MARKDOWN_CONTENT="$error_content"
        
        # Still distribute output to provide feedback
        distribute_output "$stdout_output" "$error_content"
    fi
}
```

## Data Models

### 1. Output Context Structure

```bash
# Global variables to manage output contexts
STDOUT_OUTPUT=""           # Table format for terminal display
MARKDOWN_CONTENT=""        # Markdown format for GitHub features
JSON_OUTPUT=""            # JSON format for parsing statistics
TEMP_FILES=()             # Array to track temporary files for cleanup
```

### 2. Processing Configuration

```bash
# Configuration for different output contexts
declare -A OUTPUT_CONTEXTS=(
    ["display"]="table"
    ["step-summary"]="markdown"
    ["pr-comment"]="markdown"
    ["statistics"]="json"
)
```

### 3. Content Enhancement Structure

```bash
# Structure for content enhancements per context
declare -A CONTENT_ENHANCEMENTS=(
    ["step-summary"]="workflow-info,footer"
    ["pr-comment"]="comment-id,footer"
)
```

## Implementation Details

### 1. Modified Action Execution Flow

The main execution flow will be updated to use dual output:

```bash
# Main execution flow (simplified)
main() {
    # ... existing validation ...
    
    # Run Strata with dual output
    log "Running Strata analysis with optimized output formats"
    STDOUT_OUTPUT=$(run_strata_dual_output "table" "$INPUT_PLAN_FILE" "$SHOW_DETAILS")
    STRATA_EXIT_CODE=$?
    
    # Handle errors
    if [ $STRATA_EXIT_CODE -ne 0 ]; then
        handle_dual_output_error $STRATA_EXIT_CODE "$STDOUT_OUTPUT"
    fi
    
    # Get JSON output for statistics parsing
    JSON_OUTPUT=$(run_strata "json" "$INPUT_PLAN_FILE" "false")
    
    # Parse statistics from JSON
    parse_statistics "$JSON_OUTPUT"
    
    # Distribute content to different contexts
    distribute_output "$STDOUT_OUTPUT" "$MARKDOWN_CONTENT"
    
    # Set all action outputs
    set_all_outputs
    
    exit $STRATA_EXIT_CODE
}
```

### 2. Temporary File Management

```bash
# Function to create and track temporary files
create_temp_file() {
    local temp_file=$(mktemp)
    TEMP_FILES+=("$temp_file")
    echo "$temp_file"
}

# Function to cleanup all temporary files
cleanup_temp_files() {
    for temp_file in "${TEMP_FILES[@]}"; do
        if [ -f "$temp_file" ]; then
            rm -f "$temp_file"
            log "Cleaned up temporary file: $temp_file"
        fi
    done
    TEMP_FILES=()
}

# Set trap for cleanup
trap cleanup_temp_files EXIT
```

### 3. Format Optimization

```bash
# Function to optimize content for specific contexts
optimize_content_for_context() {
    local context=$1
    local content="$2"
    
    case "$context" in
        "step-summary")
            # Add collapsible sections for better organization
            echo "$content" | sed 's/^## /### /' | add_collapsible_sections
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

# Function to add collapsible sections
add_collapsible_sections() {
    # Implementation to wrap large sections in <details> tags
    # This would be a more complex function to identify and wrap content
    cat
}

# Function to limit content size
limit_content_size() {
    local max_size=$1
    local content=$(cat)
    
    if [ ${#content} -gt $max_size ]; then
        echo "${content:0:$max_size}"
        echo ""
        echo "... (content truncated due to size limits)"
    else
        echo "$content"
    fi
}
```

## Error Handling

### 1. File Operation Errors

```bash
# Handle file operation failures gracefully
handle_file_error() {
    local operation=$1
    local file_path=$2
    local error_message=$3
    
    warning "File operation failed: $operation on $file_path - $error_message"
    
    case "$operation" in
        "create")
            warning "Could not create temporary file, falling back to stdout-only mode"
            MARKDOWN_CONTENT="$STDOUT_OUTPUT"
            ;;
        "write")
            warning "Could not write to file, using fallback content"
            MARKDOWN_CONTENT="Error: Could not generate markdown output"
            ;;
        "read")
            warning "Could not read file, using stdout content as fallback"
            MARKDOWN_CONTENT="$STDOUT_OUTPUT"
            ;;
    esac
}
```

### 2. Format Conversion Errors

```bash
# Handle format conversion failures
handle_format_error() {
    local source_format=$1
    local target_format=$2
    local error_message=$3
    
    warning "Format conversion failed: $source_format to $target_format - $error_message"
    
    # Provide fallback content
    MARKDOWN_CONTENT="## Strata Analysis Results

**Note:** Markdown formatting failed, displaying raw output:

\`\`\`
$STDOUT_OUTPUT
\`\`\`"
}
```

### 3. GitHub API Errors

```bash
# Enhanced GitHub API error handling
handle_github_api_error() {
    local operation=$1
    local http_status=$2
    local response_body=$3
    
    case "$http_status" in
        403)
            if echo "$response_body" | grep -q "rate limit"; then
                warning "GitHub API rate limit reached, will retry with backoff"
                return 1  # Indicate retry needed
            else
                warning "GitHub API permission denied for $operation"
                return 2  # Indicate permanent failure
            fi
            ;;
        404)
            warning "GitHub API resource not found for $operation"
            return 2
            ;;
        *)
            warning "GitHub API error for $operation: HTTP $http_status"
            return 1
            ;;
    esac
}
```

## Testing Strategy

### 1. Unit Testing Approach

**Dual Output Testing**:
```bash
test_dual_output_generation() {
    # Setup test environment
    local test_plan="test_plan.tfplan"
    create_test_plan "$test_plan"
    
    # Run dual output function
    local stdout_result
    stdout_result=$(run_strata_dual_output "table" "$test_plan" "false")
    local exit_code=$?
    
    # Verify stdout output
    assert_equals 0 $exit_code "Strata execution should succeed"
    assert_contains "$stdout_result" "TO ADD" "Stdout should contain table headers"
    
    # Verify markdown content was generated
    assert_not_empty "$MARKDOWN_CONTENT" "Markdown content should be generated"
    assert_contains "$MARKDOWN_CONTENT" "##" "Markdown should contain headers"
    
    # Verify content differences
    assert_not_equals "$stdout_result" "$MARKDOWN_CONTENT" "Outputs should be in different formats"
}
```

**Content Processing Testing**:
```bash
test_content_processing() {
    local test_content="## Test Content\n\nSome test data"
    
    # Test step summary processing
    local step_summary_result
    step_summary_result=$(process_markdown_for_context "step-summary" "$test_content")
    assert_contains "$step_summary_result" "Workflow Information" "Should add workflow info"
    
    # Test PR comment processing
    local pr_comment_result
    pr_comment_result=$(process_markdown_for_context "pr-comment" "$test_content")
    assert_contains "$pr_comment_result" "strata-comment-id" "Should add comment ID"
}
```

### 2. Integration Testing

**End-to-End GitHub Action Testing**:
```bash
test_github_action_integration() {
    # Setup test environment with GitHub context
    export GITHUB_EVENT_NAME="pull_request"
    export GITHUB_STEP_SUMMARY="/tmp/test_summary"
    export GITHUB_EVENT_PATH="/tmp/test_event.json"
    
    # Create test PR event
    echo '{"pull_request": {"number": 123}}' > "$GITHUB_EVENT_PATH"
    
    # Run action
    ./action.sh
    local exit_code=$?
    
    # Verify results
    assert_equals 0 $exit_code "Action should succeed"
    assert_file_exists "$GITHUB_STEP_SUMMARY" "Step summary should be created"
    assert_file_contains "$GITHUB_STEP_SUMMARY" "Terraform Plan Summary" "Should contain expected content"
}
```

### 3. Error Handling Testing

**File Operation Error Testing**:
```bash
test_file_error_handling() {
    # Create read-only directory to trigger permission error
    local readonly_dir="/tmp/readonly_test"
    mkdir -p "$readonly_dir"
    chmod 444 "$readonly_dir"
    
    # Override temp directory to trigger error
    export TMPDIR="$readonly_dir"
    
    # Run function and verify graceful handling
    local result
    result=$(run_strata_dual_output "table" "test.tfplan" "false")
    local exit_code=$?
    
    # Should not fail completely
    assert_not_equals 1 $exit_code "Should handle file errors gracefully"
    assert_not_empty "$MARKDOWN_CONTENT" "Should provide fallback content"
    
    # Cleanup
    chmod 755 "$readonly_dir"
    rm -rf "$readonly_dir"
}
```

## Security Considerations

### 1. Temporary File Security

```bash
# Secure temporary file creation
create_secure_temp_file() {
    local temp_file
    temp_file=$(mktemp)
    
    # Set restrictive permissions
    chmod 600 "$temp_file"
    
    # Verify file was created securely
    if [ ! -f "$temp_file" ]; then
        error "Failed to create secure temporary file"
        return 1
    fi
    
    TEMP_FILES+=("$temp_file")
    echo "$temp_file"
}
```

### 2. Content Sanitization

```bash
# Sanitize content for GitHub output
sanitize_github_content() {
    local content="$1"
    
    # Remove potential script tags and other dangerous content
    echo "$content" | sed 's/<script[^>]*>.*<\/script>//gi' | \
                     sed 's/<iframe[^>]*>.*<\/iframe>//gi' | \
                     sed 's/javascript:[^"'\'']*//gi'
}
```

### 3. Path Validation

```bash
# Validate file paths for security
validate_file_path() {
    local file_path="$1"
    
    # Check for path traversal attempts
    if [[ "$file_path" == *".."* ]]; then
        error "Path traversal detected in file path: $file_path"
        return 1
    fi
    
    # Ensure path is within allowed directories
    local resolved_path
    resolved_path=$(realpath "$file_path" 2>/dev/null)
    if [ $? -ne 0 ]; then
        error "Invalid file path: $file_path"
        return 1
    fi
    
    return 0
}
```

This design provides a comprehensive approach to integrating the new dual-output file system into the GitHub Action while maintaining security, reliability, and optimal user experience across different GitHub contexts.