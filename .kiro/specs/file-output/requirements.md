# File Output System Requirements Document

## Overview

This document specifies the requirements for implementing a dual-output system that allows CLI tools to simultaneously write output to stdout and save formatted output to files. This system provides flexible file output with format conversion, placeholder support, and validation.

## Core Requirements

### 1. Global Flag Definition

The system SHALL provide two persistent global flags available to all subcommands:

- `--file`: Optional file path to save output in addition to stdout
- `--file-format`: Optional format override for file output (defaults to same as stdout format)

```go
// Example flag registration
rootCmd.PersistentFlags().String("file", "", "Optional file to save the output to, in addition to stdout")
rootCmd.PersistentFlags().String("file-format", "", "Optional format for the file, defaults to the same as output")
```

### 2. Configuration Integration

The flags SHALL be bound to a configuration system (e.g., Viper) to allow:
- Command-line flag overrides
- Configuration file defaults
- Environment variable support

```go
// Example configuration binding
viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("file"))
viper.BindPFlag("output-file-format", rootCmd.PersistentFlags().Lookup("file-format"))
```

### 3. Output Settings Structure

The system SHALL implement an output settings structure that encapsulates:

```go
type OutputSettings struct {
    OutputFormat     string  // Primary output format (table, json, csv, etc.)
    OutputFile       string  // File path for saving output
    OutputFileFormat string  // Format override for file output
    // ... other output configuration
}

func (config *Config) NewOutputSettings() *OutputSettings {
    settings := &OutputSettings{
        OutputFormat:     config.GetString("output"),
        OutputFile:       config.GetString("output-file"),
        OutputFileFormat: config.GetString("output-file-format"),
    }
    // If file format not specified, use primary output format
    if settings.OutputFileFormat == "" {
        settings.OutputFileFormat = settings.OutputFormat
    }
    return settings
}
```

## Advanced Features

### 4. Placeholder System

For enhanced file output, the system SHALL support placeholder replacement in file paths:

```go
func placeholderParser(value string, context interface{}) string {
    value = strings.ReplaceAll(value, "$TIMESTAMP",
        time.Now().Format("2006-01-02T15-04-05"))
    value = strings.ReplaceAll(value, "$AWS_REGION", getAWSRegion())
    value = strings.ReplaceAll(value, "$AWS_ACCOUNTID", getAWSAccountID())
    // Add more placeholders as needed
    return value
}
```

### 5. Command-Specific File Flags

Individual commands MAY implement their own specialized file flags with enhanced features:

```go
// Example: Report command with placeholder support
cmd.Flags().StringVar(&f.Outputfile, "file", "",
    "Optional file to save the output to. Supports placeholders: $TIMESTAMP, $AWS_REGION, $AWS_ACCOUNTID)
```

### 6. File Validation System

The system SHALL implement comprehensive file validation:

#### File Existence Validation
```go
type FileExistsRule struct {
    FieldName string
    Required  bool
    GetValue  func(flags.FlagValidator) string
}

func (f *FileExistsRule) Validate(ctx context.Context, flags flags.FlagValidator, vCtx *flags.ValidationContext) error {
    value := f.GetValue(flags)
    if value == "" {
        if f.Required {
            return fmt.Errorf("file path for '%s' is required", f.FieldName)
        }
        return nil
    }
    if _, err := os.Stat(value); os.IsNotExist(err) {
        return fmt.Errorf("file '%s' specified for '%s' does not exist", value, f.FieldName)
    }
    return nil
}
```

#### File Extension Validation
```go
type FileExtensionRule struct {
    FieldName         string
    AllowedExtensions []string
    GetValue          func(flags.FlagValidator) string
}

func (f *FileExtensionRule) Validate(ctx context.Context, flags flags.FlagValidator, vCtx *flags.ValidationContext) error {
    value := f.GetValue(flags)
    if value == "" {
        return nil
    }
    ext := strings.ToLower(filepath.Ext(value))
    for _, allowedExt := range f.AllowedExtensions {
        if ext == allowedExt {
            return nil
        }
    }
    return fmt.Errorf("file '%s' has invalid extension '%s', allowed: %v",
        value, ext, f.AllowedExtensions)
}
```

## Output Format Support

### 7. Multi-Format Output System

The system SHALL support multiple output formats:

- **table**: Human-readable tabular format (default)
- **csv**: Comma-separated values
- **json**: JavaScript Object Notation
- **markdown**: Markdown format with enhanced features
- **html**: HTML format
- **dot**: DOT graph format for visualization tools

### 8. Format-Specific Enhancements

#### Markdown/HTML Enhanced Output
For certain formats, the system SHALL provide enhanced output:

```go
// Example: Add Mermaid diagrams for timeline visualization
func addMermaidDiagram(output *format.OutputArray, events []Event) {
    if output.Settings.OutputFileFormat == "markdown" ||
       output.Settings.OutputFileFormat == "html" {
        mermaidContent := generateMermaidTimeline(events)
        output.AddContent(mermaidContent)
    }
}
```

#### Format-Specific Separators
```go
func (config *Config) GetSeparator() string {
    switch config.GetString("output") {
    case "table":
        return "\r\n"
    case "dot":
        return ","
    default:
        return ", "
    }
}
```

## Implementation Pattern

### 9. Standard Output Generation Pattern

Commands SHALL follow this pattern for dual output:

```go
func generateOutput(data interface{}, settings *OutputSettings) {
    // Create output structure
    output := format.OutputArray{
        Keys:     []string{"Column1", "Column2", "Column3"},
        Settings: settings,
    }

    // Configure output
    output.Settings.Title = "My Command Output"
    output.Settings.SortKey = "Column1"

    // Populate data
    for _, item := range data {
        content := map[string]interface{}{
            "Column1": item.Field1,
            "Column2": item.Field2,
            "Column3": item.Field3,
        }
        holder := format.OutputHolder{Contents: content}
        output.AddHolder(holder)
    }

    // Write to both stdout and file (if specified)
    output.Write()
}
```

## Error Handling

### 10. Structured Error Handling

The system SHALL implement structured error handling for file operations:

```go
type ValidationError struct {
    Err      error
    Warnings []string
    Infos    []string
}

// File system errors should be caught during validation
// before command execution
func validateFileOutput(settings *OutputSettings) error {
    if settings.OutputFile != "" {
        // Check if directory exists
        dir := filepath.Dir(settings.OutputFile)
        if _, err := os.Stat(dir); os.IsNotExist(err) {
            return fmt.Errorf("output directory '%s' does not exist", dir)
        }

        // Check write permissions
        if !isWritable(dir) {
            return fmt.Errorf("no write permission for directory '%s'", dir)
        }
    }
    return nil
}
```

## Configuration Defaults

### 11. Default Configuration

The system SHALL provide sensible defaults:

```go
// Default file extensions for different file types
viper.SetDefault("templates.extensions",
    []string{"", ".yaml", ".yml", ".templ", ".tmpl", ".template", ".json"})
viper.SetDefault("tags.extensions", []string{"", ".json"})
viper.SetDefault("parameters.extensions", []string{"", ".json"})

// Output defaults
viper.SetDefault("output", "table")
viper.SetDefault("table.style", "Default")
viper.SetDefault("table.max-column-width", 50)
```

## Usage Examples

### 12. Basic Usage

```bash
# Save table output to file
command --file output.txt

# Save as JSON while displaying table on stdout
command --file output.json --file-format json

# Use placeholders in filename
command --file "report-$TIMESTAMP-$AWS_REGION.md" --file-format markdown
```

### 13. Advanced Usage

```bash
# Report with timeline visualization
report --file "report-$TIMESTAMP.md" --file-format markdown

# Export data in multiple formats
exports --file exports.csv --file-format csv
resources --output json --file resources.json
```

## Integration Requirements

### 14. Output Library Integration

The system SHALL integrate with an output formatting library that provides:

- Consistent formatting across different output types
- Table styling options
- Color and emoji support for terminal output
- File writing capabilities with format conversion

This should be done using the go-output library.

### 15. Testing Requirements

The implementation SHALL include:

- Unit tests for file validation rules
- Integration tests for file output functionality
- Tests for placeholder replacement
- Error handling tests for file system operations
- Format conversion tests

## Security Considerations

### 16. Security Requirements

The system SHALL implement security measures:

- Path traversal protection for file paths
- Permission checks before file operations
- Sanitization of user-provided file paths
- Prevention of overwriting sensitive system files

```go
func sanitizeFilePath(path string) (string, error) {
    // Clean path and prevent traversal
    clean := filepath.Clean(path)
    if strings.Contains(clean, "..") {
        return "", fmt.Errorf("path traversal not allowed: %s", path)
    }

    // Convert to absolute path
    abs, err := filepath.Abs(clean)
    if err != nil {
        return "", fmt.Errorf("invalid file path: %s", path)
    }

    return abs, nil
}
```

This requirements document provides a comprehensive specification for implementing a robust file output system that can be adapted for various CLI tools requiring dual stdout/file output capabilities with format flexibility and advanced features.