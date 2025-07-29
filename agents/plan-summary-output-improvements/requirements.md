# Plan Summary Output Improvements

## Introduction

This feature enhances the plan summary output to provide cleaner, more readable, and more useful information to users. The improvements focus on removing unnecessary empty tables, improving property change formatting to match Terraform's native output style, preserving meaningful content in details and dependencies columns, verifying dependency functionality, and implementing custom sorting that prioritizes risky changes.

## Requirements

### 1. Empty Table Suppression

**User Story:** As a user, I want empty tables to be hidden from the output, so that I can focus on relevant information without visual clutter.

**Acceptance Criteria:**
1. When a Resource Changes table would only contain no-ops (which are filtered out), then the entire table including headers SHALL NOT be displayed
2. When a provider-specific table would only contain no-ops, then it SHALL NOT be displayed
3. When displaying provider-specific table counts, then only changed resources SHALL be counted (excluding no-ops)
4. When determining if provider grouping threshold is met, then the comparison SHALL use total changed resources, not total resources
5. When a section would only contain empty tables, then the section header SHALL also be suppressed

### 2. Property Changes Formatting

**User Story:** As a user, I want property changes to be displayed in a format similar to Terraform's native output, so that I can easily understand what's changing in a familiar format.

**Acceptance Criteria:**
1. When property changes are displayed, then they SHALL use Terraform's diff-style format with:
   - `~` prefix for modified properties
   - `+` prefix for added properties
   - `-` prefix for removed properties
2. When displaying property values, then the format SHALL show:
   - Property name only (no full resource path since we're already at resource level)
   - Old value -> New value for modifications
   - Proper indentation for nested properties with full expansion
   - Values displayed using Terraform's standard formatting conventions
3. When property changes contain sensitive values, then they SHALL be masked appropriately while maintaining readability
4. When counting property changes, then each individual property SHALL be counted separately (not grouped as a single change)
5. When implementing the formatter, then a custom CollapsibleFormatter SHALL be created using go-output's APIs
6. When parsing the plan, then property-level before/after values SHALL be extracted from the Terraform plan JSON's "change" objects

### 3. Property Changes Column Content

**User Story:** As a user, I want to see the actual property changes in the property_changes column, so that I can understand what's changing without emoji replacements.

**Acceptance Criteria:**
1. When property changes are displayed in the property_changes column, then the actual change details SHALL be shown using the Terraform diff format specified in requirement 2
2. When the column content is displayed, then NO emoji replacements SHALL occur
3. When content is too long for a column, then it SHALL be appropriately formatted or truncated while maintaining readability

### 4. ~~Dependency Functionality Verification~~ (Removed)

**Note:** After investigation, Terraform plan JSON does not include dependency information in the resource changes. The `depends_on` field referenced in the original requirements does not exist in the actual plan JSON structure. This requirement has been removed as it cannot be implemented with the available data.

### 5. Custom Risk-Based Sorting

**User Story:** As a user, I want resources sorted by risk level and action type, so that I can immediately see the most important changes at the top.

**Acceptance Criteria:**
1. When resources are sorted, then the primary sort order SHALL be:
   - All dangerous items first (any resource with danger flags set)
   - Non-dangerous items second
2. When resources have the same danger level, then the secondary sort order SHALL be by action type:
   - Delete actions first
   - Replace actions second
   - Update actions third
   - Add/Create actions last
3. When resources have the same danger level and action type, then the tertiary sort SHALL be alphabetical by resource address
4. When displaying sorted results, then the sort order SHALL be consistent across all output formats
5. When processing no-op changes, then they SHALL continue to be hidden from output
6. When implementing sorting, then the existing ActionSortTransformer SHALL be added back to the rendering pipeline (it has been debugged and fixed)
7. When the ActionSortTransformer is applied, then it SHALL continue to work only for the formats it currently supports (table, markdown, HTML, CSV)

### 6. General Applicability

**User Story:** As a user, I want these improvements to apply consistently across all plan summary usage, so that I have a consistent experience regardless of context.

**Acceptance Criteria:**
1. When any plan summary command is executed, then ALL improvements SHALL be applied
2. When testing with sample files like samples/danger-sample.json, then the output SHALL demonstrate all improvements
3. When implementing these changes, then they SHALL work across all supported output formats (table, JSON, HTML, Markdown)
4. When implementing these improvements, then NO configuration options SHALL be added (these are bug fixes, not features)

## Technical Implementation Notes

### Parser Enhancements Required
1. Extract property-level before/after values from Terraform plan JSON's "change" objects
2. Update data models to store actual property-level change data

### Formatter Updates Required
1. Create custom CollapsibleFormatter for property changes using go-output APIs
2. Fix ActionSortTransformer to work correctly across all formats
3. Update provider grouping logic to count only changed resources (excluding no-ops)
4. Remove dependency column from output tables

### Data Model Updates
1. Enhance PropertyChange struct to store actual before/after values
2. Ensure proper data flow from parser through analyzer to formatter