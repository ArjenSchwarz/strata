# Color and Formatting

The go-output library provides rich text formatting capabilities including colors, emoji, and styling for enhanced user experience in terminal applications.

## Overview

The color and formatting system is built into the `OutputSettings` struct and provides methods for consistent text styling across your application. It uses the `github.com/fatih/color` package for terminal color support.

## Configuration

### Enabling Colors and Emoji

```go
settings := format.NewOutputSettings()

// Enable terminal colors
settings.UseColors = true

// Enable emoji in messages
settings.UseEmoji = true
```

### Color Support Detection

The library automatically detects terminal color support through the `fatih/color` package. Colors will be disabled automatically in environments that don't support them.

## Formatting Methods

The `OutputSettings` struct provides several methods for text formatting:

### Success Messages

Display positive feedback and successful operations.

```go
// Success with emoji/icon
message := settings.StringSuccess("Operation completed successfully")
fmt.Print(message)
```

**Output with colors and emoji:**
```
✅ Operation completed successfully
```

**Output without emoji:**
```
OK Operation completed successfully
```

### Warning Messages

Display warnings and alerts.

```go
// Single line warning
warning := settings.StringWarning("Configuration file not found")
fmt.Print(warning)

// Inline warning (no newline)
inlineWarning := settings.StringWarningInline("Invalid value")
fmt.Print(inlineWarning)
```

**Output (red color when colors enabled):**
```
Configuration file not found
```

### Failure Messages

Display errors and critical issues.

```go
// Failure with emphasis
failure := settings.StringFailure("Database connection failed")
fmt.Print(failure)
```

**Output with emoji:**
```
🚨 Database connection failed 🚨
```

**Output without emoji:**
```
!! Database connection failed !!
```

### Informational Messages

Display neutral information.

```go
// Information message
info := settings.StringInfo("Processing 150 records...")
fmt.Print(info)
```

**Output with emoji:**
```
ℹ️  Processing 150 records...
```

**Output without emoji:**
```
  Processing 150 records...
```

### Bold Text

Emphasize important text.

```go
// Bold text with newline
bold := settings.StringBold("Important Notice")
fmt.Print(bold)

// Bold text inline (no newline)
boldInline := settings.StringBoldInline("Required:")
fmt.Print(boldInline)
```

### Positive Messages

Display positive feedback (similar to success but more subtle).

```go
// Positive message with newline
positive := settings.StringPositive("All checks passed")
fmt.Print(positive)

// Positive message inline
positiveInline := settings.StringPositiveInline("Valid")
fmt.Print(positiveInline)
```

## Color Behavior

### With Colors Enabled (`UseColors = true`)

- **Success/Positive**: Green text with bold formatting
- **Warning/Failure**: Red text with bold formatting
- **Bold**: Bold formatting (no color change)
- **Info**: Regular text (no color change)

### With Colors Disabled (`UseColors = false`)

- All methods fall back to bold formatting only
- No color is applied
- Consistent appearance across different terminals

## Emoji Behavior

### With Emoji Enabled (`UseEmoji = true`)

- **Success**: ✅ checkmark
- **Failure**: 🚨 warning alarm
- **Info**: ℹ️ information symbol
- **Warning**: Uses !! without specific emoji

### With Emoji Disabled (`UseEmoji = false`)

- **Success**: "OK" text
- **Failure**: "!!" text symbols
- **Info**: No prefix
- **Warning**: No special prefix

## Usage Examples

### CLI Application Feedback

```go
func processData(settings *format.OutputSettings) {
    fmt.Print(settings.StringInfo("Starting data processing..."))

    // Simulate processing
    for i := 0; i < 100; i++ {
        if i%20 == 0 {
            progress := fmt.Sprintf("Processed %d/100 records", i)
            fmt.Print(settings.StringPositive(progress))
        }
    }

    // Check for issues
    if hasWarnings {
        fmt.Print(settings.StringWarning("Some records had validation warnings"))
    }

    // Final result
    if success {
        fmt.Print(settings.StringSuccess("Data processing completed"))
    } else {
        fmt.Print(settings.StringFailure("Data processing failed"))
    }
}
```

### Configuration Validation

```go
func validateConfig(config Config, settings *format.OutputSettings) {
    fmt.Print(settings.StringInfo("Validating configuration..."))

    if config.DatabaseURL == "" {
        fmt.Print(settings.StringFailure("Database URL is required"))
        return
    }

    if config.APIKey == "" {
        fmt.Print(settings.StringWarning("API key not provided, using default"))
    }

    if config.Debug {
        fmt.Print(settings.StringPositive("Debug mode enabled"))
    }

    fmt.Print(settings.StringSuccess("Configuration is valid"))
}
```

### Status Reporting

```go
func reportStatus(services []Service, settings *format.OutputSettings) {
    fmt.Print(settings.StringBold("Service Status Report"))

    for _, service := range services {
        status := fmt.Sprintf("%-20s", service.Name)

        switch service.Status {
        case "running":
            fmt.Print(status + settings.StringPositiveInline("RUNNING"))
        case "stopped":
            fmt.Print(status + settings.StringFailure("STOPPED"))
        case "warning":
            fmt.Print(status + settings.StringWarningInline("WARNING"))
        default:
            fmt.Print(status + settings.StringInfo("UNKNOWN"))
        }
        fmt.Println()
    }
}
```

## Integration with Output Formats

### Table Output

Colors and formatting are automatically applied to table output when `UseColors = true`:

```go
settings := format.NewOutputSettings()
settings.SetOutputFormat("table")
settings.UseColors = true
settings.TableStyle = format.TableStyles["ColoredBright"]
```

### HTML Output

Color settings don't directly affect HTML output since HTML has its own styling, but emoji settings can influence content:

```go
settings := format.NewOutputSettings()
settings.SetOutputFormat("html")
settings.UseEmoji = true  // Emoji will appear in HTML content
```

### Terminal vs File Output

When outputting to files, colors are automatically disabled by the underlying color library, but emoji settings are still respected:

```go
settings := format.NewOutputSettings()
settings.OutputFile = "report.txt"
settings.UseColors = true  // Will be ignored for file output
settings.UseEmoji = true   // Will be included in file content
```

## Best Practices

### 1. Consistent Settings

Use the same settings object throughout your application for consistent styling:

```go
// Create once, use everywhere
globalSettings := format.NewOutputSettings()
globalSettings.UseColors = true
globalSettings.UseEmoji = true
```

### 2. Environment Detection

Automatically configure based on environment:

```go
settings := format.NewOutputSettings()

// Disable colors in CI/CD environments
if os.Getenv("CI") != "" {
    settings.UseColors = false
    settings.UseEmoji = false
}

// Enable colors for interactive terminals
if isatty.IsTerminal(os.Stdout.Fd()) {
    settings.UseColors = true
}
```

### 3. User Configuration

Allow users to control formatting:

```go
func setupSettings(colorFlag, emojiFlag bool) *format.OutputSettings {
    settings := format.NewOutputSettings()
    settings.UseColors = colorFlag
    settings.UseEmoji = emojiFlag
    return settings
}

// CLI flags
var (
    colors = flag.Bool("colors", true, "Enable colored output")
    emoji  = flag.Bool("emoji", false, "Enable emoji in output")
)

settings := setupSettings(*colors, *emoji)
```

### 4. Graceful Degradation

Always provide meaningful output even when formatting is disabled:

```go
// Good: meaningful without formatting
fmt.Print(settings.StringSuccess("Backup completed"))

// Avoid: relies on formatting for meaning
fmt.Print(settings.StringPositiveInline("✓"))  // Just a checkmark
```

### 5. Testing

Test with different formatting configurations:

```go
func TestWithDifferentFormats(t *testing.T) {
    testCases := []struct {
        name      string
        useColors bool
        useEmoji  bool
    }{
        {"colors_and_emoji", true, true},
        {"colors_only", true, false},
        {"emoji_only", false, true},
        {"plain", false, false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            settings := format.NewOutputSettings()
            settings.UseColors = tc.useColors
            settings.UseEmoji = tc.useEmoji

            // Test your functionality
            result := settings.StringSuccess("test")
            // Assert expected behavior
        })
    }
}
```

## Terminal Compatibility

The formatting system is designed to work across different terminals and operating systems:

- **Windows**: Supports color through Windows Console API
- **macOS/Linux**: Full color support in most terminals
- **CI/CD**: Automatically degrades gracefully
- **File Output**: Colors are automatically stripped, emoji preserved

## Performance Considerations

- Color formatting has minimal performance impact
- String concatenation is optimized for common use cases
- Formatting methods are thread-safe
- No significant memory overhead for color support
