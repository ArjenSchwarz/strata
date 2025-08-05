# Implementation Tasks

## Task 1: Update Data Models

- [x] 1.1 Update `ChangeStatistics` struct in `models.go` to include `HighRisk` field
  - Add a new field to track high-risk changes (sensitive items with danger flag)
  - _Requirements: 2.1_

- [x] 1.2 Update `PlanConfig` struct in `config/config.go` to include `AlwaysShowSensitive` field
  - Add configuration option to control showing sensitive resources when details are disabled
  - _Requirements: 3.1, 3.2_

## Task 2: Implement Horizontal Plan Information Layout

- [x] 2.1 Modify `formatPlanInfo` method in `formatter.go` to use horizontal layout
  - Implement two-row table layout with keys in first row and values in second row
  - _Requirements: 1.1_

- [x] 2.2 Rename "Terraform Version" to "Version" in the display
  - Update the key name in the plan information output
  - _Requirements: 1.2_

- [x] 2.3 Remove the dry run indicator from the display
  - Remove the "Dry Run" entry from the plan information output
  - _Requirements: 1.3_

## Task 3: Add High-Risk Column to Statistics Summary

- [x] 3.1 Update `calculateStatistics` method in `analyzer.go` to count high-risk changes
  - Count resources that have both `IsDangerous` set to true and are in the sensitive resources list
  - _Requirements: 2.1, 2.2_

- [x] 3.2 Modify `formatStatisticsSummary` method in `formatter.go` to display high-risk count
  - Add "HIGH RISK" column to the statistics summary output
  - _Requirements: 2.1_

## Task 4: Implement Always Show Sensitive Resource Changes

- [x] 4.1 Create new method `formatSensitiveResourceChanges` in `formatter.go`
  - Implement method to filter and display only sensitive resource changes
  - _Requirements: 3.1, 3.2_

- [x] 4.2 Update `OutputSummary` method to conditionally show sensitive changes when details are disabled
  - Check config and call the new method when appropriate
  - _Requirements: 3.1, 3.2_

## Task 5: Add Markdown Output Support

- [x] 5.1 Add "markdown" as a supported output format in `plan_summary.go` command flags
  - Update the output format flag to include markdown option
  - _Requirements: 4.1_

- [x] 5.2 Update the go-output library integration to handle markdown formatting
  - Ensure the OutputSettings are properly configured for markdown output
  - _Requirements: 4.1, 4.2_

## Task 6: Update Tests

- [x] 6.1 Add unit tests for horizontal layout of plan information
  - Test the two-row table layout implementation
  - _Requirements: 1.1_

- [x] 6.2 Add unit tests for calculation of high-risk changes
  - Test the updated `calculateStatistics` method
  - _Requirements: 2.1, 2.2_

- [x] 6.3 Add unit tests for display of sensitive resources when details are disabled
  - Test the `formatSensitiveResourceChanges` method
  - _Requirements: 3.1, 3.2_

- [x] 6.4 Add unit tests for markdown output formatting
  - Test the markdown output generation
  - _Requirements: 4.1, 4.2_

## Task 7: Update Documentation

- [x] 7.1 Update README.md with new features
  - Document the UI improvements and new features
  - _Requirements: All_

- [x] 7.2 Add examples of markdown output usage
  - Include examples of using the markdown output format
  - _Requirements: 4.1, 4.2_

- [x] 7.3 Document the high-risk column in the statistics summary
  - Explain the purpose and calculation of the high-risk column
  - _Requirements: 2.1, 2.2_

- [x] 7.4 Document the always-show-sensitive feature
  - Explain how to use the always-show-sensitive feature
  - _Requirements: 3.1, 3.2_