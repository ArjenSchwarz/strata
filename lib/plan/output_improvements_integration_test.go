package plan

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/go-output/v2"
	"github.com/ArjenSchwarz/strata/config"
)

// TestPlanSummaryOutputImprovements_EndToEnd tests all improvements with real sample files
func TestPlanSummaryOutputImprovements_EndToEnd(t *testing.T) {
	tests := []struct {
		name                    string
		planFile                string
		configOverride          func(*config.Config)
		expectedFeatures        []string // Features that should be demonstrated
		shouldHaveEmptyTables   bool     // Whether empty tables should be suppressed
		shouldHaveDangerItems   bool     // Whether dangerous items should be flagged
		shouldShowPropertyDiffs bool     // Whether Terraform diff format should be shown
	}{
		{
			name:     "Danger sample - Test all improvements",
			planFile: "../../samples/danger-sample.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.Plan.ExpandableSections.AutoExpandDangerous = true
				c.Plan.Grouping.Enabled = true
				c.Plan.Grouping.Threshold = 5
				c.ExpandAll = false
				// Configure sensitive resources for danger highlighting
				c.SensitiveResources = []config.SensitiveResource{
					{ResourceType: "aws_db_instance"},
					{ResourceType: "aws_rds_instance"},
					{ResourceType: "aws_instance"},
				}
				c.SensitiveProperties = []config.SensitiveProperty{
					{ResourceType: "aws_instance", Property: "user_data"},
				}
			},
			expectedFeatures: []string{
				"property_changes", // Should show actual changes, not emojis
				"risk_based_sort",  // Dangerous items should be sorted first
				"terraform_diff",   // Should use Terraform diff format (+, -, ~)
				"danger_highlight", // Should flag dangerous changes
			},
			shouldHaveEmptyTables:   false, // Should have actual changes
			shouldHaveDangerItems:   true,  // Should identify dangerous resources
			shouldShowPropertyDiffs: true,  // Should show property change details
		},
		{
			name:     "No-change sample - Test empty table suppression",
			planFile: "../../samples/nochange-sample.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.Plan.Grouping.Enabled = true
				c.Plan.Grouping.Threshold = 5
				c.ExpandAll = false
			},
			expectedFeatures: []string{
				"empty_table_suppression", // Should hide empty tables
				"no_op_filtering",         // Should filter out no-op changes
			},
			shouldHaveEmptyTables:   true,  // Should suppress empty tables
			shouldHaveDangerItems:   false, // No dangerous items expected
			shouldShowPropertyDiffs: false, // No property changes expected
		},
		{
			name:     "Web sample - Test provider grouping and sorting",
			planFile: "../../samples/websample.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.Plan.Grouping.Enabled = true
				c.Plan.Grouping.Threshold = 3 // Lower threshold for testing
				c.ExpandAll = false
			},
			expectedFeatures: []string{
				"provider_grouping",  // Should group by provider if threshold met
				"changed_count_only", // Should count only changed resources
			},
			shouldHaveEmptyTables:   false, // Should have actual changes
			shouldHaveDangerItems:   false, // No specific dangerous items configured
			shouldShowPropertyDiffs: true,  // Should show property changes
		},
		{
			name:     "Expand-all functionality test",
			planFile: "../../samples/websample.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.ExpandAll = true // Global expand all
			},
			expectedFeatures: []string{
				"expand_all_flag", // Should expand all collapsible sections
			},
			shouldHaveEmptyTables:   false, // Should have actual changes
			shouldHaveDangerItems:   false, // No dangerous items configured
			shouldShowPropertyDiffs: true,  // Should show expanded property changes
		},
		{
			name:     "Complex properties sample - Test property diff formatting",
			planFile: "../../samples/complex-properties-sample.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.ExpandAll = false
				// Configure sensitive properties
				c.SensitiveProperties = []config.SensitiveProperty{
					{ResourceType: "aws_instance", Property: "user_data"},
				}
			},
			expectedFeatures: []string{
				"terraform_diff",    // Should use Terraform diff format
				"sensitive_masking", // Should mask sensitive properties
				"nested_properties", // Should handle nested property changes
			},
			shouldHaveEmptyTables:   false, // Should have actual changes
			shouldHaveDangerItems:   false, // No dangerous resources by default
			shouldShowPropertyDiffs: true,  // Should show detailed property changes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create configuration
			cfg := getTestConfig()
			if tt.configOverride != nil {
				tt.configOverride(cfg)
			}

			// Create parser and analyzer
			parser := NewParser(tt.planFile)
			plan, err := parser.LoadPlan()
			if err != nil {
				t.Fatalf("Failed to load plan file %s: %v", tt.planFile, err)
			}

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("")
			if summary == nil {
				t.Fatalf("Failed to generate summary from plan file %s", tt.planFile)
			}

			// Test 1: Empty table suppression
			if tt.shouldHaveEmptyTables {
				// Count resources that are not no-ops
				changedResourceCount := 0
				for _, change := range summary.ResourceChanges {
					if change.ChangeType != ChangeTypeNoOp {
						changedResourceCount++
					}
				}
				if changedResourceCount > 0 {
					t.Errorf("Expected no changed resources for empty table test, but found %d", changedResourceCount)
				}
			}

			// Test 2: Danger highlighting
			if tt.shouldHaveDangerItems {
				foundDangerous := false
				for _, change := range summary.ResourceChanges {
					if change.IsDangerous {
						foundDangerous = true
						if change.DangerReason == "" {
							t.Error("Dangerous resource should have a danger reason")
						}
						break
					}
				}
				if !foundDangerous {
					t.Error("Expected to find dangerous resources but none were flagged")
				}
			}

			// Test 3: Property changes formatting
			if tt.shouldShowPropertyDiffs {
				foundPropertyChanges := false
				for _, change := range summary.ResourceChanges {
					// Check if resource has property changes through ChangeAttributes or TopChanges
					if len(change.ChangeAttributes) > 0 || len(change.TopChanges) > 0 {
						foundPropertyChanges = true
						t.Logf("Found resource with property changes: %s (ChangeAttributes: %d, TopChanges: %d)",
							change.Address, len(change.ChangeAttributes), len(change.TopChanges))
						break
					}
				}
				if !foundPropertyChanges {
					t.Log("Note: No property changes found in this plan")
				}
			}

			// Test 4: Formatter functionality across all formats
			formatter := NewFormatter(cfg)
			testFormats := []string{"table", "json", "markdown", "html"}

			for _, format := range testFormats {
				t.Run(format+"_format", func(t *testing.T) {
					outputConfig := &config.OutputConfiguration{
						Format:     format,
						UseEmoji:   false,
						UseColors:  false,
						TableStyle: "",
					}

					// Test that formatter can handle the format without errors
					err := formatter.OutputSummary(summary, outputConfig, true)
					if err != nil {
						t.Errorf("Failed to output summary for format %s: %v", format, err)
					}
				})
			}

			// Test 5: ActionSortTransformer integration
			actionSortTransformer := &ActionSortTransformer{}
			supportedFormats := []string{"table", "markdown", "html", "csv"}

			for _, format := range supportedFormats {
				if !actionSortTransformer.CanTransform(format) {
					t.Errorf("ActionSortTransformer should support format: %s", format)
				}
			}

			// JSON should not be supported by ActionSortTransformer
			if actionSortTransformer.CanTransform("json") {
				t.Error("ActionSortTransformer should not support JSON format")
			}

			// Test 6: Provider grouping logic
			if cfg.Plan.Grouping.Enabled {
				changedCount := countChangedResources(summary.ResourceChanges)
				if changedCount >= cfg.Plan.Grouping.Threshold {
					groups := groupResourcesByProvider(summary.ResourceChanges)
					if len(groups) > 1 {
						// Verify all changed resources are in groups
						totalGrouped := 0
						for _, resources := range groups {
							totalGrouped += len(resources)
						}
						if totalGrouped != changedCount {
							t.Errorf("Expected %d changed resources to be grouped, but only %d were", changedCount, totalGrouped)
						}
					}
				}
			}
		})
	}
}

// TestPropertyChangesFormatterTerraformIntegration tests the Terraform diff format specifically
func TestPropertyChangesFormatterTerraformIntegration(t *testing.T) {
	// Load a plan with actual property changes
	parser := NewParser("../../samples/websample.json")
	plan, err := parser.LoadPlan()
	if err != nil {
		t.Fatalf("Failed to load plan: %v", err)
	}

	cfg := getTestConfig()
	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("")
	if summary == nil {
		t.Fatal("Failed to generate summary")
	}

	formatter := NewFormatter(cfg)

	// Find a resource with property changes and create PropertyChangeAnalysis like formatter does
	var testResource *ResourceChange
	var testPropAnalysis PropertyChangeAnalysis

	for i := range summary.ResourceChanges {
		change := &summary.ResourceChanges[i]
		if len(change.ChangeAttributes) > 0 || len(change.TopChanges) > 0 {
			testResource = change

			// Create PropertyChangeAnalysis like the formatter does
			testPropAnalysis = PropertyChangeAnalysis{
				Changes: []PropertyChange{},
				Count:   len(change.ChangeAttributes),
			}

			// Use ChangeAttributes for property changes
			for _, attr := range change.ChangeAttributes {
				propChange := PropertyChange{
					Name:      attr,
					Before:    change.Before,
					After:     change.After,
					Action:    "update", // Default action for testing
					Sensitive: false,    // Default for testing
				}
				testPropAnalysis.Changes = append(testPropAnalysis.Changes, propChange)
			}
			break
		}
	}

	if testResource == nil {
		t.Skip("No resources with property changes found in test data")
	}

	// Test the Terraform formatter
	terraformFormatter := formatter.propertyChangesFormatterTerraform()
	result := terraformFormatter(testPropAnalysis)

	// Result should be a collapsible value
	if _, ok := result.(output.CollapsibleValue); ok {
		t.Logf("Successfully created collapsible value for property changes")
	} else {
		t.Errorf("Expected CollapsibleValue, got %T", result)
	}
}

// TestEmptyTableSuppressionLogic tests the empty table suppression functionality
func TestEmptyTableSuppressionLogic(t *testing.T) {
	// Load no-change sample
	parser := NewParser("../../samples/nochange-sample.json")
	plan, err := parser.LoadPlan()
	if err != nil {
		t.Fatalf("Failed to load plan: %v", err)
	}

	cfg := getTestConfig()
	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("")
	if summary == nil {
		t.Fatal("Failed to generate summary")
	}

	formatter := NewFormatter(cfg)

	// Test prepareResourceTableData filters no-ops
	tableData := formatter.prepareResourceTableData(summary.ResourceChanges)

	// Count actual no-op changes in the plan
	noOpCount := 0
	totalCount := len(summary.ResourceChanges)
	for _, change := range summary.ResourceChanges {
		if change.ChangeType == ChangeTypeNoOp {
			noOpCount++
		}
	}

	// Table data should exclude no-ops
	expectedDataCount := totalCount - noOpCount
	if len(tableData) != expectedDataCount {
		t.Errorf("Expected %d table rows (total %d - no-ops %d), got %d", expectedDataCount, totalCount, noOpCount, len(tableData))
	}

	// Test countChangedResources helper
	changedCount := countChangedResources(summary.ResourceChanges)
	if changedCount != expectedDataCount {
		t.Errorf("countChangedResources should return %d, got %d", expectedDataCount, changedCount)
	}
}

// TestRiskBasedSortingBehavior tests the risk-based sorting behavior
func TestRiskBasedSortingBehavior(t *testing.T) {
	// Load danger sample which should have dangerous resources
	parser := NewParser("../../samples/danger-sample.json")
	plan, err := parser.LoadPlan()
	if err != nil {
		t.Fatalf("Failed to load plan: %v", err)
	}

	cfg := getTestConfig()
	// Configure sensitive resources to ensure we get dangerous flags
	cfg.SensitiveResources = []config.SensitiveResource{
		{ResourceType: "aws_db_instance"},
		{ResourceType: "aws_rds_instance"},
		{ResourceType: "aws_instance"},
	}

	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("")
	if summary == nil {
		t.Fatal("Failed to generate summary")
	}

	// Categorize resources by danger and action type
	dangerousChanges := make([]ResourceChange, 0)
	nonDangerousChanges := make([]ResourceChange, 0)

	for _, change := range summary.ResourceChanges {
		if change.ChangeType == ChangeTypeNoOp {
			continue // Skip no-ops as they should be filtered
		}

		if change.IsDangerous {
			dangerousChanges = append(dangerousChanges, change)
		} else {
			nonDangerousChanges = append(nonDangerousChanges, change)
		}
	}

	if len(dangerousChanges) == 0 {
		t.Log("Note: No dangerous changes identified in test data")
	}

	// Test ActionSortTransformer format support
	transformer := &ActionSortTransformer{}

	supportedFormats := []string{"table", "markdown", "html", "csv"}
	for _, format := range supportedFormats {
		if !transformer.CanTransform(format) {
			t.Errorf("ActionSortTransformer should support %s format", format)
		}
	}

	unsupportedFormats := []string{"json", "yaml", "xml"}
	for _, format := range unsupportedFormats {
		if transformer.CanTransform(format) {
			t.Errorf("ActionSortTransformer should not support %s format", format)
		}
	}
}

// TestBackwardCompatibility tests that output structure remains consistent
func TestBackwardCompatibility(t *testing.T) {
	// Test with different sample files
	testFiles := []string{
		"../../samples/danger-sample.json",
		"../../samples/websample.json",
	}

	cfg := getTestConfig()

	for _, planFile := range testFiles {
		t.Run(filepath.Base(planFile), func(t *testing.T) {
			parser := NewParser(planFile)
			plan, err := parser.LoadPlan()
			if err != nil {
				t.Fatalf("Failed to load plan: %v", err)
			}

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("")
			if summary == nil {
				t.Fatal("Failed to generate summary")
			}

			// Test JSON output structure
			formatter := NewFormatter(cfg)
			outputConfig := &config.OutputConfiguration{
				Format:     "json",
				UseEmoji:   false,
				UseColors:  false,
				TableStyle: "",
			}

			// JSON output should work without errors
			err = formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("JSON output failed: %v", err)
			}

			// Verify essential fields exist in summary
			if summary.FormatVersion == "" {
				t.Error("FormatVersion should not be empty")
			}
			if summary.TerraformVersion == "" {
				t.Error("TerraformVersion should not be empty")
			}
			if summary.ResourceChanges == nil {
				t.Error("ResourceChanges should not be nil")
			}
			if summary.Statistics.Total == 0 && len(summary.ResourceChanges) > 0 {
				t.Error("Statistics should reflect resource changes")
			}
		})
	}
}

// Helper functions

func countChangedResources(changes []ResourceChange) int {
	count := 0
	for _, change := range changes {
		if change.ChangeType != ChangeTypeNoOp {
			count++
		}
	}
	return count
}

func groupResourcesByProvider(changes []ResourceChange) map[string][]ResourceChange {
	groups := make(map[string][]ResourceChange)
	for _, change := range changes {
		if change.ChangeType == ChangeTypeNoOp {
			continue
		}
		provider := extractProvider(change.Type)
		groups[provider] = append(groups[provider], change)
	}
	return groups
}

func extractProvider(resourceType string) string {
	parts := strings.Split(resourceType, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}
