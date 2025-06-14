# Fog-Inspired Changeset Display Implementation Plan

## Project Overview

This document outlines the implementation plan for adding the first two core functionalities from fog's changeset description feature to strata. The goal is to enhance the Terraform plan summary with comprehensive plan information display and an improved resource changes table, while maintaining the existing statistics functionality.

### Target Functionality

Based on fog's changeset description feature, we aim to implement:

1. **Plan Information Display** - Basic plan context information (similar to fog's `printBasicStackInfo`)
2. **Enhanced Resource Changes Table** - Detailed resource changes with physical IDs, replacement indicators, and module information (similar to fog's `printChangeset`)
3. **Enhanced Statistics Summary** - Horizontal summary table matching fog's format with TOTAL, ADDED, REMOVED, MODIFIED, REPLACEMENTS, CONDITIONALS columns

### Current State Assessment

**✅ Already Implemented:**
- Basic CLI structure with `plan summary` command
- Terraform plan parsing using hashicorp/terraform-json
- Resource change analysis and categorisation
- Basic statistical summaries
- Multiple output formats using go-output library
- Danger detection for destructive changes

**❌ Missing Core Functionalities:**
- Plan context information display
- Physical resource ID tracking
- Replacement necessity analysis (True/Conditional/Never)
- Module hierarchy information
- Horizontal statistics summary table format
- Enhanced resource changes table structure

---

## Phase 1: Plan Information Display

### Objective
Add comprehensive plan context information similar to fog's `printBasicStackInfo` functionality.

### Target Output Format
```
Plan Information
================
Plan File: terraform.tfplan
Terraform Version: 1.6.0
Workspace: production
Backend: s3 (bucket: my-terraform-state)
Created: 2025-05-25 23:25:28
Dry Run: No
```

### Implementation Tasks

#### 1.1 Extend Data Models
- [x] **Update `PlanSummary` struct** in `lib/plan/models.go`:
  - [x] Add `PlanFile` field (string)
  - [x] Add `Workspace` field (string)
  - [x] Add `Backend` field (BackendInfo struct)
  - [x] Add `CreatedAt` field (time.Time)
  - [x] Add `IsDryRun` field (bool)

- [x] **Create `BackendInfo` struct** in `lib/plan/models.go`:
  - [x] Add `Type` field (string) - e.g., "s3", "local", "remote"
  - [x] Add `Location` field (string) - bucket name, file path, etc.
  - [x] Add `Config` field (map[string]interface{}) - additional backend config

#### 1.2 Enhance Plan Parser
- [x] **Update `Parser.LoadPlan()` method** in `lib/plan/parser.go`:
  - [x] Extract workspace information from plan
  - [x] Parse backend configuration from plan metadata
  - [x] Capture plan file creation timestamp
  - [x] Detect dry run mode from plan context

- [x] **Add helper methods** in `lib/plan/parser.go`:
  - [x] `extractWorkspaceInfo(plan *tfjson.Plan) string`
  - [x] `extractBackendInfo(plan *tfjson.Plan) BackendInfo`
  - [x] `getPlanFileInfo(filePath string) (time.Time, error)`

#### 1.3 Update Analyzer
- [x] **Enhance `GenerateSummary()` method** in `lib/plan/analyzer.go`:
  - [x] Include plan file path in summary
  - [x] Add workspace and backend information
  - [x] Set creation timestamp and dry run status

#### 1.4 Update Formatter
- [x] **Add `formatPlanInfo()` method** in `lib/plan/formatter.go`:
  - [x] Create plan information section using go-output
  - [x] Format timestamp in user-friendly format
  - [x] Handle missing/optional information gracefully

- [x] **Update `OutputSummary()` method** in `lib/plan/formatter.go`:
  - [x] Add plan information section before statistics
  - [x] Ensure proper spacing between sections

#### 1.5 Testing
- [x] **Unit tests** for plan information extraction:
  - [x] Test workspace detection
  - [x] Test backend configuration parsing
  - [x] Test timestamp handling
  - [x] Test dry run detection

- [x] **Integration tests** for plan information display:
  - [x] Test complete plan information output
  - [x] Test with different backend types
  - [x] Test with missing information

### Success Criteria
- [x] Plan information section displays before statistics
- [x] All plan context fields are populated correctly
- [x] Output format is consistent with go-output styling
- [x] Missing information is handled gracefully
- [x] Tests pass with 100% coverage for new functionality

---

## Phase 2: Enhanced Statistics Summary Table

### Objective
Redesign the statistics table to match fog's horizontal format with TOTAL, ADDED, REMOVED, MODIFIED, REPLACEMENTS, CONDITIONALS columns.

### Target Output Format
```
Summary for terraform.tfplan
============================
TOTAL | ADDED | REMOVED | MODIFIED | REPLACEMENTS | CONDITIONALS
19    | 19    | 0       | 0        | 0            | 0
```

### Implementation Tasks

#### 2.1 Enhance Data Models
- [x] **Update `ChangeStatistics` struct** in `lib/plan/models.go`:
  - [x] Add `Conditionals` field (int) - for conditional replacements
  - [x] Rename `ToReplace` to `Replacements` for clarity
  - [x] Add documentation for each field

- [x] **Add replacement analysis types** in `lib/plan/models.go`:
  - [x] Create `ReplacementType` enum (Never, Conditional, Always)
  - [x] Update `ResourceChange` struct to include `ReplacementType` field

#### 2.2 Enhance Change Analysis
- [x] **Update `analyzeResourceChanges()` method** in `lib/plan/analyzer.go`:
  - [x] Parse `ReplacePaths` field from Terraform plan
  - [x] Determine replacement necessity (Never/Conditional/Always)
  - [x] Set `ReplacementType` for each resource change

- [x] **Update `calculateStatistics()` method** in `lib/plan/analyzer.go`:
  - [x] Count definite replacements (ReplacementType: Always)
  - [x] Count conditional replacements (ReplacementType: Conditional)
  - [x] Separate replacement counts from regular changes

- [x] **Add helper methods** in `lib/plan/analyzer.go`:
  - [x] `analyzeReplacementNecessity(change *tfjson.ResourceChange) ReplacementType`
  - [x] `isConditionalReplacement(change *tfjson.ResourceChange) bool`

#### 2.3 Update Formatter
- [x] **Create `formatStatisticsSummary()` method** in `lib/plan/formatter.go`:
  - [x] Generate horizontal statistics table using go-output
  - [x] Include plan name in summary title
  - [x] Format numbers with proper alignment

- [x] **Update `OutputSummary()` method** in `lib/plan/formatter.go`:
  - [x] Replace current statistics display with new horizontal format
  - [x] Ensure proper section separation
  - [x] Maintain compatibility with all output formats (table, json, html)

#### 2.4 Configuration Updates
- [x] **Add statistics display options** in `config/config.go`:
  - [x] Add `ShowStatisticsSummary` bool field
  - [x] Add `StatisticsSummaryFormat` string field (horizontal/vertical)

- [x] **Update command flags** in `cmd/plan_summary.go`:
  - [x] Add `--stats-format` flag for summary format control
  - [x] Update help text and examples

#### 2.5 Testing
- [x] **Unit tests** for replacement analysis:
  - [x] Test ReplacePaths parsing
  - [x] Test replacement type determination
  - [x] Test statistics calculation with replacements

- [x] **Integration tests** for statistics display:
  - [x] Test horizontal table format
  - [x] Test with various change combinations
  - [x] Test output format compatibility

### Success Criteria
- [x] Statistics table displays in horizontal format matching fog's design
- [x] REPLACEMENTS and CONDITIONALS columns show accurate counts
- [x] Table formatting is consistent with go-output styling
- [x] All output formats (table, json, html) support new statistics
- [x] Backward compatibility maintained for existing functionality

---

## Phase 3: Enhanced Resource Changes Table

### Objective
Upgrade the resource changes table to include physical resource IDs, replacement indicators, and module information, similar to fog's detailed changeset display.

### Target Output Format
```
Resource Changes
================
ACTION | RESOURCE ADDRESS           | TYPE              | ID           | REPLACEMENT | MODULE      | DANGER
Add    | aws_instance.web_server   | aws_instance      | -            | Never       | -           | 
Modify | aws_s3_bucket.data       | aws_s3_bucket     | bucket-123   | Conditional | app/storage | 
Remove | aws_rds_instance.old     | aws_rds_instance  | db-456       | N/A         | -           | 
Replace| aws_ec2_instance.app     | aws_instance      | i-1234567890 | Always      | compute/web | ⚠️ Sensitive resource replacement
```

### Implementation Tasks

#### 3.1 Enhance Data Models
- [x] **Update `ResourceChange` struct** in `lib/plan/models.go`:
  - [x] Add `PhysicalID` field (string) - current physical resource ID
  - [x] Add `PlannedID` field (string) - planned physical resource ID
  - [x] Add `ModulePath` field (string) - module hierarchy path
  - [x] Add `ChangeAttributes` field ([]string) - specific attributes changing
  - [x] Update `ReplacementType` field usage
  - [x] Add danger highlights fields (`IsDangerous`, `DangerReason`, `DangerProperties`)

#### 3.2 Enhance Change Analysis
- [x] **Update `analyzeResourceChanges()` method** in `lib/plan/analyzer.go`:
  - [x] Extract physical resource IDs from plan
  - [x] Parse module path information
  - [x] Identify specific changing attributes
  - [x] Determine replacement reasoning
  - [x] Integrate danger highlights detection for sensitive resources and properties

- [x] **Add helper methods** in `lib/plan/analyzer.go`:
  - [x] `extractPhysicalID(change *tfjson.ResourceChange) string`
  - [x] `extractModulePath(address string) string`
  - [x] `getChangingAttributes(change *tfjson.ResourceChange) []string`
  - [x] `extractPlannedID(change *tfjson.ResourceChange) string`

#### 3.3 Update Formatter
- [x] **Create `formatResourceChangesTable()` method** in `lib/plan/formatter.go`:
  - [x] Generate enhanced resource changes table using go-output
  - [x] Include all new columns (ID, REPLACEMENT, MODULE, DANGER)
  - [x] Apply appropriate formatting and colours
  - [x] Handle long resource addresses and IDs
  - [x] Integrate danger highlights display with warnings and property details

- [x] **Update table column configuration**:
  - [x] Define column widths and alignment
  - [x] Add colour coding for different actions
  - [x] Format danger information with emoji indicators
  - [ ] Implement truncation for long values (optional enhancement)
  - [ ] Add sorting options (optional enhancement)

#### 3.4 Module Support
- [x] **Add module detection logic** in `lib/plan/analyzer.go`:
  - [x] Parse module hierarchy from resource addresses
  - [x] Format module paths for display
  - [x] Handle nested module structures

- [x] **Update table layout** in `lib/plan/formatter.go`:
  - [x] Show MODULE column with proper module paths
  - [x] Format module paths consistently (e.g., "app/storage")
  - [ ] Dynamic column width adjustment based on content (optional enhancement)

#### 3.5 Physical ID Handling
- [x] **Implement ID extraction** in `lib/plan/analyzer.go`:
  - [x] Parse current physical IDs from plan
  - [x] Handle missing IDs for new resources
  - [x] Parse planned IDs for resource changes

- [x] **Add ID display logic** in `lib/plan/formatter.go`:
  - [x] Show "-" for new resources (no current ID)
  - [x] Show actual IDs for existing resources
  - [x] Handle ID display for different change types
  - [ ] Truncate long IDs with ellipsis (optional enhancement)

#### 3.6 Testing
- [x] **Unit tests** for enhanced resource analysis:
  - [x] Test physical ID extraction
  - [x] Test module path parsing
  - [x] Test replacement necessity analysis
  - [x] Test danger highlights detection

- [x] **Integration tests** for resource changes table:
  - [x] Test complete table output with all columns
  - [x] Test with modules and without
  - [x] Test with various resource types
  - [x] Test column formatting and alignment
  - [x] Test danger highlights integration

### Success Criteria
- [x] Resource changes table includes all new columns (ACTION, RESOURCE, TYPE, ID, REPLACEMENT, MODULE, DANGER)
- [x] Physical IDs are displayed correctly
- [x] Replacement indicators show proper values (Never, Conditional, Always, N/A)
- [x] Module information is displayed when present
- [x] Danger highlights are properly integrated with warnings and property details
- [x] Table formatting is clean and readable
- [x] Performance remains acceptable for large plans
- [x] Full backward compatibility maintained

---

## Phase 4: Integration and Testing

### Objective
Integrate all new functionality, ensure compatibility, and provide comprehensive testing coverage.

### Implementation Tasks

#### 4.1 Output Flow Integration
- [x] **Update `OutputSummary()` method** in `lib/plan/formatter.go`:
  - [x] Implement complete output flow:
    1. Plan Information section
    2. Enhanced Statistics Summary table
    3. Enhanced Resource Changes table (with integrated danger highlights)
  - [x] Ensure proper spacing between sections
  - [x] Maintain section order consistency

- [x] **Add section control flags** in `cmd/plan_summary.go`:
  - [x] Add `--show-statistics` flag (default: true)
  - [x] Add `--details` flag for resource changes table (default: false)
  - [x] Update help text and examples with new functionality
  - [ ] Add `--show-plan-info` flag (optional enhancement for granular control)

#### 4.2 Configuration Management
- [x] **Update configuration structure** in `config/config.go`:
  - [x] Add statistics display configuration options
  - [x] Add danger highlights configuration (sensitive resources/properties)
  - [x] Maintain backward compatibility
  - [ ] Add additional display sections (optional enhancement)

- [x] **Add configuration validation**:
  - [x] Provide sensible defaults for all new options
  - [x] Handle invalid configurations gracefully
  - [x] Proper configuration binding with viper

#### 4.3 Error Handling
- [x] **Enhance error handling** throughout the codebase:
  - [x] Handle missing plan information gracefully
  - [x] Provide fallbacks for parsing failures
  - [x] Add informative error messages with context wrapping
  - [x] Handle edge cases in replacement analysis and ID extraction

#### 4.4 Performance Optimization
- [x] **Optimize plan parsing** in `lib/plan/parser.go`:
  - [x] Efficient memory usage with proper data structures
  - [x] Streamlined parsing without unnecessary allocations
  - [ ] Add progress indicators for large files (optional enhancement)

- [x] **Optimize table rendering** in `lib/plan/formatter.go`:
  - [x] Efficient string formatting using go-output library
  - [x] Proper memory management for large change sets
  - [ ] Add pagination for very large change sets (optional enhancement)

#### 4.5 Comprehensive Testing
- [x] **Unit test coverage**:
  - [x] Achieve high coverage for new functionality (24 passing tests)
  - [x] Test edge cases and error conditions
  - [x] Test replacement analysis with various scenarios
  - [x] Test danger highlights detection

- [x] **Integration test suite**:
  - [x] Test complete workflow with comprehensive test cases
  - [x] Test all output formats (table, json)
  - [x] Test with various change combinations
  - [x] Test performance with different scenarios

- [x] **Regression testing**:
  - [x] Ensure existing functionality remains unchanged
  - [x] Test backward compatibility
  - [x] Verify command-line interface consistency

#### 4.6 Documentation Updates
- [x] **Update command help text** in `cmd/plan_summary.go`:
  - [x] Add examples for new functionality (--stats-format, --show-statistics)
  - [x] Document new flags and options
  - [x] Update usage examples with comprehensive examples

- [x] **Update README and documentation**:
  - [x] Document configuration options in CLAUDE.md
  - [x] Provide comprehensive development guide
  - [ ] Add screenshots of new output format (optional enhancement)

- [x] **Update changelog** in `changelog.md`:
  - [x] Document all new features comprehensively
  - [x] Note backward compatibility maintenance
  - [x] Provide detailed feature descriptions

### Success Criteria
- [x] All phases integrate seamlessly
- [x] Complete output flow works as designed
- [x] Performance meets requirements
- [x] All tests pass with high coverage (24/24 tests passing)
- [x] Documentation is comprehensive and accurate
- [x] Backward compatibility is maintained

---

## Implementation Guidelines

### Code Quality Standards
- Follow existing code patterns and conventions
- Maintain consistent error handling approaches
- Use meaningful variable and function names
- Add comprehensive comments for complex logic
- Ensure thread safety where applicable

### Testing Requirements
- Unit tests for all new functions and methods
- Integration tests for complete workflows
- Performance tests for large plan files
- Error condition testing
- Backward compatibility testing

### Documentation Standards
- Update all relevant documentation
- Provide clear examples and usage patterns
- Document configuration options thoroughly
- Include troubleshooting guides
- Maintain changelog accuracy

### Performance Considerations
- Optimize for large Terraform plan files
- Minimize memory usage during parsing
- Implement efficient table rendering
- Consider pagination for very large outputs
- Profile and benchmark critical paths

---

## Risk Mitigation

### Technical Risks
- **Large plan file performance**: Implement streaming and pagination
- **Memory usage**: Use efficient data structures and lazy loading
- **Terraform version compatibility**: Test with multiple versions
- **Backend configuration parsing**: Provide fallbacks for unknown backends

### User Experience Risks
- **Output format changes**: Maintain backward compatibility flags
- **Information overload**: Provide granular display controls
- **Configuration complexity**: Use sensible defaults and validation

### Maintenance Risks
- **Code complexity**: Keep modules focused and well-documented
- **Test maintenance**: Automate testing with CI/CD integration
- **Documentation drift**: Update docs as part of development process

---

## Success Metrics

### Functional Metrics
- [ ] Plan information displays correctly for all supported backends
- [ ] Statistics summary matches fog's horizontal format exactly
- [ ] Resource changes table includes all required columns
- [ ] All output formats (table, json, html) work correctly
- [ ] Performance acceptable for plans with 1000+ resources

### Quality Metrics
- [ ] Unit test coverage ≥ 95% for new code
- [ ] Integration test coverage ≥ 90% for new workflows
- [ ] Zero regression in existing functionality
- [ ] Documentation completeness score ≥ 95%
- [ ] User acceptance testing passes

### User Experience Metrics
- [ ] Output format matches fog's design principles
- [ ] Information hierarchy is clear and logical
- [ ] Command-line interface remains intuitive
- [ ] Configuration options are discoverable
- [ ] Error messages are helpful and actionable

---

## Conclusion

This implementation plan provides a comprehensive roadmap for adding fog-inspired changeset display functionality to strata. The phased approach ensures systematic development while maintaining existing functionality and code quality standards.

The plan focuses on the two core missing functionalities:
1. **Plan Information Display** - Providing essential context about the plan
2. **Enhanced Resource Changes Table** - Detailed resource information with IDs, replacements, and modules

Additionally, the enhanced statistics summary table will provide a clean, horizontal format matching fog's design while preserving all existing functionality.

Each phase builds upon the previous one, ensuring a logical development progression and enabling incremental testing and validation. The comprehensive testing strategy and quality guidelines ensure the implementation meets professional standards and provides a solid foundation for future enhancements.


## Tool use and completion reminder

Remember to use the tools / functions available to you. After each phase is complete, you must check off any tasks that have been completed in full. Then stop and I will review your work.
