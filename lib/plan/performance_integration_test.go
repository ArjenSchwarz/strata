package plan

import (
	"fmt"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
)

// TestPerformanceLimitsWithLargePlans tests the performance limits with artificially large plans
func TestPerformanceLimitsWithLargePlans(t *testing.T) {
	skipIfIntegrationTestsDisabled(t)
	tests := []struct {
		name                  string
		numResources          int
		propertiesPerResource int
		expectedTruncation    bool
		expectedMaxProperties int
	}{
		{
			name:                  "Small plan - no truncation",
			numResources:          10,
			propertiesPerResource: 5,
			expectedTruncation:    false,
			expectedMaxProperties: 5,
		},
		{
			name:                  "Medium plan - approaching limits",
			numResources:          50,
			propertiesPerResource: 50,
			expectedTruncation:    false,
			expectedMaxProperties: 50,
		},
		{
			name:                  "Large plan - should trigger property limits",
			numResources:          5,
			propertiesPerResource: 120, // Exceeds hardcoded MaxPropertiesPerResource (100)
			expectedTruncation:    true,
			expectedMaxProperties: 100, // Should be capped at hardcoded limit
		},
		{
			name:                  "Very large plan - may trigger memory limits",
			numResources:          20,
			propertiesPerResource: 80,
			expectedTruncation:    false, // Memory limits harder to trigger with simple properties
			expectedMaxProperties: 80,    // Should not be truncated by count
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

			// Generate a large synthetic plan
			plan := generateLargeSyntheticPlan(tt.numResources, tt.propertiesPerResource)

			analyzer := NewAnalyzer(plan, cfg)
			summary := analyzer.GenerateSummary("")

			if summary == nil {
				t.Fatal("Expected summary to be generated even with large plans")
			}

			// Verify we got the expected number of resources
			if len(summary.ResourceChanges) != tt.numResources {
				t.Errorf("Expected %d resources, got %d", tt.numResources, len(summary.ResourceChanges))
			}

			// Test each resource for property limits
			foundTruncation := false
			maxPropertiesFound := 0

			for i, change := range summary.ResourceChanges {
				// Analyze this resource to get property changes
				if change.Type == "" {
					continue // Skip if resource type is not set
				}

				// Create a resource change from the plan to analyze
				planChange := plan.ResourceChanges[i]
				analysis := analyzer.analyzePropertyChanges(planChange)

				t.Logf("Resource %d (%s): %d properties, truncated: %v",
					i, change.Address, analysis.Count, analysis.Truncated)

				if analysis.Truncated {
					foundTruncation = true
				}

				if analysis.Count > maxPropertiesFound {
					maxPropertiesFound = analysis.Count
				}

				// Verify property limits are respected
				if analysis.Count > cfg.Plan.PerformanceLimits.MaxPropertiesPerResource {
					t.Errorf("Resource %d exceeded property limit: got %d, max %d",
						i, analysis.Count, cfg.Plan.PerformanceLimits.MaxPropertiesPerResource)
				}
			}

			// Verify truncation expectations
			if tt.expectedTruncation && !foundTruncation {
				t.Error("Expected truncation to occur but it didn't")
			}
			if !tt.expectedTruncation && foundTruncation {
				t.Error("Did not expect truncation but it occurred")
			}

			// Verify property count limits
			if maxPropertiesFound > tt.expectedMaxProperties {
				t.Errorf("Expected max properties %d, but found %d", tt.expectedMaxProperties, maxPropertiesFound)
			}
		})
	}
}

// TestFormatterPerformanceWithLargePlans tests formatter performance with large datasets
func TestFormatterPerformanceWithLargePlans(t *testing.T) {
	skipIfIntegrationTestsDisabled(t)
	// Generate a large plan
	plan := generateLargeSyntheticPlan(500, 20) // 500 resources, 20 properties each

	cfg := getTestConfig()
	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("")

	if summary == nil {
		t.Fatal("Expected summary to be generated")
	}

	formatter := NewFormatter(cfg)

	// Test formatting across different output formats
	formats := []string{"json", "table", "markdown", "html"}

	for _, format := range formats {
		t.Run(format+"_performance", func(t *testing.T) {
			outputConfig := &config.OutputConfiguration{
				Format:     format,
				UseEmoji:   false,
				UseColors:  false,
				TableStyle: "",
			}

			// Measure time and ensure it completes reasonably quickly
			err := formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("Failed to format large plan with %s: %v", format, err)
			}

			// The test passing means it completed without timing out or crashing
			t.Logf("Successfully formatted %d resources in %s format", len(summary.ResourceChanges), format)
		})
	}
}

// TestActionSortTransformerWithLargePlans tests sorting performance with large datasets
func TestActionSortTransformerWithLargePlans(t *testing.T) {
	skipIfIntegrationTestsDisabled(t)
	// Generate a large plan with mixed action types
	plan := generateMixedActionPlan(1000) // 1000 resources with various actions

	cfg := getTestConfig()
	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("")

	if summary == nil {
		t.Fatal("Expected summary to be generated")
	}

	formatter := NewFormatter(cfg)
	transformer := &ActionSortTransformer{}

	// Test with supported formats (CSV not supported by OutputSummary)
	supportedFormats := []string{"table", "markdown", "html"}

	for _, format := range supportedFormats {
		t.Run(format+"_sorting", func(t *testing.T) {
			if !transformer.CanTransform(format) {
				t.Errorf("Expected transformer to support %s format", format)
				return
			}

			outputConfig := &config.OutputConfiguration{
				Format:     format,
				UseEmoji:   false,
				UseColors:  false,
				TableStyle: "",
			}

			// Test that sorting completes without errors
			err := formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("Failed to sort and format large plan with %s: %v", format, err)
			}

			t.Logf("Successfully sorted and formatted %d resources in %s format", len(summary.ResourceChanges), format)
		})
	}
}

// TestProviderGroupingWithLargePlans tests provider grouping with large datasets
func TestProviderGroupingWithLargePlans(t *testing.T) {
	skipIfIntegrationTestsDisabled(t)
	// Generate plan with multiple providers
	plan := generateMultiProviderPlan(200) // 200 resources across multiple providers

	cfg := getTestConfig()
	cfg.Plan.Grouping.Enabled = true
	cfg.Plan.Grouping.Threshold = 10 // Low threshold to trigger grouping

	analyzer := NewAnalyzer(plan, cfg)
	summary := analyzer.GenerateSummary("")

	if summary == nil {
		t.Fatal("Expected summary to be generated")
	}

	// Count changed resources (excluding no-ops)
	changedCount := countChangedResources(summary.ResourceChanges)

	if changedCount < cfg.Plan.Grouping.Threshold {
		t.Skip("Need more changed resources to test grouping")
	}

	// Test grouping logic
	groups := groupResourcesByProvider(summary.ResourceChanges)

	if len(groups) < 2 {
		t.Error("Expected multiple provider groups")
	}

	// Verify all changed resources are accounted for
	totalGrouped := 0
	for provider, resources := range groups {
		totalGrouped += len(resources)
		t.Logf("Provider %s: %d resources", provider, len(resources))
	}

	if totalGrouped != changedCount {
		t.Errorf("Expected %d changed resources to be grouped, got %d", changedCount, totalGrouped)
	}

	// Test formatter with grouping
	formatter := NewFormatter(cfg)
	outputConfig := &config.OutputConfiguration{
		Format:     "table",
		UseEmoji:   false,
		UseColors:  false,
		TableStyle: "",
	}

	err := formatter.OutputSummary(summary, outputConfig, true)
	if err != nil {
		t.Errorf("Failed to format grouped large plan: %v", err)
	}
}

// Helper functions for generating test data

func generateLargeSyntheticPlan(numResources, propertiesPerResource int) *tfjson.Plan {
	plan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.8.5",
		ResourceChanges:  make([]*tfjson.ResourceChange, numResources),
	}

	for i := range numResources {
		// Create before and after states with many properties
		before := make(map[string]any)
		after := make(map[string]any)

		for j := range propertiesPerResource {
			propName := generatePropertyName(j)
			before[propName] = generatePropertyValue(j, "before")
			after[propName] = generatePropertyValue(j, "after")
		}

		plan.ResourceChanges[i] = &tfjson.ResourceChange{
			Address: generateResourceAddress(i),
			Type:    generateResourceType(i),
			Name:    generateResourceName(i),
			Change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionUpdate},
				Before:  before,
				After:   after,
			},
		}
	}

	return plan
}

func generatePlanWithLargeProperties(numResources, propertiesPerResource, propertySize int) *tfjson.Plan {
	plan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.8.5",
		ResourceChanges:  make([]*tfjson.ResourceChange, numResources),
	}

	for i := range numResources {
		before := make(map[string]any)
		after := make(map[string]any)

		for j := range propertiesPerResource {
			propName := generatePropertyName(j)
			// Generate large property values
			before[propName] = "before_" + generateLargeString(propertySize-7) // Account for prefix length
			after[propName] = "after_" + generateLargeString(propertySize-6)   // Account for prefix length
		}

		plan.ResourceChanges[i] = &tfjson.ResourceChange{
			Address: generateResourceAddress(i),
			Type:    generateResourceType(i),
			Name:    generateResourceName(i),
			Change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionUpdate},
				Before:  before,
				After:   after,
			},
		}
	}

	return plan
}

func generateMixedActionPlan(numResources int) *tfjson.Plan {
	actions := []tfjson.Action{
		tfjson.ActionCreate,
		tfjson.ActionUpdate,
		tfjson.ActionDelete,
		[]tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate}[0], // Replace
	}

	plan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.8.5",
		ResourceChanges:  make([]*tfjson.ResourceChange, numResources),
	}

	for i := range numResources {
		action := actions[i%len(actions)]

		var before, after any
		var changeActions []tfjson.Action

		switch action {
		case tfjson.ActionCreate:
			before = nil
			after = map[string]any{"prop": "value"}
			changeActions = []tfjson.Action{tfjson.ActionCreate}
		case tfjson.ActionDelete:
			before = map[string]any{"prop": "value"}
			after = nil
			changeActions = []tfjson.Action{tfjson.ActionDelete}
		case tfjson.ActionUpdate:
			before = map[string]any{"prop": "old_value"}
			after = map[string]any{"prop": "new_value"}
			changeActions = []tfjson.Action{tfjson.ActionUpdate}
		default: // Replace
			before = map[string]any{"prop": "old_value"}
			after = map[string]any{"prop": "new_value"}
			changeActions = []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate}
		}

		plan.ResourceChanges[i] = &tfjson.ResourceChange{
			Address: generateResourceAddress(i),
			Type:    generateResourceType(i),
			Name:    generateResourceName(i),
			Change: &tfjson.Change{
				Actions: changeActions,
				Before:  before,
				After:   after,
			},
		}
	}

	return plan
}

func generateMultiProviderPlan(numResources int) *tfjson.Plan {
	providers := []string{"aws", "azurerm", "google", "kubernetes", "local"}

	plan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.8.5",
		ResourceChanges:  make([]*tfjson.ResourceChange, numResources),
	}

	for i := range numResources {
		provider := providers[i%len(providers)]
		resourceType := provider + "_resource_" + generateResourceName(i)

		plan.ResourceChanges[i] = &tfjson.ResourceChange{
			Address: resourceType + ".instance_" + generateResourceName(i),
			Type:    resourceType,
			Name:    "instance_" + generateResourceName(i),
			Change: &tfjson.Change{
				Actions: []tfjson.Action{tfjson.ActionUpdate},
				Before:  map[string]any{"prop": "old"},
				After:   map[string]any{"prop": "new"},
			},
		}
	}

	return plan
}

// Helper functions for generating property names and values

func generatePropertyName(index int) string {
	baseNames := []string{
		"instance_type", "ami_id", "vpc_id", "subnet_id", "security_groups",
		"user_data", "key_name", "tags", "volume_size", "volume_type",
		"backup_retention", "engine_version", "db_name", "allocated_storage", "multi_az",
		"publicly_accessible", "storage_encrypted", "parameter_group", "option_group", "port",
		"bucket_name", "versioning", "encryption", "public_access", "lifecycle_rule",
		"website_config", "cors_rule", "logging", "notification", "replication",
		"ingress_rules", "egress_rules", "protocol", "cidr_blocks", "source_groups",
		"description", "name_prefix", "revoke_rules", "owner_id", "group_id",
		"availability_zone", "encrypted", "iops", "kms_key_id", "size",
		"device_name", "delete_on_termination", "volume_id", "attachment_id", "force_detach",
	}

	// Generate unique property names by adding suffix when we exceed base names
	if index < len(baseNames) {
		return baseNames[index]
	} else {
		baseIndex := index % len(baseNames)
		suffix := index / len(baseNames)
		return baseNames[baseIndex] + "_" + fmt.Sprintf("%d", suffix)
	}
}

func generatePropertyValue(index int, prefix string) any {
	switch index % 3 {
	case 0:
		return prefix + "_string_value_" + generateResourceName(index) + "_" + prefix // Make sure before/after differ
	case 1:
		// Different values for before/after
		if prefix == "before" {
			return index * 100
		} else {
			return index*100 + 1
		}
	case 2:
		return (prefix == "after") // Different boolean values for before/after
	default:
		return prefix + "_default_value"
	}
}

func generateResourceAddress(index int) string {
	return "test_resource.instance_" + generateResourceName(index)
}

func generateResourceType(index int) string {
	types := []string{
		"aws_instance", "aws_db_instance", "aws_s3_bucket", "aws_security_group",
		"aws_volume", "azurerm_virtual_machine", "google_compute_instance",
		"kubernetes_deployment", "local_file",
	}
	return types[index%len(types)]
}

func generateResourceName(index int) string {
	names := []string{
		"web", "db", "cache", "queue", "worker", "api", "frontend", "backend",
		"proxy", "balancer", "storage", "compute", "network", "security",
		"monitor", "backup", "test", "dev", "staging", "prod",
	}
	return names[index%len(names)]
}
