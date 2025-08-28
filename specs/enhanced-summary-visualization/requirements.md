# Enhanced Summary Visualization Requirements

## Introduction

The Enhanced Summary Visualization feature aims to improve how Terraform plan summaries are displayed and presented to users. This feature will enhance the readability and organization of plan summaries by providing optional resource grouping, more detailed change context, and clearer risk highlighting. The goal is to help users quickly understand the impact and risk level of proposed infrastructure changes while maintaining simplicity.

## Requirements

### 1. Progressive Disclosure with Collapsible Sections

**User Story:** As a DevOps engineer, I want to see a clean summary with the ability to expand sections for detailed information, so that I can quickly scan high-level changes while having access to comprehensive details when needed.

**Acceptance Criteria:**
1. The system SHALL use collapsible sections for all detailed information display
2. The system SHALL show essential information (resource name, change type, risk level) by default
3. The system SHALL allow expansion of detailed sections including:
   - Complete property change lists (not limited to 3)
   - Risk analysis and mitigation suggestions
   - Resource dependencies and relationships
   - Provider/service groupings
4. The system SHALL mark dangerous change groups as expanded by default
5. The system SHALL preserve the smart grouping hierarchy within collapsible sections:
   - If all resources are from the same provider, provider grouping SHALL be omitted
   - If all resources within a provider are from the same service, service grouping SHALL be omitted
6. The system SHALL NOT apply grouping when the total number of modified resources is below a configurable threshold (default: 10)
7. The system SHALL display resource counts in section headers when grouping is active

### 2. Comprehensive Change Context with Progressive Disclosure

**User Story:** As an infrastructure engineer, I want to see comprehensive context about what's changing in each resource within collapsible sections, so that I can understand the full impact without overwhelming the initial view.

**Acceptance Criteria:**
1. For resource replacements:
   - The system SHALL display ALL replacement reasons provided by Terraform in the main view
   - The system SHALL show detailed property changes in a collapsible section
   - The system SHALL include risk analysis and recommended mitigation steps in expandable sections
2. For resource updates (modifications without replacement):
   - The system SHALL show a summary of changed properties in the main view
   - The system SHALL show ALL property changes with before/after values in expandable sections
   - The system SHALL highlight sensitive property changes in the main view
3. The system SHALL show dependency information in expandable sections for all resource types
4. The system SHALL clearly indicate when a resource is being replaced versus updated in-place
5. The system SHALL provide expandable sections for:
   - Complete property change details
   - Resource dependency graphs
   - Risk assessment and mitigation recommendations
   - Related resource impacts

### 3. Enhanced Risk Analysis with Detailed Mitigation

**User Story:** As a team lead, I want clear visual indicators of risky changes with detailed risk analysis and mitigation suggestions in expandable sections, so that I can quickly identify changes that need careful review and understand how to address them safely.

**Acceptance Criteria:**
1. The system SHALL use color coding ONLY for risky changes (no rainbow effect)
2. The system SHALL use the existing sensitive resources and properties configuration to determine risk
3. The system SHALL treat all deletion operations as risky by default
4. The system SHALL treat deletions of sensitive resources as higher risk than regular deletions
5. The system SHALL provide brief explanations for why changes are considered risky in the main view
6. The system SHALL provide detailed risk analysis in expandable sections including:
   - Dependencies that may be affected
7. The system SHALL automatically expand risk detail sections for high-risk changes by default

### 4. GitHub Action Integration

**User Story:** As a CI/CD engineer, I want the enhanced summary visualization to work seamlessly in GitHub Actions with expandable sections in PR comments, so that reviewers can see clear summaries with the ability to drill down into details.

**Acceptance Criteria:**
1. The system SHALL enable expandable sections when running in GitHub Actions
2. The system SHALL produce Markdown output compatible with GitHub PR comments
3. The system SHALL respect the expand-all configuration when set via CLI or config
4. The system SHALL work correctly with the existing GitHub Action workflow

### 5. Global Expand Control

**User Story:** As a user, I want a global flag to expand all collapsible sections at once, so that I can see all details when needed without manually expanding each section.

**Acceptance Criteria:**
1. The system SHALL provide a global `--expand-all` CLI flag that expands all collapsible sections
2. The system SHALL support an `expand_all` configuration option in strata.yaml
3. The CLI flag SHALL override the configuration file setting when both are present
4. The expand-all setting SHALL be a top-level configuration, not tied to the plan section
5. The system SHALL apply expand-all to all collapsible content (sections and field values)

## Configuration

The feature will be configured through the existing strata.yaml configuration file:

```yaml
# Global expand control
expand_all: false                    # Expand all collapsible sections (default: false)

plan:
  expandable_sections:
    enabled: true                    # Enable/disable collapsible sections (default: true)
    auto_expand_dangerous: true      # Auto-expand high-risk sections (default: true)
    show_dependencies: true          # Show dependency sections (default: true)
  grouping:
    enabled: true                    # Enable/disable grouping (default: true)
    threshold: 10                    # Minimum resources to trigger grouping (default: 10)
```

## Implementation Notes

Based on the collapsible sections capability:
- All detailed information will be displayed in collapsible sections using go-output v2's expandable section support
- Main view will show essential information (resource name, change type, brief risk indicator)
- Expandable sections will provide comprehensive details without cluttering the primary view
- High-risk changes will have their detail sections automatically expanded
- Grouping will use smart hierarchy within collapsible sections
- Dependencies will be shown in expandable sections for affected resources
- The global `expand_all` flag will be implemented at the root command level
- All plan-specific configuration options will be under the existing `plan:` section with new `expandable_sections:` subsection
- GitHub Action integration will automatically use Markdown format with proper collapsible syntax

## Benefits of Collapsible Sections Integration

1. **Eliminates information density constraints** - No longer limited to showing only 3 properties
2. **Provides comprehensive context** - Users can access full details when needed
3. **Maintains clean overview** - Primary view remains uncluttered
4. **Enhances risk awareness** - Detailed mitigation guidance available on demand
5. **Improves workflow efficiency** - Quick scanning with drill-down capability