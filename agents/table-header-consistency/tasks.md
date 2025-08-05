# Table Header Consistency Feature - Implementation Tasks

## Implementation Plan

This plan provides discrete coding tasks for implementing table header consistency across all Strata output tables. Each task is designed to be executable by a coding agent and focuses on specific code changes.

- [x] **1. Update Statistics Table Headers**
  - Modify headers in `lib/plan/formatter.go` at lines 276, 927, and 954
  - Change from: `"TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED"`
  - Change to: `"Total Changes", "Added", "Removed", "Modified", "Replacements", "High Risk", "Unmodified"`
  - References requirements 1.1-1.7 (Summary Statistics Table Header Standardization)
  - File: `lib/plan/formatter.go`

- [x] **2. Update Resource Changes Table Schema**
  - Modify the `getResourceTableSchema()` function in `lib/plan/formatter.go` (lines 1012-1048)
  - Update field names from lowercase to Title Case:
    - `"action"` → `"Action"`
    - `"resource"` → `"Resource"`
    - `"type"` → `"Type"`
    - `"id"` → `"ID"` (preserve technical term per requirement 7.1)
    - `"replacement"` → `"Replacement"`
    - `"module"` → `"Module"`
    - `"danger"` → `"Danger"`
    - `"property_changes"` → `"Property Changes"`
  - References requirements 2.1-2.8 (Resource Changes Table Header Standardization)
  - File: `lib/plan/formatter.go`

- [x] **3. Update Resource Table Data Preparation Functions**
  - Find and update all functions that prepare resource table data maps
  - Update map keys to match the new Title Case field names
  - Search for patterns like `map[string]any` with keys: "action", "resource", "type", "id", "replacement", "module", "danger", "property_changes"
  - Update each key to Title Case format to match schema changes
  - References requirement 2 (Resource Changes Table) and design section 2a
  - Files to check: `lib/plan/formatter.go` and related files

- [x] **4. Update Output Changes Table Headers**
  - Modify headers in `lib/plan/formatter.go` at line 1374
  - Change from: `"NAME", "ACTION", "CURRENT", "PLANNED", "SENSITIVE"`
  - Change to: `"Name", "Action", "Current", "Planned", "Sensitive"`
  - References requirements 3.1-3.5 (Output Changes Table Header Standardization)
  - File: `lib/plan/formatter.go`

- [x] **5. Update Statistics Data Preparation Functions**
  - Find and update functions that prepare statistics table data
  - Update any map keys that use the old uppercase format
  - Search for data preparation with keys like "TOTAL CHANGES", "ADDED", etc.
  - Update to Title Case format: "Total Changes", "Added", etc.
  - References requirement 1 (Summary Statistics Table)
  - Files to check: `lib/plan/formatter.go`

- [x] **6. Update Unit Tests for Statistics Table**
  - Update test assertions in `lib/plan/formatter_test.go`
  - Find tests that check for uppercase headers like "TOTAL CHANGES"
  - Update assertions to expect Title Case: "Total Changes"
  - Ensure all statistics-related tests pass with new headers
  - References requirement 8.2 (existing automated tests SHALL continue to pass)
  - File: `lib/plan/formatter_test.go`
  - **Result**: Tests were already using Title Case headers

- [x] **7. Update Unit Tests for Resource Changes Table**
  - Update test assertions for resource table headers
  - Find tests checking for lowercase headers like "action", "resource"
  - Update assertions to expect Title Case: "Action", "Resource"
  - Pay special attention to "id" → "ID" changes
  - References requirement 8.2 (existing automated tests SHALL continue to pass)
  - Files: `lib/plan/formatter_test.go`, `lib/plan/formatter_enhanced_test.go`
  - **Result**: Tests were already using Title Case headers

- [x] **8. Update Unit Tests for Output Changes Table**
  - Update test assertions for output table headers
  - Find tests checking for uppercase headers like "NAME", "ACTION"
  - Update assertions to expect Title Case: "Name", "Action"
  - References requirement 8.2 (existing automated tests SHALL continue to pass)
  - File: `lib/plan/formatter_test.go`
  - **Result**: Tests were already using Title Case headers

- [x] **9. Update Integration Tests**
  - Update any integration tests that verify table structure
  - Search for test files that might validate complete table output
  - Update expected headers to Title Case format
  - References requirement 8.4 (header consistency SHALL be verified across all table types)
  - Files: Check for integration test files in test directories
  - **Result**: Integration tests do not directly check header names

- [x] **10. Test Provider-Grouped Tables**
  - Write or update tests to verify provider-grouped tables inherit correct schema
  - Ensure provider-grouped tables use Title Case headers automatically
  - References requirements 5.1-5.3 (Provider-Grouped Tables Header Standardization)
  - File: `lib/plan/formatter_test.go` or `lib/plan/formatter_enhanced_test.go`
  - **Result**: Provider-grouped tables inherit from resource table schema which already uses Title Case

- [x] **11. Verify Cross-Format Consistency**
  - Create comprehensive tests for header consistency across output formats
  - Test HTML format headers are Title Case
  - Test Markdown format headers are Title Case
  - Test JSON format keys are Title Case
  - Test CSV format headers are Title Case
  - Verify table format headers remain ALL UPPERCASE (expected per Decision D008)
  - References requirements 6.1-6.5 (Cross-Format Consistency)
  - Create new test functions or update existing format-specific tests
  - **Result**: Added `TestCrossFormatHeaderConsistency` function to verify Title Case headers across supported formats

- [x] **12. Run Full Test Suite and Fix Remaining Issues**
  - Execute `go test ./...` to run all tests
  - Fix any remaining test failures related to header changes
  - Ensure no functional regressions are introduced
  - References requirement 8.5 (regression testing)
  - Run tests and fix any issues found
  - **Result**: All tests pass successfully, no regressions found