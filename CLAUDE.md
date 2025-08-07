# Strata: Terraform Plan Analysis Tool - Claude Development Guide

## Project Overview

Strata is a Go CLI tool that enhances Terraform workflows by providing clear, concise summaries of Terraform plan changes, similar to changeset descriptions in deployment tools. The primary goal is to help users understand the impact of proposed infrastructure changes before applying them.

### Core Features
- Parse and summarize Terraform plan files
- Highlight potentially destructive changes
- Generate statistical summaries of resource modifications
- Support multiple output formats (table, JSON, HTML, Markdown)
- Integrate with CI/CD pipelines
- Danger highlights for sensitive resources and properties
- Progressive disclosure with collapsible sections
- Provider grouping for large plans
- Global expand-all control for comprehensive details

### Project Status
The project is currently in active development with the following phases:
1. ✅ CLI Foundation & Configuration
2. ✅ Terraform Plan Parsing
3. ✅ Summary Analysis Engine
4. ✅ Output Integration
5. ✅ Enhanced Summary Visualization (Progressive Disclosure)
6. 🔄 Future Integration Preparation

## Technical Stack

- **Language**: Go 1.24.1
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) v1.9.1
- **Configuration**: [Viper](https://github.com/spf13/viper) v1.20.1
- **Output Formatting**: [go-output](https://github.com/ArjenSchwarz/go-output) v2.1.0
- **Terraform Integration**: [terraform-json](https://github.com/hashicorp/terraform-json) v0.25.0

## Project Structure

```
strata/
├── cmd/                  # Command-line interface definitions
│   ├── root.go           # Root command and global flags
│   ├── plan.go           # Plan command group
│   ├── plan_summary.go   # Plan summary subcommand
│   └── version.go        # Version command
├── config/               # Configuration management
│   ├── config.go         # Configuration structures and helpers
│   └── validation.go     # Configuration validation
├── docs/                 # Documentation
│   ├── implementation/   # Implementation details and design docs
│   ├── research/         # Research notes and references
│   └── github-action.md  # GitHub Action documentation
├── lib/                  # Core library code
│   ├── action/           # GitHub Action shell modules
│   │   ├── binary.sh     # Binary management
│   │   ├── files.sh      # File operations
│   │   ├── github.sh     # GitHub integration
│   │   ├── security.sh   # Security validations
│   │   ├── strata.sh     # Strata execution
│   │   └── utils.sh      # Utility functions
│   └── plan/             # Terraform plan processing
│       ├── analyzer.go   # Plan analysis logic
│       ├── formatter.go  # Output formatting
│       ├── models.go     # Data structures
│       └── parser.go     # Plan file parsing
├── samples/              # Test plan files for development
├── test/                 # Test scripts and integration tests
├── testdata/            # Test data for unit tests
├── agents/              # Feature development documentation
├── action.yml           # GitHub Action definition
├── action.sh            # GitHub Action entry point
├── Makefile             # Build and development tasks
├── main.go              # Application entry point
├── strata.yaml          # Default configuration file
└── go.mod               # Go module definition
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

The project uses a comprehensive Makefile for development tasks. Key targets include:

```bash
# Build targets
make build                    # Build with version info
make build-release VERSION=1.2.3  # Build release version
make install                  # Install the application
make clean                    # Clean build artifacts and coverage files

# Testing targets
make test                     # Run Go unit tests
make test-verbose             # Run tests with verbose output and coverage
make test-coverage            # Generate test coverage report (HTML)
make benchmarks               # Run benchmark tests
make test-action              # Run GitHub Action tests
make test-action-unit         # Run GitHub Action unit tests
make test-action-integration  # Run GitHub Action integration tests

# Code quality targets
make fmt                      # Format Go code
make vet                      # Run go vet for static analysis
make lint                     # Run linter (requires golangci-lint)
make check                    # Run full validation suite (fmt, vet, lint, test)
make security-scan            # Run security analysis (requires gosec)

# Sample testing targets
make list-samples             # List available sample files
make run-sample SAMPLE=<filename>         # Run sample plan file
make run-sample-details SAMPLE=<filename> # Run sample with details
make run-all-samples          # Test all sample files quickly

# Dependency management
make deps-tidy                # Clean up go.mod and go.sum
make deps-update              # Update dependencies to latest versions

# Development utilities
make go-functions             # List all Go functions in the project
make update-v1-tag            # Update v1 tag for GitHub Action
make help                     # Show all available targets with descriptions
```

Alternative direct Go commands:
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

# Generate summary with all collapsible sections expanded
strata plan summary --expand-all terraform.tfplan

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
# Global expand control for collapsible sections
expand_all: false                    # Expand all collapsible sections (default: false)

output: table
table:
  style: ColoredBlackOnMagentaWhite

plan:
  show-details: true
  highlight-dangers: true

  # Enhanced summary visualization settings
  expandable_sections:
    enabled: true                    # Enable collapsible sections (default: true)
    auto_expand_dangerous: true      # Auto-expand high-risk sections (default: true)
    show_dependencies: true          # Show dependency information (default: true)

  grouping:
    enabled: true                    # Enable provider grouping (default: true)
    threshold: 10                    # Minimum resources to trigger grouping (default: 10)

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
   - Performs comprehensive property change analysis
   - Extracts resource dependencies and relationships
   - Performs risk assessment with automatic prioritization

3. **Formatter** (`formatter.go`):
   - Formats analysis results for display
   - Supports multiple output formats with collapsible content
   - Highlights dangerous changes with auto-expansion
   - Implements provider grouping for large plans
   - Uses go-output v2 progressive disclosure features

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
    // Enhanced summary visualization fields
    Provider         string           `json:"provider,omitempty"`
    TopChanges       []string         `json:"top_changes,omitempty"`
    ReplacementHints []string         `json:"replacement_hints,omitempty"`
}

type ResourceAnalysis struct {
    PropertyChanges     PropertyChangeAnalysis `json:"property_changes"`
    ReplacementReasons  []string              `json:"replacement_reasons"`
    RiskLevel          string                 `json:"risk_level"`
    Dependencies       DependencyInfo         `json:"dependencies"`
}

type PropertyChangeAnalysis struct {
    Changes    []PropertyChange `json:"changes"`
    Count      int             `json:"count"`
    TotalSize  int             `json:"total_size_bytes"`
    Truncated  bool            `json:"truncated"`
}

type PropertyChange struct {
    Name      string      `json:"name"`
    Path      []string    `json:"path"`
    Before    interface{} `json:"before"`
    After     interface{} `json:"after"`
    Sensitive bool        `json:"sensitive"`
    Size      int         `json:"size"`
}

type DependencyInfo struct {
    DependsOn []string `json:"depends_on"`
    UsedBy    []string `json:"used_by"`
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

**Using Makefile (recommended):**
1. `make check` - Run full validation suite (fmt, vet, lint, test)
2. `make build` - Confirm the project builds

**Individual validation steps:**
1. `make fmt` - Format Go code according to standards
2. `make vet` - Run go vet for static analysis
3. `make test` - Run all Go tests
4. `make lint` - Check for code quality issues (requires golangci-lint)

**Alternative direct commands:**
1. `gofmt ./...` - Format Go code according to standards
2. `go test ./... -v` - Run all tests with verbose output
3. `golangci-lint run` - Check for code quality issues
4. `go build -o strata` - Confirm the project builds

**For GitHub Action changes:**
- `make test-action` - Run both unit and integration tests for GitHub Action
- `make test-action-unit` - Run unit tests for Action shell modules  
- `make test-action-integration` - Run integration tests with sample plans

## Development Hooks

The project includes automated hooks for quality assurance:

1. **Go Format and Test Hook**: Automatically runs gofmt, golangci-lint, and go test when Go files are modified
2. **Documentation Update Hook**: Monitors changes to Go source files and suggests documentation updates

## Features

### Enhanced Summary Visualization
A comprehensive feature set that provides progressive disclosure and enhanced organization of Terraform plan information:

#### Progressive Disclosure
1. **Collapsible Sections**: Uses go-output v2's collapsible content APIs to provide comprehensive information without overwhelming the primary view
2. **Property Change Analysis**: All property changes are captured and made available in expandable sections
3. **Auto-Expansion**: High-risk changes and sensitive properties automatically expand for immediate visibility
4. **Cross-Format Adaptation**: Collapsible content adapts to each output format (Markdown, HTML, Table, JSON)

#### Provider Grouping
1. **Smart Grouping**: Automatically groups resources by provider when plans meet configurable thresholds (default: 10 resources)
2. **Provider Diversity Check**: Only groups when multiple providers are present - skips grouping for single-provider plans
3. **Risk-Based Expansion**: Provider groups with high-risk changes automatically expand

#### Global Expand Control
1. **CLI Flag**: `--expand-all` flag to expand all collapsible sections at once
2. **Configuration Support**: `expand_all` setting in strata.yaml for persistent behavior
3. **Override Capability**: CLI flag overrides configuration file setting

#### Danger Highlights
A key feature that allows users to define sensitive resource types and properties in the config file that trigger warnings when modified:

1. **Sensitive Resources**: Warn when sensitive resources are replaced (e.g., RDS instances)
2. **Sensitive Properties**: Warn when sensitive properties are updated (e.g., EC2 UserData)
3. **Risk Assessment**: Automated risk level calculation (critical, high, medium, low)
4. **Auto-Expansion**: High-risk changes automatically expand in collapsible sections

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

### Go Unit Tests
- Unit tests are placed alongside the code they test
- Test files are named with `_test.go` suffix
- Table-driven tests are preferred for testing multiple scenarios
- Focus on high test coverage, especially in critical paths
- Use `make test` or `go test ./...` to run all Go tests

### GitHub Action Tests
The project includes comprehensive testing for the GitHub Action:

#### Unit Tests (`make test-action-unit`)
- Tests individual shell modules in `lib/action/`
- Validates security measures, input sanitization, and utility functions
- Tests binary management and file operations
- Located in `test/action_test.sh`

#### Integration Tests (`make test-action-integration`)
- End-to-end tests with real Terraform plan files
- Tests complete workflow from plan file to output generation
- Validates GitHub-specific integrations (PR comments, step summaries)
- Uses sample plan files from `samples/` directory
- Located in `test/integration_test.sh`

#### Test Infrastructure
- Comprehensive test scripts in `test/` directory
- Sample Terraform plan files in `samples/` for development and testing
- Test data for unit tests in `testdata/` directory
- Makefile targets for easy test execution

## Documentation Philosophy

Documentation is aimed at aiding understanding rather than being overly concise. When new functionality is created, ensure the README file is updated to include relevant information for users.

## GitHub Action

Strata provides a GitHub Action for seamless CI/CD integration:

### Action Definition (`action.yml`)
The action accepts the following inputs:
- `plan-file`: Path to Terraform plan file (required)
- `output-format`: Output format (table, json, markdown) - default: markdown
- `config-file`: Path to custom Strata config file (optional)
- `show-details`: Show detailed change information - default: false
- `expand-all`: Expand all collapsible sections - default: false
- `github-token`: GitHub token for PR comments - default: `${{ github.token }}`
- `comment-on-pr`: Whether to comment on PR - default: true
- `update-comment`: Update existing comment vs create new - default: true
- `comment-header`: Custom header for PR comments - default: "🏗️ Terraform Plan Summary"

### Action Outputs
- `summary`: Plan summary text
- `has-changes`: Whether the plan contains changes
- `has-dangers`: Whether dangerous changes were detected
- `json-summary`: Full summary in JSON format
- `change-count`: Total number of changes
- `danger-count`: Number of dangerous changes

### Usage Example
```yaml
- name: Analyze Terraform Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
    output-format: markdown
    show-details: true
```

### Action Implementation (`action.sh`)
The action is implemented with modular shell scripts:
- **Security-first design** with input validation and sanitization
- **Modular architecture** with separate modules for different concerns
- **Comprehensive error handling** with structured output
- **GitHub integration** for PR comments and step summaries
- **Binary caching** for improved performance

## License

Strata is licensed under the MIT License.