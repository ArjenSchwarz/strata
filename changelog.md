# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
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
- Enhanced ChangeStatistics struct with Replacements and Conditionals fields
- Updated ResourceChange struct to include ReplacementType field
- Replacement necessity analysis in analyzer with proper categorisation
- Horizontal statistics summary table with TOTAL, ADDED, REMOVED, MODIFIED, REPLACEMENTS, CONDITIONALS columns
- Enhanced statistics calculation separating definite and conditional replacements
- Comprehensive test coverage for replacement analysis and statistics formatting
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
