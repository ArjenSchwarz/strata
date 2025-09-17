# Future Ideas for Data Pipeline Feature

This document contains ideas that were considered during requirements gathering but are out of scope for the initial implementation.

## Advanced Filtering
- Custom filter expressions with complex logic
- OR logic for combining filters
- Regex-based filtering
- Filter by specific property values

## Custom Calculations
- User-defined risk scores
- Estimated apply time calculations
- Cost implications
- Custom field calculations through configuration
- Expression language for calculations

## Advanced Configuration
- Pipeline profiles (e.g., "security-review", "quick-summary")
- Custom operation chains in configuration
- User-defined sort orders
- Complex pipeline definitions

## Interactive Features
- On-the-fly modification of pipeline operations
- TUI mode with interactive filtering
- Live preview of filter results

## Performance Features
- Specific performance targets for different plan sizes
- Memory optimization for 1000+ resource plans
- Streaming for very large datasets

## CLI Enhancements
- Flexible filter syntax (--pipeline-filter="type:delete,risk:high")
- Individual flags for each operation (--filter-dangerous, --sort-by-risk)
- Complex CLI expressions

## Extended Compatibility
- Automatic migration tools from transformers
- Deprecation warnings with timeline
- Compatibility mode for gradual migration

## Error Handling Options
- Skip failed operations and continue
- Fallback strategies for failed operations
- User-selectable error recovery modes

## Future Resource Types
- Output changes filtering
- Drift detection filtering
- Cross-resource dependency analysis