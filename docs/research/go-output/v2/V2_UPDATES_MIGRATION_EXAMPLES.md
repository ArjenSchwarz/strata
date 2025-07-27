# V2 Collapsible Content Migration Examples

This document provides step-by-step migration examples for **existing v2 users** who want to adopt the new collapsible content features. These examples show how to enhance your current v2 code with expandable/collapsible functionality.

## Table of Contents

- [Field Formatter Migration](#field-formatter-migration)
- [Built-in Collapsible Formatters](#built-in-collapsible-formatters)
- [Custom Collapsible Formatters](#custom-collapsible-formatters)
- [Collapsible Sections Migration](#collapsible-sections-migration)
- [Configuration Migration](#configuration-migration)
- [Performance Considerations](#performance-considerations)
- [Best Practices](#best-practices)

## Field Formatter Migration

### Step 1: Basic String Formatter to Collapsible

**Before (existing v2):**
```go
// Simple string formatter for error arrays
func errorFormatter(val any) any {
    if errors, ok := val.([]string); ok {
        return strings.Join(errors, ", ")
    }
    return val
}

schema := output.WithSchema(
    output.Field{
        Name: "errors",
        Type: "array",
        Formatter: errorFormatter,
    },
)
```

**After (with collapsible support):**
```go
// Enhanced formatter with collapsible support
func errorFormatter(val any) any {
    if errors, ok := val.([]string); ok && len(errors) > 0 {
        return output.NewCollapsibleValue(
            fmt.Sprintf("%d errors", len(errors)),  // Summary
            errors,                                  // Details
            output.WithExpanded(false),             // Collapsed by default
        )
    }
    return val
}

// Schema remains the same
schema := output.WithSchema(
    output.Field{
        Name: "errors",
        Type: "array",
        Formatter: errorFormatter,  // Same field structure
    },
)
```

**Migration Steps:**
1. ✅ **No breaking changes** - existing v2 code continues to work
2. ✅ **Same Field structure** - no schema changes required
3. ✅ **Enhanced return value** - formatter now returns CollapsibleValue
4. ✅ **Cross-format support** - automatically works in all output formats

**Result:**
- **Markdown**: Creates `<details><summary>3 errors</summary>...</details>`
- **JSON**: Creates `{"type": "collapsible", "summary": "3 errors", "details": [...]}`
- **Table**: Shows `3 errors [details hidden - use --expand for full view]`
- **CSV**: Creates automatic `errors_details` column

### Step 2: Path Formatter Enhancement

**Before (existing v2):**
```go
func pathFormatter(val any) any {
    path := fmt.Sprint(val)
    if len(path) > 50 {
        return "..." + path[len(path)-47:]  // Simple truncation
    }
    return path
}
```

**After (with collapsible support):**
```go
func pathFormatter(val any) any {
    path := fmt.Sprint(val)
    if len(path) > 50 {
        return output.NewCollapsibleValue(
            "..."+path[len(path)-47:],  // Shortened summary
            path,                       // Full path in details
            output.WithExpanded(false),
        )
    }
    return path  // Short paths unchanged
}
```

**Migration Benefit:** Users can now see full paths on demand instead of losing information to truncation.

## Built-in Collapsible Formatters

### Step 3: Replace Custom Formatters with Built-ins

**Before (custom implementation):**
```go
func myErrorListFormatter(val any) any {
    if errors, ok := val.([]string); ok {
        if len(errors) == 0 {
            return "No errors"
        }
        if len(errors) == 1 {
            return errors[0]
        }
        return fmt.Sprintf("%d errors: %s, ...", len(errors), errors[0])
    }
    return val
}

func myPathFormatter(val any) any {
    path := fmt.Sprint(val)
    if len(path) > 30 {
        parts := strings.Split(path, "/")
        if len(parts) > 3 {
            return fmt.Sprintf(".../%s/%s", parts[len(parts)-2], parts[len(parts)-1])
        }
    }
    return path
}

schema := output.WithSchema(
    output.Field{Name: "file", Type: "string", Formatter: myPathFormatter},
    output.Field{Name: "errors", Type: "array", Formatter: myErrorListFormatter},
)
```

**After (built-in formatters):**
```go
// Remove custom formatter functions - use built-ins instead
schema := output.WithSchema(
    output.Field{
        Name: "file", 
        Type: "string", 
        Formatter: output.FilePathFormatter(30),  // Built-in with same 30-char limit
    },
    output.Field{
        Name: "errors", 
        Type: "array", 
        Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
    },
)
```

**Migration Benefits:**
1. ✅ **Less code** - remove custom formatter implementations
2. ✅ **Better UX** - users can expand to see full details
3. ✅ **Consistent behavior** - standardized across your application
4. ✅ **Cross-format optimization** - built-ins handle all output formats properly

### Step 4: JSON/Config Formatter Migration

**Before (custom JSON formatter):**
```go
func configFormatter(val any) any {
    if config, ok := val.(map[string]any); ok {
        jsonBytes, err := json.Marshal(config)
        if err != nil {
            return fmt.Sprintf("Config (%d keys)", len(config))
        }
        if len(jsonBytes) > 100 {
            return fmt.Sprintf("Config (%d keys) [%s...]", len(config), string(jsonBytes[:50]))
        }
        return string(jsonBytes)
    }
    return val
}
```

**After (built-in with collapsible):**
```go
// Replace with built-in JSONFormatter
schema := output.WithSchema(
    output.Field{
        Name: "config",
        Type: "object",
        Formatter: output.JSONFormatter(100, output.WithExpanded(false)),
    },
)
```

**Migration Result:** Users can now see formatted JSON with proper indentation in the expandable details instead of truncated single-line output.

## Custom Collapsible Formatters

### Step 5: Advanced Custom Formatters

**Before (complex custom logic):**
```go
func performanceFormatter(val any) any {
    if metrics, ok := val.(map[string]any); ok {
        var issues []string
        if cpu, ok := metrics["cpu"].(float64); ok && cpu > 80 {
            issues = append(issues, fmt.Sprintf("High CPU: %.1f%%", cpu))
        }
        if memory, ok := metrics["memory"].(float64); ok && memory > 90 {
            issues = append(issues, fmt.Sprintf("High Memory: %.1f%%", memory))
        }
        
        if len(issues) > 0 {
            return fmt.Sprintf("%d performance issues", len(issues))
        }
        return "Performance OK"
    }
    return val
}
```

**After (collapsible with structured details):**
```go
func performanceFormatter(val any) any {
    if metrics, ok := val.(map[string]any); ok {
        var issues []string
        if cpu, ok := metrics["cpu"].(float64); ok && cpu > 80 {
            issues = append(issues, fmt.Sprintf("High CPU: %.1f%%", cpu))
        }
        if memory, ok := metrics["memory"].(float64); ok && memory > 90 {
            issues = append(issues, fmt.Sprintf("High Memory: %.1f%%", memory))
        }
        
        if len(issues) > 0 {
            return output.NewCollapsibleValue(
                fmt.Sprintf("%d performance issues", len(issues)),
                map[string]any{
                    "issues": issues,
                    "metrics": metrics,  // Full metrics in details
                    "timestamp": time.Now().Format(time.RFC3339),
                },
                output.WithExpanded(false),
            )
        }
        return "Performance OK"
    }
    return val
}
```

**Migration Benefits:**
- **Rich details**: Users can expand to see full metrics and context
- **Structured data**: Different formats render the details appropriately
- **Preserved summary**: Quick overview remains available

## Collapsible Sections Migration

### Step 6: Convert Multiple Tables to Collapsible Sections

**Before (multiple tables):**
```go
// Separate tables that might overwhelm users
doc := output.New().
    Header("System Analysis Report").
    Table("Active Users", activeUsers, output.WithKeys("Name", "LastLogin", "Activity")).
    Table("System Performance", perfMetrics, output.WithKeys("Component", "Status", "ResponseTime")).
    Table("Security Issues", securityIssues, output.WithKeys("Severity", "Issue", "Location")).
    Table("Database Stats", dbStats, output.WithKeys("Table", "RowCount", "Size")).
    Build()
```

**After (organized with collapsible sections):**
```go
// Group related tables in collapsible sections
userSection := output.NewCollapsibleTable(
    "User Activity Analysis",
    output.NewTableContent("Active Users", activeUsers, output.WithKeys("Name", "LastLogin", "Activity")),
    output.WithSectionExpanded(true),  // Important info - show by default
)

performanceSection := output.NewCollapsibleTable(
    "System Performance Metrics",
    output.NewTableContent("Performance", perfMetrics, output.WithKeys("Component", "Status", "ResponseTime")),
    output.WithSectionExpanded(false),  // Detailed info - collapsed by default
)

securitySection := output.NewCollapsibleTable(
    "Security Analysis",
    output.NewTableContent("Issues", securityIssues, output.WithKeys("Severity", "Issue", "Location")),
    output.WithSectionExpanded(true),   // Critical info - show by default
)

dbSection := output.NewCollapsibleTable(
    "Database Statistics",
    output.NewTableContent("Stats", dbStats, output.WithKeys("Table", "RowCount", "Size")),
    output.WithSectionExpanded(false),  // Supporting info - collapsed by default
)

doc := output.New().
    Header("System Analysis Report").
    Text("Click sections below to expand detailed information").
    Add(userSection).
    Add(performanceSection).
    Add(securitySection).
    Add(dbSection).
    Build()
```

**Migration Benefits:**
1. ✅ **Better organization** - related content grouped logically
2. ✅ **Progressive disclosure** - users can focus on what they need
3. ✅ **Improved UX** - less overwhelming, more navigable
4. ✅ **Format-appropriate** - sections adapt to each output format

### Step 7: Multi-Content Sections

**Before (mixed content):**
```go
doc := output.New().
    Header("Analysis Results").
    Text("Found 15 issues across 23 files").
    Table("Critical Issues", criticalIssues).
    Text("Immediate action required for critical issues").
    Table("Warnings", warnings).
    Text("Review warnings during next maintenance window").
    Build()
```

**After (organized sections):**
```go
criticalSection := output.NewCollapsibleReport(
    "Critical Issues (Immediate Action Required)",
    []output.Content{
        output.NewTextContent(fmt.Sprintf("Found %d critical issues requiring immediate attention", len(criticalIssues))),
        output.NewTableContent("Critical Issues", criticalIssues),
        output.NewTextContent("These issues may impact system stability and should be addressed immediately"),
    },
    output.WithSectionExpanded(true),  // Critical - show by default
)

warningsSection := output.NewCollapsibleReport(
    "Warnings (Review During Maintenance)",
    []output.Content{
        output.NewTextContent(fmt.Sprintf("Found %d warnings for review", len(warnings))),
        output.NewTableContent("Warnings", warnings),
        output.NewTextContent("Schedule review during next maintenance window"),
    },
    output.WithSectionExpanded(false),  // Non-critical - collapsed by default
)

doc := output.New().
    Header("Analysis Results").
    Text(fmt.Sprintf("Analysis complete: found %d total issues across %d files", len(criticalIssues)+len(warnings), fileCount)).
    Add(criticalSection).
    Add(warningsSection).
    Build()
```

## Configuration Migration

### Step 8: Add Collapsible Configuration

**Before (basic output configuration):**
```go
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithWriter(output.NewStdoutWriter()),
)
```

**After (with collapsible configuration):**
```go
out := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithWriter(output.NewStdoutWriter()),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion:      false,                              // Respect individual settings
        TableHiddenIndicator: "[expand for details]",             // Custom indicator
        MaxDetailLength:      200,                                // Limit detail length
        TruncateIndicator:    "... [see full details when expanded]",
    }),
)
```

**Migration Steps:**
1. ✅ **Add configuration** - no changes to existing renderer setup
2. ✅ **Customize behavior** - control expansion indicators and limits
3. ✅ **Per-renderer settings** - different configs for different formats

### Step 9: Multi-Format with Different Collapsible Behaviors

**Before (same configuration for all formats):**
```go
formats := []output.Format{output.Markdown, output.JSON, output.Table, output.CSV}
out := output.NewOutput(
    output.WithFormats(formats...),
    output.WithWriter(output.NewStdoutWriter()),
)
```

**After (format-specific collapsible configurations):**
```go
// Markdown for GitHub PR comments - expanded by default for reviewers
markdownOut := output.NewOutput(
    output.WithFormat(output.Markdown),
    output.WithWriter(output.NewFileWriter(".", "report.md")),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion: false,  // Use individual expansion settings
    }),
)

// Table for terminal - collapsed by default for clean overview
tableOut := output.NewOutput(
    output.WithFormat(output.Table),
    output.WithWriter(output.NewStdoutWriter()),
    output.WithCollapsibleConfig(output.CollapsibleConfig{
        GlobalExpansion:      false,
        TableHiddenIndicator: "📋 [click to expand]",
        MaxDetailLength:      100,  // Shorter for terminal
    }),
)

// JSON for API - include all data
jsonOut := output.NewOutput(
    output.WithFormat(output.JSON),
    output.WithWriter(output.NewFileWriter(".", "report.json")),
    // No special config - JSON includes all structured data
)

// CSV for analysis - automatic detail columns
csvOut := output.NewOutput(
    output.WithFormat(output.CSV),
    output.WithWriter(output.NewFileWriter(".", "report.csv")),
    // CSV automatically creates detail columns
)

// Render with different behaviors per format
ctx := context.Background()
markdownOut.Render(ctx, doc)  // GitHub-optimized
tableOut.Render(ctx, doc)     // Terminal-optimized  
jsonOut.Render(ctx, doc)      // API-optimized
csvOut.Render(ctx, doc)       // Spreadsheet-optimized
```

## Performance Considerations

### Step 10: Optimizing Large Datasets

**Before (potential performance issues with large data):**
```go
// Large error lists might create overwhelming output
func processLargeDataset(files []FileAnalysis) {
    data := make([]map[string]any, len(files))
    for i, file := range files {
        data[i] = map[string]any{
            "file":   file.Path,
            "errors": file.Errors,  // Could be 100+ errors per file
            "lines":  file.LineCount,
        }
    }
    
    doc := output.New().
        Table("Analysis", data, output.WithKeys("file", "errors", "lines")).
        Build()
}
```

**After (optimized with collapsible and limits):**
```go
func processLargeDataset(files []FileAnalysis) {
    data := make([]map[string]any, len(files))
    for i, file := range files {
        data[i] = map[string]any{
            "file":   file.Path,
            "errors": file.Errors,
            "lines":  file.LineCount,
        }
    }
    
    doc := output.New().
        Table("Analysis", data, output.WithSchema(
            output.Field{
                Name: "file",
                Type: "string",
                Formatter: output.FilePathFormatter(30),  // Shorten long paths
            },
            output.Field{
                Name: "errors",
                Type: "array",
                Formatter: output.ErrorListFormatter(
                    output.WithExpanded(false),
                    output.WithMaxLength(300),  // Limit detail length
                ),
            },
            output.Field{Name: "lines", Type: "number"},
        )).
        Build()
    
    // Configure for performance
    out := output.NewOutput(
        output.WithFormat(output.Table),
        output.WithWriter(output.NewStdoutWriter()),
        output.WithCollapsibleConfig(output.CollapsibleConfig{
            MaxDetailLength:      200,                    // Global limit
            TruncateIndicator:    "... [truncated - see CSV for full details]",
            TableHiddenIndicator: "[expand for top errors]",
        }),
    )
}
```

**Performance Benefits:**
1. ✅ **Faster rendering** - summary views load quickly
2. ✅ **Less memory** - details loaded on demand
3. ✅ **Better UX** - users aren't overwhelmed by large datasets
4. ✅ **Configurable limits** - prevent runaway detail expansion

## Best Practices

### Migration Checklist

**✅ Before Migration:**
1. **Identify repetitive data** - arrays, long strings, JSON objects
2. **Review user workflows** - what do users need to see immediately vs. on demand?
3. **Consider output formats** - how will collapsible content work in each format?
4. **Test with real data** - ensure summary views are meaningful

**✅ During Migration:**
1. **Start with built-in formatters** - ErrorListFormatter, FilePathFormatter, JSONFormatter
2. **Migrate one field at a time** - easier to test and validate
3. **Configure expansion defaults** - critical info expanded, details collapsed
4. **Test all output formats** - ensure consistent behavior

**✅ After Migration:**
1. **Gather user feedback** - are summaries helpful? Are details discoverable?
2. **Monitor performance** - large datasets should render faster
3. **Adjust configuration** - fine-tune indicators and limits based on usage
4. **Document for your team** - explain when and how to use collapsible features

### Common Patterns

**Pattern 1: Error/Warning Lists**
```go
output.Field{
    Name: "errors",
    Type: "array", 
    Formatter: output.ErrorListFormatter(output.WithExpanded(false)),
}
```

**Pattern 2: Long File Paths**
```go
output.Field{
    Name: "path",
    Type: "string",
    Formatter: output.FilePathFormatter(25),
}
```

**Pattern 3: Configuration Objects**
```go
output.Field{
    Name: "config",
    Type: "object",
    Formatter: output.JSONFormatter(100, output.WithExpanded(false)),
}
```

**Pattern 4: Report Sections**
```go
section := output.NewCollapsibleTable(
    "Detailed Analysis",
    tableContent,
    output.WithSectionExpanded(false),  // Collapsed by default
)
```

### Format-Specific Best Practices

**Markdown (GitHub PR Comments):**
- Use collapsible sections for detailed analysis results
- Expand critical issues by default
- Collapse supporting information

**JSON (API Responses):**
- Include both summary and details in structured format
- Use type indicators for programmatic processing
- Consider pagination for very large datasets

**Table (Terminal Output):**
- Use clear expansion indicators
- Limit detail length for readability
- Provide keyboard shortcuts in help text

**CSV (Spreadsheet Analysis):**
- Leverage automatic detail columns
- Ensure summaries are meaningful for filtering/sorting
- Include detail columns in documentation

## Troubleshooting

**Issue: Formatters not creating collapsible content**
```go
// ❌ Wrong: returning string instead of CollapsibleValue
func badFormatter(val any) any {
    return fmt.Sprintf("%d items", len(val.([]string)))
}

// ✅ Correct: return CollapsibleValue for expandable content
func goodFormatter(val any) any {
    if items, ok := val.([]string); ok && len(items) > 0 {
        return output.NewCollapsibleValue(
            fmt.Sprintf("%d items", len(items)),
            items,
            output.WithExpanded(false),
        )
    }
    return val
}
```

**Issue: Performance problems with large details**
```go
// ✅ Add character limits
formatter := output.ErrorListFormatter(
    output.WithMaxLength(500),  // Limit detail size
    output.WithExpanded(false),
)
```

**Issue: Inconsistent behavior across formats**
```go
// ✅ Test all formats during development
formats := []output.Format{output.Markdown, output.JSON, output.Table, output.CSV}
for _, format := range formats {
    out := output.NewOutput(output.WithFormat(format))
    err := out.Render(ctx, doc)
    // Verify behavior in each format
}
```

This migration guide provides a complete path for existing v2 users to adopt collapsible content features while maintaining compatibility with their current code.