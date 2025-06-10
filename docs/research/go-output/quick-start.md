# Quick Start

This guide will get you up and running with the go-output library in minutes.

## Installation

```bash
go get github.com/ArjenSchwarz/go-output
```

## Basic Usage

### 1. Import the Package

```go
import "github.com/ArjenSchwarz/go-output"
```

### 2. Create Your Data

```go
// Create individual data items
item1 := format.OutputHolder{
    Contents: map[string]interface{}{
        "Name":   "Alice",
        "Age":    30,
        "City":   "New York",
        "Active": true,
    },
}

item2 := format.OutputHolder{
    Contents: map[string]interface{}{
        "Name":   "Bob",
        "Age":    25,
        "City":   "London",
        "Active": false,
    },
}

// Define which keys to include in output
keys := []string{"Name", "Age", "City", "Active"}
```

### 3. Configure Output Settings

```go
// Create settings with default values
settings := format.NewOutputSettings()

// Set desired output format
settings.SetOutputFormat("table")  // Options: json, yaml, csv, table, html, markdown, mermaid, dot, drawio

// Optional: Add a title
settings.Title = "User Information"

// Optional: Enable colors and emoji
settings.UseColors = true
settings.UseEmoji = true
```

### 4. Create Output Array and Generate Output

```go
// Combine data and settings
output := format.OutputArray{
    Settings: settings,
    Contents: []format.OutputHolder{item1, item2},
    Keys:     keys,
}

// Generate output
output.Write()
```

## Complete Example

```go
package main

import "github.com/ArjenSchwarz/go-output"

func main() {
    // Create data
    users := []format.OutputHolder{
        {
            Contents: map[string]interface{}{
                "Name":   "Alice",
                "Age":    30,
                "City":   "New York",
                "Active": true,
            },
        },
        {
            Contents: map[string]interface{}{
                "Name":   "Bob",
                "Age":    25,
                "City":   "London",
                "Active": false,
            },
        },
    }

    // Configure output
    settings := format.NewOutputSettings()
    settings.SetOutputFormat("table")
    settings.Title = "User List"
    settings.UseColors = true

    // Generate output
    output := format.OutputArray{
        Settings: settings,
        Contents: users,
        Keys:     []string{"Name", "Age", "City", "Active"},
    }

    output.Write()
}
```

This will produce a formatted table like:

```
User List
┌───────┬─────┬──────────┬────────┐
│ NAME  │ AGE │ CITY     │ ACTIVE │
├───────┼─────┼──────────┼────────┤
│ Alice │ 30  │ New York │ true   │
│ Bob   │ 25  │ London   │ false  │
└───────┴─────┴──────────┴────────┘
```

## Output to File

To save output to a file instead of printing to stdout:

```go
settings.OutputFile = "users.json"
settings.SetOutputFormat("json")
```

## Different Output Formats

Change the format by simply changing the `OutputFormat`:

```go
// JSON output
settings.SetOutputFormat("json")

// YAML output
settings.SetOutputFormat("yaml")

// CSV output
settings.SetOutputFormat("csv")

// HTML output
settings.SetOutputFormat("html")

// Markdown table
settings.SetOutputFormat("markdown")
```

## Next Steps

- Learn about [Core Concepts](core-concepts.md) for deeper understanding
- Explore [Output Formats](output-formats.md) for format-specific features
- Check [Settings and Configuration](settings.md) for advanced options
- See [Examples](examples.md) for real-world usage patterns
