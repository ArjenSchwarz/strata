# Danger Highlights Feature

## Overview

The Danger Highlights feature enhances Strata's Terraform plan analysis by allowing users to define specific resource types and properties that should trigger warnings when modified. This feature addresses two key scenarios:

1. Warning when sensitive resources are being replaced
2. Warning when sensitive properties are updated, even if the resource itself is not being replaced

## Configuration

Users can define sensitive resources and properties in the configuration file:

```yaml
sensitive_resources:
  - resource_type: aws_rds_instance
  - resource_type: aws_ec2_instance
sensitive_properties:
  - resource_type: aws_ec2_instance
    property: user_data
```

### Sensitive Resources

When a resource of a type listed in `sensitive_resources` is being replaced, it will be flagged as dangerous. This is useful for highlighting changes to critical infrastructure components like databases, where replacements can cause downtime or data loss.

### Sensitive Properties

When a property listed in `sensitive_properties` is changed, the resource will be flagged as dangerous, even if the resource itself is not being replaced. This is useful for highlighting changes to properties that might cause service disruptions, like changes to user data scripts that trigger instance restarts.

## Implementation Details

The feature is implemented across several components:

1. **Configuration**: Extended to include sensitive resources and properties
2. **Analyzer**: Enhanced to detect sensitive resources and properties
3. **Models**: Extended to include danger information
4. **Formatter**: Updated to display danger highlights

### Analyzer

The analyzer checks each resource change against the sensitive resources and properties lists:

- For sensitive resources, it checks if the resource type matches and if the change is a replacement
- For sensitive properties, it checks if the resource type and property name match, and if the property value has changed

### Output

Dangerous changes are highlighted in the output with a warning symbol (⚠️) and a description of why the change is dangerous:

- For sensitive resources: "⚠️ Sensitive resource replacement"
- For sensitive properties: "⚠️ Sensitive property change: property_name"

## Usage

To use this feature, add the sensitive resources and properties to your configuration file, then run Strata as usual:

```bash
strata plan summary terraform.tfplan
```

The output will highlight any dangerous changes according to your configuration.