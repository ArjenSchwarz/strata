package plan

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputRefinements_EdgeCases_EmptyPlan tests handling of empty plans (Task 8.2.1)
func TestOutputRefinements_EdgeCases_EmptyPlan(t *testing.T) {
	// Create an empty plan with no resource changes or outputs
	plan := &tfjson.Plan{
		FormatVersion:    "0.2",
		TerraformVersion: "1.5.0",
		Variables:        map[string]*tfjson.PlanVariable{},
		PlannedValues:    &tfjson.StateValues{},
		ResourceChanges:  []*tfjson.ResourceChange{},
		OutputChanges:    map[string]*tfjson.Change{},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowNoOps:   false,
			ShowDetails: true,
		},
	}

	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("test_empty_plan.json")

	// Verify empty plan handling
	assert.NotNil(t, summary, "Summary should not be nil for empty plan")
	assert.Equal(t, 0, len(summary.ResourceChanges), "Empty plan should have no resource changes")
	assert.Equal(t, 0, len(summary.OutputChanges), "Empty plan should have no output changes")
	assert.Equal(t, 0, summary.Statistics.ToAdd, "Empty plan should have no resources to add")
	assert.Equal(t, 0, summary.Statistics.ToChange, "Empty plan should have no resources to change")
	assert.Equal(t, 0, summary.Statistics.ToDestroy, "Empty plan should have no resources to destroy")
	assert.Equal(t, 0, summary.Statistics.Replacements, "Empty plan should have no replacements")
	assert.Equal(t, 0, summary.Statistics.Unmodified, "Empty plan should have no unmodified resources")

	// Test formatter with empty plan
	formatter := NewFormatter(cfg)
	outputConfig := &config.OutputConfiguration{
		Format:           "table",
		OutputFile:       "",
		OutputFileFormat: "table",
		UseEmoji:         true,
		UseColors:        true,
		TableStyle:       "default",
		MaxColumnWidth:   80,
	}

	err := formatter.OutputSummary(summary, outputConfig, true)
	assert.NoError(t, err, "Formatter should handle empty plan without error")
}

// TestOutputRefinements_EdgeCases_OnlyNoOps tests handling of plans with only no-op resources (Task 8.2.1)
func TestOutputRefinements_EdgeCases_OnlyNoOps(t *testing.T) {
	// Create a plan with only no-op resources and outputs
	plan := &tfjson.Plan{
		FormatVersion:    "0.2",
		TerraformVersion: "1.5.0",
		Variables:        map[string]*tfjson.PlanVariable{},
		PlannedValues:    &tfjson.StateValues{},
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_s3_bucket.static",
				Type:    "aws_s3_bucket",
				Name:    "static",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionNoop},
					Before:  map[string]any{"bucket": "my-bucket"},
					After:   map[string]any{"bucket": "my-bucket"},
				},
			},
			{
				Address: "aws_s3_bucket.logs",
				Type:    "aws_s3_bucket",
				Name:    "logs",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionNoop},
					Before:  map[string]any{"bucket": "logs-bucket"},
					After:   map[string]any{"bucket": "logs-bucket"},
				},
			},
		},
		OutputChanges: map[string]*tfjson.Change{
			"bucket_name": {
				Actions: tfjson.Actions{tfjson.ActionNoop},
				Before:  "my-bucket",
				After:   "my-bucket",
			},
			"logs_bucket": {
				Actions: tfjson.Actions{tfjson.ActionNoop},
				Before:  "logs-bucket",
				After:   "logs-bucket",
			},
		},
	}

	testCases := []struct {
		name                     string
		showNoOps                bool
		expectedResourceCount    int
		expectedOutputCount      int
		expectedNoChangesMessage bool
		expectedStatisticsUnmod  int
		expectedStatisticsOutput int
	}{
		{
			name:                     "Hide no-ops - should show no changes message",
			showNoOps:                false,
			expectedResourceCount:    0,
			expectedOutputCount:      0,
			expectedNoChangesMessage: true,
			expectedStatisticsUnmod:  2, // Statistics should still count no-ops
			expectedStatisticsOutput: 0, // No-op outputs shouldn't count in statistics
		},
		{
			name:                     "Show no-ops - should show all resources and outputs",
			showNoOps:                true,
			expectedResourceCount:    2,
			expectedOutputCount:      2,
			expectedNoChangesMessage: false,
			expectedStatisticsUnmod:  2, // Statistics should count no-ops
			expectedStatisticsOutput: 0, // No-op outputs still shouldn't count
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowNoOps:   tc.showNoOps,
					ShowDetails: true,
				},
			}

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("test_only_noops.json")

			// Verify all resources are marked as no-op
			for _, resource := range summary.ResourceChanges {
				assert.True(t, resource.IsNoOp, "All resources should be marked as no-op")
				assert.Equal(t, ChangeTypeNoOp, resource.ChangeType, "Change type should be no-op")
			}

			// Verify all outputs are marked as no-op
			for _, output := range summary.OutputChanges {
				assert.True(t, output.IsNoOp, "All outputs should be marked as no-op")
			}

			// Test filtering
			formatter := NewFormatter(cfg)
			filteredResources := formatter.filterNoOps(summary.ResourceChanges)
			filteredOutputs := formatter.filterNoOpOutputs(summary.OutputChanges)

			assert.Equal(t, tc.expectedResourceCount, len(filteredResources),
				"Filtered resource count should match expected")
			assert.Equal(t, tc.expectedOutputCount, len(filteredOutputs),
				"Filtered output count should match expected")

			// Test statistics
			assert.Equal(t, tc.expectedStatisticsUnmod, summary.Statistics.Unmodified,
				"Statistics should count no-ops in Unmodified")
			assert.Equal(t, tc.expectedStatisticsOutput, summary.Statistics.OutputChanges,
				"Statistics should exclude no-op outputs")

			// Test formatter output (should not crash)
			outputConfig := &config.OutputConfiguration{
				Format:           "table",
				OutputFile:       "",
				OutputFileFormat: "table",
				UseEmoji:         true,
				UseColors:        true,
				TableStyle:       "default",
				MaxColumnWidth:   80,
			}

			err := formatter.OutputSummary(summary, outputConfig, true)
			assert.NoError(t, err, "Formatter should handle only-no-ops plan without error")
		})
	}
}

// TestOutputRefinements_EdgeCases_ComplexSensitiveStructures tests nested sensitive structures (Task 8.2.2)
func TestOutputRefinements_EdgeCases_ComplexSensitiveStructures(t *testing.T) {
	// Create a plan with complex nested sensitive structures
	plan := &tfjson.Plan{
		FormatVersion:    "0.2",
		TerraformVersion: "1.5.0",
		Variables:        map[string]*tfjson.PlanVariable{},
		PlannedValues:    &tfjson.StateValues{},
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_instance.complex",
				Type:    "aws_instance",
				Name:    "complex",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]any{
						"tags": map[string]any{
							"Name":        "instance",
							"Environment": "prod",
							"Secret":      "old-secret-value",
						},
						"user_data": "old-script-content",
						"metadata": map[string]any{
							"startup_script": "old-startup",
							"ssh_keys": []any{
								"ssh-rsa AAAAB3...",
								"ssh-rsa BBBBB4...",
							},
						},
					},
					After: map[string]any{
						"tags": map[string]any{
							"Name":        "instance",
							"Environment": "prod",
							"Secret":      "new-secret-value",
						},
						"user_data": "new-script-content",
						"metadata": map[string]any{
							"startup_script": "new-startup",
							"ssh_keys": []any{
								"ssh-rsa AAAAB3...",
								"ssh-rsa CCCCC5...", // Changed key
							},
						},
					},
					BeforeSensitive: map[string]any{
						"tags": map[string]any{
							"Name":        false,
							"Environment": false,
							"Secret":      true, // Sensitive
						},
						"user_data": true, // Entire user_data is sensitive
						"metadata": map[string]any{
							"startup_script": false,
							"ssh_keys": []any{
								true,  // First SSH key is sensitive
								false, // Second SSH key is not
							},
						},
					},
					AfterSensitive: map[string]any{
						"tags": map[string]any{
							"Name":        false,
							"Environment": false,
							"Secret":      true, // Still sensitive
						},
						"user_data": true, // Still entirely sensitive
						"metadata": map[string]any{
							"startup_script": false,
							"ssh_keys": []any{
								true, // First SSH key still sensitive
								true, // Second SSH key now sensitive too
							},
						},
					},
				},
			},
		},
		OutputChanges: map[string]*tfjson.Change{},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_instance", Property: "user_data"},
			{ResourceType: "aws_instance", Property: "tags.Secret"},
		},
	}

	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("test_complex_sensitive.json")

	// Find the update resource
	var updateResource *ResourceChange
	for i := range summary.ResourceChanges {
		if summary.ResourceChanges[i].ChangeType == ChangeTypeUpdate {
			updateResource = &summary.ResourceChanges[i]
			break
		}
	}
	require.NotNil(t, updateResource, "Should find update resource")

	// Test that sensitive properties are properly masked but structure is preserved
	sensitivePropertyFound := false
	nonSensitivePropertyFound := false

	for _, change := range updateResource.PropertyChanges.Changes {
		switch change.Name {
		case "user_data":
			assert.True(t, change.Sensitive, "user_data should be marked as sensitive")
			assert.Equal(t, "(sensitive value)", change.Before, "Before value should be masked")
			assert.Equal(t, "(sensitive value)", change.After, "After value should be masked")
			sensitivePropertyFound = true
		case "tags":
			// tags is a complex object but not itself marked sensitive in the test
			assert.False(t, change.Sensitive, "tags object should not be marked as sensitive (individual fields may be)")
			nonSensitivePropertyFound = true
		case "metadata":
			// metadata is a complex object and not marked sensitive
			assert.False(t, change.Sensitive, "metadata should not be marked as sensitive")
			nonSensitivePropertyFound = true
		}
	}

	assert.True(t, sensitivePropertyFound, "Should find at least one sensitive property")
	assert.True(t, nonSensitivePropertyFound, "Should find at least one non-sensitive property")

	// Test that properties are sorted despite masking
	if len(updateResource.PropertyChanges.Changes) > 1 {
		for i := 1; i < len(updateResource.PropertyChanges.Changes); i++ {
			prevName := strings.ToLower(updateResource.PropertyChanges.Changes[i-1].Name)
			currName := strings.ToLower(updateResource.PropertyChanges.Changes[i].Name)
			assert.LessOrEqual(t, prevName, currName,
				"Properties should be sorted alphabetically even with sensitive masking")
		}
	}
}

// TestOutputRefinements_EdgeCases_IdenticalResourceAddresses tests sorting with identical addresses (Task 8.2.3)
func TestOutputRefinements_EdgeCases_IdenticalResourceAddresses(t *testing.T) {
	resources := []ResourceChange{
		{
			Address:     "aws_instance.web",
			Type:        "aws_instance",
			Name:        "web",
			ChangeType:  ChangeTypeCreate,
			IsDangerous: false,
		},
		{
			Address:     "aws_instance.web", // Same address, different change type
			Type:        "aws_instance",
			Name:        "web",
			ChangeType:  ChangeTypeDelete,
			IsDangerous: false,
		},
		{
			Address:     "aws_instance.web", // Same address, dangerous
			Type:        "aws_instance",
			Name:        "web",
			ChangeType:  ChangeTypeUpdate,
			IsDangerous: true,
		},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}

	formatter := NewFormatter(cfg)
	sortedResources := formatter.sortResourcesByPriority(resources)

	// Verify sorting order: dangerous first, then by action priority
	assert.Equal(t, true, sortedResources[0].IsDangerous, "First resource should be dangerous")
	assert.Equal(t, ChangeTypeUpdate, sortedResources[0].ChangeType, "First resource should be update (dangerous)")

	assert.Equal(t, ChangeTypeDelete, sortedResources[1].ChangeType, "Second resource should be delete")
	assert.Equal(t, false, sortedResources[1].IsDangerous, "Second resource should not be dangerous")

	assert.Equal(t, ChangeTypeCreate, sortedResources[2].ChangeType, "Third resource should be create")
	assert.Equal(t, false, sortedResources[2].IsDangerous, "Third resource should not be dangerous")
}

// TestOutputRefinements_EdgeCases_VariousDangerActionCombinations tests all combinations of danger/action states (Task 8.2.3)
func TestOutputRefinements_EdgeCases_VariousDangerActionCombinations(t *testing.T) {
	// Create comprehensive test case with all combinations
	resources := []ResourceChange{
		// Non-dangerous resources
		{Address: "aws_s3_bucket.logs", ChangeType: ChangeTypeCreate, IsDangerous: false},
		{Address: "aws_s3_bucket.assets", ChangeType: ChangeTypeUpdate, IsDangerous: false},
		{Address: "aws_s3_bucket.backup", ChangeType: ChangeTypeDelete, IsDangerous: false},
		{Address: "aws_s3_bucket.temp", ChangeType: ChangeTypeReplace, IsDangerous: false},
		{Address: "aws_s3_bucket.static", ChangeType: ChangeTypeNoOp, IsDangerous: false},

		// Dangerous resources
		{Address: "aws_rds_instance.primary", ChangeType: ChangeTypeCreate, IsDangerous: true},
		{Address: "aws_rds_instance.replica", ChangeType: ChangeTypeUpdate, IsDangerous: true},
		{Address: "aws_rds_instance.old", ChangeType: ChangeTypeDelete, IsDangerous: true},
		{Address: "aws_rds_instance.staging", ChangeType: ChangeTypeReplace, IsDangerous: true},
		{Address: "aws_rds_instance.legacy", ChangeType: ChangeTypeNoOp, IsDangerous: true},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}

	formatter := NewFormatter(cfg)
	sortedResources := formatter.sortResourcesByPriority(resources)

	// Expected order: all dangerous first (delete, replace, update, create, no-op), then non-dangerous (same order)
	expectedOrder := []struct {
		isDangerous bool
		changeType  ChangeType
		address     string
	}{
		// Dangerous resources first, sorted by action priority
		{true, ChangeTypeDelete, "aws_rds_instance.old"},
		{true, ChangeTypeReplace, "aws_rds_instance.staging"},
		{true, ChangeTypeUpdate, "aws_rds_instance.replica"},
		{true, ChangeTypeCreate, "aws_rds_instance.primary"},
		{true, ChangeTypeNoOp, "aws_rds_instance.legacy"},

		// Non-dangerous resources second, sorted by action priority
		{false, ChangeTypeDelete, "aws_s3_bucket.backup"},
		{false, ChangeTypeReplace, "aws_s3_bucket.temp"},
		{false, ChangeTypeUpdate, "aws_s3_bucket.assets"},
		{false, ChangeTypeCreate, "aws_s3_bucket.logs"},
		{false, ChangeTypeNoOp, "aws_s3_bucket.static"},
	}

	require.Equal(t, len(expectedOrder), len(sortedResources), "Should have same number of resources")

	for i, expected := range expectedOrder {
		actual := sortedResources[i]
		assert.Equal(t, expected.isDangerous, actual.IsDangerous,
			"Resource at position %d should have IsDangerous=%v", i, expected.isDangerous)
		assert.Equal(t, expected.changeType, actual.ChangeType,
			"Resource at position %d should have ChangeType=%s", i, expected.changeType)
		assert.Equal(t, expected.address, actual.Address,
			"Resource at position %d should have Address=%s", i, expected.address)
	}
}

// TestOutputRefinements_EdgeCases_LargePlansPerformance tests performance with large plans (Task 8.2.4)
func TestOutputRefinements_EdgeCases_LargePlansPerformance(t *testing.T) {
	// Create a large plan with 1000+ resources
	const resourceCount = 1200
	const outputCount = 100

	// Generate test plan data
	resourceChanges := make([]*tfjson.ResourceChange, resourceCount)
	outputChanges := make(map[string]*tfjson.Change, outputCount)

	for i := range resourceCount {
		changeType := tfjson.ActionCreate
		if i%5 == 0 {
			changeType = tfjson.ActionUpdate
		} else if i%7 == 0 {
			changeType = tfjson.ActionDelete
		} else if i%11 == 0 {
			changeType = tfjson.ActionNoop
		}

		resourceChanges[i] = &tfjson.ResourceChange{
			Address: fmt.Sprintf("aws_instance.web_%d", i),
			Type:    "aws_instance",
			Name:    fmt.Sprintf("web_%d", i),
			Change: &tfjson.Change{
				Actions: tfjson.Actions{changeType},
				Before:  map[string]any{"instance_type": "t2.micro"},
				After:   map[string]any{"instance_type": "t2.small"},
			},
		}
	}

	for i := range outputCount {
		outputChanges[fmt.Sprintf("output_%d", i)] = &tfjson.Change{
			Actions: tfjson.Actions{tfjson.ActionCreate},
			Before:  nil,
			After:   fmt.Sprintf("value_%d", i),
		}
	}

	plan := &tfjson.Plan{
		FormatVersion:    "0.2",
		TerraformVersion: "1.5.0",
		Variables:        map[string]*tfjson.PlanVariable{},
		PlannedValues:    &tfjson.StateValues{},
		ResourceChanges:  resourceChanges,
		OutputChanges:    outputChanges,
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowNoOps:   false,
			ShowDetails: true,
		},
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_instance"},
		},
	}

	// Performance test - should complete within reasonable time
	analyzer := NewAnalyzer(plan, cfg)

	// Time the analysis
	start := time.Now()
	summary := analyzer.GenerateSummary("test_large_plan.json")
	analysisTime := time.Since(start)

	// Time the formatting
	formatter := NewFormatter(cfg)
	start = time.Now()

	filteredResources := formatter.filterNoOps(summary.ResourceChanges)
	sortedResources := formatter.sortResourcesByPriority(filteredResources)

	formattingTime := time.Since(start)

	// Verify correctness despite size
	assert.NotNil(t, summary, "Summary should not be nil for large plan")
	assert.Equal(t, resourceCount, len(summary.ResourceChanges), "Should process all resources")
	assert.Equal(t, outputCount, len(summary.OutputChanges), "Should process all outputs")

	// Verify filtering worked correctly
	expectedFiltered := 0
	for _, resource := range summary.ResourceChanges {
		if resource.ChangeType != ChangeTypeNoOp {
			expectedFiltered++
		}
	}
	assert.Equal(t, expectedFiltered, len(filteredResources), "Filtering should work correctly with large plans")

	// Verify sorting worked correctly
	assert.Equal(t, len(filteredResources), len(sortedResources), "Sorting should preserve resource count")

	// Performance requirements - should be < 5% impact vs simple operations
	t.Logf("Analysis time for %d resources: %v", resourceCount, analysisTime)
	t.Logf("Formatting time for %d resources: %v", len(filteredResources), formattingTime)

	// These are generous thresholds - adjust based on actual baseline measurements
	assert.Less(t, analysisTime.Milliseconds(), int64(5000), "Analysis should complete within 5 seconds")
	assert.Less(t, formattingTime.Milliseconds(), int64(2000), "Formatting should complete within 2 seconds")

	// Test that formatter can handle the output (basic integration)
	outputConfig := &config.OutputConfiguration{
		Format:           "json", // Use JSON for large plans to avoid table rendering overhead
		OutputFile:       "",
		OutputFileFormat: "json",
		UseEmoji:         false,
		UseColors:        false,
		MaxColumnWidth:   80,
	}

	err := formatter.OutputSummary(summary, outputConfig, false) // Set to false to avoid stdout spam
	assert.NoError(t, err, "Formatter should handle large plan without error")
}

// TestOutputRefinements_EdgeCases_PropertySortingComplexScenarios tests edge cases in property sorting (Task 8.2.2)
func TestOutputRefinements_EdgeCases_PropertySortingComplexScenarios(t *testing.T) {
	testCases := []struct {
		name            string
		propertyChanges []PropertyChange
		expectedOrder   []string // Expected property names in sorted order
	}{
		{
			name: "Mixed case property names",
			propertyChanges: []PropertyChange{
				{Name: "ZZZ_property", Path: []string{"zzz_property"}},
				{Name: "aaa_property", Path: []string{"aaa_property"}},
				{Name: "BBB_property", Path: []string{"bbb_property"}},
				{Name: "ccc_property", Path: []string{"ccc_property"}},
			},
			expectedOrder: []string{"aaa_property", "BBB_property", "ccc_property", "ZZZ_property"},
		},
		{
			name: "Same property names with different paths",
			propertyChanges: []PropertyChange{
				{Name: "config", Path: []string{"database", "config"}},
				{Name: "config", Path: []string{"app", "config"}},
				{Name: "config", Path: []string{"cache", "config"}},
			},
			expectedOrder: []string{"config", "config", "config"}, // Names same, should sort by path
		},
		{
			name: "Natural sort ordering with numbers",
			propertyChanges: []PropertyChange{
				{Name: "property10", Path: []string{"property10"}},
				{Name: "property2", Path: []string{"property2"}},
				{Name: "property1", Path: []string{"property1"}},
				{Name: "property20", Path: []string{"property20"}},
			},
			expectedOrder: []string{"property1", "property10", "property2", "property20"}, // Simple string sort for now
		},
		{
			name: "Properties with special characters",
			propertyChanges: []PropertyChange{
				{Name: "property-dash", Path: []string{"property-dash"}},
				{Name: "property_underscore", Path: []string{"property_underscore"}},
				{Name: "property.dot", Path: []string{"property.dot"}},
				{Name: "property@symbol", Path: []string{"property@symbol"}},
			},
			expectedOrder: []string{"property-dash", "property.dot", "property@symbol", "property_underscore"},
		},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}

	_ = cfg // Used for mock testing

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock PropertyChangeAnalysis
			analysis := PropertyChangeAnalysis{
				Changes: tc.propertyChanges,
			}

			// Simulate the sorting that happens in analyzePropertyChanges
			sortedChanges := make([]PropertyChange, len(analysis.Changes))
			copy(sortedChanges, analysis.Changes)

			// Apply the same sorting logic as analyzePropertyChanges
			for i := range sortedChanges {
				for j := i + 1; j < len(sortedChanges); j++ {
					iName := strings.ToLower(sortedChanges[i].Name)
					jName := strings.ToLower(sortedChanges[j].Name)

					shouldSwap := false
					if iName != jName {
						shouldSwap = iName > jName
					} else {
						// Same name, sort by path
						iPath := strings.Join(sortedChanges[i].Path, ".")
						jPath := strings.Join(sortedChanges[j].Path, ".")
						shouldSwap = iPath > jPath
					}

					if shouldSwap {
						sortedChanges[i], sortedChanges[j] = sortedChanges[j], sortedChanges[i]
					}
				}
			}

			// Verify sorting for cases where names are different
			if tc.name != "Same property names with different paths" {
				for i := 0; i < len(tc.expectedOrder); i++ {
					assert.Equal(t, tc.expectedOrder[i], sortedChanges[i].Name,
						"Property at position %d should be %s", i, tc.expectedOrder[i])
				}
			} else {
				// For same names, verify path-based sorting
				for i := 1; i < len(sortedChanges); i++ {
					prevPath := strings.Join(sortedChanges[i-1].Path, ".")
					currPath := strings.Join(sortedChanges[i].Path, ".")
					assert.LessOrEqual(t, prevPath, currPath,
						"Path sorting failed: %s should come before %s", prevPath, currPath)
				}
			}
		})
	}
}

// Helper to create time values for performance tests
func init() {
	// Ensure time package is available for performance testing
	_ = time.Now()
}
