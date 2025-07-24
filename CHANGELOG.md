# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Table style configuration support with `--table-style` flag and `table.style` config option
- Table max column width configuration with `--table-max-column-width` flag and `table.max-column-width` config option
- UNMODIFIED column to statistics summary showing no-op resource changes
- Support for handling plans with no resource changes (displays "All resources are unchanged")
- Sample file `nochange-sample.json` for testing no-change scenarios
- Test documentation output file `output.md`

### Changed
- Upgraded go-output dependency from v2.0.4 to v2.0.5
- Statistics table header from "TOTAL" to "TOTAL CHANGES" for clarity
- Improved handling of no-op changes - now tracked separately and excluded from detailed resource lists
- Enhanced table formatting with file format style support
- **GitHub Action Improvements**
  - Modified path handling to use relative paths within GITHUB_WORKSPACE
  - Enhanced error handling with proper variable exporting
  - Improved workspace directory navigation for better compatibility

### Removed
- Redundant destructive changes warning from plan summary command (now handled by formatter)
- File extension validation for plan and config files in GitHub Action (allows any file type)

### Changed
- **Code Quality Improvements**
  - Added golangci-lint configuration for consistent code quality standards
  - Updated error handling to properly check and handle return values
  - Improved memory management by properly closing files and removing temporary test files
  - Enhanced version command output formatting with proper error checking
  - Added proper documentation comments for exported functions and types
  - Updated to use modern Go constructs (any instead of interface{})
  - Improved workspace and backend information extraction from Terraform state
  - Enhanced conditional replacement analysis for resource changes

### Fixed
- **Error Handling and Resource Management**
  - Fixed unchecked return values in file operations and environment variable settings
  - Improved test cleanup by properly handling file operations and environment restoration
  - Enhanced viper flag binding with proper error checking
  - Fixed potential resource leaks in parser backend detection logic

## [1.0.0] - TBA

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

### Changed
- **Go-Output v2 Migration**
  - Migrated to go-output v2 for improved thread safety and performance
  - Updated all formatter methods to use v2-compatible data structures
  - Implemented separate rendering for stdout and file outputs
  - Enhanced error handling with proper context support

- **GitHub Action Improvements**
  - Modularized action architecture with separate library modules
  - Enhanced output headers with workflow and job context
  - Removed redundant headers and improved formatting
  - Updated test configurations and workflows

### Dependencies
- github.com/ArjenSchwarz/go-output v2.0.0 for enhanced output formatting
- github.com/hashicorp/terraform-json v0.25.0 for Terraform plan parsing
- github.com/spf13/cobra v1.9.1 for CLI framework
- github.com/spf13/viper v1.20.1 for configuration management
- github.com/mitchellh/go-homedir v1.1.0 for home directory detection

## [0.1.0] - 2025-05-24

### Added
- Initial project setup
- Basic Go module structure
- MIT License
