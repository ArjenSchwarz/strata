# Bugfix Report: Sensitive-only output file message suppression

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

In sensitive-only mode (`AlwaysShowSensitive=true`, `showDetails=false`), Strata should tell users when no sensitive changes are present.  
When `outputConfig.OutputFile` was set, stdout no longer showed `No sensitive resource changes detected.` even though the same rendered document is used for both stdout and file output.

**Reproduction steps:**
1. Configure sensitive-only mode (`AlwaysShowSensitive=true`) and run summary with details disabled.
2. Use a plan with non-sensitive changes only.
3. Set an output file and run summary; observe stdout has no no-sensitive message.

**Impact:** Users running sensitive-only mode with file output got no explicit indication that there were zero sensitive changes.

## Investigation Summary

- **Symptoms examined:** Missing no-sensitive message only when output file was configured.
- **Code inspected:** `lib/plan/formatter.go` (`handleResourceDisplay`, `handleSensitiveResourceDisplay`).
- **Hypotheses tested:** Confirmed message was guarded by `outputConfig.OutputFile == ""`; reproduced with a failing regression test.

## Discovered Root Cause

`handleSensitiveResourceDisplay` had output-destination logic mixed into document construction:
- It only appended `No sensitive resource changes detected.` when `OutputFile` was empty.
- Because the same document is rendered to stdout and file, this suppressed the message for stdout too whenever file output was enabled.

**Defect type:** Logic error (incorrect conditional gating).

**Why it occurred:** Message inclusion was tied to output configuration instead of actual sensitive-change state.

**Contributing factors:** Shared-doc rendering model (stdout + file) makes destination-specific gating unsafe in builder stage.

## Resolution for the Issue

**Changes made:**
- `lib/plan/formatter.go:1253-1270` - Removed `OutputFile`-dependent guard; always add the no-sensitive message when there are zero sensitive changes.
- `lib/plan/formatter_file_output_test.go:81-151` - Added regression test `TestFormatter_FileOutput_SensitiveOnlyNoSensitiveMessageStillShown`.

**Approach rationale:** Minimal, targeted fix aligns behavior with shared document rendering and preserves existing output structure.

**Alternatives considered:**
- Build separate documents for stdout and file output - not chosen because it is a larger architectural change for a narrow bug.

## Regression Test

**Test file:** `lib/plan/formatter_file_output_test.go`  
**Test name:** `TestFormatter_FileOutput_SensitiveOnlyNoSensitiveMessageStillShown`

**What it verifies:** In sensitive-only mode with no sensitive changes and output file enabled, stdout still includes `No sensitive resource changes detected.` and file output still works.

**Run command:** `go test ./lib/plan -run TestFormatter_FileOutput_SensitiveOnlyNoSensitiveMessageStillShown -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/formatter.go` | Fixed sensitive-only no-sensitive message logic to be output-file agnostic |
| `lib/plan/formatter_file_output_test.go` | Added regression test covering output file + stdout behavior |
| `specs/bugfixes/sensitive-only-output-file-message/report.md` | Added bugfix documentation |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Ran `make run-sample SAMPLE=danger-sample.json` successfully.
- Ran `./strata plan summary nonexistent.tfplan` and confirmed graceful error output.

## Prevention

**Recommendations to avoid similar bugs:**
- Keep destination-specific behavior out of document construction code when one document is rendered to multiple outputs.
- Add regression tests for stdout behavior when file output is enabled for display-mode code paths.

## Related

- Transit ticket: `T-258`
