# Bugfix Report: Capture array removals in property change analysis

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

When a Terraform attribute represented as an array was removed, property change analysis could omit that removal entirely.

**Reproduction steps:**
1. Run `compareObjects` for a removed list attribute (e.g. `before: {"items": [1,2]}`, `after: {}`).
2. Let map recursion call `compareObjects("items", []any{1,2}, nil, ...)`.
3. Observe no `remove` `PropertyChange` is emitted.

**Impact:** Removed list attributes were missing from property-level change output, reducing accuracy of risk and review detail.

## Investigation Summary

Applied a systematic inspection of `compareObjects` control flow for map, slice, and nil transitions.

- **Symptoms examined:** Removed list attributes had zero corresponding property changes.
- **Code inspected:** `lib/plan/analyzer.go` slice branch in `compareObjects`; `lib/plan/analyzer_utils_test.go` coverage around object comparisons.
- **Hypotheses tested:**
  - Missing map key handling (ruled out: map branch correctly visits removed keys).
  - Early-return in slice branch when `after` is nil/non-slice (confirmed).
  - Existing leaf-change detection would catch it (ruled out: complex types are intentionally excluded from leaf change creation).

## Discovered Root Cause

`compareObjects` had an unconditional early return in the `case []any` branch when `after` was `nil` or not a slice. That path did not emit a change and did not recurse, so list removals were dropped.

**Defect type:** Logic error / missing boundary condition handling.

**Why it occurred:**
- Why was the removal missing? Because the slice branch returned immediately.
- Why did it return immediately? It only handled valid `[]any` comparisons.
- Why was no fallback used? Removal/non-slice transition logic was not implemented for this branch.
- Why did tests not catch it? Existing tests covered scalar removals and array length updates, not full array attribute removal.

**Contributing factors:** Complex-type leaf filtering (`!isComplexType`) prevented fallback change creation earlier in the function, making the slice-branch gap decisive.

## Resolution for the Issue

**Changes made:**
- `lib/plan/analyzer.go:339` - Added explicit removal emission when `before` is `[]any` and `after == nil`.
- `lib/plan/analyzer.go:370` - Added fallback recursion for `before` slice + `after` non-slice transitions.
- `lib/plan/analyzer_utils_test.go:932` - Added regression case `array property removal` to assert removal is captured.

**Approach rationale:** Keep changes minimal in the existing slice branch while preserving current behavior for same-type array comparisons.

**Alternatives considered:**
- Generalize leaf change creation for all complex types - rejected as broader behavior change with higher regression risk.

## Regression Test

**Test file:** `lib/plan/analyzer_utils_test.go`
**Test name:** `TestCompareObjectsEnhanced/array_property_removal`

**What it verifies:** Removing an array-valued attribute emits a `remove` `PropertyChange` with full `Before` slice and `After: nil`.

**Run command:** `go test ./lib/plan -run TestCompareObjectsEnhanced -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/analyzer.go` | Added missing slice-removal and non-slice fallback handling in `compareObjects` |
| `lib/plan/analyzer_utils_test.go` | Added regression coverage for removed list attributes |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Ran `make run-sample SAMPLE=danger-sample.json` to verify end-to-end CLI summary output.
- Ran `./strata plan summary nonexistent.tfplan` to verify error handling remains correct.

## Prevention

**Recommendations to avoid similar bugs:**
- Add transition-focused tests for collection type changes (`slice→nil`, `slice→map`, `slice→primitive`).
- Prefer explicit fallback handling before early returns in type-switch branches.

## Related

- Transit ticket: `T-327`
- Bugfix report: `specs/bugfixes/capture-array-removals-in-property-change-analysis/report.md`
