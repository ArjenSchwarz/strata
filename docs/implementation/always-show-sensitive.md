# Always Show Sensitive Resources Feature

The "Always Show Sensitive" feature ensures that critical resource changes are never hidden, even when detailed output is disabled.

## Overview

By default, when the `--details` flag is set to false, Strata only shows the plan information and statistics summary, hiding the detailed resource changes. However, this can lead to important sensitive resource changes being overlooked.

The "Always Show Sensitive" feature addresses this by:
1. Showing sensitive resource changes even when details are disabled
2. Filtering the output to only include resources marked as dangerous
3. Providing a focused view of the most critical changes

## Configuration

To enable this feature, add the `always-show-sensitive` option to your `strata.yaml` configuration file:

```yaml
plan:
  danger-threshold: 3
  show-details: false
  highlight-dangers: true
  always-show-sensitive: true  # Always show sensitive resources even when details are disabled
```

You can also enable it via command-line flags (if implemented):

```bash
strata plan summary --details=false --always-show-sensitive=true terraform.tfplan
```

## Implementation

The feature is implemented in the `OutputSummary` method in `formatter.go`:

```go
// Output enhanced resource changes table if requested
if showDetails {
    if err := f.formatResourceChangesTable(summary, outputFormat); err != nil {
        return fmt.Errorf("failed to format resource changes table: %w", err)
    }
} else if f.config.Plan.AlwaysShowSensitive {
    // When details are disabled but AlwaysShowSensitive is enabled,
    // show only the sensitive resource changes
    if err := f.formatSensitiveResourceChanges(summary, outputFormat); err != nil {
        return fmt.Errorf("failed to format sensitive resource changes: %w", err)
    }
}
```

The `formatSensitiveResourceChanges` method filters the resource changes to only include those marked as dangerous:

```go
// Filter for sensitive resources
sensitiveChanges := []ResourceChange{}
for _, change := range summary.ResourceChanges {
    if change.IsDangerous {
        sensitiveChanges = append(sensitiveChanges, change)
    }
}
```

## Example Output

When the "Always Show Sensitive" feature is enabled and details are disabled, the output will look like this:

```
Plan Information
+----------+--------+-----------+-------------------+---------------------+
| Plan File| Version| Workspace | Backend           | Created             |
+----------+--------+-----------+-------------------+---------------------+
| test.plan| 1.6.0  | default   | local (terraform) | 2025-06-14 10:15:23 |
+----------+--------+-----------+-------------------+---------------------+

Summary for test.plan
+-------+-------+---------+----------+--------------+---------------+-----------+
| TOTAL | ADDED | REMOVED | MODIFIED | REPLACEMENTS | CONDITIONALS  | HIGH RISK |
+-------+-------+---------+----------+--------------+---------------+-----------+
| 12    | 5     | 2       | 3        | 2            | 0             | 2         |
+-------+-------+---------+----------+--------------+---------------+-----------+

Sensitive Resource Changes
+--------+------------------------+------------------+------------+-------------+--------+----------------------------------------+
| ACTION | RESOURCE               | TYPE             | ID         | REPLACEMENT | MODULE | DANGER                                 |
+--------+------------------------+------------------+------------+-------------+--------+----------------------------------------+
| Replace| aws_rds_instance.main  | aws_rds_instance | db-123456  | Always      | -      | ⚠️ Sensitive resource replacement      |
| Replace| aws_msk_cluster.events | aws_msk_cluster  | msk-789012 | Always      | kafka  | ⚠️ Sensitive resource replacement      |
+--------+------------------------+------------------+------------+-------------+--------+----------------------------------------+
```

## Use Cases

The "Always Show Sensitive" feature is particularly useful in the following scenarios:

1. **CI/CD Pipelines**: Ensure critical changes are always visible in automated workflows
2. **Large Teams**: Provide a focused view for reviewers who don't need to see all changes
3. **Change Management**: Highlight only the changes that require special attention or approval
4. **Security Reviews**: Focus security reviews on the most sensitive resources

## Best Practices

To make the most of this feature:

1. **Configure Sensitive Resources**: Ensure your `strata.yaml` includes all resource types that are critical to your infrastructure
2. **Use with `--details=false`**: This feature is most useful when combined with disabled details
3. **Review All Changes**: While this feature highlights sensitive changes, it's still important to review the full plan for complete understanding
4. **Combine with High-Risk Column**: Use the high-risk column in the statistics summary to get a quick count of sensitive dangerous changes

## Related Features

The "Always Show Sensitive" feature works in conjunction with other Strata features:

- **Danger Highlights**: Resources shown by this feature are those marked with danger indicators
- **High-Risk Column**: The count in the high-risk column corresponds to the number of resources shown when this feature is enabled
- **Markdown Output**: This feature works with all output formats, including markdown