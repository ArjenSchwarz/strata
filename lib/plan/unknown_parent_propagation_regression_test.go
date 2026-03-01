package plan

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnknownParentPropagationToNestedProperties tests that when a parent property
// is marked as unknown in after_unknown, all nested properties should also be treated as unknown
func TestUnknownParentPropagationToNestedProperties(t *testing.T) {
	tests := map[string]struct {
		plan                *tfjson.Plan
		expectedUnknownPath []string
		description         string
	}{
		"parent_object_unknown_should_propagate_to_children": {
			plan: &tfjson.Plan{
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
								"ami":           "ami-12345",
								"instance_type": "t3.micro",
								"config": map[string]any{
									"timeout":     30,
									"retry_count": 3,
									"nested": map[string]any{
										"deep_value": "test",
									},
								},
							},
							// The entire config object is unknown
							AfterUnknown: map[string]any{
								"config": true, // This should propagate to all nested properties
							},
						},
					},
				},
			},
			expectedUnknownPath: []string{
				"config",
				"config.timeout",
				"config.retry_count", 
				"config.nested",
				"config.nested.deep_value",
			},
			description: "When parent config is unknown, all nested properties should be unknown",
		},
		"mixed_parent_unknown_with_specific_children": {
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.5.0",
				ResourceChanges: []*tfjson.ResourceChange{
					{
						Address: "aws_instance.mixed",
						Type:    "aws_instance",
						Name:    "mixed",
						Change: &tfjson.Change{
							Actions: []tfjson.Action{tfjson.ActionUpdate},
							Before: map[string]any{
								"ami": "ami-12345",
								"tags": map[string]any{
									"Name":        "old-name",
									"Environment": "dev",
								},
							},
							After: map[string]any{
								"ami": "ami-67890",
								"tags": map[string]any{
									"Name":        "new-name",
									"Environment": "prod",
									"Project":     "new-project",
								},
							},
							// tags is entirely unknown
							AfterUnknown: map[string]any{
								"tags": true, // Entire tags object is unknown
							},
						},
					},
				},
			},
			expectedUnknownPath: []string{
				"tags",           // Parent unknown
				"tags.Name",      // Should inherit from parent
				"tags.Environment", // Should inherit from parent
				"tags.Project",   // Should inherit from parent
			},
			description: "When entire tags object is unknown, all nested properties should be unknown",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := getTestConfig()
			analyzer := NewAnalyzer(tc.plan, cfg)
			summary := analyzer.GenerateSummary("")

			require.NotNil(t, summary, "Summary should not be nil")
			require.Len(t, summary.ResourceChanges, 1, "Should have exactly one resource change")

			resource := summary.ResourceChanges[0]
			assert.True(t, resource.HasUnknownValues, "Resource should have unknown values")

			// Check that all expected paths are marked as unknown
			for _, expectedPath := range tc.expectedUnknownPath {
				// Check if the path is in UnknownProperties
				assert.Contains(t, resource.UnknownProperties, expectedPath,
					"Expected unknown property path %s should be identified", expectedPath)

				// Also check if property changes correctly show unknown status
				found := false
				for _, change := range resource.PropertyChanges.Changes {
					changePath := change.Name
					if len(change.Path) > 1 {
						changePath = joinPath(change.Path)
					}
					if changePath == expectedPath && change.IsUnknown {
						found = true
						assert.Equal(t, "(known after apply)", change.After,
							"Unknown property %s should display '(known after apply)'", expectedPath)
						break
					}
				}
				
				// For nested properties under unknown parents, they might not have individual PropertyChange entries
				// if they're grouped, so we mainly check UnknownProperties
				t.Logf("Path %s - Found in UnknownProperties: %v, Found in PropertyChanges: %v", 
					expectedPath, 
					contains(resource.UnknownProperties, expectedPath),
					found)
			}

			t.Logf("%s: Resource has %d unknown properties: %v", 
				tc.description, len(resource.UnknownProperties), resource.UnknownProperties)
		})
	}
}

// Helper function to join path components
func joinPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	if len(path) == 1 {
		return path[0]
	}
	
	result := path[0]
	for i := 1; i < len(path); i++ {
		// Check if the component looks like an array index
		if len(path[i]) > 0 && path[i][0] >= '0' && path[i][0] <= '9' {
			result += "[" + path[i] + "]"
		} else {
			result += "." + path[i]
		}
	}
	return result
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}