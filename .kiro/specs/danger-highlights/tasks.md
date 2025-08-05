# Implementation Tasks for Danger Highlights Feature

## Configuration Updates

1. Extend `config/config.go` with new structures:
   - Add `SensitiveResources` and `SensitiveProperties` to the Config struct
   - Create new types for `SensitiveResource` and `SensitiveProperty`
   - Update configuration parsing to handle these new fields

## Analyzer Enhancements

1. Modify `lib/plan/analyzer.go`:
   - Implement `IsSensitiveResource()` method to check if a resource type is sensitive
   - Implement `IsSensitiveProperty()` method to check if a property is sensitive for a given resource type
   - Update the analysis logic to mark resources and changes as dangerous based on these checks

## Model Updates

1. Extend `lib/plan/models.go`:
   - Add danger-related fields to the `ResourceChange` struct
   - Implement helper methods for working with dangerous changes

## Formatter Updates

1. Enhance `lib/plan/formatter.go`:
   - Update output formatting to highlight dangerous changes
   - Add danger information to all supported output formats (table, JSON, etc.)
   - Ensure consistent styling for danger highlights

## Testing

1. Create unit tests:
   - Test configuration parsing for sensitive resources and properties
   - Test sensitive resource and property detection methods
   - Test end-to-end workflow with sample configurations

## Documentation

1. Update documentation:
   - Add information about the new feature to user documentation
   - Document the configuration options for sensitive resources and properties
   - Provide examples of how to use the feature

## Integration

1. Ensure the feature works with all existing functionality:
   - Verify compatibility with different output formats
   - Check integration with the danger threshold feature
   - Test with various Terraform plan formats