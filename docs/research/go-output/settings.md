# Settings and Configuration

The `OutputSettings` struct provides comprehensive configuration options for controlling output generation. This guide covers all available settings and their usage.

## Creating Settings

### Default Settings

```go
settings := format.NewOutputSettings()
// Creates settings with sensible defaults:
// - TableStyle: table.StyleDefault
// - TableMaxColumnWidth: 50
// - MermaidSettings: empty mermaid.Settings{}
```

### Custom Settings

```go
settings := &format.OutputSettings{
    OutputFormat:        "json",
    Title:              "My Report",
    UseColors:          true,
    UseEmoji:           true,
    TableStyle:         format.TableStyles["ColoredBright"],
    TableMaxColumnWidth: 30,
    // ... other settings
}
```

## Output Control Settings

### Format Selection

```go
// Set output format (converts to lowercase automatically)
settings.SetOutputFormat("JSON")  // becomes "json"

// Supported formats:
settings.SetOutputFormat("json")     // JSON format
settings.SetOutputFormat("yaml")     // YAML format
settings.SetOutputFormat("csv")      // CSV format
settings.SetOutputFormat("table")    // Terminal table
settings.SetOutputFormat("html")     // HTML table
settings.SetOutputFormat("markdown") // Markdown table
settings.SetOutputFormat("mermaid")  // Mermaid diagrams
settings.SetOutputFormat("dot")      // DOT graphs
settings.SetOutputFormat("drawio")   // Draw.io CSV
```

### File Output

```go
// Output to file
settings.OutputFile = "/path/to/output.json"

// Different format for file output
settings.OutputFile = "report.html"
settings.OutputFileFormat = "html"  // File format different from console format

// Append to existing file (HTML only)
settings.ShouldAppend = true
```

### Default File Extensions

The library provides automatic file extension detection:

```go
ext := settings.GetDefaultExtension()
// Returns:
// ".json" for json
// ".yaml" for yaml
// ".csv" for csv
// ".html" for html
// ".md" for markdown
// ".txt" for table
// ".dot" for dot
// ".csv" for drawio
// ".mmd" for mermaid
```

## Display and Styling Settings

### Visual Enhancements

```go
// Enable terminal colors
settings.UseColors = true

// Enable emoji in output
settings.UseEmoji = true

// Add title to output
settings.Title = "Data Report"

// Add table of contents (HTML/Markdown)
settings.HasTOC = true
```

### Table Styling

```go
// Set table style from predefined styles
settings.TableStyle = format.TableStyles["ColoredBright"]

// Available table styles:
var availableStyles = []string{
    "Default",
    "Bold",
    "ColoredBright",
    "ColoredDark",
    "ColoredBlackOnBlueWhite",
    "ColoredBlackOnCyanWhite",
    "ColoredBlackOnGreenWhite",
    "ColoredBlackOnMagentaWhite",
    "ColoredBlackOnYellowWhite",
    "ColoredBlackOnRedWhite",
    "ColoredBlueWhiteOnBlack",
    "ColoredCyanWhiteOnBlack",
    "ColoredGreenWhiteOnBlack",
    "ColoredMagentaWhiteOnBlack",
    "ColoredRedWhiteOnBlack",
    "ColoredYellowWhiteOnBlack",
}

// Set maximum column width for tables
settings.TableMaxColumnWidth = 30  // Default: 50

// Add spacing between tables
settings.SeparateTables = true
```

## Data Organization Settings

### Sorting

```go
// Sort output by specific column
settings.SortKey = "Name"  // Must match a key in your Keys slice
```

### Line Splitting

```go
// Split array values into separate rows
settings.SplitLines = true

// Example:
// Without SplitLines: Tags -> "admin, user"
// With SplitLines:    Two rows, one for each tag
```

## Format-Specific Settings

### Markdown Front Matter

```go
// Add YAML front matter to Markdown output
settings.FrontMatter = map[string]string{
    "title":  "User Report",
    "author": "System Generator",
    "date":   "2024-01-15",
    "tags":   "users, report, data",
}
```

Output:
```markdown
---
title: User Report
author: System Generator
date: 2024-01-15
tags: users, report, data
---

| Name | Age |
|------|-----|
| ...  | ... |
```

### Graphical Output Settings

For Mermaid, DOT, and Draw.io formats, you need to define relationships:

```go
// Basic relationship columns
settings.AddFromToColumns("Source", "Target")

// Relationship with labels
settings.AddFromToColumnsWithLabel("Parent", "Child", "RelationType")

// Direct assignment
settings.FromToColumns = &format.FromToColumns{
    From:  "SourceNode",
    To:    "TargetNode",
    Label: "ConnectionType",
}
```

### Mermaid Settings

```go
// Create mermaid settings
mermaidSettings := &mermaid.Settings{
    AddMarkdown: true,      // Wrap in markdown code blocks
    AddHTML:     false,     // Add HTML script tags
    ChartType:   "flowchart", // "flowchart", "piechart", "ganttchart"
}

// For Gantt charts
mermaidSettings.GanttSettings = &mermaid.GanttSettings{
    LabelColumn:     "TaskName",
    StartDateColumn: "StartTime",
    DurationColumn:  "Duration",
    StatusColumn:    "Status",
}

settings.MermaidSettings = mermaidSettings
```

### Draw.io Settings

```go
// Create Draw.io header
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Set layout
header.SetLayout(drawio.LayoutVerticalFlow)
// Available layouts:
// - LayoutAuto
// - LayoutNone
// - LayoutHorizontalFlow
// - LayoutVerticalFlow
// - LayoutHorizontalTree
// - LayoutVerticalTree
// - LayoutOrganic
// - LayoutCircle

// Set spacing
header.SetSpacing(40, 100, 40) // node, level, edge spacing

// Set dimensions
header.SetHeightAndWidth("auto", "auto")
header.SetPadding(10)

// Set identity column for updating existing diagrams
header.SetIdentity("ID")

// Add connections
connection := drawio.NewConnection()
connection.From = "Parent"
connection.To = "Name"
connection.Style = drawio.DefaultConnectionStyle
header.AddConnection(connection)

settings.DrawIOHeader = header
```

## Cloud Integration Settings

### Amazon S3 Output

```go
// Set up S3 output
s3Client := // ... create S3 client
settings.SetS3Bucket(s3Client, "my-bucket", "reports/data.json")

// The S3Output structure:
type S3Output struct {
    S3Client *s3.Client
    Bucket   string
    Path     string
}
```

## Validation and Helper Methods

### Format Requirements

```go
// Check if format needs From/To columns
needsColumns := settings.NeedsFromToColumns()
// Returns true for: "dot", "mermaid"
// Returns false for: "json", "yaml", "csv", "table", "html", "markdown"
```

### Separators

```go
// Get appropriate separator for format
separator := settings.GetSeparator()
// Returns:
// "\n" for table, markdown, csv
// "," for dot
// ", " for others (default)
```

## Complete Configuration Example

```go
func createFullSettings() *format.OutputSettings {
    settings := format.NewOutputSettings()

    // Basic output control
    settings.SetOutputFormat("html")
    settings.OutputFile = "report.html"
    settings.Title = "Comprehensive Data Report"

    // Visual styling
    settings.UseColors = true
    settings.UseEmoji = true
    settings.HasTOC = true
    settings.SeparateTables = true

    // Table configuration
    settings.TableStyle = format.TableStyles["ColoredBright"]
    settings.TableMaxColumnWidth = 40

    // Data organization
    settings.SortKey = "Name"
    settings.SplitLines = false

    // Markdown front matter
    settings.FrontMatter = map[string]string{
        "title":       "Data Report",
        "description": "Automated data export",
        "version":     "1.0",
    }

    // S3 integration (if needed)
    // settings.SetS3Bucket(s3Client, "reports-bucket", "monthly/report.html")

    return settings
}
```

## Settings Validation

The library performs automatic validation:

- Missing Draw.io header causes `log.Fatal` for "drawio" format
- Missing FromToColumns causes `log.Fatal` for "dot" and "mermaid" formats
- Missing Mermaid settings causes `log.Fatal` for "mermaid" format
- Invalid S3 settings cause errors during output

## Best Practices

1. **Use Default Constructor**: Start with `NewOutputSettings()` for sensible defaults
2. **Validate Requirements**: Check format requirements before setting format
3. **Consistent Naming**: Use consistent key names that work well as headers
4. **File Extensions**: Let the library determine extensions or use `GetDefaultExtension()`
5. **Error Handling**: Be prepared for `log.Fatal` calls on configuration errors
6. **Performance**: Don't recreate settings objects unnecessarily
7. **Testing**: Test with different formats to ensure data compatibility
