# Enhanced Summary Visualization - Implementation Tasks

This document provides an actionable implementation plan for the Enhanced Summary Visualization feature, broken down into discrete coding tasks that build incrementally on each other.

## Prerequisites

- Requirements document: `agents/enhanced-summary-visualization/requirements.md`
- Design document: `agents/enhanced-summary-visualization/design.md`
- Existing codebase understanding from `lib/plan/` modules

## Implementation Tasks

### 1. Extend Data Models and Configuration

- [x] 1.1 Add new fields to ResourceChange struct in `lib/plan/models.go`
  - Add `Provider string` field for storing extracted provider name
  - Add `TopChanges []string` field for first 3 changed properties
  - Add `ReplacementHints []string` field for human-readable replacement reasons
  - References requirement: Enhanced change context (1.2) - show replacement reasons and property changes

- [x] 1.2 Extend PlanConfig struct in `config/config.go`
  - Add `GroupByProvider bool` field with mapstructure tag "group-by-provider"
  - Add `GroupingThreshold int` field with mapstructure tag "grouping-threshold"
  - Add `ShowContext bool` field with mapstructure tag "show-context"
  - References requirement: Optional resource grouping (1.1) - configurable grouping and threshold

- [x] 1.3 Write unit tests for data model extensions
  - Test ResourceChange struct serialization/deserialization with new fields
  - Test PlanConfig loading from YAML with new configuration options
  - Verify default values and field validation

### 2. Implement Provider Extraction and Caching

- [x] 2.1 Add provider extraction function to `lib/plan/analyzer.go`
  - Implement `extractProvider(resourceType string) string` function using string split
  - Handle edge cases for malformed resource types
  - References requirement: Optional resource grouping (1.1) - group by provider

- [x] 2.2 Add provider extraction caching for performance
  - Implement thread-safe cache using sync.Map for provider extraction results
  - Add cache hit/miss tracking for testing
  - References design: Performance optimizations - provider extraction caching

- [x] 2.3 Write unit tests for provider extraction
  - Test extraction for AWS resources (aws_s3_bucket → "aws")
  - Test extraction for Azure resources (azurerm_virtual_machine → "azurerm")
  - Test extraction for Google resources (google_compute_instance → "google")
  - Test edge cases and malformed resource types

### 3. Enhance Resource Change Analysis

- [ ] 3.1 Implement context extraction for replacement reasons in `lib/plan/analyzer.go`
  - Enhance existing `analyzeResourceChanges()` to populate `ReplacementHints` field
  - Extract human-readable reasons from Terraform's `ReplacePaths` data
  - Ensure replacement hints are always populated regardless of `ShowContext` setting
  - References requirement: Enhanced change context (1.2) - show all replacement reasons

- [ ] 3.2 Implement property change detection for updates
  - Add `getTopChangedProperties(change *tfjson.ResourceChange, limit int) []string` function
  - Compare before/after states using existing `equals()` function
  - Return first 3 property names that changed (not values for security)
  - Only populate when `ShowContext` is enabled
  - References requirement: Enhanced change context (1.2) - show first 3 properties changing

- [ ] 3.3 Enhance existing danger reason logic
  - Improve existing `DangerReason` field with more descriptive explanations
  - Add specific reasons for sensitive resource replacements vs deletions
  - References requirement: Risk highlighting (1.3) - brief explanations for risky changes

- [ ] 3.4 Write unit tests for context extraction
  - Test replacement reason extraction from various ReplacePaths scenarios
  - Test property change detection with different before/after state combinations
  - Test that replacement hints are always shown but property changes respect ShowContext flag
  - Test enhanced danger reason explanations

### 4. Implement Smart Resource Grouping

- [ ] 4.1 Add grouping logic function to `lib/plan/analyzer.go`
  - Implement `GroupResourcesByProvider(changes []ResourceChange) (map[string][]ResourceChange, bool)`
  - Check resource count against `GroupingThreshold` configuration
  - Detect provider diversity and skip grouping if all resources are from same provider
  - Return grouped resources and boolean indicating if grouping should be applied
  - References requirement: Optional resource grouping (1.1) - smart grouping hierarchy

- [ ] 4.2 Write comprehensive unit tests for grouping logic
  - Test threshold behavior with resources below/above threshold
  - Test single-provider scenarios (should not group)
  - Test multi-provider scenarios (should group)
  - Test edge cases with empty resource lists
  - Verify correct provider extraction and grouping

### 5. Extend Formatter for Grouped Display

- [ ] 5.1 Modify existing `formatResourceChangesTable()` in `lib/plan/formatter.go`
  - Add conditional logic to check `GroupByProvider` configuration
  - Call grouping logic and handle both grouped and ungrouped display paths
  - Maintain backward compatibility with existing ungrouped display
  - References requirement: Optional resource grouping (1.1) - optional grouping

- [ ] 5.2 Implement `formatGroupedByProvider()` function
  - Create separate go-output sections for each provider group
  - Use library's section capabilities for proper visual separation
  - Include resource counts for each provider group
  - References requirement: Optional resource grouping (1.1) - display resource counts

- [ ] 5.3 Enhance resource formatting to include context
  - Modify resource data preparation to include replacement hints (always)
  - Include property changes only when `ShowContext` is enabled
  - Ensure proper display formatting for new context fields
  - References requirement: Enhanced change context (1.2) - contextual information display

- [ ] 5.4 Write unit tests for formatter enhancements
  - Test grouped vs ungrouped output formatting
  - Test context display with ShowContext enabled/disabled
  - Test replacement hints are always displayed
  - Mock go-output library calls to verify correct section creation

### 6. Integration and Command Line Interface

- [ ] 6.1 Add new configuration flags to `cmd/plan_summary.go`
  - Add `--group-by-provider` boolean flag with viper binding
  - Add `--grouping-threshold` integer flag with default value 10
  - Add `--show-context` boolean flag with viper binding
  - Update command help text with new flag descriptions
  - References requirement: Configuration - YAML config and CLI flags

- [ ] 6.2 Update default configuration file template
  - Add new configuration section to `strata.yaml` with default values
  - Document configuration options with inline comments
  - Ensure backward compatibility with existing configurations

- [ ] 6.3 Write integration tests for complete flow
  - Create test fixtures with multi-provider Terraform plans
  - Test end-to-end flow from plan parsing to formatted output
  - Verify correct grouping behavior with different configurations
  - Test both CLI flags and configuration file options

### 7. Testing and Validation

- [ ] 7.1 Create comprehensive test fixtures
  - `testdata/single_provider_plan.json` - Plan with only AWS resources (should not group)
  - `testdata/multi_provider_plan.json` - Plan with AWS, Azure, GCP resources (should group)
  - `testdata/replacement_plan.json` - Plan with various replacement scenarios
  - `testdata/small_plan.json` - Plan below grouping threshold

- [ ] 7.2 Write end-to-end tests using test fixtures
  - Test complete analysis and formatting pipeline
  - Verify output matches expected format for each scenario
  - Test configuration variations (grouping on/off, context on/off)
  - Ensure no regressions in existing functionality

- [ ] 7.3 Add error handling and edge case tests
  - Test behavior with malformed Terraform plans
  - Test configuration loading with invalid values
  - Test graceful degradation when grouping fails
  - Verify error messages are user-friendly

### 8. Performance Optimization Implementation

- [ ] 8.1 Implement parallel processing for large resource lists
  - Add worker pool pattern for processing resources concurrently
  - Implement for both context extraction and formatting phases
  - Include configuration option to control worker count
  - References design: Performance optimizations - parallel processing

- [ ] 8.2 Add memory streaming for large outputs
  - Implement buffered writing for large formatted outputs
  - Stream output chunks instead of building complete output in memory
  - Add memory usage monitoring for testing
  - References design: Performance optimizations - memory streaming

- [ ] 8.3 Write performance tests
  - Create large test fixtures (100+ resources)
  - Benchmark performance improvements with parallel processing
  - Test memory usage with streaming vs non-streaming approaches
  - Verify performance scales appropriately with resource count

## Task Dependencies

- Tasks 1.x must be completed before all other tasks (foundation)
- Tasks 2.x and 3.x can be completed in parallel (independent enhancements)
- Task 4.x depends on completion of Task 2.x (grouping needs provider extraction)
- Task 5.x depends on completion of Tasks 3.x and 4.x (formatting needs context and grouping)
- Task 6.x depends on completion of Task 5.x (CLI needs working formatter)
- Tasks 7.x can begin after Task 6.x (integration testing)
- Task 8.x can be completed in parallel with Task 7.x (performance optimization)

## Requirements Coverage

All tasks above ensure complete coverage of the three main requirements:

1. **Optional Resource Grouping** - Covered by tasks 1.2, 2.x, 4.x, 5.1-5.2, 6.1
2. **Enhanced Change Context** - Covered by tasks 1.1, 3.1-3.2, 5.3, 6.1
3. **Risk Highlighting** - Covered by tasks 3.3, 5.3

The implementation plan follows test-driven development principles and ensures incremental progress with early validation of core functionality.