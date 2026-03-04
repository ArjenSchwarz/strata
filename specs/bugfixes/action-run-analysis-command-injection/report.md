# Bugfix Report: Action run_analysis command injection

**Date:** 2026-03-02
**Status:** Fixed

## Description of the Issue

`action.sh` built the Strata invocation as one shell string and executed it with unquoted expansion. This caused shell word-splitting of file paths (for example paths containing spaces), and plan paths starting with `-` could be parsed as options.

**Reproduction steps:**
1. Set `INPUT_PLAN_FILE` and `INPUT_CONFIG_FILE` to values containing spaces.
2. Run the action and reach `run_analysis`.
3. Observe malformed argv handling and analysis failure.

**Impact:** GitHub Action runs could fail or mis-handle user-provided file paths; argument safety was unreliable for untrusted inputs.

## Investigation Summary

Reviewed `action.sh` command construction and execution flow, then reproduced with a focused harness test in `test/action_test.sh`.

- **Symptoms examined:** analysis failures for valid files when paths contain spaces or begin with `-`
- **Code inspected:** `action.sh` `run_analysis`, existing action unit test coverage
- **Hypotheses tested:** command string splitting versus array execution; end-of-options delimiter handling

## Discovered Root Cause

`run_analysis` concatenated all arguments into a single string (`cmd="..."`) and executed it as `$cmd`, which applies shell word splitting and does not preserve argument boundaries.

**Defect type:** Missing input-safe command invocation (argument splitting bug).

**Why it occurred:** command assembly used string concatenation instead of argv array semantics.

**Contributing factors:** no explicit `--` delimiter before the plan-file positional argument.

## Resolution for the Issue

**Changes made:**
- `action.sh:243-257` - switched from string command building to bash array (`cmd=(...)`, `cmd+=(...)`) and executed via `"${cmd[@]}"`; added `--` before `INPUT_PLAN_FILE`.
- `test/action_test.sh:580-691` - added regression test `test_run_analysis_argument_safety` validating path-with-spaces and leading-dash plan file behavior using extracted `run_analysis` and a strict mock `strata` parser.

**Approach rationale:** arrays preserve each argument exactly, preventing path splitting and ensuring option parsing stops before the plan path.

**Alternatives considered:**
- Escape individual values into a string with quoting helpers - rejected as brittle and easier to regress than native argv arrays.

## Regression Test

**Test file:** `test/action_test.sh`
**Test name:** `test_run_analysis_argument_safety`

**What it verifies:** `run_analysis` passes config/plan file paths as intact arguments, uses `--` before plan file, and supports plan paths beginning with `-`.

**Run command:** `make test-action-unit`

## Affected Files

| File | Change |
|------|--------|
| `action.sh` | Safe argv array command construction and execution in `run_analysis` |
| `test/action_test.sh` | Added regression coverage for argument safety in `run_analysis` |
| `specs/bugfixes/action-run-analysis-command-injection/report.md` | Bug investigation and fix documentation |

## Verification

**Automated:**
- [x] Regression test passes
- [x] Full test suite passes
- [x] Linters/validators pass

**Manual verification:**
- `make run-sample SAMPLE=danger-sample.json`
- `./strata plan summary nonexistent.tfplan` (confirmed non-zero error handling)

## Prevention

**Recommendations to avoid similar bugs:**
- Build external commands as arrays, never concatenated strings.
- Always insert `--` before positional user-supplied paths.
- Keep regression coverage for shell argument safety in action unit tests.

## Related

- Transit ticket `T-280`
