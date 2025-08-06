# Output Refinements Feature Requirements

## Introduction

The Output Refinements feature addresses multiple quality and usability improvements for Strata's Terraform plan summary output. This feature enhances the user experience by fixing sorting issues, improving sensitive data handling, providing better configurability for output display, and ensuring consistent behavior across different output sections. The improvements focus on making the output more predictable, secure, and customizable while maintaining clarity and usefulness for infrastructure change reviews.

## Requirements

### 1. Property Sorting Improvements

**User Story:** As a DevOps engineer, I want properties in the property changes column to be sorted alphabetically, so that I can quickly locate specific properties when reviewing changes.

**Acceptance Criteria:**
1. **When** the plan summary displays property changes within each individual resource, **then** properties SHALL be sorted alphabetically by property name within that resource's property changes section only.
2. **When** multiple properties have the same name but different paths, **then** the system SHALL sort them first by name, then by path hierarchy.
3. **When** displaying property changes, **then** the alphabetical sorting SHALL be case-insensitive to ensure consistent ordering.
4. **When** properties have special characters or numbers, **then** the system SHALL follow natural sort ordering (e.g., prop1, prop2, prop10, not prop1, prop10, prop2).

### 2. Restore Sorting by Sensitivity and Action

**User Story:** As a security-conscious developer, I want changes to be sorted by sensitivity first and action second, so that I can prioritize reviewing the most critical changes.

**Acceptance Criteria:**
1. **When** displaying resource changes in the summary, **then** the system SHALL sort resources with the following priority: first by dangerous resources (IsDangerous=true), then by resources with sensitive properties, then by all other resources.
2. **When** resources have the same sensitivity level, **then** the system SHALL sort them by action type in the order: delete, replace, update, create.
3. **When** resources have the same sensitivity and action type, **then** the system SHALL sort them alphabetically by resource address.
4. **When** provider grouping is enabled, **then** the sorting SHALL be applied within each provider group independently.

### 3. Configurable No-Op Display

**User Story:** As a power user, I want the option to show or hide no-op resources in the plan summary, so that I can choose between a comprehensive view or a focused view of actual changes.

**Acceptance Criteria:**
1. **When** running the plan summary command, **then** the system SHALL provide a `--show-no-ops` flag to include resources with no changes.
2. **When** the `--show-no-ops` flag is not provided, **then** the system SHALL hide all resources marked as no-op by default.
3. **When** configuring via strata.yaml, **then** the system SHALL support a `plan.show-no-ops` configuration option (using kebab-case for consistency) with a default value of false.
4. **When** both CLI flag and configuration file specify no-op visibility, **then** the CLI flag SHALL take precedence over the configuration file setting.
5. **When** no-ops are hidden and there are no actual changes, **then** the system SHALL display a message indicating "No changes detected" rather than an empty table.
6. **When** `--show-no-ops` is enabled, **then** it SHALL work independently of the `--details` flag, allowing users to see no-ops with or without detailed property changes.
7. **When** no-ops are shown or hidden, **then** the statistics summary SHALL remain unchanged, continuing to count all resources including no-ops as it currently does.

### 4. Hide No-Op Output Changes

**User Story:** As a Terraform user, I want output changes with no actual modifications to be hidden from the summary, so that I can focus on outputs that are actually changing.

**Acceptance Criteria:**
1. **When** processing output changes, **then** the system SHALL identify outputs where before and after values are identical.
2. **When** an output has no actual change (no-op), **then** the system SHALL exclude it from the output changes table by default.
3. **When** all outputs are no-ops, **then** the system SHALL hide the entire output changes section rather than showing an empty or no-op filled table.
4. **When** the `--show-no-ops` flag is enabled, **then** the system SHALL include no-op outputs in the output changes table.
5. **When** displaying the statistics, **then** no-op outputs SHALL NOT be counted in the output change counts.

### 5. Mask Sensitive Values in Property Overview

**User Story:** As a security administrator, I want sensitive values to be masked in all property overviews, so that sensitive information is not exposed in logs or screenshots.

**Acceptance Criteria:**
1. **When** displaying property values marked as sensitive in the Terraform plan, **then** the system SHALL replace the actual value following Terraform's JSON output convention for sensitive values.
2. **When** a property is marked as sensitive, **then** both the "before" and "after" values SHALL be masked in the property changes display.
3. **When** formatting output in JSON format, **then** sensitive values SHALL be handled identically to how `terraform show -json` handles them (maintaining type information where possible).
4. **When** formatting output in table, HTML, or Markdown formats, **then** sensitive values SHALL be replaced with "(sensitive value)" text.
5. **When** a sensitive property is nested within a complex structure, **then** only the sensitive leaf values SHALL be masked while preserving the structure visibility.
6. **When** displaying sensitive values, **then** the system SHALL still indicate the type of change (created, updated, deleted) without revealing the actual values.

## Edge Cases and Considerations

### Sorting Edge Cases
- Resources with identical addresses but different types should maintain consistent ordering
- When sensitivity information is not available, resources should be treated as non-sensitive for sorting purposes
- Empty property names or paths should sort to the beginning or end consistently

### Configuration Precedence
- CLI flags always override configuration file settings
- Environment variables (if implemented) should have a defined precedence order
- Invalid configuration values should fall back to safe defaults with appropriate warnings

### Sensitive Data Handling
- Partially sensitive structures (e.g., objects with both sensitive and non-sensitive fields) need careful handling
- Sensitive data masking should be applied before any logging or output generation
- Consider providing a debug mode that warns when it would expose sensitive data

### Performance Considerations
- Sorting large numbers of resources should remain performant
- Property sorting within resources should not significantly impact rendering time
- No-op filtering should be efficient even for plans with thousands of resources

## Success Criteria

1. All open issues (#17, #18, #19, #20, #21) are resolved
2. Unit tests achieve >90% coverage for all new sorting and filtering logic
3. Sensitive values are never exposed in any output format
4. Performance regression tests show <5% impact on plan summary generation time
5. Documentation is updated to reflect new CLI flags and configuration options
6. Existing users experience no breaking changes (backward compatibility maintained)

## Technical Constraints

1. Must maintain compatibility with existing output formats (table, JSON, HTML, Markdown)
2. Must work with Terraform plan format versions currently supported
3. Configuration changes must be backward compatible with existing strata.yaml files
4. Should follow existing code patterns and architecture in the strata codebase
5. Must integrate with existing danger highlighting and progressive disclosure features

## User Experience Requirements

1. Default behavior should be the most secure and focused (hide no-ops, mask sensitive data)
2. Error messages for configuration issues should be clear and actionable
3. The sorting order should be predictable and documented
4. CLI help text should clearly explain the new flags and their effects
5. Changes should be highlighted in release notes with migration guidance if needed