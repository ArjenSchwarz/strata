# Simplified Plan Rendering Requirements

## Introduction

The current plan summary rendering architecture in Strata has become overly complex, with multiple format-specific code paths, mixed rendering approaches, and heavy data transformation layers. This complexity prevents multiple tables from rendering correctly in markdown format and makes the codebase difficult to maintain.

**Problem Statement**: When rendering plan summaries in markdown format, only the Resource Changes table appears. The Plan Information and Summary Statistics tables are missing due to architectural issues in the rendering pipeline. Previous attempts to fix this bug within the current architecture have failed because the complexity itself is the root cause.

**Solution Approach**: This feature simplifies the rendering architecture by following the proven pattern from go-output v2's collapsible-tables example (`/go-output-claude/v2/examples/collapsible-tables/main.go`), which successfully renders multiple tables with collapsible content in all formats including markdown.

## Requirements

### 1. Unified Table Creation Pattern
**User Story**: As a developer, I want a single consistent method for creating tables, so that the rendering behavior is predictable and maintainable.

**Acceptance Criteria**:
1.1. The system SHALL use `output.NewTableContent()` exclusively for all table creation.
1.2. The system SHALL NOT use `builder.Table()` method anywhere in the formatter code.
1.3. The system SHALL create tables with consistent schema definitions using `output.WithSchema()`.
1.4. The system SHALL handle table creation errors by logging the error and continuing with other tables (no fallback to alternative methods).

### 2. Simplified Renderer Configuration
**User Story**: As a developer, I want format handling delegated to go-output, so that Strata code remains format-agnostic.

**Acceptance Criteria**:
2.1. The system SHALL prepare data once and let go-output handle format-specific rendering.
2.2. The system SHALL eliminate the distinction between `getFormatFromConfig()` and `getCollapsibleFormatFromConfig()`.
2.3. The system SHALL use go-output's standard format definitions (output.Markdown, output.Table, etc.).
2.4. The system SHALL NOT contain format-specific branching logic (e.g., "if format == markdown").
2.5. The system SHALL handle format capabilities through go-output's metadata rather than custom logic.

### 3. Simplified Document Building
**User Story**: As a user, I want to see all plan information tables rendered correctly, so that I have complete visibility into my infrastructure changes.

**Acceptance Criteria**:
3.1. The system SHALL use the pattern `output.New().AddContent().AddContent().Build()` for document construction.
3.2. The system SHALL render Plan Information, Summary Statistics, and Resource Changes tables in all output formats.
3.3. The system SHALL support multiple tables in markdown format without rendering issues.
3.4. The system SHALL maintain the same table order across all output formats.

### 4. Streamlined Data Transformation
**User Story**: As a developer, I want minimal data transformation layers, so that the code is easier to understand and debug.

**Acceptance Criteria**:
4.1. The system SHALL prepare table data directly without intermediate transformation objects.
4.2. The system SHALL create `output.NewCollapsibleValue()` objects during initial data preparation.
4.3. The system SHALL eliminate duplicate formatter functions that perform the same transformations.
4.4. The system SHALL move complex data preparation logic from the formatter to the analyzer component.

### 5. Consistent Format Handling
**User Story**: As a developer, I want format-agnostic code paths, so that adding new formats or modifying existing ones is straightforward.

**Acceptance Criteria**:
5.1. The system SHALL eliminate format-specific branching in the main rendering logic.
5.2. The system SHALL handle all formats through a single code path with format-specific configuration.
5.3. The system SHALL NOT have separate logic branches for markdown vs non-markdown formats.
5.4. The system SHALL apply transformers (emoji, color, sorting) consistently across all formats.

### 6. Simplified Provider Grouping
**User Story**: As a user, I want provider grouping to work seamlessly when enabled, so that large plans are easier to review.

**Acceptance Criteria**:
6.1. When provider grouping is disabled, the system SHALL render all resources in a single table.
6.2. When provider grouping is enabled and threshold is met, the system SHALL group resources by provider.
6.3. The system SHALL use consistent section creation for provider groups across all formats.
6.4. The system SHALL maintain existing auto-expansion behavior for high-risk changes within groups.
6.5. The system SHALL maintain the current threshold check behavior (only group when resource count >= threshold).

### 7. Preserved Functionality
**User Story**: As a user, I want all existing features to continue working after the simplification, so that my workflows are not disrupted.

**Acceptance Criteria**:
7.1. The system SHALL maintain all existing collapsible/expandable content functionality through proper use of `output.NewCollapsibleValue()` during data preparation.
7.2. The system SHALL preserve danger highlighting and auto-expansion for sensitive resources.
7.3. The system SHALL continue to support the `--expand-all` flag and configuration option.
7.4. The system SHALL maintain backward compatibility with existing configuration files.
7.5. The system SHALL preserve all existing output formats (table, json, markdown, html, csv).
7.6. The system SHALL continue to support nested collapsible sections (e.g., provider groups containing expandable property changes).

### 8. Code Quality and Maintainability
**User Story**: As a developer, I want clean, well-structured code, so that future modifications are easy to implement.

**Acceptance Criteria**:
8.1. The system SHALL reduce code duplication by eliminating near-identical functions (e.g., propertyChangesFormatter vs propertyChangesFormatterDirect).
8.2. The system SHALL follow the specific pattern from `../go-output-claude/v2/examples/collapsible-tables/main.go`.
8.3. The system SHALL consolidate the multiple data transformation methods into a single approach.
8.4. The system SHALL include comments for: architectural decisions, non-obvious go-output API usage, and data structure transformations.
8.5. The system SHALL maintain or improve existing test coverage.
8.6. The system SHALL NOT require migration code (clean implementation replacement).

### 9. Error Handling and Edge Cases
**User Story**: As a user, I want the system to handle edge cases gracefully, so that rendering never fails unexpectedly.

**Acceptance Criteria**:
9.1. The system SHALL display a meaningful message when no resource changes exist.
9.2. The system SHALL handle empty or nil plan summaries without crashing.
9.3. The system SHALL log errors from table creation but continue rendering other content.
9.4. The system SHALL handle resources with missing or malformed data by showing available information.
9.5. The system SHALL properly escape special characters in resource names and values.

### 10. Testing Requirements
**User Story**: As a developer, I want comprehensive tests that validate the multi-table rendering fix, so that we can prevent regression.

**Acceptance Criteria**:
10.1. The system SHALL include tests that verify all three tables (Plan Information, Summary Statistics, Resource Changes) render in markdown format.
10.2. The system SHALL include tests for each supported output format (table, json, markdown, html, csv).
10.3. The system SHALL include tests for collapsible content functionality in supported formats.
10.4. The system SHALL include tests for provider grouping with various resource counts and thresholds.
10.5. The system SHALL include tests for edge cases (empty plans, nil data, special characters).
10.6. The system SHALL include benchmark tests to ensure no performance regression.

## Out of Scope

- Changes to the underlying plan analysis logic
- Modifications to the go-output v2 library
- New output formats beyond those currently supported
- Changes to the CLI command structure
- Modifications to configuration file format (only behavior changes)
- Migration strategy or backward compatibility code
- Format-specific behavior handling (delegated to go-output)
- Moving ActionSortTransformer logic to data preparation phase
- Changes to provider extraction logic

## Success Criteria

1. All three main tables (Plan Information, Summary Statistics, Resource Changes) render correctly in markdown format
2. Code duplication is eliminated (no more duplicate formatter functions)
3. All existing tests pass without modification
4. No user-visible behavior changes except for the fix of multi-table rendering bug
5. Developer documentation clearly explains the new simplified architecture

## Implementation Approach

The implementation will follow the exact pattern demonstrated in the go-output v2 collapsible-tables example:

1. **Data Preparation Phase**:
   - Create table data with `output.NewCollapsibleValue()` objects where needed
   - Apply consistent schema definitions using `output.WithSchema()`
   - Handle all data transformations in one place

2. **Document Building Phase**:
   - Use pattern: `output.New().AddContent(table1).AddContent(table2).Build()`
   - Create tables with `output.NewTableContent()` exclusively
   - Let go-output handle all format-specific rendering

3. **Rendering Phase**:
   - Create single output configuration
   - Apply transformers consistently
   - Let go-output manage format differences

This approach eliminates the current complexity while maintaining all functionality.