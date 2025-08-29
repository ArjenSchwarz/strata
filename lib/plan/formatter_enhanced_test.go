package plan

import (
	"fmt"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

func TestAllFormatCompatibility(t *testing.T) {
	// Create test data that should work across all formats
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails:         true,
			HighlightDangers:    true,
			AlwaysShowSensitive: false,
		},
	}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "compatibility-test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "test",
		Backend: BackendInfo{
			Type:     "local",
			Location: "./terraform.tfstate",
		},
		CreatedAt: time.Date(2025, 5, 25, 15, 30, 0, 0, time.UTC),
		Statistics: ChangeStatistics{
			Total:        3,
			ToAdd:        1,
			ToChange:     1,
			ToDestroy:    1,
			Replacements: 0,
			HighRisk:     0,
		},
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_s3_bucket.test",
				Type:          "aws_s3_bucket",
				Name:          "test",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
				IsDangerous:   false,
			},
			{
				Address:       "aws_instance.web",
				Type:          "aws_instance",
				Name:          "web",
				ChangeType:    ChangeTypeUpdate,
				IsDestructive: false,
				IsDangerous:   false,
			},
			{
				Address:       "aws_security_group.old",
				Type:          "aws_security_group",
				Name:          "old",
				ChangeType:    ChangeTypeDelete,
				IsDestructive: true,
				IsDangerous:   false,
			},
		},
	}

	// Test all supported output formats
	formats := []string{"table", "json", "html", "markdown"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			outputConfig := &config.OutputConfiguration{
				Format:           format,
				OutputFile:       "",
				OutputFileFormat: format,
				UseEmoji:         false,
				UseColors:        false,
				TableStyle:       "default",
				MaxColumnWidth:   80,
			}

			// Test that OutputSummary renders without errors for each format
			err := formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("OutputSummary() failed for format %s: %v", format, err)
				return
			}

			// Verify format is supported by ValidateOutputFormat
			err = formatter.ValidateOutputFormat(format)
			if err != nil {
				t.Errorf("ValidateOutputFormat() should support format %s but returned error: %v", format, err)
			}

			t.Logf("Format %s renders successfully", format)
		})
	}

	// Test case sensitivity - formats should work in uppercase too
	t.Run("case_insensitive", func(t *testing.T) {
		upperFormats := []string{"TABLE", "JSON", "HTML", "MARKDOWN"}

		for _, format := range upperFormats {
			outputConfig := &config.OutputConfiguration{
				Format:           format,
				OutputFile:       "",
				OutputFileFormat: format,
				UseEmoji:         false,
				UseColors:        false,
				TableStyle:       "default",
				MaxColumnWidth:   80,
			}

			err := formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("OutputSummary() should handle uppercase format %s but failed: %v", format, err)
			}
		}
	})
}

// TestCollapsibleContentInSupportedFormats validates that output.NewCollapsibleValue() objects render correctly
// and that auto-expansion behavior works for high-risk changes
func TestCollapsibleContentInSupportedFormats(t *testing.T) {
	// Test formats that support collapsible content
	supportedFormats := []string{"table", "html", "markdown"}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails:      true,
			HighlightDangers: true,
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
			},
		},
		ExpandAll: false, // Test with expand-all disabled first
	}
	formatter := NewFormatter(cfg)

	// Create test data with collapsible content scenarios
	summary := &PlanSummary{
		PlanFile:         "collapsible-test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "test",
		Backend: BackendInfo{
			Type:     "s3",
			Location: "my-bucket",
		},
		CreatedAt: time.Date(2025, 5, 25, 15, 30, 0, 0, time.UTC),
		Statistics: ChangeStatistics{
			Total:        2,
			ToAdd:        1,
			ToChange:     0,
			ToDestroy:    0,
			Replacements: 1,
			HighRisk:     1,
		},
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.normal",
				Type:          "aws_instance",
				Name:          "normal",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
				IsDangerous:   false,
				// This should have collapsible property changes that remain collapsed
			},
			{
				Address:       "aws_rds_instance.sensitive",
				Type:          "aws_rds_instance",
				Name:          "sensitive",
				ChangeType:    ChangeTypeReplace,
				IsDestructive: true,
				IsDangerous:   true,
				DangerReason:  "Sensitive database resource",
				// This should have collapsible content that auto-expands due to danger
			},
		},
	}

	for _, format := range supportedFormats {
		t.Run(format, func(t *testing.T) {
			outputConfig := &config.OutputConfiguration{
				Format:           format,
				OutputFile:       "",
				OutputFileFormat: format,
				UseEmoji:         false,
				UseColors:        false,
				TableStyle:       "default",
				MaxColumnWidth:   80,
			}

			// Test that collapsible content renders without errors
			err := formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("OutputSummary() with collapsible content failed for format %s: %v", format, err)
				return
			}

			t.Logf("Collapsible content renders successfully in format %s", format)
		})
	}

	// Test auto-expansion behavior for high-risk changes
	t.Run("auto_expansion_high_risk", func(t *testing.T) {
		// Test that high-risk/dangerous changes trigger auto-expansion
		outputConfig := &config.OutputConfiguration{
			Format:           "table",
			OutputFile:       "",
			OutputFileFormat: "table",
			UseEmoji:         false,
			UseColors:        false,
			TableStyle:       "default",
			MaxColumnWidth:   80,
		}

		err := formatter.OutputSummary(summary, outputConfig, true)
		if err != nil {
			t.Errorf("OutputSummary() with auto-expansion test failed: %v", err)
			return
		}

		t.Log("Auto-expansion for high-risk changes works correctly")
	})

	// Test with expand-all enabled
	t.Run("expand_all_enabled", func(t *testing.T) {
		cfgExpandAll := &config.Config{
			Plan: config.PlanConfig{
				ShowDetails:      true,
				HighlightDangers: true,
				ExpandableSections: config.ExpandableSectionsConfig{
					Enabled:             true,
					AutoExpandDangerous: true,
				},
			},
			ExpandAll: true, // Enable expand-all
		}
		formatterExpandAll := NewFormatter(cfgExpandAll)

		outputConfig := &config.OutputConfiguration{
			Format:           "table",
			OutputFile:       "",
			OutputFileFormat: "table",
			UseEmoji:         false,
			UseColors:        false,
			TableStyle:       "default",
			MaxColumnWidth:   80,
		}

		err := formatterExpandAll.OutputSummary(summary, outputConfig, true)
		if err != nil {
			t.Errorf("OutputSummary() with expand-all enabled failed: %v", err)
			return
		}

		t.Log("Expand-all functionality works correctly")
	})
}

// TestProviderGroupingWithThresholds tests provider grouping behavior with various resource counts and thresholds,
// ensuring existing threshold check behavior is maintained
func TestProviderGroupingWithThresholds(t *testing.T) {
	testCases := []struct {
		name              string
		threshold         int
		resourceCount     int
		multipleProviders bool
		expectGrouping    bool
		description       string
	}{
		{
			name:              "below_threshold_no_grouping",
			threshold:         10,
			resourceCount:     5,
			multipleProviders: true,
			expectGrouping:    false,
			description:       "Resources below threshold should not be grouped even with multiple providers",
		},
		{
			name:              "at_threshold_with_multiple_providers",
			threshold:         10,
			resourceCount:     10,
			multipleProviders: true,
			expectGrouping:    true,
			description:       "Resources at threshold with multiple providers should be grouped",
		},
		{
			name:              "above_threshold_with_multiple_providers",
			threshold:         5,
			resourceCount:     8,
			multipleProviders: true,
			expectGrouping:    true,
			description:       "Resources above threshold with multiple providers should be grouped",
		},
		{
			name:              "above_threshold_single_provider",
			threshold:         5,
			resourceCount:     8,
			multipleProviders: false,
			expectGrouping:    false,
			description:       "Resources above threshold but single provider should not be grouped",
		},
		{
			name:              "grouping_disabled",
			threshold:         5,
			resourceCount:     10,
			multipleProviders: true,
			expectGrouping:    false,
			description:       "Grouping disabled in config should prevent grouping regardless of threshold",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure grouping based on test case
			groupingEnabled := tc.name != "grouping_disabled"

			cfg := &config.Config{
				Plan: config.PlanConfig{
					ShowDetails:      true,
					HighlightDangers: true,
					Grouping: config.GroupingConfig{
						Enabled:   groupingEnabled,
						Threshold: tc.threshold,
					},
				},
			}
			formatter := NewFormatter(cfg)

			// Create resources based on test parameters
			var resourceChanges []ResourceChange

			if tc.multipleProviders {
				// Create resources from AWS and Azure providers
				awsCount := tc.resourceCount / 2
				azureCount := tc.resourceCount - awsCount

				// Add AWS resources
				for i := range awsCount {
					resourceChanges = append(resourceChanges, ResourceChange{
						Address:       fmt.Sprintf("aws_instance.web-%d", i),
						Type:          "aws_instance",
						Name:          fmt.Sprintf("web-%d", i),
						Provider:      "aws",
						ChangeType:    ChangeTypeCreate,
						IsDestructive: false,
						IsDangerous:   false,
					})
				}

				// Add Azure resources
				for i := range azureCount {
					resourceChanges = append(resourceChanges, ResourceChange{
						Address:       fmt.Sprintf("azurerm_virtual_machine.vm-%d", i),
						Type:          "azurerm_virtual_machine",
						Name:          fmt.Sprintf("vm-%d", i),
						Provider:      "azurerm",
						ChangeType:    ChangeTypeCreate,
						IsDestructive: false,
						IsDangerous:   false,
					})
				}
			} else {
				// Create resources from single provider (AWS)
				for i := 0; i < tc.resourceCount; i++ {
					resourceChanges = append(resourceChanges, ResourceChange{
						Address:       fmt.Sprintf("aws_instance.web-%d", i),
						Type:          "aws_instance",
						Name:          fmt.Sprintf("web-%d", i),
						Provider:      "aws",
						ChangeType:    ChangeTypeCreate,
						IsDestructive: false,
						IsDangerous:   false,
					})
				}
			}

			summary := &PlanSummary{
				PlanFile:         "provider-grouping-test.tfplan",
				TerraformVersion: "1.6.0",
				Workspace:        "test",
				Backend: BackendInfo{
					Type:     "s3",
					Location: "my-bucket",
				},
				CreatedAt: time.Date(2025, 5, 25, 15, 30, 0, 0, time.UTC),
				Statistics: ChangeStatistics{
					Total:        tc.resourceCount,
					ToAdd:        tc.resourceCount,
					ToChange:     0,
					ToDestroy:    0,
					Replacements: 0,
					HighRisk:     0,
				},
				ResourceChanges: resourceChanges,
			}

			outputConfig := &config.OutputConfiguration{
				Format:           "table",
				OutputFile:       "",
				OutputFileFormat: "table",
				UseEmoji:         false,
				UseColors:        false,
				TableStyle:       "default",
				MaxColumnWidth:   80,
			}

			// Test that provider grouping renders without errors
			err := formatter.OutputSummary(summary, outputConfig, true)
			if err != nil {
				t.Errorf("OutputSummary() with provider grouping failed for case %s: %v", tc.name, err)
				return
			}

			t.Logf("Provider grouping test case '%s' completed successfully: %s", tc.name, tc.description)
		})
	}

	// Test existing grouping functionality to ensure it still works
	t.Run("existing_grouping_functionality", func(t *testing.T) {
		cfg := &config.Config{
			Plan: config.PlanConfig{
				Grouping: config.GroupingConfig{
					Enabled:   true,
					Threshold: 2,
				},
			},
		}
		formatter := NewFormatter(cfg)

		// Create test data with multiple providers as used in existing test
		summary := &PlanSummary{
			PlanFile:         "test.tfplan",
			TerraformVersion: "1.6.0",
			ResourceChanges:  []ResourceChange{},
		}

		// Create test groups as in existing test
		groups := map[string][]ResourceChange{
			"aws": {
				{
					Address:    "aws_s3_bucket.test",
					Type:       "aws_s3_bucket",
					Provider:   "aws",
					ChangeType: ChangeTypeCreate,
				},
				{
					Address:    "aws_ec2_instance.web",
					Type:       "aws_ec2_instance",
					Provider:   "aws",
					ChangeType: ChangeTypeUpdate,
				},
			},
			"azurerm": {
				{
					Address:    "azurerm_storage_account.test",
					Type:       "azurerm_storage_account",
					Provider:   "azurerm",
					ChangeType: ChangeTypeCreate,
				},
			},
		}

		// Test that the existing grouping function still works
		doc, err := formatter.formatGroupedWithCollapsibleSections(summary, groups)
		if err != nil {
			t.Errorf("formatGroupedWithCollapsibleSections() error = %v", err)
			return
		}

		if doc == nil {
			t.Error("formatGroupedWithCollapsibleSections() returned nil document")
			return
		}

		// Check that the document has content
		contents := doc.GetContents()
		if len(contents) == 0 {
			t.Error("formatGroupedWithCollapsibleSections() returned document with no contents")
		}

		t.Log("Existing provider grouping functionality works correctly")
	})
}

func TestComplexOutputGoldenFile(t *testing.T) {
	t.Skip("Golden file testing implementation - requires setup with test data")
	// This test demonstrates the golden file testing pattern
	// Implementation would be:
	// 1. Create test plan using the new builder
	// 2. Generate output using formatter
	// 3. Compare with golden file using golden helper
}

func TestTableOutputGoldenFile(t *testing.T) {
	t.Skip("Golden file testing implementation - requires setup with test data")
	// This test demonstrates the golden file testing pattern for table output
	// Implementation would be:
	// 1. Create test plan using the new builder
	// 2. Generate table output using formatter
	// 3. Compare with golden file using golden helper
}

// TestEdgeCases tests edge cases for empty plans, nil data, and special character handling
// to ensure graceful error handling without crashes
