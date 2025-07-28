package plan

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
)

// TestEnhancedSummaryVisualization_EndToEnd tests the complete flow from plan parsing to formatted output
func TestEnhancedSummaryVisualization_EndToEnd(t *testing.T) {
	tests := []struct {
		name             string
		planFile         string
		configOverride   func(*config.Config)
		expectedOutputs  []string // Strings that should appear in resource addresses
		minResourceCount int      // Minimum number of resources expected
	}{
		{
			name:     "Simple plan with basic resources",
			planFile: "simple_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.ExpandAll = false
			},
			expectedOutputs: []string{
				"aws_instance.web_server",
				"aws_security_group.web_sg",
			},
			minResourceCount: 2,
		},
		{
			name:     "Multi-provider plan with grouping",
			planFile: "multi_provider_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.Plan.Grouping.Enabled = true
				c.Plan.Grouping.Threshold = 5 // Lower threshold for testing
				c.ExpandAll = false
			},
			expectedOutputs: []string{
				"aws_instance",
				"azurerm_virtual_network",
				"google_compute_instance",
				"kubernetes_namespace",
			},
			minResourceCount: 8,
		},
		{
			name:     "High-risk plan with sensitive resources",
			planFile: "high_risk_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.Plan.ExpandableSections.AutoExpandDangerous = true
				c.ExpandAll = false
			},
			expectedOutputs: []string{
				"aws_rds_db_instance.legacy_mysql",
				"aws_rds_db_instance.old_postgres",
				"aws_iam_role",
			},
			minResourceCount: 5,
		},
		{
			name:     "Dependencies plan with complex relationships",
			planFile: "dependencies_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.Plan.ExpandableSections.ShowDependencies = true
				c.ExpandAll = false
			},
			expectedOutputs: []string{
				"aws_vpc.main",
				"aws_subnet.public",
				"aws_internet_gateway.main",
				"aws_nat_gateway.main",
			},
			minResourceCount: 10,
		},
		{
			name:     "Expand-all flag affects processing",
			planFile: "simple_plan.json",
			configOverride: func(c *config.Config) {
				c.Plan.ExpandableSections.Enabled = true
				c.ExpandAll = true // Global expand all
			},
			expectedOutputs: []string{
				"aws_instance.web_server",
				"aws_security_group.web_sg",
			},
			minResourceCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path to test plan
			planPath := filepath.Join("../../testdata", tt.planFile)

			// Create configuration
			cfg := getTestConfig()
			if tt.configOverride != nil {
				tt.configOverride(cfg)
			}

			// Create analyzer using the plan file path
			analyzer := NewAnalyzer(nil, cfg) // Plan will be loaded internally

			// Generate summary using the plan file path
			summary := analyzer.GenerateSummary(planPath)
			if summary == nil {
				t.Fatalf("Failed to generate summary from plan file %s", tt.planFile)
			}

			// Verify we have the expected number of resources
			if len(summary.ResourceChanges) < tt.minResourceCount {
				t.Errorf("Expected at least %d resources, got %d", tt.minResourceCount, len(summary.ResourceChanges))
			}

			// Verify expected resource addresses are present
			foundAddresses := make(map[string]bool)
			for _, change := range summary.ResourceChanges {
				foundAddresses[change.Address] = true
			}

			for _, expected := range tt.expectedOutputs {
				found := false
				for address := range foundAddresses {
					if strings.Contains(address, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find resource address containing '%s', but didn't find it in: %v", expected, getAddressesList(summary.ResourceChanges))
				}
			}

			// Test formatting functionality
			formatter := NewFormatter(cfg)

			// Test basic progressive disclosure formatting
			doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
			if err != nil {
				t.Fatalf("Failed to format summary with progressive disclosure: %v", err)
			}
			if doc == nil {
				t.Fatal("Expected non-nil document from progressive disclosure formatting")
			}

			// Test grouping if enabled and applicable
			if cfg.Plan.Grouping.Enabled && len(summary.ResourceChanges) >= cfg.Plan.Grouping.Threshold {
				groups := analyzer.groupByProvider(summary.ResourceChanges)
				if len(groups) > 1 {
					doc, err = formatter.formatGroupedWithCollapsibleSections(summary, groups)
					if err != nil {
						t.Fatalf("Failed to format grouped summary: %v", err)
					}
					if doc == nil {
						t.Fatal("Expected non-nil document from grouped formatting")
					}
				}
			}
		})
	}
}

// TestProviderGrouping_Integration tests provider grouping functionality end-to-end
func TestProviderGrouping_Integration(t *testing.T) {
	// Use multi-provider plan
	planPath := filepath.Join("../../testdata", "multi_provider_plan.json")

	// Test with grouping enabled
	cfg := getTestConfig()
	cfg.Plan.Grouping.Enabled = true
	cfg.Plan.Grouping.Threshold = 5
	cfg.Plan.ExpandableSections.Enabled = true

	// Create analyzer and generate summary
	analyzer := NewAnalyzer(nil, cfg)
	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		t.Fatalf("Failed to generate summary from plan")
	}

	// Test grouping logic
	groups := analyzer.groupByProvider(summary.ResourceChanges)

	// Verify we have multiple providers
	if len(groups) < 2 {
		t.Errorf("Expected at least 2 provider groups, got %d", len(groups))
	}

	// Verify specific providers are present
	expectedProviders := []string{"aws", "azurerm", "google", "kubernetes"}
	for _, provider := range expectedProviders {
		if _, exists := groups[provider]; !exists {
			t.Errorf("Expected provider '%s' to be present in groups, available: %v", provider, getMapKeys(groups))
		}
	}

	// Verify resources are correctly grouped
	totalGroupedResources := 0
	for _, resources := range groups {
		totalGroupedResources += len(resources)
	}

	if totalGroupedResources != len(summary.ResourceChanges) {
		t.Errorf("Expected all %d resources to be grouped, but only %d were grouped", len(summary.ResourceChanges), totalGroupedResources)
	}

	// Test grouped formatting
	formatter := NewFormatter(cfg)
	doc, err := formatter.formatGroupedWithCollapsibleSections(summary, groups)
	if err != nil {
		t.Fatalf("Failed to format grouped summary: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected non-nil document from grouped formatting")
	}
}

// TestCollapsibleFormatters_Integration tests collapsible formatters with real data
func TestCollapsibleFormatters_Integration(t *testing.T) {
	// Use plan with property changes and dependencies
	planPath := filepath.Join("../../testdata", "dependencies_plan.json")

	cfg := getTestConfig()
	cfg.Plan.ExpandableSections.Enabled = true
	cfg.Plan.ExpandableSections.ShowDependencies = true

	// Test with expand-all disabled
	cfg.ExpandAll = false
	analyzer := NewAnalyzer(nil, cfg)
	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		t.Fatalf("Failed to generate summary")
	}

	formatter := NewFormatter(cfg)
	doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
	if err != nil {
		t.Fatalf("Failed to format summary: %v", err)
	}
	if doc == nil {
		t.Fatal("Expected non-nil document from progressive disclosure formatting")
	}

	// Test with expand-all enabled
	cfg.ExpandAll = true
	formatter = NewFormatter(cfg) // Create new formatter with updated config
	doc, err = formatter.formatResourceChangesWithProgressiveDisclosure(summary)
	if err != nil {
		t.Fatalf("Failed to format summary with expand-all: %v", err)
	}
	if doc == nil {
		t.Fatal("Expected non-nil document from expand-all formatting")
	}
}

// TestRiskAssessment_Integration tests risk assessment with high-risk scenarios
func TestRiskAssessment_Integration(t *testing.T) {
	// Use high-risk plan
	planPath := filepath.Join("../../testdata", "high_risk_plan.json")

	cfg := getTestConfig()
	cfg.Plan.ExpandableSections.Enabled = true
	cfg.Plan.ExpandableSections.AutoExpandDangerous = true

	// Configure sensitive resources for risk assessment
	cfg.SensitiveResources = []config.SensitiveResource{
		{ResourceType: "aws_db_instance"},
		{ResourceType: "aws_rds_db_instance"},
		{ResourceType: "aws_s3_bucket"},
		{ResourceType: "aws_iam_role"},
	}

	cfg.SensitiveProperties = []config.SensitiveProperty{
		{ResourceType: "aws_instance", Property: "user_data"},
	}

	analyzer := NewAnalyzer(nil, cfg)
	summary := analyzer.GenerateSummary(planPath)
	if summary == nil {
		t.Fatalf("Failed to generate summary")
	}

	// Verify we have resources with dangerous changes
	hasDangerousChanges := false
	for _, change := range summary.ResourceChanges {
		if change.IsDangerous {
			hasDangerousChanges = true
			break
		}
	}

	if !hasDangerousChanges {
		t.Error("Expected to find dangerous changes in high-risk plan")
	}

	// Test formatting with risk assessment
	formatter := NewFormatter(cfg)
	doc, err := formatter.formatResourceChangesWithProgressiveDisclosure(summary)
	if err != nil {
		t.Fatalf("Failed to format high-risk summary: %v", err)
	}
	if doc == nil {
		t.Fatal("Expected non-nil document from risk assessment formatting")
	}
}

// TestErrorHandling_Integration tests graceful error handling
func TestErrorHandling_Integration(t *testing.T) {
	cfg := getTestConfig()
	analyzer := NewAnalyzer(nil, cfg)

	// Test with non-existent file
	summary := analyzer.GenerateSummary("non-existent-file.json")
	if summary != nil {
		t.Error("Expected nil summary for non-existent file")
	}

	// Test with empty file path
	summary = analyzer.GenerateSummary("")
	if summary != nil {
		t.Error("Expected nil summary for empty file path")
	}
}

// Helper function to get resource addresses as a list
func getAddressesList(changes []ResourceChange) []string {
	addresses := make([]string, len(changes))
	for i, change := range changes {
		addresses[i] = change.Address
	}
	return addresses
}

// Helper function to get map keys
func getMapKeys(m map[string][]ResourceChange) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Helper function to create test configuration
func getTestConfig() *config.Config {
	return &config.Config{
		ExpandAll: false,
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
				ShowDependencies:    true,
			},
			Grouping: config.GroupingConfig{
				Enabled:   true,
				Threshold: 10,
			},
			PerformanceLimits: config.PerformanceLimitsConfig{
				MaxPropertiesPerResource: 100,
				MaxPropertySize:          1024 * 1024,       // 1MB
				MaxTotalMemory:           100 * 1024 * 1024, // 100MB
				MaxDependencyDepth:       10,
			},
		},
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_db_instance"},
			{ResourceType: "aws_rds_db_instance"},
		},
		SensitiveProperties: []config.SensitiveProperty{
			{ResourceType: "aws_instance", Property: "user_data"},
		},
	}
}
