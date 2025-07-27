# Enhanced Summary Visualization - Implementation Tasks

This document provides an actionable implementation plan for the Enhanced Summary Visualization feature, broken down into discrete coding tasks that build incrementally on each other.

## Current Implementation Status

### ✅ Already Completed
- **Basic data model enhancements**: `Provider`, `TopChanges`, `ReplacementHints` fields added to ResourceChange
- **Provider extraction**: `extractProvider()` function working with aws_, azurerm_, google_ patterns
- **Replacement hints**: `extractReplacementHints()` function extracting human-readable reasons
- **Property changes**: `getTopChangedProperties()` function (limited to 3 properties)
- **Configuration**: `GroupingThreshold`, `ShowContext` configuration options
- **go-output v2 dependency**: v2.0.5 added to go.mod

### ✅ Recently Completed
- **go-output v2 API usage**: Fixed compilation issues, project builds successfully
- **Enhanced analysis functions**: All core analysis functions implemented and tested
- **Data models**: Complete data structures for ResourceAnalysis, PropertyChangeAnalysis, DependencyInfo
- **Unit tests**: Comprehensive test coverage for all new analysis functions

### 🔄 In Progress (Partial Implementation)
- **Formatter**: Basic v2 migration working but needs collapsible formatters
- **Data models**: Enhanced fields added and comprehensive analysis structures complete

### ⏳ Not Started Yet
- **Collapsible formatters**: No collapsible content formatters implemented yet
- **Provider grouping logic**: Only extraction done, no grouping algorithm
- **Expand-all flag**: No CLI flag or configuration support
- **GitHub Actions integration**: No environment detection
- **Advanced dependency extraction**: Basic implementation done, needs full plan analysis

## Prerequisites

- Requirements document: `agents/enhanced-summary-visualization/requirements.md`
- Design document: `agents/enhanced-summary-visualization/design.md`
- Existing codebase understanding from `lib/plan/` modules

## Implementation Tasks

### 1. Update Core Data Models for go-output v2 Integration

- [x] ~~1.1 Update data models in `lib/plan/models.go`~~ **COMPLETED**
  - ✅ `Provider`, `TopChanges`, `ReplacementHints` fields already added to ResourceChange
  - ✅ Added `ResourceAnalysis` struct with `PropertyChanges`, `RiskLevel`, `Dependencies` fields
  - ✅ Added `PropertyChangeAnalysis` struct with `Changes`, `Count`, `TotalSize`, `Truncated` fields  
  - ✅ Added `PropertyChange` struct with `Name`, `Path`, `Before`, `After`, `Sensitive`, `Size` fields
  - ✅ Added `DependencyInfo` struct with `DependsOn`, `UsedBy` fields
  - References requirements: 1.6 (dependencies in expandable sections), 2.3 (ALL property changes)

- [x] ~~1.1a Complete data model updates in `lib/plan/models.go`~~ **COMPLETED**
  - ✅ Added missing `ResourceAnalysis`, `PropertyChangeAnalysis`, `PropertyChange`, `DependencyInfo` structs
  - ✅ Kept existing enhanced fields (`Provider`, `TopChanges`, `ReplacementHints`) for backward compatibility
  - ✅ Ensured new structs work alongside existing fields

- [x] ~~1.2 Update configuration structures in `config/config.go`~~ **COMPLETED**
  - ✅ Added `ExpandAll bool` field to root configuration structure
  - ✅ Enhanced existing configuration with `ExpandableSections ExpandableSectionsConfig`
  - ✅ Added `ExpandableSectionsConfig` with `Enabled`, `AutoExpandDangerous`, `ShowDependencies` fields
  - ✅ Added `GroupingConfig` and `PerformanceLimitsConfig` for enhanced configuration
  - ✅ Maintained backward compatibility with existing configuration files
  - References requirements: 5.1-5.5 (global expand control), 1.6 (expandable sections configuration)

- [x] ~~1.3 Add performance limit constants and configuration~~ **COMPLETED**
  - ✅ Added `PerformanceLimitsConfig` struct with `MaxPropertiesPerResource`, `MaxPropertySize`, `MaxTotalMemory` fields
  - ✅ Set default limits: 100 properties, 1MB property size, 100MB total memory
  - ✅ Added `GetPerformanceLimitsWithDefaults()` method for configuration validation and defaults
  - References design: Performance and scalability section

- [x] ~~1.4 Write unit tests for updated data models~~ **COMPLETED**
  - ✅ Added tests for `ResourceAnalysis` struct creation, serialization, and field access
  - ✅ Added tests for `PropertyChangeAnalysis` truncation behavior and sensitive data handling
  - ✅ Added tests for configuration loading with new `expand_all` and `expandable_sections` fields
  - ✅ Added tests for `GetPerformanceLimitsWithDefaults()` with various configuration scenarios
  - ✅ Verified backward compatibility with old configuration format

### 2. Implement Core Analysis Functions

- [x] ~~2.1 Add provider extraction in `lib/plan/analyzer.go`~~ **COMPLETED**
  - ✅ `extractProvider(resourceType string) string` function already implemented
  - ✅ Basic string splitting logic for aws_, azurerm_, google_ patterns implemented

- [x] ~~2.2 Add replacement hints extraction in `lib/plan/analyzer.go`~~ **COMPLETED**  
  - ✅ `extractReplacementHints(change *tfjson.ResourceChange) []string` function already implemented
  - ✅ Human-readable replacement reasons extraction already working

- [x] ~~2.3 Add top property changes extraction in `lib/plan/analyzer.go`~~ **COMPLETED**
  - ✅ `getTopChangedProperties(change *tfjson.ResourceChange, limit int) []string` already implemented
  - ✅ Currently limited to 3 properties but working

- [x] ~~2.4 Add enhanced property change analysis in `lib/plan/analyzer.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `analyzePropertyChanges(change *tfjson.ResourceChange, maxProps int) (PropertyChangeAnalysis, error)`
  - ✅ Extract ALL property changes with before/after values (no 3-property limit)
  - ✅ Implemented depth-limited recursive comparison for nested properties (max depth: 5)
  - ✅ Track total size and set `Truncated` flag when limits exceeded (10MB limit)
  - ✅ References requirements: 2.3 (ALL property changes with before/after values)

- [x] ~~2.5 Add simplified risk assessment in `lib/plan/analyzer.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `assessRiskLevel(change *tfjson.ResourceChange) string`
  - ✅ Return "critical", "high", "medium", or "low" based on change type and resource sensitivity
  - ✅ Use delete operations as high risk, sensitive resource deletes as critical
  - ✅ Use existing sensitive resource configuration and danger detection logic
  - ✅ References requirements: 3.1-3.5 (risk analysis), 3.7 (auto-expand high-risk)

- [x] ~~2.6 Add dependency extraction in `lib/plan/analyzer.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `extractDependenciesWithLimit(change *tfjson.ResourceChange, maxDeps int) (*DependencyInfo, error)`
  - ✅ Extract basic dependencies from explicit `depends_on` attributes
  - ✅ Apply depth limit as circuit breaker for complex dependency chains (100 deps max)
  - ✅ Basic implementation ready for future enhancement with full plan dependency analysis
  - ✅ References requirements: 3.6 (dependencies in expandable sections)

- [x] ~~2.7 Implement main resource analysis function in `lib/plan/analyzer.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `AnalyzeResource(change *tfjson.ResourceChange) (*ResourceAnalysis, error)`
  - ✅ Call property analysis, risk assessment, and dependency extraction
  - ✅ Handle errors gracefully with partial data and logging
  - ✅ Include replacement reasons from existing `extractReplacementHints()` function
  - ✅ References design: Simplified analyzer structure

- [x] ~~2.8 Write comprehensive unit tests for new analysis functions~~ **COMPLETED**
  - ✅ Test enhanced property change analysis with various data types and nesting levels
  - ✅ Test risk level assessment for different change types and resource types
  - ✅ Test dependency extraction with basic scenarios and limits
  - ✅ Test main analysis function integration and error handling
  - ✅ Test value size estimation and comparison functions
  - ✅ All tests pass successfully with comprehensive coverage

### 3. Fix and Enhance go-output v2 Integration

- [x] ~~3.1 Basic go-output v2 integration started~~ **PARTIALLY COMPLETED**
  - ✅ go-output v2.0.5 already added to go.mod
  - ✅ Basic v2 API calls already implemented in formatter.go
  - ❌ Compilation errors due to incorrect API usage (FormatTable, FormatMarkdown, etc.)
  - ❌ Need to fix API calls to match actual v2 interface

- [x] ~~3.1a Fix go-output v2 API usage in `lib/plan/formatter.go` (URGENT)~~ **COMPLETED**
  - ✅ Fixed undefined references: `output.FormatTable`, `output.FormatMarkdown`, etc.
  - ✅ Used correct v2 format constants: `output.Table.Name`, `output.Markdown.Name`, etc.
  - ✅ Fixed `output.New()` and table creation API calls to match v2 interface
  - ✅ Basic compilation working, project builds successfully

- [x] ~~3.2 Add collapsible property formatter in `lib/plan/formatter.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `propertyChangesFormatter() func(any) any`
  - ✅ Returns `output.NewCollapsibleValue` with property count summary and detailed changes
  - ✅ Auto-expands when sensitive properties are changed
  - ✅ Handles `PropertyChangeAnalysis` with truncation indicator
  - ✅ References requirements: 2.3 (expandable property changes), design: go-output v2 integration

- [x] ~~3.3 Add collapsible dependencies formatter in `lib/plan/formatter.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `dependenciesFormatter() func(any) any`
  - ✅ Returns `output.NewCollapsibleValue` with dependency count summary and detailed relationships
  - ✅ Formats "Depends On" and "Used By" lists clearly
  - ✅ Collapses by default since dependencies are supplementary information
  - ✅ References requirements: 3.6 (dependencies in expandable sections)

- [x] ~~3.4 Implement table data preparation in `lib/plan/formatter.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `prepareResourceTableData(changes []ResourceChange) []map[string]any`
  - ✅ Uses existing ResourceChange data with graceful error handling
  - ✅ Prepares data structure with `address`, `change_type`, `risk_level`, `property_changes`, `dependencies`
  - ✅ Includes replacement reasons when applicable
  - ✅ References design: Data transformation for go-output v2

- [x] ~~3.5 Add main formatting function in `lib/plan/formatter.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `formatResourceChangesWithProgressiveDisclosure(summary *PlanSummary) (*output.Document, error)`
  - ✅ Uses `output.New().Table()` with schema containing collapsible formatters
  - ✅ Ready for global expand-all setting integration from configuration
  - ✅ Built with go-output v2 document builder pattern
  - ✅ References requirements: 1.1-1.7 (progressive disclosure with collapsible sections)

- [x] ~~3.6 Write unit tests for fixed and enhanced formatters~~ **COMPLETED**
  - ✅ Fix existing formatter tests to work with v2 API
  - ✅ Test property formatter with sensitive and non-sensitive changes
  - ✅ Test dependencies formatter with various dependency patterns
  - ✅ Test table data preparation with mixed resource types
  - ✅ Mock go-output v2 components for isolated testing

### 4. Implement Provider Grouping with Collapsible Sections

- [x] ~~4.1 Add provider extraction function in `lib/plan/analyzer.go`~~ **COMPLETED**
  - ✅ `extractProvider(resourceType string) string` already implemented
  - ✅ Basic string splitting logic for aws_, azurerm_, google_ patterns working
  - ✅ Returns fallback for unrecognized patterns
  - References requirements: 1.5 (smart grouping threshold), design: Simple provider extraction

- [x] ~~4.2 Add grouping logic in `lib/plan/analyzer.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `groupByProvider(changes []ResourceChange) map[string][]ResourceChange`
  - ✅ Uses existing `extractProvider()` function and `GroupingThreshold` configuration
  - ✅ Only groups when resource count meets threshold and multiple providers present
  - ✅ Skips grouping if all resources from same provider
  - ✅ References requirements: 1.5 (smart grouping), 1.6 (omit grouping when not needed)

- [x] ~~4.3 Add grouped formatting with collapsible sections in `lib/plan/formatter.go` (NEW)~~ **COMPLETED**
  - ✅ Implemented `formatGroupedWithCollapsibleSections(summary *PlanSummary, groups map[string][]ResourceChange) (*output.Document, error)`
  - ✅ Creates sections using `builder.Section()` for each provider group  
  - ✅ Includes `hasHighRiskChanges()` helper for auto-expand logic
  - ✅ Uses go-output v2 Section API (NewCollapsibleTable API not available)
  - ✅ References requirements: 1.5 (provider grouping), design: CollapsibleSection integration

- [x] ~~4.4 Write unit tests for provider grouping (EXTEND EXISTING)~~ **COMPLETED**
  - ✅ Added comprehensive tests for `groupByProvider()` covering all scenarios
  - ✅ Test grouping logic with threshold and provider diversity
  - ✅ Test grouped formatting with collapsible sections - basic functionality test
  - ✅ Test `hasHighRiskChanges()` behavior for auto-expand logic
  - ✅ Added helper functions for testing provider extraction

### 5. Add Global Expand-All Flag Support

- [ ] 5.1 Add CLI flag to root command in `cmd/root.go`
  - Add `--expand-all` / `-e` persistent flag to root command
  - Set default value to false
  - Add flag description: "Expand all collapsible sections"
  - References requirements: 5.1 (global --expand-all CLI flag)

- [ ] 5.2 Update plan summary command in `cmd/plan_summary.go`
  - Read expand-all flag value in command execution
  - Override configuration `ExpandAll` setting when CLI flag is provided
  - Apply setting to go-output v2 `CollapsibleConfig.GlobalExpansion`
  - References requirements: 5.3 (CLI flag overrides config), design: Global expand configuration

- [ ] 5.3 Update output configuration in `lib/plan/formatter.go`
  - Implement `createOutputWithConfig(format output.Format) *output.Output`
  - Set `CollapsibleConfig.GlobalExpansion` from configuration or CLI override
  - Configure other collapsible behavior (detail length, indicators)
  - Use `output.WithCollapsibleConfig()` to apply settings
  - References requirements: 5.5 (apply to all collapsible content), design: go-output v2 integration

- [ ] 5.4 Write tests for expand-all functionality
  - Test CLI flag parsing and application
  - Test configuration override behavior
  - Test global expansion applied to all collapsible content
  - Test precedence: CLI flag > config file > default

### 6. Add GitHub Action Integration Support

- [ ] 6.1 Add GitHub Actions environment detection in `lib/plan/formatter.go`
  - Implement `isGitHubActions() bool` function checking `GITHUB_ACTIONS` environment variable
  - Automatically use Markdown format when running in GitHub Actions
  - Respect expand-all configuration for GitHub Action output
  - References requirements: 4.1-4.4 (GitHub Action integration)

- [ ] 6.2 Update command execution for GitHub Actions in `cmd/plan_summary.go`
  - Call GitHub Actions detection in command execution
  - Auto-expand dangerous changes in CI environment via configuration
  - Ensure Markdown output compatible with GitHub PR comments
  - References requirements: 4.2 (Markdown compatible), 4.3 (respect expand-all config)

- [ ] 6.3 Write tests for GitHub Actions integration
  - Test environment variable detection
  - Test automatic Markdown format selection
  - Test expand-all flag behavior in GitHub Actions environment
  - Mock environment variables for testing

### 7. Integration and End-to-End Testing

- [ ] 7.1 Create test fixtures for comprehensive scenarios
  - Create `testdata/simple_plan.json` with basic resource changes
  - Create `testdata/multi_provider_plan.json` for grouping tests
  - Create `testdata/high_risk_plan.json` with sensitive resources and deletions
  - Create `testdata/dependencies_plan.json` with resource dependencies
  - References design: Testing strategy

- [ ] 7.2 Write end-to-end integration tests
  - Test complete flow from plan parsing to formatted output with collapsible sections
  - Test provider grouping with collapsible sections and auto-expand
  - Test expand-all flag affecting all collapsible content
  - Test GitHub Actions environment behavior
  - Verify backward compatibility with existing configurations

- [ ] 7.3 Add error handling and edge case tests
  - Test behavior with malformed Terraform plans
  - Test graceful degradation when analysis fails
  - Test memory limits with large property changes
  - Test circular dependency detection
  - Ensure user-friendly error messages

- [ ] 7.4 Write performance validation tests
  - Test analysis performance with plans containing 100+ resources
  - Verify memory usage stays within reasonable bounds
  - Test that performance limits prevent excessive resource usage
  - Benchmark collapsible formatter performance vs simple display

### 8. Final Integration and Validation

- [ ] 8.1 Update command help and documentation
  - Update `cmd/plan_summary.go` help text for new flags and behavior
  - Add examples showing expandable sections in command help
  - Document the global expand-all flag usage
  - References requirements: 5.1 (CLI flag documentation)

- [ ] 8.2 Add configuration migration support
  - Handle old configuration files gracefully in `config/config.go`
  - Provide helpful warnings when deprecated config options are used
  - Ensure smooth transition from old to new configuration structure
  - References design: Backward compatibility

- [ ] 8.3 Perform final integration validation
  - Run all tests with go-output v2 integration
  - Validate that expandable sections work correctly across all output formats
  - Test complete user workflows with various configuration combinations
  - Verify no regression in existing functionality

## Task Dependencies

**CRITICAL PATH (Fix broken code first):**
1. **Task 3.1a** (Fix go-output v2 API) must be completed FIRST to restore compilation
2. **Tasks 1.1a, 1.2** (Complete data models) should be done early for foundation

**MAIN IMPLEMENTATION PATH:**
- Tasks 1.x must be completed before all others (foundation)
- **Task 3.1a is URGENT** - needed before any other formatter work
- Tasks 2.4-2.7 (new analysis functions) must be completed before 3.2+ (new collapsible formatters)
- Tasks 3.x must be completed before 4.x (basic formatting before grouping)
- Tasks 4.x and 5.x can be done in parallel after 3.x
- Task 6.x can be done after 5.x (GitHub Actions needs expand-all support)
- Tasks 7.x can begin after 6.x (integration testing)
- Tasks 8.x should be completed last (final validation)

**PRIORITY ORDER:**
1. **URGENT**: Task 3.1a (fix compilation errors)
2. **HIGH**: Tasks 1.1a, 1.2 (complete data models)
3. **HIGH**: Tasks 2.4-2.7 (new analysis functions)
4. **MEDIUM**: Tasks 3.2+ (collapsible formatters)
5. **MEDIUM**: Tasks 4.2+ (grouping logic), Tasks 5.x (expand-all flag)
6. **LOW**: Tasks 6.x (GitHub Actions), 7.x (testing), 8.x (validation)

## Requirements Coverage

This implementation plan covers all requirements:

1. **Progressive Disclosure (1.1-1.7)** - Tasks 1.x, 3.x, 4.x
2. **Comprehensive Change Context (2.1-2.3)** - Tasks 2.x, 3.1-3.2  
3. **Enhanced Risk Analysis (3.1-3.7)** - Tasks 2.2, 3.x
4. **GitHub Action Integration (4.1-4.4)** - Tasks 6.x
5. **Global Expand Control (5.1-5.5)** - Tasks 5.x

All tasks focus on code implementation that can be executed by a coding agent, building incrementally on previous work.