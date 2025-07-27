# Go-Output v2 Collapsible API Reference

This document provides a comprehensive reference for all collapsible content APIs available in go-output v2, specifically for implementing collapsible formatters for property changes and dependencies in Strata.

## Table of Contents
- [Core Interfaces](#core-interfaces)
- [CollapsibleValue API](#collapsiblevalue-api)
- [CollapsibleSection API](#collapsiblesection-api)
- [Built-in Formatters](#built-in-formatters)
- [Configuration Options](#configuration-options)
- [Usage Examples](#usage-examples)
- [Implementation Patterns](#implementation-patterns)

## Core Interfaces

### CollapsibleValue Interface

The main interface for creating expandable content in table cells and fields:

```go
type CollapsibleValue interface {
    Summary() string                              // Collapsed view text
    Details() any                                 // Expanded content (any type)
    IsExpanded() bool                            // Default expansion state
    FormatHint(format string) map[string]any     // Format-specific rendering hints
}
```

### CollapsibleSection Interface

Interface for section-level collapsible content containing multiple content blocks:

```go
type CollapsibleSection interface {
    Content // Inherits Content interface
    
    Title() string                               // Section title/summary
    Content() []Content                          // Nested content items
    IsExpanded() bool                           // Default expansion state
    Level() int                                 // Nesting level (0-3)
    FormatHint(format string) map[string]any    // Format-specific hints
}
```

## CollapsibleValue API

### Constructor

```go
func NewCollapsibleValue(summary string, details any, opts ...CollapsibleOption) *DefaultCollapsibleValue
```

**Parameters:**
- `summary`: Text shown in collapsed state
- `details`: Content shown when expanded (can be any type: string, []string, map[string]any, etc.)
- `opts`: Configuration options

### Configuration Options

```go
type CollapsibleOption func(*DefaultCollapsibleValue)

// Set default expansion state
func WithExpanded(expanded bool) CollapsibleOption

// Set maximum character length for details before truncation  
func WithMaxLength(length int) CollapsibleOption

// Set truncation indicator text
func WithTruncateIndicator(indicator string) CollapsibleOption

// Add format-specific rendering hints
func WithFormatHint(format string, hints map[string]any) CollapsibleOption
```

### Methods

```go
// Summary returns the collapsed view with fallback handling
func (d *DefaultCollapsibleValue) Summary() string

// Details returns the expanded content with character limit truncation
func (d *DefaultCollapsibleValue) Details() any

// IsExpanded returns whether this should be expanded by default
func (d *DefaultCollapsibleValue) IsExpanded() bool

// FormatHint returns renderer-specific hints for the given format
func (d *DefaultCollapsibleValue) FormatHint(format string) map[string]any

// String implements the Stringer interface for debugging
func (d *DefaultCollapsibleValue) String() string
```

## CollapsibleSection API

### Constructors

```go
// Main constructor
func NewCollapsibleSection(title string, content []Content, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection

// Helper for single table
func NewCollapsibleTable(title string, table *TableContent, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection

// Helper for multiple tables
func NewCollapsibleMultiTable(title string, tables []*TableContent, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection

// Helper for mixed content types
func NewCollapsibleReport(title string, content []Content, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection
```

### Configuration Options

```go
type CollapsibleSectionOption func(*DefaultCollapsibleSection)

// Set whether section should be expanded by default
func WithSectionExpanded(expanded bool) CollapsibleSectionOption

// Set nesting level (0-3 supported)
func WithSectionLevel(level int) CollapsibleSectionOption

// Add format-specific hints for the section
func WithSectionFormatHint(format string, hints map[string]any) CollapsibleSectionOption
```

### Methods

```go
// Title returns the section title/summary
func (cs *DefaultCollapsibleSection) Title() string

// Content returns the nested content items (returns copy)
func (cs *DefaultCollapsibleSection) Content() []Content

// IsExpanded returns whether this section should be expanded by default
func (cs *DefaultCollapsibleSection) IsExpanded() bool

// Level returns the nesting level (0-3 supported)
func (cs *DefaultCollapsibleSection) Level() int

// FormatHint provides renderer-specific hints
func (cs *DefaultCollapsibleSection) FormatHint(format string) map[string]any
```

## Built-in Formatters

These pre-built formatters return CollapsibleValue instances and can be used directly in Field.Formatter:

### ErrorListFormatter

For arrays of strings or errors:

```go
func ErrorListFormatter(opts ...CollapsibleOption) func(any) any
```

**Usage:**
```go
output.Field{
    Name: "errors",
    Type: "array",
    Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
}
```

**Behavior:**
- Input: `[]string` or `[]error`
- Summary: `"3 errors (click to expand)"`
- Details: Array of error strings
- Graceful fallback for incompatible types

### FilePathFormatter

For long file paths:

```go
func FilePathFormatter(maxLength int, opts ...CollapsibleOption) func(any) any
```

**Usage:**
```go
output.Field{
    Name: "file_path",
    Type: "string", 
    Formatter: output.FilePathFormatter(30, output.WithExpanded(false)),
}
```

**Behavior:**
- Input: `string`
- Summary: `"...ComponentName.tsx (show full path)"` (abbreviated)
- Details: Full file path
- No collapsible for paths <= maxLength

### JSONFormatter

For complex data structures:

```go
func JSONFormatter(maxLength int, opts ...CollapsibleOption) func(any) any
```

**Usage:**
```go
output.Field{
    Name: "config",
    Type: "object",
    Formatter: output.JSONFormatter(100, output.WithExpanded(false)),
}
```

**Behavior:**
- Input: Any JSON-serializable value
- Summary: `"JSON data (245 bytes)"`
- Details: Pretty-formatted JSON string
- No collapsible for small data or marshal errors

### CollapsibleFormatter

Generic formatter for custom patterns:

```go
func CollapsibleFormatter(summaryTemplate string, detailFunc func(any) any, opts ...CollapsibleOption) func(any) any
```

**Usage:**
```go
formatter := output.CollapsibleFormatter(
    "Dependencies: %v",
    func(val any) any {
        // Transform val into detailed representation
        return detailsData
    },
    output.WithExpanded(false),
)

output.Field{
    Name: "dependencies",
    Type: "array",
    Formatter: formatter,
}
```

## Configuration Options

### Global Configuration

Configure collapsible behavior per renderer:

```go
type CollapsibleConfig struct {
    GlobalExpansion      bool              // Override all IsExpanded() settings
    MaxDetailLength      int               // Character limit for details (default: 500)
    TruncateIndicator    string            // Truncation suffix (default: "[...truncated]")
    TableHiddenIndicator string            // Table collapse indicator
    HTMLCSSClasses       map[string]string // Custom CSS classes for HTML
}

// Apply to output
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion:      false,
        TableHiddenIndicator: "[click to expand]",
        MaxDetailLength:      200,
    }),
    output.WithWriter(output.NewStdoutWriter()),
)
```

## Usage Examples

### Basic Property Changes Formatter

```go
func propertyChangesFormatter(val any) any {
    if changes, ok := val.([]PropertyChange); ok && len(changes) > 0 {
        return output.NewCollapsibleValue(
            fmt.Sprintf("%d property changes", len(changes)),
            changes,
            output.WithExpanded(false),
            output.WithMaxLength(500),
        )
    }
    return val
}

// Use in schema
schema := output.WithSchema(
    output.Field{
        Name: "property_changes",
        Type: "array",
        Formatter: propertyChangesFormatter,
    },
)
```

### Dependencies Formatter

```go
func dependenciesFormatter(val any) any {
    if deps, ok := val.([]Dependency); ok && len(deps) > 0 {
        // Create summary
        summary := fmt.Sprintf("%d dependencies", len(deps))
        
        // Create structured details
        details := map[string]any{
            "count": len(deps),
            "dependencies": deps,
            "types": countDependencyTypes(deps),
        }
        
        return output.NewCollapsibleValue(
            summary,
            details,
            output.WithExpanded(false),
        )
    }
    return val
}
```

### Complex Resource Analysis

```go
func resourceAnalysisFormatter(val any) any {
    if analysis, ok := val.(ResourceAnalysis); ok {
        changes := analysis.PropertyChanges
        deps := analysis.Dependencies
        
        if len(changes) == 0 && len(deps) == 0 {
            return "No changes"
        }
        
        summary := fmt.Sprintf("%d changes, %d dependencies", len(changes), len(deps))
        
        details := map[string]any{
            "property_changes": changes,
            "dependencies": deps,
            "risk_level": analysis.RiskLevel,
            "recommendations": analysis.Recommendations,
        }
        
        return output.NewCollapsibleValue(
            summary,
            details,
            output.WithExpanded(false),
            output.WithFormatHint("markdown", map[string]any{
                "expandable": true,
                "syntax_highlight": "json",
            }),
        )
    }
    return val
}
```

### Collapsible Section for Related Tables

```go
// Create detailed analysis tables
changesTable := output.NewTableContent("Property Changes", propertyData, 
    output.WithSchema(
        output.Field{Name: "property", Type: "string"},
        output.Field{Name: "old_value", Type: "string"},
        output.Field{Name: "new_value", Type: "string"},
        output.Field{Name: "risk", Type: "string"},
    ))

depsTable := output.NewTableContent("Dependencies", dependencyData,
    output.WithSchema(
        output.Field{Name: "resource", Type: "string"},
        output.Field{Name: "dependency_type", Type: "string"},
        output.Field{Name: "impact", Type: "string"},
    ))

// Wrap in collapsible section
analysisSection := output.NewCollapsibleReport(
    "Detailed Resource Analysis",
    []output.Content{
        output.NewTextContent("Analysis of resource changes and dependencies"),
        changesTable,
        depsTable,
        output.NewTextContent("Review changes carefully before applying"),
    },
    output.WithSectionExpanded(false),
)

doc := output.New().
    Header("Terraform Plan Analysis").
    Add(analysisSection).
    Build()
```

## Implementation Patterns

### Pattern 1: Simple Array Collapsible

```go
func simpleArrayFormatter(val any) any {
    if arr, ok := val.([]string); ok && len(arr) > 2 {
        return output.NewCollapsibleValue(
            fmt.Sprintf("%d items", len(arr)),
            arr,
            output.WithExpanded(false),
        )
    }
    return val
}
```

### Pattern 2: Conditional Collapsible

```go
func conditionalFormatter(val any) any {
    if data, ok := val.(ComplexData); ok {
        // Only make collapsible if it's actually complex
        if len(data.Details) > 5 || data.HasImportantInfo {
            return output.NewCollapsibleValue(
                data.Summary,
                data.Details,
                output.WithExpanded(data.Critical),
            )
        }
    }
    return val // Return as-is for simple data
}
```

### Pattern 3: Structured Details

```go
func structuredFormatter(val any) any {
    if resource, ok := val.(TerraformResource); ok {
        details := map[string]any{
            "type": resource.Type,
            "name": resource.Name,
            "properties": resource.Properties,
            "changes": resource.Changes,
            "dependencies": resource.Dependencies,
            "metadata": map[string]any{
                "provider": resource.Provider,
                "module": resource.Module,
                "count": resource.Count,
            },
        }
        
        return output.NewCollapsibleValue(
            fmt.Sprintf("%s.%s", resource.Type, resource.Name),
            details,
            output.WithExpanded(false),
            output.WithMaxLength(300),
        )
    }
    return val
}
```

### Pattern 4: Multi-level Nested Sections

```go
// Create nested hierarchy for complex reports
subSection1 := output.NewCollapsibleTable(
    "Property Changes",
    propertyTable,
    output.WithSectionLevel(2),
    output.WithSectionExpanded(false),
)

subSection2 := output.NewCollapsibleTable(
    "Dependency Analysis", 
    dependencyTable,
    output.WithSectionLevel(2),
    output.WithSectionExpanded(false),
)

mainSection := output.NewCollapsibleReport(
    "Resource Analysis",
    []output.Content{
        output.NewTextContent("Comprehensive analysis results"),
        subSection1,
        subSection2,
    },
    output.WithSectionLevel(1),
    output.WithSectionExpanded(true),
)
```

## Cross-Format Rendering

Collapsible content automatically adapts to each output format:

| Format   | CollapsibleValue Rendering | CollapsibleSection Rendering |
|----------|----------------------------|------------------------------|
| **Markdown** | `<details><summary>Summary</summary>Details</details>` | Nested `<details>` structure |
| **JSON** | `{"type": "collapsible", "summary": "...", "details": [...]}` | Structured data with content array |
| **YAML** | YAML mapping with summary/details fields | YAML structure with nested content |
| **HTML** | Semantic `<details>` with CSS classes | Section elements with collapsible behavior |
| **Table** | Summary + expansion indicator | Section headers with indented content |
| **CSV** | Summary + automatic detail columns | Metadata comments with table data |

## Error Handling

Built-in formatters include error handling:

- **Type mismatches**: Gracefully return original value
- **Null/empty data**: Return original value or appropriate fallback
- **Nested CollapsibleValues**: Prevented to avoid infinite loops
- **Marshal errors**: Fall back to simple representation

## Performance Considerations

- **Lazy evaluation**: Details are processed only when accessed
- **Memory optimization**: Large content uses optimized processing
- **Caching**: Processed details are cached to avoid redundant work
- **Format hints**: Loaded only when needed for specific formats

This API provides comprehensive support for implementing collapsible formatters for property changes and dependencies in Strata's enhanced summary visualization feature.