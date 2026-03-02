# Bugfix Report: Statistics Total Excludes No-Op Resources

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

`ChangeStatistics.Total` was calculated as `ToAdd + ToChange + ToDestroy + Replacements`, so plans containing no-op resources reported a `Total Changes` value that was lower than the sum of displayed categories.

**Reproduction steps:**
1. Analyze a plan containing at least one no-op resource.
2. Compare `Total Changes` with `Added + Removed + Modified + Replacements + Unmodified`.
3. Observe that `Total Changes` is lower when `Unmodified > 0`.

**Impact:** Statistics were internally inconsistent and violated requirement 3.7 in `specs/output-refinements/requirements.md`, causing confusing summaries.

## Investigation Summary

Systematic debugging followed the Fagan-style workflow:

- **Phase 1 (Initial overview):** Confirmed mismatch only appears when no-op resources are present.
- **Phase 2 (Systematic inspection):** Inspected `lib/plan/analyzer.go` and found `Unmodified` is counted in the loop but omitted from the final `Total` assignment.
- **Phase 3 (Root cause analysis):** Determined this is a logic defect introduced when `Unmodified` support was added without updating the total aggregation rule.
- **Phase 4 (Solution & verification):** Added a regression test, fixed total calculation, and updated expectations asserting old behavior.

- **Symptoms examined:** Total mismatch in statistics summary for no-op plans.
- **Code inspected:** `lib/plan/analyzer.go`, `lib/plan/analyzer_statistics_test.go`, `lib/plan/output_refinements_test_compatibility_test.go`.
- **Hypotheses tested:** Parsing and display-layer issues were ruled out; defect is in statistics aggregation logic.

## Discovered Root Cause

`calculateStatistics` incremented `stats.Unmodified` for `ChangeTypeNoOp` resources but computed `stats.Total` without `stats.Unmodified`.

**Defect type:** Logic error (incomplete aggregation formula).

**Why it occurred:**
1. Total was computed from actionable categories only.
2. No-op tracking was later introduced via `Unmodified`.
3. The total formula was not updated to include the new category.
4. Existing tests encoded the old behavior, so the mismatch persisted.

**Contributing factors:** No invariant test explicitly enforcing `Total == Added + Removed + Modified + Replacements + Unmodified`.

## Resolution for the Issue

**Changes made:**
- `lib/plan/analyzer.go:994` - Updated `stats.Total` formula to include `stats.Unmodified`.
- `lib/plan/analyzer_statistics_test.go:320` - Added regression test `TestCalculateStatistics_TotalIncludesUnmodifiedResources`.
- `lib/plan/analyzer_statistics_test.go:246,260,299` - Updated expectations/comments for total behavior with no-ops.
- `lib/plan/output_refinements_test_compatibility_test.go:196` - Updated compatibility assertion to require total includes no-ops.

**Approach rationale:** Minimal targeted fix in the aggregation step aligns behavior with requirement 3.7 and existing summary schema.

**Alternatives considered:**
- Compute total via `len(changes)` - not chosen to avoid altering semantics if unsupported change types appear in future; summing explicit categories remains clear and intentional.

## Regression Test

**Test file:** `lib/plan/analyzer_statistics_test.go`
**Test name:** `TestCalculateStatistics_TotalIncludesUnmodifiedResources`

**What it verifies:** `Total` includes no-op resources and equals the sum of all resource categories.

**Run command:** `go test ./lib/plan -run TestCalculateStatistics_TotalIncludesUnmodifiedResources -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/analyzer.go` | Fixed total statistics aggregation to include unmodified resources |
| `lib/plan/analyzer_statistics_test.go` | Added regression test and updated no-op total expectations |
| `lib/plan/output_refinements_test_compatibility_test.go` | Updated compatibility test assertion for total behavior |
| `specs/bugfixes/statistics-total-excludes-no-op-resources/report.md` | Added bugfix investigation and verification report |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Ran `make build`.
- Ran `make run-sample SAMPLE=danger-sample.json` to verify normal summary output.
- Ran `./strata plan summary nonexistent-file.tfplan` to verify graceful error handling.

## Prevention

**Recommendations to avoid similar bugs:**
- Add/retain invariant tests for aggregate fields whenever new statistic categories are introduced.
- Keep requirement-linked assertions in tests when summary semantics change.
- During refactors, treat totals as derived from all exposed categories unless explicitly documented otherwise.

## Related

- Transit ticket: `T-172`
- Requirement reference: `specs/output-refinements/requirements.md` (3.7)
