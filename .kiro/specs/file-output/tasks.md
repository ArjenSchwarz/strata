# Implementation Plan

- [x] 1. Add persistent file output flags to root command
  - Add `--file` and `--file-format` persistent flags to root command in `cmd/root.go`
  - Bind flags to Viper configuration system
  - _Requirements: 1.1, 1.2_

- [x] 2. Implement placeholder resolution system in config
  - Create `resolvePlaceholders()` method in `config/config.go`
  - Support timestamp, region, account ID, and stack name placeholders
  - Add helper methods for context extraction (region, account ID)
  - _Requirements: 4.1, 4.2_

- [x] 3. Create file validation system
  - Create new `config/validation.go` file with validation functions
  - Implement `sanitizeFilePath()` function for path traversal protection
  - Implement `validateWritePermissions()` function for directory access checks
  - Implement `validateFormatSupport()` function for format validation
  - _Requirements: 6.1, 6.2, 16.1, 16.2_

- [x] 4. Extend OutputSettings in config system
  - Modify `NewOutputSettings()` method in `config/config.go` to handle file output
  - Add placeholder resolution to file path processing
  - Implement format defaulting logic (file format defaults to stdout format)
  - _Requirements: 3.1, 3.2, 3.3_

- [x] 5. Enhance go-output library integration for dual output
  - Modify `OutputArray.Write()` calls in `lib/plan/formatter.go` to support file output
  - Ensure all formatter methods (`formatPlanInfo`, `formatStatisticsSummary`, `formatResourceChangesTable`) use dual output
  - Add error handling for file write failures with graceful degradation
  - _Requirements: 9.1, 14.1, 10.1_

- [x] 6. Add file output validation to plan summary command
  - Integrate file validation into `runPlanSummary()` function in `cmd/plan_summary.go`
  - Add validation call before executing formatter
  - Implement error handling for validation failures
  - _Requirements: 10.1_

- [x] 7. Create unit tests for validation components
  - Create `config/validation_test.go` with tests for path safety validation
  - Test placeholder resolution functionality
  - Test format validation logic
  - Test permission validation
  - _Requirements: 15.1, 15.2_

- [ ] 8. Create integration tests for file output functionality
  - Create test cases in `lib/plan/formatter_test.go` for dual output
  - Test different format combinations (stdout vs file format)
  - Test file creation and content validation
  - Test error scenarios (permission denied, invalid paths)
  - _Requirements: 15.2, 15.3, 15.4_

- [ ] 9. Add security tests for path traversal prevention
  - Create security test cases in `config/validation_test.go`
  - Test malicious path inputs (../, absolute paths, etc.)
  - Test file overwrite scenarios
  - _Requirements: 15.4, 16.1, 16.2, 16.3_

- [x] 10. Update command help and documentation
  - Update flag descriptions in `cmd/plan_summary.go` to include file output flags
  - Add placeholder usage examples to command help text
  - Update README.md with file output usage examples
  - _Requirements: 12.1, 13.1_
- [ ] 11. Add format-specific enhancement tests
  - Create tests for format-specific separators in different output formats
  - Test enhanced markdown/HTML output features
  - Test format conversion between stdout and file formats
  - _Requirements: 8.1, 8.2_

- [ ] 12. Add performance tests for large datasets
  - Create benchmark tests for file output with large datasets
  - Test memory usage during dual output operations
  - Test file I/O performance with different formats
  - _Requirements: 15.1_

- [ ] 13. Enhance error handling and recovery
  - Implement graceful degradation for file write failures
  - Add structured error reporting with warnings and info messages
  - Test error recovery scenarios (fallback to stdout-only mode)
  - _Requirements: 10.1, 10.2_

- [ ] 14. Add configuration defaults implementation
  - Implement `setConfigDefaults()` function with all required defaults
  - Add default file extensions for different file types
  - Set sensible output format and table styling defaults
  - _Requirements: 11.1_

- [ ] 15. Update placeholder naming to match requirements
  - Remove `$STACKNAME` placeholder from `resolvePlaceholders()` method
  - Rename `$REGION` to `$AWS_REGION` in placeholder resolution
  - Rename `$ACCOUNTID` to `$AWS_ACCOUNTID` in placeholder resolution
  - Update helper method names accordingly (`getRegion()` to `getAWSRegion()`, etc.)
  - Update tests and documentation to reflect new placeholder names
  - _Requirements: 4.1, 4.2_