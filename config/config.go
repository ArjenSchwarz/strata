package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	markdownFormat = "markdown"
)

// SensitiveResource defines a resource type that should be flagged as sensitive
type SensitiveResource struct {
	ResourceType string `mapstructure:"resource_type"`
}

// SensitiveProperty defines a resource type and property combination that should be flagged as sensitive
type SensitiveProperty struct {
	ResourceType string `mapstructure:"resource_type"`
	Property     string `mapstructure:"property"`
}

// TableConfig holds configuration specific to table output
type TableConfig struct {
	Style          string `mapstructure:"style"`
	MaxColumnWidth int    `mapstructure:"max-column-width"`
}

// Config holds the global configuration settings
type Config struct {
	// Global expand control for collapsible sections
	ExpandAll bool `mapstructure:"expand_all"`

	// Plan-specific configuration
	Plan PlanConfig `mapstructure:"plan"`

	// Table-specific configuration
	Table TableConfig `mapstructure:"table"`

	// Sensitive resources and properties configuration
	SensitiveResources  []SensitiveResource `mapstructure:"sensitive_resources"`
	SensitiveProperties []SensitiveProperty `mapstructure:"sensitive_properties"`
}

// PlanConfig holds configuration specific to plan operations
type PlanConfig struct {
	ShowDetails             bool   `mapstructure:"show-details"`
	HighlightDangers        bool   `mapstructure:"highlight-dangers"`
	ShowStatisticsSummary   bool   `mapstructure:"show-statistics-summary"`
	StatisticsSummaryFormat string `mapstructure:"statistics-summary-format"`
	AlwaysShowSensitive     bool   `mapstructure:"always-show-sensitive"` // Always show sensitive resources even when details are disabled
	// Enhanced summary visualization fields
	GroupByProvider    bool                     `mapstructure:"group-by-provider"`   // Enable provider grouping
	GroupingThreshold  int                      `mapstructure:"grouping-threshold"`  // Minimum resources to trigger grouping
	ShowContext        bool                     `mapstructure:"show-context"`        // Show property changes
	ExpandableSections ExpandableSectionsConfig `mapstructure:"expandable_sections"` // Collapsible sections configuration
	Grouping           GroupingConfig           `mapstructure:"grouping"`            // Enhanced grouping configuration
	PerformanceLimits  PerformanceLimitsConfig  `mapstructure:"performance_limits"`  // Performance and memory limits
}

// GetLCString returns a lowercase string value for the given setting
func (config *Config) GetLCString(setting string) string {
	if viper.IsSet(setting) {
		return strings.ToLower(viper.GetString(setting))
	}
	return ""
}

// GetString returns a string value for the given setting
func (config *Config) GetString(setting string) string {
	if viper.IsSet(setting) {
		return viper.GetString(setting)
	}
	return ""
}

// GetInt returns an integer value for the given setting
func (config *Config) GetInt(setting string) int {
	if viper.IsSet(setting) {
		return viper.GetInt(setting)
	}
	return 0
}

// OutputConfiguration holds the v2 output configuration settings
type OutputConfiguration struct {
	Format           string
	OutputFile       string
	OutputFileFormat string
	UseEmoji         bool
	UseColors        bool
	TableStyle       string
	MaxColumnWidth   int
}

// NewOutputConfiguration creates a new output configuration from the global config
func (config *Config) NewOutputConfiguration() *OutputConfiguration {
	format := config.GetLCString("output")
	outputFile := config.GetLCString("output-file")
	outputFileFormat := config.GetLCString("output-file-format")

	// Apply placeholder resolution to file path if specified
	if outputFile != "" {
		outputFile = config.resolvePlaceholders(outputFile)
	}

	// Default file format to stdout format if not specified
	if outputFileFormat == "" {
		outputFileFormat = format
	}

	// Configure colors based on output format
	useColors := format != markdownFormat

	return &OutputConfiguration{
		Format:           format,
		OutputFile:       outputFile,
		OutputFileFormat: outputFileFormat,
		UseEmoji:         true,
		UseColors:        useColors,
		TableStyle:       config.GetString("table.style"),
		MaxColumnWidth:   config.GetInt("table.max-column-width"),
	}
}

// resolvePlaceholders replaces placeholder values in the given string with actual values
func (config *Config) resolvePlaceholders(value string) string {
	replacements := map[string]string{
		"$TIMESTAMP":     time.Now().Format("2006-01-02T15-04-05"),
		"$AWS_REGION":    config.getAWSRegion(),
		"$AWS_ACCOUNTID": config.getAWSAccountID(),
	}

	result := value
	for placeholder, replacement := range replacements {
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}

// getAWSRegion returns the AWS region from environment variables or config
func (config *Config) getAWSRegion() string {
	// Try environment variable first
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	// Try config setting
	if region := config.GetString("region"); region != "" {
		return region
	}
	return "unknown"
}

// getAWSAccountID returns the AWS account ID from environment variables or config
func (config *Config) getAWSAccountID() string {
	// Try environment variable first
	if accountID := os.Getenv("AWS_ACCOUNT_ID"); accountID != "" {
		return accountID
	}
	// Try config setting
	if accountID := config.GetString("account-id"); accountID != "" {
		return accountID
	}
	return "unknown"
}

// ExpandableSectionsConfig controls collapsible sections behavior
type ExpandableSectionsConfig struct {
	Enabled             bool `mapstructure:"enabled"`               // Enable collapsible sections
	AutoExpandDangerous bool `mapstructure:"auto_expand_dangerous"` // Auto-expand high-risk sections
	ShowDependencies    bool `mapstructure:"show_dependencies"`     // Show dependency sections
}

// GroupingConfig controls enhanced grouping behavior
type GroupingConfig struct {
	Enabled   bool `mapstructure:"enabled"`   // Enable provider grouping
	Threshold int  `mapstructure:"threshold"` // Minimum resources to trigger grouping
}

// PerformanceLimitsConfig defines memory and processing limits for analysis
type PerformanceLimitsConfig struct {
	MaxPropertiesPerResource int   `mapstructure:"max_properties_per_resource"` // Default: 100
	MaxPropertySize          int   `mapstructure:"max_property_size"`           // Default: 1MB (1048576 bytes)
	MaxTotalMemory           int64 `mapstructure:"max_total_memory"`            // Default: 100MB (104857600 bytes)
	MaxDependencyDepth       int   `mapstructure:"max_dependency_depth"`        // Default: 10
	MaxResourcesPerGroup     int   `mapstructure:"max_resources_per_group"`     // Default: 1000
}

// GetPerformanceLimitsWithDefaults returns performance limits with default values applied
func (config *Config) GetPerformanceLimitsWithDefaults() PerformanceLimitsConfig {
	limits := config.Plan.PerformanceLimits

	// Apply defaults for zero values
	if limits.MaxPropertiesPerResource == 0 {
		limits.MaxPropertiesPerResource = 100
	}
	if limits.MaxPropertySize == 0 {
		limits.MaxPropertySize = 1048576 // 1MB
	}
	if limits.MaxTotalMemory == 0 {
		limits.MaxTotalMemory = 104857600 // 100MB
	}
	if limits.MaxDependencyDepth == 0 {
		limits.MaxDependencyDepth = 10
	}
	if limits.MaxResourcesPerGroup == 0 {
		limits.MaxResourcesPerGroup = 1000
	}

	return limits
}

// MigrateDeprecatedConfig handles migration from old configuration format to new
func (config *Config) MigrateDeprecatedConfig() []string {
	var warnings []string

	// Set default values for new configuration sections if not present
	if !viper.IsSet("plan.expandable_sections") {
		config.Plan.ExpandableSections = ExpandableSectionsConfig{
			Enabled:             true,
			AutoExpandDangerous: true,
			ShowDependencies:    true,
		}
	}

	if !viper.IsSet("plan.grouping") {
		// Use existing threshold value if it was migrated, otherwise default to 10
		threshold := 10
		if config.Plan.GroupingThreshold > 0 {
			threshold = config.Plan.GroupingThreshold
		}

		config.Plan.Grouping = GroupingConfig{
			Enabled:   true,
			Threshold: threshold,
		}
	}

	// Remove the expand_all warning as it's not a deprecated feature but a new one
	// Users don't need to be warned about features they haven't configured yet

	return warnings
}

// ValidateConfiguration checks for invalid configuration combinations
func (config *Config) ValidateConfiguration() error {
	// Validate grouping threshold
	if config.Plan.Grouping.Threshold < 1 {
		return fmt.Errorf("plan.grouping.threshold must be at least 1, got %d", config.Plan.Grouping.Threshold)
	}

	// Validate performance limits
	limits := config.Plan.PerformanceLimits
	if limits.MaxPropertiesPerResource < 1 && limits.MaxPropertiesPerResource != 0 {
		return fmt.Errorf("plan.performance_limits.max_properties_per_resource must be positive, got %d", limits.MaxPropertiesPerResource)
	}
	if limits.MaxPropertySize < 1024 && limits.MaxPropertySize != 0 {
		return fmt.Errorf("plan.performance_limits.max_property_size must be at least 1024 bytes, got %d", limits.MaxPropertySize)
	}
	if limits.MaxTotalMemory < 1048576 && limits.MaxTotalMemory != 0 {
		return fmt.Errorf("plan.performance_limits.max_total_memory must be at least 1MB, got %d", limits.MaxTotalMemory)
	}

	return nil
}

// PrintDeprecationWarnings prints deprecation warnings to stderr
func PrintDeprecationWarnings(warnings []string) {
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Configuration Deprecation Warnings:\n")
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "   %s\n", warning)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// GetDefaultConfig returns a config with sensible defaults
func GetDefaultConfig() *Config {
	return &Config{
		ExpandAll: false,
		Plan: PlanConfig{
			ShowDetails:             true,
			HighlightDangers:        true,
			ShowStatisticsSummary:   true,
			StatisticsSummaryFormat: "horizontal",
			AlwaysShowSensitive:     true,
			ExpandableSections: ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
				ShowDependencies:    true,
			},
			Grouping: GroupingConfig{
				Enabled:   true,
				Threshold: 10,
			},
			PerformanceLimits: PerformanceLimitsConfig{
				MaxPropertiesPerResource: 100,
				MaxPropertySize:          1048576,   // 1MB
				MaxTotalMemory:           104857600, // 100MB
				MaxDependencyDepth:       10,
				MaxResourcesPerGroup:     1000,
			},
		},
		Table: TableConfig{
			Style:          "ColoredBlackOnMagentaWhite",
			MaxColumnWidth: 50,
		},
	}
}
