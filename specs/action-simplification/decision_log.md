# GitHub Action Simplification - Decision Log

## Decision Record

### 1. Architecture Consolidation
**Decision:** Consolidate from 6 shell modules to maximum 2 files
**Date:** 2025-01-18
**Rationale:**
- Current modular architecture adds complexity without proportional benefit
- Shell scripts don't benefit from the same modularization patterns as compiled languages
- Single file makes debugging and understanding the flow much easier
- Reduces source commands and dependency management

### 2. Remove Binary Compilation Fallback
**Decision:** Remove the source compilation fallback option
**Date:** 2025-01-18
**Rationale:**
- Compilation rarely works in CI environments due to missing Go toolchain
- Adds significant complexity for an edge case
- If binary download fails, it's better to fail fast with clear error
- Users who need custom builds can use different approaches

### 3. Simplify Logging Approach
**Decision:** Remove GitHub Actions group markers for standard operations
**Date:** 2025-01-18
**Rationale:**
- Current approach requires expanding every group to see what happened
- Users report this as a primary pain point
- Simple, scannable output with emoji prefixes is more user-friendly
- Reserve grouping only for genuinely verbose debug output

### 4. Direct Release URL Usage
**Decision:** Use direct GitHub release URLs instead of API calls
**Date:** 2025-01-18
**Rationale:**
- API calls add complexity and potential failure points
- Direct URLs are more reliable and faster
- Pattern is predictable: /releases/latest/download/strata-{platform}-{arch}.tar.gz
- Removes need for JSON parsing and URL construction

### 5. Remove Excessive Security Validations
**Decision:** Keep only essential file existence and readability checks
**Date:** 2025-01-18
**Rationale:**
- GitHub Actions provides a trusted, isolated environment
- Current validations (null bytes, control characters, etc.) are overkill
- Security theater adds complexity without real benefit
- Focus on actual issues like missing files

### 6. Use Existing Strata Multi-Output Capability
**Decision:** Use Strata's existing --file and --file-format flags for dual output
**Date:** 2025-01-18 (Final Revision)
**Rationale:**
- Strata ALREADY supports outputting to stdout while writing different format to file
- No changes needed to Strata at all!
- Single execution: `strata plan summary --output markdown --file /tmp/metadata.json --file-format json`
- This outputs markdown to stdout and JSON to file in one execution
- Current implementation tries to use this but with 300+ lines of unnecessary complexity
- Simplified approach:
  - Run Strata with both outputs
  - Read JSON file for action outputs
  - Use stdout for Step Summary and PR comments
  - ~40 lines of shell code instead of 300+
- Required action outputs extracted from JSON file:
  - `has-changes`: from statistics.total_changes > 0
  - `has-dangers`: from statistics.dangerous_changes > 0
  - `change-count`: from statistics.total_changes
  - `danger-count`: from statistics.dangerous_changes
  - `json-summary`: entire JSON for programmatic use
- Eliminates all the complex dual output coordination, fallback logic, and file synchronization

### 7. Standard Error Handling
**Decision:** Use bash built-in error handling instead of custom wrappers
**Date:** 2025-01-18
**Rationale:**
- `set -euo pipefail` provides robust error handling
- Custom wrapper functions obscure actual errors
- Standard patterns are well-understood by maintainers
- Simpler to debug when things go wrong

### 8. Remove Complex Cache Management
**Decision:** Rely on GitHub's cache action without custom logic
**Date:** 2025-01-18
**Rationale:**
- GitHub's cache action is battle-tested and reliable
- Custom cache validation and management adds complexity
- Binary verification can be done after extraction
- Cache misses are acceptable and quick to resolve

### 9. Simplified PR Comment Handling
**Decision:** Keep comment update capability but dramatically simplify implementation
**Date:** 2025-01-18 (Revised)
**Rationale:**
- User experience is critical - multiple comments per commit (dev/test/prod) creates clutter
- In typical workflows, a single PR might generate 3+ comments per push without updates
- However, current implementation is overly complex (~200 lines)
- Simplified approach:
  - Use GitHub API directly via curl (no GitHub CLI dependency)
  - Simple comment marker: `<!-- strata-${GITHUB_WORKFLOW}-${GITHUB_JOB} -->`
  - Single GET request to find existing comment (with jq parsing)
  - Single PATCH to update or POST to create
  - Max 2 retry attempts (no complex rate limiting)
  - If update fails, create new comment (single fallback)
- This reduces complexity from ~200 lines to ~50 lines while preserving UX
- Environment-specific markers ensure dev/test/prod maintain separate comments
- Users still control behavior via `update-comment` input parameter

### 10. Performance Over Features
**Decision:** Prioritize fast execution over edge case handling
**Date:** 2025-01-18
**Rationale:**
- CI/CD pipeline speed is critical for developer productivity
- 80/20 rule: handle common cases well, fail fast on edge cases
- Complex features add overhead for all users
- Clear errors are better than slow complex recovery attempts

### 11. Release as Minor Version (v1.5.0)
**Decision:** Release as backwards-compatible minor version, not major version
**Date:** 2025-01-18
**Rationale:**
- All user-facing inputs and outputs remain identical
- No breaking changes from user perspective
- Internal implementation details are not part of the public API
- Limited current usage means no need for parallel version support
- Simpler migration - users get improvements automatically
- Version tags: Continue using v1 tag, update to v1.5.0
- No need for deprecation period or migration guide
- Users don't need to change anything in their workflows

## Example Simplified Output Implementation

The simplified output handling would look like this (approximately 40 lines vs current 300+):

```bash
run_strata_analysis() {
  local plan_file=$1
  local output_format=$2
  local show_details=$3
  local expand_all=$4
  local json_file="/tmp/strata_metadata.json"

  echo "🔍 Analyzing Terraform plan"

  # Build command with JSON metadata output to file
  local cmd="$TEMP_DIR/strata plan summary"
  cmd="$cmd --output $output_format"
  cmd="$cmd --file $json_file --file-format json"
  [ "$show_details" = "true" ] && cmd="$cmd --show-details"
  [ "$expand_all" = "true" ] && cmd="$cmd --expand-all"
  cmd="$cmd $plan_file"

  echo "⚙️ Running: $cmd"

  # Execute and capture display output
  if display_output=$($cmd 2>&1); then
    echo "✅ Analysis successful"

    # Read JSON metadata for action outputs
    if [ -f "$json_file" ]; then
      local json=$(cat "$json_file")

      # Extract values using jq
      local total_changes=$(echo "$json" | jq -r '.statistics.total_changes // 0')
      local dangerous_changes=$(echo "$json" | jq -r '.statistics.dangerous_changes // 0')

      # Set GitHub Action outputs
      echo "has-changes=$([[ $total_changes -gt 0 ]] && echo true || echo false)" >> "$GITHUB_OUTPUT"
      echo "has-dangers=$([[ $dangerous_changes -gt 0 ]] && echo true || echo false)" >> "$GITHUB_OUTPUT"
      echo "change-count=$total_changes" >> "$GITHUB_OUTPUT"
      echo "danger-count=$dangerous_changes" >> "$GITHUB_OUTPUT"
      echo "json-summary<<EOF" >> "$GITHUB_OUTPUT"
      cat "$json_file" >> "$GITHUB_OUTPUT"
      echo "EOF" >> "$GITHUB_OUTPUT"
    fi

    # Use display output for Step Summary
    echo "$display_output" >> "$GITHUB_STEP_SUMMARY"

    # Store for PR comment if needed
    DISPLAY_OUTPUT="$display_output"
  else
    echo "❌ Analysis failed"
    exit 1
  fi
}
```

This approach requires a small enhancement to Strata to add the `--json-file` flag, but dramatically simplifies the GitHub Action implementation.

## Example Simplified PR Comment Implementation

The simplified PR comment handling would look like this (approximately 50 lines vs current 200+):

```bash
post_or_update_pr_comment() {
  local pr_number=$1
  local comment_body="$2"

  # Create unique marker for this workflow/job/environment
  local marker="<!-- strata-${GITHUB_WORKFLOW}-${GITHUB_JOB} -->"
  local body_with_marker="${marker}\n${comment_body}"

  if [ "$UPDATE_COMMENT" = "true" ]; then
    echo "🔍 Looking for existing comment to update"

    # Get all comments (single API call)
    local comments=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
      "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments")

    # Find comment with our marker
    local comment_id=$(echo "$comments" | jq -r ".[] | select(.body | contains(\"$marker\")) | .id" | head -1)

    if [ -n "$comment_id" ]; then
      echo "📝 Updating existing comment #$comment_id"

      # Update the comment
      local response=$(curl -s -w "\n%{http_code}" -X PATCH \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"body\": $(echo "$body_with_marker" | jq -R -s .)}" \
        "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}")

      local http_code="${response##*$'\n'}"

      if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo "✅ Comment updated successfully"
        return 0
      else
        echo "⚠️ Update failed, will create new comment"
      fi
    fi
  fi

  # Create new comment (either update=false or update failed)
  echo "📝 Creating new PR comment"

  local response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"body\": $(echo "$body_with_marker" | jq -R -s .)}" \
    "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/issues/${pr_number}/comments")

  local http_code="${response##*$'\n'}"

  if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
    echo "✅ Comment posted successfully"
  else
    echo "❌ Failed to post comment (HTTP $http_code)"
  fi
}
```

This maintains the user experience while dramatically reducing complexity.

### 12. Pre-release Testing Support
**Decision:** Add strata-version input parameter for testing
**Date:** 2025-01-18
**Rationale:**
- Critical for testing action changes with new Strata features
- Allows validation of pre-release versions before making them default
- Enables testing workflow: release Strata pre-release → test with action → release Strata stable
- Simple implementation:
  - Add optional `strata-version` input (default: "latest")
  - When specified, use that version in download URL
  - Example: `strata-version: "v1.4.0-beta.1"`
- Download URL becomes: `releases/download/${STRATA_VERSION}/strata-${OS}-${ARCH}.tar.gz`
- Separate cache keys for different versions prevent conflicts
- Allows users to pin to specific versions for reproducibility

## Example Implementation for Version Support

```bash
# Get version from input or use latest
STRATA_VERSION="${INPUT_STRATA_VERSION:-latest}"

if [ "$STRATA_VERSION" = "latest" ]; then
  # Get latest release URL
  download_url="https://github.com/ArjenSchwarz/strata/releases/latest/download/strata-${OS}-${ARCH}.tar.gz"
  echo "📦 Using latest Strata release"
else
  # Use specific version
  download_url="https://github.com/ArjenSchwarz/strata/releases/download/${STRATA_VERSION}/strata-${OS}-${ARCH}.tar.gz"
  echo "📦 Using Strata version: $STRATA_VERSION"
fi

# Download with simple retry
if ! curl -fsSL "$download_url" | tar -xz -C "$TEMP_DIR"; then
  if [ "$STRATA_VERSION" != "latest" ]; then
    echo "⚠️ Version $STRATA_VERSION not found, trying latest"
    download_url="https://github.com/ArjenSchwarz/strata/releases/latest/download/strata-${OS}-${ARCH}.tar.gz"
    curl -fsSL "$download_url" | tar -xz -C "$TEMP_DIR" || exit 1
  else
    echo "❌ Download failed"
    exit 1
  fi
fi

# Verify and log version
$TEMP_DIR/strata --version
```

## Implementation Priorities

1. **Phase 1**: Core simplification
   - Consolidate modules into single file
   - Simplify logging throughout
   - Remove excessive validations

2. **Phase 2**: Binary download improvement
   - Switch to direct URLs
   - Simplify retry logic
   - Remove compilation fallback
   - Add version parameter support

3. **Phase 3**: Output streamlining
   - Use existing --file and --file-format flags properly
   - Simplify GitHub integration
   - Clean up output formatting

## Release Strategy

- **Version**: v1.5.0 (minor version bump)
- **Backwards Compatibility**: 100% - no breaking changes
- **Migration**: Automatic - users don't need to change anything
- **Tag Update**: Move v1 tag to v1.5.0 after testing
- **Testing**: Use pre-release Strata versions with new `strata-version` parameter

## Risk Mitigation

- **Risk**: Breaking existing workflows
  - **Mitigation**: Maintain same inputs/outputs interface

- **Risk**: Reduced platform support
  - **Mitigation**: Clear error messages for unsupported platforms

- **Risk**: Loss of security validations
  - **Mitigation**: GitHub Actions environment is already secured

- **Risk**: Binary download failures
  - **Mitigation**: Simple retry with clear errors is better than complex fallbacks