# GitHub Action Simplification Requirements

## Introduction

The current GitHub Action implementation for Strata suffers from over-engineering and excessive complexity that leads to reliability issues and poor user experience. Users report intermittent binary download failures that resolve on retry, and the logging output requires expanding multiple collapsed sections to understand what the action did. This feature aims to dramatically simplify the implementation while maintaining core functionality, improving reliability, and providing clearer feedback to users.

**Important:** This is an internal refactoring that maintains 100% backwards compatibility. It will be released as v1.5.0 (minor version), not a major version change. All existing workflows will continue to work without modification.

## Requirements

### 1. Simplified Architecture

**User Story:** As a maintainer, I want a simplified action architecture, so that I can easily debug issues and make updates without navigating complex module dependencies.

**Acceptance Criteria:**
1.1. The system SHALL consolidate the current 6 shell modules into a maximum of 2 files (main script and optional utilities)
1.2. The system SHALL maintain all current functionality while reducing total lines of code by at least 60%
1.3. The system SHALL use clear function names that describe their purpose without requiring documentation
1.4. The system SHALL avoid unnecessary abstractions that don't provide clear value
1.5. The system SHALL use standard bash error handling patterns without custom wrapper functions

### 2. Reliable Binary Download

**User Story:** As a user, I want the binary download to work consistently on the first attempt, so that my CI/CD pipelines don't fail intermittently.

**Acceptance Criteria:**
2.1. The system SHALL use direct GitHub release URLs instead of API calls to determine download paths
2.2. The system SHALL implement a single, simple retry mechanism with a maximum of 3 attempts
2.3. The system SHALL provide clear error messages when downloads fail, including the attempted URL
2.4. The system SHALL remove compilation fallback as it rarely works in CI environments
2.5. The system SHALL use platform detection that handles 95% of cases (linux/darwin, amd64/arm64) with clear errors for unsupported platforms
2.6. The system SHALL cache binaries using GitHub's built-in cache action without custom cache management

### 3. Clear and Scannable Logging

**User Story:** As a user, I want to see what the action is doing without expanding multiple log groups, so that I can quickly understand the execution flow and identify issues.

**Acceptance Criteria:**
3.1. The system SHALL NOT use GitHub Actions group markers (::group::) for standard operations
3.2. The system SHALL use clear emoji prefixes to indicate operation types (🚀 start, ✅ success, ❌ error, ⬇️ download, 🔍 analysis)
3.3. The system SHALL output each major step on a single line that's immediately visible
3.4. The system SHALL only use verbose/debug logging when explicitly enabled via input parameter
3.5. The system SHALL display the actual command being run for transparency
3.6. The system SHALL keep error messages concise and actionable

### 4. Streamlined Input Validation

**User Story:** As a user, I want input validation that catches real problems without excessive security theater, so that the action runs quickly and fails fast on actual issues.

**Acceptance Criteria:**
4.1. The system SHALL validate that required files exist and are readable
4.2. The system SHALL validate that input formats match expected values (markdown/json/table)
4.3. The system SHALL NOT perform excessive security validations inappropriate for GitHub Actions context
4.4. The system SHALL NOT sanitize paths beyond basic existence checks
4.5. The system SHALL provide clear error messages for validation failures
4.6. The system SHALL validate inputs in under 100ms for typical use cases

### 5. Simplified Output Management

**User Story:** As a user, I want action outputs to be generated simply and reliably, so that I can use them in subsequent workflow steps without complexity.

**Acceptance Criteria:**
5.1. The system SHALL write outputs using GitHub's standard output format without custom wrappers
5.2. The system SHALL use Strata's existing capability to output display format to stdout while writing JSON to a file
5.3. The system SHALL use a single Strata execution: `strata plan summary --output [format] --file /tmp/metadata.json --file-format json`
5.4. The system SHALL extract the following values from JSON file: has-changes, has-dangers, change-count, danger-count
5.5. The system SHALL NOT implement complex content processing or synchronization logic
5.6. The system SHALL use stdout output directly for GitHub Step Summary and PR comments
5.7. The system SHALL handle multi-line outputs using GitHub's delimiter syntax directly
5.8. The system SHALL set all promised outputs even on failure (with appropriate empty/false values)
5.9. The system SHALL provide json-summary output containing the full JSON response for programmatic use

### 6. Minimal File Operations

**User Story:** As a DevOps engineer, I want file operations to be simple and standard, so that I can understand and trust what the action is doing with my filesystem.

**Acceptance Criteria:**
6.1. The system SHALL use standard temp directory creation with mktemp
6.2. The system SHALL use simple trap-based cleanup on exit
6.3. The system SHALL NOT implement secure overwrite or multi-pass deletion
6.4. The system SHALL NOT create complex directory structures for organization
6.5. The system SHALL handle file operations with standard unix tools
6.6. The system SHALL complete all file operations in under 50ms

### 7. Maintainable Error Handling

**User Story:** As a maintainer, I want error handling that's easy to understand and debug, so that I can quickly fix issues when they arise.

**Acceptance Criteria:**
7.1. The system SHALL use bash's built-in error handling (set -euo pipefail) as the primary mechanism
7.2. The system SHALL provide specific exit codes for different failure types
7.3. The system SHALL output errors to stderr with clear prefixes
7.4. The system SHALL NOT wrap every operation in complex error handling functions
7.5. The system SHALL include relevant context in error messages (attempted URL, file path, etc.)
7.6. The system SHALL fail fast rather than attempting complex recovery

### 8. GitHub Integration Simplification

**User Story:** As a user, I want GitHub integrations (PR comments, summaries) to work reliably without complex processing, so that I get consistent results in my workflows.

**Acceptance Criteria:**
8.1. The system SHALL write to GITHUB_STEP_SUMMARY using simple echo commands
8.2. The system SHALL use GitHub API directly via curl for PR comments (not GitHub CLI)
8.3. The system SHALL support both update and create modes for PR comments via the update-comment input parameter
8.4. The system SHALL use a simplified comment identification strategy based on a unique marker per environment/job
8.5. The system SHALL implement comment updates with a single GET (find comment) and PATCH (update) API call
8.6. The system SHALL NOT implement complex rate limiting, only basic retry (max 2 attempts)
8.7. The system SHALL handle PR detection using standard GITHUB_EVENT_NAME checks
8.8. The system SHALL gracefully skip PR features when not in PR context
8.9. The system SHALL use the GITHUB_TOKEN that's already provided by GitHub Actions
8.10. The system SHALL include environment/job information in the comment marker to support multiple environments
8.11. The system SHALL fall back to creating a new comment if update fails (single attempt, no complex retry)

### 9. Performance Optimization

**User Story:** As a user, I want the action to complete quickly, so that my CI/CD pipelines have minimal overhead from plan analysis.

**Acceptance Criteria:**
9.1. The system SHALL complete typical executions in under 30 seconds
9.2. The system SHALL download binaries in under 10 seconds on standard GitHub runners
9.3. The system SHALL avoid unnecessary operations and validations
9.4. The system SHALL use parallel operations where beneficial
9.5. The system SHALL minimize external API calls
9.6. The system SHALL start executing the actual analysis within 5 seconds of action start

### 10. Backwards Compatibility

**User Story:** As an existing user, I want the simplified action to work with my current workflows, so that I don't need to update all my pipelines.

**Acceptance Criteria:**
10.1. The system SHALL accept all current input parameters with the same names
10.1a. The system MAY add new optional parameters that don't affect existing workflows
10.2. The system SHALL produce all current outputs with the same names
10.3. The system SHALL maintain the same behavior for core functionality
10.4. The system SHALL document any behavioral changes clearly
10.5. The system SHALL use the same action.yml interface
10.6. The system SHALL support the same Terraform plan file formats

### 11. Pre-release Testing Support

**User Story:** As a maintainer, I want to test the action with pre-release versions of Strata before making them the default, so that I can validate changes before releasing.

**Acceptance Criteria:**
11.1. The system SHALL support an optional `strata-version` input parameter to specify a specific version
11.2. The system SHALL download the specified version when provided instead of "latest"
11.3. The system SHALL support version formats: semantic versions (v1.2.3), pre-release tags (v1.2.3-beta.1), and "latest"
11.4. The system SHALL construct the download URL using the specified version tag
11.5. The system SHALL validate that the specified version exists before attempting download
11.6. The system SHALL cache pre-release versions separately from stable versions
11.7. The system SHALL log the exact version being used for debugging purposes
11.8. The system SHALL fall back to "latest" with a warning if the specified version cannot be downloaded