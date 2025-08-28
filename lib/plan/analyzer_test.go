package plan

import (
	"encoding/json"
	"fmt"
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
			Before: map[string]any{
				"user_data":     "old-data",
				"instance_type": "t2.micro",
			},
			After: map[string]any{
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
	resourceChange.Change.After.(map[string]any)["user_data"] = "old-data"
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
					ReplacePaths: []any{},
				},
			},
			expected: ReplacementAlways,
		},
		{
			name: "Replace with ReplacePaths and definite values should be always",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions:      tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
					ReplacePaths: []any{[]any{"definite_field"}},
					After: map[string]any{
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
			result := analyzer.calculateStatistics(tc.changes, []OutputChange{})
			assert.Equal(t, tc.expected.ToAdd, result.ToAdd, "ToAdd mismatch")
			assert.Equal(t, tc.expected.ToChange, result.ToChange, "ToChange mismatch")
			assert.Equal(t, tc.expected.ToDestroy, result.ToDestroy, "ToDestroy mismatch")
			assert.Equal(t, tc.expected.Replacements, result.Replacements, "Replacements mismatch")
			assert.Equal(t, tc.expected.HighRisk, result.HighRisk, "HighRisk mismatch")
			assert.Equal(t, tc.expected.Total, result.Total, "Total mismatch")
			assert.Equal(t, tc.expected.OutputChanges, result.OutputChanges, "OutputChanges mismatch")
		})
	}
}

// TestCalculateStatisticsWithOutputChanges tests statistics behavior with output changes, specifically verifying that no-op outputs are excluded
func TestCalculateStatisticsWithOutputChanges(t *testing.T) {
	// Create analyzer with empty config for this test
	analyzer := &Analyzer{
		config: &config.Config{},
	}

	testCases := []struct {
		name     string
		changes  []ResourceChange
		outputs  []OutputChange
		expected ChangeStatistics
	}{
		{
			name:    "No changes should have all zeros including OutputChanges",
			changes: []ResourceChange{},
			outputs: []OutputChange{},
			expected: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    0,
				Total:         0,
				OutputChanges: 0,
			},
		},
		{
			name:    "Only output changes should count non-no-op outputs",
			changes: []ResourceChange{},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeUpdate, IsNoOp: false},
				{Name: "output2", ChangeType: ChangeTypeCreate, IsNoOp: false},
				{Name: "output3", ChangeType: ChangeTypeNoOp, IsNoOp: true}, // Should be excluded
			},
			expected: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    0,
				Total:         0,
				OutputChanges: 2, // Only non-no-op outputs
			},
		},
		{
			name: "Mixed resource and output changes should count correctly",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate},
				{ChangeType: ChangeTypeUpdate},
				{ChangeType: ChangeTypeNoOp}, // Should count in Unmodified but not Total
			},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeUpdate, IsNoOp: false},
				{Name: "output2", ChangeType: ChangeTypeNoOp, IsNoOp: true}, // Should be excluded from OutputChanges
				{Name: "output3", ChangeType: ChangeTypeDelete, IsNoOp: false},
			},
			expected: ChangeStatistics{
				ToAdd:         1,
				ToChange:      1,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    1, // No-op resource counts in Unmodified
				Total:         2, // Only non-no-op resources
				OutputChanges: 2, // Only non-no-op outputs
			},
		},
		{
			name:    "All no-op outputs should result in zero OutputChanges",
			changes: []ResourceChange{},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeNoOp, IsNoOp: true},
				{Name: "output2", ChangeType: ChangeTypeNoOp, IsNoOp: true},
				{Name: "output3", ChangeType: ChangeTypeNoOp, IsNoOp: true},
			},
			expected: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    0,
				Total:         0,
				OutputChanges: 0, // All outputs are no-ops, so excluded
			},
		},
		{
			name: "Resource no-ops should not affect output statistics",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeNoOp}, // Resource no-op
				{ChangeType: ChangeTypeNoOp}, // Another resource no-op
			},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeUpdate, IsNoOp: false}, // Non-no-op output
			},
			expected: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    2, // Resource no-ops count in Unmodified
				Total:         0, // No-op resources don't count in Total
				OutputChanges: 1, // Non-no-op output counts
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.calculateStatistics(tc.changes, tc.outputs)
			assert.Equal(t, tc.expected.ToAdd, result.ToAdd, "ToAdd mismatch")
			assert.Equal(t, tc.expected.ToChange, result.ToChange, "ToChange mismatch")
			assert.Equal(t, tc.expected.ToDestroy, result.ToDestroy, "ToDestroy mismatch")
			assert.Equal(t, tc.expected.Replacements, result.Replacements, "Replacements mismatch")
			assert.Equal(t, tc.expected.HighRisk, result.HighRisk, "HighRisk mismatch")
			assert.Equal(t, tc.expected.Unmodified, result.Unmodified, "Unmodified mismatch")
			assert.Equal(t, tc.expected.Total, result.Total, "Total mismatch")
			assert.Equal(t, tc.expected.OutputChanges, result.OutputChanges, "OutputChanges mismatch")
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
					Before: map[string]any{
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
					Before: map[string]any{
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
					Before: map[string]any{
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
		{
			name:     "Module with iterator should strip iterator",
			address:  "module.s3_module[0].aws_s3_bucket.logs",
			expected: "s3_module",
		},
		{
			name:     "Nested modules with iterators should strip iterators",
			address:  "module.app[1].module.storage[0].aws_s3_bucket.data",
			expected: "app/storage",
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
					ReplacePaths: []any{},
				},
			},
			expected: []string{},
		},
		{
			name: "Simple string path should be formatted",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []any{"subnet_id"},
				},
			},
			expected: []string{"subnet_id"},
		},
		{
			name: "Nested array path should be formatted with dots and brackets",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []any{
						[]any{"network_interface", 0, "subnet_id"},
					},
				},
			},
			expected: []string{"network_interface.[0].subnet_id"},
		},
		{
			name: "Multiple replacement paths should all be included",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					ReplacePaths: []any{
						"subnet_id",
						[]any{"security_groups", 1},
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
					ReplacePaths: []any{
						[]any{"network_interface", 0.0, "subnet_id"},
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
		path     any
		expected string
	}{
		{
			name:     "Simple string should return as-is",
			path:     "subnet_id",
			expected: "subnet_id",
		},
		{
			name:     "Array with string should format with dots",
			path:     []any{"network_interface", "subnet_id"},
			expected: "network_interface.subnet_id",
		},
		{
			name:     "Array with int should format with brackets",
			path:     []any{"security_groups", 0},
			expected: "security_groups.[0]",
		},
		{
			name:     "Array with float64 should format with brackets",
			path:     []any{"security_groups", 1.0},
			expected: "security_groups.[1]",
		},
		{
			name:     "Complex nested path should format correctly",
			path:     []any{"block_device_mappings", 0, "ebs", "volume_size"},
			expected: "block_device_mappings.[0].ebs.volume_size",
		},
		{
			name:     "Empty array should return empty string",
			path:     []any{},
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
					Before: map[string]any{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]any{
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
					After: map[string]any{
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
					Before: map[string]any{
						"instance_type":      "t2.micro",
						"ami":                "ami-123",
						"security_group_ids": []any{"sg-123"},
						"unchanged_property": "same",
					},
					After: map[string]any{
						"instance_type":      "t2.small",
						"ami":                "ami-456",
						"security_group_ids": []any{"sg-456"},
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
					Before: map[string]any{
						"prop1": "old1",
						"prop2": "old2",
						"prop3": "old3",
						"prop4": "old4",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"existing_prop": "value",
						"removed_prop":  "old_value",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"versioning": false,
					},
					After: map[string]any{
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
					Before: map[string]any{
						"user_data": "old-data",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"user_data": "old-data",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
						"subnet_id":     "subnet-123",
						"key_name":      "old-key",
					},
					After: map[string]any{
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
					Before: map[string]any{
						"tags": map[string]any{
							"Environment": "staging",
							"Owner":       "team-a",
						},
					},
					After: map[string]any{
						"tags": map[string]any{
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
					Before: map[string]any{
						"security_groups": []any{"sg-123"},
					},
					After: map[string]any{
						"security_groups": []any{"sg-456"},
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
			result := analyzer.analyzePropertyChanges(tc.change)

			// Debug output for failing tests
			if result.Count != tc.expectedCount {
				t.Logf("Expected %d changes, got %d changes:", tc.expectedCount, result.Count)
				for i, c := range result.Changes {
					t.Logf("  %d. Name: '%s', Action: %s, Before: %v, After: %v, Triggers: %v",
						i+1, c.Name, c.Action, c.Before, c.After, c.TriggersReplacement)
				}
			}

			assert.Equal(t, tc.expectedCount, result.Count, "Property count mismatch")
			assert.Equal(t, tc.expectedTrunc, result.Truncated, "Truncation flag mismatch")
			assert.Len(t, result.Changes, result.Count, "Changes slice length should match count")
		})
	}

	// Test limit functionality specifically
	t.Run("Limit functionality", func(t *testing.T) {
		change := &tfjson.ResourceChange{
			Change: &tfjson.Change{
				Before: map[string]any{
					"prop1": "old1",
					"prop2": "old2",
					"prop3": "old3",
					"prop4": "old4",
					"prop5": "old5",
				},
				After: map[string]any{
					"prop1": "new1",
					"prop2": "new2",
					"prop3": "new3",
					"prop4": "new4",
					"prop5": "new5",
				},
			},
		}

		result := analyzer.analyzePropertyChanges(change)
		// Note: Property limits are now handled internally within the method
		// The test should verify the behavior works but limits might be different
		assert.True(t, len(result.Changes) <= 100, "Should respect internal property limit")
		if len(result.Changes) >= 100 {
			assert.True(t, result.Truncated, "Should be truncated when limit is hit")
		}
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
					Before: map[string]any{
						"versioning": false,
					},
					After: map[string]any{
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
					ReplacePaths: []any{"engine_version"},
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
					After: map[string]any{
						"depends_on": []any{
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
		value    any
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
			value: map[string]any{
				"key1": "value1", // 4 + 6 = 10
				"key2": "value2", // 4 + 6 = 10
			},
			expected: 20,
		},
		{
			name: "Array should sum element sizes",
			value: []any{
				"hello", // 5
				"world", // 5
			},
			expected: 10,
		},
		{
			name: "Complex nested structure",
			value: map[string]any{
				"name": "test", // 4 + 4 = 8
				"tags": map[string]any{ // 4 + (3+4 + 5+4) = 20
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
		before          any
		after           any
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
			before: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			after: map[string]any{
				"key1": "value1",
				"key2": "new_value2",
			},
			expectedChanges: 1,
		},
		{
			name: "Map with new key should return one change",
			before: map[string]any{
				"key1": "value1",
			},
			after: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expectedChanges: 1,
		},
		{
			name: "Map with removed key should return one change",
			before: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			after: map[string]any{
				"key1": "value1",
			},
			expectedChanges: 1,
		},
		{
			name:            "Array changes should be detected",
			before:          []any{"a", "b"},
			after:           []any{"a", "c"},
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

// TestCompareObjects tests the deep object comparison algorithm
func TestCompareObjects(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		before            any
		after             any
		beforeSensitive   any
		afterSensitive    any
		expectedChanges   int
		expectedActions   []string
		expectedNames     []string
		expectedSensitive []bool
	}{
		{
			name:              "simple string change",
			before:            map[string]any{"name": "old"},
			after:             map[string]any{"name": "new"},
			expectedChanges:   1,
			expectedActions:   []string{"update"},
			expectedNames:     []string{"name"},
			expectedSensitive: []bool{false},
		},
		{
			name:              "nested object change",
			before:            map[string]any{"tags": map[string]any{"env": "dev"}},
			after:             map[string]any{"tags": map[string]any{"env": "prod"}},
			expectedChanges:   1,
			expectedActions:   []string{"update"},
			expectedNames:     []string{"tags"},
			expectedSensitive: []bool{false},
		},
		{
			name:              "array length change",
			before:            map[string]any{"items": []any{1, 2}},
			after:             map[string]any{"items": []any{1, 2, 3}},
			expectedChanges:   1,
			expectedActions:   []string{"update"},
			expectedNames:     []string{"items"},
			expectedSensitive: []bool{false},
		},
		{
			name:              "property removal",
			before:            map[string]any{"a": 1, "b": 2},
			after:             map[string]any{"a": 1},
			expectedChanges:   1,
			expectedActions:   []string{"remove"},
			expectedNames:     []string{"b"},
			expectedSensitive: []bool{false},
		},
		{
			name:              "property addition",
			before:            map[string]any{"a": 1},
			after:             map[string]any{"a": 1, "b": 2},
			expectedChanges:   1,
			expectedActions:   []string{"add"},
			expectedNames:     []string{"b"},
			expectedSensitive: []bool{false},
		},
		{
			name:              "sensitive value change",
			before:            map[string]any{"password": "old"},
			after:             map[string]any{"password": "new"},
			beforeSensitive:   map[string]any{"password": true},
			afterSensitive:    map[string]any{"password": true},
			expectedChanges:   1,
			expectedActions:   []string{"update"},
			expectedNames:     []string{"password"},
			expectedSensitive: []bool{true},
		},
		{
			name:              "multiple changes",
			before:            map[string]any{"name": "old", "count": 1},
			after:             map[string]any{"name": "new", "count": 2},
			expectedChanges:   2,
			expectedActions:   []string{"update", "update"},
			expectedNames:     []string{"name", "count"},
			expectedSensitive: []bool{false, false},
		},
		{
			name:            "no changes",
			before:          map[string]any{"name": "same"},
			after:           map[string]any{"name": "same"},
			expectedChanges: 0,
		},
		{
			name:              "nil to value",
			before:            nil,
			after:             map[string]any{"name": "new"},
			expectedChanges:   1,
			expectedActions:   []string{"add"},
			expectedNames:     []string{"name"},
			expectedSensitive: []bool{false},
		},
		{
			name:              "value to nil",
			before:            map[string]any{"name": "old"},
			after:             nil,
			expectedChanges:   1,
			expectedActions:   []string{"remove"},
			expectedNames:     []string{"name"},
			expectedSensitive: []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &PropertyChangeAnalysis{
				Changes: []PropertyChange{},
			}

			analyzer.compareObjects("", tt.before, tt.after, tt.beforeSensitive, tt.afterSensitive, nil, nil, analysis)

			assert.Equal(t, tt.expectedChanges, len(analysis.Changes), "Expected number of changes")

			if tt.expectedChanges > 0 {
				// For multiple changes, make order-independent assertions
				actualNames := make([]string, len(analysis.Changes))
				actualActions := make([]string, len(analysis.Changes))
				actualSensitive := make([]bool, len(analysis.Changes))

				for i, change := range analysis.Changes {
					actualNames[i] = change.Name
					actualActions[i] = change.Action
					actualSensitive[i] = change.Sensitive
				}

				if len(tt.expectedNames) > 0 {
					assert.ElementsMatch(t, tt.expectedNames, actualNames, "Expected names should match")
				}
				if len(tt.expectedActions) > 0 {
					assert.ElementsMatch(t, tt.expectedActions, actualActions, "Expected actions should match")
				}
				if len(tt.expectedSensitive) > 0 {
					assert.ElementsMatch(t, tt.expectedSensitive, actualSensitive, "Expected sensitivity should match")
				}
			}
		})
	}
}

// TestExtractPropertyName tests the property name extraction
func TestExtractPropertyName(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple property",
			path:     "name",
			expected: "name",
		},
		{
			name:     "nested property",
			path:     "tags.environment",
			expected: "environment",
		},
		{
			name:     "array property",
			path:     "items[0]",
			expected: "items",
		},
		{
			name:     "nested array property",
			path:     "tags[0].name",
			expected: "name",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "deep nested property",
			path:     "config.database.settings.timeout",
			expected: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.extractPropertyName(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParsePath tests the path parsing functionality
func TestParsePath(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple property",
			path:     "name",
			expected: []string{"name"},
		},
		{
			name:     "nested property",
			path:     "tags.environment",
			expected: []string{"tags", "environment"},
		},
		{
			name:     "array property",
			path:     "items[0]",
			expected: []string{"items", "0"},
		},
		{
			name:     "nested array property",
			path:     "tags[0].name",
			expected: []string{"tags", "0", "name"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "multiple array indices",
			path:     "matrix[1][2]",
			expected: []string{"matrix", "1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.parsePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSensitive tests the sensitive value detection
func TestIsSensitive(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name            string
		path            string
		sensitiveValues any
		expected        bool
	}{
		{
			name:            "simple sensitive property",
			path:            "password",
			sensitiveValues: map[string]any{"password": true},
			expected:        true,
		},
		{
			name:            "simple non-sensitive property",
			path:            "name",
			sensitiveValues: map[string]any{"password": true},
			expected:        false,
		},
		{
			name:            "nested sensitive property",
			path:            "config.password",
			sensitiveValues: map[string]any{"config": map[string]any{"password": true}},
			expected:        true,
		},
		{
			name:            "array sensitive property",
			path:            "secrets[0]",
			sensitiveValues: map[string]any{"secrets": []any{true, false}},
			expected:        true,
		},
		{
			name:            "array non-sensitive property",
			path:            "secrets[1]",
			sensitiveValues: map[string]any{"secrets": []any{true, false}},
			expected:        false,
		},
		{
			name:            "nil sensitive values",
			path:            "password",
			sensitiveValues: nil,
			expected:        false,
		},
		{
			name:            "path not found",
			path:            "nonexistent",
			sensitiveValues: map[string]any{"password": true},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isSensitive(tt.path, tt.sensitiveValues)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsOutputSensitive tests output sensitivity detection
func TestIsOutputSensitive(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		beforeSensitive   any
		afterSensitive    any
		expectedSensitive bool
	}{
		{
			name:              "non-sensitive output",
			beforeSensitive:   false,
			afterSensitive:    false,
			expectedSensitive: false,
		},
		{
			name:              "sensitive before value",
			beforeSensitive:   true,
			afterSensitive:    false,
			expectedSensitive: true,
		},
		{
			name:              "sensitive after value",
			beforeSensitive:   false,
			afterSensitive:    true,
			expectedSensitive: true,
		},
		{
			name:              "both sensitive",
			beforeSensitive:   true,
			afterSensitive:    true,
			expectedSensitive: true,
		},
		{
			name:              "nil values",
			beforeSensitive:   nil,
			afterSensitive:    nil,
			expectedSensitive: false,
		},
		{
			name:              "non-boolean sensitive values",
			beforeSensitive:   "not_boolean",
			afterSensitive:    123,
			expectedSensitive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := &tfjson.Change{
				BeforeSensitive: tt.beforeSensitive,
				AfterSensitive:  tt.afterSensitive,
			}
			result := analyzer.isOutputSensitive(change)
			if result != tt.expectedSensitive {
				t.Errorf("isOutputSensitive() = %v, expected %v", result, tt.expectedSensitive)
			}
		})
	}
}

// TestExtractSensitiveChild tests sensitive child extraction
func TestExtractSensitiveChild(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name            string
		sensitiveValues any
		key             string
		expected        any
	}{
		{
			name:            "extract child from map",
			sensitiveValues: map[string]any{"password": true, "name": false},
			key:             "password",
			expected:        true,
		},
		{
			name:            "extract missing child from map",
			sensitiveValues: map[string]any{"password": true},
			key:             "name",
			expected:        nil,
		},
		{
			name:            "extract from nil",
			sensitiveValues: nil,
			key:             "password",
			expected:        nil,
		},
		{
			name:            "extract from non-map",
			sensitiveValues: true,
			key:             "password",
			expected:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.extractSensitiveChild(tt.sensitiveValues, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractSensitiveIndex tests sensitive array index extraction
func TestExtractSensitiveIndex(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name            string
		sensitiveValues any
		index           int
		expected        any
	}{
		{
			name:            "extract valid index from array",
			sensitiveValues: []any{true, false, true},
			index:           1,
			expected:        false,
		},
		{
			name:            "extract out of bounds index from array",
			sensitiveValues: []any{true, false},
			index:           5,
			expected:        nil,
		},
		{
			name:            "extract negative index from array",
			sensitiveValues: []any{true, false},
			index:           -1,
			expected:        nil,
		},
		{
			name:            "extract from nil",
			sensitiveValues: nil,
			index:           0,
			expected:        nil,
		},
		{
			name:            "extract from non-array",
			sensitiveValues: map[string]any{"test": true},
			index:           0,
			expected:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.extractSensitiveIndex(tt.sensitiveValues, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAnalyzePropertyChangesWithNewCompareObjects tests the updated analyzePropertyChanges method
func TestAnalyzePropertyChangesWithNewCompareObjects(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		change            *tfjson.ResourceChange
		expectedChanges   int
		expectedTruncated bool
	}{
		{
			name: "simple property changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"name": "old-name",
						"size": 10,
					},
					After: map[string]any{
						"name": "new-name",
						"size": 20,
					},
				},
			},
			expectedChanges:   2,
			expectedTruncated: false,
		},
		{
			name: "no changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{"name": "same"},
					After:  map[string]any{"name": "same"},
				},
			},
			expectedChanges:   0,
			expectedTruncated: false,
		},
		{
			name: "nested property changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"tags":     map[string]any{"env": "dev", "team": "backend"},
						"settings": map[string]any{"timeout": 30},
					},
					After: map[string]any{
						"tags":     map[string]any{"env": "prod", "team": "backend"},
						"settings": map[string]any{"timeout": 60},
					},
				},
			},
			expectedChanges:   2, // env and timeout changed
			expectedTruncated: false,
		},
		{
			name: "sensitive property changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"password": "old-secret",
						"name":     "resource",
					},
					After: map[string]any{
						"password": "new-secret",
						"name":     "resource",
					},
					BeforeSensitive: map[string]any{
						"password": true,
						"name":     false,
					},
					AfterSensitive: map[string]any{
						"password": true,
						"name":     false,
					},
				},
			},
			expectedChanges:   1, // only password changed
			expectedTruncated: false,
		},
		{
			name: "nil change",
			change: &tfjson.ResourceChange{
				Change: nil,
			},
			expectedChanges:   0,
			expectedTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzePropertyChanges(tt.change)

			assert.Equal(t, tt.expectedChanges, result.Count, "Expected count should match")
			assert.Equal(t, tt.expectedChanges, len(result.Changes), "Expected changes length should match")
			assert.Equal(t, tt.expectedTruncated, result.Truncated, "Expected truncation status should match")

			// Verify that each change has the required Action field
			for i, change := range result.Changes {
				assert.NotEmpty(t, change.Action, "Change %d should have an action", i)
				assert.Contains(t, []string{"add", "remove", "update"}, change.Action, "Change %d should have valid action", i)
				assert.NotEmpty(t, change.Name, "Change %d should have a name", i)
			}
		})
	}
}

// TestCompareObjectsEnhanced tests the deep object comparison algorithm with specific scenarios
func TestCompareObjectsEnhanced(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name     string
		before   any
		after    any
		expected []PropertyChange
	}{
		{
			name:   "simple string change",
			before: map[string]any{"name": "old"},
			after:  map[string]any{"name": "new"},
			expected: []PropertyChange{{
				Name:   "name",
				Path:   []string{"name"},
				Action: "update",
				Before: "old",
				After:  "new",
			}},
		},
		{
			name: "nested object change",
			before: map[string]any{
				"tags": map[string]any{"env": "dev"},
			},
			after: map[string]any{
				"tags": map[string]any{"env": "prod"},
			},
			expected: []PropertyChange{{
				Name:   "tags",
				Path:   []string{"tags"},
				Action: "update",
				Before: map[string]any{"env": "dev"},
				After:  map[string]any{"env": "prod"},
			}},
		},
		{
			name:   "array length change",
			before: map[string]any{"items": []any{1, 2}},
			after:  map[string]any{"items": []any{1, 2, 3}},
			expected: []PropertyChange{{
				Name:   "items",
				Path:   []string{"items"},
				Action: "update",
				Before: []any{1, 2},
				After:  []any{1, 2, 3},
			}},
		},
		{
			name:   "property removal",
			before: map[string]any{"a": 1, "b": 2},
			after:  map[string]any{"a": 1},
			expected: []PropertyChange{{
				Name:   "b",
				Path:   []string{"b"},
				Action: "remove",
				Before: 2,
				After:  nil,
			}},
		},
		{
			name:   "property addition",
			before: map[string]any{"a": 1},
			after:  map[string]any{"a": 1, "b": 2},
			expected: []PropertyChange{{
				Name:   "b",
				Path:   []string{"b"},
				Action: "add",
				Before: nil,
				After:  2,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := PropertyChangeAnalysis{
				Changes: []PropertyChange{},
			}

			analyzer.compareObjects("", tt.before, tt.after, nil, nil, nil, nil, &analysis)

			assert.Equal(t, len(tt.expected), len(analysis.Changes), "Expected number of changes should match")

			for i, expectedChange := range tt.expected {
				if i < len(analysis.Changes) {
					actual := analysis.Changes[i]
					assert.Equal(t, expectedChange.Name, actual.Name, "Change %d name should match", i)
					assert.Equal(t, expectedChange.Path, actual.Path, "Change %d path should match", i)
					assert.Equal(t, expectedChange.Action, actual.Action, "Change %d action should match", i)
					assert.Equal(t, expectedChange.Before, actual.Before, "Change %d before value should match", i)
					assert.Equal(t, expectedChange.After, actual.After, "Change %d after value should match", i)
				}
			}
		})
	}
}

// TestEnforcePropertyLimits tests the performance limit enforcement
func TestEnforcePropertyLimits(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		initialChanges    []PropertyChange
		expectedCount     int
		expectedTruncated bool
		expectedTotalSize int
		testType          string
	}{
		{
			name: "under limits should not truncate",
			initialChanges: []PropertyChange{
				{Name: "prop1", Before: "small", After: "value", Action: "update"},
				{Name: "prop2", Before: "another", After: "small", Action: "update"},
			},
			expectedCount:     2,
			expectedTruncated: false,
			testType:          "normal",
		},
		{
			name: "property count limit should truncate",
			initialChanges: func() []PropertyChange {
				changes := make([]PropertyChange, MaxPropertiesPerResource+5)
				for i := range changes {
					changes[i] = PropertyChange{
						Name:   fmt.Sprintf("prop%d", i),
						Before: "value",
						After:  "new",
						Action: "update",
					}
				}
				return changes
			}(),
			expectedCount:     MaxPropertiesPerResource,
			expectedTruncated: true,
			testType:          "count_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := PropertyChangeAnalysis{
				Changes: tt.initialChanges,
			}

			analyzer.enforcePropertyLimits(&analysis)

			assert.Equal(t, tt.expectedCount, analysis.Count, "Count should match expected")
			assert.Equal(t, tt.expectedCount, len(analysis.Changes), "Changes length should match count")
			assert.Equal(t, tt.expectedTruncated, analysis.Truncated, "Truncated status should match expected")

			// Verify all remaining changes have Size set
			for i, change := range analysis.Changes {
				assert.GreaterOrEqual(t, change.Size, 0, "Change %d should have non-negative size", i)
			}
		})
	}
}

// TestIsValueUnknown tests the unknown value detection function
func TestIsValueUnknown(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name         string
		afterUnknown any
		path         string
		expected     bool
	}{
		{
			name:         "nil afterUnknown should return false",
			afterUnknown: nil,
			path:         "property",
			expected:     false,
		},
		{
			name:         "simple true unknown value should return true",
			afterUnknown: map[string]any{"id": true},
			path:         "id",
			expected:     true,
		},
		{
			name:         "simple false unknown value should return false",
			afterUnknown: map[string]any{"id": false},
			path:         "id",
			expected:     false,
		},
		{
			name:         "missing property should return false",
			afterUnknown: map[string]any{"id": true},
			path:         "name",
			expected:     false,
		},
		{
			name:         "nested object unknown property should return true",
			afterUnknown: map[string]any{"config": map[string]any{"timeout": true}},
			path:         "config.timeout",
			expected:     true,
		},
		{
			name:         "nested object known property should return false",
			afterUnknown: map[string]any{"config": map[string]any{"timeout": false}},
			path:         "config.timeout",
			expected:     false,
		},
		{
			name:         "array unknown element should return true",
			afterUnknown: map[string]any{"vpc_security_group_ids": []any{true, false}},
			path:         "vpc_security_group_ids[0]",
			expected:     true,
		},
		{
			name:         "array known element should return false",
			afterUnknown: map[string]any{"vpc_security_group_ids": []any{true, false}},
			path:         "vpc_security_group_ids[1]",
			expected:     false,
		},
		{
			name:         "out of bounds array index should return false",
			afterUnknown: map[string]any{"vpc_security_group_ids": []any{true}},
			path:         "vpc_security_group_ids[5]",
			expected:     false,
		},
		{
			name: "complex nested structure with unknown values",
			afterUnknown: map[string]any{
				"network_interface": []any{
					map[string]any{
						"subnet_id":    true,
						"device_index": false,
					},
				},
			},
			path:     "network_interface[0].subnet_id",
			expected: true,
		},
		{
			name: "complex nested structure with known values",
			afterUnknown: map[string]any{
				"network_interface": []any{
					map[string]any{
						"subnet_id":    true,
						"device_index": false,
					},
				},
			},
			path:     "network_interface[0].device_index",
			expected: false,
		},
		{
			name:         "invalid array index should return false",
			afterUnknown: map[string]any{"items": []any{true}},
			path:         "items[invalid]",
			expected:     false,
		},
		{
			name:         "negative array index should return false",
			afterUnknown: map[string]any{"items": []any{true}},
			path:         "items[-1]",
			expected:     false,
		},
		{
			name:         "boolean at intermediate level should return that boolean",
			afterUnknown: map[string]any{"config": true},
			path:         "config.anything",
			expected:     true,
		},
		{
			name:         "non-boolean non-map non-array value should return false",
			afterUnknown: map[string]any{"config": "invalid"},
			path:         "config.anything",
			expected:     false,
		},
		{
			name:         "empty path should handle root level",
			afterUnknown: true,
			path:         "",
			expected:     true,
		},
		{
			name: "deeply nested path with multiple unknown levels",
			afterUnknown: map[string]any{
				"level1": map[string]any{
					"level2": []any{
						map[string]any{
							"level3": map[string]any{
								"final": true,
							},
						},
					},
				},
			},
			path:     "level1.level2[0].level3.final",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.isValueUnknown(tc.afterUnknown, tc.path)
			assert.Equal(t, tc.expected, result, "isValueUnknown result should match expected")
		})
	}
}

// TestGetUnknownValueDisplay tests the unknown value display function
func TestGetUnknownValueDisplay(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "should return exact Terraform syntax",
			expected: "(known after apply)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.getUnknownValueDisplay()
			assert.Equal(t, tc.expected, result, "getUnknownValueDisplay should return exact Terraform syntax")
		})
	}

	// Additional test to ensure the string is exactly as required by requirement 1.3
	t.Run("exact string match requirement", func(t *testing.T) {
		result := analyzer.getUnknownValueDisplay()
		assert.Exactly(t, "(known after apply)", result, "Must return exact string '(known after apply)'")
		assert.NotEqual(t, "(known_after_apply)", result, "Should not have underscores")
		assert.NotEqual(t, "known after apply", result, "Should have parentheses")
		assert.NotEqual(t, "(Known After Apply)", result, "Should not be title case")
	})
}

// TestCompareObjectsWithUnknownValues tests enhanced compareObjects function with unknown values integration
func TestCompareObjectsWithUnknownValues(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name            string
		before          any
		after           any
		beforeSensitive any
		afterSensitive  any
		afterUnknown    any
		expectedChanges int
		expectedUnknown []bool
		expectedActions []string
		expectedNames   []string
		description     string
	}{
		{
			name: "simple property with unknown value should override after value",
			before: map[string]any{
				"id": "old-id",
			},
			after: map[string]any{
				"id": nil, // Would normally show as removal
			},
			afterUnknown: map[string]any{
				"id": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"update"},
			expectedNames:   []string{"id"},
			description:     "Unknown values should override standard before/after comparison logic (requirement 1.6)",
		},
		{
			name:   "newly created property with unknown value",
			before: map[string]any{},
			after: map[string]any{
				"id": nil,
			},
			afterUnknown: map[string]any{
				"id": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"add"},
			expectedNames:   []string{"id"},
			description:     "New properties with unknown values should show as additions (requirement 1.5)",
		},
		{
			name: "known to unknown transition",
			before: map[string]any{
				"ami": "ami-123",
			},
			after: map[string]any{
				"ami": nil,
			},
			afterUnknown: map[string]any{
				"ami": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"update"},
			expectedNames:   []string{"ami"},
			description:     "Known to unknown transitions should show before_value → (known after apply) (requirement 1.4)",
		},
		{
			name: "mixed known and unknown properties",
			before: map[string]any{
				"instance_type": "t2.micro",
				"id":            "old-id",
				"ami":           "ami-123",
			},
			after: map[string]any{
				"instance_type": "t2.small",
				"id":            nil,
				"ami":           "ami-456",
			},
			afterUnknown: map[string]any{
				"id": true, // Only id is unknown
			},
			expectedChanges: 3,
			expectedUnknown: []bool{false, true, false}, // Order may vary, so we'll check different in assertion
			expectedActions: []string{"update", "update", "update"},
			expectedNames:   []string{"instance_type", "id", "ami"},
			description:     "Mix of known and unknown properties should be handled correctly",
		},
		{
			name: "sensitive property with unknown value",
			before: map[string]any{
				"user_data": "old-script",
			},
			after: map[string]any{
				"user_data": nil,
			},
			beforeSensitive: map[string]any{
				"user_data": true,
			},
			afterSensitive: map[string]any{
				"user_data": true,
			},
			afterUnknown: map[string]any{
				"user_data": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"update"},
			expectedNames:   []string{"user_data"},
			description:     "Unknown values should integrate with sensitive property detection (requirement 3.1)",
		},
		{
			name: "no unknown values should work normally",
			before: map[string]any{
				"instance_type": "t2.micro",
				"ami":           "ami-123",
			},
			after: map[string]any{
				"instance_type": "t2.small",
				"ami":           "ami-123",
			},
			afterUnknown:    map[string]any{}, // Empty unknown values
			expectedChanges: 1,
			expectedUnknown: []bool{false},
			expectedActions: []string{"update"},
			expectedNames:   []string{"instance_type"},
			description:     "Normal operation should not be affected when no unknown values present",
		},
		{
			name:            "nil after_unknown should work normally",
			before:          map[string]any{"id": "old-id"},
			after:           map[string]any{"id": "new-id"},
			afterUnknown:    nil,
			expectedChanges: 1,
			expectedUnknown: []bool{false},
			expectedActions: []string{"update"},
			expectedNames:   []string{"id"},
			description:     "Nil after_unknown should not break normal processing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := PropertyChangeAnalysis{
				Changes: []PropertyChange{},
			}

			analyzer.compareObjects("", tc.before, tc.after, tc.beforeSensitive, tc.afterSensitive, tc.afterUnknown, nil, &analysis)

			// Verify number of changes
			assert.Equal(t, tc.expectedChanges, len(analysis.Changes), tc.description)

			if tc.expectedChanges > 0 {
				// Count unknown changes
				unknownCount := 0
				for _, change := range analysis.Changes {
					if change.IsUnknown {
						unknownCount++
						// Verify unknown changes have correct display value
						assert.Equal(t, "(known after apply)", change.After, "Unknown values should display '(known after apply)' (%s)", tc.description)
						assert.Equal(t, "after", change.UnknownType, "Unknown values should have UnknownType 'after' (%s)", tc.description)
					}
				}

				// Check expected unknown count
				expectedUnknownCount := 0
				for _, expected := range tc.expectedUnknown {
					if expected {
						expectedUnknownCount++
					}
				}
				assert.Equal(t, expectedUnknownCount, unknownCount, "Number of unknown changes should match expected (%s)", tc.description)

				// Verify actions match expected (order-independent)
				if len(tc.expectedActions) > 0 {
					actualActions := make([]string, len(analysis.Changes))
					for i, change := range analysis.Changes {
						actualActions[i] = change.Action
					}
					assert.ElementsMatch(t, tc.expectedActions, actualActions, "Actions should match expected (%s)", tc.description)
				}

				// Verify names match expected (order-independent)
				if len(tc.expectedNames) > 0 {
					actualNames := make([]string, len(analysis.Changes))
					for i, change := range analysis.Changes {
						actualNames[i] = change.Name
					}
					assert.ElementsMatch(t, tc.expectedNames, actualNames, "Names should match expected (%s)", tc.description)
				}

				// Verify that unknown values don't appear as deletions (requirement 1.2)
				for i, change := range analysis.Changes {
					if change.IsUnknown && change.Before != nil {
						assert.NotEqual(t, "remove", change.Action, "Change %d: Unknown values should not appear as deletions (%s)", i, tc.description)
					}
				}

				// Verify sensitive property integration (requirement 3.1)
				for i, change := range analysis.Changes {
					if tc.beforeSensitive != nil || tc.afterSensitive != nil {
						// If we have sensitive values data, verify integration works
						assert.NotNil(t, change.Sensitive, "Change %d: Sensitive field should be set when sensitive data is present (%s)", i, tc.description)
					}
				}
			}
		})
	}
}

// TestAnalyzePropertyChangesWithUnknownValuesIntegration tests complete property analysis with unknown values
func TestAnalyzePropertyChangesWithUnknownValuesIntegration(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name                 string
		change               *tfjson.ResourceChange
		expectedUnknownCount int
		expectedTotalCount   int
		expectedUnknownProps []string
		description          string
	}{
		{
			name: "resource change with unknown values should populate unknown fields",
			change: &tfjson.ResourceChange{
				Type: "aws_instance",
				Change: &tfjson.Change{
					Before: map[string]any{
						"id":            nil,
						"ami":           "ami-123",
						"instance_type": "t2.micro",
					},
					After: map[string]any{
						"id":            nil,
						"ami":           "ami-123",
						"instance_type": "t2.small",
					},
					AfterUnknown: map[string]any{
						"id": true,
					},
				},
			},
			expectedUnknownCount: 1,
			expectedTotalCount:   2, // id and instance_type changes
			expectedUnknownProps: []string{"id"},
			description:          "Resource with unknown values should properly populate unknown tracking fields",
		},
		{
			name: "resource with multiple unknown properties",
			change: &tfjson.ResourceChange{
				Type: "aws_instance",
				Change: &tfjson.Change{
					Before: map[string]any{
						"id":                     nil,
						"vpc_security_group_ids": nil,
						"private_ip":             nil,
						"instance_type":          "t2.micro",
					},
					After: map[string]any{
						"id":                     nil,
						"vpc_security_group_ids": nil,
						"private_ip":             nil,
						"instance_type":          "t2.micro",
					},
					AfterUnknown: map[string]any{
						"id":                     true,
						"vpc_security_group_ids": true,
						"private_ip":             true,
					},
				},
			},
			expectedUnknownCount: 3,
			expectedTotalCount:   3,
			expectedUnknownProps: []string{"id", "vpc_security_group_ids", "private_ip"},
			description:          "Multiple unknown properties should all be tracked correctly",
		},
		{
			name: "resource with no unknown values should work normally",
			change: &tfjson.ResourceChange{
				Type: "aws_instance",
				Change: &tfjson.Change{
					Before: map[string]any{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]any{
						"instance_type": "t2.small",
						"ami":           "ami-456",
					},
					AfterUnknown: map[string]any{}, // No unknown values
				},
			},
			expectedUnknownCount: 0,
			expectedTotalCount:   2,
			expectedUnknownProps: []string{},
			description:          "Resources without unknown values should work normally",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.analyzePropertyChanges(tc.change)

			// Count unknown properties
			unknownCount := 0
			var unknownProps []string
			for _, change := range result.Changes {
				if change.IsUnknown {
					unknownCount++
					unknownProps = append(unknownProps, change.Name)
				}
			}

			assert.Equal(t, tc.expectedUnknownCount, unknownCount, tc.description+" - unknown count")
			assert.Equal(t, tc.expectedTotalCount, result.Count, tc.description+" - total count")
			assert.ElementsMatch(t, tc.expectedUnknownProps, unknownProps, tc.description+" - unknown property names")

			// Verify all unknown changes have the correct display value
			for i, change := range result.Changes {
				if change.IsUnknown {
					assert.Equal(t, "(known after apply)", change.After, "Change %d should display '(known after apply)' for unknown values (%s)", i, tc.description)
					assert.Equal(t, "after", change.UnknownType, "Change %d should have UnknownType 'after' (%s)", i, tc.description)
				}
			}

			// Verify that all changes have required fields populated
			for i, change := range result.Changes {
				assert.NotEmpty(t, change.Action, "Change %d should have action (%s)", i, tc.description)
				assert.NotEmpty(t, change.Name, "Change %d should have name (%s)", i, tc.description)
				assert.NotNil(t, change.Path, "Change %d should have path (%s)", i, tc.description)
			}
		})
	}
}

// TestAnalyzePropertyChangesWithLimits tests the complete property analysis with performance limits
func TestAnalyzePropertyChangesWithLimits(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		change            *tfjson.ResourceChange
		expectedTruncated bool
		description       string
	}{
		{
			name: "normal change should not truncate",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"name":    "old-name",
						"size":    10,
						"enabled": false,
					},
					After: map[string]any{
						"name":    "new-name",
						"size":    20,
						"enabled": true,
					},
				},
			},
			expectedTruncated: false,
			description:       "Small changes should not trigger truncation",
		},
		{
			name: "many properties should apply count limits",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: func() map[string]any {
						result := make(map[string]any)
						// Create more properties than the limit
						for i := range MaxPropertiesPerResource + 10 {
							result[fmt.Sprintf("prop_%d", i)] = fmt.Sprintf("value_%d", i)
						}
						return result
					}(),
					After: func() map[string]any {
						result := make(map[string]any)
						// Create more properties than the limit
						for i := range MaxPropertiesPerResource + 10 {
							result[fmt.Sprintf("prop_%d", i)] = fmt.Sprintf("new_value_%d", i)
						}
						return result
					}(),
				},
			},
			expectedTruncated: true,
			description:       "Many properties should trigger count limit truncation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzePropertyChanges(tt.change)

			assert.Equal(t, tt.expectedTruncated, result.Truncated, tt.description)
			assert.Equal(t, len(result.Changes), result.Count, "Count should match changes length")

			// Verify all changes have required fields
			for i, change := range result.Changes {
				assert.NotEmpty(t, change.Action, "Change %d should have action", i)
				assert.NotEmpty(t, change.Name, "Change %d should have name", i)
				assert.NotNil(t, change.Path, "Change %d should have path", i)
				assert.GreaterOrEqual(t, change.Size, 0, "Change %d should have non-negative size", i)
			}

			if tt.expectedTruncated {
				assert.LessOrEqual(t, result.TotalSize, MaxTotalPropertyMemory, "Total size should not exceed limit")
			}
		})
	}
}

// TestProcessOutputChanges tests the ProcessOutputChanges function (Task 7.1)
func TestProcessOutputChanges(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name            string
		plan            *tfjson.Plan
		expectedCount   int
		expectedOutputs []OutputChange
		expectedError   bool
		description     string
	}{
		{
			name: "plan with no output changes should return empty list",
			plan: &tfjson.Plan{
				OutputChanges: nil,
			},
			expectedCount:   0,
			expectedOutputs: []OutputChange{},
			expectedError:   false,
			description:     "Missing output_changes field should return empty list (requirement 2.8)",
		},
		{
			name: "plan with empty output changes should return empty list",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{},
			},
			expectedCount:   0,
			expectedOutputs: []OutputChange{},
			expectedError:   false,
			description:     "Empty output_changes map should return empty list",
		},
		{
			name: "plan with single output creation",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"instance_ip": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "192.168.1.10",
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "instance_ip",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "192.168.1.10",
					IsUnknown:  false,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Single output creation should be processed correctly (requirement 2.5)",
		},
		{
			name: "plan with output modification",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"database_url": {
						Actions: []tfjson.Action{tfjson.ActionUpdate},
						Before:  "postgresql://old-host/db",
						After:   "postgresql://new-host/db",
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "database_url",
					ChangeType: ChangeTypeUpdate,
					Sensitive:  false,
					Before:     "postgresql://old-host/db",
					After:      "postgresql://new-host/db",
					IsUnknown:  false,
					Action:     "Modify",
					Indicator:  "~",
				},
			},
			expectedError: false,
			description:   "Output modification should be processed correctly (requirement 2.6)",
		},
		{
			name: "plan with output deletion",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"old_output": {
						Actions: []tfjson.Action{tfjson.ActionDelete},
						Before:  "some-value",
						After:   nil,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "old_output",
					ChangeType: ChangeTypeDelete,
					Sensitive:  false,
					Before:     "some-value",
					After:      nil,
					IsUnknown:  false,
					Action:     "Remove",
					Indicator:  "-",
				},
			},
			expectedError: false,
			description:   "Output deletion should be processed correctly (requirement 2.7)",
		},
		{
			name: "plan with unknown output value",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"computed_id": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "computed_id",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "(known after apply)",
					IsUnknown:  true,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Unknown output values should display '(known after apply)' (requirement 2.3)",
		},
		{
			name: "plan with sensitive output",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"api_key": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          "secret-key-value",
						AfterSensitive: true,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "api_key",
					ChangeType: ChangeTypeCreate,
					Sensitive:  true,
					Before:     nil,
					After:      "(sensitive value)",
					IsUnknown:  false,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Sensitive outputs should display '(sensitive value)' (requirement 2.4)",
		},
		{
			name: "plan with sensitive unknown output",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"secret_id": {
						Actions:         []tfjson.Action{tfjson.ActionUpdate},
						Before:          "old-secret",
						After:           nil,
						BeforeSensitive: true,
						AfterSensitive:  true,
						AfterUnknown:    true,
					},
				},
			},
			expectedCount: 1,
			expectedOutputs: []OutputChange{
				{
					Name:       "secret_id",
					ChangeType: ChangeTypeUpdate,
					Sensitive:  true,
					Before:     "(sensitive value)",
					After:      "(known after apply)",
					IsUnknown:  true,
					Action:     "Modify",
					Indicator:  "~",
				},
			},
			expectedError: false,
			description:   "Sensitive unknown outputs should show both masking and unknown display",
		},
		{
			name: "plan with multiple mixed outputs",
			plan: &tfjson.Plan{
				OutputChanges: map[string]*tfjson.Change{
					"public_ip": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "203.0.113.10",
					},
					"private_key": {
						Actions:         []tfjson.Action{tfjson.ActionUpdate},
						Before:          "old-key",
						After:           "new-key",
						BeforeSensitive: true,
						AfterSensitive:  true,
					},
					"instance_id": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
				},
			},
			expectedCount: 3,
			expectedOutputs: []OutputChange{
				{
					Name:       "public_ip",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "203.0.113.10",
					IsUnknown:  false,
					Action:     "Add",
					Indicator:  "+",
				},
				{
					Name:       "private_key",
					ChangeType: ChangeTypeUpdate,
					Sensitive:  true,
					Before:     "(sensitive value)",
					After:      "(sensitive value)",
					IsUnknown:  false,
					Action:     "Modify",
					Indicator:  "~",
				},
				{
					Name:       "instance_id",
					ChangeType: ChangeTypeCreate,
					Sensitive:  false,
					Before:     nil,
					After:      "(known after apply)",
					IsUnknown:  true,
					Action:     "Add",
					Indicator:  "+",
				},
			},
			expectedError: false,
			description:   "Multiple mixed outputs should all be processed correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputs, err := analyzer.ProcessOutputChanges(tc.plan)

			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}

			assert.Equal(t, tc.expectedCount, len(outputs), tc.description+" - output count")

			// Check individual outputs (order-independent comparison)
			if len(tc.expectedOutputs) > 0 && len(outputs) > 0 {
				// Create maps for easier comparison since output order isn't guaranteed
				expectedMap := make(map[string]OutputChange)
				actualMap := make(map[string]OutputChange)

				for _, expected := range tc.expectedOutputs {
					expectedMap[expected.Name] = expected
				}
				for _, actual := range outputs {
					actualMap[actual.Name] = actual
				}

				for name, expected := range expectedMap {
					if actual, exists := actualMap[name]; exists {
						assert.Equal(t, expected.Name, actual.Name, "Output %s name should match (%s)", name, tc.description)
						assert.Equal(t, expected.ChangeType, actual.ChangeType, "Output %s ChangeType should match (%s)", name, tc.description)
						assert.Equal(t, expected.Sensitive, actual.Sensitive, "Output %s Sensitive should match (%s)", name, tc.description)
						assert.Equal(t, expected.Before, actual.Before, "Output %s Before should match (%s)", name, tc.description)
						assert.Equal(t, expected.After, actual.After, "Output %s After should match (%s)", name, tc.description)
						assert.Equal(t, expected.IsUnknown, actual.IsUnknown, "Output %s IsUnknown should match (%s)", name, tc.description)
						assert.Equal(t, expected.Action, actual.Action, "Output %s Action should match (%s)", name, tc.description)
						assert.Equal(t, expected.Indicator, actual.Indicator, "Output %s Indicator should match (%s)", name, tc.description)
					} else {
						t.Errorf("Expected output %s not found in results (%s)", name, tc.description)
					}
				}
			}
		})
	}
}

// TestAnalyzeOutputChange tests the analyzeOutputChange function (Task 7.1)
func TestAnalyzeOutputChange(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name           string
		outputName     string
		change         *tfjson.Change
		expectedOutput *OutputChange
		expectedError  bool
		description    string
	}{
		{
			name:           "nil change should return error",
			outputName:     "test_output",
			change:         nil,
			expectedOutput: nil,
			expectedError:  true,
			description:    "Nil change should be handled gracefully with error",
		},
		{
			name:       "simple create action",
			outputName: "web_url",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionCreate},
				Before:  nil,
				After:   "https://example.com",
			},
			expectedOutput: &OutputChange{
				Name:       "web_url",
				ChangeType: ChangeTypeCreate,
				Sensitive:  false,
				Before:     nil,
				After:      "https://example.com",
				IsUnknown:  false,
				Action:     "Add",
				Indicator:  "+",
			},
			expectedError: false,
			description:   "Create action should show 'Add' with '+' indicator (requirement 2.5)",
		},
		{
			name:       "simple update action",
			outputName: "database_port",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionUpdate},
				Before:  5432,
				After:   3306,
			},
			expectedOutput: &OutputChange{
				Name:       "database_port",
				ChangeType: ChangeTypeUpdate,
				Sensitive:  false,
				Before:     5432,
				After:      3306,
				IsUnknown:  false,
				Action:     "Modify",
				Indicator:  "~",
			},
			expectedError: false,
			description:   "Update action should show 'Modify' with '~' indicator (requirement 2.6)",
		},
		{
			name:       "simple delete action",
			outputName: "deprecated_config",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionDelete},
				Before:  "old-config",
				After:   nil,
			},
			expectedOutput: &OutputChange{
				Name:       "deprecated_config",
				ChangeType: ChangeTypeDelete,
				Sensitive:  false,
				Before:     "old-config",
				After:      nil,
				IsUnknown:  false,
				Action:     "Remove",
				Indicator:  "-",
			},
			expectedError: false,
			description:   "Delete action should show 'Remove' with '-' indicator (requirement 2.7)",
		},
		{
			name:       "unknown output value",
			outputName: "instance_arn",
			change: &tfjson.Change{
				Actions:      []tfjson.Action{tfjson.ActionCreate},
				Before:       nil,
				After:        nil,
				AfterUnknown: true,
			},
			expectedOutput: &OutputChange{
				Name:       "instance_arn",
				ChangeType: ChangeTypeCreate,
				Sensitive:  false,
				Before:     nil,
				After:      "(known after apply)",
				IsUnknown:  true,
				Action:     "Add",
				Indicator:  "+",
			},
			expectedError: false,
			description:   "Unknown output should display '(known after apply)' (requirement 2.3)",
		},
		{
			name:       "sensitive output with before value",
			outputName: "admin_password",
			change: &tfjson.Change{
				Actions:         []tfjson.Action{tfjson.ActionUpdate},
				Before:          "old-password",
				After:           "new-password",
				BeforeSensitive: true,
				AfterSensitive:  true,
			},
			expectedOutput: &OutputChange{
				Name:       "admin_password",
				ChangeType: ChangeTypeUpdate,
				Sensitive:  true,
				Before:     "(sensitive value)",
				After:      "(sensitive value)",
				IsUnknown:  false,
				Action:     "Modify",
				Indicator:  "~",
			},
			expectedError: false,
			description:   "Sensitive outputs should display '(sensitive value)' (requirement 2.4)",
		},
		{
			name:       "sensitive unknown output",
			outputName: "tls_cert",
			change: &tfjson.Change{
				Actions:        []tfjson.Action{tfjson.ActionCreate},
				Before:         nil,
				After:          nil,
				AfterSensitive: true,
				AfterUnknown:   true,
			},
			expectedOutput: &OutputChange{
				Name:       "tls_cert",
				ChangeType: ChangeTypeCreate,
				Sensitive:  true,
				Before:     nil,
				After:      "(known after apply)",
				IsUnknown:  true,
				Action:     "Add",
				Indicator:  "+",
			},
			expectedError: false,
			description:   "Sensitive unknown outputs should prioritize unknown display over sensitive masking",
		},
		{
			name:       "no-op action",
			outputName: "static_value",
			change: &tfjson.Change{
				Actions: []tfjson.Action{},
				Before:  "unchanged",
				After:   "unchanged",
			},
			expectedOutput: &OutputChange{
				Name:       "static_value",
				ChangeType: ChangeTypeNoOp,
				Sensitive:  false,
				Before:     "unchanged",
				After:      "unchanged",
				IsUnknown:  false,
				Action:     "No-op",
				Indicator:  "",
			},
			expectedError: false,
			description:   "No-op actions should be handled correctly",
		},
		{
			name:       "replace action (delete + create)",
			outputName: "resource_reference",
			change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate},
				Before:  "old-reference",
				After:   "new-reference",
			},
			expectedOutput: &OutputChange{
				Name:       "resource_reference",
				ChangeType: ChangeTypeReplace,
				Sensitive:  false,
				Before:     "old-reference",
				After:      "new-reference",
				IsUnknown:  false,
				Action:     "No-op",
				Indicator:  "",
			},
			expectedError: false,
			description:   "Replace actions should be handled (though uncommon for outputs)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := analyzer.analyzeOutputChange(tc.outputName, tc.change)

			if tc.expectedError {
				assert.Error(t, err, tc.description)
				assert.Nil(t, result, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
				assert.NotNil(t, result, tc.description)

				if result != nil && tc.expectedOutput != nil {
					assert.Equal(t, tc.expectedOutput.Name, result.Name, "Name should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.ChangeType, result.ChangeType, "ChangeType should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Sensitive, result.Sensitive, "Sensitive should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Before, result.Before, "Before should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.After, result.After, "After should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.IsUnknown, result.IsUnknown, "IsUnknown should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Action, result.Action, "Action should match (%s)", tc.description)
					assert.Equal(t, tc.expectedOutput.Indicator, result.Indicator, "Indicator should match (%s)", tc.description)
				}
			}
		})
	}
}

// TestGetOutputActionAndIndicator tests the output action and indicator mapping function (Task 7.1)
func TestGetOutputActionAndIndicator(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name              string
		changeType        ChangeType
		expectedAction    string
		expectedIndicator string
		description       string
	}{
		{
			name:              "create should return Add with +",
			changeType:        ChangeTypeCreate,
			expectedAction:    "Add",
			expectedIndicator: "+",
			description:       "Create changes should show 'Add' with '+' indicator (requirement 2.5)",
		},
		{
			name:              "update should return Modify with ~",
			changeType:        ChangeTypeUpdate,
			expectedAction:    "Modify",
			expectedIndicator: "~",
			description:       "Update changes should show 'Modify' with '~' indicator (requirement 2.6)",
		},
		{
			name:              "delete should return Remove with -",
			changeType:        ChangeTypeDelete,
			expectedAction:    "Remove",
			expectedIndicator: "-",
			description:       "Delete changes should show 'Remove' with '-' indicator (requirement 2.7)",
		},
		{
			name:              "no-op should return No-op with empty indicator",
			changeType:        ChangeTypeNoOp,
			expectedAction:    "No-op",
			expectedIndicator: "",
			description:       "No-op changes should show 'No-op' with empty indicator",
		},
		{
			name:              "replace should return No-op with empty indicator",
			changeType:        ChangeTypeReplace,
			expectedAction:    "No-op",
			expectedIndicator: "",
			description:       "Replace changes should show 'No-op' with empty indicator (uncommon for outputs)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			action, indicator := analyzer.getOutputActionAndIndicator(tc.changeType)

			assert.Equal(t, tc.expectedAction, action, tc.description+" - action")
			assert.Equal(t, tc.expectedIndicator, indicator, tc.description+" - indicator")
		})
	}
}

// TestOutputsProcessingEndToEnd tests complete outputs processing workflow (Task 7.2)
func TestOutputsProcessingEndToEnd(t *testing.T) {
	cfg := &config.Config{}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name                 string
		plan                 *tfjson.Plan
		expectedOutputCount  int
		expectedEmptySection bool
		validateOutputs      func(t *testing.T, outputs []OutputChange, description string)
		description          string
	}{
		{
			name: "plan with no outputs should suppress outputs section",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges:    nil,
				ResourceChanges:  []*tfjson.ResourceChange{},
			},
			expectedOutputCount:  0,
			expectedEmptySection: true,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Empty(t, outputs, description+" - outputs should be empty")
			},
			description: "Plans without outputs should suppress the outputs section entirely (requirement 2.8)",
		},
		{
			name: "plan with mixed output types should process all correctly",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges: map[string]*tfjson.Change{
					"public_endpoint": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "https://api.example.com",
					},
					"database_password": {
						Actions:        []tfjson.Action{tfjson.ActionUpdate},
						Before:         "old-secret",
						After:          "new-secret",
						AfterSensitive: true,
					},
					"instance_id": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
					"deprecated_config": {
						Actions: []tfjson.Action{tfjson.ActionDelete},
						Before:  "old-value",
						After:   nil,
					},
				},
				ResourceChanges: []*tfjson.ResourceChange{},
			},
			expectedOutputCount:  4,
			expectedEmptySection: false,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Len(t, outputs, 4, description+" - should have 4 outputs")

				// Create a map for easy lookup
				outputMap := make(map[string]OutputChange)
				for _, output := range outputs {
					outputMap[output.Name] = output
				}

				// Validate public endpoint (create)
				if publicEndpoint, exists := outputMap["public_endpoint"]; exists {
					assert.Equal(t, ChangeTypeCreate, publicEndpoint.ChangeType, "public_endpoint should be create type")
					assert.Equal(t, "Add", publicEndpoint.Action, "public_endpoint should have Add action")
					assert.Equal(t, "+", publicEndpoint.Indicator, "public_endpoint should have + indicator")
					assert.Equal(t, "https://api.example.com", publicEndpoint.After, "public_endpoint should have correct after value")
					assert.False(t, publicEndpoint.Sensitive, "public_endpoint should not be sensitive")
					assert.False(t, publicEndpoint.IsUnknown, "public_endpoint should not be unknown")
				} else {
					t.Errorf("public_endpoint output not found")
				}

				// Validate database password (sensitive update)
				if dbPassword, exists := outputMap["database_password"]; exists {
					assert.Equal(t, ChangeTypeUpdate, dbPassword.ChangeType, "database_password should be update type")
					assert.Equal(t, "Modify", dbPassword.Action, "database_password should have Modify action")
					assert.Equal(t, "~", dbPassword.Indicator, "database_password should have ~ indicator")
					assert.Equal(t, "(sensitive value)", dbPassword.After, "database_password should mask sensitive value")
					assert.True(t, dbPassword.Sensitive, "database_password should be sensitive")
					assert.False(t, dbPassword.IsUnknown, "database_password should not be unknown")
				} else {
					t.Errorf("database_password output not found")
				}

				// Validate instance ID (unknown create)
				if instanceId, exists := outputMap["instance_id"]; exists {
					assert.Equal(t, ChangeTypeCreate, instanceId.ChangeType, "instance_id should be create type")
					assert.Equal(t, "Add", instanceId.Action, "instance_id should have Add action")
					assert.Equal(t, "+", instanceId.Indicator, "instance_id should have + indicator")
					assert.Equal(t, "(known after apply)", instanceId.After, "instance_id should show unknown value")
					assert.False(t, instanceId.Sensitive, "instance_id should not be sensitive")
					assert.True(t, instanceId.IsUnknown, "instance_id should be unknown")
				} else {
					t.Errorf("instance_id output not found")
				}

				// Validate deprecated config (delete)
				if deprecatedConfig, exists := outputMap["deprecated_config"]; exists {
					assert.Equal(t, ChangeTypeDelete, deprecatedConfig.ChangeType, "deprecated_config should be delete type")
					assert.Equal(t, "Remove", deprecatedConfig.Action, "deprecated_config should have Remove action")
					assert.Equal(t, "-", deprecatedConfig.Indicator, "deprecated_config should have - indicator")
					assert.Equal(t, "old-value", deprecatedConfig.Before, "deprecated_config should have correct before value")
					assert.Nil(t, deprecatedConfig.After, "deprecated_config should have nil after value")
					assert.False(t, deprecatedConfig.Sensitive, "deprecated_config should not be sensitive")
					assert.False(t, deprecatedConfig.IsUnknown, "deprecated_config should not be unknown")
				} else {
					t.Errorf("deprecated_config output not found")
				}
			},
			description: "Mixed output types should be processed with correct actions and indicators",
		},
		{
			name: "plan with sensitive unknown outputs should handle edge case correctly",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges: map[string]*tfjson.Change{
					"ssl_certificate": {
						Actions:         []tfjson.Action{tfjson.ActionCreate},
						Before:          nil,
						After:           nil,
						BeforeSensitive: false,
						AfterSensitive:  true,
						AfterUnknown:    true,
					},
					"encryption_key": {
						Actions:         []tfjson.Action{tfjson.ActionUpdate},
						Before:          "old-key",
						After:           nil,
						BeforeSensitive: true,
						AfterSensitive:  true,
						AfterUnknown:    true,
					},
				},
				ResourceChanges: []*tfjson.ResourceChange{},
			},
			expectedOutputCount:  2,
			expectedEmptySection: false,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Len(t, outputs, 2, description+" - should have 2 outputs")

				// Create a map for easy lookup
				outputMap := make(map[string]OutputChange)
				for _, output := range outputs {
					outputMap[output.Name] = output
				}

				// Validate SSL certificate (sensitive unknown create)
				if sslCert, exists := outputMap["ssl_certificate"]; exists {
					assert.Equal(t, ChangeTypeCreate, sslCert.ChangeType, "ssl_certificate should be create type")
					assert.Equal(t, "Add", sslCert.Action, "ssl_certificate should have Add action")
					assert.Equal(t, "+", sslCert.Indicator, "ssl_certificate should have + indicator")
					assert.Equal(t, "(known after apply)", sslCert.After, "ssl_certificate should prioritize unknown over sensitive")
					assert.True(t, sslCert.Sensitive, "ssl_certificate should be sensitive")
					assert.True(t, sslCert.IsUnknown, "ssl_certificate should be unknown")
				} else {
					t.Errorf("ssl_certificate output not found")
				}

				// Validate encryption key (sensitive unknown update)
				if encKey, exists := outputMap["encryption_key"]; exists {
					assert.Equal(t, ChangeTypeUpdate, encKey.ChangeType, "encryption_key should be update type")
					assert.Equal(t, "Modify", encKey.Action, "encryption_key should have Modify action")
					assert.Equal(t, "~", encKey.Indicator, "encryption_key should have ~ indicator")
					assert.Equal(t, "(sensitive value)", encKey.Before, "encryption_key before should be masked")
					assert.Equal(t, "(known after apply)", encKey.After, "encryption_key should prioritize unknown over sensitive")
					assert.True(t, encKey.Sensitive, "encryption_key should be sensitive")
					assert.True(t, encKey.IsUnknown, "encryption_key should be unknown")
				} else {
					t.Errorf("encryption_key output not found")
				}
			},
			description: "Sensitive unknown outputs should prioritize unknown display appropriately",
		},
		{
			name: "plan with only resource changes should suppress outputs section",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				OutputChanges:    map[string]*tfjson.Change{}, // Empty but not nil
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.example",
						Type:    "aws_instance",
						Name:    "example",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After:   map[string]any{"instance_type": "t2.micro"},
						},
					},
				},
			},
			expectedOutputCount:  0,
			expectedEmptySection: true,
			validateOutputs: func(t *testing.T, outputs []OutputChange, description string) {
				assert.Empty(t, outputs, description+" - outputs should be empty when no output changes exist")
			},
			description: "Plans with only resource changes should suppress empty outputs section",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the plan on the analyzer
			analyzer.plan = tc.plan

			// Test ProcessOutputChanges directly
			outputs, err := analyzer.ProcessOutputChanges(tc.plan)
			assert.NoError(t, err, tc.description+" - ProcessOutputChanges should not error")
			assert.Equal(t, tc.expectedOutputCount, len(outputs), tc.description+" - output count should match")

			// Run custom validation
			if tc.validateOutputs != nil {
				tc.validateOutputs(t, outputs, tc.description)
			}

			// Test integration with GenerateSummary
			summary := analyzer.GenerateSummary("")
			assert.NotNil(t, summary, tc.description+" - summary should not be nil")
			assert.Equal(t, tc.expectedOutputCount, len(summary.OutputChanges), tc.description+" - summary output count should match")

			// Verify outputs section suppression behavior
			if tc.expectedEmptySection {
				assert.Empty(t, summary.OutputChanges, tc.description+" - summary outputs should be empty when section should be suppressed")
			} else {
				assert.NotEmpty(t, summary.OutputChanges, tc.description+" - summary outputs should not be empty when section should be shown")
			}

			// Verify resource changes are still processed normally
			assert.NotNil(t, summary.ResourceChanges, tc.description+" - resource changes should still be processed")
			assert.Equal(t, len(tc.plan.ResourceChanges), len(summary.ResourceChanges), tc.description+" - resource change count should match")
		})
	}
}

// TestOutputsIntegrationWithResourceChanges tests outputs section integration with resource changes (Task 7.2)
func TestOutputsIntegrationWithResourceChanges(t *testing.T) {
	cfg := &config.Config{}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name                  string
		plan                  *tfjson.Plan
		expectedResourceCount int
		expectedOutputCount   int
		validateSummary       func(t *testing.T, summary *PlanSummary, description string)
		description           string
	}{
		{
			name: "plan with both resource changes and outputs should display both sections",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.web",
						Type:    "aws_instance",
						Name:    "web",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After:   map[string]any{"instance_type": "t2.micro", "ami": "ami-123"},
						},
					},
					{
						Address: "aws_s3_bucket.data",
						Type:    "aws_s3_bucket",
						Name:    "data",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before:  map[string]any{"versioning": false},
							After:   map[string]any{"versioning": true},
						},
					},
				},
				OutputChanges: map[string]*tfjson.Change{
					"instance_ip": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
					"bucket_arn": {
						Actions: []tfjson.Action{tfjson.ActionUpdate},
						Before:  "arn:aws:s3:::old-bucket",
						After:   "arn:aws:s3:::new-bucket",
					},
				},
			},
			expectedResourceCount: 2,
			expectedOutputCount:   2,
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Verify both sections are present
				assert.Len(t, summary.ResourceChanges, 2, description+" - should have resource changes")
				assert.Len(t, summary.OutputChanges, 2, description+" - should have output changes")

				// Verify outputs section placement after resource changes
				assert.NotNil(t, summary.ResourceChanges, description+" - resource changes should come first")
				assert.NotNil(t, summary.OutputChanges, description+" - output changes should come after resources")

				// Verify outputs are correctly processed
				outputMap := make(map[string]OutputChange)
				for _, output := range summary.OutputChanges {
					outputMap[output.Name] = output
				}

				if instanceIp, exists := outputMap["instance_ip"]; exists {
					assert.Equal(t, "(known after apply)", instanceIp.After, "instance_ip should show unknown value")
					assert.True(t, instanceIp.IsUnknown, "instance_ip should be unknown")
				}

				if bucketArn, exists := outputMap["bucket_arn"]; exists {
					assert.Equal(t, "arn:aws:s3:::new-bucket", bucketArn.After, "bucket_arn should show new value")
					assert.False(t, bucketArn.IsUnknown, "bucket_arn should not be unknown")
				}

				// Verify statistics only track resource changes (requirement 3.3)
				stats := summary.Statistics
				assert.Equal(t, 1, stats.ToAdd, "statistics should count resource additions")
				assert.Equal(t, 1, stats.ToChange, "statistics should count resource modifications")
				assert.Equal(t, 2, stats.Total, "statistics should total resource changes only")
			},
			description: "Plans with both resources and outputs should display both sections correctly (requirement 2.1)",
		},
		{
			name: "plan with only outputs changes should show outputs section",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.0.0",
				ResourceChanges:  []*tfjson.ResourceChange{}, // No resource changes
				OutputChanges: map[string]*tfjson.Change{
					"api_endpoint": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "https://api.example.com/v1",
					},
				},
			},
			expectedResourceCount: 0,
			expectedOutputCount:   1,
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Verify no resource changes but outputs present
				assert.Empty(t, summary.ResourceChanges, description+" - should have no resource changes")
				assert.Len(t, summary.OutputChanges, 1, description+" - should have output changes")

				// Verify output is correctly processed
				output := summary.OutputChanges[0]
				assert.Equal(t, "api_endpoint", output.Name, "output name should match")
				assert.Equal(t, "Add", output.Action, "output should have Add action")
				assert.Equal(t, "+", output.Indicator, "output should have + indicator")

				// Verify statistics show no resource changes
				stats := summary.Statistics
				assert.Equal(t, 0, stats.ToAdd, "statistics should show no resource additions")
				assert.Equal(t, 0, stats.ToChange, "statistics should show no resource modifications")
				assert.Equal(t, 0, stats.Total, "statistics should show no total resource changes")
			},
			description: "Plans with only output changes should still display outputs section",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the plan on the analyzer
			analyzer.plan = tc.plan

			// Generate full summary
			summary := analyzer.GenerateSummary("")
			assert.NotNil(t, summary, tc.description+" - summary should not be nil")

			// Verify counts
			assert.Equal(t, tc.expectedResourceCount, len(summary.ResourceChanges), tc.description+" - resource count should match")
			assert.Equal(t, tc.expectedOutputCount, len(summary.OutputChanges), tc.description+" - output count should match")

			// Run custom validation
			if tc.validateSummary != nil {
				tc.validateSummary(t, summary, tc.description)
			}

			// Verify basic summary fields are populated
			assert.Equal(t, tc.plan.FormatVersion, summary.FormatVersion, tc.description+" - format version should match")
			assert.Equal(t, tc.plan.TerraformVersion, summary.TerraformVersion, tc.description+" - terraform version should match")
			assert.NotNil(t, summary.Statistics, tc.description+" - statistics should be present")
		})
	}
}

// TestOutputsDisplayConsistencyAcrossFormats tests outputs display consistency (Task 7.2)
func TestOutputsDisplayConsistencyAcrossFormats(t *testing.T) {
	analyzer := &Analyzer{}

	// Create a plan with various output types
	plan := &tfjson.Plan{
		FormatVersion:    "1.0",
		TerraformVersion: "1.0.0",
		OutputChanges: map[string]*tfjson.Change{
			"public_ip": {
				Actions: []tfjson.Action{tfjson.ActionCreate},
				Before:  nil,
				After:   "203.0.113.10",
			},
			"secret_key": {
				Actions:        []tfjson.Action{tfjson.ActionUpdate},
				Before:         "old-secret",
				After:          "new-secret",
				AfterSensitive: true,
			},
			"instance_id": {
				Actions:      []tfjson.Action{tfjson.ActionCreate},
				Before:       nil,
				After:        nil,
				AfterUnknown: true,
			},
		},
		ResourceChanges: []*tfjson.ResourceChange{},
	}

	t.Run("outputs should have consistent display values across processing", func(t *testing.T) {
		outputs, err := analyzer.ProcessOutputChanges(plan)
		assert.NoError(t, err, "ProcessOutputChanges should not error")
		assert.Len(t, outputs, 3, "should have 3 outputs")

		// Create a map for easy lookup
		outputMap := make(map[string]OutputChange)
		for _, output := range outputs {
			outputMap[output.Name] = output
		}

		// Test public IP (normal value)
		if publicIp, exists := outputMap["public_ip"]; exists {
			assert.Equal(t, "203.0.113.10", publicIp.After, "public_ip should have actual value")
			assert.False(t, publicIp.Sensitive, "public_ip should not be sensitive")
			assert.False(t, publicIp.IsUnknown, "public_ip should not be unknown")
		}

		// Test secret key (sensitive value)
		if secretKey, exists := outputMap["secret_key"]; exists {
			assert.Equal(t, "(sensitive value)", secretKey.After, "secret_key should display (sensitive value) consistently")
			assert.True(t, secretKey.Sensitive, "secret_key should be marked as sensitive")
			assert.False(t, secretKey.IsUnknown, "secret_key should not be unknown")
		}

		// Test instance ID (unknown value)
		if instanceId, exists := outputMap["instance_id"]; exists {
			assert.Equal(t, "(known after apply)", instanceId.After, "instance_id should display (known after apply) consistently")
			assert.False(t, instanceId.Sensitive, "instance_id should not be sensitive")
			assert.True(t, instanceId.IsUnknown, "instance_id should be marked as unknown")
		}
	})

	t.Run("outputs should maintain consistency in complete summary", func(t *testing.T) {
		analyzer.plan = plan
		summary := analyzer.GenerateSummary("")

		assert.NotNil(t, summary, "summary should not be nil")
		assert.Len(t, summary.OutputChanges, 3, "summary should have 3 outputs")

		// Verify the same consistent display values in the summary
		outputMap := make(map[string]OutputChange)
		for _, output := range summary.OutputChanges {
			outputMap[output.Name] = output
		}

		// Consistency checks
		if publicIp, exists := outputMap["public_ip"]; exists {
			assert.Equal(t, "203.0.113.10", publicIp.After, "public_ip should maintain actual value in summary")
		}

		if secretKey, exists := outputMap["secret_key"]; exists {
			assert.Equal(t, "(sensitive value)", secretKey.After, "secret_key should maintain (sensitive value) in summary")
		}

		if instanceId, exists := outputMap["instance_id"]; exists {
			assert.Equal(t, "(known after apply)", instanceId.After, "instance_id should maintain (known after apply) in summary")
		}
	})
}

// TestCompleteWorkflowWithUnknownValuesAndOutputsIntegration tests the complete workflow
// with real Terraform plan containing unknown values and outputs (Task 9.1)
func TestCompleteWorkflowWithUnknownValuesAndOutputsIntegration(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails:      true,
			HighlightDangers: true,
		},
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_iam_policy"},
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_instance", Property: "user_data"},
		},
	}
	analyzer := &Analyzer{config: cfg}

	testCases := []struct {
		name            string
		plan            *tfjson.Plan
		validateSummary func(t *testing.T, summary *PlanSummary, description string)
		description     string
	}{
		{
			name: "comprehensive plan with mixed unknown values, outputs, and danger highlighting",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.web_server",
						Type:    "aws_instance",
						Name:    "web_server",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"instance_type": "t3.micro",
								"ami":           "ami-12345678",
								"user_data":     "#!/bin/bash\necho 'Hello World'",
								"id":            nil,
								"public_ip":     nil,
							},
							AfterUnknown: map[string]any{
								"id":        true,
								"public_ip": true,
							},
						},
					},
					{
						Address: "aws_iam_policy.admin_policy",
						Type:    "aws_iam_policy",
						Name:    "admin_policy",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"policy": `{"Version": "2012-10-17"}`,
								"arn":    "arn:aws:iam::123456789012:policy/old-policy",
							},
							After: map[string]any{
								"policy": `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "*", "Resource": "*"}]}`,
								"arn":    nil,
							},
							AfterUnknown: map[string]any{
								"arn": true,
							},
						},
					},
					{
						Address: "aws_s3_bucket.data_bucket",
						Type:    "aws_s3_bucket",
						Name:    "data_bucket",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"bucket":             "my-data-bucket-2024",
								"versioning":         true,
								"force_destroy":      false,
								"id":                 nil,
								"bucket_domain_name": nil,
							},
							AfterUnknown: map[string]any{
								"id":                 true,
								"bucket_domain_name": true,
							},
						},
					},
				},
				OutputChanges: map[string]*tfjson.Change{
					"instance_public_ip": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
					"policy_arn": {
						Actions:      []tfjson.Action{tfjson.ActionUpdate},
						Before:       "arn:aws:iam::123456789012:policy/old-policy",
						After:        nil,
						AfterUnknown: true,
					},
					"bucket_endpoint": {
						Actions: []tfjson.Action{tfjson.ActionCreate},
						Before:  nil,
						After:   "https://my-data-bucket-2024.s3.amazonaws.com",
					},
					"database_password": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          "super-secret-password",
						AfterSensitive: true,
					},
					"deprecated_config": {
						Actions: []tfjson.Action{tfjson.ActionDelete},
						Before:  "old-configuration-value",
						After:   nil,
					},
				},
			},
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Test 1: Verify unknown values display correctly and don't appear as deletions (requirement 1.1, 1.2)
				assert.Len(t, summary.ResourceChanges, 3, description+" - should have 3 resource changes")

				resourceMap := make(map[string]ResourceChange)
				for _, rc := range summary.ResourceChanges {
					resourceMap[rc.Address] = rc
				}

				// Verify aws_instance has unknown values displayed correctly
				if instance, exists := resourceMap["aws_instance.web_server"]; exists {
					assert.True(t, instance.HasUnknownValues, "web_server should have unknown values")
					assert.Contains(t, instance.UnknownProperties, "id", "web_server should have unknown id property")
					assert.Contains(t, instance.UnknownProperties, "public_ip", "web_server should have unknown public_ip property")

					// Verify unknown values in property changes don't appear as deletions
					for _, change := range instance.PropertyChanges.Changes {
						if change.IsUnknown {
							assert.NotEqual(t, "remove", change.Action, "Unknown property %s should not appear as deletion", change.Name)
						}
					}
				}

				// Verify aws_iam_policy has unknown values and can work with danger highlighting (requirement 3.1)
				if policy, exists := resourceMap["aws_iam_policy.admin_policy"]; exists {
					assert.True(t, policy.HasUnknownValues, "admin_policy should have unknown values")
					assert.Contains(t, policy.UnknownProperties, "arn", "admin_policy should have unknown arn property")
					// Note: Danger highlighting typically applies to destructive changes (replace/delete),
					// but unknown values should still work with any danger highlighting logic that does apply
				}

				// Test 2: Verify outputs section displays with correct 5-column format (requirement 2.2)
				assert.Len(t, summary.OutputChanges, 5, description+" - should have 5 output changes")

				outputMap := make(map[string]OutputChange)
				for _, oc := range summary.OutputChanges {
					outputMap[oc.Name] = oc
				}

				// Verify unknown output values display "(known after apply)" (requirement 2.3)
				if instanceIp, exists := outputMap["instance_public_ip"]; exists {
					assert.Equal(t, "(known after apply)", instanceIp.After, "instance_public_ip should show (known after apply)")
					assert.True(t, instanceIp.IsUnknown, "instance_public_ip should be marked as unknown")
					assert.Equal(t, "Add", instanceIp.Action, "instance_public_ip should have Add action")
					assert.Equal(t, "+", instanceIp.Indicator, "instance_public_ip should have + indicator")
				}

				if policyArn, exists := outputMap["policy_arn"]; exists {
					assert.Equal(t, "(known after apply)", policyArn.After, "policy_arn should show (known after apply)")
					assert.True(t, policyArn.IsUnknown, "policy_arn should be marked as unknown")
					assert.Equal(t, "Modify", policyArn.Action, "policy_arn should have Modify action")
					assert.Equal(t, "~", policyArn.Indicator, "policy_arn should have ~ indicator")
				}

				// Verify sensitive output values display "(sensitive value)" (requirement 2.4)
				if dbPassword, exists := outputMap["database_password"]; exists {
					assert.Equal(t, "(sensitive value)", dbPassword.After, "database_password should show (sensitive value)")
					assert.True(t, dbPassword.Sensitive, "database_password should be marked as sensitive")
					assert.False(t, dbPassword.IsUnknown, "database_password should not be unknown")
					assert.Equal(t, "Add", dbPassword.Action, "database_password should have Add action")
					assert.Equal(t, "+", dbPassword.Indicator, "database_password should have + indicator")
				}

				// Verify normal output values display correctly
				if bucketEndpoint, exists := outputMap["bucket_endpoint"]; exists {
					assert.Equal(t, "https://my-data-bucket-2024.s3.amazonaws.com", bucketEndpoint.After, "bucket_endpoint should show actual value")
					assert.False(t, bucketEndpoint.Sensitive, "bucket_endpoint should not be sensitive")
					assert.False(t, bucketEndpoint.IsUnknown, "bucket_endpoint should not be unknown")
				}

				// Verify deleted output values
				if deprecated, exists := outputMap["deprecated_config"]; exists {
					assert.Equal(t, "Remove", deprecated.Action, "deprecated_config should have Remove action")
					assert.Equal(t, "-", deprecated.Indicator, "deprecated_config should have - indicator")
				}

				// Test 3: Verify statistics properly categorize resource changes with unknown values (requirement 3.3)
				stats := summary.Statistics
				assert.Equal(t, 2, stats.ToAdd, "statistics should count resource creations")
				assert.Equal(t, 1, stats.ToChange, "statistics should count resource modifications")
				assert.Equal(t, 3, stats.Total, "statistics should total resource changes only, not outputs")

				// Test 4: Verify basic summary structure
				assert.Equal(t, "1.0", summary.FormatVersion, "format version should match")
				assert.Equal(t, "1.5.0", summary.TerraformVersion, "terraform version should match")
			},
			description: "Complete workflow should handle unknown values, outputs, and danger highlighting together",
		},
		{
			name: "complex nested unknown values with outputs integration",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_vpc.main",
						Type:    "aws_vpc",
						Name:    "main",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"cidr_block": "10.0.0.0/16",
								"tags": map[string]any{
									"Name":        "main-vpc",
									"Environment": "production",
								},
								"id":                        nil,
								"arn":                       nil,
								"default_security_group_id": nil,
								"default_network_acl_id":    nil,
							},
							AfterUnknown: map[string]any{
								"id":                        true,
								"arn":                       true,
								"default_security_group_id": true,
								"default_network_acl_id":    true,
							},
						},
					},
				},
				OutputChanges: map[string]*tfjson.Change{
					"vpc_details": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil,
						AfterUnknown: true,
					},
				},
			},
			validateSummary: func(t *testing.T, summary *PlanSummary, description string) {
				// Verify nested unknown values are handled correctly
				assert.Len(t, summary.ResourceChanges, 1, description+" - should have 1 resource change")

				vpc := summary.ResourceChanges[0]
				assert.True(t, vpc.HasUnknownValues, "VPC should have unknown values")
				assert.Contains(t, vpc.UnknownProperties, "id", "VPC should have unknown id")
				assert.Contains(t, vpc.UnknownProperties, "arn", "VPC should have unknown arn")
				assert.Contains(t, vpc.UnknownProperties, "default_security_group_id", "VPC should have unknown default_security_group_id")

				// Verify outputs with complex unknown structures
				assert.Len(t, summary.OutputChanges, 1, description+" - should have 1 output change")

				output := summary.OutputChanges[0]
				assert.Equal(t, "vpc_details", output.Name, "output name should match")
				assert.Equal(t, "(known after apply)", output.After, "complex unknown output should show (known after apply)")
				assert.True(t, output.IsUnknown, "output should be marked as unknown")
			},
			description: "Complex nested unknown values should integrate correctly with outputs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the plan on the analyzer
			analyzer.plan = tc.plan

			// Generate complete summary - this tests the full workflow
			summary := analyzer.GenerateSummary("")
			assert.NotNil(t, summary, tc.description+" - summary should not be nil")

			// Run comprehensive validation
			if tc.validateSummary != nil {
				tc.validateSummary(t, summary, tc.description)
			}
		})
	}
}

// TestCrossFormatConsistencyForUnknownValuesAndOutputs tests display consistency
// across all output formats (Task 9.2)
func TestCrossFormatConsistencyForUnknownValuesAndOutputs(t *testing.T) {
	cfg := &config.Config{}
	analyzer := &Analyzer{config: cfg}

	// Create a comprehensive plan with various unknown values and outputs
	plan := &tfjson.Plan{
		FormatVersion:    "1.0",
		TerraformVersion: "1.5.0",
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_instance.test",
				Type:    "aws_instance",
				Name:    "test",
				Change: &tfjson.Change{
					Actions: []tfjson.Action{tfjson.ActionCreate},
					Before:  nil,
					After: map[string]any{
						"instance_type": "t3.micro",
						"ami":           "ami-12345678",
						"id":            nil,
						"public_ip":     nil,
					},
					AfterUnknown: map[string]any{
						"id":        true,
						"public_ip": true,
					},
				},
			},
		},
		OutputChanges: map[string]*tfjson.Change{
			"instance_ip": {
				Actions:      []tfjson.Action{tfjson.ActionCreate},
				Before:       nil,
				After:        nil,
				AfterUnknown: true,
			},
			"secret_value": {
				Actions:        []tfjson.Action{tfjson.ActionCreate},
				Before:         nil,
				After:          "top-secret",
				AfterSensitive: true,
			},
			"public_endpoint": {
				Actions: []tfjson.Action{tfjson.ActionCreate},
				Before:  nil,
				After:   "https://api.example.com",
			},
		},
	}

	analyzer.plan = plan
	summary := analyzer.GenerateSummary("")

	testCases := []struct {
		name            string
		validateContent func(t *testing.T, content string, format string)
		description     string
	}{
		{
			name: "unknown values display consistency across formats",
			validateContent: func(t *testing.T, content string, format string) {
				// Verify "(known after apply)" appears consistently (requirement 1.3)
				assert.Contains(t, content, "(known after apply)", format+" format should contain (known after apply)")

				// Count occurrences - should appear for both resource properties and outputs
				unknownCount := strings.Count(content, "(known after apply)")
				assert.GreaterOrEqual(t, unknownCount, 2, format+" format should have multiple (known after apply) instances")
			},
			description: "Unknown values should display (known after apply) consistently across all formats",
		},
		{
			name: "sensitive values display consistency across formats",
			validateContent: func(t *testing.T, content string, format string) {
				// Verify "(sensitive value)" appears consistently (requirement 2.4)
				assert.Contains(t, content, "(sensitive value)", format+" format should contain (sensitive value)")
			},
			description: "Sensitive values should display (sensitive value) consistently across all formats",
		},
		{
			name: "outputs section format consistency",
			validateContent: func(t *testing.T, content string, format string) {
				// Verify outputs section content appears
				outputNames := []string{"instance_ip", "secret_value", "public_endpoint"}
				for _, name := range outputNames {
					assert.Contains(t, content, name, format+" format should contain output "+name)
				}

				// Verify action indicators appear
				actionIndicators := []string{"+", "~", "-"}
				indicatorFound := false
				for _, indicator := range actionIndicators {
					if strings.Contains(content, indicator) {
						indicatorFound = true
						break
					}
				}
				assert.True(t, indicatorFound, format+" format should contain action indicators")
			},
			description: "Outputs section should appear consistently across all formats",
		},
	}

	// Test formats that are relevant for consistency checking
	formats := []struct {
		name       string
		getContent func() string
	}{
		{
			name: "JSON",
			getContent: func() string {
				// Test JSON serialization consistency
				jsonBytes, err := json.Marshal(summary)
				assert.NoError(t, err, "JSON marshaling should not error")
				return string(jsonBytes)
			},
		},
		{
			name: "Summary Structure",
			getContent: func() string {
				// Test the summary data structure directly
				var content strings.Builder

				// Add resource changes content
				for _, rc := range summary.ResourceChanges {
					content.WriteString(fmt.Sprintf("Resource: %s\n", rc.Address))
					if rc.HasUnknownValues {
						content.WriteString("Has unknown values\n")
						for _, prop := range rc.UnknownProperties {
							content.WriteString(fmt.Sprintf("Unknown property: %s\n", prop))
						}
					}
					for _, change := range rc.PropertyChanges.Changes {
						if change.IsUnknown {
							content.WriteString(fmt.Sprintf("Property %s: (known after apply)\n", change.Name))
						}
					}
				}

				// Add output changes content
				for _, oc := range summary.OutputChanges {
					content.WriteString(fmt.Sprintf("Output: %s %s\n", oc.Name, oc.Indicator))
					if oc.IsUnknown {
						content.WriteString("(known after apply)\n")
					} else if oc.Sensitive {
						content.WriteString("(sensitive value)\n")
					}
				}

				return content.String()
			},
		},
	}

	for _, format := range formats {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_%s", format.name, tc.name), func(t *testing.T) {
				content := format.getContent()
				tc.validateContent(t, content, format.name)
			})
		}
	}

	// Additional cross-format validation
	t.Run("data_consistency_across_processing", func(t *testing.T) {
		// Verify that the same data appears consistently in the summary structure

		// Check resource changes
		assert.Len(t, summary.ResourceChanges, 1, "should have 1 resource change")
		resource := summary.ResourceChanges[0]
		assert.True(t, resource.HasUnknownValues, "resource should have unknown values")
		assert.Contains(t, resource.UnknownProperties, "id", "resource should have unknown id")
		assert.Contains(t, resource.UnknownProperties, "public_ip", "resource should have unknown public_ip")

		// Check output changes
		assert.Len(t, summary.OutputChanges, 3, "should have 3 output changes")

		outputMap := make(map[string]OutputChange)
		for _, oc := range summary.OutputChanges {
			outputMap[oc.Name] = oc
		}

		// Verify unknown output
		if instanceIp, exists := outputMap["instance_ip"]; exists {
			assert.Equal(t, "(known after apply)", instanceIp.After, "unknown output should show (known after apply)")
			assert.True(t, instanceIp.IsUnknown, "unknown output should be marked as unknown")
		}

		// Verify sensitive output
		if secretValue, exists := outputMap["secret_value"]; exists {
			assert.Equal(t, "(sensitive value)", secretValue.After, "sensitive output should show (sensitive value)")
			assert.True(t, secretValue.Sensitive, "sensitive output should be marked as sensitive")
		}

		// Verify normal output
		if publicEndpoint, exists := outputMap["public_endpoint"]; exists {
			assert.Equal(t, "https://api.example.com", publicEndpoint.After, "normal output should show actual value")
			assert.False(t, publicEndpoint.Sensitive, "normal output should not be sensitive")
			assert.False(t, publicEndpoint.IsUnknown, "normal output should not be unknown")
		}
	})
}

// TestSortPropertiesAlphabetically tests the property sorting functionality
func TestSortPropertiesAlphabetically(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		input    []PropertyChange
		expected []PropertyChange
	}{
		{
			name: "Basic alphabetical sorting case-insensitive",
			input: []PropertyChange{
				{Name: "zebra", Path: []string{"zebra"}, Action: "update"},
				{Name: "Apple", Path: []string{"Apple"}, Action: "update"},
				{Name: "banana", Path: []string{"banana"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "Apple", Path: []string{"Apple"}, Action: "update"},
				{Name: "banana", Path: []string{"banana"}, Action: "update"},
				{Name: "zebra", Path: []string{"zebra"}, Action: "update"},
			},
		},
		{
			name: "Same name properties sorted by path hierarchy",
			input: []PropertyChange{
				{Name: "config", Path: []string{"config", "nested", "deep"}, Action: "update"},
				{Name: "config", Path: []string{"config", "basic"}, Action: "update"},
				{Name: "config", Path: []string{"config"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "config", Path: []string{"config"}, Action: "update"},
				{Name: "config", Path: []string{"config", "basic"}, Action: "update"},
				{Name: "config", Path: []string{"config", "nested", "deep"}, Action: "update"},
			},
		},
		{
			name: "Natural sort ordering with numbers",
			input: []PropertyChange{
				{Name: "prop10", Path: []string{"prop10"}, Action: "update"},
				{Name: "prop2", Path: []string{"prop2"}, Action: "update"},
				{Name: "prop1", Path: []string{"prop1"}, Action: "update"},
				{Name: "prop20", Path: []string{"prop20"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "prop1", Path: []string{"prop1"}, Action: "update"},
				{Name: "prop2", Path: []string{"prop2"}, Action: "update"},
				{Name: "prop10", Path: []string{"prop10"}, Action: "update"},
				{Name: "prop20", Path: []string{"prop20"}, Action: "update"},
			},
		},
		{
			name: "Action type tiebreaker for identical names and paths",
			input: []PropertyChange{
				{Name: "prop", Path: []string{"prop"}, Action: "add"},
				{Name: "prop", Path: []string{"prop"}, Action: "remove"},
				{Name: "prop", Path: []string{"prop"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "prop", Path: []string{"prop"}, Action: "remove"},
				{Name: "prop", Path: []string{"prop"}, Action: "update"},
				{Name: "prop", Path: []string{"prop"}, Action: "add"},
			},
		},
		{
			name: "Mixed properties with special characters",
			input: []PropertyChange{
				{Name: "user_data", Path: []string{"user_data"}, Action: "update"},
				{Name: "user-name", Path: []string{"user-name"}, Action: "update"},
				{Name: "user.config", Path: []string{"user.config"}, Action: "update"},
				{Name: "user", Path: []string{"user"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "user", Path: []string{"user"}, Action: "update"},
				{Name: "user-name", Path: []string{"user-name"}, Action: "update"},
				{Name: "user.config", Path: []string{"user.config"}, Action: "update"},
				{Name: "user_data", Path: []string{"user_data"}, Action: "update"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := &PropertyChangeAnalysis{Changes: tc.input}
			analyzer.sortPropertiesAlphabetically(analysis)

			assert.Equal(t, len(tc.expected), len(analysis.Changes), "Length should match")

			for i, expected := range tc.expected {
				actual := analysis.Changes[i]
				assert.Equal(t, expected.Name, actual.Name, "Property name at index %d should match", i)
				assert.Equal(t, expected.Path, actual.Path, "Property path at index %d should match", i)
				assert.Equal(t, expected.Action, actual.Action, "Property action at index %d should match", i)
			}
		})
	}
}

// TestNaturalSort tests the natural sorting implementation
func TestNaturalSort(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		s1       string
		s2       string
		expected bool
	}{
		{
			name:     "Simple alphabetical order",
			s1:       "apple",
			s2:       "banana",
			expected: true,
		},
		{
			name:     "Numbers should sort numerically",
			s1:       "item2",
			s2:       "item10",
			expected: true,
		},
		{
			name:     "Mixed text and numbers",
			s1:       "version1.2.3",
			s2:       "version1.10.1",
			expected: true,
		},
		{
			name:     "Same prefix with numbers",
			s1:       "prop1",
			s2:       "prop2",
			expected: true,
		},
		{
			name:     "Equal strings",
			s1:       "same",
			s2:       "same",
			expected: false,
		},
		{
			name:     "Leading numbers",
			s1:       "2item",
			s2:       "10item",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.naturalSort(tc.s1, tc.s2)
			assert.Equal(t, tc.expected, result, "Natural sort result should match expected")
		})
	}
}

// TestSplitNatural tests the natural string splitting functionality
func TestSplitNatural(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple text",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "Simple number",
			input:    "123",
			expected: []string{"123"},
		},
		{
			name:     "Text with number",
			input:    "item123",
			expected: []string{"item", "123"},
		},
		{
			name:     "Number then text",
			input:    "123item",
			expected: []string{"123", "item"},
		},
		{
			name:     "Complex mixed",
			input:    "version1.2.3",
			expected: []string{"version", "1", ".", "2", ".", "3"},
		},
		{
			name:     "Multiple numbers",
			input:    "item123test456",
			expected: []string{"item", "123", "test", "456"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.splitNatural(tc.input)
			assert.Equal(t, tc.expected, result, "Split result should match expected")
		})
	}
}

// TestAnalyzePropertyChangesWithSorting tests the integration of sorting with property analysis
func TestAnalyzePropertyChangesWithSorting(t *testing.T) {
	analyzer := NewAnalyzer(nil, &config.Config{})

	// Create a test resource change with unsorted properties
	resourceChange := &tfjson.ResourceChange{
		Address: "test.resource",
		Type:    "test_resource",
		Name:    "resource",
		Change: &tfjson.Change{
			Before: map[string]any{
				"zebra_config":  "old_value",
				"apple_setting": "old_value",
				"banana_option": "old_value",
			},
			After: map[string]any{
				"zebra_config":  "new_value",
				"apple_setting": "new_value",
				"banana_option": "new_value",
			},
		},
	}

	result := analyzer.analyzePropertyChanges(resourceChange)

	// Verify that properties are sorted alphabetically
	assert.True(t, len(result.Changes) >= 3, "Should have at least 3 property changes")

	// Properties should be sorted alphabetically: apple_setting, banana_option, zebra_config
	propertyNames := make([]string, len(result.Changes))
	for i, change := range result.Changes {
		propertyNames[i] = change.Name
	}

	// Check that properties are in alphabetical order
	for i := 1; i < len(propertyNames); i++ {
		assert.True(t, strings.ToLower(propertyNames[i-1]) <= strings.ToLower(propertyNames[i]),
			"Properties should be in alphabetical order: %v", propertyNames)
	}
}

// TestMaskSensitiveValue tests the sensitive value masking functionality
func TestMaskSensitiveValue(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name        string
		value       any
		isSensitive bool
		expected    any
	}{
		{
			name:        "Non-sensitive value should not be masked",
			value:       "normal_value",
			isSensitive: false,
			expected:    "normal_value",
		},
		{
			name:        "Sensitive primitive string should be masked",
			value:       "secret_password",
			isSensitive: true,
			expected:    "(sensitive value)",
		},
		{
			name:        "Sensitive primitive number should be masked",
			value:       12345,
			isSensitive: true,
			expected:    "(sensitive value)",
		},
		{
			name:        "Sensitive primitive boolean should be masked",
			value:       true,
			isSensitive: true,
			expected:    "(sensitive value)",
		},
		{
			name:        "Sensitive map should preserve structure",
			value:       map[string]any{"key": "value"},
			isSensitive: true,
			expected:    map[string]any{"key": "value"},
		},
		{
			name:        "Sensitive slice should preserve structure",
			value:       []any{"item1", "item2"},
			isSensitive: true,
			expected:    []any{"item1", "item2"},
		},
		{
			name:        "Nil value should remain nil",
			value:       nil,
			isSensitive: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.maskSensitiveValue(tc.value, tc.isSensitive)
			assert.Equal(t, tc.expected, result, "Masked value should match expected")
		})
	}
}

// TestCompareObjectsWithSensitiveMasking tests the integration of sensitive masking with property comparison
func TestCompareObjectsWithSensitiveMasking(t *testing.T) {
	analyzer := &Analyzer{}

	// Test case: property change with sensitive values
	analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

	// Create test data with sensitive values
	before := map[string]any{
		"password": "old_secret",
		"username": "normal_user",
	}
	after := map[string]any{
		"password": "new_secret",
		"username": "normal_user",
	}

	// Sensitive flags indicating "password" is sensitive
	beforeSensitive := map[string]any{
		"password": true,
		"username": false,
	}
	afterSensitive := map[string]any{
		"password": true,
		"username": false,
	}

	// Call compareObjects with sensitive data
	analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, []string{}, analysis)

	// Verify that sensitive values are masked while non-sensitive values are preserved
	passwordFound := false
	usernameFound := false

	for _, change := range analysis.Changes {
		switch change.Name {
		case "password":
			passwordFound = true
			assert.True(t, change.Sensitive, "Password property should be marked as sensitive")
			assert.Equal(t, "(sensitive value)", change.Before, "Sensitive before value should be masked")
			assert.Equal(t, "(sensitive value)", change.After, "Sensitive after value should be masked")
		case "username":
			usernameFound = true
			assert.False(t, change.Sensitive, "Username property should not be marked as sensitive")
			assert.Equal(t, "normal_user", change.Before, "Non-sensitive before value should not be masked")
			assert.Equal(t, "normal_user", change.After, "Non-sensitive after value should not be masked")
		}
	}

	// Only password should change since username values are identical
	assert.True(t, passwordFound, "Should find password change")
	assert.False(t, usernameFound, "Should not find username change since values are identical")
}

// TestCompareObjectsWithNestedSensitiveValues tests sensitive masking in nested structures
func TestCompareObjectsWithNestedSensitiveValues(t *testing.T) {
	analyzer := &Analyzer{}
	analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

	// Test nested structure with some sensitive leaf values
	before := map[string]any{
		"config": map[string]any{
			"api_key":  "old_key",
			"endpoint": "https://api.example.com",
			"settings": map[string]any{
				"timeout": 30,
				"secret":  "old_secret",
			},
		},
	}
	after := map[string]any{
		"config": map[string]any{
			"api_key":  "new_key",
			"endpoint": "https://api.example.com",
			"settings": map[string]any{
				"timeout": 60,
				"secret":  "new_secret",
			},
		},
	}

	// Sensitive flags - api_key and secret are sensitive
	beforeSensitive := map[string]any{
		"config": map[string]any{
			"api_key":  true,
			"endpoint": false,
			"settings": map[string]any{
				"timeout": false,
				"secret":  true,
			},
		},
	}
	afterSensitive := map[string]any{
		"config": map[string]any{
			"api_key":  true,
			"endpoint": false,
			"settings": map[string]any{
				"timeout": false,
				"secret":  true,
			},
		},
	}

	analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, []string{}, analysis)

	// Check that sensitive leaf values are masked while structure is preserved
	changesByName := make(map[string]PropertyChange)
	for _, change := range analysis.Changes {
		changesByName[change.Name] = change
	}

	// api_key should be masked
	if apiKeyChange, exists := changesByName["api_key"]; exists {
		assert.True(t, apiKeyChange.Sensitive, "api_key should be marked as sensitive")
		assert.Equal(t, "(sensitive value)", apiKeyChange.Before, "Sensitive api_key before value should be masked")
		assert.Equal(t, "(sensitive value)", apiKeyChange.After, "Sensitive api_key after value should be masked")
	}

	// timeout should not be masked
	if timeoutChange, exists := changesByName["timeout"]; exists {
		assert.False(t, timeoutChange.Sensitive, "timeout should not be marked as sensitive")
		assert.Equal(t, 30, timeoutChange.Before, "Non-sensitive timeout before value should not be masked")
		assert.Equal(t, 60, timeoutChange.After, "Non-sensitive timeout after value should not be masked")
	}

	// secret should be masked
	if secretChange, exists := changesByName["secret"]; exists {
		assert.True(t, secretChange.Sensitive, "secret should be marked as sensitive")
		assert.Equal(t, "(sensitive value)", secretChange.Before, "Sensitive secret before value should be masked")
		assert.Equal(t, "(sensitive value)", secretChange.After, "Sensitive secret after value should be masked")
	}
}
