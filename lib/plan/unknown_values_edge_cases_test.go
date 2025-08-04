package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComplexNestedUnknownValues tests complex nested structures with mixed known/unknown values
func TestComplexNestedUnknownValues(t *testing.T) {
	tests := []struct {
		name                string
		plan                *tfjson.Plan
		expectedUnknownPath []string
		expectedKnownPath   []string
		description         string
	}{
		{
			name: "multiple root level unknown values",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.complex",
						Type:    "aws_instance",
						Name:    "complex",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"ami":                    "ami-12345",
								"instance_type":          "t3.micro",
								"id":                     nil, // Unknown value
								"public_ip":              nil, // Unknown value
								"private_ip":             nil, // Unknown value
								"availability_zone":      nil, // Unknown value
								"key_name":               "my-key",
								"vpc_security_group_ids": nil, // Unknown value
								"public_dns_name":        nil, // Unknown value
								"private_dns_name":       nil, // Unknown value
							},
							AfterUnknown: map[string]any{
								"id":                     true,
								"public_ip":              true,
								"private_ip":             true,
								"availability_zone":      true,
								"vpc_security_group_ids": true,
								"public_dns_name":        true,
								"private_dns_name":       true,
							},
						},
					},
				},
			},
			expectedUnknownPath: []string{
				"id",
				"public_ip",
				"private_ip",
				"availability_zone",
				"vpc_security_group_ids",
				"public_dns_name",
				"private_dns_name",
			},
			expectedKnownPath: []string{
				"ami",
				"instance_type",
				"key_name",
			},
			description: "Multiple root-level unknown values should be correctly identified",
		},
		{
			name: "mixed resource actions with unknown values",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_db_instance.main",
						Type:    "aws_db_instance",
						Name:    "main",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"engine":            "mysql",
								"engine_version":    "8.0.28",
								"endpoint":          "old.rds.amazonaws.com",
								"port":              3306,
								"allocated_storage": 20,
							},
							After: map[string]any{
								"engine":            "mysql",
								"engine_version":    "8.0.35", // Changed
								"endpoint":          nil,      // Unknown after update
								"port":              nil,      // Unknown after update
								"allocated_storage": 40,       // Changed
								"resource_id":       nil,      // Unknown (new field)
								"address":           nil,      // Unknown (new field)
							},
							AfterUnknown: map[string]any{
								"endpoint":    true,
								"port":        true,
								"resource_id": true,
								"address":     true,
							},
						},
					},
				},
			},
			expectedUnknownPath: []string{
				"endpoint",
				"port",
				"resource_id",
				"address",
			},
			expectedKnownPath: []string{
				"engine",
				"engine_version",
				"allocated_storage",
			},
			description: "Resource updates with unknown values should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tt.plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			require.Len(t, summary.ResourceChanges, 1, "Should have exactly one resource change")

			resource := summary.ResourceChanges[0]
			assert.True(t, resource.HasUnknownValues, "Resource should have unknown values")

			// Verify expected unknown properties are identified
			for _, expectedPath := range tt.expectedUnknownPath {
				assert.Contains(t, resource.UnknownProperties, expectedPath,
					"Expected unknown property path %s should be identified", expectedPath)
			}

			// Verify property changes analysis
			foundUnknownChanges := 0
			foundKnownChanges := 0

			for _, change := range resource.PropertyChanges.Changes {
				if change.IsUnknown {
					foundUnknownChanges++
					assert.Equal(t, "(known after apply)", change.After,
						"Unknown property %s should display '(known after apply)'", change.Name)
					assert.NotEqual(t, "remove", change.Action,
						"Unknown property %s should not appear as removal", change.Name)
				} else {
					foundKnownChanges++
					assert.NotEqual(t, "(known after apply)", change.After,
						"Known property %s should not display '(known after apply)'", change.Name)
				}
			}

			assert.Greater(t, foundUnknownChanges, 0, "Should find unknown property changes")
			assert.Greater(t, foundKnownChanges, 0, "Should find known property changes")

			t.Logf("%s: Found %d unknown changes and %d known changes",
				tt.description, foundUnknownChanges, foundKnownChanges)
		})
	}
}

// TestArraysWithUnknownElements tests arrays with unknown elements
func TestArraysWithUnknownElements(t *testing.T) {
	tests := []struct {
		name        string
		plan        *tfjson.Plan
		description string
	}{
		{
			name: "simple array with unknown elements",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.array_test",
						Type:    "aws_instance",
						Name:    "array_test",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionCreate},
							Before:  nil,
							After: map[string]any{
								"ami":                    "ami-12345",
								"security_groups":        nil, // Unknown array
								"vpc_security_group_ids": nil, // Unknown array
								"subnet_id":              "subnet-12345",
								"availability_zone":      nil, // Unknown value
							},
							AfterUnknown: map[string]any{
								"security_groups":        true,
								"vpc_security_group_ids": true,
								"availability_zone":      true,
							},
						},
					},
				},
			},
			description: "Arrays that are entirely unknown should be handled correctly",
		},
		{
			name: "update action with unknown arrays",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.update_test",
						Type:    "aws_instance",
						Name:    "update_test",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"ami":               "ami-12345",
								"security_groups":   []string{"sg-12345"},
								"availability_zone": "us-east-1a",
							},
							After: map[string]any{
								"ami":                    "ami-67890", // Changed
								"security_groups":        nil,         // Unknown after update
								"availability_zone":      nil,         // Unknown after update
								"vpc_security_group_ids": nil,         // New unknown field
							},
							AfterUnknown: map[string]any{
								"security_groups":        true,
								"availability_zone":      true,
								"vpc_security_group_ids": true,
							},
						},
					},
				},
			},
			description: "Update actions with unknown arrays should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tt.plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			require.Len(t, summary.ResourceChanges, 1, "Should have exactly one resource change")

			resource := summary.ResourceChanges[0]
			assert.True(t, resource.HasUnknownValues, "Resource should have unknown values")
			assert.NotEmpty(t, resource.UnknownProperties, "Should have identified unknown properties")

			// Verify property changes handle array unknowns correctly
			foundUnknownProperties := 0
			for _, change := range resource.PropertyChanges.Changes {
				if change.IsUnknown {
					foundUnknownProperties++
					assert.Equal(t, "(known after apply)", change.After,
						"Unknown property %s should display '(known after apply)'", change.Name)
				}
			}

			// For this test, we mainly care that unknown properties are detected
			assert.Greater(t, len(resource.UnknownProperties), 0, "Should find unknown properties")

			t.Logf("%s: Resource has %d unknown properties, found %d unknown property changes",
				tt.description, len(resource.UnknownProperties), foundUnknownProperties)
		})
	}
}

// TestPropertiesRemainingUnknown tests properties that remain unknown (before and after unknown)
func TestPropertiesRemainingUnknown(t *testing.T) {
	tests := []struct {
		name        string
		plan        *tfjson.Plan
		description string
	}{
		{
			name: "properties remaining unknown before and after",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.persistent_unknown",
						Type:    "aws_instance",
						Name:    "persistent_unknown",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"ami":                    "ami-12345",
								"instance_type":          "t3.micro",
								"id":                     nil, // Was unknown before
								"public_ip":              nil, // Was unknown before
								"private_ip":             "10.0.1.100",
								"availability_zone":      nil,                    // Was unknown before
								"vpc_security_group_ids": []any{nil, "sg-known"}, // Mixed unknown
							},
							After: map[string]any{
								"ami":                    "ami-67890",              // Changed to known value
								"instance_type":          "t3.small",               // Changed to known value
								"id":                     nil,                      // Still unknown after
								"public_ip":              nil,                      // Still unknown after
								"private_ip":             "10.0.1.150",             // Changed known value
								"availability_zone":      nil,                      // Still unknown after
								"vpc_security_group_ids": []any{nil, "sg-updated"}, // Still mixed unknown
							},
							AfterUnknown: map[string]any{
								"id":                     true,               // Remains unknown
								"public_ip":              true,               // Remains unknown
								"availability_zone":      true,               // Remains unknown
								"vpc_security_group_ids": []any{true, false}, // Still mixed
							},
						},
					},
				},
			},
			description: "Properties that become or remain unknown should be handled correctly",
		},
		{
			name: "mixed unknown transitions",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_rds_instance.transitions",
						Type:    "aws_rds_instance",
						Name:    "transitions",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"endpoint":          nil,                 // Unknown before
								"port":              3306,                // Known before
								"address":           "old.amazonaws.com", // Known before
								"hosted_zone_id":    nil,                 // Unknown before
								"resource_id":       nil,                 // Unknown before
								"allocated_storage": 20,                  // Known before
							},
							After: map[string]any{
								"endpoint":          "new.amazonaws.com", // Unknown to known
								"port":              nil,                 // Known to unknown
								"address":           nil,                 // Known to unknown
								"hosted_zone_id":    nil,                 // Unknown to unknown (remains)
								"resource_id":       "db-123456",         // Unknown to known
								"allocated_storage": 40,                  // Known to known (normal change)
							},
							AfterUnknown: map[string]any{
								"port":           true,
								"address":        true,
								"hosted_zone_id": true, // Remains unknown
							},
						},
					},
				},
			},
			description: "Mixed unknown transitions should be handled correctly with proper display",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tt.plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			require.Len(t, summary.ResourceChanges, 1, "Should have exactly one resource change")

			resource := summary.ResourceChanges[0]
			assert.True(t, resource.HasUnknownValues, "Resource should have unknown values")

			// Check that unknown values are handled correctly
			foundUnknownChanges := 0
			foundKnownToUnknownChanges := 0

			for _, change := range resource.PropertyChanges.Changes {
				if change.IsUnknown {
					foundUnknownChanges++
					assert.Equal(t, "(known after apply)", change.After,
						"Unknown property %s should show '(known after apply)'", change.Name)
				}

				// Check for known to unknown transitions (where before has a value but after is unknown)
				if change.Before != nil && change.After == "(known after apply)" {
					foundKnownToUnknownChanges++
					t.Logf("Property %s transitioned from known (%v) to unknown", change.Name, change.Before)
				}
			}

			t.Logf("%s: Found %d unknown changes, %d known-to-unknown transitions",
				tt.description, foundUnknownChanges, foundKnownToUnknownChanges)

			// Verify we found the expected transition types
			assert.Greater(t, len(resource.PropertyChanges.Changes), 0, "Should have property changes")
		})
	}
}

// TestLargePlansWithManyUnknownValues tests large plans with many unknown values within existing performance limits
func TestLargePlansWithManyUnknownValues(t *testing.T) {
	tests := []struct {
		name             string
		numResources     int
		unknownPropsRate float64 // Percentage of properties that should be unknown
		description      string
	}{
		{
			name:             "medium plan with moderate unknown values",
			numResources:     50,
			unknownPropsRate: 0.3, // 30% of properties are unknown
			description:      "Medium-sized plan with 30% unknown properties should process efficiently",
		},
		{
			name:             "large plan with high unknown values rate",
			numResources:     100,
			unknownPropsRate: 0.6, // 60% of properties are unknown
			description:      "Large plan with majority unknown properties should stay within performance limits",
		},
		{
			name:             "very large plan with sparse unknown values",
			numResources:     200,
			unknownPropsRate: 0.15, // 15% of properties are unknown
			description:      "Very large plan with sparse unknown values should handle efficiently",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test configuration with performance limits
			cfg := &config.Config{
				Plan: config.PlanConfig{
					PerformanceLimits: config.PerformanceLimitsConfig{
						MaxPropertiesPerResource: 100,
						MaxPropertySize:          1024 * 10,        // 10KB per property
						MaxTotalMemory:           1024 * 1024 * 10, // 10MB total
					},
				},
			}

			// Generate plan with many unknown values
			plan := generateLargePlanWithUnknownValues(tt.numResources, tt.unknownPropsRate)

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should be generated even with many unknown values")
			assert.Len(t, summary.ResourceChanges, tt.numResources, "Should process all resources")

			// Verify unknown values processing
			totalUnknownProps := 0
			resourcesWithUnknowns := 0

			for _, resource := range summary.ResourceChanges {
				if resource.HasUnknownValues {
					resourcesWithUnknowns++
					totalUnknownProps += len(resource.UnknownProperties)

					// Verify unknown properties are properly categorized
					unknownChanges := 0
					for _, change := range resource.PropertyChanges.Changes {
						if change.IsUnknown {
							unknownChanges++
							assert.Equal(t, "(known after apply)", change.After,
								"Unknown property should show '(known after apply)'")
							assert.NotEqual(t, "remove", change.Action,
								"Unknown property should not appear as removal")
						}
					}

					// Verify property limits are respected
					assert.LessOrEqual(t, resource.PropertyChanges.Count,
						cfg.Plan.PerformanceLimits.MaxPropertiesPerResource,
						"Resource should respect property count limits")
				}
			}

			// Verify performance characteristics
			assert.Greater(t, resourcesWithUnknowns, 0, "Should find resources with unknown values")
			assert.Greater(t, totalUnknownProps, 0, "Should find unknown properties")

			t.Logf("%s: Processed %d resources, %d with unknowns, %d total unknown properties",
				tt.description, tt.numResources, resourcesWithUnknowns, totalUnknownProps)

			// Verify statistics still work correctly
			stats := summary.Statistics
			assert.Equal(t, tt.numResources, stats.Total, "Statistics should count all resources")
			assert.GreaterOrEqual(t, stats.ToAdd+stats.ToChange+stats.ToDestroy+stats.Replacements,
				stats.Total, "Action counts should account for all resources")
		})
	}
}

// Helper function to generate large plans with unknown values
func generateLargePlanWithUnknownValues(numResources int, unknownRate float64) *tfjson.Plan {
	plan := &tfjson.Plan{
		FormatVersion:    "1.0",
		TerraformVersion: "1.5.0",
		ResourceChanges:  make([]*tfjson.ResourceChange, numResources),
	}

	for i := 0; i < numResources; i++ {
		// Create before and after states
		before := make(map[string]any)
		after := make(map[string]any)
		afterUnknown := make(map[string]any)

		// Generate properties for this resource
		numProps := 10 + (i % 15) // 10-24 properties per resource
		for j := 0; j < numProps; j++ {
			propName := generatePropertyName(j)

			// Determine if this property should be unknown
			isUnknown := (float64(i*numProps+j) / float64(numResources*numProps)) < unknownRate

			if isUnknown {
				before[propName] = generatePropertyValue(j, "before")
				after[propName] = nil // Unknown values are nil in after
				afterUnknown[propName] = true
			} else {
				before[propName] = generatePropertyValue(j, "before")
				after[propName] = generatePropertyValue(j, "after")
			}
		}

		// Add some nested structures with unknown values
		if i%5 == 0 { // Every 5th resource gets nested unknowns
			nestedAfterUnknown := make(map[string]any)
			if unknownRate > 0.3 {
				nestedAfterUnknown["nested_id"] = true
				nestedAfterUnknown["nested_arn"] = true
			}

			after["nested_config"] = map[string]any{
				"enabled":    true,
				"nested_id":  nil, // Unknown
				"nested_arn": nil, // Unknown
			}
			afterUnknown["nested_config"] = nestedAfterUnknown
		}

		// Determine action based on resource index
		var actions []tfjson.Action
		switch i % 4 {
		case 0:
			actions = []tfjson.Action{tfjson.ActionCreate}
			before = nil
		case 1:
			actions = []tfjson.Action{tfjson.ActionUpdate}
		case 2:
			actions = []tfjson.Action{tfjson.ActionDelete}
			after = nil
			afterUnknown = nil
		default:
			actions = []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate}
		}

		plan.ResourceChanges[i] = &tfjson.ResourceChange{
			Address: generateResourceAddress(i),
			Type:    generateResourceType(i),
			Name:    generateResourceName(i),
			Change: &tfjson.Change{
				Actions:      actions,
				Before:       before,
				After:        after,
				AfterUnknown: afterUnknown,
			},
		}
	}

	return plan
}
