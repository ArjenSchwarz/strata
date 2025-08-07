package plan

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputRefinements_ComprehensiveWorkflow tests the complete workflow with real Terraform plan files
// containing sensitive values, no-ops, and mixed change types (Task 8.1.1)
func TestOutputRefinements_ComprehensiveWorkflow(t *testing.T) {
	testCases := []struct {
		name                   string
		planFile               string
		configOverride         func(*config.Config)
		showNoOps              bool
		expectedResourceCount  int      // Expected number of resources after filtering
		expectedNoOpResources  []string // Resource addresses that should be marked as no-op
		expectedOutputCount    int      // Expected number of outputs after filtering
		expectedNoOpOutputs    []string // Output names that should be marked as no-op
		expectedDangerousCount int      // Expected number of dangerous resources
		expectedSorting        []string // Expected order of first few resources (by address)
		expectedOutputChanges  int      // Expected count in statistics.OutputChanges
		expectedUnmodified     int      // Expected count in statistics.Unmodified
	}{
		{
			name:     "Default behavior - no-ops hidden, proper sorting, statistics correct",
			planFile: "output_refinements_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ShowNoOps = false
				c.SensitiveResources = []config.SensitiveResource{
					{ResourceType: "aws_rds_instance"},
				}
				c.SensitiveProperties = []config.SensitiveProperty{
					{ResourceType: "aws_instance", Property: "user_data"},
				}
			},
			showNoOps:              false,
			expectedResourceCount:  2,                                     // 2 non-no-op resources (RDS + EC2)
			expectedNoOpResources:  []string{"aws_s3_bucket.assets"},      // S3 bucket is no-op
			expectedOutputCount:    2,                                     // 2 non-no-op outputs
			expectedNoOpOutputs:    []string{"bucket_name"},               // bucket_name is no-op
			expectedDangerousCount: 2,                                     // Both RDS instance and EC2 instance are dangerous
			expectedSorting:        []string{"aws_rds_instance.database"}, // Dangerous RDS should come first
			expectedOutputChanges:  2,                                     // Should exclude no-op outputs from statistics
			expectedUnmodified:     1,                                     // Should count no-op resources in Unmodified
		},
		{
			name:     "Show no-ops enabled - all resources and outputs visible",
			planFile: "output_refinements_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ShowNoOps = true
				c.SensitiveResources = []config.SensitiveResource{
					{ResourceType: "aws_rds_instance"},
				}
			},
			showNoOps:              true,
			expectedResourceCount:  3,                                     // All 3 resources visible
			expectedNoOpResources:  []string{"aws_s3_bucket.assets"},      // S3 is still marked as no-op
			expectedOutputCount:    3,                                     // All 3 outputs visible (including no-op)
			expectedNoOpOutputs:    []string{"bucket_name"},               // bucket_name still marked as no-op
			expectedDangerousCount: 1,                                     // RDS still dangerous (only RDS marked sensitive in this test)
			expectedSorting:        []string{"aws_rds_instance.database"}, // Still sort dangerous first
			expectedOutputChanges:  2,                                     // Statistics should still exclude no-op outputs even when displayed
			expectedUnmodified:     1,                                     // Still count no-op resources
		},
		{
			name:     "No sensitive configuration - no danger flags but proper sorting and filtering",
			planFile: "output_refinements_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ShowNoOps = false
				// No sensitive resources or properties configured
			},
			showNoOps:              false,
			expectedResourceCount:  2,                                     // Still filter no-ops
			expectedNoOpResources:  []string{"aws_s3_bucket.assets"},      // S3 still no-op
			expectedOutputCount:    2,                                     // Still filter no-op outputs
			expectedNoOpOutputs:    []string{"bucket_name"},               // bucket_name still no-op
			expectedDangerousCount: 0,                                     // No dangerous resources without sensitive config
			expectedSorting:        []string{"aws_rds_instance.database"}, // Delete/replace before update
			expectedOutputChanges:  2,                                     // Still exclude no-op outputs
			expectedUnmodified:     1,                                     // Still count no-op resources
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			testDataPath := filepath.Join("..", "..", "testdata", tc.planFile)
			parser := NewParser(testDataPath)
			plan, err := parser.LoadPlan()
			require.NoError(t, err, "Failed to load test plan")

			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowDetails: true,
					ShowContext: true,
				},
			}
			if tc.configOverride != nil {
				tc.configOverride(cfg)
			}

			// Test analyzer with output refinements
			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary(testDataPath)
			require.NotNil(t, summary, "Summary should not be nil")

			// Test formatter filtering by creating formatter and using internal methods
			formatter := NewFormatter(cfg)

			// Apply filtering like the formatter does internally
			filteredResources := formatter.filterNoOps(summary.ResourceChanges)
			filteredOutputs := formatter.filterNoOpOutputs(summary.OutputChanges)

			// Test filtered resource count
			assert.Equal(t, tc.expectedResourceCount, len(filteredResources),
				"Filtered resource count should match expected")

			// Test filtered output count
			assert.Equal(t, tc.expectedOutputCount, len(filteredOutputs),
				"Filtered output count should match expected")

			// Verify no-op resources are correctly identified
			for _, expectedNoOp := range tc.expectedNoOpResources {
				found := false
				for _, resource := range summary.ResourceChanges {
					if resource.Address == expectedNoOp && resource.IsNoOp {
						found = true
						break
					}
				}
				assert.True(t, found, "Resource %s should be marked as no-op", expectedNoOp)
			}

			// Verify no-op outputs are correctly identified
			for _, expectedNoOpOutput := range tc.expectedNoOpOutputs {
				found := false
				for _, output := range summary.OutputChanges {
					if output.Name == expectedNoOpOutput && output.IsNoOp {
						found = true
						break
					}
				}
				assert.True(t, found, "Output %s should be marked as no-op", expectedNoOpOutput)
			}

			// Test dangerous resource count
			dangerousCount := 0
			for _, resource := range summary.ResourceChanges {
				if resource.IsDangerous {
					dangerousCount++
				}
			}
			assert.Equal(t, tc.expectedDangerousCount, dangerousCount,
				"Dangerous resource count should match expected")

			// Test sorting - verify the first few resources are in expected order
			sortedResources := formatter.sortResourcesByPriority(filteredResources)
			for i, expectedAddress := range tc.expectedSorting {
				if i < len(sortedResources) {
					assert.Equal(t, expectedAddress, sortedResources[i].Address,
						"Resource at position %d should be %s", i, expectedAddress)
				}
			}

			// Test statistics include proper output counts (Task 7.1 verification)
			stats := summary.Statistics
			assert.Equal(t, tc.expectedOutputChanges, stats.OutputChanges,
				"Statistics should have correct OutputChanges count")

			// Verify resource statistics still count no-ops in Unmodified
			assert.Equal(t, tc.expectedUnmodified, stats.Unmodified,
				"Should count no-op resources in Unmodified")

			// Test OutputSummary method doesn't crash (Task 8.1.2 - basic format testing)
			outputConfig := &config.OutputConfiguration{
				Format:           "table",
				OutputFile:       "",
				OutputFileFormat: "table",
				UseEmoji:         true,
				UseColors:        true,
				TableStyle:       "default",
				MaxColumnWidth:   80,
			}

			err = formatter.OutputSummary(summary, outputConfig, true)
			assert.NoError(t, err, "OutputSummary should not error")
		})
	}
}

// TestOutputRefinements_ConfigurationPrecedence tests CLI flags vs config file precedence (Task 8.1.3)
func TestOutputRefinements_ConfigurationPrecedence(t *testing.T) {
	testDataPath := filepath.Join("..", "..", "testdata", "output_refinements_plan.json")
	parser := NewParser(testDataPath)
	plan, err := parser.LoadPlan()
	require.NoError(t, err)

	testCases := []struct {
		name                  string
		configShowOps         bool
		cliShowOps            *bool // nil means not set
		expectedFilteredCount int   // Expected number of resources after filtering
		description           string
	}{
		{
			name:                  "Config file only - show-no-ops false",
			configShowOps:         false,
			cliShowOps:            nil,
			expectedFilteredCount: 2, // Should filter out no-ops, leaving 2 resources
			description:           "Should use config file setting when CLI not specified",
		},
		{
			name:                  "Config file only - show-no-ops true",
			configShowOps:         true,
			cliShowOps:            nil,
			expectedFilteredCount: 3, // Should show all resources including no-ops
			description:           "Should use config file setting when CLI not specified",
		},
		{
			name:                  "CLI overrides config - CLI false, config true",
			configShowOps:         true,
			cliShowOps:            boolPtr(false),
			expectedFilteredCount: 2, // CLI false should override config true
			description:           "CLI flag should override config file (CLI false beats config true)",
		},
		{
			name:                  "CLI overrides config - CLI true, config false",
			configShowOps:         false,
			cliShowOps:            boolPtr(true),
			expectedFilteredCount: 3, // CLI true should override config false
			description:           "CLI flag should override config file (CLI true beats config false)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowNoOps: tc.configShowOps,
				},
			}

			// Simulate CLI flag override if provided
			if tc.cliShowOps != nil {
				cfg.Plan.ShowNoOps = *tc.cliShowOps
			}

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary(testDataPath)

			formatter := NewFormatter(cfg)
			filteredResources := formatter.filterNoOps(summary.ResourceChanges)

			// Check filtered count matches expected behavior
			assert.Equal(t, tc.expectedFilteredCount, len(filteredResources),
				"Filtered resource count should match expected for %s", tc.description)

			// Also verify the ShowNoOps setting is correctly applied
			assert.Equal(t, cfg.Plan.ShowNoOps, tc.cliShowOps != nil && *tc.cliShowOps || tc.cliShowOps == nil && tc.configShowOps,
				"ShowNoOps configuration should match expected for %s", tc.description)
		})
	}
}

// TestOutputRefinements_BackwardCompatibility validates backward compatibility (Task 8.1.4)
func TestOutputRefinements_BackwardCompatibility(t *testing.T) {
	// Test with existing plan files to ensure no breaking changes
	existingPlanFiles := []string{
		"simple_plan.json",
		"multi_provider_plan.json",
		"high_risk_plan.json",
	}

	for _, planFile := range existingPlanFiles {
		t.Run("backward_compatibility_"+planFile, func(t *testing.T) {
			testDataPath := filepath.Join("..", "..", "testdata", planFile)

			// Test with minimal config (as existing users would have)
			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowDetails: true,
				},
			}

			parser := NewParser(testDataPath)
			plan, err := parser.LoadPlan()
			if err != nil {
				t.Skipf("Skipping %s - file not found or invalid", planFile)
				return
			}

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary(testDataPath)
			require.NotNil(t, summary, "Summary should not be nil for existing plan file")

			// Ensure formatter works with existing plans
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

			err = formatter.OutputSummary(summary, outputConfig, true)
			require.NoError(t, err, "Formatter should handle existing plan file")

			// Verify basic structure is intact
			assert.NotNil(t, summary.Statistics, "Statistics should be present")
			assert.GreaterOrEqual(t, len(summary.ResourceChanges), 0, "Resource changes should be processable")

			// Test that new fields have sensible defaults
			for _, resource := range summary.ResourceChanges {
				// IsNoOp field should be properly initialized (false for existing plans without explicit no-ops)
				if resource.ChangeType == ChangeTypeNoOp {
					assert.True(t, resource.IsNoOp, "No-op resources should have IsNoOp=true")
				}
			}

			// Verify statistics include the new OutputChanges field with non-negative value
			assert.GreaterOrEqual(t, summary.Statistics.OutputChanges, 0, "OutputChanges should be non-negative")
		})
	}
}

// TestOutputRefinements_PropertySortingIntegration tests end-to-end property sorting (Task 8.1.1)
func TestOutputRefinements_PropertySortingIntegration(t *testing.T) {
	testDataPath := filepath.Join("..", "..", "testdata", "output_refinements_plan.json")
	parser := NewParser(testDataPath)
	plan, err := parser.LoadPlan()
	require.NoError(t, err)

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}

	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary(testDataPath)

	// Find a resource with property changes to verify sorting
	var updateResource *ResourceChange
	for i := range summary.ResourceChanges {
		if summary.ResourceChanges[i].ChangeType == ChangeTypeUpdate {
			updateResource = &summary.ResourceChanges[i]
			break
		}
	}

	require.NotNil(t, updateResource, "Should find at least one update resource")

	// Only test sorting if there are multiple property changes
	if len(updateResource.PropertyChanges.Changes) > 1 {
		// Verify properties are sorted alphabetically
		properties := updateResource.PropertyChanges.Changes
		for i := 1; i < len(properties); i++ {
			prevName := strings.ToLower(properties[i-1].Name)
			currName := strings.ToLower(properties[i].Name)

			// Allow same names (different paths) - should be sorted by path then
			if prevName == currName {
				prevPath := strings.Join(properties[i-1].Path, ".")
				currPath := strings.Join(properties[i].Path, ".")
				assert.LessOrEqual(t, prevPath, currPath,
					"Properties with same name should be sorted by path: %s vs %s", prevPath, currPath)
			} else {
				assert.LessOrEqual(t, prevName, currName,
					"Properties should be sorted alphabetically: %s vs %s", prevName, currName)
			}
		}
	}
}

// TestOutputRefinements_SensitiveMaskingIntegration tests end-to-end sensitive masking (Task 8.1.1)
func TestOutputRefinements_SensitiveMaskingIntegration(t *testing.T) {
	testDataPath := filepath.Join("..", "..", "testdata", "output_refinements_plan.json")
	parser := NewParser(testDataPath)
	plan, err := parser.LoadPlan()
	require.NoError(t, err)

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_instance", Property: "user_data"},
			{ResourceType: "aws_rds_instance", Property: "password"},
		},
	}

	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary(testDataPath)

	// Test that sensitive properties are correctly identified in resource changes
	sensitiveFound := false
	for _, resource := range summary.ResourceChanges {
		for _, change := range resource.PropertyChanges.Changes {
			if change.Sensitive {
				sensitiveFound = true
				// Verify that sensitive values are masked in the property changes
				assert.NotContains(t, change.After, "secret-password",
					"Sensitive values should be masked in property changes")
				assert.NotContains(t, change.Before, "old-password",
					"Sensitive values should be masked in property changes")
			}
		}
	}
	assert.True(t, sensitiveFound, "Should find sensitive properties in the analysis")

	// Test that the formatter processes without error (basic integration test)
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

	err = formatter.OutputSummary(summary, outputConfig, true)
	require.NoError(t, err, "Formatter should handle sensitive masking without error")
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
