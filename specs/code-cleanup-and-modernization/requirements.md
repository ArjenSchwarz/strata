# Code Cleanup and Modernization Requirements

## Introduction

This feature encompasses a comprehensive cleanup and modernization effort for the Strata codebase, focusing on three main areas: applying automated modernization suggestions, upgrading to the latest Go version (1.25.0), and improving test suite compliance with Go testing best practices. The goal is to improve code maintainability, leverage modern Go features, and ensure the test suite follows established guidelines for clarity and maintainability.

## Requirements

### 1. Automated Code Modernization

**User Story:** As a developer, I want to apply automated modernization suggestions to the codebase, so that I can leverage modern Go language features and improve code readability.

**Acceptance Criteria:**
1.1. The system SHALL run `modernize -fix -test ./...` to apply all safe automated fixes across the entire codebase
1.2. The system SHALL replace if/else conditional assignments with built-in min/max functions (identified in analyzer.go lines 344, 472, 546, 1566)
1.3. The system SHALL replace interface{} with 'any' type where appropriate (efaceany category)
1.4. The system SHALL replace manual loops with slices.Contains where appropriate (slicescontains category)
1.5. The system SHALL replace sort.Slice with slices.Sort where applicable (sortslice category)
1.6. The system SHALL replace for loops with range constructs where appropriate (rangeint category)
1.7. The system SHALL replace HasPrefix/TrimPrefix patterns with CutPrefix where applicable (stringscutprefix category)
1.8. The system SHALL replace other modernization opportunities identified by the modernize tool
1.9. The system SHALL preserve all existing functionality and behavior after modernization
1.10. The system SHALL ensure all tests pass after modernization changes

### 2. Go Version Upgrade

**User Story:** As a developer, I want to upgrade the project to Go 1.25.0, so that I can access the latest language features and performance improvements.

**Acceptance Criteria:**
2.1. The system SHALL update the go.mod file to specify Go version 1.25.0
2.2. The system SHALL run `go mod tidy` to ensure dependency compatibility
2.3. The system SHALL verify all dependencies are compatible with Go 1.25.0
2.4. The system SHALL ensure the project builds successfully with Go 1.25.0
2.5. The system SHALL ensure all tests pass with Go 1.25.0
2.6. The system SHALL update any CI/CD configurations to use Go 1.25.0
2.7. The system SHALL apply any Go 1.25-specific modernizations available in the modernize tool
2.8. The system SHALL leverage performance improvements and bug fixes available in Go 1.25.0

### 3. Test Helper Function Improvements

**User Story:** As a developer, I want test helper functions to be properly marked, so that test failures report the correct line numbers and improve debugging.

**Acceptance Criteria:**
3.1. The system SHALL identify all test helper functions across the codebase
3.2. The system SHALL add `t.Helper()` calls to all identified test helper functions
3.3. The system SHALL ensure t.Helper() is called as the first statement in each helper function
3.4. The system SHALL verify that test failure messages correctly point to the calling test after marking helpers
3.5. The system SHALL update at least the following known helper functions: testFormatterSortingBackwardCompatibility, testPropertySortingBackwardCompatibility

### 4. Test Cleanup Method Migration

**User Story:** As a developer, I want to use t.Cleanup() for test cleanup operations, so that cleanup is guaranteed to run even if tests panic.

**Acceptance Criteria:**
4.1. The system SHALL replace all instances of `defer os.Remove()` with `t.Cleanup()` in test files
4.2. The system SHALL replace all instances of `defer os.RemoveAll()` with `t.Cleanup()` in test files
4.3. The system SHALL ensure cleanup functions are registered with t.Cleanup() in the correct order
4.4. The system SHALL preserve the original cleanup logic within the t.Cleanup() callbacks
4.5. The system SHALL update approximately 20+ identified instances across the test suite

### 5. Test Variable Naming Standardization

**User Story:** As a developer, I want consistent variable naming in tests, so that test code is more readable and follows Go conventions.

**Acceptance Criteria:**
5.1. The system SHALL rename all "expected" variables to "want" in test files
5.2. The system SHALL rename all "result" and "actual" variables to "got" in test files
5.3. The system SHALL update assertion messages to use the new variable names
5.4. The system SHALL ensure naming changes are consistent across all test files
5.5. The system SHALL preserve test functionality after renaming
5.6. The system SHALL use "tc" as the variable name for test cases in table-driven tests

### 6. Parallel Test Implementation

**User Story:** As a developer, I want suitable tests to run in parallel, so that test execution time is reduced.

**Acceptance Criteria:**
6.1. The system SHALL identify unit tests that are safe to run in parallel (no shared resources)
6.2. The system SHALL add `t.Parallel()` as the first statement in identified parallel-safe tests
6.3. The system SHALL ensure proper variable capture in parallel subtests using loop variables
6.4. The system SHALL NOT parallelize integration tests that may have dependencies
6.5. The system SHALL verify test stability after enabling parallel execution
6.6. The system SHALL document any tests that cannot be parallelized and the reasons why

### 7. Test Organization and Structure

**User Story:** As a developer, I want to improve test organization and structure using available tooling, so that tests are easier to maintain and understand.

**Acceptance Criteria:**
7.1. The system SHALL identify test files that exceed 800 lines and split them by functionality
7.2. The system SHALL use the move_code_section.py script for reorganizing test functions when splitting files
7.3. The system SHALL split large test files into logical groupings (e.g., handler_auth_test.go, handler_validation_test.go)
7.4. The system SHALL target 500-800 lines as the optimal test file length range
7.5. The system SHALL ensure each test case has descriptive names for better debugging
7.6. The system SHALL preserve all existing test cases during any reorganization
7.7. The system SHALL maintain test coverage after structural changes
7.8. The system SHALL improve test readability through better organization
7.9. The system SHALL ensure all subtests use t.Run() for proper test execution
7.10. The system SHALL consider map-based test tables for tests with many cases to ensure unique names

### 8. Testify Dependency Reduction

**User Story:** As a developer, I want to minimize external testing dependencies, so that the test suite is simpler and uses standard library where possible.

**Acceptance Criteria:**
8.1. The system SHALL identify simple equality assertions using testify
8.2. The system SHALL replace simple assert.Equal with standard library comparisons where appropriate
8.3. The system SHALL replace simple require.NoError with standard if err != nil checks
8.4. The system SHALL retain testify for complex assertions that would be verbose with standard library
8.5. The system SHALL ensure test readability is maintained or improved after changes
8.6. The system SHALL document cases where testify is retained and the rationale
8.7. The system SHALL consider using github.com/google/go-cmp/cmp for complex comparisons where testify was removed

### 9. Test Coverage Standards

**User Story:** As a developer, I want to maintain appropriate test coverage levels, so that the codebase remains well-tested without coverage gaming.

**Acceptance Criteria:**
9.1. The system SHALL maintain 70-80% test coverage as the target (85% for critical components)
9.2. The system SHALL identify and remove any assertion-free tests that exist solely for coverage
9.3. The system SHALL ensure all tests contain meaningful assertions
9.4. The system SHALL document current coverage levels before and after changes
9.5. The system SHALL leverage go build -cover for integration test coverage measurement where applicable

### 10. Test Fixtures and Data Management

**User Story:** As a developer, I want test fixtures to be properly organized, so that test data is maintainable and follows Go conventions.

**Acceptance Criteria:**
10.1. The system SHALL ensure test data files are stored in testdata directories
10.2. The system SHALL implement golden file testing for complex output validation where beneficial
10.3. The system SHALL use functional builder patterns for test data creation where appropriate
10.4. The system SHALL ensure testdata directories are properly ignored by Go toolchain

### 11. Integration Test Organization

**User Story:** As a developer, I want integration tests to be properly separated, so that they can be run independently from unit tests.

**Acceptance Criteria:**
11.1. The system SHALL use environment variables (e.g., INTEGRATION) for test separation instead of build tags
11.2. The system SHALL ensure integration tests skip by default when INTEGRATION env var is not set
11.3. The system SHALL properly categorize existing tests as unit or integration tests
11.4. The system SHALL document how to run integration tests separately

### 12. Benchmark Improvements

**User Story:** As a developer, I want benchmarks to follow modern Go patterns, so that performance testing is accurate and meaningful.

**Acceptance Criteria:**
12.1. The system SHALL identify benchmarks using old patterns (for i := 0; i < b.N; i++)
12.2. The system SHALL consider migrating to B.Loop() pattern where available in Go 1.25
12.3. The system SHALL ensure all benchmark runs include memory profiling (-benchmem)
12.4. The system SHALL document how to use benchstat for statistical analysis
12.5. The system SHALL verify existing benchmarks continue to work after changes

### 13. Build and Test Validation

**User Story:** As a developer, I want all changes to be validated through the build and test pipeline, so that no regressions are introduced.

**Acceptance Criteria:**
13.1. The system SHALL run `make fmt` and ensure code formatting is correct
13.2. The system SHALL run `make vet` and ensure no static analysis issues
13.3. The system SHALL run `make lint` and ensure no linting issues
13.4. The system SHALL run `make test` and ensure all tests pass
13.5. The system SHALL run `make build` and ensure the project builds successfully
13.6. The system SHALL run `make test-action` and ensure GitHub Action tests pass
13.7. The system SHALL maintain existing test coverage metrics  
13.8. The system SHALL verify performance with existing benchmark tests
13.9. The system SHALL run tests with race detector enabled (go test -race ./...)
13.10. The system SHALL ensure all tests pass after each phase of implementation

## Implementation Phases

The implementation should follow this phased approach to minimize risk:

**Phase 1: Automated Modernization & Go Upgrade**
- Apply modernize tool fixes
- Upgrade to Go 1.25.0
- Validate build and tests

**Phase 2: Test Helper & Cleanup**
- Mark test helpers with t.Helper()
- Migrate to t.Cleanup()
- Validate test functionality

**Phase 3: Test Standardization**
- Standardize variable naming (got/want, tc for test cases)
- Ensure t.Run() usage for subtests
- Update assertion messages
- Maintain test clarity

**Phase 4: Test Organization**
- Split large test files (>800 lines)
- Organize test fixtures in testdata
- Separate integration tests
- Document test categories

**Phase 5: Test Performance**
- Enable parallel tests
- Update benchmarks to modern patterns
- Optimize test execution time
- Document parallelization decisions

**Phase 6: Optional Improvements**
- Evaluate testify usage
- Consider map-based test tables
- Apply golden file testing where beneficial
- Apply final optimizations

## Success Criteria

- All automated modernization suggestions are successfully applied
- Project successfully builds and runs on Go 1.25.0
- All tests pass with the same or better coverage (70-80% target, 85% for critical)
- Test execution time is reduced through parallelization
- Code follows modern Go idioms and best practices
- Test suite complies with all Go unit testing guidelines
- No functionality is broken or behavior changed
- Race detector finds no issues
- All 38 Go unit testing rules are properly addressed
- Changes are reviewed and approved through standard code review process

## Testing Guidelines Compliance

This feature ensures compliance with all 38 Go unit testing guidelines:
- Test organization (Rules 1-5): File placement, splitting, and structure
- Table-driven testing (Rules 6-8): Map-based tables, descriptive names, t.Run() usage
- Test coverage (Rules 9-11): 70-80% target, meaningful tests, integration coverage
- Parallel testing (Rules 12-14): t.Parallel(), variable capture, synctest consideration
- Mocking (Rules 15-17): Prefer integration, function-based doubles, Testcontainers
- Test fixtures (Rules 18-20): testdata directory, golden files, builder patterns
- Integration testing (Rules 21-22): Environment variable separation, skip by default
- Benchmarking (Rules 23-25): B.Loop(), benchmem, benchstat
- Test helpers (Rules 26-28): t.Helper(), t.Cleanup(), got/want naming
- Assertions (Rules 29-30): cmp.Diff, testify for complex cases
- Naming (Rules 31-32): Go conventions, tc for test cases
- Code quality (Rules 33-35): fmt, test execution, linting
- Modern patterns (Rules 36-38): Interfaces, dependency injection, behavior testing