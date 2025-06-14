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
			name: "Replace with ReplacePaths and computed values should be conditional",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions:      tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
					ReplacePaths: []interface{}{[]interface{}{"computed_field"}},
					After: map[string]interface{}{
						"computed_field": nil, // null value indicates computed
					},
				},
			},
			expected: ReplacementConditional,
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

func TestIsConditionalReplacementPath(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		change   *tfjson.ResourceChange
		path     interface{}
		expected bool
	}{
		{
			name: "Path with null value should be conditional",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: map[string]interface{}{
						"field": nil,
					},
				},
			},
			path:     []interface{}{"field"},
			expected: true,
		},
		{
			name: "Path with definite value should not be conditional",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: map[string]interface{}{
						"field": "value",
					},
				},
			},
			path:     []interface{}{"field"},
			expected: false,
		},
		{
			name: "Path with nested null value should be conditional",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: map[string]interface{}{
						"nested": map[string]interface{}{
							"field": nil,
						},
					},
				},
			},
			path:     []interface{}{"nested", "field"},
			expected: true,
		},
		{
			name: "No after state should not be conditional",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					After: nil,
				},
			},
			path:     []interface{}{"field"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.isConditionalReplacementPath(tc.change, tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCalculateStatistics(t *testing.T) {
	analyzer := &Analyzer{}

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
				Conditionals: 0,
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
				Conditionals: 0,
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
				Conditionals: 0,
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
				Conditionals: 0,
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
				Conditionals: 0,
				Total:        1,
			},
		},
		{
			name: "Replace with conditional replacement should increment Conditionals and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementConditional},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				Conditionals: 1,
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
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementConditional},
			},
			expected: ChangeStatistics{
				ToAdd:        1,
				ToChange:     1,
				ToDestroy:    1,
				Replacements: 1,
				Conditionals: 1,
				Total:        5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.calculateStatistics(tc.changes)
			assert.Equal(t, tc.expected, result)
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
