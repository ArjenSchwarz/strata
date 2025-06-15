# High-Risk Column in Statistics Summary

The high-risk column is a new addition to the statistics summary table that provides visibility into particularly risky changes in your Terraform plans.

## Overview

The high-risk column counts the number of resources that are marked as dangerous. This includes ANY resource with the `IsDangerous` flag set, which covers:
1. Sensitive resources that are being replaced
2. Resources with sensitive property changes
3. Any other resource flagged as dangerous by the analysis engine

The count includes ALL dangerous resources, regardless of whether they are in the sensitive resources configuration or not.

This provides a quick way to identify the most critical changes in your plan that require careful review.

## Implementation

The high-risk count is calculated in the `calculateStatistics` method in `analyzer.go`:

```go
// Count high-risk changes (any resource with the dangerous flag set)
if change.IsDangerous {
    stats.HighRisk++
}
```

The count is then displayed in the statistics summary table as the "HIGH RISK" column.

## Configuration

While the high-risk column counts ALL dangerous resources, you can configure which resources trigger the danger flag by setting up sensitive resources in your `strata.yaml` configuration file:

```yaml
sensitive_resources:
  - resource_type: aws_db_instance
  - resource_type: aws_rds_cluster
  - resource_type: aws_elasticache_cluster
  - resource_type: aws_elasticsearch_domain
  - resource_type: aws_msk_cluster
  - resource_type: aws_neptune_cluster
```

## Example Output

In the statistics summary table, the high-risk column appears alongside other change counts:

```
+-------+-------+---------+----------+--------------+---------------+-----------+
| TOTAL | ADDED | REMOVED | MODIFIED | REPLACEMENTS | CONDITIONALS  | HIGH RISK |
+-------+-------+---------+----------+--------------+---------------+-----------+
| 12    | 5     | 2       | 3        | 2            | 0             | 2         |
+-------+-------+---------+----------+--------------+---------------+-----------+
```

## Use Cases

The high-risk column is particularly useful in the following scenarios:

1. **Large Infrastructure Changes**: When making changes to many resources, it helps to quickly identify which ones require special attention
2. **CI/CD Pipelines**: Automated processes can use this count to determine whether manual approval is required
3. **Team Reviews**: Helps reviewers focus on the most critical changes first
4. **Change Management**: Provides a quantifiable metric for assessing the risk level of a planned change

## Best Practices

To make the most of the high-risk column:

1. **Configure Sensitive Resources**: Ensure your `strata.yaml` includes all resource types that are critical to your infrastructure
2. **Review High-Risk Changes First**: Always start by reviewing the resources counted in the high-risk column
3. **Set Thresholds**: Consider setting policies based on the high-risk count (e.g., requiring additional approvals when the count is above a certain threshold)
4. **Document Justifications**: For each high-risk change, document the reason and mitigation strategy

## Related Features

The high-risk column works in conjunction with other Strata features:

- **Danger Highlights**: Resources counted in the high-risk column are also highlighted with danger indicators in the detailed resource changes table
- **Always Show Sensitive**: When using the `always-show-sensitive` configuration option, high-risk resources will be displayed even when detailed output is disabled