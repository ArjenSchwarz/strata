# Strata: Terraform Plan Analysis Tool - Claude Development Guide

## Project Overview

Strata is a Go CLI tool that enhances Terraform workflows by providing clear, concise summaries of Terraform plan changes, similar to changeset descriptions in deployment tools. The primary goal is to help users understand the impact of proposed infrastructure changes before applying them.

### Core Features
- Parse and summarize Terraform plan files
- Highlight potentially destructive changes
- Generate statistical summaries of resource modifications
- Support multiple output formats (table, JSON, HTML)
- Integrate with CI/CD pipelines
- Danger highlights for sensitive resources and properties

### Project Status
The project is currently in active development with the following phases:
1. ✅ CLI Foundation & Configuration
2. ✅ Terraform Plan Parsing
3. ✅ Summary Analysis Engine
4. ✅ Output Integration
5. 🔄 Future Integration Preparation

## Technical Stack

- **Language**: Go 1.24.1
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) v1.9.1
- **Configuration**: [Viper](https://github.com/spf13/viper) v1.20.1
- **Output Formatting**: [go-output](https://github.com/ArjenSchwarz/go-output) v1.4.0
- **Terraform Integration**: [terraform-json](https://github.com/hashicorp/terraform-json) v0.25.0

## Project Structure

```
strata/
├── cmd/                  # Command-line interface definitions
│   ├── root.go           # Root command and global flags
│   ├── plan.go           # Plan command group
│   └── plan_summary.go   # Plan summary subcommand
├── config/               # Configuration management
│   └── config.go         # Configuration structures and helpers
├── docs/                 # Documentation
│   ├── implementation/   # Implementation details and design docs
│   └── research/         # Research notes and references
├── lib/                  # Core library code
│   └── plan/             # Terraform plan processing
│       ├── analyzer.go   # Plan analysis logic
│       ├── formatter.go  # Output formatting
│       ├── models.go     # Data structures
│       └── parser.go     # Plan file parsing
├── main.go               # Application entry point
├── strata.yaml           # Default configuration file
└── go.mod                # Go module definition
```

## Architecture Patterns

### Command Pattern
The project follows the command pattern using Cobra:
- Root command (`cmd/root.go`) defines global flags and configuration
- Subcommands (`cmd/plan.go`, `cmd/plan_summary.go`) implement specific functionality
- Commands delegate to library code for core logic

### Layered Architecture
1. **Command Layer** (`cmd/`): Handles CLI interaction, flags, and user input
2. **Library Layer** (`lib/`): Contains core business logic
3. **Configuration Layer** (`config/`): Manages application settings

### Data Flow
```
User Input → Command Layer → Library Layer → Output Formatting → User Display
```

## Development Guidelines

### Prerequisites
- Go 1.24.1 or higher
- Terraform 1.6+ (for testing)

### Build System
```bash
# Build the project
go build -o strata

# Run tests
go test ./...

# Install locally
go install
```

### Common Commands
```bash
# Generate a plan summary from a Terraform plan file
strata plan summary terraform.tfplan

# Generate summary with JSON output
strata plan summary --output json terraform.tfplan

# Generate summary with custom danger threshold
strata plan summary --danger-threshold 5 terraform.tfplan

# Use a custom config file
strata --config custom-config.yaml plan summary terraform.tfplan
```

### Code Style & Conventions

1. **Go Formatting**:
   - Use `gofmt` or `go fmt` to format code
   - Follow standard Go naming conventions (CamelCase for exported, camelCase for unexported)

2. **Error Handling**:
   - Use descriptive error messages with context
   - Wrap errors using `fmt.Errorf("failed to X: %w", err)`
   - Return errors rather than handling them internally when appropriate

3. **Documentation**:
   - Document all exported functions, types, and constants
   - Include examples in documentation where helpful
   - Keep implementation notes in the docs/implementation directory

4. **Testing**:
   - Write unit tests for all core functionality
   - Use table-driven tests where appropriate
   - Aim for high test coverage, especially in critical paths

### Naming Conventions
- **Files**: Use lowercase with underscores for multi-word filenames
- **Packages**: Use lowercase, single-word names
- **Types**: Use CamelCase for type names
- **Functions**: Use CamelCase for exported functions, camelCase for unexported
- **Variables**: Use descriptive names that indicate purpose

## Configuration

Strata uses a YAML configuration file with the following default locations:
- Current directory: `./strata.yaml`
- Home directory: `~/.strata.yaml`

Example configuration:
```yaml
output: table
table:
  style: ColoredBlackOnMagentaWhite
plan:
  danger-threshold: 3
  show-details: true
  highlight-dangers: true
sensitive_resources:
  - resource_type: AWS::RDS::DBInstance
  - resource_type: AWS::EC2::Instance
sensitive_properties:
  - resource_type: AWS::EC2::Instance
    property: UserData
```

## Key Components

### Plan Processing Module (`lib/plan/`)

1. **Parser** (`parser.go`):
   - Loads and validates Terraform plan files
   - Converts raw plan data to internal structures

2. **Analyzer** (`analyzer.go`):
   - Processes plan data to extract meaningful information
   - Categorizes changes and calculates statistics
   - Detects sensitive resources and properties for danger highlights

3. **Formatter** (`formatter.go`):
   - Formats analysis results for display
   - Supports multiple output formats
   - Highlights dangerous changes

4. **Models** (`models.go`):
   - Defines data structures used throughout the module
   - Includes helper methods for data manipulation

### Key Data Structures

```go
type ResourceChange struct {
    Address          string          `json:"address"`
    Type             string          `json:"type"`
    Name             string          `json:"name"`
    ChangeType       ChangeType      `json:"change_type"`
    IsDestructive    bool            `json:"is_destructive"`
    ReplacementType  ReplacementType `json:"replacement_type"`
    // Danger highlights fields
    IsDangerous      bool     `json:"is_dangerous"`
    DangerReason     string   `json:"danger_reason"`
    DangerProperties []string `json:"danger_properties"`
}

type PlanSummary struct {
    FormatVersion    string           `json:"format_version"`
    TerraformVersion string           `json:"terraform_version"`
    ResourceChanges  []ResourceChange `json:"resource_changes"`
    OutputChanges    []OutputChange   `json:"output_changes"`
    Statistics       ChangeStatistics `json:"statistics"`
}
```

## Local Validation Process

Before finishing any task, always run the following commands:

1. `gofmt` - Format Go code according to standards
2. `go test ./... -v` - Run all tests with verbose output
3. `golangci-lint run` - Check for code quality issues
4. Optionally `go build -o strata` - Confirm the project builds

## Development Hooks

The project includes automated hooks for quality assurance:

1. **Go Format and Test Hook**: Automatically runs gofmt, golangci-lint, and go test when Go files are modified
2. **Documentation Update Hook**: Monitors changes to Go source files and suggests documentation updates

## Features

### Danger Highlights
A key feature that allows users to define sensitive resource types and properties in the config file that trigger warnings when modified:

1. **Sensitive Resources**: Warn when sensitive resources are replaced (e.g., RDS instances)
2. **Sensitive Properties**: Warn when sensitive properties are updated (e.g., EC2 UserData)

This feature integrates with the existing plan analysis workflow and highlights potential risks in infrastructure deployments.

### Change Classification
- **Create**: New resources being added
- **Update**: Existing resources being modified
- **Delete**: Resources being removed
- **Replace**: Resources being destroyed and recreated
- **No-op**: No changes to resources

### Statistical Analysis
- Counts of different change types
- Identification of destructive changes
- Danger threshold warnings

## Testing Strategy

- Unit tests are placed alongside the code they test
- Test files are named with `_test.go` suffix
- Table-driven tests are preferred for testing multiple scenarios
- Focus on high test coverage, especially in critical paths

## Documentation Philosophy

Documentation is aimed at aiding understanding rather than being overly concise. When new functionality is created, ensure the README file is updated to include relevant information for users.

## License

Strata is licensed under the MIT License.