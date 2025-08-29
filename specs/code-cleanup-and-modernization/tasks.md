# Code Cleanup and Modernization - Implementation Tasks

## Overview
This document contains the implementation tasks for modernizing the Strata codebase through Go 1.25.0 upgrade, automated pattern modernization, and test suite improvements. Each task is designed to be executable by a coding agent with clear objectives and implementation steps.

## Implementation Tasks

### Phase 1: Foundation - Go Upgrade and Automated Modernization

- [x] **1. Upgrade Go version to 1.25.0**
  - Edit go.mod file to specify Go version 1.25.0
  - Run go mod tidy to update dependencies 
  - Verify all dependencies are compatible with Go 1.25.0
  - Update any CI/CD configuration files that specify Go version
  - References requirement 2.1, 2.2, 2.3, 2.6

- [x] **2. Apply automated modernization patterns**
  - Run modernize tool with command `modernize -fix -test ./...` to apply all safe automated fixes
  - Run gofmt and goimports to fix formatting and imports
  - References requirements 1.1-1.8

- [x] **3. Validate Go 1.25.0 compatibility and modernization**
  - Run make build to ensure successful compilation with Go 1.25.0
  - Run make test to verify all existing tests pass with new version and modernization changes
  - Run make test-action to ensure GitHub Action compatibility
  - Run benchmarks to verify no performance regressions
  - References requirements 1.9, 1.10, 2.4, 2.5, 2.8

### Phase 2: Test Organization and Helper Functions

- [x] **4. Split config/validation_test.go file**
  - Split validation_test.go (1,006 lines) into multiple files by validation category
  - Create validation_file_test.go for file validator tests (304 lines)
  - Create validation_config_test.go for configuration tests (121 lines)
  - Create validation_result_test.go for result and error tests (117 lines)
  - Create validation_security_test.go for security tests (490 lines)
  - Use move_code_section.py script to safely move test functions
  - References requirements 7.1, 7.2, 7.3, 7.5

- [x] **5. Split lib/plan/analyzer_test.go file**
  - Split analyzer_test.go (5,335 lines) into focused test files
  - Create analyzer_test.go for core resource analysis tests (1,029 lines)
  - Create analyzer_statistics_test.go for statistics calculation tests (319 lines)
  - Create analyzer_properties_test.go for property analysis tests (503 lines)
  - Create analyzer_objects_test.go for object comparison tests (326 lines)
  - Create analyzer_outputs_test.go for output processing tests (1,374 lines)
  - Create analyzer_utils_test.go for utility functions tests (1,833 lines)
  - Use move_code_section.py script to organize test functions by functionality
  - References requirements 7.1, 7.2, 7.3

- [x] **6. Split lib/plan/formatter_test.go file**
  - Split formatter_test.go (2,860 lines) into logical groupings
  - Create formatter_basic_test.go for basic formatting tests
  - Create formatter_enhanced_test.go for enhanced features tests
  - Create formatter_edge_cases_test.go for edge case tests
  - Use move_code_section.py script for safe reorganization
  - References requirements 7.1, 7.2, 7.3

- [x] **7. Mark test helper functions with t.Helper()**
  - Identify all test helper functions across the codebase using grep or AST analysis
  - Add t.Helper() as first statement to testFormatterSortingBackwardCompatibility function
  - Add t.Helper() as first statement to testPropertySortingBackwardCompatibility function
  - Add t.Helper() to all other identified helper functions
  - Verify test failure messages correctly point to calling tests
  - References requirements 3.1, 3.2, 3.3, 3.4, 3.5

### Phase 3: Test Cleanup and Naming Standardization

- [x] **8. Migrate defer cleanup to t.Cleanup() pattern**
  - Find all instances of defer os.Remove() in test files
  - Replace with t.Cleanup(func() { os.Remove(...) })
  - Find all instances of defer os.RemoveAll() in test files
  - Replace with t.Cleanup(func() { os.RemoveAll(...) })
  - Preserve original cleanup order and error handling
  - References requirements 4.1, 4.2, 4.3, 4.4

- [x] **9. Standardize test variable naming conventions**
  - Replace all "expected" variables with "want" in test files
  - Replace all "result" and "actual" variables with "got" in test files
  - Update assertion messages to use new variable names
  - Ensure "tc" is used for test case variables in table-driven tests
  - References requirements 5.1, 5.2, 5.3, 5.4, 5.6

- [x] **10. Ensure all subtests use t.Run()**
  - Identify table-driven tests not using t.Run()
  - Wrap each test case execution with t.Run(name, func(t *testing.T) {...})
  - Add descriptive names for each subtest
  - Consider converting slice-based test tables to map[string]struct for unique names
  - References requirements 7.9, 7.10

### Phase 4: Test Performance Optimization

- [x] **11. Implement parallel test execution**
  - Identify unit tests without shared resources or dependencies
  - Add t.Parallel() as first statement in identified tests
  - Capture loop variables properly in parallel subtests
  - Document tests that cannot be parallelized and reasons
  - References requirements 6.1, 6.2, 6.3, 6.6

- [x] **12. Organize test fixtures and data**
  - Ensure all test data files are in testdata directories
  - Implement golden file testing pattern for complex output validation where needed
  - Create functional builder patterns for test data creation where appropriate
  - Verify testdata directories are properly ignored by Go toolchain
  - References requirements 10.1, 10.2, 10.3, 10.4

- [x] **13. Separate integration tests from unit tests**
  - Identify tests that are actually integration tests
  - Add INTEGRATION environment variable checks to integration tests
  - Implement t.Skip() when INTEGRATION env var is not set
  - Document how to run integration tests separately in README
  - References requirements 11.1, 11.2, 11.3, 11.4

### Phase 5: Optional Improvements and Testify Reduction

- [ ] **14. Reduce testify dependency usage**
  - Find simple assert.Equal calls that can use standard library
  - Replace with standard if statements and t.Error/t.Fatal
  - Find simple require.NoError calls
  - Replace with if err != nil { t.Fatal(err) }
  - Keep testify for complex object comparisons
  - Document cases where testify is retained
  - References requirements 8.1, 8.2, 8.3, 8.6

- [ ] **15. Update benchmarks to modern patterns**
  - Identify benchmarks using old for loop patterns
  - Consider migrating to B.Loop() pattern if available in Go 1.25
  - Ensure all benchmark runs include -benchmem flag
  - Document benchstat usage for statistical analysis
  - References requirements 12.1, 12.2, 12.3, 12.4

### Phase 6: Final Validation

- [ ] **16. Run comprehensive validation suite**
  - Run make fmt to ensure code formatting
  - Run make vet for static analysis
  - Run make lint for code quality
  - Run make test to verify all tests pass
  - Run make build to confirm compilation
  - Run make test-action for GitHub Action tests
  - Run go test -race ./... for race condition detection
  - References requirements 13.1-13.9

- [ ] **17. Document test coverage and performance metrics**
  - Run go test -cover ./... to measure test coverage
  - Ensure coverage maintains 70-80% target (85% for critical components)
  - Run benchmarks to compare performance before and after changes
  - Document test execution time improvements from parallelization
  - Create summary of all changes made
  - References requirements 9.1, 9.4, 13.7, 13.8

## Success Criteria

All tasks must be completed with:
- All existing tests passing
- Test coverage maintained or improved (70-80% target)
- No functionality broken or behavior changed  
- All validation commands passing successfully
- Race detector finding no issues
- Performance benchmarks showing no regressions