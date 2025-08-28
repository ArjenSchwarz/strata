# Terraform Unknown Values and Outputs Feature Requirements

## Introduction

This feature enhances the Strata Terraform plan summary tool to provide complete parity with Terraform's native output by:
1. Properly displaying unknown values that will be determined after apply (instead of showing them as deleted)
2. Adding a dedicated outputs section to show output value changes

## Requirements

### 1. Unknown Values Display
**User Story**: As a DevOps engineer, I want to see when resource values will be known after apply, so that I can distinguish between actual deletions and values that are simply not yet determined.

**Acceptance Criteria**:
1.1. **When** a resource property value is unknown (present in `after_unknown`), **then** the system shall display it with the exact text `(known after apply)`.
1.2. **When** a property has an unknown value, **then** the system shall NOT display it as being deleted (e.g., `aws_iam_policy.wildcard_admin_policy.arn` in wildcards-sample.json).
1.3. **When** formatting unknown values, **then** the system shall use `(known after apply)` as a string across all output formats (table, JSON, HTML, Markdown).
1.4. **When** a property changes from a known value to unknown, **then** the system shall show `before_value → (known after apply)`.
1.5. **When** a property is newly created with an unknown value, **then** the system shall indicate it as `+ property = (known after apply)`.
1.6. **When** processing `after_unknown` field, **then** properties marked as unknown shall override standard before/after comparison logic.
1.7. **When** a property remains unknown (unknown before and after), **then** the system shall show `(known after apply) → (known after apply)`.

### 2. Outputs Display
**User Story**: As a Terraform user, I want to see changes to output values in the plan summary, so that I can understand what outputs will change after applying the plan.

**Acceptance Criteria**:
2.1. **When** the plan contains output changes, **then** the system shall display them in a dedicated "Output Changes" section after resource changes.
2.2. **When** displaying outputs, **then** the system shall show a 5-column table: NAME, ACTION, CURRENT, PLANNED, SENSITIVE.
2.3. **When** an output value is unknown, **then** the system shall show `(known after apply)` in the PLANNED column.
2.4. **When** outputs are sensitive, **then** the system shall display `(sensitive value)` with ⚠️ indicator in the SENSITIVE column.
2.5. **When** outputs are added, **then** the system shall show ACTION as "Add" with + indicator.
2.6. **When** outputs are updated, **then** the system shall show ACTION as "Modify" with ~ indicator.
2.7. **When** outputs are removed, **then** the system shall show ACTION as "Remove" with - indicator.
2.8. **When** the outputs section is empty, **then** the system shall suppress the section entirely (no "no changes" message).

### 3. Integration Requirements
**User Story**: As a Strata user, I want the unknown values and outputs features to integrate seamlessly with existing functionality, so that I have a cohesive experience.

**Acceptance Criteria**:
3.1. **When** unknown values appear in dangerous or sensitive properties, **then** the system shall still apply appropriate danger highlighting.
3.2. **When** using collapsible sections, **then** unknown values shall be included in the expanded details.
3.3. **When** generating statistics, **then** the system shall properly categorize resource changes involving unknown values (statistics track resource changes only, not outputs).
3.4. **When** using different output formats, **then** unknown values and outputs shall be consistently represented.

## Implementation Specifications

### Unknown Value Processing
- **Detection**: Process `after_unknown` field in Terraform JSON to identify unknown properties
- **Representation**: Use exact Terraform syntax `(known after apply)` for all output formats
- **Integration**: Unknown values take precedence over property change analysis
- **Data Structure**: Properties in `after_unknown` should not be treated as deletions

### Outputs Section Design
- **Placement**: Appears after resource changes section
- **Format**: 5-column table: NAME, ACTION, CURRENT, PLANNED, SENSITIVE
- **Visual Indicators**: 
  - Create: "Add" with + indicator
  - Update: "Modify" with ~ indicator  
  - Delete: "Remove" with - indicator
- **Sensitive Values**: Display "(sensitive value)" with ⚠️ indicator
- **Not Collapsible**: Simple section without progressive disclosure initially

### Edge Case Handling
Following Terraform's standard behavior:
- **Properties remaining unknown**: Display `(known after apply)` for both before and after
- **Unknown to known transitions**: Show before value → new value
- **Known to unknown transitions**: Show before value → `(known after apply)`
- **Complex nested structures**: Show structure with `(known after apply)` for unknown nested values
- **Circular dependencies**: Process as Terraform handles them (defer until more info available)

### Cross-Format Consistency
- **JSON**: Unknown values as string `"(known after apply)"`
- **Table/HTML/Markdown**: Display `(known after apply)` as formatted text
- **Statistics**: Unknown values count as their respective change types (create/update/delete)