# Technical Stack & Development Guidelines

## Tech Stack

### Core Dependencies
- **Language**: Go 1.24.1
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) v1.9.1
- **Configuration**: [Viper](https://github.com/spf13/viper) v1.20.1
- **Output Formatting**: [go-output](https://github.com/ArjenSchwarz/go-output) v1.4.0
- **Terraform Integration**: [terraform-json](https://github.com/hashicorp/terraform-json) v0.25.0
- **Testing**: [testify](https://github.com/stretchr/testify) v1.10.0

### Additional Dependencies
- **Home Directory**: [go-homedir](https://github.com/mitchellh/go-homedir) v1.1.0
- **AWS SDK**: aws-sdk-go-v2 (for context placeholders)
- **Table Formatting**: [go-pretty](https://github.com/jedib0t/go-pretty) v6.4.9
- **Color Output**: [fatih/color](https://github.com/fatih/color) v1.16.0

## Development Environment

### Prerequisites
- Go 1.24.1 or higher
- Terraform 1.6+ (for testing and sample data)
- golangci-lint (for code quality)
- make (for build automation)

### Build System

The project uses a Makefile for standardized build operations:

```bash
# Build with version information
make build

# Build release version (requires VERSION)
make build-release VERSION=1.2.3

# Run all tests
make test

# Run linting
make lint

# Test with sample data
make run-sample

# Run sample with detailed output
make run-sample-details

# Format code
make fmt

# Clean build artifacts
make clean
```

### Manual Build Commands

```bash
# Standard build
go build -o strata

# Build with version injection
go build -ldflags "-X github.com/ArjenSchwarz/strata/cmd.Version=1.2.3 \
                   -X github.com/ArjenSchwarz/strata/cmd.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
                   -X github.com/ArjenSchwarz/strata/cmd.GitCommit=$(git rev-parse --short HEAD)" .

# Run tests
go test ./...

# Install locally
go install
```

## Common Usage Patterns

### Basic Commands

```bash
# Generate plan summary
strata plan summary terraform.tfplan

# JSON output
strata plan summary --output json terraform.tfplan

# Custom danger threshold
strata plan summary --danger-threshold 5 terraform.tfplan

# File output with different format
strata plan summary --file report.html --file-format html terraform.tfplan

# Dynamic file naming
strata plan summary --file "report-$TIMESTAMP-$AWS_REGION.json" --file-format json terraform.tfplan

# Version information
strata --version
strata version
strata version --output json
```

### Configuration Usage

```bash
# Use custom config file
strata --config custom-config.yaml plan summary terraform.tfplan

# Debug mode
strata --debug plan summary terraform.tfplan

# Verbose output
strata --verbose plan summary terraform.tfplan
```

## Code Style & Conventions

### Go Standards
- Use `gofmt` or `go fmt` for consistent formatting
- Follow standard Go naming conventions
- Use CamelCase for exported identifiers, camelCase for unexported
- Add comprehensive documentation for all exported functions and types

### Error Handling Patterns
- Use descriptive error messages with context
- Wrap errors using `fmt.Errorf("operation failed: %w", err)`
- Return errors rather than handling them internally
- Validate inputs early with clear error messages

### Testing Standards
- Write unit tests for all core functionality
- Use table-driven tests for multiple scenarios
- Aim for high test coverage, especially in critical paths
- Include integration tests for CLI commands
- Test error conditions and edge cases

### Documentation Requirements
- Document all exported functions, types, and constants
- Include usage examples in documentation
- Keep implementation notes in docs/implementation/
- Update README.md for user-facing changes

## Configuration System

### File Locations
Strata uses YAML configuration with the following precedence:
1. `--config` flag specified file
2. `./strata.yaml` (current directory)
3. `~/.strata.yaml` (home directory)

### Configuration Structure
```yaml
# Output settings
output: table
table:
  style: ColoredBlackOnMagentaWhite

# Plan analysis settings
plan:
  danger-threshold: 3
  show-details: true
  highlight-dangers: true
  always-show-sensitive: true

# File output settings
output-file: "reports/plan-$TIMESTAMP.json"
output-file-format: json

# Sensitivity configuration
sensitive_resources:
  - resource_type: aws_db_instance
  - resource_type: aws_ec2_instance

sensitive_properties:
  - resource_type: aws_instance
    property: user_data
```

## Architecture Patterns

### Command Pattern with Cobra
- Root command defines global flags and configuration
- Subcommands implement specific functionality
- Commands delegate to library code for core logic
- Clear separation between CLI and business logic

### Layered Architecture
1. **Command Layer** (`cmd/`): CLI interaction and user input
2. **Library Layer** (`lib/`): Core business logic organized by domain
3. **Configuration Layer** (`config/`): Application settings management

### Data Flow
```
User Input → Command Parsing → Configuration Loading → 
Library Processing → Output Formatting → Display/File Output
```

## Version Management

### Build-time Injection
Version information is injected at build time using ldflags:
- `Version`: Semantic version string
- `BuildTime`: ISO 8601 timestamp
- `GitCommit`: Short Git commit hash
- `GoVersion`: Go compiler version

### Version Commands
- `strata --version`: Quick version check
- `strata version`: Detailed version information
- `strata version --output json`: Machine-readable version data

## File Output System

### Features
- Dual output: simultaneous stdout and file output
- Format flexibility: different formats for stdout and file
- Dynamic naming: placeholder support for contextual file names
- Security: path traversal protection and validation

### Placeholders
- `$TIMESTAMP`: Current timestamp (2006-01-02T15-04-05)
- `$AWS_REGION`: AWS region from context
- `$AWS_ACCOUNTID`: AWS account ID from context

## GitHub Action Integration

The project includes a GitHub Action that:
- Analyzes Terraform plans in CI/CD workflows
- Posts summaries to GitHub step summaries
- Comments on pull requests with plan analysis
- Supports multiple output formats and configurations
- Handles caching and error scenarios gracefully