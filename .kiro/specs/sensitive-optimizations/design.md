# Sensitive Resource Optimization Feature Design

## Overview

The Sensitive Resource Optimization feature enhances Strata's existing sensitive resource filtering functionality by improving performance, adding configuration validation, and providing more context in error messages. This design addresses the three key requirements: performance optimization, configuration validation, and enhanced error messaging.

## Architecture

The feature will build upon the existing sensitive resource filtering architecture in Strata:

1. **Performance Optimization**: Replace linear lookups with optimized data structures
2. **Configuration Validation**: Add validation at configuration load time
3. **Enhanced Error Messages**: Improve error handling throughout the sensitive resource filtering process

The changes will primarily affect the following components:
- Configuration loading and validation in `config/config.go`
- Sensitive resource and property checking in `lib/plan/analyzer.go`
- Error handling throughout the codebase

## Components and Interfaces

### Configuration Validation

We'll enhance the configuration loading process to validate sensitive resource and property definitions:

```go
// ValidationResult holds validation errors and warnings
type ValidationResult struct {
    Errors   []string
    Warnings []string
}

// New function in config/config.go
func (c *Config) Validate() *ValidationResult {
    result := &ValidationResult{
        Errors:   []string{},
        Warnings: []string{},
    }
    
    // Validate sensitive resources
    resourceTypeMap := make(map[string]bool)
    for i, sr := range c.SensitiveResources {
        if sr.ResourceType == "" {
            result.Errors = append(result.Errors, 
                fmt.Sprintf("sensitive_resources[%d]: resource_type cannot be empty", i))
            continue
        }
        
        if !isValidResourceType(sr.ResourceType) {
            result.Errors = append(result.Errors, 
                fmt.Sprintf("sensitive_resources[%d]: invalid resource_type format: %s (expected format: provider_resource)", i, sr.ResourceType))
            continue
        }
        
        // Check for duplicates - treat as warning since it's not fatal
        if resourceTypeMap[sr.ResourceType] {
            result.Warnings = append(result.Warnings, 
                fmt.Sprintf("sensitive_resources[%d]: duplicate resource_type: %s (previous definition will be ignored)", i, sr.ResourceType))
        }
        resourceTypeMap[sr.ResourceType] = true
    }
    
    // Validate sensitive properties
    propertyMap := make(map[string]bool)
    for i, sp := range c.SensitiveProperties {
        if sp.ResourceType == "" {
            result.Errors = append(result.Errors, 
                fmt.Sprintf("sensitive_properties[%d]: resource_type cannot be empty", i))
            continue
        }
        
        if !isValidResourceType(sp.ResourceType) {
            result.Errors = append(result.Errors, 
                fmt.Sprintf("sensitive_properties[%d]: invalid resource_type format: %s (expected format: provider_resource)", i, sp.ResourceType))
            continue
        }
        
        if sp.Property == "" {
            result.Errors = append(result.Errors, 
                fmt.Sprintf("sensitive_properties[%d]: property cannot be empty", i))
            continue
        }
        
        // Check for duplicates - treat as warning since it's not fatal
        propertyKey := sp.ResourceType + ":" + sp.Property
        if propertyMap[propertyKey] {
            result.Warnings = append(result.Warnings, 
                fmt.Sprintf("sensitive_properties[%d]: duplicate property definition: %s.%s (previous definition will be ignored)", 
                    i, sp.ResourceType, sp.Property))
        }
        propertyMap[propertyKey] = true
    }
    
    return result
}

// Helper method to check if validation failed
func (vr *ValidationResult) HasErrors() bool {
    return len(vr.Errors) > 0
}

// Helper method to check if there are warnings
func (vr *ValidationResult) HasWarnings() bool {
    return len(vr.Warnings) > 0
}

// Helper method to format all issues
func (vr *ValidationResult) FormatIssues() string {
    var issues []string
    
    if len(vr.Errors) > 0 {
        issues = append(issues, "ERRORS:")
        for _, err := range vr.Errors {
            issues = append(issues, "  - "+err)
        }
    }
    
    if len(vr.Warnings) > 0 {
        if len(issues) > 0 {
            issues = append(issues, "")
        }
        issues = append(issues, "WARNINGS:")
        for _, warn := range vr.Warnings {
            issues = append(issues, "  - "+warn)
        }
    }
    
    return strings.Join(issues, "\n")
}

// Helper function to validate resource type format
func isValidResourceType(resourceType string) bool {
    // Basic validation: non-empty, contains provider prefix (e.g., aws_)
    return resourceType != "" && strings.Contains(resourceType, "_")
}
```

### Performance Optimization

We'll optimize the sensitive resource and property lookups by using maps instead of linear searches:

```go
// New fields in Analyzer struct
type Analyzer struct {
    // Existing fields
    plan   *tfjson.Plan
    config *config.Config
    
    // New fields for optimized lookups
    sensitiveResourceMap map[string]bool
    sensitivePropertyMap map[string]map[string]bool
}

// Modified NewAnalyzer function
func NewAnalyzer(plan *tfjson.Plan, cfg *config.Config) *Analyzer {
    analyzer := &Analyzer{
        plan:   plan,
        config: cfg,
    }
    
    // Initialize lookup maps
    analyzer.buildLookupMaps()
    
    return analyzer
}

// New method to build lookup maps
func (a *Analyzer) buildLookupMaps() {
    // Initialize maps
    a.sensitiveResourceMap = make(map[string]bool)
    a.sensitivePropertyMap = make(map[string]map[string]bool)
    
    // Skip if no config
    if a.config == nil {
        return
    }
    
    // Build sensitive resource map
    for _, sr := range a.config.SensitiveResources {
        a.sensitiveResourceMap[sr.ResourceType] = true
    }
    
    // Build sensitive property map
    for _, sp := range a.config.SensitiveProperties {
        if _, exists := a.sensitivePropertyMap[sp.ResourceType]; !exists {
            a.sensitivePropertyMap[sp.ResourceType] = make(map[string]bool)
        }
        a.sensitivePropertyMap[sp.ResourceType][sp.Property] = true
    }
}

// Optimized IsSensitiveResource method
func (a *Analyzer) IsSensitiveResource(resourceType string) bool {
    return a.sensitiveResourceMap[resourceType]
}

// Optimized IsSensitiveProperty method
func (a *Analyzer) IsSensitiveProperty(resourceType string, propertyName string) bool {
    if propertyMap, exists := a.sensitivePropertyMap[resourceType]; exists {
        return propertyMap[propertyName]
    }
    return false
}
```

### Enhanced Error Messages

We'll improve error handling throughout the sensitive resource filtering process:

```go
// Enhanced checkSensitiveProperties method with better error handling
func (a *Analyzer) checkSensitiveProperties(change *tfjson.ResourceChange) ([]string, error) {
    var sensitiveProps []string
    
    // If there's no change or no config, return empty
    if change.Change.Before == nil || change.Change.After == nil || a.config == nil {
        return sensitiveProps, nil
    }
    
    // Extract before and after as maps
    beforeMap, beforeOk := change.Change.Before.(map[string]interface{})
    afterMap, afterOk := change.Change.After.(map[string]interface{})
    
    if !beforeOk {
        return nil, fmt.Errorf("failed to process resource %s: Before state is not a valid map", change.Address)
    }
    
    if !afterOk {
        return nil, fmt.Errorf("failed to process resource %s: After state is not a valid map", change.Address)
    }
    
    // Check each property to see if it's changed and if it's sensitive
    for propName := range afterMap {
        // Skip if property doesn't exist in before (new property)
        beforeVal, exists := beforeMap[propName]
        if !exists {
            continue
        }
        
        afterVal := afterMap[propName]
        
        // If property has changed and is sensitive, add to list
        equal, err := equals(beforeVal, afterVal)
        if err != nil {
            return nil, fmt.Errorf("failed to compare property %s in resource %s: %w", 
                propName, change.Address, err)
        }
        
        if !equal && a.IsSensitiveProperty(change.Type, propName) {
            sensitiveProps = append(sensitiveProps, propName)
        }
    }
    
    return sensitiveProps, nil
}

// Enhanced equals function with error handling
func equals(a, b interface{}) (bool, error) {
    // Handle nil cases
    if a == nil && b == nil {
        return true, nil
    }
    if a == nil || b == nil {
        return false, nil
    }
    
    // Handle maps specially since they're not directly comparable
    aMap, aIsMap := a.(map[string]interface{})
    bMap, bIsMap := b.(map[string]interface{})
    
    if aIsMap && bIsMap {
        // If maps have different lengths, they're not equal
        if len(aMap) != len(bMap) {
            return false, nil
        }
        
        // Check each key-value pair
        for k, aVal := range aMap {
            bVal, exists := bMap[k]
            if !exists {
                return false, nil
            }
            
            // Recursively compare values
            equal, err := equals(aVal, bVal)
            if err != nil {
                return false, fmt.Errorf("error comparing map values for key %s: %w", k, err)
            }
            
            if !equal {
                return false, nil
            }
        }
        return true, nil
    }
    
    // Handle slices specially since they're not directly comparable
    aSlice, aIsSlice := a.([]interface{})
    bSlice, bIsSlice := b.([]interface{})
    
    if aIsSlice && bIsSlice {
        // If slices have different lengths, they're not equal
        if len(aSlice) != len(bSlice) {
            return false, nil
        }
        
        // Check each element
        for i, aVal := range aSlice {
            bVal := bSlice[i]
            // Recursively compare values
            equal, err := equals(aVal, bVal)
            if err != nil {
                return false, fmt.Errorf("error comparing slice values at index %d: %w", i, err)
            }
            
            if !equal {
                return false, nil
            }
        }
        return true, nil
    }
    
    // For non-map and non-slice types, use direct comparison
    return a == b, nil
}

// AnalysisError represents an error that occurred during analysis
type AnalysisError struct {
    ResourceAddress string
    ErrorType       string
    Message         string
    Cause           error
}

func (ae AnalysisError) Error() string {
    if ae.Cause != nil {
        return fmt.Sprintf("%s error for resource %s: %s (caused by: %v)", 
            ae.ErrorType, ae.ResourceAddress, ae.Message, ae.Cause)
    }
    return fmt.Sprintf("%s error for resource %s: %s", 
        ae.ErrorType, ae.ResourceAddress, ae.Message)
}

// AnalysisResult holds the results and any errors encountered during analysis
type AnalysisResult struct {
    Changes []ResourceChange
    Errors  []AnalysisError
}

// Enhanced analyzeResourceChanges method with error aggregation
func (a *Analyzer) analyzeResourceChanges() (*AnalysisResult, error) {
    if a.plan.ResourceChanges == nil {
        return &AnalysisResult{Changes: []ResourceChange{}, Errors: []AnalysisError{}}, nil
    }
    
    result := &AnalysisResult{
        Changes: make([]ResourceChange, 0, len(a.plan.ResourceChanges)),
        Errors:  []AnalysisError{},
    }
    
    for _, rc := range a.plan.ResourceChanges {
        changeType := FromTerraformAction(rc.Change.Actions)
        replacementType, err := a.analyzeReplacementNecessity(rc)
        if err != nil {
            // Log error but continue processing other resources
            result.Errors = append(result.Errors, AnalysisError{
                ResourceAddress: rc.Address,
                ErrorType:       "ReplacementAnalysis",
                Message:         "failed to analyze replacement necessity",
                Cause:           err,
            })
            // Use default replacement type and continue
            replacementType = ReplacementTypeNone
        }
        
        change := ResourceChange{
            Address:          rc.Address,
            Type:             rc.Type,
            Name:             rc.Name,
            ChangeType:       changeType,
            IsDestructive:    changeType.IsDestructive(),
            ReplacementType:  replacementType,
            PhysicalID:       a.extractPhysicalID(rc),
            PlannedID:        a.extractPlannedID(rc),
            ModulePath:       a.extractModulePath(rc.Address),
            ChangeAttributes: a.getChangingAttributes(rc),
            Before:           rc.Change.Before,
            After:            rc.Change.After,
            IsDangerous:      false,
            DangerReason:     "",
            DangerProperties: []string{},
            ErrorContext:     "",
        }
        
        // Check if this is a sensitive resource
        if a.IsSensitiveResource(rc.Type) && changeType == ChangeTypeReplace {
            change.IsDangerous = true
            change.DangerReason = "Sensitive resource replacement"
        }
        
        // Check for sensitive properties
        dangerProps, err := a.checkSensitiveProperties(rc)
        if err != nil {
            // Log error but continue processing
            result.Errors = append(result.Errors, AnalysisError{
                ResourceAddress: rc.Address,
                ErrorType:       "SensitivePropertyCheck",
                Message:         "failed to check sensitive properties",
                Cause:           err,
            })
            change.ErrorContext = "Error checking sensitive properties: " + err.Error()
        } else if len(dangerProps) > 0 {
            change.IsDangerous = true
            change.DangerProperties = dangerProps
            if change.DangerReason == "" {
                change.DangerReason = "Sensitive property change"
            } else {
                change.DangerReason += " and sensitive property change"
            }
        }
        
        result.Changes = append(result.Changes, change)
    }
    
    return result, nil
}
```

### Integration with Configuration Loading

We'll integrate the configuration validation with the configuration loading process:

```go
// In cmd/root.go or wherever configuration is loaded
func initConfig() {
    // Existing configuration loading code...
    
    // After loading configuration, validate it
    validationResult := config.Validate()
    
    // Handle validation results
    if validationResult.HasWarnings() {
        fmt.Fprintf(os.Stderr, "Configuration warnings:\n%s\n", 
            formatWarnings(validationResult.Warnings))
    }
    
    if validationResult.HasErrors() {
        fmt.Fprintf(os.Stderr, "Configuration validation failed:\n%s\n", 
            formatErrors(validationResult.Errors))
        os.Exit(1)
    }
}

// Helper functions for formatting validation messages
func formatErrors(errors []string) string {
    var formatted []string
    for i, err := range errors {
        formatted = append(formatted, fmt.Sprintf("%d. %s", i+1, err))
    }
    return strings.Join(formatted, "\n")
}

func formatWarnings(warnings []string) string {
    var formatted []string
    for i, warn := range warnings {
        formatted = append(formatted, fmt.Sprintf("%d. %s", i+1, warn))
    }
    return strings.Join(formatted, "\n")
}
```

## Data Models

We'll extend the existing models to support the new functionality:

```go
// Enhanced ResourceChange struct with structured error context
type ResourceChange struct {
    // Existing fields...
    
    // Enhanced error context for debugging and user feedback
    ErrorContext    string            `json:"error_context,omitempty"`
    ErrorSeverity   ErrorSeverity     `json:"error_severity,omitempty"`
    ErrorDetails    map[string]string `json:"error_details,omitempty"`
    SuggestedFix    string            `json:"suggested_fix,omitempty"`
}

// ErrorSeverity indicates the severity level of errors
type ErrorSeverity string

const (
    ErrorSeverityLow      ErrorSeverity = "low"
    ErrorSeverityMedium   ErrorSeverity = "medium"
    ErrorSeverityHigh     ErrorSeverity = "high"
    ErrorSeverityCritical ErrorSeverity = "critical"
)

// ErrorContextBuilder helps build structured error contexts
type ErrorContextBuilder struct {
    resourceAddress string
    errorType       string
    details         map[string]string
}

func NewErrorContextBuilder(resourceAddress, errorType string) *ErrorContextBuilder {
    return &ErrorContextBuilder{
        resourceAddress: resourceAddress,
        errorType:       errorType,
        details:         make(map[string]string),
    }
}

func (ecb *ErrorContextBuilder) AddDetail(key, value string) *ErrorContextBuilder {
    ecb.details[key] = value
    return ecb
}

func (ecb *ErrorContextBuilder) BuildContext() string {
    var parts []string
    parts = append(parts, fmt.Sprintf("Resource: %s", ecb.resourceAddress))
    parts = append(parts, fmt.Sprintf("Error Type: %s", ecb.errorType))
    
    for key, value := range ecb.details {
        parts = append(parts, fmt.Sprintf("%s: %s", key, value))
    }
    
    return strings.Join(parts, " | ")
}

func (ecb *ErrorContextBuilder) BuildSuggestedFix(errorType string) string {
    switch errorType {
    case "SensitivePropertyCheck":
        return "Verify that the resource properties in your Terraform plan are valid and properly formatted"
    case "ReplacementAnalysis":
        return "Check if the resource configuration has breaking changes that require replacement"
    case "ConfigurationValidation":
        return "Review your strata.yaml configuration file for syntax errors or invalid resource type definitions"
    default:
        return "Review the error details and consult the documentation for troubleshooting steps"
    }
}
```

## Error Handling

The feature will follow these error handling principles:

1. **Validation Errors**: Configuration validation errors will be aggregated and reported together
2. **Processing Errors**: Errors during plan analysis will include context about the specific resource or property
3. **Error Propagation**: Errors will be propagated up the call stack with proper context
4. **User-Friendly Messages**: Error messages will be formatted for readability and actionability

## Testing Strategy

We'll implement the following tests:

1. **Performance Tests**:
   - Benchmark tests for IsSensitiveResource and IsSensitiveProperty methods
   - Comparison of performance before and after optimization
   - Tests with large numbers of sensitive resources and properties
   - **Performance Validation Framework**: Automated tests to ensure the 500ms requirement is met for 100+ resource plans

2. **Validation Tests**:
   - Unit tests for configuration validation
   - Tests for various invalid configurations
   - Tests for duplicate detection

3. **Error Handling Tests**:
   - Tests for error messages in various scenarios
   - Tests for error propagation
   - Tests for error aggregation

Test cases will include:
- Valid and invalid resource type formats
- Duplicate sensitive resource and property definitions
- Edge cases in property comparison
- Performance with large Terraform plans
- Various error scenarios with expected error messages
- **Integration tests** with real Terraform plan files
- **Edge case validation** for malformed configurations and plan files
- **Error aggregation scenarios** with multiple simultaneous failures
- **Configuration migration testing** for backward compatibility

### Performance Testing Framework

To ensure we meet the performance requirements, we'll implement a comprehensive performance testing framework:

```go
// Performance testing utilities in lib/plan/analyzer_performance_test.go
func BenchmarkSensitiveResourceFiltering(b *testing.B) {
    // Test with varying plan sizes
    planSizes := []int{50, 100, 200, 500, 1000}
    
    for _, size := range planSizes {
        b.Run(fmt.Sprintf("Plan_%d_resources", size), func(b *testing.B) {
            plan := generateMockPlan(size)
            config := generateSensitiveConfig(10) // 10 sensitive resources
            
            b.ResetTimer()
            b.StartTimer()
            
            for i := 0; i < b.N; i++ {
                analyzer := NewAnalyzer(plan, config)
                _, err := analyzer.Analyze()
                if err != nil {
                    b.Fatalf("Analysis failed: %v", err)
                }
            }
            
            b.StopTimer()
        })
    }
}

// Performance validation test to ensure 500ms requirement
func TestPerformanceRequirement(t *testing.T) {
    // Generate a plan with 100+ resources
    plan := generateMockPlan(150)
    config := generateSensitiveConfig(20)
    
    analyzer := NewAnalyzer(plan, config)
    
    start := time.Now()
    _, err := analyzer.Analyze()
    duration := time.Since(start)
    
    if err != nil {
        t.Fatalf("Analysis failed: %v", err)
    }
    
    // Verify performance requirement
    maxDuration := 500 * time.Millisecond
    if duration > maxDuration {
        t.Errorf("Performance requirement not met: analysis took %v, expected < %v", 
            duration, maxDuration)
    }
    
    t.Logf("Analysis completed in %v (requirement: < %v)", duration, maxDuration)
}

// Benchmark comparison test for before/after optimization
func BenchmarkSensitiveResourceLookup(b *testing.B) {
    resourceTypes := generateResourceTypes(1000)
    sensitiveTypes := resourceTypes[:50] // 50 sensitive out of 1000
    
    b.Run("LinearSearch", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            for _, rt := range resourceTypes {
                _ = linearSearchSensitive(sensitiveTypes, rt)
            }
        }
    })
    
    b.Run("MapLookup", func(b *testing.B) {
        sensitiveMap := buildSensitiveMap(sensitiveTypes)
        b.ResetTimer()
        
        for i := 0; i < b.N; i++ {
            for _, rt := range resourceTypes {
                _ = sensitiveMap[rt]
            }
        }
    })
}

// Helper functions for generating test data
func generateMockPlan(resourceCount int) *tfjson.Plan {
    // Generate a realistic Terraform plan with specified number of resources
    // Include mix of creates, updates, deletes, and replaces
}

func generateSensitiveConfig(sensitiveCount int) *config.Config {
    // Generate configuration with specified number of sensitive resources/properties
}

func generateResourceTypes(count int) []string {
    // Generate realistic resource type names
}
```

### Performance Monitoring Integration

We'll also add runtime performance monitoring to track actual performance in production:

```go
// Performance monitoring in analyzer.go
func (a *Analyzer) Analyze() (*PlanSummary, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        resourceCount := len(a.plan.ResourceChanges)
        
        // Log performance metrics for monitoring
        log.Printf("Analysis completed: %d resources in %v", resourceCount, duration)
        
        // Optional: emit metrics to monitoring system
        if duration > 500*time.Millisecond && resourceCount > 100 {
            log.Printf("WARNING: Performance requirement not met: %d resources took %v", 
                resourceCount, duration)
        }
    }()
    
    // Existing analysis logic...
}
```

### Integration and Edge Case Testing

We'll add comprehensive integration and edge case testing:

```go
// Integration tests in lib/plan/analyzer_integration_test.go
func TestRealTerraformPlanIntegration(t *testing.T) {
    testCases := []struct {
        name         string
        planFile     string
        configFile   string
        expectErrors bool
        expectCount  int
    }{
        {
            name:         "AWS Infrastructure Plan",
            planFile:     "testdata/aws_infrastructure.tfplan",
            configFile:   "testdata/aws_sensitive_config.yaml",
            expectErrors: false,
            expectCount:  25,
        },
        {
            name:         "Large Multi-Provider Plan",
            planFile:     "testdata/large_multi_provider.tfplan", 
            configFile:   "testdata/multi_provider_config.yaml",
            expectErrors: false,
            expectCount:  150,
        },
        {
            name:         "Malformed Plan File",
            planFile:     "testdata/malformed.tfplan",
            configFile:   "testdata/basic_config.yaml",
            expectErrors: true,
            expectCount:  0,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Load real plan and config files
            plan, err := loadTerraformPlan(tc.planFile)
            if tc.expectErrors && err != nil {
                return // Expected error case
            }
            require.NoError(t, err)
            
            config, err := loadConfig(tc.configFile)
            require.NoError(t, err)
            
            analyzer := NewAnalyzer(plan, config)
            result, err := analyzer.Analyze()
            
            if tc.expectErrors {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Len(t, result.ResourceChanges, tc.expectCount)
            }
        })
    }
}

// Edge case testing for configuration validation
func TestConfigurationEdgeCases(t *testing.T) {
    testCases := []struct {
        name            string
        config          *Config
        expectErrors    int
        expectWarnings  int
        expectedMessage string
    }{
        {
            name: "Empty Resource Types",
            config: &Config{
                SensitiveResources: []SensitiveResource{
                    {ResourceType: ""},
                    {ResourceType: "aws_instance"},
                },
            },
            expectErrors:    1,
            expectWarnings:  0,
            expectedMessage: "resource_type cannot be empty",
        },
        {
            name: "Invalid Resource Type Format", 
            config: &Config{
                SensitiveResources: []SensitiveResource{
                    {ResourceType: "invalid-format"},
                    {ResourceType: "aws_instance"},
                },
            },
            expectErrors:    1,
            expectWarnings:  0,
            expectedMessage: "invalid resource_type format",
        },
        {
            name: "Duplicate Definitions",
            config: &Config{
                SensitiveResources: []SensitiveResource{
                    {ResourceType: "aws_instance"},
                    {ResourceType: "aws_instance"},
                },
            },
            expectErrors:    0,
            expectWarnings:  1,
            expectedMessage: "duplicate resource_type",
        },
        {
            name: "Mixed Valid and Invalid",
            config: &Config{
                SensitiveResources: []SensitiveResource{
                    {ResourceType: ""},
                    {ResourceType: "aws_instance"},
                    {ResourceType: "aws_instance"}, // duplicate
                },
                SensitiveProperties: []SensitiveProperty{
                    {ResourceType: "aws_instance", Property: ""},
                    {ResourceType: "aws_instance", Property: "user_data"},
                },
            },
            expectErrors:    2, // empty resource_type + empty property
            expectWarnings:  1, // duplicate resource_type
            expectedMessage: "",
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := tc.config.Validate()
            
            assert.Len(t, result.Errors, tc.expectErrors, "Expected %d errors, got %d", tc.expectErrors, len(result.Errors))
            assert.Len(t, result.Warnings, tc.expectWarnings, "Expected %d warnings, got %d", tc.expectWarnings, len(result.Warnings))
            
            if tc.expectedMessage != "" {
                allMessages := strings.Join(append(result.Errors, result.Warnings...), " ")
                assert.Contains(t, allMessages, tc.expectedMessage)
            }
        })
    }
}

// Error aggregation testing
func TestErrorAggregationScenarios(t *testing.T) {
    // Create a plan with multiple problematic resources
    plan := &tfjson.Plan{
        ResourceChanges: []*tfjson.ResourceChange{
            // Resource with invalid before/after states
            {
                Address: "aws_instance.invalid_state",
                Type:    "aws_instance", 
                Change: &tfjson.Change{
                    Before: "invalid_json", // This should cause parsing errors
                    After:  map[string]interface{}{"user_data": "new_value"},
                },
            },
            // Resource with valid state changes
            {
                Address: "aws_instance.valid",
                Type:    "aws_instance",
                Change: &tfjson.Change{
                    Before: map[string]interface{}{"user_data": "old_value"},
                    After:  map[string]interface{}{"user_data": "new_value"},
                },
            },
            // Another problematic resource
            {
                Address: "aws_rds_instance.problematic",
                Type:    "aws_rds_instance",
                Change: &tfjson.Change{
                    Before: nil,
                    After:  "invalid_json",
                },
            },
        },
    }
    
    config := &Config{
        SensitiveResources: []SensitiveResource{{ResourceType: "aws_rds_instance"}},
        SensitiveProperties: []SensitiveProperty{{ResourceType: "aws_instance", Property: "user_data"}},
    }
    
    analyzer := NewAnalyzer(plan, config)
    result, err := analyzer.analyzeResourceChanges()
    
    // Should not fail completely, but should collect errors
    assert.NoError(t, err, "Analysis should not fail completely due to individual resource errors")
    assert.NotNil(t, result)
    assert.Greater(t, len(result.Errors), 0, "Should have collected errors from problematic resources")
    assert.Greater(t, len(result.Changes), 0, "Should still process valid resources")
    
    // Verify error details
    errorTypes := make(map[string]int)
    for _, analysisErr := range result.Errors {
        errorTypes[analysisErr.ErrorType]++
    }
    
    assert.Contains(t, errorTypes, "SensitivePropertyCheck", "Should have sensitive property check errors")
}