# Data Pipeline Integration Design

## Overview

This feature replaces the current ActionSortTransformer's string-based sorting implementation with data-level sorting. The ActionSortTransformer currently parses rendered table strings using regex patterns to re-sort them, which is fragile and inefficient. By moving sorting to the data level before rendering, we eliminate string manipulation and improve maintainability.

## Architecture

### Current Architecture

```
Data Preparation → Table Rendering → String Transformation (ActionSortTransformer) → Output
                                      ↑
                                      Regex parsing, row extraction, re-sorting
```

### Proposed Architecture

```
Data Preparation → Sort Data → Apply Decorations → Table Rendering → Output
                   ↑           ↑
                   Sort raw     Add emoji/styling
                   data         after sorting
```

### Key Architectural Changes

1. **Sorting Location**: Move sorting from post-render transformer to pre-render data preparation
2. **Decoration Timing**: Apply emoji warnings and styling AFTER sorting for cleaner logic
3. **Data Structure**: Use existing `[]map[string]any` structure from `prepareResourceTableData`
4. **Integration Point**: Two simple functions in the data preparation flow

## Components and Interfaces

### 1. Simple Implementation (15 lines of code)

**Location**: `lib/plan/formatter.go`

```go
// sortResourceTableData sorts table data by danger, action priority, then alphabetically
func sortResourceTableData(tableData []map[string]any) {
    sort.SliceStable(tableData, func(i, j int) bool {
        a, b := tableData[i], tableData[j]

        // 1. Compare danger status (using IsDangerous flag)
        dangerA, _ := a["IsDangerous"].(bool)
        dangerB, _ := b["IsDangerous"].(bool)
        if dangerA != dangerB {
            return dangerA // dangerous items first
        }

        // 2. Compare raw action priority (before decoration)
        actionA, _ := a["ActionType"].(string)
        actionB, _ := b["ActionType"].(string)
        priorityA := getActionPriority(actionA)
        priorityB := getActionPriority(actionB)
        if priorityA != priorityB {
            return priorityA < priorityB
        }

        // 3. Alphabetical by resource address
        resourceA, _ := a["Resource"].(string)
        resourceB, _ := b["Resource"].(string)
        return resourceA < resourceB
    })
}

func getActionPriority(action string) int {
    switch action {
    case "Remove":  return 0
    case "Replace": return 1
    case "Modify":  return 2
    case "Add":     return 3
    default:        return 4
    }
}
```

### 2. Apply Decorations After Sorting

```go
// applyDecorations adds emoji and styling to sorted data
func applyDecorations(tableData []map[string]any) {
    for _, row := range tableData {
        actionType, _ := row["ActionType"].(string)
        isDangerous, _ := row["IsDangerous"].(bool)

        // Apply emoji decoration based on danger flag
        if isDangerous {
            row["Action"] = "⚠️ " + actionType
        } else {
            row["Action"] = actionType
        }

        // Remove internal fields used for sorting
        delete(row, "ActionType")
        delete(row, "IsDangerous")
    }
}
```

### 3. Modified prepareResourceTableData

```go
func (f *Formatter) prepareResourceTableData(changes []ResourceChange) []map[string]any {
    tableData := make([]map[string]any, 0, len(changes))

    for _, change := range changes {
        if change.ChangeType == ChangeTypeNoOp {
            continue
        }

        row := map[string]any{
            "ActionType":  getActionDisplay(change.ChangeType), // Raw action without emoji
            "IsDangerous": change.IsDangerous,                   // Flag for sorting
            "Resource":    change.Address,
            // ... other fields ...
        }
        tableData = append(tableData, row)
    }

    // Step 1: Sort the raw data
    sortResourceTableData(tableData)

    // Step 2: Apply decorations after sorting
    applyDecorations(tableData)

    return tableData
}
```

### 4. Complete Removal of ActionSortTransformer

Delete entirely:
- `ActionSortTransformer` struct and all methods (~200 lines)
- Regex pattern caching variables (~15 lines)
- Registration in output pipeline (~5 lines)

## Data Models

No changes to existing data structures. The solution uses temporary internal fields during sorting that are removed after decoration:

```go
// Temporary fields used during sorting (removed after decoration)
"ActionType":  string  // Raw action without emoji ("Remove", "Add", etc.)
"IsDangerous": bool    // Danger flag for sorting

// Final visible fields (unchanged from current implementation)
"Action":    string  // Decorated action ("⚠️ Remove")
"Resource":  string  // Resource address
// ... other existing fields ...
```

## Error Handling

Simple and robust - type assertions with zero values as defaults:

```go
// Safe type assertions return zero values on failure
dangerA, _ := a["IsDangerous"].(bool)  // false if missing
actionA, _ := a["ActionType"].(string)  // "" if missing
```

No panics, no complex error handling needed.

## Testing Strategy

### Simple Test Plan

1. **Add sorting test** - Test the sort function with a few examples
2. **Run existing tests** - All existing tests should pass unchanged
3. **Compare output** - Run sample files, verify output is identical

That's it. No complex test infrastructure needed.

## Implementation Steps

1. **Add two functions** (~30 lines total):
   - `sortResourceTableData()` - 15 lines
   - `applyDecorations()` - 10 lines
   - `getActionPriority()` - 5 lines

2. **Modify prepareResourceTableData** (2 lines):
   - Call sort function
   - Call decoration function

3. **Delete ActionSortTransformer** (~220 lines removed):
   - Delete struct and methods
   - Remove transformer registration
   - Delete regex patterns

4. **Test** (15 minutes):
   - Run `make test`
   - Run sample files
   - Verify output matches

Total time: ~30 minutes

## Why This Works

1. **Provider Grouping**: Since `prepareResourceTableData` is called for each provider group separately, sorting naturally happens within each group. No special handling needed.

2. **Format Handling**: go-output v2 already handles format-specific logic. We sort at the data level, output formats work unchanged.

3. **Clean Separation**: Sorting logic is separate from decoration logic - much cleaner and easier to understand.

4. **Minimal Changes**: ~30 lines added, ~220 lines removed. Net reduction in code complexity.

## Summary

This design replaces 200+ lines of fragile regex-based string parsing with 30 lines of clean data sorting. By separating sorting from decoration, the logic becomes trivial:

1. Sort raw data by danger, action, then name
2. Add emoji decorations after sorting
3. Delete the old transformer

The implementation is so simple it barely needs a design document.