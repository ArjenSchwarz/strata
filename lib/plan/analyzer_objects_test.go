package plan

import (
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
)

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
