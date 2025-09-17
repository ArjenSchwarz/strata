---
references:
    - specs/data-pipeline/requirements.md
    - specs/data-pipeline/design.md
    - specs/data-pipeline/decision_log.md
    - docs/research/go-output/v2/API.md
---
# Data Pipeline Integration Tasks

- [x] 1. Create unit tests for sorting functionality
  - Write unit test for sortResourceTableData function testing danger sorting
  - Write unit test for getActionPriority function with all action types
  - Write unit test for applyDecorations function verifying emoji application and field cleanup
  - Test edge cases: empty data, missing fields, null values
  - References: Requirements 1.3, 4.1, 4.2

- [x] 2. Implement sorting helper functions
  - Create getActionPriority function in lib/plan/formatter.go that maps action types to numeric priorities
  - Create sortResourceTableData function that sorts by danger status, action priority, then alphabetically
  - Create applyDecorations function that adds emoji indicators and removes internal sorting fields
  - References: Requirements 1.2, 1.3

- [x] 3. Integrate sorting into data preparation flow
  - Modify prepareResourceTableData to store raw ActionType and IsDangerous fields in row data
  - Call sortResourceTableData after building tableData
  - Call applyDecorations after sorting to add emoji and clean up internal fields
  - Ensure collapsible content and provider grouping continue to work
  - References: Requirements 2.1, 2.4, 2.5

- [x] 4. Create integration tests for output verification
  - Write test comparing sorted output with current output for sample plans
  - Test sorting within provider groups when grouping is enabled
  - Verify all output formats produce identical results to current implementation
  - Run tests with all sample plan files in samples/ directory
  - References: Requirements 1.4, 2.3, 4.1, 4.3

- [x] 5. Remove ActionSortTransformer and clean up code
  - Delete ActionSortTransformer struct and all its methods from the codebase
  - Remove registration of ActionSortTransformer in the output pipeline initialization
  - Delete cached regex pattern variables used for string parsing
  - Remove hasDangerIndicator function if only used by ActionSortTransformer
  - Clean up any imports and dependencies no longer needed
  - References: Requirements 3.1, 3.2, 3.3, 3.4, 3.5

- [x] 6. Write benchmark tests
  - Create benchmark test for sortResourceTableData with various data sizes
  - Create benchmark comparing new implementation vs old string parsing approach
  - Document performance improvements in code comments
  - References: Requirements 4.4

- [ ] 7. Update tests for new implementation
  - Update any existing unit tests that depend on ActionSortTransformer
  - Verify all existing tests pass with the new implementation
  - Ensure test coverage remains high for the formatter module
  - References: Requirements 1.5, 1.6, 4.1
