# Strata Codebase Structure

## Project Root Structure
```
strata/
├── cmd/                  # Command-line interface definitions
│   ├── root.go          # Root command and global flags
│   ├── plan.go          # Plan command group
│   ├── plan_summary.go  # Plan summary subcommand
│   └── version.go       # Version command
├── config/              # Configuration management
│   ├── config.go        # Configuration structures and helpers
│   └── validation.go    # Configuration validation
├── docs/                # Documentation
│   ├── implementation/  # Implementation details and design docs
│   ├── research/        # Research notes and references
│   └── github-action.md # GitHub Action documentation
├── lib/                 # Core library code
│   ├── action/          # GitHub Action shell modules
│   └── plan/            # Terraform plan processing
│       ├── analyzer.go  # Plan analysis logic
│       ├── formatter.go # Output formatting
│       ├── models.go    # Data structures
│       └── parser.go    # Plan file parsing
├── samples/             # Test plan files for development
├── test/                # Test scripts and integration tests
├── testdata/            # Test data for unit tests
├── specs/               # Feature specifications and development docs
├── action.yml           # GitHub Action definition
├── action.sh            # GitHub Action entry point
├── Makefile            # Build and development tasks
├── main.go             # Application entry point
├── strata.yaml         # Default configuration file
├── go.mod              # Go module definition
├── .golangci.yml       # Linter configuration
└── CLAUDE.md           # Development guide for Claude AI

## Key Components

### Command Layer (cmd/)
- Handles CLI interaction using Cobra framework
- Defines global flags and configuration
- Implements subcommands for specific functionality

### Library Layer (lib/plan/)
- **parser.go**: Loads and validates Terraform plan files
- **analyzer.go**: Processes plan data, categorizes changes, detects sensitive resources
- **formatter.go**: Formats analysis results, supports multiple output formats
- **models.go**: Core data structures (ResourceChange, PlanSummary, etc.)

### Configuration (config/)
- Manages application settings using Viper
- Validates configuration files
- Handles default and custom config locations

### Testing
- Unit tests alongside Go code (*_test.go files)
- Integration test scripts in test/ directory
- Sample Terraform plans in samples/
- Test data in testdata/

### GitHub Action (lib/action/)
- Modular shell scripts for GitHub integration
- Security-first design with input validation
- Binary management and caching
- PR comment and step summary generation

## Specs Directory
Active feature development with specifications in specs/:
- enhanced-summary-visualization
- terraform-unknown-values-and-outputs
- code-cleanup-and-modernization
- output-refinements
- llm-analysis
- simplified-plan-rendering