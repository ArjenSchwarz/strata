# Bugfix Report: File output permission check fails if .strata_write_test exists

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

`validateDirectoryPermissions` attempted to create a fixed marker file (`.strata_write_test`) using `O_EXCL` to verify directory write access. If that file already existed (left over from a previous interrupted run or another process), the check returned `PERMISSION_DENIED` even when the directory was writable.

**Reproduction steps:**
1. Create a writable directory.
2. Create a file named `.strata_write_test` inside it.
3. Call `validateDirectoryPermissions` for an output file in that directory and observe a false permission error.

**Impact:** File output validation could fail incorrectly, blocking valid writes and surfacing misleading permission errors.

## Investigation Summary

A focused investigation was performed on the file-output validation path.

- **Symptoms examined:** false `PERMISSION_DENIED` when marker file already existed.
- **Code inspected:** `config/validation.go`, `config/validation_file_test.go`.
- **Hypotheses tested:** non-writable directory vs fixed-name collision; collision hypothesis confirmed via regression test.

## Discovered Root Cause

The permission check used a deterministic filename (`.strata_write_test`) together with exclusive creation (`O_EXCL`). Existing-file collisions were treated the same as genuine permission failures.

**Defect type:** Logic error (false negative in permission validation)

**Why it occurred:** The implementation conflated "cannot create because file exists" with "cannot write to directory".

**Contributing factors:** Reuse of a fixed test filename across runs/processes.

## Resolution for the Issue

**Changes made:**
- `config/validation.go:100` - replaced fixed filename + `os.OpenFile(...O_EXCL...)` with `os.CreateTemp(dir, ".strata_write_test_*")` and cleanup of the created temp file.
- `config/validation_file_test.go:174` - added regression subtest for pre-existing `.strata_write_test` file.

**Approach rationale:** `os.CreateTemp` guarantees a unique file name in the target directory while still validating real write access.

**Alternatives considered:**
- Remove existing fixed file and retry - rejected because it can delete unrelated files and remains race-prone.

## Regression Test

**Test file:** `config/validation_file_test.go`
**Test name:** `TestFileValidator_ValidateDirectoryPermissions/existing_test_marker_file_should_not_fail_permission_check`

**What it verifies:** A pre-existing `.strata_write_test` no longer causes permission validation to fail for writable directories.

**Run command:** `go test ./config -run TestFileValidator_ValidateDirectoryPermissions -count=1`

## Affected Files

| File | Change |
|------|--------|
| `config/validation.go` | Use unique temp file for directory write-permission validation |
| `config/validation_file_test.go` | Add regression coverage for existing marker file case |
| `specs/bugfixes/file-output-permission-check-fails-if-strata-write-test-exists/report.md` | Document investigation, fix, and verification |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- `make run-sample SAMPLE=danger-sample.json` still produces expected summary output.
- `./strata plan summary nonexistent.tfplan` still returns a clear error and usage output.

## Prevention

**Recommendations to avoid similar bugs:**
- Prefer `os.CreateTemp` over fixed probe filenames for filesystem writability checks.
- Treat `EEXIST` separately from permission errors when using explicit file probes.
- Keep regression tests for edge cases involving leftover artifacts from interrupted runs.

## Related

- Transit ticket: `T-185`
