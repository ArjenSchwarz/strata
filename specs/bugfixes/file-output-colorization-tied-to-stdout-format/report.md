# Bugfix Report: File output colorization tied to stdout format

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

Colorization was controlled by `OutputConfiguration.UseColors`, which is derived from stdout output format.  
When stdout was non-markdown but `--file-format markdown` was used, file rendering still followed the stdout color flag and attempted to attach the color transformer for file output.

**Reproduction steps:**
1. Configure stdout output as a non-markdown format (for example `--output table`).
2. Configure file output as markdown (`--file-format markdown`).
3. Run `strata plan summary` with file output enabled and observe file output colorization decision tied to stdout settings.

**Impact:** Incorrect output behavior coupling between stdout and file output; markdown file output could receive color transformation logic intended for terminal output.

## Investigation Summary

Reviewed output configuration construction and rendering paths for stdout and file outputs independently.

- **Symptoms examined:** File output transformer selection reused stdout color flag.
- **Code inspected:** `config/config.go`, `lib/plan/formatter.go`, `lib/plan/formatter_file_output_test.go`.
- **Hypotheses tested:** Verified current code path applies `outputConfig.UseColors` to both stdout and file render options; added regression test for per-format color decision.

## Discovered Root Cause

`OutputSummary` used a single boolean (`UseColors`) for both destinations.  
No per-output format guard was applied before appending the color transformer for file output.

**Defect type:** Logic error (output destination coupling)

**Why it occurred:** Color usage logic was implemented once for stdout and reused for file output without considering that file format may differ from stdout format.

**Contributing factors:** Shared output configuration flag used across two independent render targets.

## Resolution for the Issue

**Changes made:**
- `lib/plan/formatter.go` - Introduced `shouldUseColorTransformer(useColors, outputFormat)` and used it separately for stdout and file transformer selection, disabling colors for markdown output.
- `lib/plan/formatter_file_output_test.go` - Added regression test for per-output color decision behavior.

**Approach rationale:** Centralized color decision in a small helper to keep logic explicit and consistent across both output targets.

**Alternatives considered:**
- Add separate config booleans for stdout/file colors - not chosen to keep changes minimal and avoid broader config surface changes.

## Regression Test

**Test file:** `lib/plan/formatter_file_output_test.go`  
**Test name:** `TestFormatter_FileOutputColorDecision`

**What it verifies:** Colors are enabled for table output when requested, disabled for markdown output even when `UseColors` is true, and disabled globally when `UseColors` is false.

**Run command:** `go test ./lib/plan -run TestFormatter_FileOutputColorDecision -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/formatter.go` | Added per-output color decision helper and applied it to stdout and file transformer selection |
| `lib/plan/formatter_file_output_test.go` | Added regression test for output-format-aware color decision |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Ran `make run-sample SAMPLE=danger-sample.json` to verify end-to-end summary output.
- Ran `./strata plan summary nonexistent.tfplan` and confirmed graceful error handling.

## Prevention

**Recommendations to avoid similar bugs:**
- Keep output-destination-specific decisions (stdout/file) isolated instead of sharing a single flag.
- Add focused tests whenever stdout and file output settings can diverge.
- Prefer helper functions for format-specific decisions to reduce duplicated conditional logic.

## Related

- Transit ticket `T-272`
