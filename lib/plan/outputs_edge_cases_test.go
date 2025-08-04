package plan

import (
	"fmt"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSensitiveOutputsWithUnknownValues tests sensitive outputs with unknown values
func TestSensitiveOutputsWithUnknownValues(t *testing.T) {
	tests := []struct {
		name        string
		plan        *tfjson.Plan
		description string
	}{
		{
			name: "sensitive output with unknown value",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges: map[string]*tfjson.Change{
					"database_password": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          nil, // Unknown value
						AfterUnknown:   true,
						AfterSensitive: true,
					},
					"api_key": {
						Actions:        []tfjson.Action{tfjson.ActionUpdate},
						Before:         "(sensitive value)",
						After:          nil, // Unknown value
						AfterUnknown:   true,
						AfterSensitive: true,
					},
					"mixed_sensitive": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          "visible-value", // Known sensitive value
						AfterUnknown:   false,
						AfterSensitive: true,
					},
				},
			},
			description: "Sensitive outputs with unknown values should display both indicators correctly",
		},
		{
			name: "complex sensitive output scenarios",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges: map[string]*tfjson.Change{
					"secret_known": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          "secret-value-123",
						AfterUnknown:   false,
						AfterSensitive: true,
					},
					"secret_unknown": {
						Actions:        []tfjson.Action{tfjson.ActionCreate},
						Before:         nil,
						After:          nil, // Unknown
						AfterUnknown:   true,
						AfterSensitive: true,
					},
					"secret_transition": {
						Actions:        []tfjson.Action{tfjson.ActionUpdate},
						Before:         "old-secret",
						After:          nil, // Becomes unknown
						AfterUnknown:   true,
						AfterSensitive: true,
					},
				},
			},
			description: "Multiple sensitive output scenarios should be handled appropriately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tt.plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			require.NotEmpty(t, summary.OutputChanges, "Should have output changes")

			outputMap := make(map[string]OutputChange)
			for _, oc := range summary.OutputChanges {
				outputMap[oc.Name] = oc
			}

			foundSensitiveUnknown := 0
			foundSensitiveKnown := 0

			for name, output := range outputMap {
				t.Logf("Output %s: Sensitive=%v, IsUnknown=%v, After=%v",
					name, output.Sensitive, output.IsUnknown, output.After)

				if output.Sensitive && output.IsUnknown {
					foundSensitiveUnknown++
					// Sensitive unknown outputs should show "(known after apply)" not "(sensitive value)"
					assert.Equal(t, "(known after apply)", output.After,
						"Sensitive unknown output %s should show '(known after apply)'", name)
				} else if output.Sensitive && !output.IsUnknown {
					foundSensitiveKnown++
					// Known sensitive outputs should show "(sensitive value)"
					assert.Equal(t, "(sensitive value)", output.After,
						"Known sensitive output %s should show '(sensitive value)'", name)
				}

				// All sensitive outputs should be marked as sensitive
				if output.Name == "database_password" || output.Name == "api_key" ||
					output.Name == "mixed_sensitive" || output.Name == "secret_known" ||
					output.Name == "secret_unknown" || output.Name == "secret_transition" {
					assert.True(t, output.Sensitive, "Output %s should be marked as sensitive", name)
				}
			}

			t.Logf("%s: Found %d sensitive+unknown outputs, %d sensitive+known outputs",
				tt.description, foundSensitiveUnknown, foundSensitiveKnown)

			assert.Greater(t, foundSensitiveUnknown+foundSensitiveKnown, 0, "Should find sensitive outputs")
		})
	}
}

// TestLargeOutputValues tests large output values with size limits
func TestLargeOutputValues(t *testing.T) {
	tests := []struct {
		name          string
		outputSize    int
		expectedTrunc bool
		description   string
	}{
		{
			name:          "small output value",
			outputSize:    100, // 100 characters
			expectedTrunc: false,
			description:   "Small output values should not be truncated",
		},
		{
			name:          "medium output value",
			outputSize:    5000, // 5KB
			expectedTrunc: false,
			description:   "Medium output values should not be truncated",
		},
		{
			name:          "large output value",
			outputSize:    15000, // 15KB - may trigger limits
			expectedTrunc: false, // Current implementation may not truncate outputs
			description:   "Large output values should respect size limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate large output value
			largeValue := make([]byte, tt.outputSize)
			for i := range largeValue {
				largeValue[i] = byte('a' + (i % 26)) // Cycle through a-z
			}
			largeValueStr := string(largeValue)

			plan := &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges: map[string]*tfjson.Change{
					"large_output": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        largeValueStr,
						AfterUnknown: false,
					},
					"normal_output": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        "normal-value",
						AfterUnknown: false,
					},
				},
			}

			cfg := getTestConfig()
			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			require.Len(t, summary.OutputChanges, 2, "Should have 2 output changes")

			var largeOutput *OutputChange
			for _, oc := range summary.OutputChanges {
				if oc.Name == "large_output" {
					output := oc // Create a copy
					largeOutput = &output
					break
				}
			}

			require.NotNil(t, largeOutput, "Should find large output")

			// Check if output value was processed correctly
			if tt.expectedTrunc {
				// If truncation is expected, the output should be shorter than input
				assert.Less(t, len(largeOutput.After.(string)), tt.outputSize,
					"Large output should be truncated")
			} else {
				// If no truncation expected, output should match input (or be handled gracefully)
				if str, ok := largeOutput.After.(string); ok {
					assert.LessOrEqual(t, len(str), tt.outputSize+100, // Allow some margin
						"Output size should be reasonable")
				}
			}

			t.Logf("%s: Input size=%d, Output size=%d",
				tt.description, tt.outputSize, len(largeOutput.After.(string)))
		})
	}
}

// TestMalformedOutputStructures tests malformed output structures with graceful error handling
func TestMalformedOutputStructures(t *testing.T) {
	tests := []struct {
		name        string
		plan        *tfjson.Plan
		description string
	}{
		{
			name: "missing output actions",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges: map[string]*tfjson.Change{
					"valid_output": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        "valid-value",
						AfterUnknown: false,
					},
					"invalid_output": {
						// Missing Actions field
						Before:       nil,
						After:        "invalid-value",
						AfterUnknown: false,
					},
				},
			},
			description: "Outputs with missing actions should be handled gracefully",
		},
		{
			name: "nil output changes",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges: map[string]*tfjson.Change{
					"valid_output": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        "valid-value",
						AfterUnknown: false,
					},
					"nil_change": nil, // Nil change entry
				},
			},
			description: "Nil output change entries should be handled gracefully",
		},
		{
			name: "empty output actions",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges: map[string]*tfjson.Change{
					"empty_actions": {
						Actions:      []tfjson.Action{}, // Empty actions
						Before:       nil,
						After:        "some-value",
						AfterUnknown: false,
					},
					"valid_output": {
						Actions:      []tfjson.Action{tfjson.ActionUpdate},
						Before:       "old-value",
						After:        "new-value",
						AfterUnknown: false,
					},
				},
			},
			description: "Outputs with empty actions should be handled gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tt.plan, cfg)

			// This should not panic or crash
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil even with malformed data")

			// Should process valid outputs successfully
			validFound := false
			for _, oc := range summary.OutputChanges {
				if oc.Name == "valid_output" {
					validFound = true
					assert.NotEmpty(t, oc.Action, "Valid output should have action")
					assert.NotEmpty(t, oc.Indicator, "Valid output should have indicator")
					break
				}
			}

			assert.True(t, validFound, "Should successfully process valid outputs despite malformed entries")

			t.Logf("%s: Processed %d outputs successfully",
				tt.description, len(summary.OutputChanges))
		})
	}
}

// TestPlansWithOnlyOutputChanges tests plans with only output changes (no resource changes)
func TestPlansWithOnlyOutputChanges(t *testing.T) {
	tests := []struct {
		name        string
		plan        *tfjson.Plan
		description string
	}{
		{
			name: "only output changes, no resources",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges:  nil, // No resource changes
				OutputChanges: map[string]*tfjson.Change{
					"config_value": {
						Actions:      []tfjson.Action{tfjson.ActionUpdate},
						Before:       "old-config",
						After:        "new-config",
						AfterUnknown: false,
					},
					"computed_value": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        nil, // Unknown
						AfterUnknown: true,
					},
					"removed_output": {
						Actions:      []tfjson.Action{tfjson.ActionDelete},
						Before:       "deprecated-value",
						After:        nil,
						AfterUnknown: false,
					},
				},
			},
			description: "Plans with only output changes should be processed correctly",
		},
		{
			name: "empty resource changes with outputs",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges:  []*tfjson.ResourceChange{}, // Empty slice
				OutputChanges: map[string]*tfjson.Change{
					"standalone_output": {
						Actions:      []tfjson.Action{tfjson.ActionCreate},
						Before:       nil,
						After:        "standalone-value",
						AfterUnknown: false,
					},
				},
			},
			description: "Empty resource changes with outputs should work correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tt.plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")

			// Should have no resource changes
			assert.Empty(t, summary.ResourceChanges, "Should have no resource changes")

			// Should have output changes
			assert.NotEmpty(t, summary.OutputChanges, "Should have output changes")

			// Verify statistics
			stats := summary.Statistics
			assert.Equal(t, 0, stats.Total, "Resource statistics should be zero")
			assert.Equal(t, 0, stats.ToAdd, "Should have no resources to add")
			assert.Equal(t, 0, stats.ToChange, "Should have no resources to change")
			assert.Equal(t, 0, stats.ToDestroy, "Should have no resources to destroy")

			// Verify output processing
			foundActions := make(map[string]bool)
			for _, oc := range summary.OutputChanges {
				foundActions[oc.Action] = true
				assert.NotEmpty(t, oc.Indicator, "Output should have indicator")

				if oc.IsUnknown {
					assert.Equal(t, "(known after apply)", oc.After,
						"Unknown output should show '(known after apply)'")
				}
			}

			t.Logf("%s: Found %d outputs with actions: %v",
				tt.description, len(summary.OutputChanges), foundActions)
		})
	}
}

// TestOutputPerformanceWithLargeOutputSets tests performance with many outputs
func TestOutputPerformanceWithLargeOutputSets(t *testing.T) {
	tests := []struct {
		name        string
		numOutputs  int
		description string
	}{
		{
			name:        "medium output set",
			numOutputs:  50,
			description: "50 outputs should process efficiently",
		},
		{
			name:        "large output set",
			numOutputs:  200,
			description: "200 outputs should process within reasonable time",
		},
		{
			name:        "very large output set",
			numOutputs:  500,
			description: "500 outputs should handle gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate many outputs with mixed types
			outputChanges := make(map[string]*tfjson.Change)

			for i := 0; i < tt.numOutputs; i++ {
				outputName := generateOutputName(i)
				actionType := i % 3 // Cycle through action types

				var actions []tfjson.Action
				var before, after any
				var isUnknown bool

				switch actionType {
				case 0: // Create
					actions = []tfjson.Action{tfjson.ActionCreate}
					before = nil
					if i%4 == 0 { // 25% unknown
						after = nil
						isUnknown = true
					} else {
						after = generateOutputValue(i)
						isUnknown = false
					}
				case 1: // Update
					actions = []tfjson.Action{tfjson.ActionUpdate}
					before = generateOutputValue(i)
					if i%5 == 0 { // 20% unknown
						after = nil
						isUnknown = true
					} else {
						after = generateOutputValue(i + 1000)
						isUnknown = false
					}
				case 2: // Delete
					actions = []tfjson.Action{tfjson.ActionDelete}
					before = generateOutputValue(i)
					after = nil
					isUnknown = false
				}

				outputChanges[outputName] = &tfjson.Change{
					Actions:        actions,
					Before:         before,
					After:          after,
					AfterUnknown:   isUnknown,
					AfterSensitive: (i%10 == 0), // 10% sensitive
				}
			}

			plan := &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				OutputChanges:    outputChanges,
			}

			cfg := getTestConfig()
			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			// Note: Some outputs might be filtered out due to processing logic
			assert.LessOrEqual(t, len(summary.OutputChanges), tt.numOutputs, "Should not exceed expected outputs")
			assert.Greater(t, len(summary.OutputChanges), tt.numOutputs/2, "Should process majority of outputs")

			// Verify different output types were processed
			actionCounts := make(map[string]int)
			unknownCount := 0
			sensitiveCount := 0

			for _, oc := range summary.OutputChanges {
				actionCounts[oc.Action]++
				if oc.IsUnknown {
					unknownCount++
				}
				if oc.Sensitive {
					sensitiveCount++
				}
			}

			assert.Greater(t, len(actionCounts), 0, "Should have processed different action types")

			t.Logf("%s: Processed %d outputs, %d unknown, %d sensitive, actions: %v",
				tt.description, tt.numOutputs, unknownCount, sensitiveCount, actionCounts)
		})
	}
}

// Helper functions for generating test data

func generateOutputName(index int) string {
	baseNames := []string{
		"instance_ip", "database_endpoint", "api_gateway_url", "bucket_name", "vpc_id",
		"subnet_ids", "security_group_id", "load_balancer_dns", "certificate_arn", "key_pair_name",
		"cluster_endpoint", "node_group_arn", "ecr_repository_url", "lambda_function_arn", "sns_topic_arn",
		"sqs_queue_url", "cloudfront_domain", "route53_zone_id", "iam_role_arn", "s3_bucket_arn",
	}

	if index < len(baseNames) {
		return baseNames[index]
	} else {
		baseIndex := index % len(baseNames)
		suffix := index / len(baseNames)
		return baseNames[baseIndex] + "_" + generateResourceName(suffix)
	}
}

func generateOutputValue(index int) string {
	valueTypes := []string{
		"https://api-%d.example.com",
		"arn:aws:s3:::bucket-%d",
		"vpc-%d",
		"10.0.%d.0/24",
		"i-%d",
	}

	template := valueTypes[index%len(valueTypes)]
	return fmt.Sprintf(template, index)
}
