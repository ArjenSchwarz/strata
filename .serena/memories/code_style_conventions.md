# Strata Code Style and Conventions

## Go Language Standards

### Formatting
- Use `gofmt` or `go fmt` for standard Go formatting
- Enforce formatting with `make fmt` before commits
- Standard Go indentation (tabs)

### Naming Conventions
- **Packages**: lowercase, single-word names (e.g., `plan`, `config`)
- **Files**: lowercase with underscores for multi-word (e.g., `plan_summary.go`)
- **Types**: CamelCase for type names (e.g., `ResourceChange`)
- **Functions**: CamelCase for exported, camelCase for unexported
- **Variables**: descriptive names indicating purpose (e.g., `changeType`, not just `ct`)
- **Constants**: CamelCase or ALL_CAPS depending on scope

### Code Patterns
- Use `any` instead of `interface{}`
- Use `comparable` for type constraints
- Prefer modern Go idioms (min/max functions, strings.CutPrefix, etc.)
- Use struct tags for JSON serialization

### Error Handling
- Always handle errors explicitly
- Use descriptive error messages with context
- Wrap errors: `fmt.Errorf("failed to parse plan: %w", err)`
- Return errors rather than panicking
- Log errors at appropriate levels

### Documentation
- Document all exported functions, types, and constants
- Include examples in documentation where helpful
- Use clear, concise comments
- Keep implementation notes in docs/implementation/

### Testing Conventions
- Write unit tests for all core functionality
- Place test files alongside code (*_test.go)
- Use table-driven tests for multiple scenarios
- Aim for high test coverage in critical paths
- Test names should be descriptive (TestAnalyzer_ProcessPlan_WithDangerousChanges)
- Use testify for assertions when needed

## Project-Specific Patterns

### Data Structures
- Use JSON struct tags for serialization
- Include omitempty for optional fields
- Use internal fields with `json:"-"` tag when needed

### Command Pattern
- Commands delegate to library code
- Keep business logic in lib/, not cmd/
- Use Cobra for command structure

### Configuration
- YAML-based configuration files
- Support multiple config locations
- Use Viper for configuration management

## Code Quality Tools

### Linting
- golangci-lint with project-specific config
- Enabled linters: govet, ineffassign, misspell, staticcheck, revive, goconst, gocritic, unconvert
- Run with `make lint`

### Static Analysis
- Use `go vet` for basic static analysis
- Security scanning with gosec (optional)

## Modern Go Features
- Use Go 1.24.5 features where appropriate
- Leverage generics when beneficial
- Use context for cancellation and timeouts
- Prefer io.Reader/Writer interfaces

## Import Organization
Standard library imports first, then third-party, then local packages:
```go
import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    
    "github.com/ArjenSchwarz/strata/lib/plan"
)
```