# Plan Summary Output Improvements - Implementation Tasks

## Implementation Plan

This document outlines the discrete coding tasks required to implement the plan summary output improvements feature. Each task builds incrementally on previous tasks and focuses on test-driven development with early validation.

### 1. Property Change Extraction Infrastructure

- [x] **1.1** Implement deep object comparison algorithm in analyzer.go
  - Create `compareObjects` function with recursive comparison logic for maps, slices, and primitives
  - Handle type changes, nil values, and property additions/removals
  - Add helper functions: `extractPropertyName`, `parsePath`, `isSensitive`, `extractSensitiveChild`, `extractSensitiveIndex`
  - References requirement 2 (Property Changes Formatting - property-level before/after values extraction)

- [x] **1.2** Enhance PropertyChange struct in models.go
  - Add `Action` field (string) for "add", "remove", "update" actions
  - Ensure `Before` and `After` fields can hold any value type
  - Update existing struct fields to support the new data flow
  - References requirement 2 (Property Changes Formatting - actual before/after values)

- [x] **1.3** Create comprehensive unit tests for property comparison
  - Test simple property changes (string, number, boolean)
  - Test nested object changes with proper path extraction
  - Test array modifications and length changes
  - Test property additions and removals
  - Test sensitive value handling and masking
  - References requirement 2 (Property Changes Formatting - property-level extraction validation)

### 2. Enhanced Property Analysis

- [x] **2.1** Update analyzePropertyChanges method in analyzer.go
  - Replace placeholder implementation with real deep comparison using `compareObjects`
  - Extract actual before/after values from ResourceChange.Change structure
  - Handle sensitive values using BeforeSensitive and AfterSensitive fields
  - Set proper Count field based on individual property changes
  - References requirement 2 (Property Changes Formatting - individual property counting)

- [x] **2.2** Implement performance limits for property extraction
  - Add constants: MaxPropertiesPerResource (100), MaxPropertyValueSize (10KB), MaxTotalPropertyMemory (10MB)
  - Create `enforcePropertyLimits` function to truncate excessive data
  - Update PropertyChangeAnalysis to include `Truncated` and `TotalSize` fields
  - Add size estimation logic for property values
  - References technical implementation notes (performance and memory management)

- [x] **2.3** Add unit tests for enhanced property analysis
  - Test that individual properties are counted correctly
  - Test performance limit enforcement
  - Test truncation behavior with large datasets
  - Test memory usage calculations
  - Verify proper handling of complex nested structures
  - References requirement 2 (Property Changes Formatting - individual property counting validation)

### 3. Terraform-Style Property Change Formatting

- [x] **3.1** Create custom CollapsibleFormatter for property changes in formatter.go
  - Implement `propertyChangesFormatterTerraform` function returning collapsible content
  - Create `formatPropertyChange` function with `+`, `-`, `~` prefixes for actions
  - Implement `formatValue` function handling different value types (string, map, array, primitives)
  - Handle sensitive value masking with "(sensitive value hidden)" text
  - Use go-output's `NewCollapsibleValue` API for expandable content
  - References requirement 2 (Property Changes Formatting - Terraform diff-style format)

- [x] **3.2** Implement auto-expansion logic for sensitive changes
  - Auto-expand collapsible sections when sensitive properties are detected
  - Respect global expand-all configuration from config
  - Add warning indicators for sensitive content in summary text
  - References requirement 2 (Property Changes Formatting - sensitive values handling)

- [x] **3.3** Add comprehensive formatter tests
  - Test Terraform-style formatting for add, remove, update actions
  - Test sensitive value masking in formatted output
  - Test complex value formatting (maps, arrays, nested objects)
  - Test auto-expansion behavior for sensitive changes
  - Test collapsible content creation and structure
  - References requirement 2 (Property Changes Formatting - format validation)

### 4. Empty Table Suppression Logic

- [x] **4.1** Implement empty table detection in formatter.go
  - Create `prepareResourceTableData` function that filters out no-op changes
  - Update `OutputSummary` method to check for empty table data before creating tables
  - Suppress entire table creation when no changed resources exist after filtering
  - References requirement 1 (Empty Table Suppression - no-op filtering)

- [x] **4.2** Update provider grouping logic for changed resources only
  - Create `countChangedResources` function excluding no-ops from counts
  - Update `shouldGroupByProvider` to use changed resource count for threshold comparison
  - Modify `groupResourcesByProvider` to exclude no-ops from grouping
  - Update provider count displays to show only changed resources
  - References requirement 1 (Empty Table Suppression - changed resource counting)

- [x] **4.3** Add unit tests for empty table suppression
  - Test table suppression when only no-op changes exist
  - Test provider grouping threshold calculations with changed resources only
  - Test provider count accuracy excluding no-ops
  - Test section header suppression when all tables are empty
  - References requirement 1 (Empty Table Suppression - validation of suppression logic)

### 5. Dependencies Column Removal and Cleanup

- [x] **5.1** Remove dependencies functionality from formatter.go
  - Remove `dependencies` field from `getResourceTableSchema()` function
  - Remove dependencies row data from `prepareResourceTableData()` function
  - Remove `dependenciesFormatterDirect()` function entirely
  - Update table column definitions to exclude dependencies
  - References requirement 4 removal (Dependencies functionality removed per decision log)

- [x] **5.2** Clean up dependency-related data structures
  - Remove `DependencyInfo` struct from models.go if not used elsewhere
  - Remove any dependency-related analyzer code that's no longer needed
  - Update tests to remove dependency-related test cases
  - References requirement 4 removal (Dependencies cleanup)

### 6. Risk-Based Sorting Implementation

- [x] **6.1** Re-enable ActionSortTransformer in rendering pipeline
  - Add ActionSortTransformer back to stdoutOptions in `OutputSummary` method
  - Ensure transformer is added after emoji transformer but before output
  - Verify transformer only applies to supported formats (table, markdown, HTML, CSV)
  - References requirement 5 (Custom Risk-Based Sorting - transformer re-enablement)

- [x] **6.2** Add integration tests for risk-based sorting
  - Test sorting order: dangerous items first, then by action type (delete, replace, update, add)
  - Test alphabetical tertiary sorting by resource address
  - Test sorting consistency across different output formats
  - Test that no-op changes remain hidden from output
  - References requirement 5 (Custom Risk-Based Sorting - sort order validation)

### 7. Integration and End-to-End Testing

- [x] **7.1** Create end-to-end tests with sample data
  - Test all improvements using `samples/danger-sample.json`
  - Test empty table suppression with `samples/nochange-sample.json`
  - Create new test data for complex property change scenarios
  - Verify output improvements across all supported formats (table, JSON, HTML, Markdown)
  - References requirement 6 (General Applicability - cross-format consistency)

- [x] **7.2** Add performance and memory tests
  - Test with large plans (1000+ resources) to verify performance limits work
  - Test with resources having many property changes
  - Verify memory usage stays within defined limits
  - Test truncation behavior under load
  - References technical implementation notes (performance testing)

- [x] **7.3** Validate backward compatibility
  - Ensure JSON output structure remains unchanged (only values in property_changes field differ)
  - Verify table column names and order remain consistent
  - Test that existing parsers can still process the output
  - References design document (Backward Compatibility section)

### 8. Final Integration and Validation

- [ ] **8.1** Run comprehensive test suite
  - Execute all unit tests with `go test ./...`
  - Run integration tests with all sample files
  - Verify no regressions in existing functionality
  - Test all output formats work correctly with improvements

- [ ] **8.2** Verify all requirements are implemented
  - Empty table suppression working across all formats
  - Property changes displayed in Terraform diff format
  - Property changes column shows actual changes, not emojis
  - Risk-based sorting prioritizes dangerous changes
  - All improvements apply universally to plan summary commands
  - References all requirements from requirements.md

## Notes

- Each task builds incrementally on previous tasks
- All tasks focus on writing, modifying, or testing code
- Test-driven development is prioritized with comprehensive unit and integration tests
- Performance and memory considerations are built into the implementation
- The implementation follows the decisions made in the DECISION_LOG.md