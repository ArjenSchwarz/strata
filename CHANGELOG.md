# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **GitHub Action JSON Body Creation**: Enhanced error handling in PR comment creation and update functions by safely creating JSON bodies with proper error checking and validation before making GitHub API calls, preventing potential failures during comment operations
- **GitHub Action Output Formatting**: Fixed GitHub Action output formatting by using heredoc format (`<<EOF`) for multiline outputs instead of simple assignment, added proper line endings for JSON output, and ensured consistent output formatting across all GitHub Action output variables (`summary`, `json-summary`) to prevent potential parsing issues in GitHub workflows.
- **GitHub Action Security**: Enhanced cleanup process to remove analysis JSON file from workspace for improved security by preventing potential information disclosure
- **Output File Configuration**: Fixed output file path handling to preserve case sensitivity by using `GetString` instead of `GetLCString` for the `output-file` configuration parameter, ensuring file paths maintain their original casing.

### Changed
- **GitHub Action Binary Download Enhancement**: Improved binary download system with enhanced version tag resolution, MD5 checksum verification replacing SHA256, platform-specific checksum handling with fallback mechanisms, better error reporting with expected vs actual checksum logging, and graceful degradation when checksum verification fails. Enhanced retry logic with fallback to latest version when specific versions are not found.
- **GitHub Action CLI Flag Standardization**: Updated CLI flag from `--show-details` to `--details` for consistency with main application interface and improved user experience.

### Removed
- **Test Infrastructure Cleanup**: Removed obsolete danger threshold validation tests and related test functions that are no longer relevant after configuration simplification. Streamlined test suite by removing unused validation functions and improving test maintainability.

### Changed
- **Test Infrastructure Refactoring**: Renamed action simplified test to foundation test and reorganized test files from root to test directory. Removed obsolete comprehensive unit and error handling tests while adding new foundation and comprehensive test targets to Makefile for cleaner test organization.

### Changed
- **GitHub Action Architecture Simplification**: Completely redesigned GitHub Action with single-file implementation reducing code from ~1970 lines across 7 files to ~400 lines in one script. Removed all shell modules (`lib/action/` directory) and consolidated functionality into simplified `action.sh` with improved reliability, faster execution, and clearer logging.
- **GitHub Action Documentation**: Updated README.md with comprehensive documentation for the new simplified architecture, including new `strata-version` parameter for version control, improved features list highlighting reliability improvements, and enhanced usage examples with version pinning capabilities.

### Removed
- **Shell Module Dependencies**: Removed 6 shell modules (`binary.sh`, `files.sh`, `github.sh`, `security.sh`, `strata.sh`, `utils.sh`) from `lib/action/` directory as part of architecture simplification, consolidating all functionality into the main action script.

### Added
- **Comprehensive Integration Test Suite**: Implemented complete integration test infrastructure for GitHub Action with four main components: Master Test Runner (`run_integration_test_suite.sh`) orchestrating all tests, End-to-End Integration Tests (`test_comprehensive_integration.sh`) validating full action execution with all sample plan files, Backwards Compatibility Tests (`test_backwards_compatibility.sh`) ensuring existing workflows continue working, and Performance Benchmarks (`test_performance_benchmarks.sh`) measuring binary download, analysis startup, and total execution time against thresholds (<10s, <5s, <30s respectively). Includes comprehensive test reports, requirements traceability, and integration with existing Makefile structure.

### Changed
- **GitHub Action Simplification Tasks**: Updated task completion status in `specs/action-simplification/tasks.md` to mark task 8 and all subtasks (8.1-8.3) as completed, reflecting the successful implementation of the comprehensive integration test suite covering end-to-end functionality, backwards compatibility validation, and performance benchmarking.

### Added
- **GitHub Features Test Suite**: Implemented comprehensive test suite (`test/test_github_features.sh`) with 882 lines of tests covering all GitHub integration features including Step Summary generation, PR context detection, comment marker generation, comment creation and updates, and error handling scenarios. Tests validate GitHub Action requirements 8.1-8.11 with mock API responses, environment variable handling, and graceful fallbacks for non-PR contexts.
- **Strata Execution Component Testing**: Added comprehensive test suite for Strata execution component (`test/test_strata_execution.sh`) with 915 lines of tests covering command construction with all parameters, dual output system (display to stdout + JSON to file), JSON parsing for statistics extraction, output generation for all formats (markdown/json/table/html), and error handling scenarios. Includes mock Strata binary creation for testing different scenarios (success, failure, dangerous changes) and validation of GitHub Action output generation.

### Fixed
- **GitHub Action Code Quality**: Fixed all shellcheck warnings in action_simplified.sh by removing unused variables (SCRIPT_NAME, comment_on_pr) and separating variable declarations from assignments to avoid masking return values
- **Test Code Quality**: Fixed all shellcheck warnings in `test/action_simplified_test.sh` by separating variable declarations from assignments to avoid masking return values and using underscore for unused loop variable

### Changed
- **Task Management**: Updated GitHub Action simplification tasks document to reflect completion of task group 5 (GitHub integration features) including all three subtasks: test implementation for GitHub features, Step Summary generation, and PR comment functionality with API integration.
- **Code Documentation**: Enhanced code comments in `action_simplified.sh` for better documentation of temporary directory creation and cleanup operations

### Technical Notes
- **GitHub Integration Test Coverage**: The new test suite covers Step Summary writing to GITHUB_STEP_SUMMARY, PR context detection via GITHUB_EVENT_NAME, comment marker generation with workflow/job names, API calls for comment creation/updates, and graceful handling of non-PR contexts. Includes mock curl functions for testing GitHub API interactions without external dependencies.
- **Test Infrastructure**: The test suite validates the existing `run_analysis` and `extract_outputs` functions in `action_simplified.sh` which implement the dual output capability using Strata's native `--file` and `--file-format` flags to output display format to stdout while writing JSON metadata to file for GitHub Action outputs.
- **GitHub Action Integration**: Tests confirm proper extraction of statistics from JSON metadata files including total_changes, dangerous_changes, and proper setting of GitHub Action outputs (has-changes, has-dangers, change-count, danger-count, summary, json-summary).

### Previous Unreleased Changes
- **GitHub Action Simplification Specifications**: Added comprehensive requirements and decision log for simplifying the GitHub Action implementation. Features a major refactoring plan to reduce complexity by 60%, streamline binary downloads, improve error messages, and enhance user experience while maintaining 100% backwards compatibility for release as v1.5.0.
- **Implementation Tasks Document**: Created detailed task breakdown for GitHub Action simplification with 9 major categories and 32 subtasks, including test-first development approach, unit and integration testing requirements, and performance benchmarks.
- **Simplified GitHub Action Implementation**: Created single-file GitHub Action implementation (`action_simplified.sh`) that reduces code complexity by 60% while maintaining full backwards compatibility. Features modular error handling, comprehensive logging, retry mechanisms for downloads, and unified security validation. Includes complete test infrastructure with unit tests for core foundation, error handling, cleanup mechanisms, and validation functions.

## [1.4.0] - 2025-09-18

### Added
- **Data Pipeline Sorting System**: Replaced the legacy ActionSortTransformer with a modern data-level sorting pipeline that provides 5-10x performance improvement. Features danger-first sorting, proper action priority ordering (Remove → Replace → Modify → Add), and comprehensive test coverage with 279 lines of unit tests covering all sorting scenarios and edge cases.
- **Performance Benchmarking Infrastructure**: Added comprehensive benchmark test suite supporting data sizes from 10 to 10,000 resources with memory allocation tracking. Benchmarks demonstrate significant performance improvements over the previous regex-based approach.

### Changed
- **Architecture Modernization**: Refactored sorting logic from string-based table parsing to data-level operations, eliminating ~200 lines of regex-based code. Enhanced code maintainability with consistent constant references and cleaner separation between data decoration and sorting logic.

### Fixed
- **Output Filtering**: Enhanced plan summary output to intelligently hide empty string property values for addition and deletion actions, reducing noise while preserving meaningful change information for update actions.

## [1.3.0] - 2025-08-29

### Changed
- **Go Language Modernization**: Upgraded to Go 1.25.0 with modern patterns including `any` type alias, `for range n` syntax, and `slices.Contains()`
- **Test Suite Organization**: Refactored large test files into focused modules following Go testing best practices with proper `t.Cleanup()` and standardized naming conventions
- **Enhanced Benchmark Infrastructure**: Added comprehensive benchmark targets, memory profiling, and performance comparison capabilities to Makefile
- **Reduced External Dependencies**: Replaced testify assertions with standard library equivalents across test suite
- **Code Quality Improvements**: Fixed Go linting issues, improved code formatting consistency, and enhanced constant usage patterns

### Added
- **Development Configuration**: Serena MCP integration, comprehensive project memories, and enhanced Claude development environment
- **Code Modernization Feature**: Complete feature specifications with requirements, decision log, and implementation plans for systematic code modernization
- **Test Infrastructure Enhancements**: Integration test helpers, golden test utilities, and comprehensive test fixtures for improved maintainability

### Fixed
- **Standards Compliance**: Resolved Go linting issues and improved code organization for better maintainability

## [1.2.7] - 2025-08-28

### Fixed
- **Terraform Plan Parsing**:
  - Fixed issue where Strata couldn't properly handle plan files in subdirectories
  - Changed parser to use `filepath.Base()` when executing `terraform show` to ensure the correct relative path is used within the plan's directory context

### Changed
- **Documentation Organization**:
  - Reorganized feature specifications from `agents/` directory to `specs/` directory for better project structure
  - Added new LLM analysis feature specifications with requirements and decision log documentation

## [1.2.6] - 2025-08-13

### Changed
- **Formatter Code Quality**:
  - Refactored formatter indentation to use Unicode En space constants for consistent spacing across all output formats
  - Added `indent` and `nestedIndent` constants to eliminate magic number usage and improve code maintainability
  - Updated all property change formatting to use standardized indentation constants
  - Enhanced test suite to validate consistent Unicode En space formatting across all formatter functions

### Fixed
- **Nested Property Indentation**:
  - Fixed incorrect indentation for nested property changes in formatter output
  - Corrected nested object change formatting to use proper Unicode En space alignment
  - Resolved inconsistent spacing in nested map and array property displays

## [1.2.5] - 2025-08-13

### Added
- **Development Documentation**:
  - Added comprehensive GitHub Copilot instructions for Strata development (`.github/copilot-instructions.md`)
  - Enhanced formatting capabilities for better property change display

### Changed
- **Nested Property Formatting**:
  - Improved nested property formatting with better visual hierarchy for complex objects
  - Enhanced property change display with clearer structure for nested values like maps and arrays
  - Added comprehensive test suite for nested property formatting validation

## [1.2.4] - 2025-08-08

### Added
- **Output Refinements Feature**:
  - Implemented comprehensive output refinements feature addressing issues #17-#21
  - Added `ShowNoOps` field to `PlanConfig` struct with `--show-no-ops` CLI flag for configurable no-op display control
  - Added `IsNoOp` field to `ResourceChange` and `OutputChange` structs for internal no-op resource tracking
  - Implemented `filterNoOps` and `filterNoOpOutputs` methods for filtering based on configuration
  - Enhanced `OutputSummary` method to integrate no-op filtering with "No changes detected" message when appropriate
  - Added alphabetical property sorting (case-insensitive) with natural sort ordering and path hierarchy sorting
  - Implemented immediate sensitive value masking with `maskSensitiveValue` helper function for security by default
  - Enhanced ActionSortTransformer with improved danger detection and table sorting by danger indicators, action priority, then alphabetically

- **Enhanced Testing Infrastructure**:
  - Added comprehensive unit tests in `lib/plan/comparison_consistency_test.go` for comparison consistency validation
  - Added extensive Output Refinements test suite with edge cases, integration tests, and backward compatibility validation
  - Added performance testing with large plans (1000+ resources) to ensure <5% performance impact
  - Enhanced unit tests for statistics behavior with output changes and no-op exclusion logic

- **Build System and Development Tools**:
  - Added comprehensive Makefile targets (test-verbose, test-coverage, benchmarks, security-scan, dependency management)
  - Added sample testing targets (list-samples, run-all-samples) for improved development workflow
  - Added development utilities including go-functions target for code analysis and enhanced help documentation

### Changed
- **Code Quality and API Improvements**:
  - Standardized comparison functions in `lib/plan/analyzer.go` - replaced custom `equals()` with `reflect.DeepEqual()` (58 lines removed)
  - Refactored `hasDangerIndicator` from stateless method to package function for better API design and testability
  - Enhanced `calculateStatistics` method to properly handle output changes with no-op exclusion logic
  - Updated all callers of `calculateStatistics` to pass output changes for comprehensive statistics tracking

- **Documentation and Configuration**:
  - Updated CLAUDE.md with comprehensive project structure, build system documentation, and GitHub Action implementation details
  - Enhanced README.md with version information, detailed example output, and improved installation instructions
  - Updated default strata.yaml configuration with show-no-ops option documentation and proper commenting
  - Updated Claude settings to include mcp__devtools__search_packages in allowlist

### Removed
- **Test Code Cleanup**:
  - Removed disabled memory usage tracking test in performance_integration_test.go marked for future implementation

### Fixed
- **Comparison Function Standardization**:
  - Resolved inconsistent comparison function usage that could potentially cause subtle bugs in change detection
  - Enhanced output analysis to detect outputs with identical before/after values using deep equality comparison

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
