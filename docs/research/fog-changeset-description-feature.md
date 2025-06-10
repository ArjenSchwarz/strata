# Changeset Summary Display in Fog Deploy

This document describes the detailed functionality of the changeset summary display feature in the `fog deploy` command, which provides users with a comprehensive overview of CloudFormation changes before deployment.

## Overview

The changeset summary is a critical part of fog's deployment workflow that presents users with detailed information about what changes will be made to their CloudFormation stack. This feature is implemented across several functions in [`cmd/describe_changeset.go`](cmd/describe_changeset.go) and [`cmd/deploy.go`](cmd/deploy.go), providing both tabular displays and interactive decision points.

## Core Functionality

### 1. Changeset Information Display

The summary begins with basic stack information via [`printBasicStackInfo`](cmd/describe_changeset.go):

- **Stack Name**: The name of the CloudFormation stack being modified
- **AWS Account**: Shows account alias and ID (e.g., "myaccount (123456789012)")
- **Region**: The AWS region where the stack exists
- **Action**: Either "Create" (for new stacks) or "Update" (for existing stacks)
- **Dry Run Status**: Indicates if this is a dry run operation

### 2. Resource Changes Table

The main changeset display is handled by [`printChangeset`](cmd/describe_changeset.go) and shows:

#### Table Columns
- **Action**: The type of change (Add, Remove, Modify) - "Remove" actions are displayed in bold
- **CfnName**: The logical ID of the resource in the CloudFormation template
- **Type**: The AWS resource type (e.g., `AWS::EC2::Instance`, `AWS::S3::Bucket`)
- **ID**: The physical resource ID (empty for new resources)
- **Replacement**: Indicates if the resource will be replaced (`True`, `Conditional`, `Never`)
- **Module**: (Optional) Shows module information for resources managed by CloudFormation modules

#### Special Handling
- Resources with no changes display: "No changes to resources have been found, but there are still changes to other parts of the stack"
- Empty changesets show appropriate messaging via [`DeployChangesetMessageNoResourceChanges`](lib/texts/deployments.go)

### 3. Summary Statistics

The system generates a summary table with aggregate counts:

- **Total**: Total number of resource changes
- **Added**: Number of resources being created
- **Removed**: Number of resources being deleted
- **Modified**: Number of resources being updated
- **Replacements**: Number of resources requiring replacement (`Replacement: True`)
- **Conditionals**: Number of resources with conditional replacement (`Replacement: Conditional`)

### 4. Potentially Destructive Changes

The [`printDangerTable`](cmd/describe_changeset.go) function identifies and highlights dangerous operations:

#### Criteria for "Dangerous" Changes
- **Remove actions**: Any resource being deleted
- **Replacement operations**: Resources with `Replacement: True` or `Replacement: Conditional`
- **Detailed impact analysis**: Uses [`GetDangerDetails`](lib/changesets.go) to show specific attributes requiring recreation

#### Danger Details Format
Details are formatted as: `{Evaluation}: {Attribute} - {CausingEntity}`

Examples:
- `Static: Properties.BucketName - BucketName`
- `Dynamic: Properties.Tags - Tags`

Only changes with `RequiresRecreation` set to `Always` or `Conditional` are included.

### 5. Console Integration

The summary includes a direct link to the AWS Console for detailed changeset review:
- URL format: `https://{region}.console.aws.amazon.com/cloudformation/home?region={region}#/stacks/changesets/changes?stackId={stackId}&changeSetId={changeSetId}`
- Generated via [`GenerateChangesetUrl`](lib/changesets.go)

## Interactive Workflow

### User Decision Points

After displaying the changeset summary, the system prompts for user action:

1. **Deploy Confirmation**: "Do you want to deploy this change set?" (via [`DeployChangesetMessageDeployConfirm`](lib/texts/deployments.go))
2. **Delete Confirmation**: If declined, "Do you want to delete this change set?" (via [`DeployChangesetMessageDeleteConfirm`](lib/texts/deployments.go))

### Non-Interactive Mode

With the `--non-interactive` flag:
- Automatically proceeds with deployment
- Displays: "Non-interactive mode: Automatically deploying the change set for you." (via [`DeployChangesetMessageAutoDeploy`](lib/texts/deployments.go))

### Dry Run Mode

With the `--dry-run` flag:
- Shows the changeset summary
- Automatically deletes the changeset without deployment
- Displays: "Dry run: Change set has been successfully created." (via [`DeployChangesetMessageDryrunSuccess`](lib/texts/deployments.go))

## Implementation Details

### Data Sources

The changeset data comes from AWS CloudFormation APIs:
- [`DescribeChangeSet`](lib/stacks.go) calls retrieve changeset details
- [`AddChangeset`](lib/stacks.go) processes the raw AWS response into structured data
- [`ChangesetChanges`](lib/changesets.go) struct holds individual resource change information

### Module Support

The system detects and displays CloudFormation module information:
- Sets [`HasModule`](lib/changesets.go) flag when modules are present
- Displays module hierarchy: `{LogicalIdHierarchy}({TypeHierarchy})`
- Adjusts table layout to include Module column when applicable

### Output Formatting

Uses the [`go-output`](https://github.com/ArjenSchwarz/go-output) library for consistent table formatting:
- Configurable via [`outputsettings`](cmd/deploy.go)
- Supports sorting by Type column
- Separate tables for better visual separation (`SeparateTables: true`)

### Error Handling

Comprehensive error handling for various scenarios:
- **No Changes**: Detects when changesets contain no modifications
- **Creation Failures**: Handles changeset creation errors with detailed messaging
- **Retrieval Failures**: Manages API failures when fetching changeset data

## User Experience Features

### Visual Enhancements
- **Bold formatting** for destructive actions (Remove operations)
- **Color coding** for different message types (success, warning, failure)
- **Structured layout** with clear section headers and spacing

### Information Hierarchy
1. Stack context (account, region, action type)
2. Detailed resource changes
3. Summary statistics
4. Dangerous operations highlight
5. External links for detailed review

### Consistency
- Standardized messaging via [`lib/texts/deployments.go`](lib/texts/deployments.go)
- Consistent table formatting across all changeset displays
- Unified approach to user confirmations and feedback

## Integration Points

The changeset summary integrates with other fog features:
- **Deployment Logging**: Changes are recorded in [`DeploymentLog`](lib/logging.go) for audit trails
- **Report Generation**: Summary data feeds into deployment reports
- **Configuration**: Respects user-defined output preferences and formatting settings

This comprehensive changeset summary ensures users have complete visibility into infrastructure changes before committing to deployment, supporting both interactive review and automated deployment workflows.