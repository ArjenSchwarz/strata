# Bugfix Report: Output Replace Action Shown as No-op

**Date:** 2025-01-03
**Status:** Fixed

## Description of the Issue

Terraform output changes with replace actions (delete + create) were incorrectly displayed as "No-op" instead of "Replace" in the plan summary output.

**Reproduction steps:**
1. Create a Terraform plan with an output that has both delete and create actions (replace)
2. Run `strata plan summary` on the plan file
3. Observe that the output action is shown as "No-op" instead of "Replace"

**Impact:** Medium - Users would see misleading "No-op" actions for output changes that were actually replacements, reducing clarity in plan summaries and creating inconsistency with resource change handling.

## Investigation Summary

Systematic debugging revealed the issue was in the `getOutputActionAndIndicator` function in `lib/plan/analyzer.go`.

- **Symptoms examined:** Replace actions showing as "No-op" in output tables
- **Code inspected:** `getOutputActionAndIndicator` function, `FromTerraformAction` function, test cases
- **Hypotheses tested:** 
  - Incorrect action detection in `FromTerraformAction` (ruled out - correctly returns `ChangeTypeReplace`)
  - Missing case handling in `getOutputActionAndIndicator` (confirmed as root cause)

## Discovered Root Cause

The `getOutputActionAndIndicator` function did not handle the `ChangeTypeReplace` case, causing replace actions to fall through to the default case and return "No-op".

**Defect type:** Missing case handling in switch statement

**Why it occurred:** The function was implemented with cases for Create, Update, and Delete, but Replace was overlooked despite being a valid ChangeType.

**Contributing factors:** 
- Replace actions are less common for outputs than resources
- Existing tests were expecting the incorrect "No-op" behavior, masking the bug

## Resolution for the Issue

**Changes made:**
- `lib/plan/analyzer.go:925` - Added `case ChangeTypeReplace: return "Replace", "~"` to handle replace actions
- `lib/plan/analyzer_outputs_test.go:580` - Updated test to expect "Replace" with "~" indicator
- `lib/plan/analyzer_outputs_test.go:503` - Updated test to expect correct replace action behavior

**Approach rationale:** Added the missing case to align with how resource changes handle replace actions, maintaining consistency across the codebase.

**Alternatives considered:**
- Map replace to "Modify" - Rejected as it would be misleading since replace is more destructive than modify
- Keep as "No-op" - Rejected as it's factually incorrect and inconsistent

## Regression Test

**Test file:** `lib/plan/analyzer_outputs_test.go`
**Test name:** `TestGetOutputActionAndIndicator/replace_should_return_Replace_with_~`

**What it verifies:** That replace actions return "Replace" with "~" indicator instead of "No-op"

**Run command:** `go test -v ./lib/plan -run TestGetOutputActionAndIndicator`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/analyzer.go` | Added missing case for ChangeTypeReplace |
| `lib/plan/analyzer_outputs_test.go` | Updated tests to expect correct behavior |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Build completes successfully

**Manual verification:**
- Created and ran test program confirming replace actions now return "Replace" with "~" indicator
- Verified ChangeType is correctly identified as "replace"

## Prevention

**Recommendations to avoid similar bugs:**
- When adding new ChangeType values, ensure all switch statements handling ChangeType are updated
- Consider using exhaustive switch linting rules to catch missing cases
- Review test expectations more critically - tests expecting "wrong" behavior can mask bugs

## Related

- Transit ticket: T-87
- Commit: be0b1542c35e4927ca42ad5ea7873c2c597e1e59