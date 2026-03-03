# Bugfix Report: Handle Remote Backend Workspaces Prefix in Location Display

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

`Parser.extractBackendLocation` only handled Terraform remote backend configs that use `workspaces.name`.
When a remote backend used `workspaces.prefix`, backend location detection fell back to `terraform.tfstate`, which is incorrect for Terraform Cloud/Enterprise remote state.

**Reproduction steps:**
1. Configure a remote backend with `organization` and `workspaces.prefix` (without `workspaces.name`).
2. Set current workspace (for example `TF_WORKSPACE=prod`).
3. Parse backend location and observe it returns `terraform.tfstate` instead of a remote location.

**Impact:** Plan summaries displayed incorrect backend locations for common remote backend prefix setups.

## Investigation Summary

Reviewed backend parsing flow in `lib/plan/parser.go` and found remote backend handling only checked `workspaces.name`.

- **Symptoms examined:** Remote backend location displayed as local tfstate path.
- **Code inspected:** `lib/plan/parser.go`, `lib/plan/parser_test.go`.
- **Hypotheses tested:** Missing prefix branch in remote backend logic was causing fallback behavior.

## Discovered Root Cause

The remote backend formatter lacked support for `workspaces.prefix`, so prefix-based configurations could not be resolved and always hit fallback logic.

**Defect type:** Logic error (incomplete condition handling)

**Why it occurred:** Implementation initially covered only explicit workspace names and omitted Terraform's prefix-based workspace mapping mode.

**Contributing factors:** No regression test for `workspaces.prefix` behavior.

## Resolution for the Issue

**Changes made:**
- `lib/plan/parser.go` - Added `getCurrentWorkspace()` helper and extended remote backend location extraction to support `workspaces.prefix` by combining prefix + current workspace.
- `lib/plan/parser_test.go` - Added regression test validating remote backend prefix location rendering.

**Approach rationale:** Minimal, targeted change in parser logic that preserves existing behavior and only extends remote backend handling for the missing supported mode.

**Alternatives considered:**
- Resolve workspace only from separate summary context - rejected because backend location extraction should remain self-contained and deterministic.

## Regression Test

**Test file:** `lib/plan/parser_test.go`
**Test name:** `TestParser_extractBackendLocation_RemoteWorkspacePrefix`

**What it verifies:** Remote backend with `workspaces.prefix` returns `app.terraform.io/<org>/<prefix><workspace>` when current workspace is available.

**Run command:** `go test ./lib/plan -run TestParser_extractBackendLocation_RemoteWorkspacePrefix -count=1`

## Affected Files

| File | Change |
|------|--------|
| `lib/plan/parser.go` | Added prefix-based remote backend location handling |
| `lib/plan/parser_test.go` | Added regression test for prefix mode |
| `specs/bugfixes/handle-remote-backend-workspaces-prefix-in-location-display/report.md` | Added bugfix documentation |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- Ran `make run-sample SAMPLE=danger-sample.json` to confirm normal summary behavior.
- Ran `./strata plan summary nonexistent.tfplan` to confirm graceful error handling remains intact.

## Prevention

**Recommendations to avoid similar bugs:**
- Add parser tests for backend config variants (name vs prefix paths).
- Reuse shared workspace resolution helpers for all backend branches requiring workspace context.

## Related

- Transit ticket: `T-205`
