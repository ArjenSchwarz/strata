# Go-Output v2 Migration Guide

This guide provides comprehensive instructions for migrating from go-output v1 to v2. Version 2 is a complete redesign that eliminates global state, provides thread safety, and maintains exact key ordering while preserving all v1 functionality.

## Table of Contents

- [Overview](#overview)
- [Breaking Changes](#breaking-changes)
- [Automated Migration](#automated-migration)
- [Migration Patterns](#migration-patterns)
  - [Basic Output](#basic-output)
  - [Multiple Tables](#multiple-tables)
  - [Output Settings](#output-settings)
  - [Progress Indicators](#progress-indicators)
  - [File Output](#file-output)
  - [S3 Output](#s3-output)
  - [Chart and Diagram Output](#chart-and-diagram-output)
- [Feature-by-Feature Migration](#feature-by-feature-migration)
- [Common Issues](#common-issues)
- [Examples](#examples)

## Overview

Go-Output v2 introduces a clean, modern API while maintaining feature parity with v1. The main changes are:

- **No Global State**: All state is encapsulated in instances
- **Builder Pattern**: Fluent API for document construction
- **Functional Options**: Configuration through option functions
- **Key Order Preservation**: Exact user-specified key ordering is maintained
- **Thread Safety**: Safe for concurrent use

## Breaking Changes

### 1. Import Path
```go
// v1
import "github.com/ArjenSchwarz/go-output/format"

// v2
import "github.com/ArjenSchwarz/go-output/v2"
```

### 2. OutputArray Replaced with Builder Pattern
```go
// v1
output := &format.OutputArray{
    Keys: []string{"Name", "Age"},
}

// v2
doc := output.New().
    Table("", data, output.WithKeys("Name", "Age")).
    Build()
```

### 3. OutputSettings Replaced with Functional Options
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFormat = "table"
settings.UseEmoji = true

// v2
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithTransformer(&output.EmojiTransformer{}),
)
```

### 4. Write() Method Requires Context
```go
// v1
output.Write()

// v2
output.NewOutput().Render(ctx, doc)
```

### 5. Keys Field Replaced with Schema Options
```go
// v1
output.Keys = []string{"Name", "Age", "Status"}

// v2
output.WithKeys("Name", "Age", "Status")
// or
output.WithSchema(
    output.Field{Name: "Name"},
    output.Field{Name: "Age"},
    output.Field{Name: "Status"},
)
```

## Automated Migration

Use the included migration tool to automatically convert most v1 code:

```bash
# Install the migration tool
go install github.com/ArjenSchwarz/go-output/v2/migrate/cmd/migrate@latest

# Migrate a single file
migrate -file main.go

# Migrate an entire directory
migrate -source ./myproject

# Dry run to see changes without applying them
migrate -source ./myproject -dry-run

# Verbose mode for detailed information
migrate -source ./myproject -verbose
```

The migration tool handles approximately 80% of common v1 usage patterns. Manual adjustments may be needed for complex scenarios.

## Migration Patterns

### Basic Output

#### Simple Table Output
```go
// v1
output := &format.OutputArray{}
output.AddContents(map[string]interface{}{
    "Name": "Alice",
    "Age":  30,
})
output.Write()

// v2
ctx := context.Background()
doc := output.New().
    Table("", []map[string]interface{}{
        {"Name": "Alice", "Age": 30},
    }).
    Build()

output.NewOutput(output.WithFormat(output.Table)).Render(ctx, doc)
```

#### Table with Key Ordering
```go
// v1
output := &format.OutputArray{
    Keys: []string{"ID", "Name", "Status"},
}
output.AddContents(data)
output.Write()

// v2
ctx := context.Background()
doc := output.New().
    Table("", data, output.WithKeys("ID", "Name", "Status")).
    Build()

output.NewOutput(output.WithFormat(output.Table)).Render(ctx, doc)
```

### Multiple Tables

#### Multiple Tables with Different Keys
```go
// v1
output := &format.OutputArray{}

// First table
output.Keys = []string{"Name", "Email"}
output.AddContents(userData)
output.AddToBuffer()

// Second table
output.Keys = []string{"ID", "Status", "Time"}
output.AddContents(statusData)
output.AddToBuffer()

output.Write()

// v2
ctx := context.Background()
doc := output.New().
    Table("Users", userData, output.WithKeys("Name", "Email")).
    Table("Status", statusData, output.WithKeys("ID", "Status", "Time")).
    Build()

output.NewOutput(output.WithFormat(output.Table)).Render(ctx, doc)
```

#### Tables with Headers
```go
// v1
output := &format.OutputArray{}
output.AddHeader("User Report")
output.Keys = []string{"Name", "Role"}
output.AddContents(users)
output.Write()

// v2
ctx := context.Background()
doc := output.New().
    Header("User Report").
    Table("", users, output.WithKeys("Name", "Role")).
    Build()

output.NewOutput(output.WithFormat(output.Table)).Render(ctx, doc)
```

### Output Settings

#### Basic Settings Migration
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFormat = "json"
settings.OutputFile = "report.json"
settings.UseEmoji = true
settings.UseColors = true

output := &format.OutputArray{
    Settings: settings,
}

// v2
out := output.NewOutput(
    output.WithFormat(output.JSON),
    output.WithWriter(output.NewFileWriter(".", "report.json")),
    output.WithTransformer(&output.EmojiTransformer{}),
    output.WithTransformer(&output.ColorTransformer{}),
)
```

#### Table Styling
```go
// v1
settings := format.NewOutputSettings()
settings.TableStyle = "ColoredBright"

// v2
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithTableStyle("ColoredBright"),
)
```

#### Multiple Output Formats
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFormat = "table"
settings.OutputFile = "report.html"
settings.OutputFileFormat = "html"

// v2
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithFormat(output.HTML),
    output.WithWriter(&output.StdoutWriter{}),
    output.WithWriter(output.NewFileWriter(".", "report.html")),
)
```

### Progress Indicators

#### Basic Progress
```go
// v1
settings := format.NewOutputSettings()
p := format.NewProgress(settings)
p.SetTotal(100)
p.SetColor(format.ProgressColorGreen)

for i := 0; i < 100; i++ {
    p.Increment(1)
    p.SetStatus(fmt.Sprintf("Processing item %d", i))
}
p.Complete()

// v2
p := output.NewProgress(output.Table,
    output.WithProgressColor(output.ProgressColorGreen),
)
p.SetTotal(100)

for i := 0; i < 100; i++ {
    p.Increment(1)
    p.SetStatus(fmt.Sprintf("Processing item %d", i))
}
p.Complete()
```

#### Progress with Output
```go
// v1
settings := format.NewOutputSettings()
settings.SetOutputFormat("table")
settings.ProgressOptions = format.ProgressOptions{
    Color: format.ProgressColorBlue,
    Status: "Loading data",
}

// v2
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithProgress(output.NewProgress(output.Table,
        output.WithProgressColor(output.ProgressColorBlue),
        output.WithProgressStatus("Loading data"),
    )),
)
```

### File Output

#### Simple File Output
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFile = "report.csv"
settings.OutputFormat = "csv"

// v2
out := output.NewOutput(
    output.WithFormat(output.CSV),
    output.WithWriter(output.NewFileWriter(".", "report.csv")),
)
```

#### Multiple File Outputs
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFile = "report.json"
settings.OutputFileFormat = "json"
settings.OutputFormat = "table" // For stdout

// v2
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithFormat(output.JSON),
    output.WithWriter(&output.StdoutWriter{}),
    output.WithWriter(output.NewFileWriter(".", "report.json")),
)
```

### S3 Output

```go
// v1
settings := format.NewOutputSettings()
settings.OutputS3Bucket = "my-bucket"
settings.OutputS3Key = "reports/output.json"

// v2
s3Client := s3.NewFromConfig(cfg)
out := output.NewOutput(
    output.WithFormat(output.JSON),
    output.WithWriter(output.NewS3Writer(s3Client, "my-bucket", "reports/output.json")),
)
```

### Chart and Diagram Output

#### DOT Format (Graphviz)
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFormat = "dot"
settings.DotFromColumn = "source"
settings.DotToColumn = "target"

// v2
doc := output.New().
    Graph("Network", data, 
        output.WithFromTo("source", "target"),
    ).
    Build()

out := output.NewOutput(output.WithFormat(output.DOT))
```

#### Mermaid Charts
```go
// v1
settings := format.NewOutputSettings()
settings.OutputFormat = "mermaid"
settings.MermaidSettings = &mermaid.Settings{
    ChartType: "gantt",
}

// v2
doc := output.New().
    Chart("Project Timeline", ganttData,
        output.WithChartType("gantt"),
    ).
    Build()

out := output.NewOutput(output.WithFormat(output.Mermaid))
```

#### Draw.io Diagrams
```go
// v1
drawio.SetHeaderValues(drawio.Header{
    Label: "%Name%",
    Style: "%Image%",
    Layout: "horizontalflow",
})

// v2
doc := output.New().
    DrawIO("Architecture", data,
        output.WithDrawIOLayout("horizontalflow"),
        output.WithDrawIOLabel("%Name%"),
        output.WithDrawIOStyle("%Image%"),
    ).
    Build()

out := output.NewOutput(output.WithFormat(output.DrawIO))
```

## Feature-by-Feature Migration

### Sorting
```go
// v1
settings.SortKey = "Name"

// v2
output.WithTransformer(&output.SortTransformer{
    Key: "Name",
    Ascending: true,
})
```

### Line Splitting
```go
// v1
settings.LineSplitColumn = "Description"
settings.LineSplitSeparator = ","

// v2
output.WithTransformer(&output.LineSplitTransformer{
    Column: "Description",
    Separator: ",",
})
```

### Table of Contents (Markdown)
```go
// v1
settings.HasTOC = true

// v2
output.WithTOC(true)
```

### Front Matter (Markdown)
```go
// v1
settings.FrontMatter = map[string]string{
    "title": "Report",
    "date": "2024-01-01",
}

// v2
output.WithFrontMatter(map[string]string{
    "title": "Report",
    "date": "2024-01-01",
})
```

## Common Issues

### 1. Context Required
v2 requires a context for all rendering operations:
```go
ctx := context.Background()
// or with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### 2. Key Order Not Preserved
Always use `WithKeys()` or `WithSchema()` to ensure key order:
```go
// This may not preserve order
doc := output.New().Table("", data).Build()

// This will preserve order
doc := output.New().
    Table("", data, output.WithKeys("ID", "Name", "Status")).
    Build()
```

### 3. Multiple Outputs
v2 handles multiple outputs more elegantly:
```go
// Create once, render to multiple formats/destinations
doc := output.New().Table("", data).Build()

out := output.NewOutput(
    output.WithFormat(output.JSON),
    output.WithFormat(output.CSV),
    output.WithWriter(&output.StdoutWriter{}),
    output.WithWriter(output.NewFileWriter(".", "report.csv")),
)

err := out.Render(ctx, doc)
```

### 4. Error Handling
v2 provides better error context:
```go
err := out.Render(ctx, doc)
if err != nil {
    var renderErr *output.RenderError
    if errors.As(err, &renderErr) {
        log.Printf("Failed to render %s: %v", renderErr.Format, renderErr.Cause)
    }
}
```

## Examples

### Complete Example: Report Generation
```go
package main

import (
    "context"
    "log"
    
    "github.com/ArjenSchwarz/go-output/v2"
)

func main() {
    ctx := context.Background()
    
    // Sample data
    users := []map[string]interface{}{
        {"ID": 1, "Name": "Alice", "Role": "Admin"},
        {"ID": 2, "Name": "Bob", "Role": "User"},
    }
    
    stats := []map[string]interface{}{
        {"Metric": "Total Users", "Value": 2},
        {"Metric": "Active Sessions", "Value": 5},
    }
    
    // Build document
    doc := output.New().
        Header("System Report").
        Table("Users", users, output.WithKeys("ID", "Name", "Role")).
        Table("Statistics", stats, output.WithKeys("Metric", "Value")).
        Build()
    
    // Configure output
    out := output.NewOutput(
        // Multiple formats
        output.WithFormat(output.Table),
        output.WithFormat(output.JSON),
        
        // Multiple destinations
        output.WithWriter(&output.StdoutWriter{}),
        output.WithWriter(output.NewFileWriter(".", "report.json")),
        
        // Transformers
        output.WithTransformer(&output.ColorTransformer{}),
        
        // Table styling
        output.WithTableStyle("ColoredBright"),
    )
    
    // Render
    if err := out.Render(ctx, doc); err != nil {
        log.Fatalf("Failed to render: %v", err)
    }
}
```

### Example: Progress with Data Processing
```go
package main

import (
    "context"
    "time"
    
    "github.com/ArjenSchwarz/go-output/v2"
)

func processData(ctx context.Context) {
    // Create progress indicator
    progress := output.NewProgress(output.Table,
        output.WithProgressColor(output.ProgressColorGreen),
        output.WithProgressStatus("Processing records"),
    )
    
    // Set total items
    progress.SetTotal(100)
    
    // Process data
    for i := 0; i < 100; i++ {
        // Simulate work
        time.Sleep(50 * time.Millisecond)
        
        // Update progress
        progress.Increment(1)
        progress.SetStatus(fmt.Sprintf("Processing record %d/100", i+1))
        
        // Check context cancellation
        select {
        case <-ctx.Done():
            progress.Fail(ctx.Err())
            return
        default:
        }
    }
    
    // Complete
    progress.Complete()
}
```

## Need Help?

- Check the [API documentation](https://pkg.go.dev/github.com/ArjenSchwarz/go-output/v2)
- Review the [examples](https://github.com/ArjenSchwarz/go-output/tree/main/v2/examples)
- Report issues at [GitHub Issues](https://github.com/ArjenSchwarz/go-output/issues)

The migration tool can handle most common patterns automatically. For complex migrations, refer to this guide and the API documentation.