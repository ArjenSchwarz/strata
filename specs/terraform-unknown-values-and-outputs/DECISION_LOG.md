# Decision Log: Terraform Unknown Values and Outputs Feature

## Feature Name Decision
**Decision**: Use `terraform-unknown-values-and-outputs` as the feature name
**Date**: 2025-08-03
**Context**: User proposed this feature to handle unknown values display and add outputs section
**Rationale**: Name clearly describes both aspects of the feature (unknown values + outputs)

## Unknown Value Representation
**Decision**: Use exact Terraform syntax `(known after apply)` 
**Date**: 2025-08-03
**Context**: User specified to use exact Terraform syntax for familiarity
**Rationale**: Users are already familiar with this text from Terraform's native output

## Data Format Consistency
**Decision**: Display `(known after apply)` as string across all output formats
**Date**: 2025-08-03
**Context**: Need consistency across table, JSON, HTML, and Markdown formats
**Rationale**: Even JSON can represent this as a string value for simplicity

## Problem Identification
**Decision**: Focus on `after_unknown` field processing to fix "deletion" display issue
**Date**: 2025-08-03
**Context**: User identified specific issue in `wildcards-sample.json` where `aws_iam_policy.wildcard_admin_policy.arn` shows as deleted but should show as unknown
**Rationale**: Concrete example proves the problem exists and needs `after_unknown` field processing

## Outputs Section Design
**Decision**: Use 5-column table format placed after resource changes
**Date**: 2025-08-03
**Context**: UI/UX agent recommended table structure and placement
**Rationale**: 
- Follows existing visual hierarchy
- 5 columns provide all necessary information: NAME, ACTION, CURRENT, PLANNED, SENSITIVE
- Placement after resources maintains logical flow

## Collapsible Behavior
**Decision**: Outputs section will NOT be collapsible initially
**Date**: 2025-08-03
**Context**: User specified outputs shouldn't have as much information as resources
**Rationale**: Simpler implementation, fewer details than resource changes

## Sensitive Output Handling
**Decision**: Display `(sensitive value)` with ⚠️ indicator
**Date**: 2025-08-03
**Context**: Need to handle sensitive outputs appropriately
**Rationale**: Follows existing sensitive value patterns in codebase

## Empty Outputs Section
**Decision**: Suppress empty outputs section (no "no changes" message)
**Date**: 2025-08-03
**Context**: UI/UX consideration for when no output changes exist
**Rationale**: Cleaner interface when no outputs are changing

## Edge Case Handling
**Decision**: Follow Terraform's standard behavior for all edge cases
**Date**: 2025-08-03
**Context**: User specified to investigate and follow Terraform's behavior
**Rationale**: 
- Maintains consistency with user expectations
- Research revealed complex edge cases that Terraform already handles
- Reduces implementation complexity by following established patterns

## Performance Considerations
**Decision**: No specific performance constraints required
**Date**: 2025-08-03
**Context**: User confirmed no performance concerns
**Rationale**: Focus on functionality over optimization initially

## Data Structure Integration
**Decision**: Unknown values override standard before/after comparison logic
**Date**: 2025-08-03
**Context**: Properties in `after_unknown` should not appear as deletions
**Rationale**: Fixes the core issue where unknown values are misinterpreted as deleted properties

## Visual Indicators for Actions
**Decision**: Use consistent action indicators across all change types
**Date**: 2025-08-03
**Context**: Need visual consistency between resource and output changes
**Rationale**: 
- Add: "Add" with + indicator
- Update: "Modify" with ~ indicator  
- Delete: "Remove" with - indicator
- Maintains consistency with existing resource change indicators

## Implementation Priority
**Decision**: Implement unknown values first, then outputs section
**Date**: 2025-08-03
**Context**: Two distinct features that can be implemented incrementally
**Rationale**: Unknown values fix is more critical (fixes existing bug), outputs are enhancement

## Statistics Scope
**Decision**: Statistics are for resource changes only, not outputs
**Date**: 2025-08-03
**Context**: User clarified that ChangeStatistics struct tracks resource-level changes exclusively
**Rationale**: 
- Statistics (Total, ToAdd, ToChange, etc.) represent resource changes only
- Output changes are displayed separately in their own section
- Unknown properties only affect the "x properties changed" count in collapsible sections
- Maintains consistency with existing statistics structure and purpose