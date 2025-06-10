# Advanced Usage

This guide covers advanced patterns, complex scenarios, and best practices for using the go-output library effectively in production environments.

## Complex Data Transformations

### Working with Nested Data Structures

```go
package main

import (
    "github.com/ArjenSchwarz/go-output"
)

func processNestedData() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Service", "Region", "Instances", "Config"},
    }

    // Complex nested data
    serviceData := map[string]interface{}{
        "Service": "web-app",
        "Region":  "us-east-1",
        "Instances": []string{"i-123", "i-456", "i-789"},
        "Config": map[string]interface{}{
            "CPU":    "2 cores",
            "Memory": "4GB",
            "Storage": "20GB SSD",
        },
    }

    output.AddContents(serviceData)

    // The library automatically handles:
    // - Arrays as comma-separated strings (JSON) or newline-separated (table)
    // - Maps as JSON strings
    // - Complex types are serialized appropriately per format

    output.Settings.SetOutputFormat("table")
    output.Write()
}
```

### Data Aggregation and Grouping

```go
func aggregateData() {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Department", "Total Cost", "Instance Count", "Services"},
    }

    // Simulate aggregated data from multiple sources
    departments := map[string][]map[string]interface{}{
        "Engineering": {
            {"service": "api", "cost": 150.50, "instances": 3},
            {"service": "web", "cost": 89.20, "instances": 2},
        },
        "Marketing": {
            {"service": "analytics", "cost": 45.80, "instances": 1},
            {"service": "cms", "cost": 67.30, "instances": 1},
        },
    }

    for dept, services := range departments {
        var totalCost float64
        var totalInstances int
        var serviceNames []string

        for _, service := range services {
            totalCost += service["cost"].(float64)
            totalInstances += service["instances"].(int)
            serviceNames = append(serviceNames, service["service"].(string))
        }

        output.AddContents(map[string]interface{}{
            "Department":     dept,
            "Total Cost":     fmt.Sprintf("$%.2f", totalCost),
            "Instance Count": totalInstances,
            "Services":       serviceNames, // Will be formatted per output type
        })
    }

    output.Settings.SortKey = "Total Cost"
    output.Settings.SetOutputFormat("table")
    output.Write()
}
```

## Multi-Format Output Workflows

### Pipeline Processing

```go
func pipelineProcessing() {
    // Common data preparation
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Resource", "Type", "Status", "Cost", "Owner"},
    }

    // Add data once
    resources := []map[string]interface{}{
        {"Resource": "prod-db", "Type": "RDS", "Status": "running", "Cost": 245.67, "Owner": "team-backend"},
        {"Resource": "staging-app", "Type": "EC2", "Status": "stopped", "Cost": 0.00, "Owner": "team-frontend"},
        {"Resource": "backup-s3", "Type": "S3", "Status": "active", "Cost": 12.34, "Owner": "team-ops"},
    }

    for _, resource := range resources {
        output.AddContents(resource)
    }

    // Generate multiple output formats
    formats := []string{"table", "csv", "json", "html", "markdown"}

    for _, format := range formats {
        output.Settings.SetOutputFormat(format)
        output.Settings.OutputFile = fmt.Sprintf("resources.%s", format)
        output.Settings.Title = "Cloud Resources Report"
        output.Write()
    }

    // Generate diagram formats
    output.Settings.AddFromToColumns("Owner", "Resource")
    output.Settings.SetOutputFormat("mermaid")
    output.Settings.OutputFile = "resources.mmd"
    output.Write()
}
```

### Conditional Output Generation

```go
func conditionalOutput(resourceCount int, outputType string, verbose bool) {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
    }

    // Adjust output based on data size
    if resourceCount > 100 {
        output.Settings.SetOutputFormat("csv") // More efficient for large datasets
        output.Settings.SplitLines = false     // Keep data compact
    } else {
        output.Settings.SetOutputFormat("table")
        output.Settings.SplitLines = true // More readable for small datasets
    }

    // Conditional keys based on verbosity
    if verbose {
        output.Keys = []string{"ID", "Name", "Type", "Status", "Created", "Modified", "Owner", "Tags", "Cost"}
    } else {
        output.Keys = []string{"Name", "Type", "Status", "Cost"}
    }

    // Conditional styling
    if outputType == "presentation" {
        output.Settings.UseColors = true
        output.Settings.UseEmoji = true
        output.Settings.TableStyle = format.TableStyles["ColoredBright"]
    } else {
        output.Settings.UseColors = false
        output.Settings.UseEmoji = false
        output.Settings.TableStyle = format.TableStyles["Default"]
    }

    // Add data and generate output
    // ... (data population logic)
    output.Write()
}
```

## Error Handling and Validation

### Robust Error Handling

```go
func robustOutputGeneration(data []map[string]interface{}) error {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Name", "Value", "Status"},
    }

    // Validate data before processing
    if len(data) == 0 {
        return fmt.Errorf("no data provided for output generation")
    }

    // Validate required keys exist in data
    requiredKeys := map[string]bool{
        "Name":   true,
        "Value":  true,
        "Status": true,
    }

    for i, item := range data {
        for key := range requiredKeys {
            if _, exists := item[key]; !exists {
                return fmt.Errorf("required key '%s' missing from data item %d", key, i)
            }
        }
        output.AddContents(item)
    }

    // Validate output settings
    if output.Settings.OutputFormat == "mermaid" && output.Settings.FromToColumns == nil {
        return fmt.Errorf("mermaid format requires FromToColumns to be set")
    }

    if output.Settings.OutputFormat == "drawio" && !output.Settings.DrawIOHeader.IsSet() {
        return fmt.Errorf("drawio format requires DrawIOHeader to be configured")
    }

    // Safe output generation with recovery
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic during output generation: %v", r)
        }
    }()

    output.Write()
    return nil
}
```

### Data Validation and Sanitization

```go
func sanitizeAndValidateData(rawData []map[string]interface{}) []map[string]interface{} {
    sanitized := make([]map[string]interface{}, 0, len(rawData))

    for _, item := range rawData {
        cleanItem := make(map[string]interface{})

        for key, value := range item {
            // Handle nil values
            if value == nil {
                cleanItem[key] = ""
                continue
            }

            // Sanitize strings
            if str, ok := value.(string); ok {
                // Remove control characters
                cleaned := strings.Map(func(r rune) rune {
                    if r < 32 && r != '\t' && r != '\n' && r != '\r' {
                        return -1
                    }
                    return r
                }, str)
                cleanItem[key] = strings.TrimSpace(cleaned)
                continue
            }

            // Validate numeric values
            if num, ok := value.(float64); ok {
                if math.IsInf(num, 0) || math.IsNaN(num) {
                    cleanItem[key] = 0
                    continue
                }
            }

            cleanItem[key] = value
        }

        sanitized = append(sanitized, cleanItem)
    }

    return sanitized
}
```

## Performance Optimization

### Large Dataset Handling

```go
func handleLargeDataset(dataSource <-chan map[string]interface{}) {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"ID", "Timestamp", "Event", "Source"},
    }

    // Use CSV for large datasets (most efficient)
    output.Settings.SetOutputFormat("csv")
    output.Settings.SplitLines = false

    // Process data in batches to avoid memory issues
    batchSize := 1000
    currentBatch := 0

    for data := range dataSource {
        output.AddContents(data)
        currentBatch++

        // Write batch and reset buffer
        if currentBatch >= batchSize {
            output.Settings.OutputFile = fmt.Sprintf("batch_%d.csv", currentBatch/batchSize)
            output.Write()

            // Reset for next batch
            output.Contents = make([]format.OutputHolder, 0, batchSize)
            currentBatch = 0
        }
    }

    // Write remaining data
    if currentBatch > 0 {
        output.Settings.OutputFile = fmt.Sprintf("batch_final.csv")
        output.Write()
    }
}
```

### Memory-Efficient Streaming

```go
func streamingOutput(dataReader io.Reader) {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"Record", "Value"},
    }

    output.Settings.SetOutputFormat("csv")
    output.Settings.OutputFile = "streamed_output.csv"

    scanner := bufio.NewScanner(dataReader)
    recordCount := 0

    for scanner.Scan() {
        line := scanner.Text()

        // Process line into data structure
        record := map[string]interface{}{
            "Record": recordCount,
            "Value":  line,
        }

        output.AddContents(record)
        recordCount++

        // Write every 100 records to manage memory
        if recordCount%100 == 0 {
            output.AddToBuffer() // Add to internal buffer without writing to file
        }
    }

    // Final write
    output.Write()
}
```

## Integration Patterns

### Web API Integration

```go
type APIHandler struct {
    outputSettings *format.OutputSettings
}

func (h *APIHandler) GetResourcesHandler(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    outputFormat := r.URL.Query().Get("format")
    if outputFormat == "" {
        outputFormat = "json"
    }

    verbose := r.URL.Query().Get("verbose") == "true"

    // Create output array
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
    }

    // Configure based on request
    output.Settings.SetOutputFormat(outputFormat)
    output.Settings.UseColors = false // No colors for API responses

    if verbose {
        output.Keys = []string{"ID", "Name", "Type", "Status", "Created", "Owner", "Tags"}
    } else {
        output.Keys = []string{"ID", "Name", "Status"}
    }

    // Fetch and add data
    resources := fetchResources() // Your data source
    for _, resource := range resources {
        output.AddContents(resource)
    }

    // Set appropriate content type
    switch outputFormat {
    case "json":
        w.Header().Set("Content-Type", "application/json")
    case "csv":
        w.Header().Set("Content-Type", "text/csv")
        w.Header().Set("Content-Disposition", "attachment; filename=resources.csv")
    case "html":
        w.Header().Set("Content-Type", "text/html")
    default:
        w.Header().Set("Content-Type", "text/plain")
    }

    // Generate output directly to response
    if outputFormat == "json" {
        json.NewEncoder(w).Encode(output.GetContentsMapRaw())
    } else {
        // Use buffer for other formats
        output.AddToBuffer()
        w.Write(buffer.Bytes())
    }
}
```

### CLI Application Integration

```go
type CLIConfig struct {
    OutputFormat string
    OutputFile   string
    Verbose      bool
    UseColors    bool
    SortBy       string
}

func generateCLIOutput(config CLIConfig, data []map[string]interface{}) {
    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
    }

    // Configure from CLI flags
    output.Settings.SetOutputFormat(config.OutputFormat)
    output.Settings.OutputFile = config.OutputFile
    output.Settings.UseColors = config.UseColors && terminal.IsTerminal()
    output.Settings.UseEmoji = config.UseColors
    output.Settings.SortKey = config.SortBy

    // Adjust keys based on verbosity
    if config.Verbose {
        output.Keys = []string{"ID", "Name", "Type", "Status", "Created", "Modified", "Owner", "Tags", "Cost"}
    } else {
        output.Keys = []string{"Name", "Type", "Status"}
    }

    // Add data
    for _, item := range data {
        output.AddContents(item)
    }

    // Generate output
    output.Write()

    // Provide user feedback
    if config.OutputFile != "" {
        fmt.Printf("%s Output written to %s\n",
            output.Settings.StringSuccess("✓"),
            config.OutputFile)
    }
}
```

## Custom Extensions

### Custom Output Format

```go
func (output OutputArray) toCustomFormat() []byte {
    var buffer bytes.Buffer

    buffer.WriteString("CUSTOM FORMAT HEADER\n")
    buffer.WriteString("Generated: " + time.Now().Format(time.RFC3339) + "\n\n")

    for _, item := range output.Contents {
        buffer.WriteString("RECORD START\n")
        for _, key := range output.Keys {
            if value, exists := item.Contents[key]; exists {
                buffer.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
            }
        }
        buffer.WriteString("RECORD END\n\n")
    }

    return buffer.Bytes()
}

// Extend the Write method to support custom format
func (output OutputArray) WriteWithCustom() {
    if output.Settings.OutputFormat == "custom" {
        result := output.toCustomFormat()
        err := format.PrintByteSlice(result, output.Settings.OutputFile, output.Settings.S3Bucket)
        if err != nil {
            log.Fatal(err.Error())
        }
        return
    }

    // Fall back to standard Write method
    output.Write()
}
```

### Custom Styling Functions

```go
type CustomOutputSettings struct {
    *format.OutputSettings
    HighlightThreshold float64
    CriticalThreshold  float64
}

func (c *CustomOutputSettings) FormatMetric(value interface{}) string {
    if num, ok := value.(float64); ok {
        switch {
        case num >= c.CriticalThreshold:
            return c.StringFailure(fmt.Sprintf("%.2f", num))
        case num >= c.HighlightThreshold:
            return c.StringWarning(fmt.Sprintf("%.2f", num))
        default:
            return c.StringSuccess(fmt.Sprintf("%.2f", num))
        }
    }
    return c.toString(value)
}

func useCustomStyling() {
    customSettings := &CustomOutputSettings{
        OutputSettings:     format.NewOutputSettings(),
        HighlightThreshold: 100.0,
        CriticalThreshold:  500.0,
    }

    customSettings.UseColors = true
    customSettings.UseEmoji = true

    // Process data with custom formatting
    // ... (implementation details)
}
```

## Testing Strategies

### Unit Testing Output Generation

```go
func TestOutputGeneration(t *testing.T) {
    tests := []struct {
        name     string
        format   string
        data     []map[string]interface{}
        expected string
    }{
        {
            name:   "JSON output",
            format: "json",
            data: []map[string]interface{}{
                {"name": "test", "value": 123},
            },
            expected: `[{"name":"test","value":123}]`,
        },
        {
            name:   "CSV output",
            format: "csv",
            data: []map[string]interface{}{
                {"name": "test", "value": 123},
            },
            expected: "name,value\ntest,123\n",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            output := format.OutputArray{
                Settings: format.NewOutputSettings(),
                Keys:     []string{"name", "value"},
            }

            output.Settings.SetOutputFormat(tt.format)

            for _, item := range tt.data {
                output.AddContents(item)
            }

            // Capture output
            var result []byte
            switch tt.format {
            case "json":
                result = output.toJSON()
            case "csv":
                result = output.toCSV()
            }

            if string(result) != tt.expected {
                t.Errorf("Expected %s, got %s", tt.expected, string(result))
            }
        })
    }
}
```

### Integration Testing

```go
func TestFullWorkflow(t *testing.T) {
    tempDir := t.TempDir()

    output := format.OutputArray{
        Settings: format.NewOutputSettings(),
        Keys:     []string{"service", "status", "cost"},
    }

    output.Settings.SetOutputFormat("csv")
    output.Settings.OutputFile = filepath.Join(tempDir, "test_output.csv")

    // Add test data
    testData := []map[string]interface{}{
        {"service": "web", "status": "running", "cost": 150.50},
        {"service": "db", "status": "stopped", "cost": 0.00},
    }

    for _, item := range testData {
        output.AddContents(item)
    }

    // Generate output
    output.Write()

    // Verify file was created
    if _, err := os.Stat(output.Settings.OutputFile); os.IsNotExist(err) {
        t.Fatalf("Output file was not created")
    }

    // Verify file contents
    content, err := ioutil.ReadFile(output.Settings.OutputFile)
    if err != nil {
        t.Fatalf("Failed to read output file: %v", err)
    }

    expected := "service,status,cost\nweb,running,150.5\ndb,stopped,0\n"
    if string(content) != expected {
        t.Errorf("File content mismatch. Expected: %s, Got: %s", expected, string(content))
    }
}
```

## Best Practices Summary

### 1. Performance
- Use CSV format for large datasets
- Implement batching for very large data
- Reset buffers between operations
- Consider streaming for real-time data

### 2. Error Handling
- Validate data before processing
- Check format requirements
- Implement recovery mechanisms
- Provide meaningful error messages

### 3. Code Organization
- Separate data preparation from output generation
- Use factory functions for common configurations
- Implement configuration validation
- Create reusable components

### 4. Testing
- Test each output format separately
- Validate file outputs
- Test error conditions
- Use temporary directories for file tests

### 5. Maintenance
- Document custom extensions
- Version configuration schemas
- Monitor performance metrics
- Keep dependencies updated

This advanced guide provides the foundation for building robust, scalable applications using the go-output library in production environments.
