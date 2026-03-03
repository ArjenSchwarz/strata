# Bugfix Report: macOS checksum verification in action_simplified.sh

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

action_simplified.sh always executed `sha256sum` during release tarball verification. On macOS runners, `sha256sum` is typically not installed, so the command failed and `set -e` aborted `download_strata`.

**Reproduction steps:**
1. Run `download_strata` from `action_simplified.sh` in an environment where `sha256sum` is unavailable.
2. Ensure a valid checksum exists and `shasum` is available.
3. Observe `sha256sum: command not found` and script exit.

**Impact:** GitHub Action runs on macOS could fail before analysis starts, preventing plan summary output.

## Investigation Summary

- **Symptoms examined:** `download_strata` exited with code 127 when `sha256sum` was missing.
- **Code inspected:** `action_simplified.sh` checksum verification block in `download_strata`.
- **Hypotheses tested:** Missing checksum binary vs bad checksums file vs download failures; failure reproduced specifically from unconditional `sha256sum` execution.

## Discovered Root Cause

`download_strata` used `sha256sum` unconditionally and had no macOS-compatible fallback. Because the script runs with `set -euo pipefail`, missing command execution terminated the action.

**Defect type:** Missing platform compatibility handling.

**Why it occurred:** The simplified action path did not replicate the md5 fallback pattern already present in `action.sh`.

**Contributing factors:** Strict shell error mode (`set -e`) made the missing command fatal.

## Resolution for the Issue

**Changes made:**
- `action_simplified.sh:220-226` - Added checksum tool detection: prefer `sha256sum`, fallback to `shasum -a 256`, otherwise log warning and skip checksum verification.
- `test/test_binary_download.sh:416-499,679` - Added regression test `test_action_simplified_checksum_fallback` that executes `download_strata` from `action_simplified.sh` with `sha256sum` intentionally unavailable.

**Approach rationale:** This keeps verification enabled on Linux, supports macOS natively, and avoids hard-failing when SHA-256 tooling is unavailable.

**Alternatives considered:**
- Fail hard when no SHA-256 tool exists — rejected to avoid breaking macOS runner compatibility.

## Regression Test

**Test file:** `test/test_binary_download.sh`
**Test name:** `test_action_simplified_checksum_fallback`

**What it verifies:** `action_simplified.sh` successfully verifies checksums via `shasum` when `sha256sum` is unavailable.

**Run command:** `./test/test_binary_download.sh`

## Affected Files

| File | Change |
|------|--------|
| `action_simplified.sh` | Added SHA-256 command fallback and safe behavior when tool is unavailable |
| `test/test_binary_download.sh` | Added regression harness covering macOS-style missing `sha256sum` |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Reproduced original failure path (missing `sha256sum`) and confirmed success after fix using the same harness conditions.

## Prevention

**Recommendations to avoid similar bugs:**
- Apply explicit command-availability checks for platform-specific utilities in shell actions.
- Add compatibility-focused tests for macOS/Linux tool differences when introducing checksum or hashing steps.

## Related

- Transit ticket: T-284
