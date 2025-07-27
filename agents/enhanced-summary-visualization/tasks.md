# Enhanced Summary Visualization with Collapsible Sections - Implementation Tasks

This document provides an actionable implementation plan for the Enhanced Summary Visualization feature with collapsible sections support, broken down into discrete coding tasks that build incrementally on each other.

## Important Note on Existing Completed Tasks

Some tasks marked as completed (1.x, 2.x, 3.x) were implemented based on the original limited design. These need to be updated for the new collapsible sections approach, which removes the "first 3 properties" limitation and adds comprehensive progressive disclosure.

## Prerequisites

- Requirements document: `agents/enhanced-summary-visualization/requirements.md`
- Design document: `agents/enhanced-summary-visualization/design.md`
- Existing codebase understanding from `lib/plan/` modules

## Implementation Tasks

### 1. Update Data Models for Collapsible Sections (NEEDS REWORK)

- [x] ~~1.1 Add new fields to ResourceChange struct~~ - **COMPLETED BUT NEEDS UPDATE**
  - **DEPRECATED**: `TopChanges []string` field (limited to 3 properties)
  - **DEPRECATED**: `ShowContext` configuration approach
  - **NEW TASK NEEDED**: Update to comprehensive analysis approach

- [x] ~~1.2 Extend PlanConfig struct~~ - **COMPLETED BUT NEEDS UPDATE**
  - **DEPRECATED**: Simple boolean flags approach
  - **NEW TASK NEEDED**: Replace with progressive disclosure configuration structure

- [x] ~~1.3 Write unit tests for data model extensions~~ - **COMPLETED BUT NEEDS UPDATE**
  - **NEW TASK NEEDED**: Update tests for new comprehensive data model

**NEW TASKS TO REPLACE COMPLETED ONES:**

- [ ] 1.4 Update ResourceChange struct for collapsible sections in `lib/plan/models.go`
  - Add `ComprehensiveAnalysis *ComprehensiveChangeAnalysis` field
  - Add `SummaryView string` field for main view display
  - Add `RiskLevel string` field ("low", "medium", "high", "critical")
  - Add `HasExpandableSections bool` field
  - Remove or deprecate `TopChanges` field (replaced by comprehensive analysis)

- [ ] 1.5 Add CollapsibleSection and related structs to `lib/plan/models.go`
  - Add `CollapsibleSection` struct with Title, Content, AutoExpand, SectionType
  - Add `ComprehensiveChangeAnalysis` struct with all analysis fields
  - Add `PropertyChange`, `RiskAssessment`, `DependencyInfo` structs
  - References requirement: Progressive disclosure with collapsible sections

- [ ] 1.6 Update PlanConfig for progressive disclosure in `config/config.go`
  - Replace simple flags with `ProgressiveDisclosure ProgressiveDisclosureConfig`
  - Replace simple flags with `Grouping GroupingConfig`
  - Add new configuration structs with comprehensive options
  - Maintain backward compatibility with existing config files

- [ ] 1.7 Write unit tests for updated data models
  - Test new ResourceChange fields and comprehensive analysis structure
  - Test CollapsibleSection serialization and functionality
  - Test updated PlanConfig loading with progressive disclosure options
  - Test backward compatibility with old configuration formats

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

### 3. Implement Comprehensive Change Analysis (NEEDS MAJOR REWORK)

- [x] ~~3.1 Implement limited context extraction~~ - **COMPLETED BUT NEEDS UPDATE**
  - **DEPRECATED**: Limited replacement hints approach
  - **NEW TASK NEEDED**: Comprehensive analysis with risk assessment

- [x] ~~3.2 Implement limited property change detection~~ - **COMPLETED BUT NEEDS UPDATE**
  - **DEPRECATED**: "First 3 properties" limitation
  - **NEW TASK NEEDED**: ALL property changes with comprehensive details

- [x] ~~3.3 Enhance existing danger reason logic~~ - **COMPLETED BUT NEEDS UPDATE**
  - **PARTIALLY USABLE**: Basic danger detection can be extended
  - **NEW TASK NEEDED**: Comprehensive risk assessment with mitigation

- [x] ~~3.4 Write unit tests for limited context extraction~~ - **COMPLETED BUT NEEDS UPDATE**
  - **NEW TASK NEEDED**: Tests for comprehensive analysis

**NEW TASKS TO REPLACE COMPLETED ONES:**

- [ ] 3.5 Implement comprehensive change analysis in `lib/plan/analyzer.go`
  - Add `AnalyzeChangeComprehensively(change *tfjson.ResourceChange) *ComprehensiveChangeAnalysis`
  - Extract ALL property changes with before/after values (no limit)
  - Generate comprehensive risk assessment with impact analysis
  - Extract dependencies and relationships between resources
  - References requirement: Comprehensive change context with progressive disclosure

- [ ] 3.6 Implement risk assessment engine in `lib/plan/analyzer.go`
  - Add `assessRisk(change *tfjson.ResourceChange) RiskAssessment` function
  - Determine risk levels: low, medium, high, critical
  - Generate impact assessments and potential consequences
  - Determine auto-expand behavior for high-risk changes
  - References requirement: Enhanced risk analysis with detailed mitigation

- [ ] 3.7 Implement mitigation suggestion engine in `lib/plan/analyzer.go`
  - Add `generateMitigationSuggestions(change, risk) []string` function
  - Provide actionable recommendations based on resource type and risk
  - Include alternative deployment strategies for high-risk changes
  - Generate specific mitigation steps for different scenarios
  - References requirement: Risk mitigation guidance

- [ ] 3.8 Implement dependency extraction in `lib/plan/analyzer.go`
  - Add `extractDependencies(change *tfjson.ResourceChange) DependencyInfo` function
  - Identify resources that this change depends on
  - Identify resources that depend on this change
  - Extract relationship information for display in collapsible sections
  - References requirement: Dependency visualization

- [ ] 3.9 Write comprehensive unit tests for new analysis functions
  - Test comprehensive change analysis with various resource types
  - Test risk assessment accuracy and auto-expand logic
  - Test mitigation suggestion generation for different scenarios
  - Test dependency extraction and relationship identification

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

### 5. Implement Progressive Disclosure Formatter (MAJOR REWORK NEEDED)

**DEPRECATED TASKS:**
- [ ] ~~5.1 Simple grouped display~~ - **NEEDS COMPLETE REWORK**
- [ ] ~~5.2 Basic provider grouping~~ - **NEEDS ENHANCEMENT FOR COLLAPSIBLE SECTIONS**
- [ ] ~~5.3 Limited context display~~ - **NEEDS COMPREHENSIVE REWORK**
- [ ] ~~5.4 Basic formatter tests~~ - **NEEDS UPDATE FOR NEW APPROACH**

**NEW TASKS FOR PROGRESSIVE DISCLOSURE:**

- [ ] 5.5 Implement progressive disclosure formatter in `lib/plan/formatter.go`
  - Add `formatResourceChangesWithProgressiveDisclosure()` function
  - Create main view with essential information only (resource name, change type, risk level)
  - Generate collapsible sections for each resource using comprehensive analysis
  - Use go-output library's new collapsible sections capability
  - References requirement: Progressive disclosure with collapsible sections

- [ ] 5.6 Implement collapsible section creation in `lib/plan/formatter.go`
  - Add `createCollapsibleSections(change, analysis) []CollapsibleSection` function
  - Create property changes section with ALL properties (no limit)
  - Create risk analysis section with mitigation suggestions
  - Create dependencies section when enabled
  - Create replacement reasons section for replacements
  - Implement auto-expand logic for high-risk sections

- [ ] 5.7 Implement provider grouping with collapsible sections
  - Update `formatGroupedByProvider()` to use collapsible framework
  - Create provider-level collapsible sections
  - Auto-expand provider groups containing high-risk changes
  - Nest resource-level collapsible sections within provider groups
  - References requirement: Progressive disclosure within provider grouping

- [ ] 5.8 Implement section content formatting
  - Add `formatPropertyChanges(properties []PropertyChange)` for detailed property display
  - Add `formatRiskAnalysis(risk, mitigations)` for comprehensive risk display
  - Add `formatDependencies(deps DependencyInfo)` for relationship visualization
  - Respect sensitive property handling in detailed views
  - Include before/after values with proper formatting

- [ ] 5.9 Write comprehensive tests for progressive disclosure formatter
  - Test collapsible section creation and content
  - Test auto-expand logic for high-risk resources
  - Test provider grouping with nested collapsible sections
  - Test that main view remains clean while details are comprehensive
  - Mock go-output collapsible sections functionality

### 6. Update Command Line Interface for Progressive Disclosure

**NEEDS SIGNIFICANT UPDATES:**

- [ ] 6.1 Update configuration flags in `cmd/plan_summary.go`
  - **DEPRECATED**: `--show-context` flag (replaced by progressive disclosure)
  - **UPDATE**: Enhance `--group-by-provider` to work with collapsible sections
  - **NEW**: Add `--progressive-disclosure` flag (default: true)
  - **NEW**: Add `--auto-expand-dangerous` flag (default: true) 
  - **NEW**: Add `--show-dependencies` flag (default: true)
  - **NEW**: Add `--show-mitigation` flag (default: true)
  - Update command help text for new progressive disclosure approach

- [ ] 6.2 Update default configuration file template
  - Replace old configuration structure with progressive disclosure section
  - Add comprehensive configuration options with inline documentation
  - Ensure backward compatibility with old config files
  - Provide migration path from old to new configuration format
  - Update `strata.yaml` template with new structure

- [ ] 6.3 Write integration tests for progressive disclosure flow
  - Create test fixtures with multi-provider Terraform plans
  - Test end-to-end flow with collapsible sections enabled
  - Verify auto-expand behavior for high-risk changes
  - Test both CLI flags and configuration file options for new features
  - Test backward compatibility with old configuration files

### 7. Testing and Validation for Progressive Disclosure

**ENHANCED TESTING FOR COLLAPSIBLE SECTIONS:**

- [ ] 7.1 Create comprehensive test fixtures for progressive disclosure
  - `testdata/single_provider_plan.json` - Plan with only AWS resources (should not group)
  - `testdata/multi_provider_plan.json` - Plan with AWS, Azure, GCP resources (should group)
  - `testdata/high_risk_plan.json` - Plan with dangerous changes requiring auto-expand
  - `testdata/complex_dependencies_plan.json` - Plan with complex resource dependencies
  - `testdata/sensitive_properties_plan.json` - Plan with sensitive property changes
  - `testdata/large_plan.json` - Plan with 50+ resources to test performance and display
  - `testdata/replacement_plan.json` - Plan with various replacement scenarios

- [ ] 7.2 Write end-to-end tests for progressive disclosure
  - Test complete analysis and formatting pipeline with collapsible sections
  - Verify collapsible sections are created correctly for each resource
  - Test auto-expand behavior for high-risk changes
  - Test comprehensive property change display (no 3-property limit)
  - Test risk analysis and mitigation suggestion display
  - Test dependency visualization in collapsible sections
  - Ensure no regressions in existing functionality

- [ ] 7.3 Add enhanced error handling and edge case tests
  - Test behavior with malformed Terraform plans
  - Test progressive disclosure configuration loading with invalid values
  - Test graceful degradation when collapsible section creation fails
  - Test fallback to simple display when go-output collapsible sections unavailable
  - Test memory usage with very large plans containing comprehensive analysis
  - Verify error messages are user-friendly for new features

### 8. Performance Optimization for Progressive Disclosure

**ENHANCED PERFORMANCE REQUIREMENTS:**

- [ ] 8.1 Implement parallel processing for comprehensive analysis
  - Add worker pool pattern for processing comprehensive change analysis concurrently
  - Parallelize risk assessment, dependency extraction, and mitigation generation
  - Implement parallel collapsible section creation for large resource lists
  - Include configuration option to control worker count
  - **NEW REQUIREMENT**: Handle increased computational load from comprehensive analysis

- [ ] 8.2 Add memory streaming for collapsible sections
  - Implement buffered writing for large formatted outputs with collapsible sections
  - Stream collapsible section content instead of building complete sections in memory
  - Optimize memory usage for comprehensive property change data
  - Add memory usage monitoring for progressive disclosure features
  - **NEW REQUIREMENT**: Handle memory impact of storing ALL property changes

- [ ] 8.3 Write performance tests for progressive disclosure
  - Create very large test fixtures (200+ resources) to test comprehensive analysis impact
  - Benchmark performance with vs without comprehensive analysis enabled
  - Test memory usage with progressive disclosure vs simple display
  - Verify performance scales appropriately with increased analysis depth
  - Test collapsible section creation performance with large datasets
  - **NEW REQUIREMENT**: Ensure comprehensive analysis doesn't degrade performance significantly

- [ ] 8.4 Implement progressive analysis loading (NEW)
  - Add option to lazy-load comprehensive analysis only when sections are expanded
  - Implement caching for expensive analysis operations
  - Add progressive analysis depth based on resource risk level
  - Optimize analysis pipeline to focus on high-risk resources first
  - **NEW REQUIREMENT**: Support scenarios where full analysis isn't always needed

## Updated Task Dependencies for Progressive Disclosure

**IMPORTANT**: Some dependencies have changed due to the progressive disclosure approach:

- **Tasks 1.4-1.7** must be completed before all other tasks (new foundation)
- **Tasks 2.x** can continue as-is (provider extraction still needed)
- **Tasks 3.5-3.9** must be completed before Tasks 5.x (comprehensive analysis needed for collapsible sections)
- **Task 4.x** depends on completion of Task 2.x (grouping needs provider extraction)
- **Tasks 5.5-5.9** depend on completion of Tasks 3.5-3.9 and 4.x (progressive disclosure formatter needs comprehensive analysis)
- **Task 6.x** depends on completion of Task 5.x (CLI needs working progressive disclosure formatter)
- **Tasks 7.x** can begin after Task 6.x (integration testing)
- **Task 8.x** can be completed in parallel with Task 7.x (performance optimization)

## Updated Requirements Coverage

All tasks above ensure complete coverage of the updated requirements:

1. **Progressive Disclosure with Collapsible Sections** - Covered by tasks 1.4-1.7, 5.5-5.9, 6.1-6.2
2. **Comprehensive Change Context with Progressive Disclosure** - Covered by tasks 3.5-3.9, 5.6, 5.8
3. **Enhanced Risk Analysis with Detailed Mitigation** - Covered by tasks 3.6-3.7, 5.6, 5.8

## Implementation Strategy for Existing Completed Tasks

Since tasks 1.1-1.3, 3.1-3.4 were completed based on the old design:

1. **Immediate Priority**: Complete tasks 1.4-1.7 to update data models
2. **Refactor Existing**: Update completed analysis code (3.1-3.4) to support comprehensive approach  
3. **Extend Rather Than Replace**: Build on existing provider extraction and basic analysis
4. **Backward Compatibility**: Ensure old configurations continue to work during transition

The implementation plan maintains the test-driven development approach while accommodating the significant enhancement from collapsible sections capability.