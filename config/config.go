package config

import (
	"os"
	"strings"
	"time"

	format "github.com/ArjenSchwarz/go-output"
	"github.com/spf13/viper"
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

// Config holds the global configuration settings
type Config struct {
	// Plan-specific configuration
	Plan PlanConfig `mapstructure:"plan"`

	// Sensitive resources and properties configuration
	SensitiveResources  []SensitiveResource `mapstructure:"sensitive_resources"`
	SensitiveProperties []SensitiveProperty `mapstructure:"sensitive_properties"`
}

// PlanConfig holds configuration specific to plan operations
type PlanConfig struct {
	DangerThreshold         int    `mapstructure:"danger-threshold"`
	ShowDetails             bool   `mapstructure:"show-details"`
	HighlightDangers        bool   `mapstructure:"highlight-dangers"`
	ShowStatisticsSummary   bool   `mapstructure:"show-statistics-summary"`
	StatisticsSummaryFormat string `mapstructure:"statistics-summary-format"`
	AlwaysShowSensitive     bool   `mapstructure:"always-show-sensitive"` // Always show sensitive resources even when details are disabled
}

func (config *Config) GetLCString(setting string) string {
	if viper.IsSet(setting) {
		return strings.ToLower(viper.GetString(setting))
	}
	return ""
}

func (config *Config) GetString(setting string) string {
	if viper.IsSet(setting) {
		return viper.GetString(setting)
	}
	return ""
}

func (config *Config) GetStringSlice(setting string) []string {
	if viper.IsSet(setting) {
		return viper.GetStringSlice(setting)
	}
	return []string{}
}

func (config *Config) GetBool(setting string) bool {
	return viper.GetBool(setting)
}

func (config *Config) GetInt(setting string) int {
	if viper.IsSet(setting) {
		return viper.GetInt(setting)
	}
	return 0
}

func (config *Config) GetSeparator() string {
	switch config.GetLCString("output") {
	case "table":
		return "\r\n"
	case "markdown":
		return "\n"
	case "dot":
		return ","
	default:
		return ", "
	}
}

func (config *Config) GetFieldOrEmptyValue(value string) string {
	if value != "" {
		return value
	}
	switch config.GetLCString("output") {
	case "json":
		return ""
	case "markdown":
		return "-"
	default:
		return "-"
	}
}

// GetTimezoneLocation gets the location object you can use in a time object
// based on the timezone specified in your settings.
func (config *Config) GetTimezoneLocation() *time.Location {
	location, err := time.LoadLocation(config.GetString("timezone"))
	if err != nil {
		panic(err)
	}
	return location
}

func (config *Config) NewOutputSettings() *format.OutputSettings {
	settings := format.NewOutputSettings()
	settings.UseEmoji = true
	settings.UseColors = true
	settings.SetOutputFormat(config.GetLCString("output"))
	settings.OutputFile = config.GetLCString("output-file")
	settings.OutputFileFormat = config.GetLCString("output-file-format")
	// settings.ShouldAppend = config.GetBool("output.append")
	settings.TableStyle = format.TableStyles[config.GetString("table.style")]
	settings.TableMaxColumnWidth = config.GetInt("table.max-column-width")

	// Configure markdown settings if needed
	if config.GetLCString("output") == "markdown" {
		// Ensure markdown tables are properly formatted
		settings.UseColors = false
		settings.UseEmoji = true
	}

	// Apply placeholder resolution to file path if specified
	if settings.OutputFile != "" {
		settings.OutputFile = config.resolvePlaceholders(settings.OutputFile)
	}

	// Default file format to stdout format if not specified
	if settings.OutputFileFormat == "" {
		settings.OutputFileFormat = settings.OutputFormat
	}

	return settings
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
