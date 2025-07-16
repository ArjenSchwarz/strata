# File Output System Design Document

## Overview

This design document outlines the implementation of a dual-output system for the Strata CLI tool that enables simultaneous output to stdout and file destinations with format conversion capabilities. The system extends the existing go-output integration to provide flexible file output with placeholder support, validation, and enhanced formatting options.

### Design Goals

- **Seamless Integration**: Extend existing output system without breaking current functionality
- **Format Flexibility**: Support different formats for stdout and file output
- **Placeholder Support**: Enable dynamic file naming with contextual placeholders
- **Validation**: Comprehensive file path and format validation
- **Security**: Prevent path traversal and unauthorized file access
- **Extensibility**: Support future output formats and features

## Architecture

### High-Level Architecture

The file output system follows a layered architecture that integrates with the existing Strata codebase:

```
┌─────────────────────────────────────────────────────────────┐
│                    Command Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │   Root Flags    │  │ Command Flags   │                  │
│  │ --file          │  │ --file (local)  │                  │
│  │ --file-format   │  │                 │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                Configuration Layer                          │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │            Enhanced OutputSettings                      │ │
│  │  • OutputFormat (stdout)                               │ │
│  │  • OutputFile (file path)                              │ │
│  │  • OutputFileFormat (file format override)             │ │
│  │  • Placeholder resolution                              │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                 Validation Layer                            │
│  ┌─────────────────┐  ┌─────────────────┐                  │
│  │  File Validator │  │ Format Validator│                  │
│  │  • Path safety  │  │ • Format support│                  │
│  │  • Permissions  │  │ • Extension map │                  │
│  │  • Existence    │  │                 │                  │
│  └─────────────────┘  └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Output Layer                              │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              go-output Integration                      │ │
│  │  • Dual output writing                                 │ │
│  │  • Format-specific enhancements                        │ │
│  │  • Error handling and recovery                         │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Integration Points

The system integrates with existing Strata components:

1. **Root Command**: Adds persistent flags for global file output
2. **Config System**: Extends `Config` and `OutputSettings` structures
3. **go-output Library**: Leverages existing formatting capabilities
4. **Command Pattern**: Maintains existing command structure and flow

## Components and Interfaces

### 1. Flag Management Component

**Location**: `cmd/root.go` (persistent flags) and individual commands (local flags)

```go
// Persistent flags added to root command
rootCmd.PersistentFlags().String("file", "", "Optional file to save the output to, in addition to stdout")
rootCmd.PersistentFlags().String("file-format", "", "Optional format for the file, defaults to the same as output")

// Viper binding for configuration integration
viper.BindPFlag("output-file", rootCmd.PersistentFlags().Lookup("file"))
viper.BindPFlag("output-file-format", rootCmd.PersistentFlags().Lookup("file-format"))
```

**Design Rationale**: 
- Persistent flags ensure availability across all subcommands
- Viper integration maintains consistency with existing configuration system
- Optional nature preserves backward compatibility

### 2. Enhanced Configuration Component

**Location**: `config/config.go`

```go
// Enhanced OutputSettings creation
func (config *Config) NewOutputSettings() *format.OutputSettings {
    settings := format.NewOutputSettings()
    // ... existing configuration ...
    
    // New file output configuration
    settings.OutputFile = config.GetString("output-file")
    settings.OutputFileFormat = config.GetString("output-file-format")
    
    // Apply placeholder resolution
    if settings.OutputFile != "" {
        settings.OutputFile = config.resolvePlaceholders(settings.OutputFile)
    }
    
    // Default file format to stdout format if not specified
    if settings.OutputFileFormat == "" {
        settings.OutputFileFormat = settings.OutputFormat
    }
    
    return settings
}
```

**Design Rationale**:
- Extends existing `NewOutputSettings()` method to maintain consistency
- Placeholder resolution happens at configuration time for early validation
- Format defaulting logic centralizes decision-making

### 3. Placeholder Resolution Component

**Location**: `config/config.go` (new methods)

```go
func (config *Config) resolvePlaceholders(value string) string {
    replacements := map[string]string{
        "$TIMESTAMP":     time.Now().Format("2006-01-02T15-04-05"),
        "$AWS_REGION":    config.getAWSRegion(),
        "$AWS_ACCOUNTID": config.getAWSAccountID(),
    }
    
    result := value
    for placeholder, replacement := range replacements {
        result = strings.ReplaceAll(result, placeholder, replacement)
    }
    
    return result
}
```

**Design Rationale**:
- Map-based approach allows easy extension of placeholders
- Timestamp format avoids filesystem-incompatible characters
- Context-aware placeholders (region, account) enhance file organization

### 4. Validation Component

**Location**: `config/validation.go` (new file)

```go
type FileValidator struct {
    config *Config
}

func (fv *FileValidator) ValidateFileOutput(settings *format.OutputSettings) error {
    if settings.OutputFile == "" {
        return nil // No file output, nothing to validate
    }
    
    // Validate file path safety
    if err := fv.validatePathSafety(settings.OutputFile); err != nil {
        return fmt.Errorf("file path validation failed: %w", err)
    }
    
    // Validate directory permissions
    if err := fv.validateDirectoryPermissions(settings.OutputFile); err != nil {
        return fmt.Errorf("directory permission validation failed: %w", err)
    }
    
    // Validate format support
    if err := fv.validateFormatSupport(settings.OutputFileFormat); err != nil {
        return fmt.Errorf("format validation failed: %w", err)
    }
    
    return nil
}
```

**Design Rationale**:
- Separate validation component maintains single responsibility
- Early validation prevents runtime errors
- Comprehensive checks ensure security and functionality

### 5. Enhanced Output Integration

**Location**: Integration with go-output library through existing patterns

The system leverages the existing go-output library's capabilities while extending them for dual output:

```go
// Enhanced output generation pattern
func generateOutput(data interface{}, settings *format.OutputSettings) error {
    output := format.OutputArray{
        Keys:     []string{"Column1", "Column2", "Column3"},
        Settings: settings,
    }
    
    // Configure output
    output.Settings.Title = "Command Output"
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
    
    // Add format-specific enhancements
    if err := addFormatEnhancements(&output); err != nil {
        return fmt.Errorf("failed to add format enhancements: %w", err)
    }
    
    // Write to both stdout and file
    return output.Write()
}
```

**Design Rationale**:
- Maintains existing output generation patterns
- go-output library handles dual output internally
- Format enhancements are applied before writing

## Data Models

### 1. Enhanced OutputSettings Structure

The existing `format.OutputSettings` from go-output library will be extended to include:

```go
// These fields will be added to the existing OutputSettings
type OutputSettings struct {
    // ... existing fields ...
    
    // File output configuration
    OutputFile       string // File path for saving output
    OutputFileFormat string // Format override for file output
    
    // Placeholder context (for resolution)
    PlaceholderContext map[string]string
}
```

### 2. Validation Result Structure

```go
type ValidationResult struct {
    Valid    bool
    Errors   []error
    Warnings []string
    Infos    []string
}

func (vr *ValidationResult) HasErrors() bool {
    return len(vr.Errors) > 0
}

func (vr *ValidationResult) AddError(err error) {
    vr.Errors = append(vr.Errors, err)
    vr.Valid = false
}
```

### 3. Placeholder Context Structure

```go
type PlaceholderContext struct {
    Timestamp string
    Region    string
    AccountID string
    StackName string
    // Additional context fields as needed
}

func (pc *PlaceholderContext) ToMap() map[string]string {
    return map[string]string{
        "$TIMESTAMP": pc.Timestamp,
        "$REGION":    pc.Region,
        "$ACCOUNTID": pc.AccountID,
        "$STACKNAME": pc.StackName,
    }
}
```

## Error Handling

### 1. Validation Errors

The system implements structured error handling for different validation scenarios:

```go
type FileOutputError struct {
    Type    string // "validation", "permission", "format", "write"
    Path    string
    Format  string
    Message string
    Cause   error
}

func (e *FileOutputError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s error for %s: %s (caused by: %v)", 
            e.Type, e.Path, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s error for %s: %s", e.Type, e.Path, e.Message)
}
```

### 2. Error Recovery Strategy

The system implements graceful degradation:

1. **File Write Failure**: Continue with stdout output, log warning
2. **Format Conversion Failure**: Fall back to stdout format for file
3. **Placeholder Resolution Failure**: Use original path, log warning
4. **Permission Errors**: Fail fast with clear error message

```go
func handleFileOutputError(err error, settings *format.OutputSettings) error {
    switch e := err.(type) {
    case *FileOutputError:
        switch e.Type {
        case "write":
            // Log warning and continue with stdout only
            log.Warnf("Failed to write to file %s: %v", e.Path, e.Cause)
            settings.OutputFile = "" // Disable file output
            return nil
        case "validation", "permission":
            // These are fatal errors
            return err
        case "format":
            // Fall back to stdout format
            log.Warnf("Unsupported file format %s, using %s instead", 
                settings.OutputFileFormat, settings.OutputFormat)
            settings.OutputFileFormat = settings.OutputFormat
            return nil
        }
    }
    return err
}
```

### 3. Validation Error Aggregation

Multiple validation errors are collected and presented together:

```go
func (fv *FileValidator) ValidateAll(settings *format.OutputSettings) *ValidationResult {
    result := &ValidationResult{Valid: true}
    
    validators := []func(*format.OutputSettings) error{
        fv.validatePathSafety,
        fv.validateDirectoryPermissions,
        fv.validateFormatSupport,
    }
    
    for _, validator := range validators {
        if err := validator(settings); err != nil {
            result.AddError(err)
        }
    }
    
    return result
}
```

## Testing Strategy

### 1. Unit Testing Approach

**File Validation Tests**:
```go
func TestFileValidator_ValidatePathSafety(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        wantErr  bool
        errType  string
    }{
        {
            name:    "valid relative path",
            path:    "output/report.json",
            wantErr: false,
        },
        {
            name:    "path traversal attempt",
            path:    "../../../etc/passwd",
            wantErr: true,
            errType: "validation",
        },
        {
            name:    "absolute path",
            path:    "/tmp/output.json",
            wantErr: false,
        },
    }
    
    validator := &FileValidator{}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.validatePathSafety(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("validatePathSafety() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Placeholder Resolution Tests**:
```go
func TestConfig_ResolvePlaceholders(t *testing.T) {
    config := &Config{}
    
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "timestamp placeholder",
            input:    "report-$TIMESTAMP.json",
            expected: "report-2025-01-15T10-30-45.json", // Mock timestamp
        },
        {
            name:     "multiple placeholders",
            input:    "$REGION-$ACCOUNTID-$TIMESTAMP.md",
            expected: "us-east-1-123456789-2025-01-15T10-30-45.md",
        },
        {
            name:     "no placeholders",
            input:    "simple-report.json",
            expected: "simple-report.json",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := config.resolvePlaceholders(tt.input)
            if result != tt.expected {
                t.Errorf("resolvePlaceholders() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### 2. Integration Testing

**End-to-End File Output Tests**:
```go
func TestFileOutput_Integration(t *testing.T) {
    tempDir := t.TempDir()
    
    tests := []struct {
        name           string
        outputFormat   string
        fileFormat     string
        fileName       string
        expectStdout   bool
        expectFile     bool
    }{
        {
            name:         "table to file as json",
            outputFormat: "table",
            fileFormat:   "json",
            fileName:     "output.json",
            expectStdout: true,
            expectFile:   true,
        },
        {
            name:         "json to markdown file",
            outputFormat: "json",
            fileFormat:   "markdown",
            fileName:     "output.md",
            expectStdout: true,
            expectFile:   true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test data and configuration
            // Execute command with file output
            // Verify stdout and file contents
        })
    }
}
```

### 3. Security Testing

**Path Traversal Prevention Tests**:
```go
func TestSecurity_PathTraversal(t *testing.T) {
    maliciousPaths := []string{
        "../../../etc/passwd",
        "..\\..\\..\\windows\\system32\\config\\sam",
        "/etc/shadow",
        "~/../../etc/passwd",
    }
    
    validator := &FileValidator{}
    for _, path := range maliciousPaths {
        t.Run(fmt.Sprintf("block_%s", path), func(t *testing.T) {
            err := validator.validatePathSafety(path)
            if err == nil {
                t.Errorf("Expected path traversal to be blocked for: %s", path)
            }
        })
    }
}
```

### 4. Performance Testing

**Large File Output Tests**:
```go
func BenchmarkFileOutput_LargeDataset(b *testing.B) {
    // Generate large dataset
    data := generateLargeTestData(10000) // 10k records
    
    settings := &format.OutputSettings{
        OutputFormat:     "json",
        OutputFile:       "benchmark_output.json",
        OutputFileFormat: "json",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        generateOutput(data, settings)
    }
}
```

## Implementation Phases

### Phase 1: Core Infrastructure
1. Add persistent flags to root command
2. Extend configuration system
3. Implement basic placeholder resolution
4. Create validation framework

### Phase 2: File Output Integration
1. Integrate with go-output library
2. Implement dual output writing
3. Add format-specific enhancements
4. Error handling and recovery

### Phase 3: Advanced Features
1. Enhanced placeholder system
2. Format-specific optimizations
3. Security hardening
4. Performance optimizations

### Phase 4: Testing and Documentation
1. Comprehensive test suite
2. Integration tests
3. Security testing
4. Documentation updates

## Security Considerations

### 1. Path Traversal Prevention

```go
func sanitizeFilePath(path string) (string, error) {
    // Clean path and resolve any relative components
    clean := filepath.Clean(path)
    
    // Check for path traversal attempts
    if strings.Contains(clean, "..") {
        return "", fmt.Errorf("path traversal not allowed: %s", path)
    }
    
    // Convert to absolute path for consistency
    abs, err := filepath.Abs(clean)
    if err != nil {
        return "", fmt.Errorf("invalid file path: %s", path)
    }
    
    return abs, nil
}
```

### 2. Permission Validation

```go
func validateWritePermissions(filePath string) error {
    dir := filepath.Dir(filePath)
    
    // Check if directory exists
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        return fmt.Errorf("directory does not exist: %s", dir)
    }
    
    // Test write permissions by creating a temporary file
    tempFile := filepath.Join(dir, ".strata_write_test")
    file, err := os.Create(tempFile)
    if err != nil {
        return fmt.Errorf("no write permission for directory: %s", dir)
    }
    file.Close()
    os.Remove(tempFile)
    
    return nil
}
```

### 3. File Overwrite Protection

```go
func checkFileOverwrite(filePath string) error {
    if _, err := os.Stat(filePath); err == nil {
        // File exists - could implement confirmation prompt
        // For now, we'll allow overwrite but log a warning
        log.Warnf("File %s will be overwritten", filePath)
    }
    return nil
}
```

This design provides a comprehensive, secure, and extensible file output system that integrates seamlessly with the existing Strata codebase while maintaining backward compatibility and following Go best practices.
##
 Format Support and Configuration

### Supported Output Formats

The system supports all formats specified in the requirements:

- **table**: Human-readable tabular format (default)
- **csv**: Comma-separated values for data processing
- **json**: JavaScript Object Notation for API integration
- **markdown**: Markdown format with enhanced features
- **html**: HTML format for web display
- **dot**: DOT graph format for visualization tools

### Format-Specific Enhancements

As specified in requirements, the system provides enhanced output for certain formats:

```go
// Add format-specific enhancements before writing
func addFormatEnhancements(output *format.OutputArray) error {
    if output.Settings.OutputFileFormat == "markdown" || 
       output.Settings.OutputFileFormat == "html" {
        // Add Mermaid diagrams for timeline visualization
        if timeline := generateTimelineData(output); timeline != "" {
            output.AddContent(timeline)
        }
        
        // Add enhanced formatting for better readability
        output.Settings.MarkdownEnhancements = true
    }
    return nil
}
```

### Format-Specific Separators

As required, the system supports different separators based on output format:

```go
func (config *Config) GetSeparator() string {
    switch config.GetString("output") {
    case "table":
        return "\r\n"
    case "dot":
        return ","
    case "csv":
        return ","
    default:
        return ", "
    }
}
```

### Configuration Defaults

The system provides sensible defaults as specified in requirements:

```go
func setConfigDefaults() {
    // Default file extensions for different file types
    viper.SetDefault("templates.extensions", 
        []string{"", ".yaml", ".yml", ".templ", ".tmpl", ".template", ".json"})
    viper.SetDefault("tags.extensions", []string{"", ".json"})
    viper.SetDefault("parameters.extensions", []string{"", ".json"})
    
    // Output defaults
    viper.SetDefault("output", "table")
    viper.SetDefault("table.style", "Default")
    viper.SetDefault("table.max-column-width", 50)
    
    // File output defaults
    viper.SetDefault("output-file", "")
    viper.SetDefault("output-file-format", "")
}
```

**Design Rationale**:
- Comprehensive format support addresses all requirements
- Enhanced features for markdown/HTML improve user experience
- Sensible defaults reduce configuration burden
- Format-specific separators ensure proper data formatting

## Enhanced Validation Details

### Detailed Validation Implementation

The validation component includes comprehensive checks as specified in requirements:

```go
// validatePathSafety implements path traversal protection
func (fv *FileValidator) validatePathSafety(filePath string) error {
    clean := filepath.Clean(filePath)
    if strings.Contains(clean, "..") {
        return fmt.Errorf("path traversal not allowed: %s", filePath)
    }
    return nil
}

// validateDirectoryPermissions checks write permissions and creates directories
func (fv *FileValidator) validateDirectoryPermissions(filePath string) error {
    dir := filepath.Dir(filePath)
    
    // Check if directory exists, create if it doesn't
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return fmt.Errorf("cannot create directory %s: %w", dir, err)
        }
    }
    
    // Test write permissions
    tempFile := filepath.Join(dir, ".strata_write_test")
    file, err := os.Create(tempFile)
    if err != nil {
        return fmt.Errorf("no write permission for directory: %s", dir)
    }
    file.Close()
    os.Remove(tempFile)
    
    return nil
}

// validateFormatSupport checks if the output format is supported
func (fv *FileValidator) validateFormatSupport(format string) error {
    supportedFormats := []string{"table", "json", "csv", "markdown", "html", "dot"}
    for _, supported := range supportedFormats {
        if format == supported {
            return nil
        }
    }
    return fmt.Errorf("unsupported output format: %s", format)
}
```

## Enhanced Placeholder Support

### Extended Placeholder Resolution

The placeholder system supports all placeholders mentioned in requirements:

```go
func (config *Config) resolvePlaceholders(value string) string {
    replacements := map[string]string{
        "$TIMESTAMP":     time.Now().Format("2006-01-02T15-04-05"),
        "$AWS_REGION":    config.getAWSRegion(),
        "$AWS_ACCOUNTID": config.getAWSAccountID(),
    }
    
    result := value
    for placeholder, replacement := range replacements {
        result = strings.ReplaceAll(result, placeholder, replacement)
    }
    
    return result
}

// Helper methods for context extraction
func (config *Config) getAWSRegion() string {
    // Extract region from AWS configuration or environment
    if region := config.GetString("aws.region"); region != "" {
        return region
    }
    return os.Getenv("AWS_REGION")
}

func (config *Config) getAWSAccountID() string {
    // Extract account ID from AWS configuration or environment
    if accountID := config.GetString("aws.account-id"); accountID != "" {
        return accountID
    }
    return os.Getenv("AWS_ACCOUNT_ID")
}
```

**Design Rationale**:
- Extended placeholder support addresses all requirements
- Environment variable fallbacks provide flexibility
- Context-aware extraction enhances file organization

This enhanced design now fully addresses all requirements specified in the requirements document, providing comprehensive coverage of global flag definition, configuration integration, placeholder support, validation, format support, error handling, and security considerations while maintaining seamless integration with the existing Strata codebase.