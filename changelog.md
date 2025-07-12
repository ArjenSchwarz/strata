# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Code Quality and Linting Improvements
  * Enhanced golangci-lint configuration with revive linter settings for exported function documentation
  * Added comprehensive package documentation comments across all cmd/ packages
  * Improved error handling with proper blank identifier usage for ignored return values
  * Refactored conditional logic to use switch statements for better readability
  * Updated type definitions with detailed documentation comments for error codes and model types
  * Replaced interface{} with any type for better Go 1.18+ compatibility
  * Removed unused functions and improved code organization
  * Enhanced constant definitions with proper documentation and consistent naming
  * Improved function parameter handling with blank identifiers for unused parameters

### Added
- Enhanced Error Handling and Recovery System
  * Added comprehensive error recovery mechanisms with context-aware suggestions
  * Added intelligent error pattern analysis for authentication, permissions, network, and timeout issues
  * Added provider-specific error handling for AWS, Azure, and GCP
  * Added enhanced error messages with detailed recovery actions and troubleshooting steps
  * Added user guidance system with interactive error recovery suggestions
  * Added structured error context with workflow information and command details

- Improved Resource Cleanup and Management
  * Added automatic temporary file cleanup with deferred cleanup handlers
  * Added comprehensive resource tracking during workflow execution
  * Added graceful cleanup on both success and failure scenarios
  * Added cleanup validation and error reporting for failed cleanup operations
  * Added support for multiple resource types (files, directories, etc.)

- Enhanced Terraform Executor Reliability
  * Added robust pipe error handling with system resource suggestions
  * Added enhanced command start error detection with specific recovery guidance
  * Added intelligent output analysis for plan and apply failures
  * Added timeout handling with context-specific suggestions based on operation type
  * Added comprehensive plan file validation and creation error handling
  * Added output truncation for better error context management

- Comprehensive Test Coverage
  * Added extensive unit tests for error recovery mechanisms
  * Added test coverage for temporary resource cleanup functionality
  * Added tests for enhanced error analysis and pattern matching
  * Added validation tests for different error scenarios and edge cases
  * Added mock-based testing for external dependencies

- Terraform Workflow Command
  * Added `strata apply` command for complete Terraform plan and apply workflow
  * Implemented interactive workflow with plan analysis, user prompts, and apply execution
  * Added comprehensive error handling with user-friendly messages and recovery suggestions
  * Added CI/CD environment detection with automatic non-interactive mode
  * Added support for destructive change warnings and confirmation prompts
  * Added configuration support for Terraform workflow settings
  * Added audit logging for CI/CD environments with structured output
  * Added real-time output streaming during Terraform command execution
  * Added timeout handling and progress indicators for long-running operations
  * Added backend detection and validation for remote state management
  * Added comprehensive test coverage for workflow components

- Enhanced Error Handling System
  * Added comprehensive StrataError type with structured error information
  * Added error builders for common Terraform and workflow scenarios
  * Added context information, suggestions, and recovery actions for all errors
  * Added error classification (user errors, system errors, critical errors)
  * Added user-friendly error message formatting with emojis and structured output
  * Added proper exit code handling for different error types

- Terraform Execution Engine
  * Added TerraformExecutor interface and implementation for command execution
  * Added real-time output streaming with progress indicators
  * Added version detection and installation validation
  * Added backend configuration detection and validation
  * Added comprehensive output parsing for plan and apply results
  * Added timeout handling and cancellation support

- Configuration Enhancements
  * Added TerraformConfig section with path, arguments, and backend settings
  * Added BackendConfig for remote state configuration
  * Added validation for all configuration values with helpful error messages
  * Added support for environment variables and timeout settings

- Development Infrastructure
  * Added golangci-lint configuration with comprehensive linter rules
  * Added agent workflow documentation with design, requirements, and tasks
  * Added invalid configuration test file for validation testing

### Changed
- GitHub Action Modularization
  - Refactored monolithic action.sh into modular architecture with separate library modules
  - Created lib/action/ directory structure with specialized modules:
    - utils.sh: Core utility functions (logging, error handling, JSON parsing, downloads)
    - security.sh: Input validation, sanitization, and security checks
    - files.sh: Temporary file management and secure cleanup operations
    - binary.sh: Strata binary downloading, caching, and compilation
    - strata.sh: Strata execution functions and dual output processing
    - github.sh: GitHub integration (PR comments, step summaries, output distribution)
  - Enhanced main action.sh to orchestrate modular components with cleaner separation of concerns
  - Improved code maintainability and testability through modular design
  - Added comprehensive test suite with modular function testing
  - Updated test runner with enhanced reporting and requirements coverage analysis
  - Enhanced Claude development settings with additional bash command permissions

### Added
- File Output System
  - Added comprehensive file output functionality with dual output support (stdout + file)
  - Added file output validation system with security checks for path traversal prevention
  - Added placeholder support for dynamic file naming ($TIMESTAMP, $AWS_REGION, $AWS_ACCOUNTID)
  - Added --file and --file-format global flags for specifying output file and format
  - Added FileValidator with comprehensive validation for file paths, formats, and permissions
  - Added file output configuration options in config files (output-file, output-file-format)
  - Added extensive unit tests for file validation and placeholder resolution
  - Added integration tests for file output functionality
  - Updated README.md with comprehensive file output documentation and examples
  - Updated command help text with file output usage examples and placeholder documentation
  - Enhanced config system with placeholder resolution for AWS context and timestamps
  - Added security features including path sanitization and directory permission validation

### Added
- GitHub Release Automation
  - Added automated release workflow for building and publishing cross-platform binaries
  - Added support for Linux, Windows, and macOS on both amd64 and arm64 architectures
  - Implemented build-time version injection with git commit, build time, and version information
  - Added automatic binary publishing to GitHub releases with LICENSE and README.md files
  - Enhanced Claude development settings with additional GitHub CLI permissions

### Added
- Version Information System
  - Added version command with detailed version information display
  - Added --version flag support to root command
  - Implemented build-time version injection via ldflags
  - Added VersionInfo struct with version, build time, git commit, and Go version
  - Added JSON output format support for version command
  - Added comprehensive version helper functions with graceful handling of missing information
  - Enhanced Makefile with version injection support for builds
  - Added build-release target requiring explicit version specification
  - Updated README.md with version information usage examples and build instructions
  - Added comprehensive unit tests for version functionality
  - Removed unused toggle flag from root command
- Minor Code Improvements
  - Fixed loop variable usage in analyzer.go for better code clarity
- GitHub Action Integration
  - Created composite GitHub Action for Terraform plan analysis in CI/CD workflows
  - Added action.yml with comprehensive input/output definitions
  - Added action.sh with binary management, caching, and execution logic
  - Implemented automatic binary download with platform detection and checksum verification
  - Added fallback compilation from source when binary download fails
  - Integrated GitHub Step Summary with rich Markdown formatting
  - Added pull request comment functionality with update/create logic
  - Implemented GitHub API integration with retry logic and rate limiting
  - Added comprehensive error handling and input validation
  - Created extensive test suite with unit and integration tests
  - Added GitHub workflow files for automated testing
  - Updated README.md with detailed GitHub Action usage documentation
- Enhanced Makefile with additional development targets
  - Added test-action-unit and test-action-integration targets
  - Added fmt, lint, clean, install, and help targets
  - Improved development workflow with comprehensive target documentation
- Enhanced error handling in formatter
  - Added ValidateOutputFormat method to validate output format before processing
  - Added input validation for all formatter methods
  - Added comprehensive error messages with proper context
  - Added unit tests for error handling scenarios
- Updated high-risk column documentation to clarify that ALL dangerous resources are counted

### Added
- UI Improvements
  - Added horizontal plan information layout for better readability
  - Added high-risk column in statistics summary to highlight sensitive dangerous changes
  - Added always-show-sensitive feature to display critical changes even when details are disabled
  - Added markdown output format support for documentation and pull requests
  - Added comprehensive documentation for all new UI features
- Added Claude settings file with appropriate permissions for development tasks

### Added
- Added comprehensive project documentation
  - Added CLAUDE.md with detailed development guide
  - Expanded README.md with usage instructions and feature documentation
  - Added sample documentation from fog project for reference
- Added project logo in docs/images directory
- Added Makefile with standard development targets
  - Added build target for compiling the application
  - Added test target for running all tests
  - Added run-sample target for executing sample files with plan summary command
  - Added run-sample-details target for executing sample files with detailed output
- Added test_makefile.sh script for validating Makefile targets
- Reorganized sample files into dedicated samples directory
  - Moved danger-sample.json, k8ssample.json, and websample.json to samples directory
- Added sample Terraform plan JSON file (danger-sample.json) demonstrating sensitive resource replacements and property changes
- Danger Highlights feature for identifying sensitive resource replacements and property changes
  - Enhanced equals function in analyzer to properly handle slice comparisons
  - Implementation of sensitive resource detection to flag replacements of critical infrastructure
  - Implementation of sensitive property detection to identify risky property changes
  - Added IsSensitiveResource and IsSensitiveProperty methods to the analyzer
  - Added checkSensitiveProperties method to detect property changes
  - Extended ResourceChange model with danger-related fields
- Configuration options for defining sensitive resources and properties
- Detection of sensitive resource replacements with warning indicators
- Detection of sensitive property changes with detailed property listings
- Enhanced output formatting with danger information in resource changes table
- Documentation for the Danger Highlights feature
- Unit tests for sensitive resource and property detection
- Initial CLI foundation with Cobra framework
- `plan summary` command for Terraform plan analysis
- Configuration management for output formats and danger thresholds
- Support for multiple output formats (table, json, html)
- Danger threshold configuration for highlighting destructive changes
- Command-line flags for customising plan summary behaviour
- Terraform plan file parsing with support for both binary and JSON formats
- Plan data models for resource and output changes
- Plan analysis engine with change categorisation and statistics
- Destructive change detection and warning system
- Comprehensive output formatting using go-output library
- Table, JSON, and HTML output formats with proper go-output integration
- Unified formatter that leverages go-output's built-in capabilities
- Enhanced table output with icons, colours, and structured data presentation
- Comprehensive implementation plan for fog-inspired changeset display functionality
- Documentation for implementing plan information display and enhanced resource changes table
- Phased development approach with detailed checklists for systematic implementation
- **Phase 1: Plan Information Display** - Added comprehensive plan context information display
- Plan information section showing plan file, Terraform version, workspace, backend, creation time, and dry run status
- Enhanced data models with BackendInfo struct and extended PlanSummary fields
- Parser helper methods for extracting workspace, backend, and file information
- Formatter integration for plan information display with proper go-output styling
- Plan information section displays before statistics with proper spacing
- **Phase 2: Enhanced Statistics Summary Table** - Added horizontal statistics table matching fog's format
- ReplacementType enum with Never, Conditional, and Always values for replacement analysis
- Enhanced ChangeStatistics struct with Replacements and Conditionals fields with comprehensive field documentation
- Updated ResourceChange struct to include ReplacementType field
- Replacement necessity analysis with ReplacePaths parsing from Terraform plans
- Advanced replacement type determination distinguishing between definite and conditional replacements
- Helper methods for analyzing conditional replacement paths and computed values
- Horizontal statistics summary table with TOTAL, ADDED, REMOVED, MODIFIED, REPLACEMENTS, CONDITIONALS columns
- Enhanced statistics calculation properly separating definite and conditional replacements
- New configuration options for statistics display control (ShowStatisticsSummary, StatisticsSummaryFormat)
- Added --stats-format command flag for controlling summary format (horizontal/vertical)
- Added --show-statistics command flag for toggling statistics display
- Updated command help text with examples for new flags
- Comprehensive unit tests for replacement analysis covering ReplacePaths parsing and type determination
- Integration tests for statistics display with various change combinations and output format compatibility
- Full backward compatibility maintained with existing functionality
- All output formats (table, json, html) support the new enhanced statistics
- **Phase 3: Enhanced Resource Changes Table** - Added detailed resource changes table with physical IDs, replacement indicators, and module information
- Enhanced ResourceChange struct with PhysicalID, PlannedID, ModulePath, and ChangeAttributes fields
- Physical ID extraction from Terraform plan before/after states with proper handling for new/deleted resources
- Module path parsing from resource addresses with support for nested module hierarchies
- Enhanced resource changes table with ACTION, RESOURCE, TYPE, ID, REPLACEMENT, MODULE columns
- Proper display formatting for different change types (Add, Modify, Remove, Replace, No-op)
- ID display logic showing "-" for new resources, actual IDs for existing resources, and "N/A" for deleted resources
- Module path extraction and formatting for clear hierarchy display
- Complete fog-inspired changeset display implementation with all three phases integrated

### Changed
- Updated root command description to reflect Strata's purpose as a Terraform helper tool
- Refactored output formatting to use go-output library consistently
- Simplified formatter implementation to leverage library capabilities
- Renamed ToReplace field to Replacements in ChangeStatistics for clarity
- Enhanced OutputSummary method to include both plan information and horizontal statistics

### Dependencies
- Added github.com/ArjenSchwarz/go-output v1.4.0 for output formatting
- Added github.com/hashicorp/terraform-json v0.25.0 for Terraform plan parsing
- Added github.com/spf13/cobra v1.9.1 for CLI framework
- Added github.com/spf13/viper v1.20.1 for configuration management
- Added github.com/mitchellh/go-homedir v1.1.0 for home directory detection

## [0.1.0] - 2025-05-24

### Added
- Initial project setup
- Basic Go module structure
- MIT License
