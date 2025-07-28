package config

import (
	"strings"
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

func TestConfig_ExpandAllFlag(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent map[string]any
		expected    bool
	}{
		{
			name: "expand_all set to true",
			yamlContent: map[string]any{
				"expand_all": true,
			},
			expected: true,
		},
		{
			name: "expand_all set to false",
			yamlContent: map[string]any{
				"expand_all": false,
			},
			expected: false,
		},
		{
			name:        "expand_all not specified",
			yamlContent: map[string]any{},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()

			for key, value := range tt.yamlContent {
				v.Set(key, value)
			}

			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			if config.ExpandAll != tt.expected {
				t.Errorf("ExpandAll = %v, expected %v", config.ExpandAll, tt.expected)
			}
		})
	}
}

func TestExpandableSectionsConfig_Loading(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent map[string]any
		expectedESC ExpandableSectionsConfig
	}{
		{
			name: "all expandable sections fields set",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"expandable_sections": map[string]any{
						"enabled":               true,
						"auto_expand_dangerous": true,
						"show_dependencies":     true,
					},
				},
			},
			expectedESC: ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
				ShowDependencies:    true,
			},
		},
		{
			name: "partial expandable sections config",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"expandable_sections": map[string]any{
						"enabled": false,
					},
				},
			},
			expectedESC: ExpandableSectionsConfig{
				Enabled:             false,
				AutoExpandDangerous: false,
				ShowDependencies:    false,
			},
		},
		{
			name:        "no expandable sections config",
			yamlContent: map[string]any{},
			expectedESC: ExpandableSectionsConfig{
				Enabled:             false,
				AutoExpandDangerous: false,
				ShowDependencies:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()

			for key, value := range tt.yamlContent {
				v.Set(key, value)
			}

			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			esc := config.Plan.ExpandableSections
			if esc.Enabled != tt.expectedESC.Enabled {
				t.Errorf("ExpandableSections.Enabled = %v, expected %v", esc.Enabled, tt.expectedESC.Enabled)
			}
			if esc.AutoExpandDangerous != tt.expectedESC.AutoExpandDangerous {
				t.Errorf("ExpandableSections.AutoExpandDangerous = %v, expected %v", esc.AutoExpandDangerous, tt.expectedESC.AutoExpandDangerous)
			}
			if esc.ShowDependencies != tt.expectedESC.ShowDependencies {
				t.Errorf("ExpandableSections.ShowDependencies = %v, expected %v", esc.ShowDependencies, tt.expectedESC.ShowDependencies)
			}
		})
	}
}

func TestGetPerformanceLimitsWithDefaults(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent map[string]any
		expected    PerformanceLimitsConfig
	}{
		{
			name: "all performance limits specified",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"performance_limits": map[string]any{
						"max_properties_per_resource": 50,
						"max_property_size":           524288,   // 512KB
						"max_total_memory":            52428800, // 50MB
						"max_dependency_depth":        5,
						"max_resources_per_group":     500,
					},
				},
			},
			expected: PerformanceLimitsConfig{
				MaxPropertiesPerResource: 50,
				MaxPropertySize:          524288,
				MaxTotalMemory:           52428800,
				MaxDependencyDepth:       5,
				MaxResourcesPerGroup:     500,
			},
		},
		{
			name:        "no performance limits - defaults applied",
			yamlContent: map[string]any{},
			expected: PerformanceLimitsConfig{
				MaxPropertiesPerResource: 100,
				MaxPropertySize:          1048576,   // 1MB
				MaxTotalMemory:           104857600, // 100MB
				MaxDependencyDepth:       10,
				MaxResourcesPerGroup:     1000,
			},
		},
		{
			name: "partial performance limits - defaults for missing",
			yamlContent: map[string]any{
				"plan": map[string]any{
					"performance_limits": map[string]any{
						"max_properties_per_resource": 200,
						"max_total_memory":            209715200, // 200MB
					},
				},
			},
			expected: PerformanceLimitsConfig{
				MaxPropertiesPerResource: 200,
				MaxPropertySize:          1048576,   // Default 1MB
				MaxTotalMemory:           209715200, // Specified 200MB
				MaxDependencyDepth:       10,        // Default 10
				MaxResourcesPerGroup:     1000,      // Default 1000
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()

			for key, value := range tt.yamlContent {
				v.Set(key, value)
			}

			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			limits := config.GetPerformanceLimitsWithDefaults()

			if limits.MaxPropertiesPerResource != tt.expected.MaxPropertiesPerResource {
				t.Errorf("MaxPropertiesPerResource = %v, expected %v", limits.MaxPropertiesPerResource, tt.expected.MaxPropertiesPerResource)
			}
			if limits.MaxPropertySize != tt.expected.MaxPropertySize {
				t.Errorf("MaxPropertySize = %v, expected %v", limits.MaxPropertySize, tt.expected.MaxPropertySize)
			}
			if limits.MaxTotalMemory != tt.expected.MaxTotalMemory {
				t.Errorf("MaxTotalMemory = %v, expected %v", limits.MaxTotalMemory, tt.expected.MaxTotalMemory)
			}
			if limits.MaxDependencyDepth != tt.expected.MaxDependencyDepth {
				t.Errorf("MaxDependencyDepth = %v, expected %v", limits.MaxDependencyDepth, tt.expected.MaxDependencyDepth)
			}
			if limits.MaxResourcesPerGroup != tt.expected.MaxResourcesPerGroup {
				t.Errorf("MaxResourcesPerGroup = %v, expected %v", limits.MaxResourcesPerGroup, tt.expected.MaxResourcesPerGroup)
			}
		})
	}
}

func TestValidateConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: Config{
				Plan: PlanConfig{
					Grouping: GroupingConfig{
						Enabled:   true,
						Threshold: 10,
					},
					PerformanceLimits: PerformanceLimitsConfig{
						MaxPropertiesPerResource: 100,
						MaxPropertySize:          1048576,
						MaxTotalMemory:           104857600,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid grouping threshold",
			config: Config{
				Plan: PlanConfig{
					Grouping: GroupingConfig{
						Enabled:   true,
						Threshold: 0,
					},
				},
			},
			expectError: true,
			errorMsg:    "plan.grouping.threshold must be at least 1",
		},
		{
			name: "invalid max properties per resource",
			config: Config{
				Plan: PlanConfig{
					Grouping: GroupingConfig{
						Enabled:   true,
						Threshold: 10,
					},
					PerformanceLimits: PerformanceLimitsConfig{
						MaxPropertiesPerResource: -1,
					},
				},
			},
			expectError: true,
			errorMsg:    "plan.performance_limits.max_properties_per_resource must be positive",
		},
		{
			name: "invalid max property size",
			config: Config{
				Plan: PlanConfig{
					Grouping: GroupingConfig{
						Enabled:   true,
						Threshold: 10,
					},
					PerformanceLimits: PerformanceLimitsConfig{
						MaxPropertiesPerResource: 100,
						MaxPropertySize:          500, // Less than 1024
					},
				},
			},
			expectError: true,
			errorMsg:    "plan.performance_limits.max_property_size must be at least 1024 bytes",
		},
		{
			name: "invalid max total memory",
			config: Config{
				Plan: PlanConfig{
					Grouping: GroupingConfig{
						Enabled:   true,
						Threshold: 10,
					},
					PerformanceLimits: PerformanceLimitsConfig{
						MaxPropertiesPerResource: 100,
						MaxPropertySize:          1048576,
						MaxTotalMemory:           500000, // Less than 1MB
					},
				},
			},
			expectError: true,
			errorMsg:    "plan.performance_limits.max_total_memory must be at least 1MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateConfiguration()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	// Test that all required fields have sensible defaults
	if config.ExpandAll != false {
		t.Errorf("Expected ExpandAll to be false, got %v", config.ExpandAll)
	}

	if !config.Plan.ShowDetails {
		t.Errorf("Expected ShowDetails to be true")
	}

	if !config.Plan.HighlightDangers {
		t.Errorf("Expected HighlightDangers to be true")
	}

	if !config.Plan.ExpandableSections.Enabled {
		t.Errorf("Expected ExpandableSections.Enabled to be true")
	}

	if !config.Plan.ExpandableSections.AutoExpandDangerous {
		t.Errorf("Expected ExpandableSections.AutoExpandDangerous to be true")
	}

	if !config.Plan.Grouping.Enabled {
		t.Errorf("Expected Grouping.Enabled to be true")
	}

	if config.Plan.Grouping.Threshold != 10 {
		t.Errorf("Expected Grouping.Threshold to be 10, got %d", config.Plan.Grouping.Threshold)
	}

	if config.Plan.PerformanceLimits.MaxPropertiesPerResource != 100 {
		t.Errorf("Expected MaxPropertiesPerResource to be 100, got %d", config.Plan.PerformanceLimits.MaxPropertiesPerResource)
	}

	// Test that validation passes for default config
	if err := config.ValidateConfiguration(); err != nil {
		t.Errorf("Default config should be valid, got error: %v", err)
	}
}
