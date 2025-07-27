# Go-Output v2 API Documentation

## Overview

Go-Output v2 is a complete redesign of the library providing thread-safe document generation with preserved key ordering and multiple output formats. This API documentation covers all public interfaces and methods.

**Version**: v2.0.0
**Go Version**: 1.24+
**Import Path**: `github.com/ArjenSchwarz/go-output/v2`

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    output "github.com/ArjenSchwarz/go-output/v2"
)

func main() {
    // Create a document using the builder pattern
    doc := output.New().
        Table("Users", []map[string]any{
            {"Name": "Alice", "Age": 30, "Status": "Active"},
            {"Name": "Bob", "Age": 25, "Status": "Inactive"},
        }, output.WithKeys("Name", "Age", "Status")).
        Text("This is additional text content").
        Build()

    // Create output with multiple formats
    out := output.NewOutput(
        output.WithFormats(output.Table, output.JSON),
        output.WithWriter(output.NewStdoutWriter()),
    )

    // Render the document
    if err := out.Render(context.Background(), doc); err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Core Concepts

### Document-Builder Pattern

The v2 API eliminates global state by using an immutable Document-Builder pattern:

- **Document**: Immutable container for content and metadata
- **Builder**: Fluent API for constructing documents with thread-safe operations
- **Content**: Interface implemented by all content types

### Key Order Preservation

A fundamental feature that preserves exact user-specified column ordering:

- Key order is **never** alphabetized or reordered
- Each table maintains its own independent key ordering
- Supports multiple tables with different key sets

## Public API Reference

### Core Types

#### Document

Represents an immutable collection of content to be rendered.

```go
type Document struct {
    // Internal fields (not exported)
}

// GetContents returns a copy of the document's contents
func (d *Document) GetContents() []Content

// GetMetadata returns a copy of the document's metadata
func (d *Document) GetMetadata() map[string]any
```

**Thread Safety**: All methods are thread-safe using RWMutex.

#### Builder

Constructs documents using a fluent API pattern.

```go
type Builder struct {
    // Internal fields (not exported)
}

// New creates a new document builder
func New() *Builder

// Build finalizes and returns the document
func (b *Builder) Build() *Document

// HasErrors returns true if any errors occurred during building
func (b *Builder) HasErrors() bool

// Errors returns all errors that occurred during building
func (b *Builder) Errors() []error

// SetMetadata sets a metadata key-value pair
func (b *Builder) SetMetadata(key string, value any) *Builder
```

**Thread Safety**: All methods are thread-safe using Mutex.

### Content Types

#### Content Interface

All content types implement this interface:

```go
type Content interface {
    // Type returns the content type
    Type() ContentType

    // ID returns a unique identifier for this content
    ID() string

    // Encoding interfaces for efficient serialization
    encoding.TextAppender
    encoding.BinaryAppender
}
```

#### ContentType

Defines the type of content:

```go
type ContentType int

const (
    ContentTypeTable   ContentType = iota // Tabular data
    ContentTypeText                       // Unstructured text
    ContentTypeRaw                        // Format-specific content
    ContentTypeSection                    // Grouped content
)

// String returns the string representation
func (ct ContentType) String() string
```

#### TableContent

Represents tabular data with preserved key ordering:

```go
// NewTableContent creates a new table content
func NewTableContent(title string, data any, opts ...TableOption) (*TableContent, error)

// Methods
func (t *TableContent) Type() ContentType
func (t *TableContent) ID() string
func (t *TableContent) Title() string
func (t *TableContent) Schema() *Schema
func (t *TableContent) Records() []Record
```

**Key Features**:
- Preserves exact key order as specified by user
- Supports various data types ([]map[string]any, []Record, etc.)
- Thread-safe read operations

#### TextContent

Represents unstructured text with styling:

```go
// NewTextContent creates a new text content
func NewTextContent(text string, opts ...TextOption) *TextContent

// Methods
func (t *TextContent) Type() ContentType
func (t *TextContent) ID() string
func (t *TextContent) Text() string
func (t *TextContent) Style() TextStyle
```

#### RawContent

Represents format-specific content:

```go
// NewRawContent creates a new raw content
func NewRawContent(format string, data []byte, opts ...RawOption) (*RawContent, error)

// Methods
func (r *RawContent) Type() ContentType
func (r *RawContent) ID() string
func (r *RawContent) Format() string
func (r *RawContent) Data() []byte
```

#### SectionContent

Represents grouped content with hierarchical structure:

```go
// NewSectionContent creates a new section content
func NewSectionContent(title string, opts ...SectionOption) *SectionContent

// Methods
func (s *SectionContent) Type() ContentType
func (s *SectionContent) ID() string
func (s *SectionContent) Title() string
func (s *SectionContent) Level() int
func (s *SectionContent) Contents() []Content
func (s *SectionContent) AddContent(content Content)
```

### Builder Methods

#### Table Creation

```go
// Table adds a table with preserved key ordering
func (b *Builder) Table(title string, data any, opts ...TableOption) *Builder
```

**Table Options**:
- `WithKeys(keys ...string)` - Explicit key ordering (recommended)
- `WithSchema(fields ...Field)` - Full schema with formatters
- `WithAutoSchema()` - Auto-detect schema from data

**Example**:
```go
doc := output.New().
    Table("Users", userData, output.WithKeys("Name", "Email", "Status")).
    Table("Orders", orderData, output.WithKeys("ID", "Date", "Amount")).
    Build()
```

#### Text Content

```go
// Text adds text content with optional styling
func (b *Builder) Text(text string, opts ...TextOption) *Builder

// Header adds a header text (v1 compatibility)
func (b *Builder) Header(text string) *Builder
```

**Text Options**:
- `WithBold(bold bool)` - Bold text
- `WithItalic(italic bool)` - Italic text
- `WithColor(color string)` - Text color
- `WithHeader(header bool)` - Header styling

#### Raw Content

```go
// Raw adds format-specific raw content
func (b *Builder) Raw(format string, data []byte, opts ...RawOption) *Builder
```

**Supported Formats**: `html`, `css`, `js`, `json`, `xml`, `yaml`, `markdown`, `text`, `csv`, `dot`, `mermaid`, `drawio`, `svg`

#### Section Grouping

```go
// Section groups content under a heading
func (b *Builder) Section(title string, fn func(*Builder), opts ...SectionOption) *Builder
```

**Example**:
```go
doc := output.New().
    Section("User Data", func(b *output.Builder) {
        b.Table("Active Users", activeUsers, output.WithKeys("Name", "Email"))
        b.Table("Inactive Users", inactiveUsers, output.WithKeys("Name", "LastLogin"))
    }).
    Build()
```

#### Graph and Chart Methods

```go
// Graph adds graph content with edges
func (b *Builder) Graph(title string, edges []Edge) *Builder

// Chart adds a generic chart content
func (b *Builder) Chart(title, chartType string, data any) *Builder

// GanttChart adds a Gantt chart with tasks
func (b *Builder) GanttChart(title string, tasks []GanttTask) *Builder

// PieChart adds a pie chart with slices
func (b *Builder) PieChart(title string, slices []PieSlice, showData bool) *Builder

// DrawIO adds Draw.io diagram content
func (b *Builder) DrawIO(title string, records []Record, header DrawIOHeader) *Builder
```

### Schema System

#### Schema

Defines table structure with key ordering:

```go
type Schema struct {
    // Internal fields (not exported)
}

// NewSchemaFromKeys creates a schema from key names
func NewSchemaFromKeys(keys []string) *Schema

// DetectSchemaFromData auto-detects schema from data
func DetectSchemaFromData(data any) *Schema

// Methods
func (s *Schema) GetKeyOrder() []string
func (s *Schema) SetKeyOrder(keys []string)
func (s *Schema) FindField(name string) *Field
func (s *Schema) AddField(field Field)
func (s *Schema) GetFields() []Field
```

#### Field

Defines individual table columns:

```go
type Field struct {
    Name      string                    // Field name
    Type      string                    // Data type hint
    Formatter func(any) any           // Custom formatter (can return CollapsibleValue)
    Hidden    bool                      // Hide from output
}
```

### Collapsible Content System

The v2 library provides comprehensive support for collapsible content that adapts to each output format, enabling summary/detail views for complex data.

#### CollapsibleValue Interface

Core interface for creating expandable content in table cells:

```go
type CollapsibleValue interface {
    Summary() string                              // Collapsed view text
    Details() any                                 // Expanded content (any type)
    IsExpanded() bool                            // Default expansion state
    FormatHint(format string) map[string]any     // Format-specific rendering hints
}
```

**Usage**: Field formatters can return CollapsibleValue instances to create expandable content.

#### DefaultCollapsibleValue

Standard implementation with configuration options:

```go
type DefaultCollapsibleValue struct {
    // Internal fields (not exported)
}

// NewCollapsibleValue creates a collapsible value with options
func NewCollapsibleValue(summary string, details any, opts ...CollapsibleOption) *DefaultCollapsibleValue

// Configuration options
func WithExpanded(expanded bool) CollapsibleOption
func WithMaxLength(length int) CollapsibleOption
func WithFormatHint(format string, hints map[string]any) CollapsibleOption
```

**Example**:
```go
// Create collapsible error list
errorValue := output.NewCollapsibleValue(
    "3 errors found",
    []string{"Missing import", "Unused variable", "Type error"},
    output.WithExpanded(false),
    output.WithMaxLength(200),
)
```

#### Built-in Collapsible Formatters

Pre-built formatters for common patterns:

```go
// Error list formatter - collapses arrays of strings/errors
func ErrorListFormatter(opts ...CollapsibleOption) func(any) any

// File path formatter - shortens long paths with expandable details
func FilePathFormatter(maxLength int, opts ...CollapsibleOption) func(any) any

// JSON formatter - collapses large JSON objects
func JSONFormatter(maxLength int, opts ...CollapsibleOption) func(any) any

// Custom collapsible formatter
func CollapsibleFormatter(summaryTemplate string, detailFunc func(any) any, opts ...CollapsibleOption) func(any) any
```

**Usage Example**:
```go
schema := output.WithSchema(
    output.Field{
        Name: "errors",
        Type: "array",
        Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
    },
    output.Field{
        Name: "path",
        Type: "string", 
        Formatter: output.FilePathFormatter(30),
    },
    output.Field{
        Name: "config",
        Type: "object",
        Formatter: output.JSONFormatter(100),
    },
)
```

#### CollapsibleSection Interface

Interface for section-level collapsible content:

```go
type CollapsibleSection interface {
    Title() string                               // Section title/summary
    Content() []Content                          // Nested content items
    IsExpanded() bool                           // Default expansion state
    Level() int                                 // Nesting level (0-3)
    FormatHint(format string) map[string]any    // Format-specific hints
}
```

**Usage**: Create collapsible sections containing entire tables or content blocks.

#### DefaultCollapsibleSection

Standard implementation for collapsible sections:

```go
type DefaultCollapsibleSection struct {
    // Internal fields (not exported)
}

// NewCollapsibleSection creates a collapsible section
func NewCollapsibleSection(title string, content []Content, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection

// Helper constructors
func NewCollapsibleTable(title string, table *TableContent, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection
func NewCollapsibleReport(title string, content []Content, opts ...CollapsibleSectionOption) *DefaultCollapsibleSection

// Configuration options
func WithSectionExpanded(expanded bool) CollapsibleSectionOption
func WithSectionLevel(level int) CollapsibleSectionOption
func WithSectionFormatHint(format string, hints map[string]any) CollapsibleSectionOption
```

**Example**:
```go
// Create collapsible table section
analysisTable := output.NewTableContent("Analysis Results", data)
section := output.NewCollapsibleTable(
    "Detailed Code Analysis",
    analysisTable,
    output.WithSectionExpanded(false),
)

// Create multi-content section
reportSection := output.NewCollapsibleReport(
    "Performance Report",
    []output.Content{
        output.NewTextContent("Analysis complete"),
        analysisTable,
        output.NewTextContent("All systems operational"),
    },
    output.WithSectionExpanded(true),
)
```

#### Cross-Format Rendering

Collapsible content adapts automatically to each output format:

| Format   | CollapsibleValue Rendering | CollapsibleSection Rendering |
|----------|----------------------------|------------------------------|
| Markdown | `<details><summary>` HTML elements | Nested `<details>` structure |
| JSON     | `{"type": "collapsible", "summary": "...", "details": [...]}` | Structured data with content array |
| YAML     | YAML mapping with summary/details fields | YAML structure with nested content |
| HTML     | Semantic `<details>` with CSS classes | Section elements with collapsible behavior |
| Table    | Summary + expansion indicator | Section headers with indented content |
| CSV      | Summary + automatic detail columns | Metadata comments with table data |

#### Renderer Configuration

Control collapsible behavior globally per renderer:

```go
type CollapsibleConfig struct {
    GlobalExpansion      bool              // Override all IsExpanded() settings
    MaxDetailLength      int               // Character limit for details (default: 500)
    TruncateIndicator    string            // Truncation suffix (default: "[...truncated]")
    TableHiddenIndicator string            // Table collapse indicator
    HTMLCSSClasses       map[string]string // Custom CSS classes for HTML
}

// Apply configuration to renderers
tableOutput := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion:      false,
        TableHiddenIndicator: "[click to expand]",
        MaxDetailLength:      200,
    }),
)

htmlOutput := output.NewOutput(
    output.WithFormat(output.HTML),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        HTMLCSSClasses: map[string]string{
            "details": "my-collapsible",
            "summary": "my-summary",
            "content": "my-details",
        },
    }),
)
```

#### Complete Example

```go
package main

import (
    "context"
    output "github.com/ArjenSchwarz/go-output/v2"
)

func main() {
    // Data with complex nested information
    analysisData := []map[string]any{
        {
            "file": "/very/long/path/to/project/src/components/UserProfile.tsx",
            "errors": []string{
                "Missing import for React",
                "Unused variable 'userData'",
                "Type annotation missing for 'props'",
            },
            "config": map[string]any{
                "eslint": true,
                "typescript": true,
                "prettier": false,
                "rules": []string{"no-unused-vars", "explicit-return-type"},
            },
        },
    }
    
    // Create table with collapsible formatters
    table := output.NewTableContent("Code Analysis", analysisData,
        output.WithSchema(
            output.Field{
                Name: "file",
                Type: "string",
                Formatter: output.FilePathFormatter(25), // Shorten long paths
            },
            output.Field{
                Name: "errors", 
                Type: "array",
                Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
            },
            output.Field{
                Name: "config",
                Type: "object",
                Formatter: output.JSONFormatter(50, output.WithExpanded(false)),
            },
        ))
    
    // Wrap in collapsible section
    section := output.NewCollapsibleTable(
        "Detailed Analysis Results",
        table,
        output.WithSectionExpanded(false),
    )
    
    // Build document
    doc := output.New().
        Header("Project Analysis Report").
        Text("Analysis completed successfully. Click sections to expand details.").
        Add(section).
        Build()
    
    // Render with custom configuration
    out := output.NewOutput(
        output.WithFormats(output.Markdown, output.JSON, output.Table),
        output.WithWriter(output.NewStdoutWriter()),
        output.WithCollapsibleConfig(output.CollapsibleConfig{
            TableHiddenIndicator: "[expand for details]",
            MaxDetailLength:      100,
        }),
    )
    
    if err := out.Render(context.Background(), doc); err != nil {
        panic(err)
    }
}
```

### Output System

#### Output

Manages rendering and writing:

```go
type Output struct {
    // Internal fields (not exported)
}

// NewOutput creates a new Output instance
func NewOutput(opts ...OutputOption) *Output

// Render processes a document through all configured components
func (o *Output) Render(ctx context.Context, doc *Document) error

// RenderTo is a convenience method using background context
func (o *Output) RenderTo(doc *Document) error

// Close cleans up resources
func (o *Output) Close() error
```

#### Output Options

Configuration options for Output:

```go
// Format options
func WithFormat(format Format) OutputOption
func WithFormats(formats ...Format) OutputOption

// Writer options
func WithWriter(writer Writer) OutputOption
func WithWriters(writers ...Writer) OutputOption

// Transformer options
func WithTransformer(transformer Transformer) OutputOption
func WithTransformers(transformers ...Transformer) OutputOption

// Progress options
func WithProgress(progress Progress) OutputOption

// v1 compatibility options
func WithTableStyle(style string) OutputOption
func WithTOC(enabled bool) OutputOption
func WithFrontMatter(fm map[string]string) OutputOption
func WithMetadata(key string, value any) OutputOption
```

### Format System

#### Format

Represents an output format:

```go
type Format struct {
    Name     string           // Format name
    Renderer Renderer         // Renderer implementation
    Options  map[string]any   // Format-specific options
}
```

#### Built-in Formats

Pre-configured formats:

```go
var (
    JSON     Format  // JSON output
    YAML     Format  // YAML output
    CSV      Format  // CSV output
    HTML     Format  // HTML output
    Table    Format  // Table output
    Markdown Format  // Markdown output
    DOT      Format  // Graphviz DOT output
    Mermaid  Format  // Mermaid diagram output
    DrawIO   Format  // Draw.io CSV output
)

// Table style variants
var (
    TableDefault       Format
    TableBold          Format
    TableColoredBright Format
    TableLight         Format
    TableRounded       Format
)
```

#### Format Constructors

```go
// TableWithStyle creates a table format with specified style
func TableWithStyle(styleName string) Format

// MarkdownWithToC creates markdown with table of contents
func MarkdownWithToC(enabled bool) Format

// MarkdownWithFrontMatter creates markdown with front matter
func MarkdownWithFrontMatter(frontMatter map[string]string) Format

// MarkdownWithOptions creates markdown with ToC and front matter
func MarkdownWithOptions(includeToC bool, frontMatter map[string]string) Format
```

### Renderer Interface

Custom renderers implement this interface:

```go
type Renderer interface {
    // Format returns the output format name
    Format() string

    // Render converts the document to bytes
    Render(ctx context.Context, doc *Document) ([]byte, error)

    // RenderTo streams output to a writer
    RenderTo(ctx context.Context, doc *Document, w io.Writer) error

    // SupportsStreaming indicates if streaming is supported
    SupportsStreaming() bool
}
```

### Writer System

#### Writer Interface

```go
type Writer interface {
    // Write outputs rendered data
    Write(ctx context.Context, format string, data []byte) error
}
```

#### Built-in Writers

Pre-implemented writers:

```go
// StdoutWriter writes to standard output
func NewStdoutWriter() Writer

// FileWriter writes to files with pattern support
func NewFileWriter(rootDir, pattern string) (Writer, error)

// S3Writer writes to AWS S3
func NewS3Writer(region, bucket, keyPattern string) (Writer, error)

// MultiWriter writes to multiple destinations
func NewMultiWriter(writers ...Writer) Writer
```

**File Pattern Examples**:
- `"report.{format}"` → `report.json`, `report.csv`
- `"output/{format}/data.{ext}"` → `output/json/data.json`

### Transformer System

#### Transformer Interface

```go
type Transformer interface {
    // Name returns the transformer name
    Name() string

    // Transform modifies the input bytes
    Transform(ctx context.Context, input []byte, format string) ([]byte, error)

    // CanTransform checks if this transformer applies
    CanTransform(format string) bool

    // Priority determines transform order (lower = earlier)
    Priority() int
}
```

#### Built-in Transformers

Pre-implemented transformers with two usage patterns:

**Direct Struct Instantiation** (for simple transformers):
```go
// Basic emoji conversion - no constructor needed
&EmojiTransformer{}

// Remove color codes - no constructor needed  
&RemoveColorsTransformer{}
```

**Constructor Functions** (for configurable transformers):
```go
// Color transformers
func NewColorTransformer() *ColorTransformer
func NewColorTransformerWithScheme(scheme ColorScheme) *ColorTransformer

// Sorting transformers
func NewSortTransformer(key string, ascending bool) *SortTransformer
func NewSortTransformerAscending(key string) *SortTransformer

// Line splitting transformers
func NewLineSplitTransformer(separator string) *LineSplitTransformer
func NewLineSplitTransformerDefault() *LineSplitTransformer

// Enhanced transformers with format awareness
func NewEnhancedEmojiTransformer() *EnhancedEmojiTransformer
func NewEnhancedColorTransformer() *EnhancedColorTransformer
func NewEnhancedSortTransformer(key string, ascending bool) *EnhancedSortTransformer

// Format-aware wrapper for existing transformers
func NewFormatAwareTransformer(transformer Transformer) *FormatAwareTransformer

// Transform pipeline for multiple transformers
func NewTransformPipeline() *TransformPipeline
```

**Struct Instantiation with Configuration** (alternative to constructors):
```go
// Configure transformers directly
&SortTransformer{Key: "Name", Ascending: true}
&LineSplitTransformer{Column: "Description", Separator: ","}
&ColorTransformer{Scheme: ColorScheme{Success: "green", Error: "red"}}
```

**ColorScheme Structure**:
```go
type ColorScheme struct {
    Success string // Color for positive/success values
    Warning string // Color for warning values
    Error   string // Color for error/failure values
    Info    string // Color for informational values
}
```

### Progress System

#### Progress Interface

```go
type Progress interface {
    // Core progress methods
    SetTotal(total int)
    SetCurrent(current int)
    Increment(delta int)
    SetStatus(status string)
    Complete()
    Fail(err error)

    // v1 compatibility methods
    SetColor(color ProgressColor)
    IsActive() bool
    SetContext(ctx context.Context)

    // v2 enhancements
    Close() error
}
```

#### Progress Types

```go
// ProgressColor for visual feedback
type ProgressColor int

const (
    ProgressColorDefault ProgressColor = iota
    ProgressColorGreen   // Success state
    ProgressColorRed     // Error state
    ProgressColorYellow  // Warning state
    ProgressColorBlue    // Informational state
)
```

#### Progress Constructors

```go
// NewProgress creates format-aware progress
func NewProgress(format string, opts ...ProgressOption) Progress

// NewProgressForFormats creates progress for multiple formats
func NewProgressForFormats(formats []Format, opts ...ProgressOption) Progress

// NewNoOpProgress creates a no-operation progress indicator
func NewNoOpProgress() Progress
```

### Error Handling

#### Error Types

Structured error types for different failure modes:

```go
// RenderError indicates rendering failure
type RenderError struct {
    Format   string
    Renderer string
    Content  string
    Cause    error
}

// ValidationError indicates invalid input
type ValidationError struct {
    Field   string
    Value   any
    Message string
}

// TransformError indicates transformation failure
type TransformError struct {
    Transformer string
    Format      string
    Input       []byte
    Cause       error
}

// WriterError indicates writer failure
type WriterError struct {
    Writer    string
    Format    string
    Operation string
    Cause     error
}
```

All error types implement `error` and can be unwrapped using `errors.Unwrap()`.

### Utility Functions

#### Data Types

```go
// Record represents a table row
type Record map[string]any

// GenerateID creates unique content identifiers
func GenerateID() string
```

#### Helper Functions

```go
// Helper functions for testing and validation
func ValidateNonNil(name string, value any) error
func ValidateSliceNonEmpty(name string, slice any) error
func FailFast(validators ...error) error
```

## Common Usage Patterns

### Basic Table Output

```go
doc := output.New().
    Table("Results", []map[string]any{
        {"Name": "Alice", "Score": 95},
        {"Name": "Bob", "Score": 87},
    }, output.WithKeys("Name", "Score")).
    Build()

out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithWriter(output.NewStdoutWriter()),
)

err := out.Render(context.Background(), doc)
```

### Multiple Formats and Destinations

```go
doc := output.New().
    Table("Data", data, output.WithKeys("ID", "Name", "Status")).
    Text("Report generated on: " + time.Now().Format(time.RFC3339)).
    Build()

fileWriter, _ := output.NewFileWriter("./output", "report.{format}")

out := output.NewOutput(
    output.WithFormats(output.JSON, output.CSV, output.HTML),
    output.WithWriter(output.NewStdoutWriter()),
    output.WithWriter(fileWriter),
)

err := out.Render(context.Background(), doc)
```

### Mixed Content Document

```go
doc := output.New().
    Header("System Report").
    Section("User Statistics", func(b *output.Builder) {
        b.Table("Active Users", activeUsers, output.WithKeys("Name", "LastLogin"))
        b.Table("User Roles", roles, output.WithKeys("Role", "Count"))
    }).
    Section("System Health", func(b *output.Builder) {
        b.Text("All systems operational").
        b.Table("Metrics", metrics, output.WithKeys("Metric", "Value", "Status"))
    }).
    Build()
```

### Transformer Usage Patterns

```go
// Pattern 1: Direct struct instantiation (simple transformers)
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithTransformer(&output.EmojiTransformer{}),
    output.WithTransformer(&output.RemoveColorsTransformer{}),
    output.WithWriter(output.NewStdoutWriter()),
)

// Pattern 2: Constructor functions (configurable transformers)
colorTransformer := output.NewColorTransformerWithScheme(output.ColorScheme{
    Success: "green",
    Info:    "blue", 
    Warning: "yellow",
    Error:   "red",
})
sortTransformer := output.NewSortTransformer("Name", true)

out = output.NewOutput(
    output.WithFormat(output.Table),
    output.WithTransformers(colorTransformer, sortTransformer),
    output.WithWriter(output.NewStdoutWriter()),
)

// Pattern 3: Struct instantiation with configuration
out = output.NewOutput(
    output.WithFormat(output.Table),
    output.WithTransformer(&output.SortTransformer{Key: "Name", Ascending: true}),
    output.WithTransformer(&output.LineSplitTransformer{Column: "Description", Separator: ","}),
    output.WithWriter(output.NewStdoutWriter()),
)
```

### Progress Tracking

```go
progress := output.NewProgress(output.FormatTable,
    output.WithProgressColor(output.ProgressColorGreen),
    output.WithProgressStatus("Processing data"),
)

out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithWriter(output.NewStdoutWriter()),
    output.WithProgress(progress),
)
```

### Graph and Chart Generation

```go
// Flow chart
doc := output.New().
    Graph("Process Flow", []output.Edge{
        {From: "Start", To: "Process", Label: "begin"},
        {From: "Process", To: "End", Label: "complete"},
    }).
    Build()

// Gantt chart
tasks := []output.GanttTask{
    {ID: "task1", Title: "Design", StartDate: "2024-01-01", EndDate: "2024-01-15"},
    {ID: "task2", Title: "Development", StartDate: "2024-01-16", EndDate: "2024-02-15"},
}

doc = output.New().
    GanttChart("Project Timeline", tasks).
    Build()

out := output.NewOutput(
    output.WithFormat(output.Mermaid),
    output.WithWriter(output.NewStdoutWriter()),
)
```

### Collapsible Content Patterns

#### Simple Field Collapsible Content

```go
// Data with error arrays and long paths
data := []map[string]any{
    {
        "file": "/very/long/path/to/project/src/components/UserDashboard.tsx",
        "errors": []string{"Import missing", "Unused variable", "Type error"},
        "warnings": []string{"Deprecated API", "Performance concern"},
    },
}

// Create table with collapsible formatters
doc := output.New().
    Table("Analysis Results", data, output.WithSchema(
        output.Field{
            Name: "file",
            Type: "string",
            Formatter: output.FilePathFormatter(25), // Shorten paths > 25 chars
        },
        output.Field{
            Name: "errors",
            Type: "array", 
            Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
        },
        output.Field{
            Name: "warnings",
            Type: "array",
            Formatter: output.ErrorListFormatter(output.WithExpanded(true)),
        },
    )).
    Build()

// Render to GitHub-compatible markdown
out := output.NewOutput(
    output.WithFormat(output.Markdown),
    output.WithWriter(output.NewStdoutWriter()),
)
```

#### Custom Collapsible Formatter

```go
// Custom formatter for configuration objects
func configFormatter(val any) any {
    if config, ok := val.(map[string]any); ok {
        configJSON, _ := json.MarshalIndent(config, "", "  ")
        if len(configJSON) > 100 {
            return output.NewCollapsibleValue(
                fmt.Sprintf("Config (%d keys)", len(config)),
                string(configJSON),
                output.WithExpanded(false),
                output.WithMaxLength(200),
            )
        }
    }
    return val
}

schema := output.WithSchema(
    output.Field{Name: "name", Type: "string"},
    output.Field{
        Name: "config",
        Type: "object",
        Formatter: configFormatter,
    },
)
```

#### Collapsible Sections for Report Organization

```go
// Create detailed analysis tables
usersTable := output.NewTableContent("User Analysis", userData)
performanceTable := output.NewTableContent("Performance Metrics", perfData)
securityTable := output.NewTableContent("Security Issues", securityData)

// Wrap tables in collapsible sections
userSection := output.NewCollapsibleTable(
    "User Activity Analysis",
    usersTable,
    output.WithSectionExpanded(true), // Expanded by default
)

performanceSection := output.NewCollapsibleTable(
    "Performance Analysis", 
    performanceTable,
    output.WithSectionExpanded(false), // Collapsed by default
)

// Multi-content section
securitySection := output.NewCollapsibleReport(
    "Security Report",
    []output.Content{
        output.NewTextContent("Security scan completed with 5 issues found"),
        securityTable,
        output.NewTextContent("Immediate action required for critical issues"),
    },
    output.WithSectionExpanded(false),
)

// Build comprehensive document
doc := output.New().
    Header("System Analysis Report").
    Text("Generated on " + time.Now().Format("2006-01-02 15:04:05")).
    Add(userSection).
    Add(performanceSection).
    Add(securitySection).
    Build()
```

#### Nested Collapsible Sections

```go
// Create nested hierarchy (max 3 levels)
subSection1 := output.NewCollapsibleTable(
    "Database Performance",
    dbTable,
    output.WithSectionLevel(2),
    output.WithSectionExpanded(false),
)

subSection2 := output.NewCollapsibleTable(
    "API Response Times",
    apiTable,
    output.WithSectionLevel(2),
    output.WithSectionExpanded(false),
)

mainSection := output.NewCollapsibleReport(
    "Infrastructure Analysis",
    []output.Content{
        output.NewTextContent("Infrastructure health check results"),
        subSection1,
        subSection2,
    },
    output.WithSectionLevel(1),
    output.WithSectionExpanded(true),
)

doc := output.New().Add(mainSection).Build()
```

#### Cross-Format Collapsible Rendering

```go
// Same data rendered in multiple formats with different behaviors
data := []map[string]any{
    {"errors": []string{"Error 1", "Error 2", "Error 3"}},
}

table := output.NewTableContent("Issues", data, output.WithSchema(
    output.Field{
        Name: "errors",
        Type: "array",
        Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
    },
))

doc := output.New().Add(table).Build()

// Markdown: GitHub <details> elements
markdownOut := output.NewOutput(
    output.WithFormat(output.Markdown),
    output.WithWriter(output.NewFileWriter(".", "report.md")),
)

// JSON: Structured collapsible data
jsonOut := output.NewOutput(
    output.WithFormat(output.JSON),
    output.WithWriter(output.NewFileWriter(".", "report.json")),
)

// Table: Terminal-friendly with expansion indicators
tableOut := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithWriter(output.NewStdoutWriter()),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        TableHiddenIndicator: "[expand to view all errors]",
    }),
)

// CSV: Automatic detail columns for spreadsheet analysis
csvOut := output.NewOutput(
    output.WithFormat(output.CSV),
    output.WithWriter(output.NewFileWriter(".", "report.csv")),
)

// Render all formats
ctx := context.Background()
markdownOut.Render(ctx, doc)  // Creates expandable <details> 
jsonOut.Render(ctx, doc)      // Creates {"type": "collapsible", ...}
tableOut.Render(ctx, doc)     // Shows: "3 errors [expand to view all errors]"
csvOut.Render(ctx, doc)       // Creates: errors, errors_details columns
```

#### Global Expansion Control

```go
// Development/debug mode: expand all content
debugOut := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion: true, // Override all IsExpanded() settings
    }),
    output.WithWriter(output.NewStdoutWriter()),
)

// Production mode: respect individual expansion settings
prodOut := output.NewOutput(
    output.WithFormat(output.Markdown),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion: false, // Use individual IsExpanded() values
        MaxDetailLength: 500,   // Limit detail length
        TruncateIndicator: "... (truncated)",
    }),
    output.WithWriter(output.NewStdoutWriter()),
)
```

#### Advanced Collapsible Configuration

```go
// Custom HTML output with branded styling
htmlOut := output.NewOutput(
    output.WithFormat(output.HTML),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        HTMLCSSClasses: map[string]string{
            "details": "company-collapsible",
            "summary": "company-summary",
            "content": "company-details",
        },
    }),
    output.WithWriter(output.NewFileWriter(".", "report.html")),
)

// Custom table output with branded indicators
tableOut := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        TableHiddenIndicator: "📋 Click to expand detailed information",
        MaxDetailLength:      150,
        TruncateIndicator:    "... [see full details in expanded view]",
    }),
    output.WithWriter(output.NewStdoutWriter()),
)
```

## Thread Safety

All v2 components are designed to be thread-safe:

- **Document**: Immutable after Build(), safe for concurrent reads
- **Builder**: Thread-safe during construction using mutexes
- **Output**: Thread-safe configuration and rendering
- **Content**: Immutable after creation

## Performance Considerations

- **Memory Efficiency**: Uses encoding.TextAppender and encoding.BinaryAppender interfaces
- **Concurrent Rendering**: Processes multiple formats in parallel
- **Streaming Support**: Large datasets can be streamed to avoid memory issues
- **Key Ordering**: No performance penalty - preserves user order without sorting

## Migration from v1

For migration guidance, see:
- `MIGRATION.md` - Complete migration guide
- `MIGRATION_EXAMPLES.md` - Before/after examples
- `MIGRATION_QUICK_REFERENCE.md` - Quick lookup reference

## Error Handling Best Practices

1. **Context Cancellation**: Always pass context for cancellable operations
2. **Error Wrapping**: Use structured error types for detailed debugging
3. **Early Validation**: Builder validates inputs during construction
4. **Resource Cleanup**: Call `Close()` on Output when done

## Extension Points

The v2 API is designed for extensibility:

- **Custom Renderers**: Implement `Renderer` interface for new formats
- **Custom Transformers**: Implement `Transformer` interface for data processing
- **Custom Writers**: Implement `Writer` interface for new destinations
- **Custom Content**: Implement `Content` interface for specialized content types

For more examples and advanced usage, see the `/examples` directory in the repository.