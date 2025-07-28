# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Configuration Migration System**:
  - Added `MigrateDeprecatedConfig()` function for graceful handling of configuration format changes
  - Automatic migration of old grouping threshold settings to new nested configuration structure
  - Default value provisioning for new configuration sections (expandable_sections, grouping)
  - User-friendly warnings about new features available in updated configuration format
  - Comprehensive test coverage for migration scenarios including edge cases and validation
- **Enhanced Documentation and Help Text**:
  - Expanded plan summary command help with detailed progressive disclosure explanations
  - Added comprehensive examples for --expand-all flag usage across output formats
  - Documentation of provider grouping behavior and risk analysis features
  - File output examples with dynamic naming and format combinations
  - Updated Claude development guide with latest architectural changes and testing procedures
  - Task completion tracking for enhanced summary visualization feature implementation
- **Progressive Disclosure Integration**:
  - Complete integration of provider grouping with collapsible sections in formatter
  - Enhanced resource changes formatting with conditional grouping logic based on provider diversity
  - Intelligent fallback from grouped to standard progressive disclosure for single-provider plans
  - Sensitive resource filtering with progressive disclosure for always-show-sensitive mode
  - Updated formatter to use new collapsible document building instead of legacy table-only approach
- **Comprehensive Test Suite for Enhanced Summary Visualization**:
  - End-to-end integration tests covering complete flow from plan parsing to formatted output with 5 comprehensive test scenarios
  - Error handling tests for malformed Terraform plans, graceful degradation, memory limits, and circular dependency detection
  - Performance validation tests including benchmarks for small/medium/large plans (10/100/1000 resources) with performance targets (<100ms/<1s/<10s)
  - Memory usage verification tests ensuring system stays under 500MB limit for large plans
  - Test fixtures with realistic Terraform plan JSON data for simple, multi-provider, high-risk, and dependency scenarios
  - Collapsible formatter performance tests comparing overhead vs simple display (max 3x allowed)
- **GitHub Action expand-all Support**:
  - Added `expand-all` input parameter to GitHub Action configuration with default `false`
  - GitHub Action now supports `expand-all: true` to expand all collapsible sections in PR comments
  - Updated action.sh to validate and sanitize expand-all boolean input parameter
  - Enhanced lib/action/strata.sh to pass `--expand-all` flag through to strata command
  - Added INPUT_EXPAND_ALL environment variable mapping in action configuration
  - Proper parameter flow: action.yml → action.sh → strata.sh → strata command
  - Maintains backward compatibility - default behavior unchanged when expand-all not specified
  - Added comprehensive logging to track expand-all parameter through execution chain
- **Global Expand-All Flag Support**:
  - Added `--expand-all` / `-e` CLI flag to expand all collapsible sections globally
  - CLI flag bound to Viper configuration system with automatic config file override
  - Enhanced plan summary command to load expandable sections, grouping, and performance limits configuration
  - Implemented `createOutputWithConfig()` function using go-output v2 `RendererConfig` with `ForceExpansion` for proper global expansion control
  - Updated collapsible formatters to use renderer-level global expansion instead of custom logic
  - Property changes formatter auto-expands sensitive properties, with global expansion handled by `ForceExpansion`
  - Dependencies formatter respects global expansion through renderer configuration
  - Comprehensive test coverage for expand-all functionality with property changes, dependencies, and output configuration tests
  - Supports Table, Markdown, HTML, and CSV formats with collapsible-enabled renderers
- **Provider Grouping with Collapsible Sections**:
  - `groupByProvider()` function for smart resource grouping by provider with configurable thresholds
  - `formatGroupedWithCollapsibleSections()` function for provider-based resource organization using go-output v2 sections
  - `hasHighRiskChanges()` helper function for auto-expansion of high-risk provider groups
  - `getResourceTableSchema()` for consistent table schema across grouped displays
  - Smart grouping logic that only activates with multiple providers and sufficient resource count (default: 10)
  - Provider diversity checking to skip grouping when all resources are from same provider
  - Comprehensive unit test coverage with 7 test cases for grouping logic and 6 test cases for risk detection
  - Integration with existing collapsible formatters (property changes, dependencies) within provider sections
- **Comprehensive Unit Tests for Enhanced Formatters**:
  - Unit tests for `propertyChangesFormatter()` with sensitive and non-sensitive property change scenarios
  - Unit tests for `dependenciesFormatter()` with various dependency patterns (depends-on, used-by, no dependencies)
  - Unit tests for `prepareResourceTableData()` with mixed resource types and risk assessment validation
  - Unit tests for `formatResourceChangesWithProgressiveDisclosure()` main formatting function
  - Error handling tests for invalid input types and edge cases
  - Property change details formatting tests with sensitive data masking validation
  - Fixed existing formatter tests for go-output v2 API compatibility
  - Integration tests for go-output v2 component interaction and document creation
- **go-output v2 Collapsible Content Integration**:
  - Fixed go-output v2 API usage with correct format constants (`output.Table.Name` instead of `output.FormatTable`)
  - Implemented `propertyChangesFormatter()` for collapsible property changes with sensitive data auto-expansion
  - Implemented `dependenciesFormatter()` for collapsible dependency information display
  - Created `prepareResourceTableData()` function to transform ResourceChange data for v2 table display
  - Added `formatResourceChangesWithProgressiveDisclosure()` main formatting function using v2 document builder
  - Comprehensive collapsible API reference documentation for go-output v2 integration

### Fixed
- Compilation errors in formatter.go due to incorrect go-output v2 format constant usage
- ActionSortTransformer now uses correct v2 format constants for proper format detection
- Analyzer nil plan handling - Added plan loading fallback in GenerateSummary method to prevent nil pointer errors

### Changed
- Claude settings - Added MCP context7 documentation tools and copy command permissions for enhanced development capabilities

### Added
- Core analysis functions for enhanced summary visualization:
  - `analyzePropertyChanges()` function with comprehensive property change extraction (no 3-property limit)
  - `assessRiskLevel()` function with 4-tier risk assessment (critical, high, medium, low)  
  - `extractDependenciesWithLimit()` function for basic resource dependency extraction
  - `AnalyzeResource()` function integrating all analysis components with performance limits
  - Depth-limited recursive comparison for nested properties (max depth: 5)
  - Memory tracking and truncation safeguards (10MB total limit, 100 properties/resource)
  - Value size estimation for performance monitoring
  - Comprehensive unit test coverage for all new analysis functions (16 new test cases)
- New agent documentation files for enhanced summary visualization feature planning
- Configuration support for enhanced summary visualization (group-by-provider, grouping-threshold, show-context options)  
- Extended ResourceChange model with Provider, TopChanges, and ReplacementHints fields
- Provider extraction and caching functionality in plan analyzer
- Enhanced context extraction for resource replacement reasons using Terraform ReplacePaths data
- Property change detection for update operations showing first 3 changed properties
- Context-aware danger evaluation with descriptive resource-specific and property-specific reasons
- Comprehensive test coverage for new configuration options and data models (47 new test cases)
- Global `expand_all` configuration option and CLI flag for collapsible sections
- Property change truncation with size limits for performance optimization
- Implementation status tracking in tasks documentation
- API clarification notes for go-output v2 implementation
- Core data models for progressive disclosure with go-output v2 integration:
  - `ResourceAnalysis` struct for comprehensive resource change analysis
  - `PropertyChangeAnalysis` struct for detailed property change tracking with truncation support
  - `PropertyChange` struct for individual property changes with before/after values and sensitive data handling
  - `DependencyInfo` struct for resource dependency relationships
- Enhanced configuration structures:
  - `ExpandableSectionsConfig` for collapsible sections behavior control
  - `GroupingConfig` for enhanced provider grouping configuration
  - `PerformanceLimitsConfig` with memory and processing limits (100 properties/resource, 1MB property size, 100MB total memory)
  - `GetPerformanceLimitsWithDefaults()` helper method for configuration validation
- Comprehensive unit tests for new data models with JSON serialization, sensitive data handling, and backward compatibility validation

### Enhanced
- Simplified design approach for enhanced summary visualization feature
- Streamlined risk assessment to basic risk levels (low, medium, high, critical)
- Simplified data structures for property change analysis
- Enhanced GitHub Actions integration with expandable sections

### Changed
- Enhanced plan analyzer with smart grouping capabilities and context extraction
- Improved data models to support provider-based resource grouping and change context
- Updated configuration structure from `progressive_disclosure` to `expandable_sections`
- Removed complex risk assessment components (mitigation suggestions, detailed scoring)
- Modified property change analysis to focus on essential information with performance limits
- Updated API usage examples for go-output v2 compatibility
- Replaced basic danger reason logic with enhanced evaluation system providing specific reasons for:
  - Database replacements (RDS instances, database clusters)
  - Compute instance replacements (EC2, Azure VMs)
  - Storage replacements (S3 buckets, storage accounts)
  - Security rule replacements (security groups, firewalls)
  - Network infrastructure replacements (VPCs, networks)
  - Credential changes (passwords, secrets)
  - Authentication key changes (API keys, tokens)
  - User data modifications
  - Security configuration changes
- All deletion operations are now considered risky by default with appropriate danger messaging

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
