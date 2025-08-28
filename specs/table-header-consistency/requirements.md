# Table Header Consistency Feature Requirements

## Introduction

This feature addresses the critical inconsistency in table header formatting across Strata's Terraform plan summary output. Currently, the four main tables use three different header capitalization styles (Title Case, ALL UPPERCASE, and all lowercase), creating a fragmented user experience that reduces readability and professional appearance. This feature will standardize all table headers to use consistent Title Case formatting across all output formats.

## Requirements

### 1. Summary Statistics Table Header Standardization

**User Story**: As a DevOps engineer using Strata to review Terraform plan summaries, I want the Summary Statistics table headers to use consistent Title Case formatting, so that I can quickly scan statistical information without cognitive dissonance from inconsistent visual styles.

**Acceptance Criteria**:
1. WHEN the system displays the Summary Statistics table, THEN the "TOTAL CHANGES" header SHALL be formatted as "Total Changes"
2. WHEN the system displays the Summary Statistics table, THEN the "ADDED" header SHALL be formatted as "Added"
3. WHEN the system displays the Summary Statistics table, THEN the "REMOVED" header SHALL be formatted as "Removed"
4. WHEN the system displays the Summary Statistics table, THEN the "MODIFIED" header SHALL be formatted as "Modified"
5. WHEN the system displays the Summary Statistics table, THEN the "REPLACEMENTS" header SHALL be formatted as "Replacements"
6. WHEN the system displays the Summary Statistics table, THEN the "HIGH RISK" header SHALL be formatted as "High Risk"
7. WHEN the system displays the Summary Statistics table, THEN the "UNMODIFIED" header SHALL be formatted as "Unmodified"

### 2. Resource Changes Table Header Standardization

**User Story**: As a user reviewing Terraform resource changes, I want the Resource Changes table headers to use consistent Title Case formatting, so that I can easily understand the column purposes and scan resource information effectively.

**Acceptance Criteria**:
1. WHEN the system displays the Resource Changes table, THEN the "action" header SHALL be formatted as "Action"
2. WHEN the system displays the Resource Changes table, THEN the "resource" header SHALL be formatted as "Resource"
3. WHEN the system displays the Resource Changes table, THEN the "type" header SHALL be formatted as "Type"
4. WHEN the system displays the Resource Changes table, THEN the "id" header SHALL be formatted as "ID"
5. WHEN the system displays the Resource Changes table, THEN the "replacement" header SHALL be formatted as "Replacement"
6. WHEN the system displays the Resource Changes table, THEN the "module" header SHALL be formatted as "Module"
7. WHEN the system displays the Resource Changes table, THEN the "danger" header SHALL be formatted as "Danger"
8. WHEN the system displays the Resource Changes table, THEN the "property_changes" header SHALL be formatted as "Property Changes"

### 3. Output Changes Table Header Standardization

**User Story**: As a user reviewing Terraform output changes, I want the Output Changes table headers to use consistent Title Case formatting, so that I have a uniform experience when scanning output modifications.

**Acceptance Criteria**:
1. WHEN the system displays the Output Changes table, THEN the "NAME" header SHALL be formatted as "Name"
2. WHEN the system displays the Output Changes table, THEN the "ACTION" header SHALL be formatted as "Action"
3. WHEN the system displays the Output Changes table, THEN the "CURRENT" header SHALL be formatted as "Current"
4. WHEN the system displays the Output Changes table, THEN the "PLANNED" header SHALL be formatted as "Planned"
5. WHEN the system displays the Output Changes table, THEN the "SENSITIVE" header SHALL be formatted as "Sensitive"

### 4. Sensitive Resource Changes Table Header Standardization

**User Story**: As a security-conscious user reviewing sensitive resource changes, I want the Sensitive Resource Changes table headers to use consistent Title Case formatting, so that I can focus on security implications without distraction from formatting inconsistencies.

**Acceptance Criteria**:
1. WHEN the system displays the Sensitive Resource Changes table, THEN the "ACTION" header SHALL be formatted as "Action"
2. WHEN the system displays the Sensitive Resource Changes table, THEN the "RESOURCE" header SHALL be formatted as "Resource"
3. WHEN the system displays the Sensitive Resource Changes table, THEN the "TYPE" header SHALL be formatted as "Type"
4. WHEN the system displays the Sensitive Resource Changes table, THEN the "ID" header SHALL be formatted as "ID"
5. WHEN the system displays the Sensitive Resource Changes table, THEN the "REPLACEMENT" header SHALL be formatted as "Replacement"
6. WHEN the system displays the Sensitive Resource Changes table, THEN the "MODULE" header SHALL be formatted as "Module"
7. WHEN the system displays the Sensitive Resource Changes table, THEN the "DANGER" header SHALL be formatted as "Danger"

### 5. Provider-Grouped Tables Header Standardization

**User Story**: As a user working with large Terraform plans, I want provider-grouped table headers to use consistent Title Case formatting, so that I can navigate grouped resources with the same visual consistency as ungrouped tables.

**Acceptance Criteria**:
1. WHEN the system displays provider-grouped tables, THEN all headers SHALL use the same Title Case formatting as the main Resource Changes table
2. WHEN provider grouping is enabled, THEN header formatting SHALL remain consistent with non-grouped tables
3. WHEN dynamically generated provider tables are displayed, THEN headers SHALL follow Title Case conventions

### 6. Cross-Format Consistency

**User Story**: As a user who consumes Strata output in different formats (terminal, web, documentation), I want table headers to maintain consistent Title Case formatting across supported output formats, so that I have a uniform experience when viewing results in formats that support custom header formatting.

**Acceptance Criteria**:
1. WHEN output is generated in HTML format, THEN all headers SHALL use Title Case formatting
2. WHEN output is generated in Markdown format, THEN all headers SHALL use Title Case formatting
3. WHEN output is generated in JSON format, THEN headers SHALL use Title Case formatting in structured output
4. WHEN switching between supported output formats, THEN header formatting SHALL remain consistent
5. WHEN output is generated in table format, THEN headers will display in ALL UPPERCASE as determined by the table format renderer (this is acceptable and cannot be changed)

### 7. Technical Term Preservation

**User Story**: As a technical user, I want abbreviations and technical terms in headers to maintain their correct capitalization, so that technical accuracy is preserved while achieving visual consistency.

**Acceptance Criteria**:
1. WHEN displaying the "ID" header, THEN it SHALL remain "ID" and not be changed to "Id"
2. WHEN displaying technical abbreviations in headers, THEN they SHALL maintain their established capitalization conventions
3. WHEN applying Title Case formatting, THEN proper nouns and technical terms SHALL preserve their correct spelling

### 8. Implementation and Testing

**User Story**: As a developer maintaining Strata, I want the header formatting changes to be implemented correctly and thoroughly tested, so that the changes work reliably across all scenarios without breaking existing functionality.

**Acceptance Criteria**:
1. WHEN changes are implemented, THEN they SHALL modify the correct header definition mechanisms in the codebase
2. WHEN the implementation is complete, THEN all existing automated tests SHALL continue to pass
3. WHEN the changes are deployed, THEN no breaking changes SHALL be introduced to the output structure or API
4. WHEN testing is performed, THEN header consistency SHALL be verified across all table types and output formats
5. WHEN regression testing is performed, THEN existing functionality SHALL remain unaffected