# Bugfix Report: Output Replace Action Shown as No-op

**Date:** 2026-03-02  
**Status:** Fixed

## Description of the Issue

Output changes with replace actions could be treated as no-op when `before` and `after` happened to be equal, which hid the output change by default and made the replace effectively appear as no-op behavior.

**Reproduction steps:**
1. Create a plan JSON with an output change using replace actions (`["delete","create"]`) and equal `before`/`after` values.
2. Run `strata plan summary <plan.json>` with default settings.
3. Observe the output change is filtered away as no-op.

**Impact:** Medium. Users can miss real output replacement events in summaries, especially in plans where replacement happens but rendered values are unchanged.

## Investigation Summary

Systematic inspection followed four phases: problem framing, code inspection, Five Whys root-cause analysis, and fix verification.

- **Symptoms examined:** Output replacement disappeared in default output when values were equal.
- **Code inspected:** `lib/plan/analyzer.go` (`analyzeOutputChange`), output filtering in `lib/plan/formatter.go`, and output tests in `lib/plan/analyzer_outputs_test.go`.
- **Hypotheses tested:**
  - Action mapping was broken (ruled out; replace already mapped correctly).
  - No-op detection logic was over-broad (confirmed).

## Discovered Root Cause

`analyzeOutputChange` marked outputs as no-op using `reflect.DeepEqual(change.Before, change.After)`.  
This ignored Terraform action metadata and incorrectly flagged some replace/create/update outputs as no-op when values matched.

**Defect type:** Logic error (incorrect no-op classification rule).

**Why it occurred:**  
1. No-op was inferred from value equality.  
2. Value equality was assumed to mean no change.  
3. Terraform action semantics were not used as source-of-truth for output operation type.

**Contributing factors:** Replace outputs are less common and existing tests did not assert `IsNoOp` for this scenario.

## Resolution for the Issue

**Changes made:**
- `lib/plan/analyzer.go` - Updated output no-op detection to use `changeType == ChangeTypeNoOp` instead of `reflect.DeepEqual(before, after)`.
- `lib/plan/analyzer_outputs_test.go` - Added regression test `TestAnalyzeOutputChangeReplaceWithEqualValuesIsNotNoOp`.

**Approach rationale:** Terraform action metadata is authoritative for operation intent; equal values should not reclassify replace actions as no-op.

**Alternatives considered:**
- Keep equality-based fallback for all actions — rejected because it repeats the same misclassification.

## Regression Test

**Test file:** `lib/plan/analyzer_outputs_test.go`  
**Test name:** `TestAnalyzeOutputChangeReplaceWithEqualValuesIsNotNoOp`

**What it verifies:** Replace output actions remain actionable (`IsNoOp == false`) even when before/after display values are equal.

**Run command:** `go test ./lib/plan -run TestAnalyzeOutputChangeReplaceWithEqualValuesIsNotNoOp -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/analyzer.go` | Corrected output no-op classification logic |
| `lib/plan/analyzer_outputs_test.go` | Added focused regression test |
| `specs/bugfixes/output-replace-action-no-op/report.md` | Updated bugfix documentation |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Verified `go run . plan summary /tmp/t87-repro-same.json` now shows output action `Replace` in the Output Changes table.
- Verified `make run-sample SAMPLE=danger-sample.json` works normally.
- Verified `./strata plan summary nonexistent.tfplan` still returns a clear load error.

## Prevention

**Recommendations to avoid similar bugs:**
- Use Terraform action metadata as primary classification input for no-op decisions.
- Add regression tests for equal-value outputs with non-no-op actions (replace/create/update).

## Related

- Transit ticket: T-87
