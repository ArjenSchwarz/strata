# Danger Highlights Feature Design

## Overview

The Danger Highlights feature enhances Strata's Terraform plan analysis by allowing users to define specific resource types and properties that should trigger warnings when modified. This feature addresses two key scenarios:

1. Warning when sensitive resources are being replaced
2. Warning when sensitive properties are updated, even if the resource itself is not being replaced

This feature will help users identify potentially risky changes in their Terraform plans before applying them, improving safety in infrastructure deployments.

## Architecture

The Danger Highlights feature will integrate with the existing plan analysis workflow:

1. Configuration parsing will be extended to include sensitive resources and properties
2. The plan analyzer will check resources and properties against these configurations
3. When matches are found, the analyzer will flag them as dangerous
4. The formatter will highlight these dangers in the output

## Components and Interfaces

### Configuration Extension

We'll extend the existing configuration structure in `config/config.go` to include:

```go
type Config struct {
    // Existing fields...
    
    // New fields for danger highlights
    SensitiveResources []SensitiveResource `mapstructure:"sensitive_resources"`
    SensitiveProperties []SensitiveProperty `mapstructure:"sensitive_properties"`
}

type SensitiveResource struct {
    ResourceType string `mapstructure:"resource_type"`
}

type SensitiveProperty struct {
    ResourceType string `mapstructure:"resource_type"`
    Property     string `mapstructure:"property"`
}
```

### Plan Analyzer Enhancement

The plan analyzer in `lib/plan/analyzer.go` will be extended with:

1. A method to check if a resource is in the sensitive resources list
2. A method to check if a property change matches the sensitive properties list
3. Logic to mark resources and changes as dangerous based on these checks

```go
func (a *Analyzer) IsSensitiveResource(resourceType string) bool {
    // Check if resource type is in sensitive resources list
}

func (a *Analyzer) IsSensitiveProperty(resourceType string, propertyName string) bool {
    // Check if resource type and property combination is in sensitive properties list
}
```

### Output Formatting

The formatter in `lib/plan/formatter.go` will be enhanced to:

1. Include danger highlight information in the output
2. Apply special formatting (colors, symbols) to highlight dangerous changes

## Data Models

We'll extend the existing models in `lib/plan/models.go`:

```go
type ResourceChange struct {
    // Existing fields...
    
    // New fields for danger highlights
    IsDangerous       bool
    DangerReason      string // e.g., "Sensitive resource replacement", "Sensitive property change"
    DangerProperties  []string // List of dangerous property changes
}
```

## Error Handling

The feature will follow existing error handling patterns:

1. Configuration validation will ensure sensitive resource and property definitions are valid
2. If invalid configurations are detected, appropriate error messages will be returned
3. Errors during analysis will be propagated up the call stack with context

## Testing Strategy

We'll implement the following tests:

1. Unit tests for the new configuration parsing
2. Unit tests for the sensitive resource and property detection methods
3. Integration tests that verify the end-to-end workflow with sample configurations
4. Table-driven tests for various combinations of resource types and properties

Test cases will include:
- Resources that match sensitive resource types
- Resources with properties that match sensitive property definitions
- Resources that don't match any sensitive definitions
- Edge cases like partial matches or case sensitivity

## Implementation Plan

1. Extend the configuration structure in `config/config.go`
2. Update the analyzer in `lib/plan/analyzer.go` to detect sensitive resources and properties
3. Modify the formatter in `lib/plan/formatter.go` to highlight dangerous changes
4. Add unit tests for the new functionality
5. Update documentation to explain the new feature