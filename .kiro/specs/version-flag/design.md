# Design Document

## Overview

The version flag feature will add standard version display functionality to the Strata CLI tool. This will be implemented using Cobra's built-in version support combined with Go's build-time variable injection. The design follows standard CLI conventions and integrates with the existing command structure and output formatting system.

## Architecture

The version functionality will be implemented through two main approaches:

1. **Global Version Flag**: Using Cobra's built-in `--version` flag on the root command
2. **Version Subcommand**: A dedicated `version` command for more detailed output and formatting options

### Component Integration

```
Root Command (cmd/root.go)
├── Version Flag (--version) → Display version and exit
└── Version Subcommand (cmd/version.go) → Detailed version info with formatting options
```

## Components and Interfaces

### Version Variable

A package-level variable will store the version information:

```go
// In cmd/root.go or separate version.go file
var (
    Version   = "dev"     // Set via ldflags during build
    BuildTime = "unknown" // Set via ldflags during build  
    GitCommit = "unknown" // Set via ldflags during build
)
```

### Root Command Enhancement

The root command will be enhanced with:
- Version string assignment using `rootCmd.Version`
- Custom version template for consistent formatting

### Version Subcommand

A new `version` subcommand will provide:
- Detailed version information display
- Support for multiple output formats (table, JSON)
- Integration with existing output formatting system

### Build Integration

Version information will be injected at build time using Go's ldflags:

```bash
go build -ldflags "-X github.com/ArjenSchwarz/strata/cmd.Version=1.2.3 \
                   -X github.com/ArjenSchwarz/strata/cmd.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
                   -X github.com/ArjenSchwarz/strata/cmd.GitCommit=$(git rev-parse HEAD)"
```

## Data Models

### Version Information Structure

```go
type VersionInfo struct {
    Version   string `json:"version"`
    BuildTime string `json:"build_time,omitempty"`
    GitCommit string `json:"git_commit,omitempty"`
    GoVersion string `json:"go_version"`
}
```

### Output Formats

1. **Human-readable** (default):
   ```
   strata version 1.2.3
   Built: 2025-01-15T10:30:00Z
   Commit: abc123def456
   Go: go1.24.1
   ```

2. **JSON format**:
   ```json
   {
     "version": "1.2.3",
     "build_time": "2025-01-15T10:30:00Z",
     "git_commit": "abc123def456",
     "go_version": "go1.24.1"
   }
   ```

## Error Handling

- **Missing version information**: Display "dev" or "unknown" for development builds
- **Invalid output format**: Return error with supported format list
- **JSON marshaling errors**: Return error with descriptive message

The version functionality should be robust and never cause the application to crash, even with missing build information.

## Testing Strategy

### Unit Tests

1. **Version Display Tests**:
   - Test version string formatting
   - Test JSON output generation
   - Test handling of missing version information

2. **Command Integration Tests**:
   - Test `--version` flag behavior
   - Test `version` subcommand execution
   - Test output format selection

3. **Build Integration Tests**:
   - Test version injection during build process
   - Test default values when no version is provided

### Test Data

- Mock version information for consistent testing
- Test cases for various version string formats
- Edge cases for missing or malformed version data

## Implementation Notes

### Cobra Integration

- Use `rootCmd.Version` to set the version string for the `--version` flag
- Use `rootCmd.SetVersionTemplate()` to customize version output format
- Create separate `version` subcommand for extended functionality

### Output Formatting

- Leverage existing output formatting patterns from the project
- Support the same output formats as other commands (table, JSON)
- Maintain consistency with the application's overall CLI design

### Build Process

- Update Makefile to include version injection
- Document build process for maintainers
- Ensure version information is available in development builds