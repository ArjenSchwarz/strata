# Implementation Tasks: Terraform Unknown Values and Outputs Feature

## Overview
Convert the feature design into a series of prompts for a code-generation LLM that will implement each step in a test-driven manner. Implementation follows two phases: unknown values processing first, then outputs section.

## Implementation Tasks

### 1. Enhance Data Models for Unknown Values Support
- [ ] 1.1. Add unknown value fields to `PropertyChange` struct in `lib/plan/models.go`
  - Add `IsUnknown bool` field (requirement 1.6)
  - Add `UnknownType string` field to track before/after/both unknown states (requirement 1.7)
  - Update JSON tags for proper serialization across output formats (requirement 3.4)

- [ ] 1.2. Add unknown value tracking to `ResourceChange` struct in `lib/plan/models.go`
  - Add `HasUnknownValues bool` field to indicate resource contains unknown properties (requirement 1.2)
  - Add `UnknownProperties []string` field to list unknown property paths (requirement 1.5)
  - Ensure integration with existing danger highlighting fields (requirement 3.1)

### 2. Implement Unknown Value Detection in Analyzer
- [ ] 2.1. Create helper function for unknown value detection in `lib/plan/analyzer.go`
  - Implement `isValueUnknown(afterUnknown any, path string) bool` function (requirement 1.6)
  - Implement `getUnknownValueDisplay() string` function returning "(known after apply)" (requirement 1.3)
  - Add comprehensive test cases for nested object unknown value detection

- [ ] 2.2. Enhance `compareObjects` function to handle unknown values in `lib/plan/analyzer.go`
  - Add `afterUnknown any` parameter to function signature (requirement 1.6)
  - Implement logic to check unknown status before standard property comparison (requirement 1.6)
  - Handle edge cases: unknown to known, known to unknown, and remaining unknown transitions (requirements 1.4, 1.7)
  - Ensure unknown values override deletion detection logic (requirement 1.2)

- [ ] 2.3. Update `ProcessResourceChange` function in `lib/plan/analyzer.go`
  - Extract `after_unknown` field from Terraform JSON change data (requirement 1.6)
  - Pass `after_unknown` data to enhanced `compareObjects` function
  - Populate `HasUnknownValues` and `UnknownProperties` fields in `ResourceChange` (requirement 1.5)
  - Maintain integration with existing danger highlighting (requirement 3.1)

### 3. Write Comprehensive Tests for Unknown Values Processing
- [ ] 3.1. Create unit tests for unknown value detection functions in `lib/plan/analyzer_test.go`
  - Test `isValueUnknown` with various `after_unknown` structures
  - Test `getUnknownValueDisplay` returns exact "(known after apply)" string (requirement 1.3)
  - Test edge cases: nested objects, arrays, complex data types

- [ ] 3.2. Create integration tests for enhanced `compareObjects` function
  - Test unknown value detection with real Terraform plan data (requirement 1.1)
  - Test property change classification with unknown values (requirements 1.4, 1.5, 1.7)
  - Test integration with existing sensitive property detection (requirement 3.1)
  - Verify unknown values don't appear as deletions (requirement 1.2)

### 4. Update Output Formatting for Unknown Values Display
- [ ] 4.1. Enhance property change formatting in `lib/plan/formatter.go`
  - Update table format to display "(known after apply)" for unknown values (requirement 1.3)
  - Update JSON format to include unknown value information (requirement 3.4)
  - Update HTML and Markdown formats for consistent unknown value display (requirement 3.4)
  - Ensure unknown values integrate with collapsible sections (requirement 3.2)

- [ ] 4.2. Test unknown value display across all output formats
  - Create test cases for table, JSON, HTML, and Markdown output (requirement 3.4)
  - Verify "(known after apply)" appears consistently across formats (requirement 1.3)
  - Test integration with danger highlighting for unknown sensitive properties (requirement 3.1)

### 5. Enhance Data Models for Outputs Support
- [ ] 5.1. Add output change fields to `OutputChange` struct in `lib/plan/models.go`
  - Add `IsUnknown bool` field for unknown output values (requirement 2.3)
  - Add `Action string` field for "Add", "Modify", "Remove" actions (requirements 2.5, 2.6, 2.7)
  - Add `Indicator string` field for "+", "~", "-" visual indicators (requirements 2.5, 2.6, 2.7)
  - Update JSON tags for proper serialization (requirement 3.4)

- [ ] 5.2. Add outputs tracking to `PlanSummary` struct in `lib/plan/models.go`
  - Ensure `OutputChanges []OutputChange` field exists for outputs section (requirement 2.1)
  - Ensure outputs are properly displayed in plan summary (requirement 2.1)

### 6. Implement Outputs Processing in Analyzer
- [ ] 6.1. Create outputs processing function in `lib/plan/analyzer.go`
  - Implement `ProcessOutputChanges(plan *tfjson.Plan) ([]OutputChange, error)` function (requirement 2.1)
  - Extract output changes from plan's `OutputChanges` field
  - Handle missing `output_changes` field gracefully (return empty list, requirement 2.8)

- [ ] 6.2. Create individual output analysis function in `lib/plan/analyzer.go`
  - Implement `analyzeOutputChange(name string, change *tfjson.Change) (*OutputChange, error)` function
  - Detect output change actions: create, update, delete (requirements 2.5, 2.6, 2.7)
  - Apply unknown value detection using existing logic (requirement 2.3)
  - Handle sensitive output detection with "(sensitive value)" display (requirement 2.4)

- [ ] 6.3. Update main analysis workflow in `lib/plan/analyzer.go`
  - Integrate outputs processing into `GenerateSummary` function (requirement 2.1)
  - Add output changes to plan summary structure
  - Add outputs to plan summary for display (requirement 2.1)

### 7. Write Comprehensive Tests for Outputs Processing
- [ ] 7.1. Create unit tests for outputs processing functions in `lib/plan/analyzer_test.go`
  - Test `ProcessOutputChanges` with various output change scenarios
  - Test `analyzeOutputChange` for create, update, delete actions (requirements 2.5, 2.6, 2.7)
  - Test sensitive output handling with "(sensitive value)" display (requirement 2.4)
  - Test unknown output values display "(known after apply)" (requirement 2.3)

- [ ] 7.2. Create integration tests for end-to-end outputs processing
  - Test outputs section integration with resource changes (requirement 2.1)
  - Test empty outputs section suppression (requirement 2.8)
  - Test outputs display consistency across all formats (requirement 3.4)

### 8. Implement Outputs Section Formatting
- [ ] 8.1. Add outputs section rendering to `lib/plan/formatter.go`
  - Implement 5-column table format: NAME, ACTION, CURRENT, PLANNED, SENSITIVE (requirement 2.2)
  - Place outputs section after resource changes section (requirement 2.1)
  - Suppress section entirely when no output changes exist (requirement 2.8)

- [ ] 8.2. Implement outputs formatting across all output formats
  - Add outputs table support for table format with visual indicators (requirements 2.5, 2.6, 2.7)
  - Add outputs section support for JSON format (requirement 3.4)
  - Add outputs section support for HTML and Markdown formats (requirement 3.4)
  - Ensure sensitive value masking with ⚠️ indicator (requirement 2.4)

### 9. Create End-to-End Integration Tests
- [ ] 9.1. Create comprehensive integration tests in `lib/plan/analyzer_test.go`
  - Test complete workflow with real Terraform plan containing unknown values and outputs
  - Verify unknown values display correctly and don't appear as deletions (requirements 1.1, 1.2)
  - Verify outputs section displays with correct 5-column format (requirement 2.2)
  - Test integration with existing danger highlighting (requirement 3.1)

- [ ] 9.2. Create cross-format consistency tests
  - Test unknown values display consistency across table, JSON, HTML, Markdown (requirement 3.4)
  - Test outputs section consistency across all output formats (requirement 3.4)
  - Verify "(known after apply)" and "(sensitive value)" display consistently (requirements 1.3, 2.4)

### 10. Add Performance and Edge Case Testing
- [ ] 10.1. Create edge case tests for unknown values processing
  - Test complex nested structures with mixed known/unknown values
  - Test arrays with unknown elements
  - Test properties remaining unknown (before and after unknown) (requirement 1.7)
  - Test large plans with many unknown values within existing performance limits

- [ ] 10.2. Create edge case tests for outputs processing
  - Test sensitive outputs with unknown values (requirements 2.3, 2.4)
  - Test large output values with size limits
  - Test malformed output structures with graceful error handling
  - Test plans with only outputs changes (no resource changes)

### 11. Update Statistics and Integration Points
- [ ] 11.1. Update resource statistics calculation in `lib/plan/analyzer.go`
  - Ensure proper categorization of resource changes involving unknown values (requirement 3.3)
  - Verify unknown properties appear in "x properties changed" count for collapsible sections
  - Maintain existing statistics structure for resource changes only

- [ ] 11.2. Final integration verification
  - Test collapsible sections include unknown values in expanded details (requirement 3.2)
  - Test danger highlighting still functions with unknown values (requirement 3.1)
  - Test provider grouping works correctly with unknown values and outputs
  - Verify global expand-all control includes outputs section

## Implementation Notes

- Follow decision log priorities: implement unknown values first (phase 1), then outputs (phase 2)
- All tasks focus on code implementation, testing, and integration within development environment
- Each task builds incrementally on previous tasks with immediate testing validation
- Unknown values take precedence over standard property comparison logic (decision log requirement)
- Maintain backward compatibility - all changes are additive to existing functionality
- Use exact Terraform syntax "(known after apply)" across all output formats per decision log