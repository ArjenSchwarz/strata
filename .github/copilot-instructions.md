# Strata: Terraform Plan Analysis Tool - GitHub Copilot Instructions

**ALWAYS follow these instructions first. Only search for additional context if the information here is incomplete or found to be in error.**

Strata is a Go CLI tool that enhances Terraform workflows by providing clear, concise summaries of Terraform plan changes. It parses Terraform plan files and generates comprehensive summaries with danger detection, progressive disclosure, and multiple output formats.

## Working Effectively

### Prerequisites and Setup
- **Go 1.24.5+** is required (project uses Go 1.24.5)
- **Terraform 1.6+** for integration testing (optional for core development)
- **golangci-lint v2** for linting (optional but recommended)

### Core Development Workflow

#### 1. Bootstrap and Build (NEVER CANCEL - Times: 30s first build, <1s subsequent)
```bash
# Build with version information (includes dependency download on first run)
make build                    # Takes ~30 seconds first time, <1 second subsequent builds

# Clean build without artifacts
make clean && make build      # <1 second for clean builds

# Build with specific version
make build VERSION=1.2.3

# Build release version (requires VERSION parameter)
make build-release VERSION=1.2.3
```

#### 2. Testing (NEVER CANCEL - Times: <10s unit tests, <1s action tests)
```bash
# Run all Go unit tests - comprehensive test suite with 85% coverage
make test                     # Takes ~8 seconds, NEVER CANCEL

# Run tests with verbose output and coverage report
make test-verbose             # Takes ~3 seconds, NEVER CANCEL

# Run benchmark tests to verify performance
make benchmarks               # Takes ~13 seconds, NEVER CANCEL

# Generate HTML coverage report
make test-coverage            # Creates coverage.html file

# Test GitHub Action components
make test-action-unit         # Takes <1 second, tests shell modules
make test-action-integration  # Requires Terraform, tests end-to-end workflows
make test-action              # Runs both unit and integration tests
```

#### 3. Code Quality (NEVER CANCEL - Times: <5s total)
```bash
# Format Go code according to standards
make fmt                      # Takes <1 second

# Run static analysis
make vet                      # Takes ~3 seconds

# Run linter (requires golangci-lint v2 - note: v1 will fail with config error)
make lint                     # May fail if wrong golangci-lint version installed

# Run complete validation suite (excluding lint if not available)
make fmt vet test             # Takes ~10 seconds total, NEVER CANCEL

# Full validation including lint (if golangci-lint v2 is available)
make check                    # Takes ~15 seconds total, NEVER CANCEL
```

### Validation Scenarios

#### Core CLI Functionality Testing
Always test your changes with these scenarios to ensure the tool works correctly:

```bash
# List available sample files for testing
make list-samples

# Test basic functionality with simple sample
make run-sample SAMPLE=simpleadd-sample.json

# Test detailed output with complex sample
make run-sample-details SAMPLE=complex-properties-sample.json

# Test danger detection with dangerous changes
make run-sample SAMPLE=danger-sample.json

# Test all samples quickly to verify no regressions
make run-all-samples          # Takes <1 second, tests all 8 sample files

# Test different output formats
./strata plan summary --output json samples/simpleadd-sample.json
./strata plan summary --output markdown samples/danger-sample.json
./strata plan summary --expand-all samples/complex-properties-sample.json
```

#### Manual Validation Scenarios
After making changes, ALWAYS run through these user scenarios to ensure functionality:

1. **Basic Plan Analysis**:
   ```bash
   ./strata plan summary samples/simpleadd-sample.json
   # Verify: Shows plan info, statistics, and resource changes
   ```

2. **Danger Detection**:
   ```bash
   ./strata plan summary samples/danger-sample.json
   # Verify: Shows ⚠️ warnings for dangerous changes
   ```

3. **Progressive Disclosure**:
   ```bash
   ./strata plan summary --expand-all samples/complex-properties-sample.json
   # Verify: All property details are expanded and visible
   ```

4. **Output Formats**:
   ```bash
   ./strata plan summary --output json samples/simpleadd-sample.json | jq .
   ./strata plan summary --output markdown samples/danger-sample.json | head -20
   # Verify: JSON is valid, Markdown renders correctly
   ```

5. **Error Handling**:
   ```bash
   ./strata plan summary nonexistent-file.tfplan
   # Verify: Shows helpful error message and usage information
   ```

6. **Version Information**:
   ```bash
   ./strata --version
   ./strata version
   ./strata version --output json | jq .
   # Verify: Version info displays correctly in all formats
   ```

### Build Times and Timeout Guidelines

**CRITICAL: Set appropriate timeouts and NEVER CANCEL long-running commands**

- **First build**: 30 seconds (includes dependency download) - Set timeout to 60+ seconds
- **Subsequent builds**: <1 second - Set timeout to 30+ seconds  
- **Unit tests**: ~8 seconds - Set timeout to 30+ seconds, NEVER CANCEL
- **Benchmarks**: ~13 seconds - Set timeout to 60+ seconds, NEVER CANCEL
- **Action tests**: <1 second - Set timeout to 30+ seconds
- **All samples**: <1 second - Set timeout to 30+ seconds
- **Code formatting**: <1 second - Set timeout to 30+ seconds
- **Static analysis**: ~3 seconds - Set timeout to 30+ seconds
- **Linting**: ~2 seconds - Set timeout to 30+ seconds (if golangci-lint v2 available)

## Repository Structure

### Key Directories
- **`cmd/`**: CLI command definitions using Cobra framework
  - `root.go`: Root command and global flags
  - `plan.go`: Plan command group  
  - `plan_summary.go`: Plan summary subcommand (main functionality)
  - `version.go`: Version command
- **`lib/plan/`**: Core Terraform plan processing logic
  - `analyzer.go`: Plan analysis and danger detection
  - `formatter.go`: Output formatting for multiple formats
  - `models.go`: Data structures and types
  - `parser.go`: Plan file parsing
- **`config/`**: Configuration management with Viper
- **`samples/`**: Test plan files for development and validation (8 samples available)
- **`test/`**: GitHub Action test scripts and integration tests
- **`docs/`**: Documentation and implementation details
- **`.github/`**: GitHub-specific files including workflows and this instruction file

### Important Files
- **`Makefile`**: Comprehensive build system with all development targets
- **`strata.yaml`**: Default configuration file with expandable sections and danger settings
- **`action.yml`**: GitHub Action definition
- **`action.sh`**: GitHub Action implementation
- **`.golangci.yml`**: Linter configuration (requires golangci-lint v2)

## Common Development Tasks

### Adding New Features
1. **Always build and test first**: `make clean build test`
2. **Add tests** for new functionality in appropriate `*_test.go` files
3. **Test with samples**: Use `make run-sample SAMPLE=<relevant-sample>` 
4. **Validate output formats**: Test table, JSON, and markdown outputs
5. **Run full validation**: `make fmt vet test` before committing

### Debugging Issues
1. **Use verbose testing**: `make test-verbose` to see detailed test output
2. **Test specific samples**: Use samples that trigger the issue you're investigating
3. **Check error handling**: Test with invalid inputs to ensure graceful failures
4. **Verify output**: Always manually inspect the generated output

### Modifying Output Formats
1. **Test all formats**: table, JSON, HTML, markdown
2. **Verify progressive disclosure**: Test both collapsed and expanded (`--expand-all`) modes
3. **Check danger highlighting**: Use `samples/danger-sample.json` to verify warnings
4. **Validate collapsible content**: Ensure expandable sections work correctly

### Performance Considerations
- **Memory limits**: The tool processes large plans efficiently with built-in limits
- **Property analysis**: Comprehensive property change capture with smart truncation
- **Progressive disclosure**: Balances detail availability with clean presentation
- **Benchmark validation**: Always run `make benchmarks` after performance-related changes

## Configuration

The tool uses `strata.yaml` for configuration:

```yaml
output: table
plan:
  show-details: true
  highlight-dangers: true
  expandable_sections:
    enabled: true
    auto_expand_dangerous: true
sensitive_resources:
  - resource_type: aws_db_instance
  - resource_type: aws_instance  
sensitive_properties:
  - resource_type: aws_instance
    property: user_data
```

## GitHub Action Integration

The project includes a comprehensive GitHub Action at `action.yml`:
- **Unit tests**: Validate individual shell modules
- **Integration tests**: End-to-end testing with real Terraform plans  
- **Security measures**: Input validation and sanitization
- **Caching**: Binary caching for improved performance

Test the action components:
```bash
make test-action-unit         # Test shell module functions
make test-action-integration  # Test complete workflows (requires Terraform)
```

## Troubleshooting

### Build Issues
- **Dependency problems**: Run `make deps-tidy` to clean up modules
- **Version issues**: Use `make build-release VERSION=x.y.z` for specific versions
- **Cache issues**: Use `make clean` then rebuild

### Test Failures  
- **Linting failures**: Ensure golangci-lint v2 is installed, not v1
- **Sample failures**: Use `make list-samples` to see available test files
- **Coverage issues**: Run `make test-coverage` for detailed coverage report

### Runtime Issues
- **Config errors**: Check `strata.yaml` syntax and configuration options
- **Plan parsing errors**: Verify plan file is valid JSON or binary Terraform plan
- **Output issues**: Test different output formats to isolate format-specific problems

## Known Working Commands Reference

These commands are validated to work correctly:

```bash
# Build and setup
make build                                    # <1s (after first build)
make clean build                             # <1s  
make deps-tidy                               # Clean dependencies

# Testing (all NEVER CANCEL)
make test                                    # ~8s, comprehensive test suite
make test-verbose                            # ~3s, detailed output  
make benchmarks                              # ~13s, performance validation
make test-action-unit                        # <1s, action component tests

# Code quality  
make fmt                                     # <1s, format code
make vet                                     # ~3s, static analysis
make fmt vet test                            # ~10s, validation without lint

# Sample testing
make list-samples                            # List 8 available samples
make run-sample SAMPLE=simpleadd-sample.json       # Basic functionality
make run-sample SAMPLE=danger-sample.json          # Danger detection  
make run-sample-details SAMPLE=complex-properties-sample.json  # Detailed output
make run-all-samples                         # <1s, test all samples

# CLI usage
./strata --version                           # Version info
./strata version --output json               # JSON version info
./strata plan summary --help                 # Command help
./strata plan summary samples/simpleadd-sample.json            # Basic usage
./strata plan summary --output json samples/simpleadd-sample.json    # JSON output
./strata plan summary --output markdown --expand-all samples/danger-sample.json  # Full markdown
```

**Validation requirement**: Always test at least one complete end-to-end scenario after making changes:
1. Build the tool: `make build`
2. Run a sample: `make run-sample SAMPLE=danger-sample.json`  
3. Verify the output shows plan information, statistics, and properly highlighted dangerous changes
4. Test error handling: `./strata plan summary nonexistent.tfplan`
5. Verify graceful error message with usage information

This ensures your changes maintain the tool's core functionality and user experience.