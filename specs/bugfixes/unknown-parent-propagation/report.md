# Bugfix Report: Unknown Parent Propagation to Nested Properties

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

When Terraform marks a parent object as unknown in `after_unknown` (for example, `"tags": true`), nested properties were not always propagated as unknown.
The failure occurred when the parent object was treated as a grouped nested-object change and `before`/`after` values were equal.

**Reproduction steps:**
1. Build a resource change where `before.tags` and `after.tags` are identical maps.
2. Set `after_unknown` to `{"tags": true}`.
3. Run summary generation and observe `UnknownProperties` missing `tags` and nested paths.

**Impact:** Unknown values were under-reported, which hid uncertainty in nested plan fields and reduced trust in risk visibility.

## Investigation Summary

A systematic inspection focused on unknown-value handling in `compareObjects` and downstream unknown-property extraction.

- **Symptoms examined:** Missing `HasUnknownValues`, empty `UnknownProperties`, and no `(known after apply)` output for nested fields under unknown parents.
- **Code inspected:** `lib/plan/analyzer.go` (`compareObjects`, nested-object branch, unknown collection logic).
- **Hypotheses tested:**
  - Unknown traversal (`isValueUnknown`) failed for parent booleans (ruled out).
  - Child propagation failed on recursive map/array traversal (ruled out).
  - Nested-object shortcut skipped unknown handling when values were equal (confirmed).

## Discovered Root Cause

In the nested-object shortcut path of `compareObjects`, property-change creation was guarded by `!reflect.DeepEqual(before, after)` and then returned early.
For unknown parent objects with equal `before`/`after`, this condition prevented both the unknown parent change and `collectNestedUnknownProperties`, so propagation never occurred.

**Defect type:** Logic error (incorrect condition in control flow).

**Why it occurred:**
- The branch assumed only value differences should produce grouped nested-object changes.
- Unknown semantics require recording uncertainty even when concrete values compare equal.
- Early return prevented fallback recursion that could have surfaced unknown children.

**Contributing factors:**
- Grouped nested-object optimization combines presentation and traversal decisions in one conditional.

## Resolution for the Issue

**Changes made:**
- `lib/plan/analyzer.go:195` - Updated grouped nested-object condition to also create a change when `isUnknown` is true.
- `lib/plan/unknown_parent_propagation_regression_test.go` - Added regression case for unknown parent with unchanged nested values.

**Approach rationale:**
This is the smallest safe fix: keep existing grouping behavior, but ensure unknown semantics are always preserved.

**Alternatives considered:**
- Remove early return and recurse always for grouped nested objects — rejected due larger behavioral and output impact.

## Regression Test

**Test file:** `lib/plan/unknown_parent_propagation_regression_test.go`
**Test name:** `TestUnknownParentPropagationToNestedProperties/parent_unknown_with_unchanged_nested_object_should_still_propagate`

**What it verifies:** Unknown parent markers (`after_unknown` boolean on object) propagate to parent and nested properties even when `before` and `after` values are identical.

**Run command:** `go test ./lib/plan -run TestUnknownParentPropagationToNestedProperties -v`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/analyzer.go` | Fixed grouped nested-object condition to honor unknown parent propagation even for equal before/after values. |
| `lib/plan/unknown_parent_propagation_regression_test.go` | Added failing-then-passing regression scenario for unchanged nested object with unknown parent. |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Ran `make run-sample SAMPLE=danger-sample.json` to confirm summary output remains healthy.
- Ran `./strata plan summary nonexistent.tfplan` and verified graceful error output with usage text.

## Prevention

**Recommendations to avoid similar bugs:**
- Treat unknown-state semantics as independent from value-diff checks in grouped/object fast paths.
- Add regression tests that combine unknown markers with equal before/after values.
- Keep unknown propagation checks in both recursive and shortcut code paths.

## Related

- Transit task: `T-146`
