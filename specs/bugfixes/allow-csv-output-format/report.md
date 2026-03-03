# Bugfix Report: Allow CSV Output Format in Formatter Validation

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

`strata plan summary --output csv` failed even though CSV rendering exists in the formatter.

**Reproduction steps:**
1. Build the CLI (`make build`).
2. Run `./strata plan summary --output csv samples/simpleadd-sample.json`.
3. Observe: `unsupported output format 'csv'. Supported formats: table, json, html, markdown`.

**Impact:** Users could not use CSV output for plan summary despite CSV support in format conversion logic.

## Investigation Summary

Validated the CLI error path and traced output format handling in formatter validation and conversion.

- **Symptoms examined:** CLI rejected `--output csv` with unsupported format error.
- **Code inspected:** `lib/plan/formatter.go`, `lib/plan/formatter_basic_test.go`, `cmd/root.go`.
- **Hypotheses tested:** Confirmed `getFormatFromConfig` supported CSV while `ValidateOutputFormat` did not.

## Discovered Root Cause

`Formatter.ValidateOutputFormat` used a hardcoded allowlist that omitted `csv`, while `getFormatFromConfig` already supported `csv`.

**Defect type:** Missing validation case / allowlist mismatch.

**Why it occurred:** Supported format definitions were duplicated in different formatter functions and diverged.

**Contributing factors:** Validation and conversion paths were not kept in sync by tests for CSV.

## Resolution for the Issue

**Changes made:**
- `lib/plan/formatter.go:48` - Added `csv` to `ValidateOutputFormat` supported formats.
- `lib/plan/formatter_basic_test.go:21` - Added regression test coverage for `csv` in `TestFormatter_ValidateOutputFormat`.
- `cmd/root.go:81` - Updated `--output` help text to include `csv`.

**Approach rationale:** Minimal, targeted fix to align validation behavior with existing conversion support.

**Alternatives considered:**
- Derive validation directly from conversion mapping - not chosen to keep change scope minimal for this ticket.

## Regression Test

**Test file:** `lib/plan/formatter_basic_test.go`
**Test name:** `TestFormatter_ValidateOutputFormat/csv`

**What it verifies:** `csv` is accepted by formatter output validation.

**Run command:** `go test ./lib/plan -run TestFormatter_ValidateOutputFormat -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/formatter.go` | Added `csv` to supported output formats in validation. |
| `lib/plan/formatter_basic_test.go` | Added regression test case for `csv`. |
| `cmd/root.go` | Updated CLI help text to list CSV output support. |
| `specs/bugfixes/allow-csv-output-format/report.md` | Added bugfix report. |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- `./strata plan summary --output csv samples/simpleadd-sample.json` now renders CSV output successfully.
- `./strata plan summary --help` now lists CSV in output formats.
- `./strata plan summary nonexistent-file.tfplan` still returns a graceful error and usage output.

## Prevention

**Recommendations to avoid similar bugs:**
- Keep supported output formats defined in one shared source of truth.
- Ensure compatibility tests include every supported user-facing output format.
- Keep CLI help text and validation allowlists synchronized with renderer support.

## Related

- Transit ticket: `T-313`
