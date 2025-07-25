package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestPlanConfig_NewFieldsLoadFromYAML(t *testing.T) {
	tests := []struct {
		name         string
		yamlContent  map[string]any
		expectedPlan PlanConfig
	}{
		{
			name: "all new fields set",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"show-details":       true,
					"highlight-dangers":  true,
					"group-by-provider":  true,
					"grouping-threshold": 15,
					"show-context":       true,
				},
			},
			expectedPlan: PlanConfig{
				ShowDetails:       true,
				HighlightDangers:  true,
				GroupByProvider:   true,
				GroupingThreshold: 15,
				ShowContext:       true,
			},
		},
		{
			name: "new fields with defaults",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"show-details":       false,
					"highlight-dangers":  false,
					"group-by-provider":  false,
					"grouping-threshold": 0,
					"show-context":       false,
				},
			},
			expectedPlan: PlanConfig{
				ShowDetails:       false,
				HighlightDangers:  false,
				GroupByProvider:   false,
				GroupingThreshold: 0,
				ShowContext:       false,
			},
		},
		{
			name: "mixed existing and new fields",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"show-details":              true,
					"highlight-dangers":         true,
					"show-statistics-summary":   true,
					"statistics-summary-format": "table",
					"always-show-sensitive":     true,
					"group-by-provider":         true,
					"grouping-threshold":        10,
					"show-context":              false,
				},
			},
			expectedPlan: PlanConfig{
				ShowDetails:             true,
				HighlightDangers:        true,
				ShowStatisticsSummary:   true,
				StatisticsSummaryFormat: "table",
				AlwaysShowSensitive:     true,
				GroupByProvider:         true,
				GroupingThreshold:       10,
				ShowContext:             false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new viper instance for isolation
			v := viper.New()

			// Set the yaml content
			for key, value := range tt.yamlContent {
				v.Set(key, value)
			}

			// Unmarshal into config struct
			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			// Verify the plan configuration fields
			planConfig := config.Plan

			if planConfig.ShowDetails != tt.expectedPlan.ShowDetails {
				t.Errorf("ShowDetails = %v, expected %v", planConfig.ShowDetails, tt.expectedPlan.ShowDetails)
			}
			if planConfig.HighlightDangers != tt.expectedPlan.HighlightDangers {
				t.Errorf("HighlightDangers = %v, expected %v", planConfig.HighlightDangers, tt.expectedPlan.HighlightDangers)
			}
			if planConfig.ShowStatisticsSummary != tt.expectedPlan.ShowStatisticsSummary {
				t.Errorf("ShowStatisticsSummary = %v, expected %v", planConfig.ShowStatisticsSummary, tt.expectedPlan.ShowStatisticsSummary)
			}
			if planConfig.StatisticsSummaryFormat != tt.expectedPlan.StatisticsSummaryFormat {
				t.Errorf("StatisticsSummaryFormat = %v, expected %v", planConfig.StatisticsSummaryFormat, tt.expectedPlan.StatisticsSummaryFormat)
			}
			if planConfig.AlwaysShowSensitive != tt.expectedPlan.AlwaysShowSensitive {
				t.Errorf("AlwaysShowSensitive = %v, expected %v", planConfig.AlwaysShowSensitive, tt.expectedPlan.AlwaysShowSensitive)
			}

			// Test the new fields
			if planConfig.GroupByProvider != tt.expectedPlan.GroupByProvider {
				t.Errorf("GroupByProvider = %v, expected %v", planConfig.GroupByProvider, tt.expectedPlan.GroupByProvider)
			}
			if planConfig.GroupingThreshold != tt.expectedPlan.GroupingThreshold {
				t.Errorf("GroupingThreshold = %v, expected %v", planConfig.GroupingThreshold, tt.expectedPlan.GroupingThreshold)
			}
			if planConfig.ShowContext != tt.expectedPlan.ShowContext {
				t.Errorf("ShowContext = %v, expected %v", planConfig.ShowContext, tt.expectedPlan.ShowContext)
			}
		})
	}
}

func TestPlanConfig_DefaultValues(t *testing.T) {
	// Test that default values are properly handled when fields are not specified
	v := viper.New()

	// Set only some basic config without the new fields
	v.Set("plan", map[string]any{
		"show-details": true,
	})

	var config Config
	err := v.Unmarshal(&config)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	planConfig := config.Plan

	// Verify that new fields have their zero values when not specified
	if planConfig.GroupByProvider != false {
		t.Errorf("GroupByProvider should default to false, got %v", planConfig.GroupByProvider)
	}
	if planConfig.GroupingThreshold != 0 {
		t.Errorf("GroupingThreshold should default to 0, got %v", planConfig.GroupingThreshold)
	}
	if planConfig.ShowContext != false {
		t.Errorf("ShowContext should default to false, got %v", planConfig.ShowContext)
	}

	// Verify existing fields still work
	if planConfig.ShowDetails != true {
		t.Errorf("ShowDetails should be true as specified, got %v", planConfig.ShowDetails)
	}
}
