package config

import (
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
	GroupByProvider   bool `mapstructure:"group-by-provider"`  // Enable provider grouping
	GroupingThreshold int  `mapstructure:"grouping-threshold"` // Minimum resources to trigger grouping
	ShowContext       bool `mapstructure:"show-context"`       // Show property changes
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
