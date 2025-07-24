# Strata GitHub Action

The Strata GitHub Action provides seamless integration of Terraform plan analysis into your CI/CD workflows. It automatically analyzes Terraform plans and provides clear, actionable summaries through GitHub's native features.

## Quick Start

Add the following step to your workflow:

```yaml
- name: Analyze Terraform Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
```

## Features

### 🔍 Plan Analysis
- Comprehensive analysis of Terraform plan files (binary or JSON format)
- Statistical summaries of resource changes
- Identification of potentially dangerous changes
- Support for custom danger thresholds and configuration

### 📊 GitHub Integration
- **Step Summaries**: Rich Markdown output in GitHub step summaries
- **Pull Request Comments**: Automatic commenting on pull requests
- **Workflow Outputs**: Structured outputs for downstream workflow steps
- **Comment Management**: Smart comment updating to reduce noise

### ⚡ Performance
- Binary caching for faster execution
- Platform-specific binary downloads
- Fallback compilation from source
- Optimized for GitHub Actions runners

### 🔒 Security
- Minimal permission requirements
- Input validation and sanitization
- Secure token handling
- Rate limiting and error handling

## Configuration

### Inputs

| Input | Description | Required | Default | Example |
|-------|-------------|----------|---------|---------|
| `plan-file` | Path to Terraform plan file | ✅ | | `terraform.tfplan` |
| `output-format` | Output format for analysis | ❌ | `markdown` | `table`, `json`, `markdown` |
| `config-file` | Path to custom Strata config | ❌ | | `.strata.yaml` |
| `highlight-dangers` | Highlight potentially dangerous changes | ❌ | `true` | `true`, `false` |
| `show-details` | Show detailed change info | ❌ | `false` | `true`, `false` |
| `github-token` | GitHub token for API access | ❌ | `${{ github.token }}` | `${{ secrets.GITHUB_TOKEN }}` |
| `comment-on-pr` | Enable PR commenting | ❌ | `true` | `true`, `false` |
| `update-comment` | Update existing comments | ❌ | `true` | `true`, `false` |
| `comment-header` | Custom comment header | ❌ | `🏗️ Terraform Plan Summary` | `🚀 Infrastructure Changes` |

### Outputs

| Output | Description | Type | Example |
|--------|-------------|------|---------|
| `summary` | Plan summary text | String | `"Plan contains 3 changes..."` |
| `has-changes` | Whether plan has changes | Boolean | `true` |
| `has-dangers` | Whether dangerous changes detected | Boolean | `false` |
| `json-summary` | Full summary in JSON format | String | `{"totalChanges": 3, ...}` |
| `change-count` | Total number of changes | Number | `3` |
| `danger-count` | Number of dangerous changes | Number | `0` |

## Usage Examples

### Basic Terraform Workflow

```yaml
name: Terraform
on:
  pull_request:
    paths: ['**.tf', '**.tfvars']

permissions:
  contents: read
  pull-requests: write

jobs:
  plan:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v3
      
    - name: Terraform Init
      run: terraform init
      
    - name: Terraform Plan
      run: terraform plan -out=terraform.tfplan
      
    - name: Analyze Plan
      uses: ArjenSchwarz/strata@v1
      with:
        plan-file: terraform.tfplan
```

### Advanced Configuration

```yaml
- name: Analyze Terraform Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
    output-format: markdown
    config-file: .strata.yaml
    highlight-dangers: true
    show-details: true
    comment-header: "🚀 Infrastructure Changes"
    update-comment: false  # Always create new comments
```

### Multi-Environment Workflow

```yaml
name: Multi-Environment Terraform
on:
  pull_request:
    paths: ['**.tf', '**.tfvars']

permissions:
  contents: read
  pull-requests: write

jobs:
  plan:
    strategy:
      matrix:
        environment: [dev, staging, prod]
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v3
      
    - name: Terraform Plan
      run: |
        cd environments/${{ matrix.environment }}
        terraform init
        terraform plan -out=terraform.tfplan
      
    - name: Analyze Plan
      uses: ArjenSchwarz/strata@v1
      with:
        plan-file: environments/${{ matrix.environment }}/terraform.tfplan
        comment-header: "🏗️ ${{ matrix.environment }} Infrastructure Changes"
        highlight-dangers: true
        show-details: true
```

### Using Outputs in Downstream Steps

```yaml
- name: Analyze Plan
  id: strata
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan

- name: Check for Dangerous Changes
  if: steps.strata.outputs.has-dangers == 'true'
  run: |
    echo "⚠️ Dangerous changes detected!"
    echo "Danger count: ${{ steps.strata.outputs.danger-count }}"
    exit 1

- name: Notify Slack on Changes
  if: steps.strata.outputs.has-changes == 'true'
  uses: 8398a7/action-slack@v3
  with:
    status: custom
    custom_payload: |
      {
        text: "Terraform plan has ${{ steps.strata.outputs.change-count }} changes",
        attachments: [{
          color: "${{ steps.strata.outputs.has-dangers == 'true' && 'danger' || 'good' }}",
          text: "${{ steps.strata.outputs.summary }}"
        }]
      }
```

### Custom Configuration File

Create a `.strata.yaml` file in your repository:

```yaml
output: markdown
plan:
  show-details: true
  highlight-dangers: true
  always-show-sensitive: true

sensitive_resources:
  - resource_type: aws_db_instance
  - resource_type: aws_rds_cluster
  - resource_type: aws_ecs_service

sensitive_properties:
  - resource_type: aws_instance
    property: user_data
  - resource_type: aws_launch_template
    property: user_data
```

Then reference it in your workflow:

```yaml
- name: Analyze Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
    config-file: .strata.yaml
```

## Output Examples

### Step Summary

The action automatically generates rich step summaries:

```markdown
# 🏗️ Terraform Plan Summary

⚠️ **Plan contains changes with potential risks**

## Statistics Summary
| TO ADD | TO CHANGE | TO DESTROY | REPLACEMENTS | HIGH RISK |
|--------|-----------|------------|--------------|----------|
| 2 | 1 | 0 | 1 | 1 |

## Resource Changes
[Detailed table of resource changes]

<details>
<summary>📋 Detailed Changes</summary>
[Detailed change information]
</details>

---
*Generated by [Strata](https://github.com/ArjenSchwarz/strata)*
```

### Pull Request Comments

Comments are automatically posted to pull requests:

```markdown
## 🏗️ Terraform Plan Summary

<!-- strata-comment-id: terraform-plan -->

✅ **Plan contains changes**

**Statistics:**
- 📝 **Changes**: 3
- ⚠️ **Dangerous**: 0
- 🔄 **Replacements**: 1

**Plan Summary:**
[Summary table]

<details>
<summary>📋 Detailed Changes</summary>
[Detailed changes]
</details>

---
*Generated by [Strata](https://github.com/ArjenSchwarz/strata) in [workflow run](link)*
```

## Permissions

The action requires minimal permissions:

```yaml
permissions:
  contents: read        # To checkout code and read files
  pull-requests: write  # To comment on pull requests (optional)
```

If you don't need PR comments, you can omit the `pull-requests: write` permission:

```yaml
- name: Analyze Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
    comment-on-pr: false
```

## Troubleshooting

### Common Issues

#### Plan File Not Found
```
Error: Plan file does not exist: terraform.tfplan
```

**Solution**: Ensure the plan file is generated before running the action:
```yaml
- name: Terraform Plan
  run: terraform plan -out=terraform.tfplan

- name: Analyze Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
```

#### Permission Denied for PR Comments
```
Warning: Failed to create comment (HTTP 403)
```

**Solution**: Add the required permission:
```yaml
permissions:
  contents: read
  pull-requests: write
```

#### Binary Download Failed
```
Warning: Failed to download binary after multiple attempts, falling back to compilation
```

**Solution**: This is usually temporary. The action will automatically fall back to compiling from source. If the issue persists, check your network connectivity or GitHub's status.

#### Invalid Plan File Format
```
Warning: Strata analysis failed with exit code 1
```

**Solution**: Ensure you're using a valid Terraform plan file:
- Binary format: `terraform plan -out=terraform.tfplan`
- JSON format: `terraform show -json terraform.tfplan > plan.json`

### Debug Mode

Enable debug output for troubleshooting:

```yaml
- name: Analyze Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
  env:
    ACTIONS_STEP_DEBUG: true
```

### Rate Limiting

The action handles GitHub API rate limiting automatically:
- Checks rate limit before making requests
- Implements exponential backoff for retries
- Waits for rate limit reset when necessary

## Best Practices

### 1. Use in Pull Request Workflows
```yaml
on:
  pull_request:
    paths: ['**.tf', '**.tfvars']
```

### 2. Highlight Dangerous Changes
```yaml
with:
  highlight-dangers: true  # Always highlight potentially dangerous changes
```

### 3. Use Custom Headers for Multi-Environment
```yaml
with:
  comment-header: "🏗️ ${{ matrix.environment }} Infrastructure Changes"
```

### 4. Enable Details for Complex Plans
```yaml
with:
  show-details: true  # For comprehensive change information
```

### 5. Use Configuration Files for Consistency
```yaml
with:
  config-file: .strata.yaml  # Consistent configuration across workflows
```

## Integration with Other Tools

### Terraform Cloud/Enterprise

```yaml
- name: Download Plan from Terraform Cloud
  run: |
    # Download plan file from Terraform Cloud
    terraform show -json > plan.json

- name: Analyze Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: plan.json
```

### Atlantis

```yaml
# In your atlantis.yaml
workflows:
  default:
    plan:
      steps:
      - run: terraform plan -out=terraform.tfplan
      - run: |
          # Use Strata in Atlantis custom workflows
          strata plan summary terraform.tfplan
```

### Terragrunt

```yaml
- name: Terragrunt Plan
  run: |
    terragrunt plan -out=terraform.tfplan

- name: Analyze Plan
  uses: ArjenSchwarz/strata@v1
  with:
    plan-file: terraform.tfplan
```

## Contributing

If you encounter issues or have suggestions for the GitHub Action:

1. Check existing [issues](https://github.com/ArjenSchwarz/strata/issues)
2. Create a new issue with:
   - Workflow configuration
   - Error messages
   - Expected vs actual behavior
3. For feature requests, describe the use case and expected behavior

## License

The Strata GitHub Action is licensed under the MIT License, same as the main Strata project.