# Markdown Output Format

Strata now supports markdown output format, which is useful for including plan summaries in documentation, pull requests, and other markdown-based contexts.

## Usage

To generate a plan summary in markdown format, use the `--output markdown` flag:

```bash
strata plan summary --output markdown terraform.tfplan
```

You can also save the output to a file:

```bash
strata plan summary --output markdown --file plan-summary.md terraform.tfplan
```

## Example Output

A plan summary in markdown format looks like this:

```markdown
# Plan Information

| Plan File | Version | Workspace | Backend | Created |
|-----------|---------|-----------|---------|---------|
| terraform.tfplan | 1.6.0 | default | local (terraform.tfstate) | 2025-06-14 10:15:23 |

# Summary for terraform.tfplan

| TOTAL | ADDED | REMOVED | MODIFIED | REPLACEMENTS | CONDITIONALS | HIGH RISK |
|-------|-------|---------|----------|--------------|-------------|-----------|
| 5     | 2     | 1       | 1        | 1            | 0           | 1         |

# Resource Changes

| ACTION  | RESOURCE           | TYPE          | ID           | REPLACEMENT | MODULE | DANGER |
|---------|-------------------|---------------|--------------|-------------|--------|--------|
| Add     | aws_instance.web  | aws_instance  | -            | Never       | -      |        |
| Add     | aws_s3_bucket.logs| aws_s3_bucket | -            | Never       | -      |        |
| Modify  | aws_vpc.main      | aws_vpc       | vpc-123456   | Never       | -      |        |
| Remove  | aws_rds_instance.old | aws_rds_instance | db-789012 | N/A       | -      |        |
| Replace | aws_rds_instance.new | aws_rds_instance | db-345678 | Always    | -      | ⚠️ Sensitive resource replacement |
```

## Integration with Pull Requests

Markdown output is particularly useful for including plan summaries in pull request descriptions or comments. This helps reviewers quickly understand the changes being made without having to run the plan themselves.

### GitHub Actions Example

Here's an example of how to use Strata with GitHub Actions to automatically add a plan summary to pull requests:

```yaml
name: Terraform Plan

on:
  pull_request:
    paths:
      - 'terraform/**'

jobs:
  terraform:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.6.0

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.1'

      - name: Install Strata
        run: |
          go install github.com/yourusername/strata@latest

      - name: Terraform Init
        run: terraform init
        working-directory: ./terraform

      - name: Terraform Plan
        run: terraform plan -out=terraform.tfplan
        working-directory: ./terraform

      - name: Generate Plan Summary
        run: |
          strata plan summary --output markdown --file plan-summary.md terraform.tfplan
        working-directory: ./terraform

      - name: Comment on PR
        uses: actions/github-script@v6
        with:
          github-token: ${{ secrets.GITHUB_TOKEN }}
          script: |
            const fs = require('fs');
            const planSummary = fs.readFileSync('./terraform/plan-summary.md', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Terraform Plan Summary\n\n' + planSummary
            });
```

## Benefits of Markdown Output

1. **Documentation Integration**: Easily include plan summaries in project documentation
2. **Pull Request Reviews**: Provide clear, formatted plan summaries for code reviewers
3. **Persistent Records**: Save plan summaries as markdown files for future reference
4. **Formatting Consistency**: Maintain consistent formatting across different environments
5. **Accessibility**: Make plan summaries accessible to team members without direct access to the Terraform environment