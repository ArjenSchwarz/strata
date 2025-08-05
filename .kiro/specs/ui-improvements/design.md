# Design Document: UI Improvements

## Overview

This design document outlines the implementation approach for enhancing the UI clarity of the Strata Terraform plan analysis tool. The improvements focus on making the output more readable, providing better visibility into high-risk changes, and adding support for markdown output format.

The changes will be implemented in the existing codebase, primarily in the `lib/plan` package, with modifications to the formatter and analyzer components. The design aims to maintain compatibility with the current architecture while enhancing the user experience.

## Architecture

The UI improvements will be implemented within the existing architecture of the Strata tool:

1. **Command Layer** (`cmd/plan_summary.go`): 
   - Add support for markdown output format
   - Update flag handling for new features

2. **Library Layer** (`lib/plan`):
   - Enhance the `Formatter` to support two-row table layout for plan information
   - Modify the `Analyzer` to calculate high-risk changes
   - Update the output formatting to support markdown

3. **Configuration Layer** (`config/config.go`):
   - Add configuration options for new features

## Components and Interfaces

### Plan Information Display

The plan information display will be modified to use a two-row table layout inspired by the fog tool. This will be implemented in the `formatPlanInfo` method of the `Formatter` struct.

```go
// Current vertical layout:
// Key: Value
// Key: Value
// Key: Value

// New two-row table layout:
// | Key1      | Key2      | Key3      |
// | Value1    | Value2    | Value3    |
```

Key changes:
- Modify the `formatPlanInfo` method to arrange items in a two-row table with keys in the first row and values in the second row
- Add a title above the table (e.g., "Plan Information")
- Rename "Terraform Version" to "Version" in the display
- Remove the dry run indicator from the display

### Risk Visibility in Summary

The statistics summary will be enhanced to include a "High Risk" column that counts the number of sensitive items with a danger flag. This will be implemented by:

1. Updating the `ChangeStatistics` struct in `models.go` to include a `HighRisk` field
2. Modifying the `calculateStatistics` method in `analyzer.go` to count high-risk changes
3. Updating the `formatStatisticsSummary` method in `formatter.go` to display the high-risk count

### Sensitive Resource Changes Display

The formatter will be modified to always show sensitive resource changes, even when detailed output is disabled:

1. Update the `OutputSummary` method in `formatter.go` to conditionally show sensitive changes
2. Add a new method `formatSensitiveResourceChanges` to display only sensitive changes when `showDetails` is false

### Markdown Output Support

Support for markdown output will be added using the existing go-output library:

1. Add "markdown" as a supported output format in the command flags
2. Update the go-output library integration to handle markdown formatting
3. Ensure consistent formatting between table and markdown outputs

## Data Models

### Changes to Existing Models

The `ChangeStatistics` struct in `models.go` will be updated to include a high-risk count:

```go
type ChangeStatistics struct {
    ToAdd        int `json:"to_add"`
    ToChange     int `json:"to_change"`
    ToDestroy    int `json:"to_destroy"`
    Replacements int `json:"replacements"`
    Conditionals int `json:"conditionals"`
    HighRisk     int `json:"high_risk"`    // New field for high-risk changes
    Total        int `json:"total"`
}
```

### Configuration Updates

The `PlanConfig` struct in `config/config.go` will be updated to include options for the new features:

```go
type PlanConfig struct {
    DangerThreshold         int    `mapstructure:"danger-threshold"`
    ShowDetails             bool   `mapstructure:"show-details"`
    HighlightDangers        bool   `mapstructure:"highlight-dangers"`
    ShowStatisticsSummary   bool   `mapstructure:"show-statistics-summary"`
    StatisticsSummaryFormat string `mapstructure:"statistics-summary-format"`
    AlwaysShowSensitive     bool   `mapstructure:"always-show-sensitive"` // New field
}
```

## Implementation Details

### Two-Row Table Plan Information Layout

The `formatPlanInfo` method will be updated to use a two-row table layout:

1. Create a table with two rows - the first row containing the keys and the second row containing the values
2. Add a title above the table (e.g., "Plan Information")
3. Modify the keys to be more concise (e.g., "Terraform Version" → "Version")
4. Remove the "Dry Run" entry from the display
5. Ensure proper alignment and spacing between columns

### High-Risk Column in Summary

The implementation will:

1. Update the `calculateStatistics` method to count resources that have both `IsDangerous` set to true and are in the sensitive resources list
2. Add the "HIGH RISK" column to the statistics summary output

### Always Show Sensitive Resource Changes

When `showDetails` is false but sensitive changes exist:

1. Create a new method `formatSensitiveResourceChanges` that filters the resource changes to only include those with `IsDangerous` set to true
2. Call this method from `OutputSummary` when `showDetails` is false but sensitive changes exist

### Markdown Output Support

The implementation will:

1. Add "markdown" as a supported output format in the command flags
2. Update the `OutputSettings` to handle markdown format
3. Ensure the go-output library correctly formats tables in markdown

## Error Handling

The existing error handling approach will be maintained:

1. Functions will return errors when operations fail
2. Errors will be wrapped with context using `fmt.Errorf("failed to X: %w", err)`
3. The command layer will handle and display errors to the user

## Testing Strategy

The following tests will be added or updated:

1. **Unit Tests**:
   - Test the two-row table layout of plan information
   - Test the calculation of high-risk changes
   - Test the display of sensitive resources when details are disabled
   - Test markdown output formatting

2. **Integration Tests**:
   - Test the end-to-end flow with different output formats
   - Test with various combinations of flags and settings

3. **Manual Testing**:
   - Verify the readability of the two-row table layout
   - Confirm that high-risk changes are correctly identified
   - Ensure sensitive resources are always displayed
   - Validate markdown output in different environments

## Conclusion

These UI improvements will enhance the usability of the Strata tool by making the output more readable and providing better visibility into high-risk changes. The addition of markdown output support will also make it easier to include plan summaries in documentation and pull requests.

The implementation will maintain compatibility with the existing architecture while adding new features that address the requirements specified in the requirements document.