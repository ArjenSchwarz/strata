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
