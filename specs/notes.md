# Go-Output v2 Data Transformation Pipeline Analysis for Strata

## Current State Analysis

### How Strata Currently Uses go-output v2
1. **Output Formatting**: Uses go-output v2 for rendering plan summaries in multiple formats (table, JSON, HTML, Markdown)
2. **Collapsible Content**: Leverages collapsible sections for progressive disclosure
3. **Byte-Level Transformers**: Currently uses byte-level transformers for post-rendering modifications:
   - `ActionSortTransformer`: Sorts table rows by danger indicators and action priority
   - `EmojiTransformer`: Adds emoji to output
   - `ColorTransformer`: Adds color highlighting

### Pain Points with Current Approach

The `ActionSortTransformer` demonstrates several limitations of byte-level transformation:

1. **String Parsing Overhead**: Must parse rendered table output to identify rows
2. **Fragile Pattern Matching**: Relies on string patterns like "| action" and "| resource" 
3. **Format-Specific Logic**: Only works with certain table formats
4. **Re-rendering Issues**: Transforms already-rendered content, which can break formatting
5. **Limited Scope**: Can only sort what's already rendered, cannot filter or aggregate

## New Data Transformation Pipeline Benefits

The go-output v2 data transformation pipeline (introduced in v2.1.0+) provides:

### Core Capabilities
1. **Data-Level Operations**: Transform structured data before rendering
2. **Fluent API**: Chain operations with method chaining
3. **Format-Aware**: Operations can adapt based on target format
4. **Performance Optimized**: Operations are reordered for optimal execution
5. **Immutable**: Returns new documents without modifying originals

### Available Operations
- **Filter**: Filter records based on predicates
- **Sort**: Sort by one or more columns with custom comparators
- **Limit**: Restrict output to first N records
- **GroupBy**: Group records and apply aggregations
- **AddColumn**: Add calculated fields based on existing data

## Opportunities for Strata

### 1. Replace ActionSortTransformer
Instead of parsing rendered tables:
```go
// Current: Byte-level sorting after rendering
transformer := &ActionSortTransformer{}
output.WithTransformer(transformer)

// New: Data-level sorting before rendering
doc.Pipeline().
    SortWith(func(a, b Record) int {
        // Sort by danger, then action priority, then name
    }).
    Execute()
```

### 2. Advanced Filtering Capabilities
Enable users to filter plan output:
- Show only destructive changes
- Show only changes to specific resource types
- Show only changes above a risk threshold
- Hide no-op changes

### 3. Resource Aggregation
Group and summarize changes:
- Group by provider (AWS, Azure, etc.)
- Group by resource type
- Group by action type
- Show counts and statistics per group

### 4. Calculated Fields
Add computed information:
- Risk scores
- Estimated apply time
- Dependency impact analysis
- Cost implications

### 5. Progressive Disclosure Enhancement
Use pipeline to control what's shown initially:
- Filter to show critical changes first
- Limit initial display to top N most important changes
- Expand to show all changes on demand

### 6. Performance Improvements
- Eliminate string parsing overhead
- Reduce memory usage for large plans
- Enable streaming for very large datasets

## Implementation Considerations

### Backward Compatibility
- Keep existing transformer support for legacy behavior
- Provide migration path from byte transformers to data pipeline
- Maintain existing CLI interface

### Configuration Integration
- Add pipeline operations to strata.yaml configuration
- Support CLI flags for common operations (--filter-dangerous, --limit=10)
- Allow custom pipeline definitions

### Testing Strategy
- Ensure pipeline operations produce same results as current transformers
- Add benchmarks to measure performance improvements
- Test with large plan files to validate scalability

## Next Steps
1. Define requirements for data pipeline integration
2. Identify priority use cases
3. Design configuration schema for pipeline operations
4. Plan migration strategy for existing transformers