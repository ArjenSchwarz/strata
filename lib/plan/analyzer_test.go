package plan

import (
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
