# Strata Development Commands

## Build Commands
```bash
# Build with version info (dev version)
make build

# Build release version
make build-release VERSION=1.2.3

# Install locally
make install

# Clean build artifacts
make clean
```

## Testing Commands
```bash
# Run all Go unit tests
make test

# Run tests with verbose output and coverage
make test-verbose

# Generate HTML coverage report
make test-coverage

# Run benchmark tests
make benchmarks

# Test GitHub Action components
make test-action          # Run all action tests
make test-action-unit     # Unit tests for shell modules
make test-action-integration  # Integration tests
```

## Code Quality Commands
```bash
# Format Go code
make fmt

# Run static analysis
make vet

# Run linter (requires golangci-lint)
make lint

# Run full validation suite (fmt, vet, lint, test)
make check

# Security scanning (requires gosec)
make security-scan
```

## Development Commands
```bash
# Run Strata with a sample file
make run-sample SAMPLE=filename.json

# Run with detailed output
make run-sample-details SAMPLE=filename.json

# List available sample files
make list-samples

# Test all sample files
make run-all-samples

# Direct execution
go run . plan summary terraform.tfplan
go run . plan summary --output json terraform.tfplan
go run . plan summary --expand-all terraform.tfplan
```

## Dependency Management
```bash
# Clean up go.mod and go.sum
make deps-tidy

# Update dependencies to latest
make deps-update
```

## Git Commands (Darwin/macOS)
```bash
git status
git diff
git add .
git commit -m "message"
git push
git log --oneline -10
```

## System Commands (Darwin-specific)
```bash
# List files (with color on macOS)
ls -la

# Find files
find . -name "*.go"

# Search in files (with color)
grep -r "pattern" .

# Check Go version
go version

# Run tests with race detector
go test -race ./...
```

## Project Execution
```bash
# After building
./strata plan summary terraform.tfplan
./strata plan summary --output json terraform.tfplan
./strata plan summary --expand-all terraform.tfplan
./strata --config custom.yaml plan summary terraform.tfplan
./strata version
./strata --version
```

## Important Makefile Targets
```bash
make help         # Show all available targets
make go-functions # List all Go functions in project
make update-v1-tag # Update v1 tag for GitHub Action
```