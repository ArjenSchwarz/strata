# Agent Development Guidelines

## Tools & Commands

1. Always use available tools when possible
2. Never include `cd .` when running terminal commands
3. Use the Makefile targets for common operations:
   - `make build` - Build with version information
   - `make test` - Run all tests
   - `make run-sample` - Test with sample files
   - `make lint` - Run golangci-lint

## Local Validation

Before completing any task, run these validation steps in order:

1. **Format Code**: `gofmt -w .` or `make fmt`
2. **Run Tests**: `go test ./... -v` or `make test`
3. **Lint Code**: `golangci-lint run` or `make lint`
4. **Build Check**: `make build` to confirm project builds with version info
5. **Sample Test**: `make run-sample` to verify functionality with test data

## Code Style & Standards

### Go Best Practices
- Prefer simple, readable solutions over complex ones
- Use `gofmt` for consistent formatting
- Follow standard Go naming conventions (CamelCase for exported, camelCase for unexported)
- Add comprehensive comments for all exported functions, types, and constants
- Keep functions focused, testable, and with single responsibilities
- Use table-driven tests for comprehensive test coverage

### Error Handling
- Use descriptive error messages with context
- Wrap errors using `fmt.Errorf("failed to X: %w", err)`
- Return errors rather than handling them internally when appropriate
- Validate inputs early and provide clear error messages

### Project-Specific Patterns
- Follow the layered architecture: Command → Library → Output
- Use the existing configuration system (Viper + YAML)
- Maintain separation between CLI commands and core logic
- Use the go-output library for consistent formatting
- Follow the established patterns in lib/plan/ for new functionality

## Documentation Requirements

### Code Documentation
- Document all exported functions, types, and constants with Go doc comments
- Include usage examples in documentation where helpful
- Keep implementation notes in docs/implementation/ directory
- Update README.md for any new user-facing functionality

### User Documentation
- Update README.md with new features, commands, or configuration options
- Include practical examples and use cases
- Document configuration options in the example strata.yaml
- Add help text to CLI commands using Cobra's help system

### Testing Documentation
- Write comprehensive unit tests for all core functionality
- Use table-driven tests for testing multiple scenarios
- Include integration tests for CLI commands
- Document test scenarios and expected behaviors

## Version Management

- Version information is injected at build time via ldflags
- Use `make build` to build with proper version information
- Version command should display version, build time, Git commit, and Go version
- Follow semantic versioning for releases

## Integration Considerations

- Maintain compatibility with existing Terraform workflows
- Ensure GitHub Action integration continues to work
- Test file output functionality with various formats
- Validate configuration file parsing and validation
- Consider CI/CD pipeline integration requirements