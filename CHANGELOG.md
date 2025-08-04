# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Terraform Unknown Values and Outputs Integration Tests**:
  - Added comprehensive end-to-end integration tests in `lib/plan/analyzer_test.go` for complete workflow validation
  - Added `TestCompleteWorkflowWithUnknownValuesAndOutputsIntegration` test with comprehensive unknown values, outputs, and danger highlighting validation
  - Added `TestCrossFormatConsistencyForUnknownValuesAndOutputs` test for display consistency across all output formats
  - Verified unknown values display correctly and don't appear as deletions (requirements 1.1, 1.2)
  - Validated outputs section displays with correct 5-column format (requirement 2.2)
  - Ensured "(known after apply)" and "(sensitive value)" display consistently across formats (requirements 1.3, 2.4)
  - Tested integration with existing danger highlighting functionality (requirement 3.1)
  - Completed tasks 9.1 and 9.2 for end-to-end integration testing
- **Terraform Outputs Section Formatter Implementation**:
  - Implemented `handleOutputDisplay` function in formatter.go for outputs section rendering after resource changes (requirement 2.1)
  - Added `createOutputChangesData` function to create 5-column output table data: NAME, ACTION, CURRENT, PLANNED, SENSITIVE (requirement 2.2)
  - Added `formatOutputValue` function with proper handling of sensitive values "(sensitive value)" and unknown values "(known after apply)" (requirements 2.3, 2.4)
  - Implemented outputs section suppression when no output changes exist (requirement 2.8)
  - Added outputs table integration with go-output builder pattern for cross-format compatibility
  - Completed tasks 8.1 and 8.2 for outputs section formatting implementation
- **Comprehensive Terraform Outputs Testing**:
  - Complete test coverage for `ProcessOutputChanges` function with all output change scenarios
  - Unit tests for `analyzeOutputChange` covering create, update, delete actions with proper indicators (+, ~, -)
  - Test validation for sensitive output masking with "(sensitive value)" display
  - Test validation for unknown output values with "(known after apply)" display
  - Integration tests for end-to-end outputs processing workflow
  - Edge case testing for sensitive unknown outputs with proper precedence handling
  - Task completion tracking for terraform-unknown-values-and-outputs feature (tasks 7.1, 7.2)
- **Terraform Outputs Processing Implementation**:
  - Added `ProcessOutputChanges` function in analyzer.go for comprehensive output change processing with graceful error handling (requirement 2.1)
  - Added `analyzeOutputChange` function for individual output analysis with action detection (create/Add, update/Modify, delete/Remove) and visual indicators (+, ~, -) (requirements 2.5-2.7)
  - Added `getOutputActionAndIndicator` helper function for consistent action naming and visual indicator mapping across output formats
  - Enhanced output processing with unknown value detection using existing `after_unknown` logic, displaying "(known after apply)" for unknown output values (requirement 2.3)
  - Implemented sensitive output handling with "(sensitive value)" masking and ⚠️ indicator support (requirement 2.4)
  - Added graceful handling of missing `output_changes` field, returning empty list instead of errors (requirement 2.8)
  - Maintained backward compatibility through legacy `analyzeOutputChanges` method that internally calls new processing functions
  - Integrated outputs processing into main `GenerateSummary` workflow for seamless plan analysis
- **Terraform Unknown Values and Outputs Support - Data Models**:
  - Added unknown values support to `PropertyChange` struct with `IsUnknown` and `UnknownType` fields for tracking before/after/both unknown states (requirement 1.6, 1.7)
  - Added unknown values tracking to `ResourceChange` struct with `HasUnknownValues` and `UnknownProperties` fields for complete unknown value visibility (requirement 1.2, 1.5)
  - Added outputs support fields to `OutputChange` struct with `IsUnknown`, `Action`, and `Indicator` fields for comprehensive output change display (requirement 2.3, 2.5-2.7)
  - Enhanced data models with proper JSON serialization tags for cross-format consistency (requirement 3.4)
- **Terraform Unknown Values Detection Implementation**:
  - Implemented `isValueUnknown` helper function to detect unknown values in `after_unknown` field with proper path navigation (requirement 1.6)
  - Enhanced `compareObjects` function to handle `afterUnknown` parameter and detect unknown values throughout object hierarchy
  - Added `getUnknownValueDisplay` function returning "(known after apply)" for consistent Terraform syntax display (requirement 1.3)
  - Implemented unknown values override logic to prevent false deletion detection for unknown properties (requirement 1.2)
  - Enhanced property change analysis to populate `HasUnknownValues` and `UnknownProperties` fields in resource changes (requirement 1.5)
- **Terraform Unknown Values Testing Infrastructure**:
  - Added comprehensive unit tests for unknown value detection functions (`isValueUnknown`, `getUnknownValueDisplay`)
  - Added integration tests for enhanced `compareObjects` function with unknown values processing
  - Added test coverage for unknown values integration with sensitive property detection
  - Added test verification that unknown values don't appear as deletions (requirement 1.2)
  - Added test validation of exact "(known after apply)" string display (requirement 1.3)
  - Enhanced test coverage for edge cases including nested objects, arrays, and complex data types
- **Terraform Unknown Values and Outputs Feature Planning**:
  - Added comprehensive requirements documentation for handling unknown values and output changes
  - Added detailed design specification for implementing `after_unknown` field processing and outputs section
  - Added feature decision log documenting key implementation choices and technical decisions
  - Added feature task breakdown for development implementation phases
  - Added UI/UX analysis for output changes display with visual hierarchy recommendations
- **Development Environment Enhancements**:
  - Added additional MCP tool permissions for internet search and URL fetching capabilities
- **Terraform Unknown Values Display**:
  - Support for displaying Terraform unknown values as `(known after apply)` without quotes in all output formats
  - Comprehensive test coverage for unknown value formatting across table, JSON, HTML, and Markdown formats

### Changed
- **Output Formatting**:
  - Enhanced `formatValueWithContext` function to detect and properly format unknown values
  - Updated property change formatting to display `(known after apply)` without quotes, matching Terraform's exact syntax

### Fixed
- **Test Compatibility**:
  - Updated JSON serialization tests to include new unknown values and outputs fields in expected output
  - Fixed `TestResourceChange_SerializationWithNewFields` and `TestResourceAnalysis_Serialization` test expectations
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
