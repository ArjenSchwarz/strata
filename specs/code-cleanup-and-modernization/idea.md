# Test Suite Cleanup Plan for Strata

## Overview
After reviewing all 27 Go test files against the testing guidelines, I've identified several areas where the test suite can be improved to follow Go best practices and the project's testing guidelines.

## Key Findings

### Current State
- **301 test functions and subtests** across the codebase
- **Good coverage** with table-driven tests widely used
- **Using testify** (assert/require) in 16 files - guideline prefers standard library when possible
- **No parallel tests** - missing `t.Parallel()` calls
- **No test helpers marked** - missing `t.Helper()` in helper functions
- **Using defer for cleanup** instead of `t.Cleanup()`
- **Slice-based test tables** instead of map-based (guideline prefers maps)
- **Inconsistent naming** - using "expected" instead of "want", "result" instead of "got"

## Proposed Changes

### 1. Test Helper Marking (Priority: High)
- Add `t.Helper()` to helper functions in test files
- Functions to update: `testFormatterSortingBackwardCompatibility`, `testPropertySortingBackwardCompatibility`, etc.

### 2. Cleanup Method Migration (Priority: Medium)
- Replace `defer os.Remove()` and `defer os.RemoveAll()` with `t.Cleanup()`
- Approximately 20+ instances to update

### 3. Naming Convention Standardization (Priority: Medium)
- Rename test variables from "expected" → "want" and "result/actual" → "got"
- Update across all test files for consistency

### 4. Parallel Test Implementation (Priority: Low)
- Add `t.Parallel()` to suitable test functions
- Focus on unit tests that don't share resources
- Skip for integration tests that may have dependencies

### 5. Test Table Migration (Priority: Low)
- Consider converting critical test tables from slice-based to map-based
- Prioritize tests with many cases where unique naming would help

### 6. Reduce Testify Dependency (Priority: Low)
- Where simple equality checks are used, replace with standard library
- Keep testify for complex assertions that would be verbose otherwise

### 7. Test Documentation (Priority: Low)
- Add documentation comments to exported test helpers
- Document complex test scenarios

## Implementation Order

1. **Phase 1: Helper Functions & Cleanup**
   - Mark all test helpers with `t.Helper()`
   - Replace defer with `t.Cleanup()`

2. **Phase 2: Naming Conventions**
   - Standardize to got/want naming
   - Update assertion messages

3. **Phase 3: Parallel Tests**
   - Add `t.Parallel()` to independent unit tests
   - Ensure proper variable capture in subtests

4. **Phase 4: Optional Improvements**
   - Evaluate testify usage for potential simplification
   - Consider map-based tables for complex test suites

## Files Requiring Most Attention

1. `lib/plan/analyzer_test.go` - Core test file, needs helper marking and naming updates
2. `lib/plan/formatter_test.go` - Multiple helper-like functions need `t.Helper()`
3. `lib/plan/integration_test.go` - Heavy defer usage, convert to `t.Cleanup()`
4. `lib/plan/performance_test.go` - Good candidate for parallel execution

## Estimated Scope

- **Files to modify**: ~20-25 test files
- **Lines affected**: ~500-800 (mostly mechanical changes)
- **Risk level**: Low (test-only changes)
- **Testing approach**: Run `make test` after each phase to ensure no breakage