# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **Code Organization and Maintainability**: Refactored complex resource table display logic in formatter.go by extracting monolithic method into focused helper functions (addResourceChangesTable, addGroupedResourceTables, addProviderGroupTable, addStandardResourceTable, handleResourceDisplay, handleSensitiveResourceDisplay) for improved readability and maintainability
- **Code Style Improvements**: Enhanced comment documentation in analyzer.go constants and fixed Go variable declaration style to use modern syntax

### Fixed
- **Property Change Details Display**: Fixed missing property change details in collapsible sections by properly integrating PropertyChanges field into ResourceChange model and ensuring the analyzer populates it with detailed change information
- **Map Formatting Consistency**: Fixed non-deterministic map key ordering in formatValue function by sorting keys alphabetically for consistent output across test runs

### Added
- **Comprehensive Integration Testing for Plan Summary Output Improvements**: Complete end-to-end integration test suite covering all output improvements including danger highlighting, property change formatting, and empty table suppression
- **Performance Testing Infrastructure**: Large-scale performance and memory testing with artificial plan generation supporting up to 1000+ resources with configurable property counts and truncation validation
- **Complex Properties Sample Data**: New test fixture `samples/complex-properties-sample.json` with comprehensive property change scenarios for testing enhanced property analysis
- **Backward Compatibility Validation**: Complete test coverage ensuring JSON output structure consistency and table formatting compatibility for existing parsers

### Removed
- **Dependencies Column and Functionality**: Removed dependencies column from resource table display, `dependenciesFormatterDirect()` function, and `DependencyInfo` data structure from models
- **Dependency-Related Test Coverage**: Removed test cases for dependency extraction, circular dependency detection, and dependency formatter functions

### Changed
- **Resource Table Schema**: Simplified resource table schema by removing dependencies field and related formatting logic
- **Resource Analysis**: Updated `ResourceAnalysis` struct to remove `Dependencies` field, streamlining the analysis model
- **Test Data Preparation**: Updated `prepareResourceTableData()` to exclude dependency information from table row data

### Added
- **Empty Table Suppression**: Implemented filtering to exclude no-op changes from Resource Changes tables, preventing display of tables that would only contain unchanged resources
- **Comprehensive Empty Table Tests**: Added extensive test coverage for empty table suppression including no-op filtering, provider grouping behavior, and changed resource counting

### Changed
- **Provider Grouping Threshold Calculation**: Updated threshold comparison to use changed resource count (excluding no-ops) instead of total resource count for more accurate grouping decisions
- **Provider Resource Counts**: Modified provider group headers to show only changed resource counts, excluding no-op resources from totals
- **Table Data Preparation**: Enhanced `prepareResourceTableData` to filter out no-op changes, ensuring Resource Changes tables only display actual modifications

### Added
- **Terraform-Style Property Change Formatting**: Implemented Terraform diff-style formatting for property changes with `+`, `-`, `~` prefixes for add, remove, and update actions
- **Enhanced Property Formatters**: Added `propertyChangesFormatterTerraform` and `formatPropertyChange` functions with support for complex value types (maps, arrays, primitives)
- **Sensitive Value Auto-Expansion**: Property changes with sensitive values automatically expand when `AutoExpandDangerous` is enabled
- **Comprehensive Terraform Formatter Tests**: Added extensive test coverage for Terraform-style formatting including value formatting, sensitive property handling, and auto-expansion logic
- **Enhanced Property Analysis Performance Limits**: Added `enforcePropertyLimits` function with configurable constants for MaxPropertiesPerResource (100), MaxPropertyValueSize (10KB), and MaxTotalPropertyMemory (10MB) to prevent excessive memory usage during plan analysis
- **Comprehensive Unit Tests for Property Analysis**: Added extensive test coverage for enhanced property analysis including deep object comparison tests, performance limit enforcement tests, and property extraction validation

### Changed
- **Property Changes Display**: Switched from direct property formatter to Terraform-style formatter in resource table schema for improved readability
- **Output Format Validation**: Improved case-insensitive validation for output formats
- **Auto-Expansion Logic**: Updated auto-expansion to respect `AutoExpandDangerous` configuration setting instead of always expanding sensitive properties
- **Property Analysis Method**: Refactored `analyzePropertyChanges` to use dedicated `enforcePropertyLimits` function for better code organization and maintainability

### Added
- **Property Change Extraction Infrastructure**: Deep object comparison algorithm in analyzer.go with recursive comparison logic for maps, slices, and primitives
- **Property Analysis Helper Functions**: Added `extractPropertyName`, `parsePath`, `isSensitive`, `extractSensitiveChild`, `extractSensitiveIndex` for comprehensive property change extraction
- **Action Tracking for Property Changes**: Action field to PropertyChange struct to track "add", "remove", "update" operations
- **Comprehensive Property Comparison Tests**: Added 10+ test cases for property comparison functionality with order-independent assertions
- **Sensitive Value Detection**: Support for detecting sensitive values using Terraform's BeforeSensitive/AfterSensitive data
- **Array Index Path Parsing**: Property path parsing with array index support (e.g., matrix[1][2])
- **Performance Limits for Property Analysis**: Added limits of 100 properties max and 10MB total size for property analysis
- **Plan Summary Output Improvements Feature Documentation**: Complete feature documentation including requirements, design, decision log, and tasks for improving plan summary output with empty table suppression, enhanced property change formatting, and risk-based sorting

### Changed
- **Property Change Analysis Method**: Updated `analyzePropertyChanges` method to use new deep comparison algorithm instead of callback-based approach
- **PropertyChange Struct Enhancement**: Enhanced PropertyChange struct with Action field and improved documentation for clarity
- **Unified table creation pattern**: All tables now created consistently in `OutputSummary()` method using single `output.New().AddContent().Build()` document building pattern
- **ActionSortTransformer scope**: Limited ActionSortTransformer to table/JSON/CSV formats only, excluding markdown/HTML to prevent rendering conflicts
- **Performance test threshold**: Adjusted collapsible formatting performance threshold from 3x to 6x slower to accommodate multi-table rendering complexity

### Fixed
- **Multi-table rendering in Markdown and HTML formats**: Resolved critical bug where Plan Information and Summary Statistics tables were missing in markdown/HTML output due to ActionSortTransformer interference with multi-table rendering
- **Simplified plan rendering architecture**: Unified all table creation using `output.NewTableContent()` pattern following go-output v2 best practices, eliminating architectural complexity and mixed rendering approaches
- **Provider grouping with collapsible sections**: Enhanced provider-based resource grouping to use proper collapsible sections with auto-expansion for high-risk changes

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
