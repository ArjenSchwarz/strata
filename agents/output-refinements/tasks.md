# Output Refinements Implementation Tasks

## Implementation Overview

Convert the feature design into a series of prompts for implementing the Output Refinements feature in a test-driven manner. These tasks focus on enhancing existing components rather than creating new architectural layers, following the pragmatic design approach outlined in the design document.

## Implementation Tasks

### 1. Configuration Enhancement
- [x] 1.1 Add `ShowNoOps bool` field to `PlanConfig` struct in `config/config.go` with `mapstructure:"show-no-ops"` tag
  - Implement field with proper YAML mapping for kebab-case configuration
  - Ensure default value is `false` to hide no-ops by default (Requirement 3.2)
  - Write unit tests for configuration loading and parsing

- [x] 1.2 Add `--show-no-ops` CLI flag to plan summary command in `cmd/plan_summary.go`
  - Implement boolean flag that overrides configuration file setting
  - Ensure CLI flag takes precedence over config file (Requirement 3.4)
  - Write unit tests for flag parsing and precedence behavior

### 2. Property Sorting and Sensitive Value Masking
- [x] 2.1 Enhance `analyzePropertyChanges` function in `lib/plan/analyzer.go` to implement alphabetical property sorting
  - Sort properties case-insensitively by property name within each resource (Requirement 1.1, 1.3)
  - For properties with same name, sort by path hierarchy (Requirement 1.2)  
  - Implement natural sort ordering for special characters and numbers (Requirement 1.4)
  - Write comprehensive unit tests for all sorting scenarios

- [x] 2.2 Implement immediate sensitive value masking in `compareObjects` function in `lib/plan/analyzer.go`
  - Add `maskSensitiveValue` helper function that returns "(sensitive value)" for table/HTML/Markdown formats (Requirement 5.4)
  - Mask sensitive values immediately during property extraction to ensure security by default (Requirement 5.1, 5.2)
  - Handle nested structures by masking only sensitive leaf values while preserving structure (Requirement 5.5)
  - Follow Terraform's JSON output convention for JSON format sensitive values (Requirement 5.3)
  - Write unit tests for masking behavior across all scenarios

### 3. No-Op Resource and Output Detection  
- [x] 3.1 Add `IsNoOp bool` field to `ResourceChange` struct in `lib/plan/models.go`
  - Add JSON tag with `json:"-"` to exclude from output serialization
  - Update model to support internal no-op tracking

- [x] 3.2 Add `IsNoOp bool` field to `OutputChange` struct in `lib/plan/models.go`  
  - Add JSON tag with `json:"-"` for internal use only
  - Support detection of outputs with identical before/after values

- [x] 3.3 Enhance `AnalyzePlan` function in `lib/plan/analyzer.go` to detect no-op resources and outputs
  - Mark resources with `ChangeTypeNoOp` by setting `IsNoOp = true`
  - Detect output changes where before and after values are identical using `reflect.DeepEqual` (Requirement 4.1)
  - Write unit tests for no-op detection logic

### 4. No-Op Filtering Implementation
- [x] 4.1 Implement `filterNoOps` method in `lib/plan/formatter.go` for resource filtering
  - Filter out resources where `ChangeType == ChangeTypeNoOp` when `ShowNoOps` is false (Requirement 3.2)
  - Preserve original slice when `ShowNoOps` is true
  - Write unit tests for filtering behavior

- [x] 4.2 Implement `filterNoOpOutputs` method in `lib/plan/formatter.go` for output filtering
  - Filter out outputs where `IsNoOp == true` when `ShowNoOps` is false (Requirement 4.2)
  - Hide entire output changes section when all outputs are no-ops (Requirement 4.3)
  - Include no-op outputs when `--show-no-ops` flag is enabled (Requirement 4.4)
  - Write unit tests for output filtering scenarios

- [x] 4.3 Enhance `OutputSummary` method in `lib/plan/formatter.go` to integrate no-op filtering
  - Apply filtering based on `f.config.Plan.ShowNoOps` configuration
  - Display "No changes detected" message when no actual changes exist (Requirement 3.5)
  - Ensure statistics remain unchanged and count all resources including no-ops (Requirement 3.7)
  - Write integration tests for complete filtering workflow

### 5. Resource Priority Sorting Implementation  
- [x] 5.1 Implement `sortResourcesByPriority` method in `lib/plan/formatter.go` for danger and action-based sorting
  - Sort first by `IsDangerous` flag with dangerous resources first (Requirement 2.1)
  - Sort second by action type: delete > replace > update > create (Requirement 2.2)
  - Sort third alphabetically by resource address as tiebreaker (Requirement 2.3) 
  - Apply sorting within provider groups independently when provider grouping is enabled (Requirement 2.4)
  - Write unit tests for all sorting priority scenarios

- [x] 5.2 Integrate priority sorting into `FormatPlan` method in `lib/plan/formatter.go`
  - Call `sortResourcesByPriority` after filtering but before format-specific rendering
  - Ensure sorting works correctly with existing provider grouping feature
  - Preserve existing progressive disclosure and collapsible sections functionality
  - Write unit tests for sorting integration

### 6. Enhanced ActionSortTransformer for Table Output
- [ ] 6.1 Enhance `hasDangerIndicator` method in `lib/plan/formatter.go` ActionSortTransformer
  - Improve danger detection logic to identify dangerous resources in table rows
  - Use existing danger column regex patterns for detection
  - Handle edge cases where danger indicators might be ambiguous
  - Write unit tests for danger indicator detection

- [ ] 6.2 Update `Transform` method in ActionSortTransformer to implement enhanced table sorting
  - Sort table rows first by danger indicators, then by action priority, then alphabetically
  - Maintain existing action priority mapping (delete=0, replace=1, update=2, create=3, noop=4)
  - Ensure transformation works correctly with filtered content
  - Write unit tests for enhanced table transformation

### 7. Statistics Handling for No-Ops
- [ ] 7.1 Ensure statistics calculation in `AnalyzePlan` continues to count all resources including no-ops
  - Verify that `ChangeStatistics` includes no-op resources regardless of display settings (Requirement 3.7)
  - Exclude no-op outputs from output change counts in statistics (Requirement 4.5)
  - Write unit tests to verify statistics behavior with hidden no-ops

### 8. Integration Testing and Validation
- [ ] 8.1 Create comprehensive integration tests in `test/integration/output_refinements_test.go`
  - Test complete workflow with real Terraform plan files containing sensitive values, no-ops, and mixed change types
  - Verify all output formats (table, JSON, HTML, Markdown) handle enhancements correctly
  - Test configuration precedence between CLI flags and config file settings
  - Validate backward compatibility with existing plan files and configurations

- [ ] 8.2 Create edge case tests for all enhancement scenarios
  - Test empty plans with only no-op resources and outputs
  - Test plans with mixed sensitive/non-sensitive properties in nested structures
  - Test sorting with identical resource addresses and various combinations of danger/action states
  - Test performance with large plans (1000+ resources) to ensure <5% impact

- [ ] 8.3 Update existing unit tests to accommodate enhanced behavior
  - Review and update formatter tests to work with new sorting and filtering behavior
  - Update analyzer tests to verify property sorting and sensitive masking
  - Ensure all existing functionality continues to work with enhancements

### 9. Documentation and Configuration Updates
- [ ] 9.1 Update CLI help text in `cmd/plan_summary.go` to document the new `--show-no-ops` flag
  - Provide clear explanation of flag behavior and interaction with configuration
  - Include examples of usage scenarios

- [ ] 9.2 Update default `strata.yaml` configuration file to include new option with documentation
  - Add commented example: `# show-no-ops: false  # Show no-op resources (default: false)`
  - Ensure configuration documentation reflects all available options

## Testing Strategy

Each implementation task includes specific unit tests to ensure:
- **Property Sorting**: Case-insensitive, natural ordering, path-based tiebreaking
- **Sensitive Masking**: Immediate masking, format-specific handling, nested structure support  
- **No-Op Detection**: Accurate identification of resources and outputs with no changes
- **Filtering Logic**: Correct application of show/hide logic based on configuration
- **Sorting Priority**: Proper danger > action > alphabetical ordering within provider groups
- **Configuration**: CLI flag precedence over config file settings
- **Backward Compatibility**: All existing functionality preserved

Integration tests will validate the complete feature using real Terraform plan files and ensure performance requirements are met (<5% impact on plan summary generation time).

## Implementation Notes

- All enhancements build incrementally on existing components following the pragmatic design approach
- Security is implemented by default with immediate sensitive value masking during analysis
- Each task references specific requirements from the requirements document
- Implementation maintains full backward compatibility while fixing existing sorting regressions
- Focus on test-driven development with comprehensive coverage for all new functionality