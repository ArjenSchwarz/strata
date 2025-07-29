# Decision Log - Plan Summary Output Improvements

## Decisions Made

### Requirements Clarifications (Post-Design Review)

#### Empty Tables
**Decision:** Empty tables specifically refer to:
- Resource Changes tables that would only contain no-ops (which are filtered out)
- Provider-specific tables when they would only contain no-ops
- Provider counts should only include changed resources, not total resources
- Provider grouping threshold should compare against total changed resources, not total resources

**Rationale:** No-ops are filtered out, so tables showing only filtered content are misleading

#### Property Changes Data Source
**Decision:** Enhance the parser to extract property-level before/after values from Terraform plan JSON's "change" objects
**Rationale:** Current implementation uses resource-level data incorrectly; proper data exists in the plan

#### Dependencies Parsing
**Decision:** Parse the "depends_on" field from Terraform plan JSON as the initial implementation
**Rationale:** This data exists in the plan and should be extracted properly

#### Sorting Implementation
**Decision:** Re-enable the existing ActionSortTransformer which has been debugged and fixed
**Rationale:** The transformer was previously working but disabled, it has been debugged and should be re-enabled

#### Configuration
**Decision:** No configuration options for these improvements - they are bug fixes
**Rationale:** These are corrections to how the tool should have worked from the beginning

### 1. Empty Table Suppression
**Decision:** Apply empty table suppression to ALL output formats (table, JSON, HTML, Markdown)
**Rationale:** There is no reason to show empty tables in any format - they add visual clutter without providing value

### 2. Property Changes Formatting
**Decision:** 
- Show property changes in Terraform-style format at the property level only (no full resource paths)
- Display nested objects in full with proper indentation
- Create a custom CollapsibleFormatter using go-output's capabilities
- Fix property count to accurately reflect individual property changes, not treat everything as a single property

**Rationale:** 
- Users are already at the resource level, so full paths are redundant
- Terraform's native format is familiar and readable
- Current implementation incorrectly counts all changes as a single property change

### 3. Property Changes Column
**Decision:** Show actual property change details in Terraform diff format, no emoji
**Rationale:** The column was incorrectly showing emoji instead of useful change information

### 4. Dependencies
**Decision:** 
- Show only direct dependencies
- Handle circular dependencies by marking them explicitly (e.g., "Dependency A (circular)")

**Rationale:** 
- Direct dependencies are sufficient for understanding relationships
- Circular dependencies need special handling to avoid infinite loops

### 5. Risk-Based Sorting
**Decision:** 
- Primary sort: Dangerous items first (anything with danger flags)
- Secondary sort: Action type (delete, replace, update, add)
- Tertiary sort: Alphabetical by resource address
- Continue hiding no-op changes

**Rationale:** 
- Dangerous changes need immediate attention
- Destructive actions (delete/replace) are more risky than constructive ones
- Alphabetical sorting provides consistency

### 6. Scope
**Decision:** Apply these improvements universally to all plan summary commands
**Rationale:** These are general improvements that enhance readability and usefulness across all contexts