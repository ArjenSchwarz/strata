# Strata GitHub Action - Integration Test Suite

This directory contains comprehensive integration tests for the Strata GitHub Action, implementing all requirements from Task 8 of the GitHub Action Simplification specification.

## Test Suite Overview

The integration test suite consists of four main components:

### 1. Master Test Runner
- **File**: `run_integration_test_suite.sh`
- **Purpose**: Orchestrates all integration tests and provides consolidated reporting
- **Usage**: `./test/run_integration_test_suite.sh`

### 2. End-to-End Integration Tests (Task 8.1)
- **File**: `test_comprehensive_integration.sh`
- **Purpose**: Tests full action execution with all sample plan files
- **Coverage**:
  - Test harness for full action execution
  - Tests with each sample plan file in `samples/`
  - All output formats (markdown, json, table, html)
  - Dangerous changes detection and reporting
  - Performance measurement (<30 seconds total execution)

### 3. Backwards Compatibility Tests (Task 8.2)
- **File**: `test_backwards_compatibility.sh`
- **Purpose**: Ensures existing workflows continue working without modification
- **Coverage**:
  - Minimal inputs (just plan-file) compatibility
  - All current input parameters work correctly
  - All outputs generated with same names
  - Existing workflows continue without changes
  - action.yml interface unchanged (except strata-version addition)
  - Terraform plan file format support maintained

### 4. Performance Benchmarks (Task 8.3)
- **File**: `test_performance_benchmarks.sh`
- **Purpose**: Measures and validates performance against requirements
- **Coverage**:
  - Binary download time (<10 seconds)
  - Analysis startup time (<5 seconds)
  - Total execution time (<30 seconds)
  - Performance baseline comparison
  - Performance improvement documentation

## Running the Tests

### Run All Tests (Recommended)
```bash
# Run the complete integration test suite
./test/run_integration_test_suite.sh
```

### Run Individual Test Suites
```bash
# End-to-end integration tests only
./test/test_comprehensive_integration.sh

# Backwards compatibility tests only
./test/test_backwards_compatibility.sh

# Performance benchmarks only
./test/test_performance_benchmarks.sh
```

## Test Requirements

### Prerequisites
- Strata repository root directory
- Sample plan files in `samples/` directory
- Executable `action.sh` and valid `action.yml`
- Basic Unix tools (mktemp, jq, etc.)

### Environment
- Tests create temporary directories for isolation
- No network dependencies (uses local sample files)
- Tests clean up automatically on success
- Preserves artifacts on failure for debugging

## Test Output

### Success Indicators
- All tests return exit code 0 on success
- Green checkmarks for passed tests
- Performance metrics within thresholds
- Comprehensive test reports generated

### Failure Handling
- Clear error messages with context
- Test artifacts preserved for debugging
- Detailed logs for failed operations
- Performance metrics even on partial failure

## Test Reports

Each test suite generates detailed reports:

- **Master Report**: Consolidated view of all test results
- **Integration Report**: End-to-end functionality validation
- **Compatibility Report**: Backwards compatibility verification
- **Performance Report**: Performance benchmark results with metrics

Reports are saved in temporary directories (preserved on failure) and include:
- Test execution summary
- Detailed results for each test case
- Performance metrics and analysis
- Recommendations for improvements

## Integration with Development Workflow

### Local Development
```bash
# Quick validation during development
./test/run_integration_test_suite.sh

# Focus on specific aspects
./test/test_backwards_compatibility.sh  # Before API changes
./test/test_performance_benchmarks.sh   # After performance optimizations
```

### CI/CD Integration
The test suite is designed to integrate with continuous integration:

```bash
# Add to your CI pipeline
make test-action-integration  # If using Makefile
# or
./test/run_integration_test_suite.sh
```

### Makefile Integration
The tests integrate with the existing Makefile structure:
- `make test-action` - Includes integration tests
- `make test-action-integration` - Runs integration tests specifically

## Performance Benchmarking

The performance tests measure key metrics:

- **Binary Download**: Time to download and cache Strata binary
- **Analysis Startup**: Time from binary execution to analysis start
- **Total Execution**: Complete action execution time
- **Consistency**: Performance variance across multiple runs

### Performance Thresholds
- Binary download: ≤10 seconds
- Analysis startup: ≤5 seconds
- Total execution: ≤30 seconds

### Performance Monitoring
Results include baseline comparisons and recommendations for optimization.

## Extending the Test Suite

### Adding New Test Cases
1. Add test functions following the existing pattern
2. Use the established logging functions (`log_test`, `log_pass`, `log_fail`)
3. Create unique temporary files to avoid conflicts
4. Clean up environment variables after tests

### Testing New Features
1. Add feature-specific tests to appropriate suite
2. Update requirements documentation if needed
3. Consider performance impact for new features
4. Ensure backwards compatibility is maintained

## Troubleshooting

### Common Issues
- **Permission errors**: Ensure test scripts are executable (`chmod +x`)
- **Missing samples**: Verify `samples/` directory contains `.json` files
- **Timeout issues**: Check system load and network connectivity
- **Path issues**: Run tests from repository root directory

### Debug Mode
For detailed debugging, examine the preserved test artifacts in temporary directories when tests fail.

## Requirements Traceability

This test suite implements all requirements from the GitHub Action Simplification specification:

- **8.1**: End-to-end integration tests ✅
- **8.2**: Backwards compatibility tests ✅
- **8.3**: Performance benchmarks ✅
- **9.1-9.6**: Performance requirements validation ✅
- **10.1-10.6**: Backwards compatibility requirements validation ✅

The comprehensive nature of these tests ensures the GitHub Action maintains quality, performance, and compatibility standards throughout development and release cycles.