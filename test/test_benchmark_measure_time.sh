#!/bin/bash

# Regression test for T-1228: Performance benchmark tests discard measured durations
#
# Bug: In test/test_performance_benchmarks.sh the call sites used:
#     duration=$(measure_time timeout 60 ./action.sh > "$action_log" 2>&1)
# The redirection applied to measure_time itself, so the duration that
# measure_time prints on stdout was redirected into the log file instead of
# being captured into $duration. The command substitution therefore captured
# an empty string, making every subsequent arithmetic comparison unreliable.
#
# This test exercises the measure_time helper the same way the benchmark
# callers do and asserts that:
#   1. duration is captured as a non-empty integer (the bug left it empty)
#   2. the timed command's stdout/stderr lands in the log file, not the duration

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BENCHMARK_SCRIPT="$SCRIPT_DIR/test_performance_benchmarks.sh"

FAILED=0

fail() {
    echo "[FAIL] $1"
    FAILED=1
}

pass() {
    echo "[PASS] $1"
}

# Extract just the measure_time function definition from the benchmark script so
# we can call it without running the whole suite (which would invoke action.sh
# and download binaries).
extract_measure_time() {
    awk '/^measure_time\(\) \{/{flag=1} flag{print} flag&&/^\}/{exit}' "$BENCHMARK_SCRIPT"
}

if ! grep -q "^measure_time()" "$BENCHMARK_SCRIPT"; then
    fail "Could not find measure_time definition in $BENCHMARK_SCRIPT"
    exit 1
fi

# Source the extracted function into the current shell.
eval "$(extract_measure_time)"

# Verify the benchmark script no longer redirects the command substitution that
# wraps measure_time (the root cause of the bug). The fix routes the timed
# command's output through measure_time's own log-file argument instead.
# Inspect only the executable call sites (lines containing a measure_time
# command substitution), ignoring comments, and flag any that still carry an
# output redirection on the outer command substitution.
if grep -nE '=\$\(measure_time ' "$BENCHMARK_SCRIPT" \
    | grep -vE '^[0-9]+:[[:space:]]*#' \
    | grep -E '> *"?\$?[A-Za-z_/]' >/dev/null; then
    fail "Found measure_time call substitution with an output redirection on the outer command (reintroduces T-1228)"
else
    pass "No measure_time call substitutions redirect the outer command"
fi

# A fake command that writes to stdout AND stderr, mimicking ./action.sh output.
fake_command() {
    echo "stdout line that belongs in the log"
    echo "stderr line that belongs in the log" >&2
    return 0
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
action_log="$TMP_DIR/action.log"

# Call measure_time exactly the way the fixed benchmark callers do: pass the log
# file path so measure_time redirects the timed command's output, and capture
# the printed duration via command substitution.
duration=""
if duration=$(measure_time "$action_log" fake_command); then
    : # ran successfully
else
    fail "measure_time returned non-zero for a successful command"
fi

# 1. duration must be a non-empty integer (the bug left it empty).
if [[ "$duration" =~ ^[0-9]+$ ]]; then
    pass "Captured duration is a non-empty integer: '$duration'"
else
    fail "Captured duration is not a valid integer: '$duration' (bug discards the measured duration)"
fi

# 2. The command's output must be in the log file, not lost or mixed with the duration.
if [ -f "$action_log" ] && grep -q "stdout line that belongs in the log" "$action_log" \
    && grep -q "stderr line that belongs in the log" "$action_log"; then
    pass "Timed command stdout and stderr were written to the log file"
else
    fail "Timed command output did not reach the log file"
fi

# 3. The log file must NOT contain the duration value (that would mean the
#    redirection swallowed measure_time's printed duration, as in the bug).
if grep -qx "$duration" "$action_log"; then
    fail "Log file contains the duration value, indicating the redirection captured measure_time's output (T-1228 bug)"
else
    pass "Duration value is not leaked into the log file"
fi

echo ""
if [ "$FAILED" -eq 0 ]; then
    echo "All measure_time regression checks passed."
    exit 0
else
    echo "measure_time regression checks failed."
    exit 1
fi
