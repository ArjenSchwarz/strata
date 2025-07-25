# Enhanced Summary Visualization Requirements

## Introduction

The Enhanced Summary Visualization feature aims to improve how Terraform plan summaries are displayed and presented to users. This feature will enhance the readability and organization of plan summaries by providing optional resource grouping, more detailed change context, and clearer risk highlighting. The goal is to help users quickly understand the impact and risk level of proposed infrastructure changes while maintaining simplicity.

## Requirements

### 1. Optional Resource Grouping

**User Story:** As a DevOps engineer, I want resources to be optionally grouped by logical categories when dealing with many changes, so that I can better understand which parts of my infrastructure are being modified without cluttering simple plans.

**Acceptance Criteria:**
1. The system SHALL apply smart grouping hierarchy when grouping is enabled:
   - If all resources are from the same provider, provider grouping SHALL be omitted
   - If all resources within a provider are from the same service, service grouping SHALL be omitted
2. The system SHALL group by provider (e.g., AWS, Azure, GCP) only when multiple providers are present
3. The system SHALL group by service type (e.g., EC2, RDS, S3) only when multiple services are present within a provider
4. The system SHALL NOT apply grouping when the total number of modified resources (added/changed/deleted) is below a configurable threshold (default: 10)
5. The system SHALL allow users to completely disable grouping via configuration
6. The system SHALL allow users to configure the grouping threshold via the configuration file
7. The system SHALL display resource counts for each group when grouping is active

### 2. Enhanced Change Context

**User Story:** As an infrastructure engineer, I want to see more context about what's changing in each resource, so that I can understand the impact without opening the full plan.

**Acceptance Criteria:**
1. For resource replacements:
   - The system SHALL display ALL replacement reasons provided by Terraform
   - The system SHALL NOT show individual property changes when showing replacement reasons
2. For resource updates (modifications without replacement):
   - The system SHALL show the first 3 properties that are being changed
   - The system SHALL show before/after values for these properties when available
3. The system SHALL NOT show property details for newly created resources
4. The system SHALL NOT show property details for deleted resources
5. The system SHALL clearly indicate when a resource is being replaced versus updated in-place
6. The system MAY optionally indicate dependencies between resources being changed (disabled by default)

### 3. Risk Highlighting

**User Story:** As a team lead, I want clear visual indicators of risky changes, so that I can quickly identify changes that need careful review.

**Acceptance Criteria:**
1. The system SHALL use color coding ONLY for risky changes (no rainbow effect)
2. The system SHALL use the existing sensitive resources and properties configuration to determine risk
3. The system SHALL treat all deletion operations as risky by default
4. The system SHALL treat deletions of sensitive resources as higher risk than regular deletions
5. The system SHALL provide brief explanations for why changes are considered risky (e.g., "Sensitive resource deletion", "Database replacement")

## Configuration

The feature will be configured through the existing strata.yaml configuration file:

```yaml
plan:
  grouping:
    enabled: true              # Enable/disable grouping (default: true)
    threshold: 10              # Minimum resources to trigger grouping (default: 10)
  context:
    show_dependencies: false   # Show resource dependencies (default: false)
```

## Implementation Notes

Based on the requirements clarification:
- Change context will show the first 3 changed properties for updates only
- Resource replacements will show all Terraform-provided replacement reasons instead of property changes
- Grouping will use smart hierarchy that omits unnecessary levels when all resources share the same provider or service
- Risk explanations will be brief, single-phrase descriptions
- All configuration options will be under the existing `plan:` section

Do the requirements look good or do you want additional changes?