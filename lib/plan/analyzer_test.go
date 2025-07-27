package plan

import (
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

func TestIsSensitiveResource(t *testing.T) {
	// Create a test config with sensitive resources
	cfg := &config.Config{
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_instance"},
			{ResourceType: "aws_ec2_instance"},
		},
	}

	// Create analyzer with the config
	analyzer := &Analyzer{
		config: cfg,
	}

	// Test cases
	testCases := []struct {
		name         string
		resourceType string
		expected     bool
	}{
		{
			name:         "Sensitive resource should return true",
			resourceType: "aws_rds_instance",
			expected:     true,
		},
		{
			name:         "Another sensitive resource should return true",
			resourceType: "aws_ec2_instance",
			expected:     true,
		},
		{
			name:         "Non-sensitive resource should return false",
			resourceType: "aws_s3_bucket",
			expected:     false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.IsSensitiveResource(tc.resourceType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsSensitiveProperty(t *testing.T) {
	// Create a test config with sensitive properties
	cfg := &config.Config{
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_ec2_instance", Property: "user_data"},
			{ResourceType: "aws_lambda_function", Property: "source_code_hash"},
		},
	}

	// Create analyzer with the config
	analyzer := &Analyzer{
		config: cfg,
	}

	// Test cases
	testCases := []struct {
		name         string
		resourceType string
		property     string
		expected     bool
	}{
		{
			name:         "Sensitive property should return true",
			resourceType: "aws_ec2_instance",
			property:     "user_data",
			expected:     true,
		},
		{
			name:         "Another sensitive property should return true",
			resourceType: "aws_lambda_function",
			property:     "source_code_hash",
			expected:     true,
		},
		{
			name:         "Non-sensitive property should return false",
			resourceType: "aws_ec2_instance",
			property:     "instance_type",
			expected:     false,
		},
		{
			name:         "Sensitive property on wrong resource should return false",
			resourceType: "aws_s3_bucket",
			property:     "user_data",
			expected:     false,
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.IsSensitiveProperty(tc.resourceType, tc.property)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCheckSensitiveProperties(t *testing.T) {
	// Create a test config with sensitive properties
	cfg := &config.Config{
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_ec2_instance", Property: "user_data"},
		},
	}

	// Create analyzer with the config
	analyzer := &Analyzer{
		config: cfg,
	}

	// Create a test resource change
	resourceChange := &tfjson.ResourceChange{
		Type: "aws_ec2_instance",
		Change: &tfjson.Change{
			Before: map[string]interface{}{
				"user_data":     "old-data",
				"instance_type": "t2.micro",
			},
			After: map[string]interface{}{
				"user_data":     "new-data",
				"instance_type": "t2.micro",
			},
		},
	}

	// Test the function
	result := analyzer.checkSensitiveProperties(resourceChange)

	// Should find one sensitive property change
	assert.Len(t, result, 1)
	assert.Contains(t, result, "user_data")

	// Test with unchanged sensitive property
	resourceChange.Change.After.(map[string]interface{})["user_data"] = "old-data"
	result = analyzer.checkSensitiveProperties(resourceChange)

	// Should find no sensitive property changes
	assert.Len(t, result, 0)
}

func TestAnalyzeReplacementNecessity(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		change   *tfjson.ResourceChange
		expected ReplacementType
	}{
		{
			name: "Non-destructive change should be never",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
				},
			},
			expected: ReplacementNever,
		},
		{
			name: "Delete operation should be never",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete},
				},
			},
			expected: ReplacementNever,
		},
		{
			name: "Replace without ReplacePaths should be always",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
				},
			},
			expected: ReplacementAlways,
		},
		{
			name: "Replace with empty ReplacePaths should be always",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions:      tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
					ReplacePaths: []interface{}{},
				},
			},
			expected: ReplacementAlways,
		},
		{
			name: "Replace with ReplacePaths and definite values should be always",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions:      tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
					ReplacePaths: []interface{}{[]interface{}{"definite_field"}},
					After: map[string]interface{}{
						"definite_field": "definite_value",
					},
				},
			},
			expected: ReplacementAlways,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.analyzeReplacementNecessity(tc.change)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateStatistics(t *testing.T) {
	// Create a test config with sensitive resources
	cfg := &config.Config{
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_instance"},
			{ResourceType: "aws_dynamodb_table"},
		},
	}

	// Create analyzer with the config
	analyzer := &Analyzer{
		config: cfg,
	}

	testCases := []struct {
		name     string
		changes  []ResourceChange
		expected ChangeStatistics
	}{
		{
			name:    "Empty changes should have all zeros",
			changes: []ResourceChange{},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        0,
			},
		},
		{
			name: "Single create should increment ToAdd and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate, ReplacementType: ReplacementNever},
			},
			expected: ChangeStatistics{
				ToAdd:        1,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Single update should increment ToChange and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeUpdate, ReplacementType: ReplacementNever},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     1,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Single delete should increment ToDestroy and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeDelete, ReplacementType: ReplacementNever},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    1,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Replace with always replacement should increment Replacements and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 1,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Dangerous sensitive resource should increment HighRisk",
			changes: []ResourceChange{
				{
					Type:        "aws_rds_instance",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 1,
				HighRisk:     1,
				Total:        1,
			},
		},
		{
			name: "Non-dangerous sensitive resource should not increment HighRisk",
			changes: []ResourceChange{
				{
					Type:        "aws_rds_instance",
					ChangeType:  ChangeTypeUpdate,
					IsDangerous: false,
				},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     1,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Dangerous non-sensitive resource should increment HighRisk",
			changes: []ResourceChange{
				{
					Type:        "aws_s3_bucket",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 1,
				HighRisk:     1,
				Total:        1,
			},
		},
		{
			name: "Mixed changes should calculate correctly",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeUpdate, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeDelete, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
				{
					Type:        "aws_rds_instance",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
				{
					Type:        "aws_dynamodb_table",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			expected: ChangeStatistics{
				ToAdd:        1,
				ToChange:     1,
				ToDestroy:    1,
				Replacements: 3,
				HighRisk:     2,
				Total:        6,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.calculateStatistics(tc.changes)
			assert.Equal(t, tc.expected.ToAdd, result.ToAdd, "ToAdd mismatch")
			assert.Equal(t, tc.expected.ToChange, result.ToChange, "ToChange mismatch")
			assert.Equal(t, tc.expected.ToDestroy, result.ToDestroy, "ToDestroy mismatch")
			assert.Equal(t, tc.expected.Replacements, result.Replacements, "Replacements mismatch")
			assert.Equal(t, tc.expected.HighRisk, result.HighRisk, "HighRisk mismatch")
			assert.Equal(t, tc.expected.Total, result.Total, "Total mismatch")
		})
	}
}

func TestExtractPhysicalID(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		change   *tfjson.ResourceChange
		expected string
	}{
		{
			name: "New resource should return dash",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: nil,
				},
			},
			expected: "-",
		},
		{
			name: "Existing resource with ID should return ID",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"id": "resource-123",
					},
				},
			},
			expected: "resource-123",
		},
		{
			name: "Existing resource without ID should return dash",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"name": "resource-name",
					},
				},
			},
			expected: "-",
		},
		{
			name: "Existing resource with empty ID should return dash",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"id": "",
					},
				},
			},
			expected: "-",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.extractPhysicalID(tc.change)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractModulePath(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "Root resource should return dash",
			address:  "aws_instance.web",
			expected: "-",
		},
		{
			name:     "Single module should return module name",
			address:  "module.web.aws_instance.server",
			expected: "web",
		},
		{
			name:     "Nested modules should return path",
			address:  "module.app.module.storage.aws_s3_bucket.data",
			expected: "app/storage",
		},
		{
			name:     "Complex nested path should parse correctly",
			address:  "module.infrastructure.module.vpc.module.subnets.aws_subnet.private",
			expected: "infrastructure/vpc/subnets",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.extractModulePath(tc.address)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractProvider(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "AWS resource should extract aws",
			resourceType: "aws_s3_bucket",
			expected:     "aws",
		},
		{
			name:         "AWS EC2 resource should extract aws",
			resourceType: "aws_ec2_instance",
			expected:     "aws",
		},
		{
			name:         "Azure resource should extract azurerm",
			resourceType: "azurerm_virtual_machine",
			expected:     "azurerm",
		},
		{
			name:         "Google resource should extract google",
			resourceType: "google_compute_instance",
			expected:     "google",
		},
		{
			name:         "Kubernetes resource should extract kubernetes",
			resourceType: "kubernetes_deployment",
			expected:     "kubernetes",
		},
		{
			name:         "HashiCorp Vault resource should extract vault",
			resourceType: "vault_policy",
			expected:     "vault",
		},
		{
			name:         "Resource without underscore should return as-is",
			resourceType: "data",
			expected:     "data",
		},
		{
			name:         "Empty string should return unknown",
			resourceType: "",
			expected:     "unknown",
		},
		{
			name:         "Single underscore should extract first part",
			resourceType: "provider_",
			expected:     "provider",
		},
		{
			name:         "Complex resource type should extract first part",
			resourceType: "aws_db_parameter_group",
			expected:     "aws",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.extractProvider(tc.resourceType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractProviderCaching(t *testing.T) {
	analyzer := &Analyzer{}

	// Test that caching works by calling the same resource type multiple times
	resourceType := "aws_s3_bucket"

	// First call should compute and cache the result
	result1 := analyzer.extractProvider(resourceType)
	assert.Equal(t, "aws", result1)

	// Second call should return cached result
	result2 := analyzer.extractProvider(resourceType)
	assert.Equal(t, "aws", result2)

	// Verify cache contains the entry
	cached, ok := analyzer.providerCache.Load(resourceType)
	assert.True(t, ok, "Cache should contain the entry")
	assert.Equal(t, "aws", cached.(string))
}

func TestExtractProviderEdgeCases(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "Resource starting with underscore",
			resourceType: "_resource_type",
			expected:     "unknown",
		},
		{
			name:         "Resource with multiple consecutive underscores",
			resourceType: "aws__s3__bucket",
			expected:     "aws",
		},
		{
			name:         "Resource ending with underscore",
			resourceType: "aws_s3_bucket_",
			expected:     "aws",
		},
		{
			name:         "Single character provider",
			resourceType: "a_resource",
			expected:     "a",
		},
		{
			name:         "Numeric provider",
			resourceType: "123_resource",
			expected:     "123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.extractProvider(tc.resourceType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractReplacementHints(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		change   *tfjson.ResourceChange
		expected []string
	}{
		{
			name: "No replacement paths should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: nil,
				},
			},
			expected: []string{},
		},
		{
			name: "Empty replacement paths should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []interface{}{},
				},
			},
			expected: []string{},
		},
		{
			name: "Simple string path should be formatted",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []interface{}{"subnet_id"},
				},
			},
			expected: []string{"subnet_id"},
		},
		{
			name: "Nested array path should be formatted with dots and brackets",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []interface{}{
						[]interface{}{"network_interface", 0, "subnet_id"},
					},
				},
			},
			expected: []string{"network_interface.[0].subnet_id"},
		},
		{
			name: "Multiple replacement paths should all be included",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []interface{}{
						"subnet_id",
						[]interface{}{"security_groups", 1},
						"availability_zone",
					},
				},
			},
			expected: []string{
				"subnet_id",
				"security_groups.[1]",
				"availability_zone",
			},
		},
		{
			name: "Float64 indices should be converted to int",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []interface{}{
						[]interface{}{"network_interface", 0.0, "subnet_id"},
					},
				},
			},
			expected: []string{"network_interface.[0].subnet_id"},
		},
		{
			name: "Nil change should return empty",
			change: &tfjson.ResourceChange{
				Change: nil,
			},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.extractReplacementHints(tc.change)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatReplacePath(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		path     interface{}
		expected string
	}{
		{
			name:     "Simple string should return as-is",
			path:     "subnet_id",
			expected: "subnet_id",
		},
		{
			name:     "Array with string should format with dots",
			path:     []interface{}{"network_interface", "subnet_id"},
			expected: "network_interface.subnet_id",
		},
		{
			name:     "Array with int should format with brackets",
			path:     []interface{}{"security_groups", 0},
			expected: "security_groups.[0]",
		},
		{
			name:     "Array with float64 should format with brackets",
			path:     []interface{}{"security_groups", 1.0},
			expected: "security_groups.[1]",
		},
		{
			name:     "Complex nested path should format correctly",
			path:     []interface{}{"block_device_mappings", 0, "ebs", "volume_size"},
			expected: "block_device_mappings.[0].ebs.volume_size",
		},
		{
			name:     "Empty array should return empty string",
			path:     []interface{}{},
			expected: "",
		},
		{
			name:     "Unsupported type should return empty string",
			path:     123,
			expected: "",
		},
		{
			name:     "Nil should return empty string",
			path:     nil,
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.formatReplacePath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetTopChangedProperties(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowContext: true,
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name     string
		change   *tfjson.ResourceChange
		limit    int
		expected []string
	}{
		{
			name: "ShowContext disabled should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]interface{}{
						"instance_type": "t2.small",
						"ami":           "ami-456",
					},
				},
			},
			limit:    3,
			expected: []string{},
		},
		{
			name: "Non-update operation should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
					Before:  nil,
					After: map[string]interface{}{
						"instance_type": "t2.micro",
					},
				},
			},
			limit:    3,
			expected: []string{},
		},
		{
			name: "Changed properties should be detected",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"instance_type":      "t2.micro",
						"ami":                "ami-123",
						"security_group_ids": []interface{}{"sg-123"},
						"unchanged_property": "same",
					},
					After: map[string]interface{}{
						"instance_type":      "t2.small",
						"ami":                "ami-456",
						"security_group_ids": []interface{}{"sg-456"},
						"unchanged_property": "same",
					},
				},
			},
			limit:    3,
			expected: []string{"instance_type", "ami", "security_group_ids"},
		},
		{
			name: "Limit should be respected",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"prop1": "old1",
						"prop2": "old2",
						"prop3": "old3",
						"prop4": "old4",
					},
					After: map[string]interface{}{
						"prop1": "new1",
						"prop2": "new2",
						"prop3": "new3",
						"prop4": "new4",
					},
				},
			},
			limit:    2,
			expected: []string{}, // We'll check length separately since map iteration order is not guaranteed
		},
		{
			name: "Removed properties should be detected",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"existing_prop": "value",
						"removed_prop":  "old_value",
					},
					After: map[string]interface{}{
						"existing_prop": "value",
					},
				},
			},
			limit:    3,
			expected: []string{"removed_prop (removed)"},
		},
		{
			name: "No changes should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]interface{}{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
				},
			},
			limit:    3,
			expected: []string{},
		},
		{
			name: "Nil before/after should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before:  nil,
					After:   nil,
				},
			},
			limit:    3,
			expected: []string{},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Only enable ShowContext for tests that expect results
			if i == 0 {
				analyzer.config.Plan.ShowContext = false
			} else {
				analyzer.config.Plan.ShowContext = true
			}

			result := analyzer.getTopChangedProperties(tc.change, tc.limit)

			// Special handling for limit test case
			if tc.name == "Limit should be respected" {
				assert.Len(t, result, tc.limit, "Should respect the limit")
				// Check that all returned properties are valid (from the test data)
				validProps := []string{"prop1", "prop2", "prop3", "prop4"}
				for _, prop := range result {
					assert.Contains(t, validProps, prop, "Returned property should be valid")
				}
			} else if len(tc.expected) > 0 {
				// For tests that expect specific properties, check that all expected are present
				// but allow for different ordering since map iteration order is not guaranteed
				assert.Len(t, result, len(tc.expected), "Number of properties should match")
				for _, expected := range tc.expected {
					assert.Contains(t, result, expected, "Expected property should be present")
				}
			} else {
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestEvaluateResourceDanger(t *testing.T) {
	cfg := &config.Config{
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_instance"},
			{ResourceType: "aws_ec2_instance"},
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_ec2_instance", Property: "user_data"},
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name           string
		change         *tfjson.ResourceChange
		changeType     ChangeType
		expectedDanger bool
		expectedReason string
	}{
		{
			name: "Regular deletion should be dangerous",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
			},
			changeType:     ChangeTypeDelete,
			expectedDanger: true,
			expectedReason: "Resource deletion",
		},
		{
			name: "Sensitive resource deletion should be dangerous with specific reason",
			change: &tfjson.ResourceChange{
				Type: "aws_rds_instance",
			},
			changeType:     ChangeTypeDelete,
			expectedDanger: true,
			expectedReason: "Sensitive resource deletion",
		},
		{
			name: "Sensitive resource replacement should be dangerous",
			change: &tfjson.ResourceChange{
				Type: "aws_rds_instance",
			},
			changeType:     ChangeTypeReplace,
			expectedDanger: true,
			expectedReason: "Database replacement",
		},
		{
			name: "EC2 instance replacement should have specific reason",
			change: &tfjson.ResourceChange{
				Type: "aws_ec2_instance",
			},
			changeType:     ChangeTypeReplace,
			expectedDanger: true,
			expectedReason: "Compute instance replacement",
		},
		{
			name: "Non-sensitive resource update should not be dangerous",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"versioning": false,
					},
					After: map[string]interface{}{
						"versioning": true,
					},
				},
			},
			changeType:     ChangeTypeUpdate,
			expectedDanger: false,
			expectedReason: "",
		},
		{
			name: "Sensitive property change should be dangerous",
			change: &tfjson.ResourceChange{
				Type: "aws_ec2_instance",
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"user_data": "old-data",
					},
					After: map[string]interface{}{
						"user_data": "new-data",
					},
				},
			},
			changeType:     ChangeTypeUpdate,
			expectedDanger: true,
			expectedReason: "User data modification",
		},
		{
			name: "Non-sensitive resource creation should not be dangerous",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
			},
			changeType:     ChangeTypeCreate,
			expectedDanger: false,
			expectedReason: "",
		},
		{
			name: "Multiple danger reasons should be combined",
			change: &tfjson.ResourceChange{
				Type: "aws_ec2_instance",
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"user_data": "old-data",
					},
					After: map[string]interface{}{
						"user_data": "new-data",
					},
				},
			},
			changeType:     ChangeTypeDelete,
			expectedDanger: true,
			expectedReason: "Sensitive resource deletion and User data modification",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dangerous, reason := analyzer.evaluateResourceDanger(tc.change, tc.changeType)
			assert.Equal(t, tc.expectedDanger, dangerous, "Danger evaluation mismatch")
			assert.Equal(t, tc.expectedReason, reason, "Danger reason mismatch")
		})
	}
}

func TestGetSensitiveResourceReason(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "RDS instance should return database replacement",
			resourceType: "aws_rds_instance",
			expected:     "Database replacement",
		},
		{
			name:         "Database cluster should return database replacement",
			resourceType: "aws_rds_cluster",
			expected:     "Database replacement",
		},
		{
			name:         "EC2 instance should return compute replacement",
			resourceType: "aws_ec2_instance",
			expected:     "Compute instance replacement",
		},
		{
			name:         "Azure VM should return compute replacement",
			resourceType: "azurerm_virtual_machine",
			expected:     "Compute instance replacement",
		},
		{
			name:         "S3 bucket should return storage replacement",
			resourceType: "aws_s3_bucket",
			expected:     "Storage replacement",
		},
		{
			name:         "Security group should return security replacement",
			resourceType: "aws_security_group",
			expected:     "Security rule replacement",
		},
		{
			name:         "VPC should return network replacement",
			resourceType: "aws_vpc",
			expected:     "Network infrastructure replacement",
		},
		{
			name:         "Unknown resource should return generic replacement",
			resourceType: "custom_resource",
			expected:     "Sensitive resource replacement",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.getSensitiveResourceReason(tc.resourceType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetSensitivePropertyReason(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name       string
		properties []string
		expected   string
	}{
		{
			name:       "Single password property should return credential change",
			properties: []string{"password"},
			expected:   "Credential change",
		},
		{
			name:       "Single secret property should return credential change",
			properties: []string{"secret_key"},
			expected:   "Credential change",
		},
		{
			name:       "Single key property should return authentication key change",
			properties: []string{"api_key"},
			expected:   "Authentication key change",
		},
		{
			name:       "Single token property should return authentication key change",
			properties: []string{"access_token"},
			expected:   "Authentication key change",
		},
		{
			name:       "User data property should return user data modification",
			properties: []string{"user_data"},
			expected:   "User data modification",
		},
		{
			name:       "Security policy property should return security configuration change",
			properties: []string{"security_policy"},
			expected:   "Security configuration change",
		},
		{
			name:       "Unknown single property should return property-specific reason",
			properties: []string{"custom_property"},
			expected:   "Sensitive property change: custom_property",
		},
		{
			name:       "Multiple properties should return generic reason",
			properties: []string{"password", "api_key"},
			expected:   "Multiple sensitive properties changed",
		},
		{
			name:       "Empty properties should return multiple reason",
			properties: []string{},
			expected:   "Multiple sensitive properties changed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.getSensitivePropertyReason(tc.properties)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAnalyzePropertyChanges(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			PerformanceLimits: config.PerformanceLimitsConfig{
				MaxPropertiesPerResource: 2, // Set low for testing
			},
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name          string
		change        *tfjson.ResourceChange
		expectedCount int
		expectedTrunc bool
		expectedError bool
	}{
		{
			name: "Nil change should return empty",
			change: &tfjson.ResourceChange{
				Change: nil,
			},
			expectedCount: 0,
			expectedTrunc: false,
			expectedError: false,
		},
		{
			name: "Simple property change should be detected",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]interface{}{
						"instance_type": "t2.small",
						"ami":           "ami-123",
					},
				},
			},
			expectedCount: 1,
			expectedTrunc: false,
			expectedError: false,
		},
		{
			name: "Multiple changes should respect limit",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
						"subnet_id":     "subnet-123",
						"key_name":      "old-key",
					},
					After: map[string]interface{}{
						"instance_type": "t2.small",
						"ami":           "ami-456",
						"subnet_id":     "subnet-456",
						"key_name":      "new-key",
					},
				},
			},
			expectedCount: 4, // All changes detected (limit not reached)
			expectedTrunc: false,
			expectedError: false,
		},
		{
			name: "Nested property changes should be detected",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"tags": map[string]interface{}{
							"Environment": "staging",
							"Owner":       "team-a",
						},
					},
					After: map[string]interface{}{
						"tags": map[string]interface{}{
							"Environment": "production",
							"Owner":       "team-a",
						},
					},
				},
			},
			expectedCount: 1, // Only Environment tag changed
			expectedTrunc: false,
			expectedError: false,
		},
		{
			name: "Array changes should be detected",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]interface{}{
						"security_groups": []interface{}{"sg-123"},
					},
					After: map[string]interface{}{
						"security_groups": []interface{}{"sg-456"},
					},
				},
			},
			expectedCount: 1,
			expectedTrunc: false,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.analyzePropertyChanges(tc.change, 10)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, result.Count, "Property count mismatch")
				assert.Equal(t, tc.expectedTrunc, result.Truncated, "Truncation flag mismatch")
				assert.Len(t, result.Changes, result.Count, "Changes slice length should match count")
			}
		})
	}

	// Test limit functionality specifically
	t.Run("Limit functionality", func(t *testing.T) {
		change := &tfjson.ResourceChange{
			Change: &tfjson.Change{
				Before: map[string]interface{}{
					"prop1": "old1",
					"prop2": "old2",
					"prop3": "old3",
					"prop4": "old4",
					"prop5": "old5",
				},
				After: map[string]interface{}{
					"prop1": "new1",
					"prop2": "new2",
					"prop3": "new3",
					"prop4": "new4",
					"prop5": "new5",
				},
			},
		}

		result, err := analyzer.analyzePropertyChanges(change, 2) // Limit to 2 properties
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Count, "Should respect property limit")
		assert.True(t, result.Truncated, "Should be truncated when limit is hit")
		assert.Len(t, result.Changes, 2, "Changes slice should contain exactly 2 items")
	})
}

func TestAssessRiskLevel(t *testing.T) {
	cfg := &config.Config{
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_instance"},
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name         string
		change       *tfjson.ResourceChange
		expectedRisk string
	}{
		{
			name: "Regular resource deletion should be high risk",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete},
				},
			},
			expectedRisk: "high",
		},
		{
			name: "Sensitive resource deletion should be critical risk",
			change: &tfjson.ResourceChange{
				Type: "aws_rds_instance",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete},
				},
			},
			expectedRisk: "critical",
		},
		{
			name: "Regular resource replacement should be medium risk",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
				},
			},
			expectedRisk: "medium",
		},
		{
			name: "Sensitive resource replacement should be high risk",
			change: &tfjson.ResourceChange{
				Type: "aws_rds_instance",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
				},
			},
			expectedRisk: "high",
		},
		{
			name: "Sensitive resource update should be medium risk",
			change: &tfjson.ResourceChange{
				Type: "aws_rds_instance",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
				},
			},
			expectedRisk: "medium",
		},
		{
			name: "Regular resource update should be low risk",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
				},
			},
			expectedRisk: "low",
		},
		{
			name: "Resource creation should be low risk",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
				},
			},
			expectedRisk: "low",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.assessRiskLevel(tc.change)
			assert.Equal(t, tc.expectedRisk, result)
		})
	}
}

func TestExtractDependenciesWithLimit(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name              string
		change            *tfjson.ResourceChange
		expectedDependsOn int
		expectedUsedBy    int
		expectedError     bool
	}{
		{
			name: "No dependencies should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: map[string]interface{}{
						"instance_type": "t2.micro",
					},
				},
			},
			expectedDependsOn: 0,
			expectedUsedBy:    0,
			expectedError:     false,
		},
		{
			name: "Explicit depends_on should be extracted",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: map[string]interface{}{
						"depends_on": []interface{}{
							"aws_vpc.main",
							"aws_security_group.web",
						},
					},
				},
			},
			expectedDependsOn: 2,
			expectedUsedBy:    0,
			expectedError:     false,
		},
		{
			name: "Non-array depends_on should be ignored",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: map[string]interface{}{
						"depends_on": "aws_vpc.main",
					},
				},
			},
			expectedDependsOn: 0,
			expectedUsedBy:    0,
			expectedError:     false,
		},
		{
			name: "Nil change should return empty",
			change: &tfjson.ResourceChange{
				Change: nil,
			},
			expectedDependsOn: 0,
			expectedUsedBy:    0,
			expectedError:     false,
		},
		{
			name: "Nil after state should return empty",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: nil,
				},
			},
			expectedDependsOn: 0,
			expectedUsedBy:    0,
			expectedError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.extractDependenciesWithLimit(tc.change, 10)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.DependsOn, tc.expectedDependsOn, "DependsOn length mismatch")
				assert.Len(t, result.UsedBy, tc.expectedUsedBy, "UsedBy length mismatch")
			}
		})
	}
}

func TestAnalyzeResource(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			PerformanceLimits: config.PerformanceLimitsConfig{
				MaxPropertiesPerResource: 100,
				MaxPropertySize:          1048576,
				MaxTotalMemory:           104857600,
			},
		},
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_instance"},
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name            string
		change          *tfjson.ResourceChange
		expectedError   bool
		expectedRisk    string
		hasReplacements bool
	}{
		{
			name: "Simple resource change should be analyzed",
			change: &tfjson.ResourceChange{
				Type: "aws_s3_bucket",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"versioning": false,
					},
					After: map[string]interface{}{
						"versioning": true,
					},
				},
			},
			expectedError:   false,
			expectedRisk:    "low",
			hasReplacements: false,
		},
		{
			name: "Sensitive resource replacement should be high risk",
			change: &tfjson.ResourceChange{
				Type: "aws_rds_instance",
				Change: &tfjson.Change{
					Actions:      tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
					ReplacePaths: []interface{}{"engine_version"},
				},
			},
			expectedError:   false,
			expectedRisk:    "high",
			hasReplacements: true,
		},
		{
			name: "Resource with dependencies should extract them",
			change: &tfjson.ResourceChange{
				Type: "aws_ec2_instance",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
					After: map[string]interface{}{
						"depends_on": []interface{}{
							"aws_vpc.main",
						},
					},
				},
			},
			expectedError:   false,
			expectedRisk:    "low",
			hasReplacements: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeResource(tc.change)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedRisk, result.RiskLevel, "Risk level mismatch")

				if tc.hasReplacements {
					assert.Greater(t, len(result.ReplacementReasons), 0, "Should have replacement reasons")
				}

				// Verify all fields are populated
				assert.NotNil(t, result.PropertyChanges, "PropertyChanges should not be nil")
				assert.NotNil(t, result.ReplacementReasons, "ReplacementReasons should not be nil")
				assert.NotEmpty(t, result.RiskLevel, "RiskLevel should not be empty")
			}
		})
	}
}

func TestEstimateValueSize(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		value    interface{}
		expected int
	}{
		{
			name:     "Nil should return 0",
			value:    nil,
			expected: 0,
		},
		{
			name:     "String should return length",
			value:    "hello",
			expected: 5,
		},
		{
			name:     "Int should return 8 bytes",
			value:    42,
			expected: 8,
		},
		{
			name:     "Float64 should return 8 bytes",
			value:    3.14,
			expected: 8,
		},
		{
			name:     "Bool should return 1 byte",
			value:    true,
			expected: 1,
		},
		{
			name: "Map should sum key and value sizes",
			value: map[string]interface{}{
				"key1": "value1", // 4 + 6 = 10
				"key2": "value2", // 4 + 6 = 10
			},
			expected: 20,
		},
		{
			name: "Array should sum element sizes",
			value: []interface{}{
				"hello", // 5
				"world", // 5
			},
			expected: 10,
		},
		{
			name: "Complex nested structure",
			value: map[string]interface{}{
				"name": "test", // 4 + 4 = 8
				"tags": map[string]interface{}{ // 4 + (3+4 + 5+4) = 20
					"env":   "prod",
					"owner": "team",
				},
			},
			expected: 28,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.estimateValueSize(tc.value)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCompareValues(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name            string
		before          interface{}
		after           interface{}
		expectedChanges int
	}{
		{
			name:            "Identical values should return no changes",
			before:          "same",
			after:           "same",
			expectedChanges: 0,
		},
		{
			name:            "Different primitive values should return one change",
			before:          "old",
			after:           "new",
			expectedChanges: 1,
		},
		{
			name: "Map with one change should return one change",
			before: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			after: map[string]interface{}{
				"key1": "value1",
				"key2": "new_value2",
			},
			expectedChanges: 1,
		},
		{
			name: "Map with new key should return one change",
			before: map[string]interface{}{
				"key1": "value1",
			},
			after: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expectedChanges: 1,
		},
		{
			name: "Map with removed key should return one change",
			before: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			after: map[string]interface{}{
				"key1": "value1",
			},
			expectedChanges: 1,
		},
		{
			name:            "Array changes should be detected",
			before:          []interface{}{"a", "b"},
			after:           []interface{}{"a", "c"},
			expectedChanges: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			changes := []PropertyChange{}

			err := analyzer.compareValues(tc.before, tc.after, nil, 0, 5, func(pc PropertyChange) bool {
				changes = append(changes, pc)
				return true
			})

			assert.NoError(t, err)
			assert.Len(t, changes, tc.expectedChanges, "Number of changes should match expected")
		})
	}
}

func TestGroupByProvider(t *testing.T) {
	testCases := []struct {
		name           string
		config         *config.Config
		changes        []ResourceChange
		expectedGroups int
		expectedEmpty  bool
		groupNames     []string
	}{
		{
			name: "Grouping disabled should return empty",
			config: &config.Config{
				Plan: config.PlanConfig{
					Grouping: config.GroupingConfig{
						Enabled: false,
					},
				},
			},
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "azurerm_storage_account"},
			},
			expectedGroups: 0,
			expectedEmpty:  true,
		},
		{
			name: "Below threshold should return empty",
			config: &config.Config{
				Plan: config.PlanConfig{
					Grouping: config.GroupingConfig{
						Enabled:   true,
						Threshold: 10,
					},
				},
			},
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "azurerm_storage_account"},
			},
			expectedGroups: 0,
			expectedEmpty:  true,
		},
		{
			name: "Single provider should return empty",
			config: &config.Config{
				Plan: config.PlanConfig{
					Grouping: config.GroupingConfig{
						Enabled:   true,
						Threshold: 2,
					},
				},
			},
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "aws_ec2_instance"},
				{Type: "aws_rds_instance"},
			},
			expectedGroups: 0,
			expectedEmpty:  true,
		},
		{
			name: "Multiple providers above threshold should group",
			config: &config.Config{
				Plan: config.PlanConfig{
					Grouping: config.GroupingConfig{
						Enabled:   true,
						Threshold: 3,
					},
				},
			},
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "aws_ec2_instance"},
				{Type: "azurerm_storage_account"},
				{Type: "google_compute_instance"},
			},
			expectedGroups: 3,
			expectedEmpty:  false,
			groupNames:     []string{"aws", "azurerm", "google"},
		},
		{
			name: "Default threshold of 10 should be used",
			config: &config.Config{
				Plan: config.PlanConfig{
					Grouping: config.GroupingConfig{
						Enabled:   true,
						Threshold: 0, // Should default to 10
					},
				},
			},
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "azurerm_storage_account"},
			},
			expectedGroups: 0,
			expectedEmpty:  true,
		},
		{
			name:   "Nil config should return empty",
			config: nil,
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "azurerm_storage_account"},
			},
			expectedGroups: 0,
			expectedEmpty:  true,
		},
		{
			name: "Mixed providers with sufficient resources should group",
			config: &config.Config{
				Plan: config.PlanConfig{
					Grouping: config.GroupingConfig{
						Enabled:   true,
						Threshold: 5,
					},
				},
			},
			changes: []ResourceChange{
				{Type: "aws_s3_bucket"},
				{Type: "aws_ec2_instance"},
				{Type: "aws_rds_instance"},
				{Type: "azurerm_storage_account"},
				{Type: "azurerm_virtual_machine"},
				{Type: "google_compute_instance"},
			},
			expectedGroups: 3,
			expectedEmpty:  false,
			groupNames:     []string{"aws", "azurerm", "google"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analyzer := &Analyzer{config: tc.config}
			result := analyzer.groupByProvider(tc.changes)

			if tc.expectedEmpty {
				assert.Empty(t, result, "Should return empty groups")
			} else {
				assert.Len(t, result, tc.expectedGroups, "Number of groups should match expected")

				// Check that all expected group names are present
				for _, expectedGroup := range tc.groupNames {
					assert.Contains(t, result, expectedGroup, "Should contain group: %s", expectedGroup)
				}

				// Check that total resources in groups equals input resources
				totalResourcesInGroups := 0
				for _, group := range result {
					totalResourcesInGroups += len(group)
				}
				assert.Equal(t, len(tc.changes), totalResourcesInGroups, "Total resources in groups should equal input")

				// Check that resources are in correct groups
				for provider, resources := range result {
					for _, resource := range resources {
						expectedProvider := extractProviderFromType(resource.Type)
						assert.Equal(t, provider, expectedProvider, "Resource should be in correct provider group")
					}
				}
			}
		})
	}
}

// Helper function to extract provider for testing
func extractProviderFromType(resourceType string) string {
	parts := strings.Split(resourceType, "_")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}
