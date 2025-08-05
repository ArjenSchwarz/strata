# Table Header Consistency Feature - Design Document

## Overview

This feature standardizes table header formatting across all Strata output tables to use consistent Title Case formatting. The current implementation has inconsistent header capitalization (ALL UPPERCASE, all lowercase, and mixed styles) across the four main table types, creating a fragmented user experience.

### Problem Statement

Strata currently displays four types of tables with inconsistent header formatting:
- **Summary Statistics**: "TOTAL CHANGES", "ADDED", "REMOVED" (ALL UPPERCASE)
- **Resource Changes**: "action", "resource", "type", "property_changes" (all lowercase) 
- **Output Changes**: "NAME", "ACTION", "CURRENT" (ALL UPPERCASE)
- **Sensitive Resource Changes**: "ACTION", "RESOURCE", "TYPE" (ALL UPPERCASE)

This inconsistency reduces readability and creates a unprofessional appearance across different output formats.

### Solution Approach

Based on Decision D003 ("Simple Implementation Approach"), we will directly modify field names in table schemas to use Title Case formatting without implementing complex consistency enforcement systems.

## Architecture

### Design Principles

1. **Direct Implementation**: Modify field names directly in schema definitions (Decision D003)
2. **Title Case Standardization**: All headers use Title Case format (Decision D001)
3. **Technical Accuracy**: Preserve correct capitalization for technical terms like "ID" (Decision D002)
4. **Comprehensive Coverage**: Apply changes to all table types including provider-grouped tables (Decision D004)
5. **Format Limitations**: Accept that table format headers remain ALL UPPERCASE (Decision D008)

### High-Level Architecture

```
Formatter Schemas → Field Name Changes → Header Display
      ↓                    ↓                   ↓
getResourceTableSchema() → "Action"        → Rendered Headers
getOutputTableSchema()   → "Property Changes" → (Title Case in
getStatsTableSchema()    → "High Risk"     →  supported formats)
```

### Component Overview

The implementation affects three main components in `/lib/plan/formatter.go`:

1. **Statistics Table Schema** (lines 276, 927, 954)
2. **Resource Changes Table Schema** (lines 1012-1048) 
3. **Output Changes Table Schema** (line 1374)
4. **Provider-Grouped Tables** (inherit from Resource Changes schema)

## Components and Interfaces

### Critical Implementation Note

**\u26a0\ufe0f This implementation requires both schema changes AND data preparation changes.** Simply changing field names in schemas without updating the corresponding data preparation functions will result in empty table columns.

### 1. Statistics Table Schema Changes

**Location**: `lib/plan/formatter.go` lines 276, 927, 954

**Current Implementation**:
```go
output.WithKeys("TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED")
```

**Required Changes**:
```go
output.WithKeys("Total Changes", "Added", "Removed", "Modified", "Replacements", "High Risk", "Unmodified")
```

**Impact**: This change affects the Summary Statistics table displayed in all plan summaries.

### 2. Resource Changes Table Schema Changes

**Location**: `lib/plan/formatter.go` function `getResourceTableSchema()` (lines 1012-1048)

**Current Implementation**:
```go
[]output.Field{
    {Name: "action", Type: "string"},
    {Name: "resource", Type: "string", Formatter: output.FilePathFormatter(50)},
    {Name: "type", Type: "string"},
    {Name: "id", Type: "string"},
    {Name: "replacement", Type: "string"},
    {Name: "module", Type: "string"},
    {Name: "danger", Type: "string"},
    {Name: "property_changes", Type: "object", Formatter: f.propertyChangesFormatterTerraform()},
}
```

**Required Changes**:
```go
[]output.Field{
    {Name: "Action", Type: "string"},
    {Name: "Resource", Type: "string", Formatter: output.FilePathFormatter(50)},
    {Name: "Type", Type: "string"},
    {Name: "ID", Type: "string"},           // Technical term preserved (Decision D002)
    {Name: "Replacement", Type: "string"},
    {Name: "Module", Type: "string"},
    {Name: "Danger", Type: "string"},
    {Name: "Property Changes", Type: "object", Formatter: f.propertyChangesFormatterTerraform()},
}
```

**Impact**: This change affects the main Resource Changes table and all provider-grouped tables that inherit this schema.

### 2a. Resource Changes Data Preparation Changes

**\u26a0\ufe0f CRITICAL**: The schema changes above require corresponding updates to data preparation functions.

**Location**: Data preparation functions that create table data maps

**Current Data Preparation**:
```go
tableData := map[string]any{
    "action":           change.ChangeType.String(),
    "resource":         change.Address,
    "type":             change.Type,
    "id":               change.PhysicalID,
    "replacement":      change.ReplacementType.String(),
    "module":           change.ModulePath,
    "danger":           dangerStatus,
    "property_changes": analysis.PropertyChanges,
}
```

**Required Data Preparation Changes**:
```go
tableData := map[string]any{
    "Action":           change.ChangeType.String(),
    "Resource":         change.Address,
    "Type":             change.Type,
    "ID":               change.PhysicalID,
    "Replacement":      change.ReplacementType.String(),
    "Module":           change.ModulePath,
    "Danger":           dangerStatus,
    "Property Changes": analysis.PropertyChanges,
}
```

**Functions Requiring Updates**:
- All functions that prepare resource table data
- Provider grouping functions that create similar data structures

### 3. Output Changes Table Schema Changes

**Location**: `lib/plan/formatter.go` line 1374

**Current Implementation**:
```go
output.WithKeys("NAME", "ACTION", "CURRENT", "PLANNED", "SENSITIVE")
```

**Required Changes**:
```go
output.WithKeys("Name", "Action", "Current", "Planned", "Sensitive")
```

**Impact**: This change affects the Output Changes table when Terraform outputs are modified.

### 4. Provider-Grouped Tables

**Inheritance**: Provider-grouped tables automatically inherit the Resource Changes schema through the `getResourceTableSchema()` function, so no separate changes are required (Decision D004).

## Data Models

### Header Mapping Rules

The following mapping rules apply for converting current headers to Title Case:

| Current Header | New Header | Rule Applied |
|----------------|------------|--------------|
| TOTAL CHANGES | Total Changes | Title Case with space preserved |
| ADDED | Added | Simple Title Case |
| REMOVED | Removed | Simple Title Case |
| MODIFIED | Modified | Simple Title Case |
| REPLACEMENTS | Replacements | Simple Title Case |
| HIGH RISK | High Risk | Title Case with space preserved |
| UNMODIFIED | Unmodified | Simple Title Case |
| action | Action | Simple Title Case |
| resource | Resource | Simple Title Case |
| type | Type | Simple Title Case |
| id | ID | Technical term preserved |
| replacement | Replacement | Simple Title Case |
| module | Module | Simple Title Case |
| danger | Danger | Simple Title Case |
| property_changes | Property Changes | Title Case with underscore → space |
| NAME | Name | Simple Title Case |
| ACTION | Action | Simple Title Case |
| CURRENT | Current | Simple Title Case |
| PLANNED | Planned | Simple Title Case |
| SENSITIVE | Sensitive | Simple Title Case |

### Format Compatibility

The changes affect headers differently based on output format:

| Format | Header Rendering | Supports Title Case |
|--------|------------------|-------------------|
| Table | ALL UPPERCASE (fixed) | No (Decision D008) |
| Markdown | As specified | Yes |
| HTML | As specified | Yes |
| JSON | As specified | Yes |
| CSV | As specified | Yes |

**Note**: Table format headers will always display in ALL UPPERCASE due to the table formatter implementation. This is acceptable per Decision D008.

## Error Handling

### Error Prevention Strategy

1. **Schema Validation**: No additional validation required as field names are static strings
2. **Test Coverage**: Existing tests will need comprehensive updates
3. **Data Preparation**: Functions that prepare table data must be updated to match new field names

### Potential Issues and Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Data Binding Failures | Medium | Update all data preparation functions to use new field names |
| Test Failures | Medium | Update test assertions systematically |
| User Confusion | Low | Changes improve consistency and readability |

## Testing Strategy

### Test Coverage Areas

1. **Unit Tests**:
   - `lib/plan/formatter_test.go`: Update assertions for new header names
   - Verify all table types display correct Title Case headers
   - Test provider-grouped tables inherit correct schema

2. **Integration Tests**:
   - Verify header consistency across output formats
   - Test with real Terraform plan files
   - Validate no functional regressions

3. **Format-Specific Tests**:
   - Markdown: Verify Title Case headers in table markdown
   - HTML: Verify Title Case headers in HTML tables
   - JSON: Verify Title Case keys in JSON output
   - CSV: Verify Title Case column headers

### Test Update Requirements

**Current Test Patterns to Update**:

```go
// OLD: Tests expecting uppercase headers
if row["TOTAL CHANGES"] != 10 {
    t.Errorf("Expected TOTAL CHANGES to be 10, got %v", row["TOTAL CHANGES"])
}

// NEW: Tests expecting Title Case headers  
if row["Total Changes"] != 10 {
    t.Errorf("Expected Total Changes to be 10, got %v", row["Total Changes"])
}
```

**Files Requiring Test Updates**:
- `lib/plan/formatter_test.go`
- `lib/plan/formatter_enhanced_test.go`
- `lib/plan/formatter_create_test.go`
- Any integration tests that verify table structure

## Implementation Plan

### Phase 1: Data Preparation Updates
1. **Identify Data Functions**: Find all functions that prepare table data
2. **Update Data Keys**: Change map keys to match new field names:
   ```go
   // OLD:
   "property_changes": analysis
   // NEW:
   "Property Changes": analysis
   ```
3. **Function Updates Required**:
   - `prepareResourceTableData()`
   - `createStatisticsSummaryDataV2()`
   - Any other data preparation functions

### Phase 2: Schema Changes
1. Update Statistics table headers in all three locations
2. Update Resource Changes table schema in `getResourceTableSchema()`
3. Update Output Changes table headers

### Phase 3: Test Updates
1. Update unit tests to expect Title Case headers
2. Update integration tests
3. Verify no functional regressions
4. Test with real Terraform plan files

### Phase 4: Documentation
1. Update any documentation referencing old header names
2. Update examples if needed

### Implementation Order

The changes should be implemented in this order to minimize test failures:

1. **formatter.go Changes**: Update all schema definitions first
2. **Test Updates**: Update all test assertions immediately after
3. **Documentation**: Update any relevant documentation

## Validation and Verification

### Success Criteria

1. ✅ All table headers use consistent Title Case formatting
2. ✅ Technical terms like "ID" maintain correct capitalization
3. ✅ No functional regressions in plan analysis or output generation
4. ✅ All automated tests pass with updated assertions
5. ✅ Headers display correctly across all supported output formats (except table format)
6. ✅ **Data Binding Integrity**: All table data correctly maps to new field names

### Manual Testing Checklist

- [ ] Summary Statistics table shows Title Case headers
- [ ] Resource Changes table shows Title Case headers
- [ ] Output Changes table shows Title Case headers
- [ ] Provider-grouped tables show Title Case headers
- [ ] Markdown output has Title Case table headers
- [ ] HTML output has Title Case table headers
- [ ] JSON output has Title Case keys
- [ ] CSV output has Title Case column headers
- [ ] Table format headers remain ALL UPPERCASE (expected behavior)

## Assumptions and Constraints

### Assumptions

1. **Go-output v2 Behavior**: Field names in schema directly become table headers (confirmed)
2. **Format Compatibility**: All claimed formats (Markdown, HTML, JSON, CSV) will correctly display Title Case headers (confirmed)
3. **No Display Name Support**: go-output v2 doesn't support separate display names (Decision D005)

### Constraints

1. **Table Format Headers**: Cannot be changed from ALL UPPERCASE (Decision D008)
2. **Technical Accuracy**: Must preserve correct capitalization for abbreviations (Decision D002)
3. **Data Binding**: All data preparation functions must be updated to match new field names

### Dependencies

- **go-output v2.1.0**: Current library version supports field-name-as-header approach
- **Existing Test Suite**: Must be updated to reflect new header names
- **No External Dependencies**: Changes are internal to Strata codebase

## Alternative Approaches

### Alternative 1: Contribute DisplayName Support to go-output v2

**Approach**: Contribute a `DisplayName` field to go-output v2 that allows separate API keys and display headers.

**Pros**:
- Cleaner separation of concerns
- Benefits entire go-output community
- Future-proof solution

**Cons**:
- Requires external library contribution
- Longer implementation timeline
- Dependency on upstream acceptance

### Alternative 2: Current Direct Approach (SELECTED)

**Approach**: Directly modify field names in schemas to achieve Title Case headers.

**Pros**:
- Simple and straightforward implementation
- Achieves consistency goals immediately
- No external dependencies
- Aligns with Decision D003 (Simple Implementation Approach)

**Cons**:
- Field names and display names are coupled
- Future consistency enforcement would require additional tooling

### Recommendation

**Selected Approach**: Alternative 2 (Current Direct Approach) aligns with the project decision to keep the implementation simple and direct.

## Future Considerations

### Potential Enhancements

While not required for this feature, future improvements could include:

1. **Header Customization**: User-configurable header styles
2. **Internationalization**: Support for different languages/locales
3. **Dynamic Headers**: Context-aware header formatting

### Maintenance Notes

- Future table additions should follow Title Case convention
- New contributors should be aware of the Title Case standard
- Documentation should specify header formatting expectations
- Consider automated consistency checks in CI/CD

---

This design implements a simple, direct approach to achieving header consistency across all Strata table outputs while preserving technical accuracy and following the project's preference for straightforward solutions.