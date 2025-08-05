# Project Structure & Organization

## Directory Structure

```
strata/
├── .github/              # GitHub workflows and templates
├── .kiro/                # Kiro IDE configuration and steering
├── agent/                # GitHub Action implementation
│   └── terraform-workflow/
├── cmd/                  # Command-line interface definitions
│   ├── root.go           # Root command and global flags
│   ├── plan.go           # Plan command group
│   ├── plan_summary.go   # Plan summary subcommand
│   ├── version.go        # Version command implementation
│   └── version_test.go   # Version command tests
├── config/               # Configuration management
│   ├── config.go         # Configuration structures and helpers
│   ├── validation.go     # Configuration validation logic
│   └── validation_test.go # Configuration validation tests
├── docs/                 # Documentation
│   ├── github-action.md  # GitHub Action documentation
│   ├── images/           # Documentation images and assets
│   ├── implementation/   # Implementation details and design docs
│   └── research/         # Research notes and references
├── lib/                  # Core library code
│   ├── action/           # GitHub Action specific logic
│   ├── errors/           # Error handling utilities
│   ├── plan/             # Terraform plan processing
│   ├── terraform/        # Terraform integration utilities
│   └── workflow/         # Workflow and automation helpers
├── samples/              # Sample Terraform plans for testing
├── test/                 # Integration tests and test utilities
├── test_debug/           # Debug test artifacts
├── main.go               # Application entry point
├── strata.yaml           # Default configuration file
├── Makefile              # Build automation
├── action.yml            # GitHub Action definition
├── action.sh             # GitHub Action script
└── go.mod                # Go module definition
```

## Architecture Patterns

### Command Pattern with Cobra

The project follows the command pattern using Cobra CLI framework:

- **Root Command** (`cmd/root.go`):
  - Defines global flags (`--config`, `--debug`, `--verbose`, `--file`, `--file-format`)
  - Handles configuration loading and validation
  - Sets up logging and output systems

- **Command Groups** (`cmd/plan.go`):
  - Organizes related commands under logical groups
  - Provides command-specific help and documentation

- **Subcommands** (`cmd/plan_summary.go`, `cmd/version.go`):
  - Implement specific functionality with focused responsibilities
  - Delegate core logic to library modules
  - Handle user input validation and error reporting

### Layered Architecture

The project uses a clean layered architecture with clear separation of concerns:

1. **Presentation Layer** (`cmd/`):
   - Handles CLI interaction, argument parsing, and user input
   - Manages output formatting and display
   - Delegates business logic to library layer

2. **Business Logic Layer** (`lib/`):
   - Contains core application logic organized by domain
   - Implements plan analysis, configuration management, and workflow automation
   - Provides reusable components for different use cases

3. **Data Layer** (`config/`, file I/O):
   - Manages configuration loading, validation, and persistence
   - Handles file operations and data serialization
   - Provides abstraction over external data sources

### Data Flow Architecture

```
User Input → CLI Parsing → Configuration Loading → 
Business Logic Processing → Output Formatting → 
Display/File Output + Error Handling
```

## Module Organization

### Plan Processing (`lib/plan/`)

The plan processing module follows domain-driven design principles:

1. **Models** (`models.go`):
   - Defines core data structures and types
   - Includes helper methods and validation logic
   - Provides serialization and deserialization capabilities

2. **Parser** (`parser.go`):
   - Loads and validates Terraform plan files (binary and JSON)
   - Converts raw plan data to internal structures
   - Handles different plan file formats and versions

3. **Analyzer** (`analyzer.go`):
   - Processes plan data to extract meaningful insights
   - Categorizes changes and calculates statistics
   - Implements danger detection and sensitivity analysis

4. **Formatter** (`formatter.go`):
   - Formats analysis results for different output types
   - Supports multiple output formats (table, JSON, HTML, markdown)
   - Handles file output with dynamic naming and placeholders

### Configuration Management (`config/`)

Configuration management follows a validation-first approach:

1. **Configuration Structures** (`config.go`):
   - Defines configuration data structures with validation tags
   - Provides default values and environment variable support
   - Implements configuration merging and precedence rules

2. **Validation Logic** (`validation.go`):
   - Validates configuration values and constraints
   - Provides detailed error messages for invalid configurations
   - Supports custom validation rules for complex scenarios

### Error Handling (`lib/errors/`)

Centralized error handling with context and categorization:
- Custom error types for different failure scenarios
- Error wrapping with contextual information
- Structured error reporting for CLI and programmatic use

### GitHub Action Integration (`lib/action/`, `agent/`)

GitHub Action support with workflow automation:
- Action metadata and input/output definitions
- Integration with GitHub API for PR comments and summaries
- Caching and performance optimization for CI/CD environments

## Naming Conventions

### File and Directory Naming
- **Files**: Use lowercase with underscores for multi-word filenames (`plan_summary.go`)
- **Directories**: Use lowercase, single-word names when possible (`cmd`, `lib`, `config`)
- **Test Files**: Use `_test.go` suffix for test files

### Go Code Naming
- **Packages**: Use lowercase, single-word names that reflect functionality
- **Types**: Use CamelCase for type names (`PlanSummary`, `ConfigOptions`)
- **Functions**: Use CamelCase for exported functions, camelCase for unexported
- **Variables**: Use descriptive names that indicate purpose and scope
- **Constants**: Use CamelCase or UPPER_CASE depending on scope and usage

### Configuration and Documentation
- **Config Files**: Use lowercase with hyphens (`strata.yaml`, `github-action.md`)
- **Documentation**: Use descriptive names with appropriate extensions
- **Sample Files**: Use descriptive prefixes (`sample_`, `test_`, `debug_`)

## Testing Strategy

### Test Organization
- **Unit Tests**: Placed alongside the code they test with `_test.go` suffix
- **Integration Tests**: Located in `test/` directory for cross-module testing
- **Sample Data**: Stored in `samples/` directory for consistent testing

### Test Patterns
- **Table-Driven Tests**: Preferred for testing multiple scenarios and edge cases
- **Test Fixtures**: Reusable test data and helper functions
- **Mock Objects**: For testing external dependencies and error conditions

### Test Coverage
- Focus on critical paths and business logic
- Include error conditions and edge cases
- Validate configuration parsing and validation
- Test CLI commands and output formatting

## Build and Deployment

### Build System (`Makefile`)
- Standardized build targets for development and release
- Version injection using ldflags for build-time metadata
- Automated testing and quality checks
- Sample data testing and validation

### GitHub Action (`action.yml`, `action.sh`)
- Self-contained action definition with inputs and outputs
- Shell script wrapper for cross-platform compatibility
- Integration with GitHub workflow ecosystem
- Caching and performance optimization

### Release Management
- Semantic versioning with build metadata
- Automated version injection during build process
- Release artifacts and distribution management
- Documentation and changelog maintenance