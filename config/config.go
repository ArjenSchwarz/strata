// Package config provides configuration management functionality for the Strata application.
package config

import (
	"os"
	"strings"
	"time"

	format "github.com/ArjenSchwarz/go-output"
	"github.com/spf13/viper"
)

const (
	outputFormatMarkdown = "markdown"
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

	// Terraform workflow configuration
	Terraform TerraformConfig `mapstructure:"terraform"`

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

// TerraformConfig holds configuration specific to Terraform workflow operations
type TerraformConfig struct {
	// Path to Terraform binary
	Path string `mapstructure:"path"`

	// Default arguments for terraform plan
	PlanArgs []string `mapstructure:"plan-args"`

	// Default arguments for terraform apply
	ApplyArgs []string `mapstructure:"apply-args"`

	// Backend configuration
	Backend BackendConfig `mapstructure:"backend"`

	// Danger threshold for highlighting risks
	DangerThreshold int `mapstructure:"danger-threshold"`

	// Whether to show detailed output by default
	ShowDetails bool `mapstructure:"show-details"`

	// Default timeout for operations
	Timeout string `mapstructure:"timeout"`

	// Environment variables to set for Terraform commands
	Environment map[string]string `mapstructure:"environment"`
}

// BackendConfig holds configuration for Terraform backends
type BackendConfig struct {
	// Type of backend (e.g., s3, gcs, azurerm)
	Type string `mapstructure:"type"`

	// Backend-specific configuration
	Config map[string]interface{} `mapstructure:"config"`

	// Lock timeout
	LockTimeout string `mapstructure:"lock-timeout"`

	// Disable locking
	DisableLocking bool `mapstructure:"disable-locking"`
}

// GetLCString returns the lowercase string value for the given setting.
func (config *Config) GetLCString(setting string) string {
	if viper.IsSet(setting) {
		return strings.ToLower(viper.GetString(setting))
	}
	return ""
}

// GetString returns the string value for the given setting.
func (config *Config) GetString(setting string) string {
	if viper.IsSet(setting) {
		return viper.GetString(setting)
	}
	return ""
}

// GetStringSlice returns the string slice value for the given setting.
func (config *Config) GetStringSlice(setting string) []string {
	if viper.IsSet(setting) {
		return viper.GetStringSlice(setting)
	}
	return []string{}
}

// GetBool returns the boolean value for the given setting.
func (config *Config) GetBool(setting string) bool {
	return viper.GetBool(setting)
}

// GetInt returns the integer value for the given setting.
func (config *Config) GetInt(setting string) int {
	if viper.IsSet(setting) {
		return viper.GetInt(setting)
	}
	return 0
}

// GetSeparator returns the appropriate separator string based on the output format.
func (config *Config) GetSeparator() string {
	switch config.GetLCString("output") {
	case "table":
		return "\r\n"
	case outputFormatMarkdown:
		return "\n"
	case "dot":
		return ","
	default:
		return ", "
	}
}

// GetFieldOrEmptyValue returns the value if not empty, otherwise returns an appropriate empty value based on output format.
func (config *Config) GetFieldOrEmptyValue(value string) string {
	if value != "" {
		return value
	}
	switch config.GetLCString("output") {
	case "json":
		return ""
	case outputFormatMarkdown:
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

// NewOutputSettings creates and configures a new OutputSettings instance based on the current configuration.
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
	if config.GetLCString("output") == outputFormatMarkdown {
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
