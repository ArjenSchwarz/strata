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
			got := analyzer.IsSensitiveResource(tc.resourceType)
			if got != tc.expected {
				t.Errorf("IsSensitiveResource() = %v, want %v", got, tc.expected)
			}
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
			got := analyzer.IsSensitiveProperty(tc.resourceType, tc.property)
			if got != tc.expected {
				t.Errorf("IsSensitiveProperty() = %v, want %v", got, tc.expected)
			}
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
			got := analyzer.analyzeReplacementNecessity(tc.change)
			if got != tc.expected {
				t.Errorf("analyzeReplacementNecessity() = %v, want %v", got, tc.expected)
			}
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
			got := analyzer.extractPhysicalID(tc.change)
			if got != tc.expected {
				t.Errorf("extractPhysicalID() = %q, want %q", got, tc.expected)
			}
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
			got := analyzer.extractModulePath(tc.address)
			if got != tc.expected {
				t.Errorf("extractModulePath() = %q, want %q", got, tc.expected)
			}
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
			got := analyzer.extractProvider(tc.resourceType)
			if got != tc.expected {
				t.Errorf("extractProvider() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestExtractProviderCaching(t *testing.T) {
	analyzer := &Analyzer{}

	// Test that caching works by calling the same resource type multiple times
	resourceType := "aws_s3_bucket"

	// First call should compute and cache the result
	result1 := analyzer.extractProvider(resourceType)
	if result1 != "aws" {
		t.Errorf("extractProvider() first call = %q, want %q", result1, "aws")
	}

	// Second call should return cached result
	result2 := analyzer.extractProvider(resourceType)
	if result2 != "aws" {
		t.Errorf("extractProvider() second call = %q, want %q", result2, "aws")
	}

	// Verify cache contains the entry
	cached, ok := analyzer.providerCache.Load(resourceType)
	assert.True(t, ok, "Cache should contain the entry")
	if cached.(string) != "aws" {
		t.Errorf("cached provider = %q, want %q", cached.(string), "aws")
	}
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
			got := analyzer.extractProvider(tc.resourceType)
			if got != tc.expected {
				t.Errorf("extractProvider() = %q, want %q", got, tc.expected)
			}
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
			got := analyzer.extractReplacementHints(tc.change)
			if len(got) != len(tc.expected) {
				t.Errorf("extractReplacementHints() length = %d, want %d", len(got), len(tc.expected))
			}
			for i, expected := range tc.expected {
				if i >= len(got) || got[i] != expected {
					t.Errorf("extractReplacementHints()[%d] = %q, want %q", i, got[i], expected)
				}
			}
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
			got := analyzer.formatReplacePath(tc.path)
			if got != tc.expected {
				t.Errorf("formatReplacePath() = %q, want %q", got, tc.expected)
			}
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
				if len(tc.expected) != len(result) {
					t.Errorf("getTopChangedProperties() length = %d, want %d", len(result), len(tc.expected))
				}
				for i, expected := range tc.expected {
					if i >= len(result) || result[i] != expected {
						t.Errorf("getTopChangedProperties()[%d] = %q, want %q", i, result[i], expected)
					}
				}
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
			if dangerous != tc.expectedDanger {
				t.Errorf("evaluateResourceDanger() dangerous = %v, want %v", dangerous, tc.expectedDanger)
			}
			if reason != tc.expectedReason {
				t.Errorf("evaluateResourceDanger() reason = %q, want %q", reason, tc.expectedReason)
			}
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
			got := analyzer.getSensitiveResourceReason(tc.resourceType)
			if got != tc.expected {
				t.Errorf("getSensitiveResourceReason() = %q, want %q", got, tc.expected)
			}
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
			got := analyzer.getSensitivePropertyReason(tc.properties)
			if got != tc.expected {
				t.Errorf("getSensitivePropertyReason() = %q, want %q", got, tc.expected)
			}
		})
	}
}
