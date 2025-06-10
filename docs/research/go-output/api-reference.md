# API Reference

Complete reference documentation for all public functions, types, and methods in the go-output library.

## Package Structure

```
github.com/ArjenSchwarz/go-output
├── format          // Main package
├── drawio           // Draw.io integration
├── mermaid          // Mermaid diagram generation
└── templates        // HTML templates
```

## Core Types

### OutputHolder

Represents a single record in the output.

```go
type OutputHolder struct {
    Contents map[string]interface{}
}
```

**Fields:**
- `Contents`: Key-value pairs representing the data for this record

### OutputArray

Main container for all output data and configuration.

```go
type OutputArray struct {
    Settings *OutputSettings
    Contents []OutputHolder
    Keys     []string
}
```

**Fields:**
- `Settings`: Configuration for output generation
- `Contents`: Collection of data records
- `Keys`: Column headers/field names to include in output

### OutputSettings

Configuration options for output generation.

```go
type OutputSettings struct {
    HasTOC                bool
    DrawIOHeader          drawio.Header
    FrontMatter           map[string]string
    FromToColumns         *FromToColumns
    MermaidSettings       *mermaid.Settings
    OutputFile            string
    OutputFileFormat      string
    OutputFormat          string
    S3Bucket             S3Output
    SeparateTables       bool
    ShouldAppend         bool
    SplitLines           bool
    SortKey              string
    TableMaxColumnWidth  int
    TableStyle           table.Style
    Title                string
    UseColors            bool
    UseEmoji             bool
}
```

## Core Functions

### NewOutputSettings

Creates a new OutputSettings instance with default values.

```go
func NewOutputSettings() *OutputSettings
```

**Returns:**
- `*OutputSettings`: New settings instance with defaults

**Example:**
```go
settings := format.NewOutputSettings()
settings.UseColors = true
settings.SetOutputFormat("table")
```

## OutputArray Methods

### AddHolder

Adds an OutputHolder to the array.

```go
func (output *OutputArray) AddHolder(holder OutputHolder)
```

**Parameters:**
- `holder`: The OutputHolder to add

**Example:**
```go
holder := format.OutputHolder{
    Contents: map[string]interface{}{
        "name": "service1",
        "status": "running",
    },
}
output.AddHolder(holder)
```

### AddContents

Creates and adds an OutputHolder from a map.

```go
func (output *OutputArray) AddContents(contents map[string]interface{})
```

**Parameters:**
- `contents`: Key-value pairs to add as a new record

**Example:**
```go
output.AddContents(map[string]interface{}{
    "name":   "web-server",
    "status": "active",
    "port":   8080,
})
```

### Write

Generates and outputs the data in the configured format.

```go
func (output OutputArray) Write()
```

**Behavior:**
- Outputs to stdout if no OutputFile is specified
- Creates/appends to file if OutputFile is set
- Uploads to S3 if S3Bucket is configured
- Formats data according to OutputFormat setting

**Example:**
```go
output.Settings.SetOutputFormat("json")
output.Settings.OutputFile = "data.json"
output.Write()
```

### AddToBuffer

Adds the formatted output to an internal buffer without writing to file.

```go
func (output OutputArray) AddToBuffer()
```

**Use Cases:**
- Building multi-section outputs
- Combining multiple data sets
- Memory-efficient streaming

### GetContentsMap

Returns data as a slice of string maps.

```go
func (output OutputArray) GetContentsMap() []map[string]string
```

**Returns:**
- `[]map[string]string`: All data with values converted to strings

**Example:**
```go
stringMaps := output.GetContentsMap()
for _, row := range stringMaps {
    fmt.Printf("Name: %s, Status: %s\n", row["name"], row["status"])
}
```

### GetContentsMapRaw

Returns data as a slice of interface{} maps (preserving original types).

```go
func (output OutputArray) GetContentsMapRaw() []map[string]interface{}
```

**Returns:**
- `[]map[string]interface{}`: All data with original types preserved

### KeysAsInterface

Returns keys as a slice of interface{} for table headers.

```go
func (output OutputArray) KeysAsInterface() []interface{}
```

**Returns:**
- `[]interface{}`: Keys converted to interface{} slice

### ContentsAsInterfaces

Returns all record values as interface{} slices.

```go
func (output OutputArray) ContentsAsInterfaces() [][]interface{}
```

**Returns:**
- `[][]interface{}`: Each record as a slice of values

## OutputSettings Methods

### SetOutputFormat

Sets the output format (case-insensitive).

```go
func (settings *OutputSettings) SetOutputFormat(format string)
```

**Parameters:**
- `format`: Output format ("json", "csv", "table", "html", "markdown", "yaml", "mermaid", "drawio", "dot")

**Example:**
```go
settings.SetOutputFormat("TABLE") // Converted to "table"
```

### AddFromToColumns

Configures columns for relationship-based formats (mermaid, dot).

```go
func (settings *OutputSettings) AddFromToColumns(from string, to string)
```

**Parameters:**
- `from`: Source column name
- `to`: Target column name

**Example:**
```go
settings.AddFromToColumns("Parent", "Child")
```

### AddFromToColumnsWithLabel

Configures columns for relationships with labels.

```go
func (settings *OutputSettings) AddFromToColumnsWithLabel(from string, to string, label string)
```

**Parameters:**
- `from`: Source column name
- `to`: Target column name
- `label`: Column containing edge labels

### GetDefaultExtension

Returns the default file extension for the current output format.

```go
func (settings *OutputSettings) GetDefaultExtension() string
```

**Returns:**
- `string`: File extension (e.g., ".json", ".csv", ".html")

### NeedsFromToColumns

Checks if the current format requires relationship columns.

```go
func (settings *OutputSettings) NeedsFromToColumns() bool
```

**Returns:**
- `bool`: True if format requires FromToColumns

### GetSeparator

Returns the appropriate separator for multi-value fields.

```go
func (settings *OutputSettings) GetSeparator() string
```

**Returns:**
- `string`: Separator string (", " for most formats, "\n" for table/markdown)

## Color and Styling Methods

### StringFailure

Formats text as a failure message.

```go
func (settings *OutputSettings) StringFailure(text interface{}) string
```

**Parameters:**
- `text`: Text to format

**Returns:**
- `string`: Formatted text with red color and failure indicators

**Example:**
```go
fmt.Println(settings.StringFailure("Operation failed"))
// Output: 🚨 Operation failed 🚨 (with red color if enabled)
```

### StringWarning

Formats text as a warning message.

```go
func (settings *OutputSettings) StringWarning(text string) string
```

**Parameters:**
- `text`: Text to format

**Returns:**
- `string`: Formatted text with yellow/orange color and warning indicators

### StringWarningInline

Formats text as an inline warning (no extra spacing).

```go
func (settings *OutputSettings) StringWarningInline(text string) string
```

### StringSuccess

Formats text as a success message.

```go
func (settings *OutputSettings) StringSuccess(text interface{}) string
```

**Parameters:**
- `text`: Text to format

**Returns:**
- `string`: Formatted text with green color and success indicators

### StringPositive

Formats text positively.

```go
func (settings *OutputSettings) StringPositive(text string) string
```

### StringPositiveInline

Formats text positively (inline version).

```go
func (settings *OutputSettings) StringPositiveInline(text string) string
```

### StringInfo

Formats text as informational.

```go
func (settings *OutputSettings) StringInfo(text interface{}) string
```

**Parameters:**
- `text`: Text to format

**Returns:**
- `string`: Formatted text with info indicators

### StringBold

Formats text in bold.

```go
func (settings *OutputSettings) StringBold(text string) string
```

### StringBoldInline

Formats text in bold (inline version).

```go
func (settings *OutputSettings) StringBoldInline(text string) string
```

## Draw.io Package

### Header Type

```go
type Header struct {
    // Private fields - use methods to configure
}
```

### NewHeader

Creates a new Draw.io header.

```go
func NewHeader(label string, style string, ignore string) Header
```

**Parameters:**
- `label`: Node label template with placeholders
- `style`: Node style template with placeholders
- `ignore`: Comma-separated list of columns to ignore

**Returns:**
- `Header`: Configured header instance

### DefaultHeader

Creates a header with default settings.

```go
func DefaultHeader() Header
```

**Returns:**
- `Header`: Header with label="%Name%", style="%Image%", ignore="Image"

### Header Methods

#### SetLayout

```go
func (header *Header) SetLayout(layout string)
```

**Parameters:**
- `layout`: Layout algorithm (see Layout Constants)

#### SetSpacing

```go
func (header *Header) SetSpacing(nodespacing, levelspacing, edgespacing int)
```

**Parameters:**
- `nodespacing`: Space between nodes (default: 40)
- `levelspacing`: Space between hierarchy levels (default: 100)
- `edgespacing`: Space between parallel edges (default: 40)

#### GetSpacing

```go
func (header *Header) GetSpacing() (nodespacing, levelspacing, edgespacing int)
```

**Returns:**
- Current spacing values

#### SetHeightAndWidth

```go
func (header *Header) SetHeightAndWidth(height, width string)
```

**Parameters:**
- `height`: Node height ("auto", number in px, or "@ColumnName")
- `width`: Node width ("auto", number in px, or "@ColumnName")

#### SetIdentity

```go
func (header *Header) SetIdentity(columnname string)
```

**Parameters:**
- `columnname`: Column to use as unique identifier

#### SetParent

```go
func (header *Header) SetParent(parent, parentStyle string)
```

**Parameters:**
- `parent`: Column containing parent references
- `parentStyle`: CSS style for parent containers

#### AddConnection

```go
func (header *Header) AddConnection(connection Connection)
```

**Parameters:**
- `connection`: Connection configuration

#### IsSet

```go
func (header *Header) IsSet() bool
```

**Returns:**
- `bool`: True if header is properly configured

### Connection Type

```go
type Connection struct {
    From   string `json:"from"`
    To     string `json:"to"`
    Invert bool   `json:"invert"`
    Label  string `json:"label"`
    Style  string `json:"style"`
}
```

### NewConnection

Creates a new connection with defaults.

```go
func NewConnection() Connection
```

**Returns:**
- `Connection`: Default connection configuration

### Layout Constants

```go
const (
    LayoutAuto           = "auto"
    LayoutNone           = "none"
    LayoutHorizontalFlow = "horizontalflow"
    LayoutVerticalFlow   = "verticalflow"
    LayoutHorizontalTree = "horizontaltree"
    LayoutVerticalTree   = "verticaltree"
    LayoutOrganic        = "organic"
    LayoutCircle         = "circle"
)
```

### Style Constants

```go
const DefaultConnectionStyle = "curved=1;endArrow=blockThin;endFill=1;fontSize=11;"
const BidirectionalConnectionStyle = "curved=1;endArrow=blockThin;endFill=1;fontSize=11;startArrow=blockThin;startFill=1;"
const DefaultParentStyle = "swimlane;whiteSpace=wrap;html=1;childLayout=stackLayout;horizontal=1;horizontalStack=0;resizeParent=1;resizeLast=0;collapsible=1;"
```

### Draw.io Functions

#### CreateCSV

Creates a Draw.io compatible CSV file.

```go
func CreateCSV(drawIOHeader Header, headerRow []string, contents []map[string]string, filename string)
```

**Parameters:**
- `drawIOHeader`: Header configuration
- `headerRow`: Column names
- `contents`: Data rows
- `filename`: Output filename (empty for stdout)

#### GetContentsFromFileAsStringMaps

Reads Draw.io CSV and returns data as string maps.

```go
func GetContentsFromFileAsStringMaps(filename string) []map[string]string
```

**Parameters:**
- `filename`: CSV file to read

**Returns:**
- `[]map[string]string`: Parsed data rows

#### AWSShape

Returns AWS service shape for Draw.io.

```go
func AWSShape(group string, title string) string
```

**Parameters:**
- `group`: AWS service group (e.g., "Compute", "Storage")
- `title`: Service name (e.g., "EC2", "S3")

**Returns:**
- `string`: Shape definition for Draw.io

## Mermaid Package

### Settings Type

```go
type Settings struct {
    ChartType     string
    Direction     string
    GanttSettings *GanttSettings
    // Other chart-specific settings
}
```

### Chart Types

- `"flowchart"`: Flow diagrams
- `"piechart"`: Pie charts
- `"ganttchart"`: Gantt charts

## Utility Functions

### PrintByteSlice

Outputs byte data to file or stdout.

```go
func PrintByteSlice(contents []byte, outputFile string, targetBucket S3Output) error
```

**Parameters:**
- `contents`: Data to output
- `outputFile`: Target file path (empty for stdout)
- `targetBucket`: S3 configuration (optional)

**Returns:**
- `error`: Error if operation fails

## Type Definitions

### S3Output

```go
type S3Output struct {
    S3Client *s3.Client
    Bucket   string
    Path     string
}
```

### FromToColumns

```go
type FromToColumns struct {
    From  string
    To    string
    Label string
}
```

### TableStyles Map

```go
var TableStyles = map[string]table.Style{
    "Default":                    table.StyleDefault,
    "Bold":                       table.StyleBold,
    "ColoredBright":              table.StyleColoredBright,
    "ColoredDark":                table.StyleColoredDark,
    // ... additional styles
}
```

## Error Handling

Most methods that can fail return errors or use `log.Fatal()` for unrecoverable errors. Key error scenarios:

- Invalid output format
- Missing required configuration (e.g., FromToColumns for mermaid)
- File I/O errors
- S3 upload failures
- Data validation failures

## Thread Safety

The library is not thread-safe. If using from multiple goroutines:

- Create separate OutputArray instances per goroutine
- Use mutex protection when sharing OutputSettings
- Avoid concurrent modifications to the same OutputArray

## Memory Considerations

- Large datasets: Use CSV format and consider batching
- Buffer management: Call `AddToBuffer()` for complex outputs
- Reset buffers: Library manages internal buffers automatically
- S3 uploads: Data is loaded into memory before upload

This API reference covers all public interfaces. For implementation details and examples, see the other documentation files.
