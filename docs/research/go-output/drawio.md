# Draw.io Integration

The go-output library provides comprehensive support for generating Draw.io/diagrams.net compatible CSV files for creating visual diagrams from data.

## Overview

Draw.io (now diagrams.net) supports importing CSV files to automatically generate diagrams. The go-output library creates properly formatted CSV files with special header comments that control how the diagram is rendered.

## Basic Usage

```go
package main

import (
    "github.com/ArjenSchwarz/go-output"
    "github.com/ArjenSchwarz/go-output/drawio"
)

func main() {
    // Create output array
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Name", "Type", "Parent", "Description"},
    }

    // Configure Draw.io header
    header := drawio.NewHeader("%Name%", "%Type%", "Type")
    output.Settings.DrawIOHeader = header
    output.Settings.SetOutputFormat("drawio")

    // Add data
    output.AddContents(map[string]interface{}{
        "Name":        "Frontend",
        "Type":        "application",
        "Parent":      "System",
        "Description": "Web application frontend",
    })

    output.AddContents(map[string]interface{}{
        "Name":        "Backend",
        "Type":        "service",
        "Parent":      "System",
        "Description": "API backend service",
    })

    // Generate Draw.io CSV
    output.Write()
}
```

## Header Configuration

The Draw.io header controls how nodes and connections appear in the diagram:

### Creating Headers

```go
// Basic header with label, style, and ignore columns
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Default header (label: %Name%, style: %Image%, ignore: Image)
header := drawio.DefaultHeader()
```

### Header Properties

| Property | Description | Example |
|----------|-------------|---------|
| **Label** | Node display text with placeholders | `%Name%` |
| **Style** | Node style with placeholders | `%Type%` |
| **Ignore** | Comma-separated list of columns to exclude from metadata | `"Type,Internal"` |

## Layout Options

Draw.io supports multiple automatic layout algorithms:

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Set layout type
header.SetLayout(drawio.LayoutAuto)           // Default auto layout
header.SetLayout(drawio.LayoutHorizontalFlow) // Left-to-right flow
header.SetLayout(drawio.LayoutVerticalFlow)   // Top-to-bottom flow
header.SetLayout(drawio.LayoutHorizontalTree) // Horizontal tree
header.SetLayout(drawio.LayoutVerticalTree)   // Vertical tree
header.SetLayout(drawio.LayoutOrganic)        // Organic/force-directed
header.SetLayout(drawio.LayoutCircle)         // Circular layout
header.SetLayout(drawio.LayoutNone)           // Manual positioning
```

## Node Styling

### Basic Node Configuration

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Set node dimensions
header.SetHeightAndWidth("100", "150")  // Fixed size in pixels
header.SetHeightAndWidth("auto", "auto") // Auto-size (default)
header.SetHeightAndWidth("@Height", "@Width") // Use column values

// Set padding for auto-sized nodes
header.SetPadding(10)
```

### Advanced Styling with Placeholders

```go
// Use column values in styles
header := drawio.NewHeader(
    "%Name%\n%Description%",                    // Multi-line label
    "shape=%Shape%;fillColor=%Color%;fontSize=12", // Dynamic styling
    "Shape,Color"                               // Hide style columns
)
```

## Connections and Relationships

### Simple Connections

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Create connection from Parent column to Name column
connection := drawio.NewConnection()
connection.From = "Parent"
connection.To = "Name"
connection.Style = drawio.DefaultConnectionStyle

header.AddConnection(connection)
```

### Bidirectional Connections

```go
connection := drawio.NewConnection()
connection.From = "Source"
connection.To = "Target"
connection.Style = drawio.BidirectionalConnectionStyle
connection.Label = "%ConnectionType%"

header.AddConnection(connection)
```

### Multiple Connection Types

```go
// Network connections
networkConn := drawio.NewConnection()
networkConn.From = "FromNode"
networkConn.To = "ToNode"
networkConn.Style = "curved=1;endArrow=blockThin;endFill=1;strokeColor=#0066CC;"

// Data flow connections
dataConn := drawio.NewConnection()
dataConn.From = "Source"
dataConn.To = "Destination"
dataConn.Style = "curved=0;endArrow=classic;strokeColor=#FF6600;"
dataConn.Label = "%DataType%"

header.AddConnection(networkConn)
header.AddConnection(dataConn)
```

## Hierarchical Diagrams

### Parent-Child Relationships

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Configure identity for node updates
header.SetIdentity("ID")

// Set up parent-child relationships
header.SetParent("ParentID", drawio.DefaultParentStyle)

// Custom parent style
customParentStyle := "swimlane;whiteSpace=wrap;html=1;childLayout=stackLayout;horizontal=1;"
header.SetParent("ParentID", customParentStyle)
```

### Organizational Charts

```go
func createOrgChart() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Name", "Title", "Manager", "Department"},
    }

    header := drawio.NewHeader(
        "%Name%\n%Title%",
        "shape=mxgraph.flowchart.person;fillColor=#dae8fc;strokeColor=#6c8ebf;",
        "Department"
    )

    // Connect employees to managers
    connection := drawio.NewConnection()
    connection.From = "Manager"
    connection.To = "Name"
    connection.Style = "curved=0;endArrow=none;endFill=0;"

    header.AddConnection(connection)
    header.SetLayout(drawio.LayoutVerticalTree)

    output.Settings.DrawIOHeader = header
    output.Settings.SetOutputFormat("drawio")

    // Add employee data
    output.AddContents(map[string]interface{}{
        "Name":       "John Smith",
        "Title":      "CEO",
        "Manager":    "",
        "Department": "Executive",
    })

    output.AddContents(map[string]interface{}{
        "Name":       "Jane Doe",
        "Title":      "CTO",
        "Manager":    "John Smith",
        "Department": "Technology",
    })

    output.Write()
}
```

## AWS Architecture Diagrams

The library includes built-in AWS shapes for creating architecture diagrams:

```go
import "github.com/ArjenSchwarz/go-output/drawio"

// Get AWS service shape
ec2Shape := drawio.AWSShape("Compute", "EC2")
s3Shape := drawio.AWSShape("Storage", "S3")
rdsShape := drawio.AWSShape("Database", "RDS")

// Create header with AWS shapes
header := drawio.NewHeader(
    "%ServiceName%",
    drawio.AWSShape("%Category%", "%ServiceType%"),
    "Category,ServiceType"
)
```

### AWS Infrastructure Example

```go
func createAWSArchitecture() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"ServiceName", "Category", "ServiceType", "ConnectedTo"},
    }

    header := drawio.NewHeader(
        "%ServiceName%",
        drawio.AWSShape("%Category%", "%ServiceType%"),
        "Category,ServiceType"
    )

    // Connect services
    connection := drawio.NewConnection()
    connection.From = "ServiceName"
    connection.To = "ConnectedTo"
    header.AddConnection(connection)

    output.Settings.DrawIOHeader = header
    output.Settings.SetOutputFormat("drawio")

    // Add AWS services
    output.AddContents(map[string]interface{}{
        "ServiceName": "Web Server",
        "Category":    "Compute",
        "ServiceType": "EC2",
        "ConnectedTo": "Database",
    })

    output.AddContents(map[string]interface{}{
        "ServiceName": "Database",
        "Category":    "Database",
        "ServiceType": "RDS",
        "ConnectedTo": "",
    })

    output.Write()
}
```

## Advanced Configuration

### Spacing and Layout Control

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Set spacing between elements
header.SetSpacing(
    50,  // Node spacing
    120, // Level spacing
    30   // Edge spacing
)

// Get current spacing values
nodeSpacing, levelSpacing, edgeSpacing := header.GetSpacing()
```

### Manual Positioning

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Use manual positioning (layout must be "none")
header.SetLayout(drawio.LayoutNone)
header.SetLeftAndTopColumns("X", "Y")

// Data should include X and Y coordinates
output.AddContents(map[string]interface{}{
    "Name": "Node1",
    "Type": "service",
    "X":    100,
    "Y":    200,
})
```

### Links and Metadata

```go
header := drawio.NewHeader("%Name%", "%Type%", "Type")

// Set link column for clickable nodes
header.SetLink("URL")

// Set namespace to avoid ID conflicts
header.SetNamespace("myapp-")

// Data with links
output.AddContents(map[string]interface{}{
    "Name": "Service Dashboard",
    "Type": "dashboard",
    "URL":  "https://dashboard.example.com",
})
```

## File Operations

### Reading Draw.io CSV Files

```go
// Read header and contents from existing CSV
headers, contents := drawio.GetHeaderAndContentsFromFile("diagram.csv")

// Read as string maps
data := drawio.GetContentsFromFileAsStringMaps("diagram.csv")

// Process the data
for _, row := range data {
    fmt.Printf("Node: %s, Type: %s\n", row["Name"], row["Type"])
}
```

### Creating CSV Files Directly

```go
// Create CSV without using OutputArray
header := drawio.NewHeader("%Name%", "%Type%", "Type")
headerRow := []string{"Name", "Type", "Parent"}
contents := []map[string]string{
    {"Name": "Frontend", "Type": "app", "Parent": "System"},
    {"Name": "Backend", "Type": "service", "Parent": "System"},
}

drawio.CreateCSV(header, headerRow, contents, "output.csv")
```

## Best Practices

### 1. Use Descriptive Labels

```go
// Good: Multi-line labels with context
header := drawio.NewHeader(
    "%ServiceName%\n%Environment%\n%Version%",
    "%ServiceType%",
    "ServiceType"
)

// Avoid: Single cryptic labels
header := drawio.NewHeader("%ID%", "%T%", "T")
```

### 2. Consistent Styling

```go
// Define style constants
const (
    ServiceStyle  = "shape=process;fillColor=#dae8fc;strokeColor=#6c8ebf;"
    DatabaseStyle = "shape=cylinder3;fillColor=#f8cecc;strokeColor=#b85450;"
    QueueStyle    = "shape=parallelogram;fillColor=#fff2cc;strokeColor=#d6b656;"
)

// Use style mapping
header := drawio.NewHeader(
    "%Name%",
    "%StyleType%",  // Map to style constants
    "StyleType"
)
```

### 3. Organize Complex Diagrams

```go
// Use parent-child relationships for grouping
header := drawio.NewHeader("%Name%", "%Type%", "Type")
header.SetIdentity("ID")
header.SetParent("ParentID", drawio.DefaultParentStyle)

// Group related services
output.AddContents(map[string]interface{}{
    "ID":       "frontend-group",
    "Name":     "Frontend Services",
    "Type":     "group",
    "ParentID": "",
})

output.AddContents(map[string]interface{}{
    "ID":       "web-app",
    "Name":     "Web Application",
    "Type":     "service",
    "ParentID": "frontend-group",
})
```

### 4. Error Handling

```go
// Validate header configuration
if !header.IsSet() {
    log.Fatal("Draw.io header is not properly configured")
}

// Check required settings
if output.Settings.OutputFormat == "drawio" && !output.Settings.DrawIOHeader.IsSet() {
    log.Fatal("DrawIO header must be configured for drawio output format")
}
```

## Connection Styles Reference

| Style | Description | Use Case |
|-------|-------------|----------|
| `DefaultConnectionStyle` | Standard directed arrow | General relationships |
| `BidirectionalConnectionStyle` | Two-way arrow | Bidirectional communication |
| `curved=0;endArrow=none;` | Straight line, no arrow | Grouping/containment |
| `curved=1;endArrow=classic;` | Curved line with classic arrow | Data flow |
| `strokeColor=#FF0000;strokeWidth=3;` | Colored thick line | Critical paths |

## Layout Algorithm Comparison

| Layout | Best For | Description |
|--------|----------|-------------|
| **Auto** | General purpose | Automatic best-fit layout |
| **HorizontalFlow** | Process flows | Left-to-right progression |
| **VerticalFlow** | Hierarchies | Top-to-bottom flow |
| **HorizontalTree** | Org charts | Horizontal tree structure |
| **VerticalTree** | Decision trees | Vertical tree structure |
| **Organic** | Network diagrams | Force-directed positioning |
| **Circle** | Cycle diagrams | Circular arrangement |
| **None** | Custom layouts | Manual positioning |

The Draw.io integration provides powerful capabilities for creating professional diagrams directly from your data, with extensive customization options for styling, layout, and interactivity.
