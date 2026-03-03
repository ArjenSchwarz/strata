# Bugfix Report: Dot Output Format Validation Mismatch

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

`OutputFileFormat=dot` was accepted by configuration validation, but the formatter had no `dot` handler and silently rendered table output.

**Reproduction steps:**
1. Configure file output with `OutputFileFormat` set to `dot`.
2. Run plan summary rendering with file output enabled.
3. Observe validation passes but output is table-formatted instead of dot (or an explicit error).

**Impact:** Misleading behaviour for file-output users; an unsupported format appeared valid and produced incorrect output.

## Investigation Summary

Validation and rendering paths were compared to confirm format support parity.

- **Symptoms examined:** `dot` passed `validateFormatSupport` but was not rendered as dot.
- **Code inspected:** `config/validation.go`, `config/validation_file_test.go`, `lib/plan/formatter.go`.
- **Hypotheses tested:** (1) dot renderer existed under a different name (ruled out), (2) fallback path in formatter was being used (confirmed).

## Discovered Root Cause

The supported-format allowlist in config validation drifted from actual formatter capabilities.

**Defect type:** Missing validation alignment (logic/config mismatch)

**Why it occurred:** `dot` remained in `validateFormatSupport` while `Formatter.getFormatFromConfig` only supports json/csv/html/markdown/table and defaults unknown formats to table.

**Contributing factors:** No dedicated regression test enforcing rejection of unsupported-but-fallback formats.

## Resolution for the Issue

**Changes made:**
- `config/validation.go:118` - Removed `dot` from supported format list and updated supported-format comment.
- `config/validation_file_test.go:104` - Updated table-driven expectation so `dot` is unsupported.
- `config/validation_file_test.go:145` - Added regression test `TestFileValidator_ValidateFormatSupport_DotFormatRejected`.

**Approach rationale:** Removing `dot` from validation is the smallest safe fix that restores consistency without introducing new renderer behavior.

**Alternatives considered:**
- Add dot rendering support in formatter - not chosen because no existing dot rendering implementation/path exists in current formatter logic.

## Regression Test

**Test file:** `config/validation_file_test.go`
**Test name:** `TestFileValidator_ValidateFormatSupport_DotFormatRejected`

**What it verifies:** `dot` is rejected with an unsupported format error, preventing silent table fallback.

**Run command:** `go test ./config -run TestFileValidator_ValidateFormatSupport_DotFormatRejected -count=1`

## Affected Files

| File | Change |
|------|--------|
| `config/validation.go` | Removed stale `dot` support from validator allowlist |
| `config/validation_file_test.go` | Added/updated tests to enforce `dot` rejection |
| `specs/bugfixes/dot-format-validation/report.md` | Added formal bugfix report |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Built binary: `make build`
- End-to-end sample run: `make run-sample SAMPLE=danger-sample.json`
- Error handling check: `./strata plan summary nonexistent.tfplan` shows clear error + usage

## Prevention

**Recommendations to avoid similar bugs:**
- Keep validator format allowlists aligned with formatter switch cases.
- Add parity tests for accepted formats versus renderer-supported formats.
- Prefer explicit unsupported-format errors over renderer fallback for unknown values.

## Related

- Transit ticket: `T-186`
