# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Statistics Calculation Enhancement**:
  - Enhanced `calculateStatistics` method to properly handle output changes by accepting both resource changes and output changes as parameters
  - Added output change counting that excludes no-op outputs per requirement 4.5, ensuring statistics accurately reflect only meaningful changes
  - Updated all callers of `calculateStatistics` to pass output changes for comprehensive statistics tracking

### Added
- **Comprehensive Testing Infrastructure for Output Refinements**:
  - Added `output_refinements_edge_cases_test.go` with extensive edge case testing including empty plans, no-op only plans, complex sensitive structures, and identical resource addresses
  - Added `output_refinements_integration_test.go` with complete end-to-end integration tests covering real Terraform plan workflows and all output formats
  - Added `output_refinements_test_compatibility_test.go` with backward compatibility validation ensuring existing functionality remains unaffected
  - Added `testdata/output_refinements_plan.json` with comprehensive test data for validation scenarios
  - Performance testing with large plans (1000+ resources) to ensure <5% performance impact
  - Enhanced unit tests for statistics behavior with output changes, verifying no-op exclusion logic

### Added
- **Enhanced ActionSortTransformer for Table Output**:
  - Enhanced `hasDangerIndicator` method with improved danger detection logic to identify dangerous resources in table rows using existing danger column regex patterns and handle edge cases where danger indicators might be ambiguous
  - Updated `Transform` method to implement enhanced table sorting that sorts table rows first by danger indicators, then by action priority (delete=0, replace=1, update=2, create=3, noop=4), then alphabetically by resource address
  - Added `extractAction`, `getActionPriority`, and `extractResourceAddress` helper methods for robust table data extraction and sorting logic
  - Comprehensive unit tests for danger indicator detection, table sorting, action extraction, and resource address parsing with edge case handling
  - Completed tasks 6.1 and 6.2 from Output Refinements feature implementation

### Added
- **No-Op Filtering Implementation for Output Refinements feature**:
  - Implemented `filterNoOps` method in `lib/plan/formatter.go` for resource filtering that removes resources with `ChangeTypeNoOp` when `ShowNoOps` is false
  - Implemented `filterNoOpOutputs` method in `lib/plan/formatter.go` for output filtering that removes outputs with `IsNoOp == true` when `ShowNoOps` is false
  - Enhanced `OutputSummary` method to integrate no-op filtering based on `f.config.Plan.ShowNoOps` configuration
  - Added "No changes detected" message display when no actual changes exist after filtering
  - Ensured statistics remain unchanged and count all resources including no-ops (preserves original behavior)
  - Completed tasks 4.1, 4.2, and 4.3 from Output Refinements feature implementation

- **No-op resource and output detection for Output Refinements feature**:
  - Added `IsNoOp` field to `ResourceChange` struct for internal no-op resource tracking
  - Added `IsNoOp` field to `OutputChange` struct for internal no-op output tracking  
  - Enhanced resource analysis to automatically mark resources with no changes (`ChangeTypeNoOp`)
  - Enhanced output analysis to detect outputs with identical before/after values using deep equality comparison
- **Property Sorting and Sensitive Value Masking**:
  - Implemented alphabetical property sorting within resources (case-insensitive) with natural sort ordering for numbers and special characters
  - Added path hierarchy sorting for same-named properties to ensure consistent ordering
  - Implemented immediate sensitive value masking during property extraction for security by default
  - Added `maskSensitiveValue` helper function that returns "(sensitive value)" for primitive types while preserving structure for nested objects
  - Enhanced property comparison logic to mask sensitive values immediately in `compareObjects` function
  - Added comprehensive unit tests for property sorting, natural sort ordering, sensitive value masking, and nested structure handling

- **No-Op Resources Display Control**:
  - Added `ShowNoOps` field to `PlanConfig` struct with `mapstructure:"show-no-ops"` tag for YAML configuration support
  - Added `--show-no-ops` CLI flag to plan summary command with boolean override capability
  - Implemented CLI flag precedence over configuration file setting for flexible usage
  - Added comprehensive unit tests for flag parsing, configuration loading, and precedence behavior

- **Output Refinements Feature Documentation**:
  - Added comprehensive design documentation for output refinements feature addressing issues #17-#21 with property sorting, no-op display control, and sensitive value masking
  - Added detailed decision log documenting architectural choices, security considerations, and stakeholder feedback integration
  - Added complete requirements specification covering property sorting improvements, sensitivity-based resource ordering, configurable no-op display, and comprehensive sensitive value masking
  - Added implementation task breakdown with phased development approach and testing strategy

### Changed
- **Development Tool Configuration**:
  - Updated Claude settings to include mcp__devtools__search_packages in allowlist for enhanced package management capabilities

## [1.2.3] - 2025-08-06

### Fixed
- Fixed table header key mismatch in sensitive resource display table where data was using Title Case keys ("Action", "Resource", etc.) but table creation was using uppercase keys ("ACTION", "RESOURCE", etc.), which would have resulted in empty table columns when displaying sensitive resources in summary mode

## [1.2.2] - 2025-08-05

### Added
- **Table Header Consistency Feature Documentation**:
  - Added comprehensive design documentation, decision log, and UI/UX improvements analysis for table header consistency feature
  - Added detailed implementation specifications with design principles, component architecture, and testing strategy
  - Enhanced test coverage for formatter functionality with additional test scenarios

### Changed
- **Documentation Organization**:
  - Reorganized feature documentation from generic UI-IMPROVEMENTS.md to specific feature-based directory structure
  - Updated table header formatting implementation in formatter.go and associated tests for improved consistency

### Removed
- **Legacy Documentation**:
  - Removed generic UI-IMPROVEMENTS.md documentation file in favor of feature-specific organization

## [1.2.1] - 2025-08-04

### Fixed
- **Table Output Cleanup**:
  - Removed three unnecessary columns from Resource Changes table that were accidentally introduced during terraform-unknown-values-and-outputs feature implementation
  - Removed `has_unknown_values`, `unknown_properties`, and `property_change_details` fields from table display as they were implementation details that leaked into the user interface
  - Table now correctly shows only the 8 expected columns: ACTION, RESOURCE, TYPE, ID, REPLACEMENT, MODULE, DANGER, and PROPERTY_CHANGES
  - Updated test expectations to align with actual feature requirements rather than testing implementation artifacts

### Changed
- **Test Improvements**:
  - Simplified unknown value formatting test to focus on actual requirements (displaying "(known after apply)" text) rather than internal field structure
  - Improved test maintainability by removing dependencies on internal implementation details

## [1.2.0] - 2025-08-04

### Added
- **Terraform Unknown Values and Outputs Support**:
  - Added support for Terraform unknown values, displaying them as "(known after apply)" to match Terraform's syntax
  - Added comprehensive outputs section that displays after resource changes with 5-column format (NAME, ACTION, CURRENT, PLANNED, SENSITIVE)
  - Enhanced data models with `IsUnknown`, `UnknownType`, and `HasUnknownValues` fields for tracking unknown states
  - Implemented unknown value detection in `after_unknown` field with proper path navigation
  - Added sensitive output handling with "(sensitive value)" masking and ⚠️ indicator support
  - Outputs section automatically suppressed when no output changes exist

### Changed
- **Performance Optimizations**:
  - Cached compiled regex patterns in ActionSortTransformer to eliminate redundant compilation overhead
  - Optimized table row parsing functions to use pre-compiled regex patterns instead of runtime compilation
  - Enhanced `formatValueWithContext` function to detect and properly format unknown values

- **Code Quality Improvements**:
  - Refactored string literals to named constants throughout analyzer and formatter for improved maintainability
  - Enhanced output formatting to display "(known after apply)" without quotes, matching Terraform's exact syntax

### Fixed
- **Test Compatibility**:
  - Updated JSON serialization tests to include new unknown values and outputs fields
  - Maintained backward compatibility for existing JSON output structure

## [1.1.6] - 2025-08-02

### Fixed
- **Nested Property Display Improvements**:
  - Fixed nested properties like tags to display with proper visual hierarchy and indentation instead of inline format
  - Enhanced property change analysis to treat nested objects (tags, metadata, labels, config objects) as single grouped property changes
  - Implemented cross-format compatible indentation using Unicode En spaces (U+2002) to ensure consistent visual spacing in table, markdown, HTML, and JSON outputs
  - Updated property change formatter to detect complex nested values and apply appropriate nested formatting with visual indentation
  - Enhanced nested map and array formatting with proper line breaks and hierarchical display structure

### Changed
- **Property Analysis Logic**:
  - Modified `compareObjects` function in analyzer.go to identify and group nested object properties instead of splitting them into individual property changes
  - Added `shouldTreatAsNestedObject` helper function to detect common nested property patterns (tags, metadata, labels, etc.)
  - Updated property change formatter to use context-aware formatting with `formatValueWithContext` method
  - Enhanced test expectations to match new nested property display format with proper indentation

### Dependencies
- **Go-Output Library**:
  - Updated go-output dependency from v2.1.1 to v2.1.2 for improved nested content rendering support

## [1.1.5] - 2025-08-01

### Fixed
- **Property Change Formatting**:
  - Fixed create action property formatting to properly display Terraform diff-style `+` prefix for new resources
  - Simplified property change formatter by removing unnecessary context-based special handling
  - Unified property change display logic for consistent formatting across all resource actions

### Changed
- **Dependency Updates**:
  - Updated go-output dependency from v2.1.0 to v2.1.1 for improved collapsible content handling

- **Sample File Naming Standardization**:
  - Renamed `k8ssample.json` to `k8s-sample.json` for consistent hyphenated naming convention
  - Renamed `websample.json` to `web-sample.json` for consistent hyphenated naming convention
  - Added new `wildcards-sample.json` test fixture for wildcard IAM policy testing scenarios
  - Updated all test references to use the new standardized sample file names

### Fixed
- **Property Change Analysis Improvements**:
  - Refactored property comparison logic in analyzer.go to eliminate duplicate property change detection
  - Unified compareObjects function to handle replacement path checking and sensitive value processing in a single pass
  - Improved property change formatter to handle context-aware formatting with proper data structure mapping
  - Enhanced replacement path handling using simplified string-based matching for better performance

### Changed
- **Code Simplification**:
  - Simplified replacement path handling from complex slice operations to string-based matching
  - Streamlined property change detection by removing redundant comparison functions
  - Enhanced formatter data structure to use map-based context passing instead of wrapper types

### Removed
- **Deprecated Functions**:
  - Removed deprecated compareObjectsWithReplacePaths function in favor of unified compareObjects
  - Eliminated deduplicatePropertyChanges function as improved comparison logic prevents duplicates at source
  - Removed PropertyChangesWithContext wrapper type in favor of direct map-based data passing

## [1.1.3] - 2025-07-30

### Added
- **Configuration Enhancements**:
  - Added `use_emoji` configuration option to allow users to control emoji usage in output
  - Added `max_detail_length` configuration for collapsible sections with default 10KB limit to prevent excessive content
  - Enhanced collapsible content with proper truncation limits based on configuration
- **Terraform-Style Property Change Formatting**:
  - Implemented Terraform diff-style formatting for property changes with `+`, `-`, `~` prefixes for add, remove, and update actions
  - Added `propertyChangesFormatterTerraform` and `formatPropertyChange` functions with support for complex value types (maps, arrays, primitives)
  - Property changes with sensitive values automatically expand when `AutoExpandDangerous` is enabled
- **Enhanced Property Analysis**:
  - Deep object comparison algorithm in analyzer.go with recursive comparison logic for maps, slices, and primitives
  - Property analysis helper functions: `extractPropertyName`, `parsePath`, `isSensitive`, `extractSensitiveChild`, `extractSensitiveIndex`
  - Action tracking for property changes to track "add", "remove", "update" operations
  - Array index path parsing with support for complex paths (e.g., matrix[1][2])
  - Performance limits: 100 properties max per resource, 10KB per property value, 10MB total memory
- **Empty Table Suppression**:
  - Implemented filtering to exclude no-op changes from Resource Changes tables
  - Enhanced provider grouping to use changed resource count (excluding no-ops) for threshold calculations
- **Comprehensive Testing Infrastructure**:
  - Complete end-to-end integration test suite covering all output improvements
  - Performance testing with artificial plan generation supporting up to 1000+ resources
  - New test fixture `samples/complex-properties-sample.json` for property change scenarios
  - Backward compatibility validation for JSON output structure consistency

### Changed
- **Output and Formatting Improvements**:
  - Updated sensitive value masking to use consistent "(sensitive value)" format across all output modes
  - Changed default table style from "ColoredBlackOnMagentaWhite" to "default" for better accessibility
  - Switched from direct property formatter to Terraform-style formatter for improved readability
  - Enhanced auto-expansion to respect `AutoExpandDangerous` configuration setting
- **Configuration Structure**:
  - Replaced `show_dependencies` field with `max_detail_length` in expandable sections configuration
  - Improved case-insensitive validation for output formats
- **Code Organization and Architecture**:
  - Refactored complex resource table display logic by extracting monolithic methods into focused helper functions
  - Enhanced comment documentation and fixed Go variable declaration style to use modern syntax
  - Unified table creation pattern using single `output.New().AddContent().Build()` document building pattern
  - Limited ActionSortTransformer to table/JSON/CSV formats only, excluding markdown/HTML
  - Adjusted collapsible formatting performance threshold from 3x to 6x slower
- **Resource Analysis and Display**:
  - Updated `analyzePropertyChanges` method to use new deep comparison algorithm
  - Enhanced PropertyChange struct with Action field and improved documentation
  - Modified provider group headers to show only changed resource counts, excluding no-op resources

### Fixed
- **Output Sensitivity Handling**:
  - Fixed output sensitivity detection to properly check boolean values from Terraform's BeforeSensitive/AfterSensitive fields
  - Enhanced output change analysis to properly detect and mask sensitive output values
- **Property Change Processing**:
  - Fixed missing property change details in collapsible sections by properly integrating PropertyChanges field
  - Improved null value handling in property comparison to avoid redundant checks
  - Fixed non-deterministic map key ordering in formatValue function by sorting keys alphabetically
- **Multi-table Rendering**:
  - Resolved critical bug where Plan Information and Summary Statistics tables were missing in markdown/HTML output
  - Simplified plan rendering architecture using `output.NewTableContent()` pattern
  - Enhanced provider-based resource grouping to use proper collapsible sections with auto-expansion

### Removed
- **Dependencies Column and Functionality**:
  - Removed dependencies column from resource table display
  - Removed `dependenciesFormatterDirect()` function and `DependencyInfo` data structure from models
  - Removed test cases for dependency extraction, circular dependency detection, and dependency formatter functions
  - Updated `ResourceAnalysis` struct to remove `Dependencies` field, streamlining the analysis model

## [1.1.0] - 2025-07-28

### Added
- **Enhanced Summary Visualization with Progressive Disclosure**:
  - Complete enhanced summary visualization feature implementation with progressive disclosure and collapsible sections
  - Smart provider-based resource grouping with configurable thresholds (activates with multiple providers and 10+ resources by default)
  - Global `--expand-all` / `-e` CLI flag to expand all collapsible sections across all output formats
  - Comprehensive resource analysis with property change detection, dependency extraction, and 4-tier risk assessment
  - Advanced collapsible formatters for property changes, dependencies, and provider-grouped resources using go-output v2
  - Performance optimization with memory limits, property truncation, and processing safeguards
- **Configuration Migration System**:
  - Automatic migration of deprecated configuration formats with graceful fallback handling
  - New configuration structure with `expandable_sections`, `grouping`, and `performance_limits` sections
  - User-friendly warnings and default value provisioning for new configuration options
- **GitHub Action Enhancements**:
  - Added `expand-all` input parameter to GitHub Action with full parameter flow support
  - Enhanced GitHub Action integration with collapsible sections in PR comments
  - Maintains backward compatibility with existing GitHub Action configurations
- **Core Data Models and Analysis**:
  - New `ResourceAnalysis`, `PropertyChangeAnalysis`, `PropertyChange`, and `DependencyInfo` data structures
  - Enhanced `ResourceChange` model with Provider, TopChanges, and ReplacementHints fields
  - Comprehensive analysis functions including `analyzePropertyChanges()`, `assessRiskLevel()`, and `AnalyzeResource()`
  - Enhanced context extraction for resource replacement reasons using Terraform ReplacePaths data
- **Comprehensive Testing Infrastructure**:
  - End-to-end integration tests with 5 comprehensive scenarios covering the complete flow
  - Performance validation tests with benchmarks for small/medium/large plans
  - Error handling tests for malformed plans, memory limits, and edge cases
  - Unit tests for all new formatters, analysis functions, and data models
  - Test fixtures with realistic Terraform plan JSON data for various scenarios

### Fixed
- Compilation errors in formatter.go due to incorrect go-output v2 format constant usage
- ActionSortTransformer format detection with correct v2 format constants
- Analyzer nil plan handling with proper fallback in GenerateSummary method

### Changed
- Updated configuration structure from `progressive_disclosure` to `expandable_sections`
- Enhanced plan analyzer with smart grouping capabilities and context extraction
- Streamlined risk assessment to focus on essential information with performance limits
- Enhanced danger evaluation system with specific reasons for database, compute, storage, security, and network changes
- All deletion operations now considered risky by default with appropriate danger messaging
- Updated formatter to use go-output v2 collapsible document building instead of legacy table-only approach

## [1.0.0] - 2025-07-24

### Added
- **Complete Terraform Plan Analysis Tool**
  - CLI foundation with Cobra framework for plan summary command
  - Terraform plan file parsing with support for both binary and JSON formats
  - Plan analysis engine with change categorization and statistics
  - Destructive change detection and warning system
  - Support for multiple output formats (table, json, html, markdown)
  - Configuration management with YAML config files

- **Danger Highlights System**
  - Sensitive resource replacement detection with configurable resource types
  - Sensitive property change detection with detailed property listings
  - Enhanced output formatting with danger information and warning indicators
  - Configuration options for defining sensitive resources and properties
  - High-risk column in statistics summary to highlight dangerous changes

- **Advanced Output Features**
  - File output system with dual output support (stdout + file)
  - Dynamic file naming with placeholder support ($TIMESTAMP, $AWS_REGION, $AWS_ACCOUNTID)
  - Action sorting to prioritize dangerous changes (Remove > Replace > Modify > Add)
  - Enhanced statistics table with horizontal layout and comprehensive change tracking
  - Plan information display showing context, versions, and metadata
  - Always-show-sensitive feature to display critical changes even when details disabled
  - Table style configuration support with `--table-style` flag and `table.style` config option
  - Table max column width configuration with `--table-max-column-width` flag and `table.max-column-width` config option
  - UNMODIFIED column to statistics summary showing no-op resource changes
  - Support for handling plans with no resource changes (displays "All resources are unchanged")

- **GitHub Action Integration**
  - Composite GitHub Action for CI/CD workflows
  - Automated binary download with platform detection and checksum verification
  - Pull request comment functionality with update/create logic
  - GitHub Step Summary integration with rich Markdown formatting
  - Comprehensive error handling and input validation

- **Version Management**
  - Version command with detailed build information
  - Build-time version injection via ldflags
  - JSON output format support for version information
  - Automated release workflow for cross-platform binaries (Linux, Windows, macOS on amd64/arm64)

- **Development Infrastructure**
  - Comprehensive test suites for all functionality
  - Enhanced Makefile with development targets
  - Project documentation and development guides
  - Sample Terraform plan files for testing

### Dependencies
- github.com/ArjenSchwarz/go-output v2.0.5 for enhanced output formatting
- github.com/hashicorp/terraform-json v0.25.0 for Terraform plan parsing
- github.com/spf13/cobra v1.9.1 for CLI framework
- github.com/spf13/viper v1.20.1 for configuration management
- github.com/mitchellh/go-homedir v1.1.0 for home directory detection

## [0.1.0] - 2025-05-24

### Added
- Initial project setup
- Basic Go module structure
- MIT License
