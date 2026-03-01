# Bug Fix Report: Unknown Parent Propagation to Nested Properties

**Bug ID**: T-146  
**Date**: 2025-01-01  
**Severity**: Medium  
**Status**: Fixed  

## Problem Statement

When a parent property is marked as unknown in Terraform's `after_unknown` field (e.g., `"config": true`), the nested properties within that parent (e.g., `config.timeout`, `config.retry_count`) were not being identified as unknown values. Only the parent property itself was tracked in the `UnknownProperties` list.

### Expected Behavior
- When `after_unknown` contains `"config": true`, all nested properties under `config` should be treated as unknown
- Properties like `config.timeout`, `config.retry_count`, `config.nested.deep_value` should appear in the `UnknownProperties` list
- These nested properties should display `(known after apply)` in the output

### Actual Behavior
- Only the parent property `config` was identified as unknown
- Nested properties like `config.timeout` were not in the `UnknownProperties` list
- Nested properties appeared as regular changes instead of unknown values

## Root Cause Analysis

The issue was in the `compareObjects` function in `lib/plan/analyzer.go`. When recursively processing nested properties, the function only passed the specific child's `afterUnknown` value to the recursive call, not considering that a parent could be marked as entirely unknown with a boolean `true`.

**Specific Issues:**

1. **Incomplete Unknown Propagation**: When `afterUnknown` contained `{"config": true}`, the recursive call for `config.timeout` would receive `afterUnknownChild = true`, but the `isValueUnknown` function expected to navigate a structure, not just receive a boolean.

2. **Missing Nested Property Collection**: When a nested object was treated as a single change (due to `shouldTreatAsNestedObject` logic), individual nested properties weren't being added to the `UnknownProperties` list.

## Solution

### 1. Enhanced Unknown Propagation Logic

Modified the `compareObjects` function to properly propagate unknown status from parent to child properties:

```go
// Extract afterUnknown for the child property
var afterUnknownChild any
if afterUnknown != nil {
    if afterUnknownMap, ok := afterUnknown.(map[string]any); ok {
        afterUnknownChild = afterUnknownMap[key]
    } else if parentUnknown, ok := afterUnknown.(bool); ok && parentUnknown {
        // If parent is entirely unknown (boolean true), propagate to all children
        afterUnknownChild = true
    }
}
```

This change was applied to:
- Map property processing (object properties)
- Array element processing 
- Nil-to-map transitions
- Nil-to-array transitions

### 2. Nested Unknown Properties Collection

Added a new helper function `collectNestedUnknownProperties` that recursively collects all nested property paths when a parent object is marked as unknown:

```go
func (a *Analyzer) collectNestedUnknownProperties(basePath string, value any, analysis *PropertyChangeAnalysis) {
    // Recursively traverse the structure and add each nested property as unknown
}
```

This ensures that when a nested object like `config` is treated as a single change but is unknown, all its nested properties are still individually tracked in the `UnknownProperties` list.

## Files Modified

- `lib/plan/analyzer.go`: Enhanced unknown propagation logic and added nested property collection
- `lib/plan/unknown_parent_propagation_regression_test.go`: Added comprehensive regression tests

## Verification

### Regression Test
Created `TestUnknownParentPropagationToNestedProperties` with two test cases:

1. **Parent Object Unknown**: Tests that when `config` is entirely unknown, all nested properties (`config.timeout`, `config.retry_count`, `config.nested.deep_value`) are properly identified as unknown.

2. **Mixed Scenario**: Tests that when `tags` is entirely unknown, all its nested properties are identified as unknown.

### Test Results
- All existing tests continue to pass
- New regression tests pass
- Unknown value detection works correctly for both granular and parent-level unknown markers

## Impact

### Positive
- Nested properties under unknown parents are now correctly identified as unknown
- `UnknownProperties` list is complete and accurate
- Output correctly shows `(known after apply)` for all unknown nested properties
- Maintains backward compatibility with existing functionality

### Risk Assessment
- **Low Risk**: Changes are additive and don't modify existing logic paths
- All existing tests pass, indicating no regressions
- The fix handles edge cases (arrays, nil transitions) comprehensively

## Future Considerations

1. **Performance**: The nested property collection adds some overhead for unknown nested objects, but this is minimal and only occurs when objects are actually unknown.

2. **Deep Nesting**: The current implementation handles arbitrary nesting depth, but very deep structures might generate many individual property entries.

3. **Array Handling**: The fix properly handles arrays marked as entirely unknown, propagating the unknown status to all elements.

## Validation Steps

1. ✅ Regression test passes
2. ✅ All existing unknown value tests pass  
3. ✅ Full test suite passes
4. ✅ Manual verification with sample Terraform plans
5. ✅ Edge cases (arrays, deep nesting) handled correctly

## Conclusion

The bug has been successfully fixed with a comprehensive solution that handles both the immediate issue and related edge cases. The fix ensures that Terraform's `after_unknown` field is properly interpreted when parents are marked as entirely unknown, providing users with accurate and complete information about which properties will be known after apply.