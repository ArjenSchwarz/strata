# Simplified Plan Rendering Implementation Tasks

## Overview

This implementation plan converts the simplified plan rendering design into discrete, actionable coding tasks. The primary goal is to fix the multi-table rendering bug in markdown format while simplifying the rendering architecture through targeted changes following the proven go-output v2 pattern.

## Implementation Tasks

### 1. Create Comprehensive Test Suite for Multi-Table Rendering
- [x] 1.1. Create test that validates all three tables (Plan Information, Summary Statistics, Resource Changes) render correctly in markdown format (Requirement 10.1)
  - Write `TestMarkdownMultiTableRendering` function in `lib/plan/formatter_test.go`
  - Test should verify presence of all three table headers: "### Plan Information", "### Summary Statistics", "### Resource Changes"
  - Validate table structure with expected column headers for each table
  - Use test plan summary data that includes all table types
- [x] 1.2. Create format compatibility tests for all supported output formats (Requirement 10.2)
  - Write `TestAllFormatCompatibility` that tests table, json, html, markdown formats (corrected from csv)
  - Ensure each format renders without errors
  - Verify output contains expected content structure for each format
- [x] 1.3. Create collapsible content functionality tests (Requirement 10.3)
  - Write `TestCollapsibleContentInSupportedFormats` to verify expandable sections work
  - Test that `output.NewCollapsibleValue()` objects render correctly
  - Validate auto-expansion behavior for high-risk changes
- [x] 1.4. Create provider grouping tests with various thresholds (Requirement 10.4)
  - Write `TestProviderGroupingWithThresholds` to test grouping behavior
  - Test scenarios: below threshold (no grouping), above threshold (grouping enabled)
  - Verify existing threshold check behavior is maintained (Decision #7)
- [x] 1.5. Create edge case tests for empty plans and malformed data (Requirement 10.5)
  - Write `TestEdgeCases` for nil plan summaries, empty resource changes, missing data
  - Test special character handling in resource names and values (Requirement 9.5)
  - Verify graceful error handling without crashes (Requirement 9.2)

### 2. Enable Multi-Table Rendering (Primary Bug Fix)
- [x] 2.1. Remove artificial table disabling in formatter.go lines 189-191 (Requirement 3.2, Decision #13)
  - Remove the commented-out Plan Information and Summary Statistics table creation
  - Remove the TODO comment about "go-output v2 multi-table rendering bug"
  - Update comments to reflect that all tables are now enabled
- [x] 2.2. Implement Plan Information table using NewTableContent pattern (Requirement 1.1, Decision #1)
  - Modify `OutputSummary` function to create Plan Information table using `output.NewTableContent()`
  - Use existing `createPlanInfoDataV2()` method for data preparation
  - Apply schema definition with `output.WithKeys("Plan File", "Version", "Workspace", "Backend", "Created")`
  - Add conservative error handling that logs warnings but continues operation (Decision #15)
- [x] 2.3. Implement Summary Statistics table using NewTableContent pattern (Requirement 1.1)
  - Create Summary Statistics table in `OutputSummary` using `output.NewTableContent()`
  - Use existing `createStatisticsSummaryDataV2()` method for data preparation  
  - Apply schema with keys: "TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED"
  - Implement same conservative error handling pattern
- [x] 2.4. Convert Resource Changes table to use NewTableContent consistently (Requirement 1.2)
  - Modify existing Resource Changes table creation to use `output.NewTableContent()` instead of mixed methods
  - Keep existing data preparation logic (`prepareResourceTableData`) unchanged
  - Apply existing schema definition using `output.WithSchema()`
  - Maintain existing collapsible content functionality
- [x] 2.5. Implement unified document building using output.New().AddContent() pattern (Requirement 3.1)
  - Replace current document building with `output.New().AddContent(table1).AddContent(table2).AddContent(table3).Build()`
  - Ensure tables are added in consistent order: Plan Information, Summary Statistics, Resource Changes
  - Maintain same table order across all output formats (Requirement 3.4)

### 3. Simplify Format Handling Logic
- [x] 3.1. Remove markdown-specific format branching (Requirement 5.1, Decision #3)
  - Remove the conditional logic in lines 254-264 that treats markdown differently
  - Eliminate distinction between `getFormatFromConfig()` and `getCollapsibleFormatFromConfig()` (Requirement 2.2)
  - Use single code path: `stdoutFormat := f.getFormatFromConfig(outputConfig.Format)`
- [x] 3.2. Consolidate format configuration methods (Requirement 2.3, Decision #3)
  - Evaluate if `getCollapsibleFormatFromConfig()` method can be removed or consolidated
  - Ensure all formats use go-output's standard format definitions (output.Markdown, output.Table, etc.)
  - Remove format-specific branching logic throughout the formatter
- [x] 3.3. Update format handling to be delegated to go-output (Requirement 2.1, Decision #9)
  - Remove any Strata-specific format handling logic
  - Let go-output handle format capabilities through metadata rather than custom logic (Requirement 2.5)
  - Ensure consistent application of transformers across all formats (Requirement 5.4)

### 4. Code Consolidation and Cleanup
- [x] 4.1. Remove duplicate property formatter functions (Requirement 8.1, Decision #16)
  - Remove `propertyChangesFormatter()` function and keep `propertyChangesFormatterDirect()`
  - Verify that `propertyChangesFormatterDirect()` correctly returns `NewCollapsibleValue` objects
  - Update all references to use the Direct version
- [x] 4.2. Remove duplicate dependency formatter functions (Requirement 8.1)
  - Remove `dependenciesFormatter()` function and keep `dependenciesFormatterDirect()`
  - Verify the Direct version maintains existing functionality
  - Update all function calls to use the retained version
- [x] 4.3. Clean up unused code paths and update comments (Requirement 8.4)
  - Remove any unused format-specific code paths
  - Add architectural decision comments for key changes
  - Document non-obvious go-output API usage patterns
  - Update function documentation to reflect consolidated approach

### 5. Provider Grouping Integration
- [x] 5.1. Ensure provider grouping works with new table creation pattern (Requirement 6.2, 6.3)
  - Verify that provider grouping logic continues to work with `NewTableContent` pattern
  - Test that grouped resources render correctly in all formats using consistent section creation
  - Maintain existing threshold check behavior (Decision #7)
- [x] 5.2. Preserve auto-expansion behavior for high-risk changes within groups (Requirement 6.4)
  - Ensure that high-risk changes within provider groups still auto-expand
  - Verify that `output.NewCollapsibleValue()` objects work correctly within grouped content
  - Test that danger highlighting continues to function within provider groups

### 6. Validation and Integration Testing
- [ ] 6.1. Run comprehensive regression tests to ensure no functionality loss (Requirement 7.1-7.6)
  - Execute full test suite to verify all existing tests pass unchanged
  - Validate that collapsible/expandable content continues working (Requirement 7.1)
  - Verify danger highlighting and auto-expansion for sensitive resources (Requirement 7.2)
  - Test `--expand-all` flag and configuration option (Requirement 7.3)
  - Confirm backward compatibility with existing configuration files (Requirement 7.4)
- [ ] 6.2. Performance validation to prevent regression (Requirement 8.5, Decision #5)
  - Create benchmark tests comparing rendering times before/after changes
  - Monitor memory consumption during rendering of large plans
  - Test with existing large plan fixtures to ensure no performance degradation
  - Document any performance improvements gained from simplification
- [ ] 6.3. Manual testing across all output formats (Requirement 10.2)
  - Test real Terraform plan files with all supported formats: table, json, markdown, html, csv
  - Verify that all three tables appear correctly in each format
  - Test edge cases: empty plans, large plans, plans with special characters
  - Validate that nested collapsible sections work (provider groups with expandable property changes)

### 7. Error Handling Enhancement
- [ ] 7.1. Implement conservative error handling for table creation failures (Requirement 9.3, Decision #15)
  - Add error handling that logs warnings but continues operation when individual tables fail to create
  - Ensure that failure of one table doesn't prevent other tables from rendering
  - Display meaningful messages when no resource changes exist (Requirement 9.1)
- [ ] 7.2. Add proper error handling for edge cases (Requirement 9.2, 9.4)
  - Handle empty or nil plan summaries without crashing
  - Handle resources with missing or malformed data by showing available information
  - Ensure proper escaping of special characters in resource names and values (Requirement 9.5)

## Success Validation

After completing all tasks, verify:
1. All three tables (Plan Information, Summary Statistics, Resource Changes) render correctly in markdown format
2. All existing tests pass without modification  
3. No performance regression in rendering times or memory usage
4. All output formats continue to work identically except for the multi-table fix
5. Code duplication is reduced while maintaining all functionality

## Implementation Notes

- Follow the specific pattern from go-output v2's collapsible-tables example
- Maintain conservative approach - preserve existing functionality while fixing the core bug
- Each task builds incrementally on previous tasks
- Test-driven development approach with validation at each step
- Focus on targeted fix rather than complete architectural overhaul (Decision #12)