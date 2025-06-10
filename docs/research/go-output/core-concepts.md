# Core Concepts

Understanding the core data structures and concepts is essential for effectively using the go-output library.

## OutputHolder

An `OutputHolder` represents a single entity or record in your dataset. It contains a map of key-value pairs where keys are column names and values can be any Go type.

### Structure

```go
type OutputHolder struct {
    Contents map[string]interface{}
}
```

### Usage

```go
// Simple data
user := format.OutputHolder{
    Contents: map[string]interface{}{
        "ID":     123,
        "Name":   "John Doe",
        "Email":  "john@example.com",
        "Active": true,
        "Score":  95.5,
    },
}

// Complex data with nested structures
server := format.OutputHolder{
    Contents: map[string]interface{}{
        "Name":      "web-server-01",
        "Status":    "running",
        "Uptime":    "5d 12h 30m",
        "Memory":    "8.2GB",
        "CPU":       "45%",
        "Processes": []string{"nginx", "php-fpm", "mysql"},
        "Metadata":  map[string]string{"env": "production", "region": "us-east-1"},
    },
}
```

### Best Practices

- Use consistent key names across OutputHolders
- Keep data types consistent for the same keys
- Use descriptive key names that will work well as column headers
- For complex data, consider flattening nested structures

## OutputArray

The `OutputArray` is the main container that holds your data collection, output configuration, and column definitions.

### Structure

```go
type OutputArray struct {
    Settings *OutputSettings  // Configuration for output generation
    Contents []OutputHolder   // Collection of data records
    Keys     []string        // Column names to include in output
}
```

### Creating an OutputArray

```go
// Method 1: Direct creation
output := format.OutputArray{
    Settings: settings,
    Contents: []format.OutputHolder{user1, user2, user3},
    Keys:     []string{"Name", "Email", "Status"},
}

// Method 2: Build incrementally
output := format.OutputArray{
    Settings: format.NewOutputSettings(),
    Contents: make([]format.OutputHolder, 0),
    Keys:     []string{},
}

// Add data
output.Contents = append(output.Contents, user1)
output.Contents = append(output.Contents, user2)
output.Keys = []string{"Name", "Email", "Status"}
```

### Key Management

The `Keys` slice determines which columns appear in the output and their order:

```go
// Only show specific columns
output.Keys = []string{"Name", "Status"}

// Show all available keys (you need to collect them yourself)
allKeys := collectAllKeys(output.Contents)
output.Keys = allKeys

// Custom order
output.Keys = []string{"Status", "Name", "ID", "Email"}
```

## OutputSettings

The `OutputSettings` struct controls how your data is formatted and where it's output.

### Structure Overview

```go
type OutputSettings struct {
    // Output control
    OutputFormat      string
    OutputFile        string
    OutputFileFormat  string

    // Display options
    Title            string
    UseColors        bool
    UseEmoji         bool
    HasTOC          bool
    SeparateTables  bool
    ShouldAppend    bool
    SplitLines      bool

    // Sorting and organization
    SortKey         string

    // Table styling
    TableStyle          table.Style
    TableMaxColumnWidth int

    // Graphical output settings
    FromToColumns   *FromToColumns
    MermaidSettings *mermaid.Settings
    DrawIOHeader    drawio.Header

    // Markdown specific
    FrontMatter map[string]string

    // S3 integration
    S3Bucket S3Output
}
```

### Default Settings

```go
settings := format.NewOutputSettings()
// Creates settings with:
// - TableStyle: table.StyleDefault
// - TableMaxColumnWidth: 50
// - MermaidSettings: empty mermaid.Settings{}
```

### Common Configurations

```go
settings := format.NewOutputSettings()

// Basic output configuration
settings.SetOutputFormat("json")
settings.Title = "Data Export"
settings.OutputFile = "export.json"

// Visual enhancements
settings.UseColors = true
settings.UseEmoji = true
settings.HasTOC = true

// Table styling
settings.TableStyle = format.TableStyles["ColoredBright"]
settings.TableMaxColumnWidth = 30
settings.SeparateTables = true

// Sorting
settings.SortKey = "Name"
```

## Data Type Handling

The library handles Go data types intelligently:

### Supported Types

- **Primitives**: string, int, float64, bool
- **Collections**: []string, []interface{}
- **Maps**: map[string]string, map[string]interface{}
- **Custom**: Any type that can be converted to string

### Type Conversion

For most output formats, values are converted to strings using the internal `toString()` method:

```go
// Automatic conversion examples:
"hello"           → "hello"
123               → "123"
45.67             → "45.67"
true              → "true"
[]string{"a","b"} → "a, b" (with appropriate separator)
```

### Special Handling

- **JSON/YAML**: Preserves original data types
- **Mermaid Pie Charts**: Requires float64 values for numeric data
- **DrawIO**: Uses string representations for CSV import
- **Split Lines**: Arrays can be split into separate rows when `SplitLines = true`

## Relationships and Connections

For graphical outputs (Mermaid, DOT, DrawIO), you can define relationships between data items.

### FromToColumns

```go
type FromToColumns struct {
    From  string  // Source column name
    To    string  // Target column name
    Label string  // Optional edge label column
}

// Usage
settings.AddFromToColumns("Parent", "Child")
// or with label
settings.AddFromToColumnsWithLabel("Source", "Target", "ConnectionType")
```

### Example Data for Relationships

```go
connections := []format.OutputHolder{
    {
        Contents: map[string]interface{}{
            "Parent": "Database",
            "Child":  "API Server",
            "Type":   "queries",
        },
    },
    {
        Contents: map[string]interface{}{
            "Parent": "API Server",
            "Child":  "Web Frontend",
            "Type":   "HTTP",
        },
    },
}

settings.AddFromToColumnsWithLabel("Parent", "Child", "Type")
```

## Memory and Performance

### Buffer Management

The library uses an internal buffer for certain operations:

- Buffer is automatically managed
- Reset after each write operation
- Used for incremental content building

### Large Data Sets

For large datasets:

- Consider processing in chunks
- Use file output instead of stdout
- Be mindful of memory usage with complex nested data

### Best Practices

1. **Consistent Structure**: Keep OutputHolder contents consistent
2. **Key Management**: Define Keys explicitly for predictable output
3. **Type Consistency**: Use the same data types for the same logical fields
4. **Memory Efficiency**: Don't hold unnecessary data in memory
5. **Error Handling**: The library uses log.Fatal for errors, plan accordingly
