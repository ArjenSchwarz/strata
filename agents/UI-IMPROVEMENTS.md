# UI/UX Improvements for Terraform Output Changes Display

## Summary
This document provides design recommendations for implementing an output changes section in the Strata Terraform plan summary tool. The analysis is based on the existing interface patterns, visual hierarchy, and user experience considerations within the current codebase.

## Critical Issues

### Issue: Missing Output Changes Display
**Current State**: The tool analyzes output changes (`OutputChange` struct exists) but does not display them in any output format.
**Problem**: Users cannot see planned changes to Terraform outputs, which are crucial for understanding infrastructure modifications that affect external integrations, API endpoints, and resource references.
**Recommendation**: Implement a dedicated "Output Changes" table section that appears after the Resource Changes section.
**Impact**: Provides complete visibility into all planned infrastructure changes, improving user confidence and decision-making.
**Implementation Notes**: The `OutputChange` struct already exists in `/Users/arjenschwarz/projects/personal/strata/lib/plan/models.go` with proper fields (Name, ChangeType, Sensitive, Before, After).

## High Priority Improvements

### Issue: Visual Hierarchy and Placement
**Current State**: Resource Changes section is the final content section before rendering.
**Problem**: Without proper positioning, output changes could appear disconnected from the overall plan summary.
**Recommendation**: 
- Place "Output Changes" section immediately after "Resource Changes" section
- Use the same visual styling pattern as existing tables (`output.NewTableContent`)
- Maintain consistent section spacing and typography hierarchy
**Impact**: Creates logical flow from infrastructure resources to their exposed outputs.
**Implementation Notes**: Add to `handleResourceDisplay` function in `/Users/arjenschwarz/projects/personal/strata/lib/plan/formatter.go` after line 1250.

### Issue: Table Format and Column Design
**Current State**: No output display exists.
**Problem**: Need to design appropriate column structure for different output types and states.
**Recommendation**: Use a simplified table structure with these columns:
- **NAME**: Output name (string, left-aligned)
- **ACTION**: Change type with visual indicators (Create→"Add", Update→"Modify", Delete→"Remove")
- **CURRENT**: Current value display with sensitivity handling
- **PLANNED**: New value display with "known after apply" support
- **SENSITIVE**: Clear indicator for sensitive outputs (⚠️ icon or text)

**Impact**: Provides clear, scannable information about output changes.
**Implementation Notes**: Follow the same pattern as `getResourceTableSchema()` but with simplified columns for outputs.

### Issue: Sensitive Value Display
**Current State**: Resource changes handle sensitive values with masking.
**Problem**: Outputs often contain sensitive information (API keys, passwords, connection strings) that require careful display handling.
**Recommendation**:
- Sensitive outputs: Display "(sensitive value)" for both current and planned values
- Include clear visual indicator (⚠️ symbol) in SENSITIVE column
- Consider auto-expansion of sensitive outputs when `AutoExpandDangerous` is enabled
- Non-sensitive outputs: Show actual values with proper formatting

**Impact**: Maintains security while providing necessary visibility into output changes.
**Implementation Notes**: Leverage existing `formatValue` function with sensitivity parameter.

### Issue: Change Type Visual Indicators
**Current State**: Resource changes use action-specific indicators and sorting.
**Problem**: Need consistent visual language for output change types.
**Recommendation**:
- **Create**: "Add" with + indicator or green highlighting
- **Update**: "Modify" with ~ indicator or yellow highlighting  
- **Delete**: "Remove" with - indicator or red highlighting
- Apply same action-based sorting as resources (Remove → Replace → Modify → Add)

**Impact**: Maintains visual consistency with resource changes, improving user comprehension.
**Implementation Notes**: Reuse `getActionDisplay()` function and extend `ActionSortTransformer` to handle output tables.

## Medium Priority Enhancements

### Issue: Unknown Value Handling
**Current State**: Resource changes handle computed values.
**Problem**: Terraform outputs often show "(known after apply)" which needs clear presentation.
**Recommendation**:
- Display "(known after apply)" in italic or subdued styling
- Consider special formatting for computed values vs. actual values
- Ensure consistent handling across all output formats (table, JSON, HTML, Markdown)

**Impact**: Clear differentiation between known and computed values reduces user confusion.
**Implementation Notes**: Add special case handling in value formatting functions.

### Issue: Empty State Handling
**Current State**: Resource changes show "All resources are unchanged" when no changes exist.
**Problem**: Need appropriate messaging when no output changes are present.
**Recommendation**:
- When outputs exist but have no changes: "All outputs unchanged"
- When no outputs exist: Suppress the section entirely (follow existing pattern)
- Maintain consistency with resource change empty state handling

**Impact**: Reduces visual clutter while providing clear status information.
**Implementation Notes**: Follow the same pattern as `prepareResourceTableData` filtering.

### Issue: Integration with Collapsible Sections
**Current State**: Resource changes support progressive disclosure.
**Problem**: Outputs are typically simpler than resources and may not need collapsible content initially.
**Recommendation**:
- Start with non-collapsible implementation for simplicity
- Consider future enhancement for collapsible output details if outputs become complex
- Ensure compatibility with global `--expand-all` flag

**Impact**: Maintains interface simplicity while preserving future extensibility.
**Implementation Notes**: Use basic table without collapsible formatters initially.

## Low Priority Suggestions

### Issue: Output Grouping
**Current State**: Resources support provider-based grouping.
**Problem**: Outputs don't have natural grouping like resources do by provider.
**Recommendation**: 
- No grouping for initial implementation
- Consider future grouping by module path if needed for large infrastructures
- Maintain consistency with resource grouping thresholds if implemented

**Impact**: Keeps initial implementation focused while allowing future enhancements.

### Issue: Cross-Format Consistency
**Current State**: Resource changes render consistently across table, JSON, HTML, Markdown.
**Problem**: Output changes need same level of cross-format support.
**Recommendation**:
- Ensure output table renders properly in all supported formats
- Test sensitive value masking across formats
- Verify "(known after apply)" formatting in each output type

**Impact**: Maintains professional output quality across all user preferences.

## Positive Observations

The existing codebase demonstrates several excellent patterns that should be preserved:

1. **Consistent Table Creation**: The `output.NewTableContent()` pattern provides reliable table rendering across formats
2. **Sensitive Value Handling**: Existing sensitive value masking in resource changes provides a proven template
3. **Action-Based Sorting**: The `ActionSortTransformer` creates intuitive ordering that should extend to outputs
4. **Configuration Integration**: The config system supports sensitivity settings that can be leveraged for outputs
5. **Error Handling**: Conservative error handling with logging ensures graceful degradation

## Implementation Priority

1. **Phase 1** (Critical): Basic output table with standard columns and sensitive value masking
2. **Phase 2** (High): Visual indicators, action sorting, and proper value formatting
3. **Phase 3** (Medium): Unknown value handling and empty state management
4. **Phase 4** (Low): Advanced features like grouping or collapsible sections if needed

This phased approach ensures essential functionality is delivered quickly while maintaining code quality and user experience standards established in the existing codebase.