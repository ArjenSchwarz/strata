/*
Copyright © 2025 Arjen Schwarz <developer@arjen.eu>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/ArjenSchwarz/strata/lib/errors"
	"github.com/ArjenSchwarz/strata/lib/workflow"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Execute Terraform plan and apply workflow",
	Long: `Execute a complete Terraform workflow that includes planning, analysis, and applying changes.

This command wraps the Terraform plan and apply commands, providing a streamlined workflow
that displays a summary of planned changes and allows you to review them before applying.
The workflow is inspired by deployment tools like Fog for CloudFormation.

The workflow:
1. Execute 'terraform plan' to generate a plan
2. Analyze and display a summary of the planned changes
3. Prompt for user action (apply, view details, or cancel)
4. Execute 'terraform apply' if approved

Examples:
  # Run the complete workflow in current directory
  strata apply

  # Run with custom Terraform binary path
  strata apply --terraform-path /usr/local/bin/terraform

  # Run in non-interactive mode (auto-approve)
  strata apply --non-interactive

  # Run with custom working directory
  strata apply --working-dir /path/to/terraform

  # Run with custom plan arguments
  strata apply --plan-args "-var-file=prod.tfvars"

  # Run with custom apply arguments
  strata apply --apply-args "-parallelism=5"

  # Force apply in non-interactive mode even with destructive changes
  strata apply --non-interactive --force`,
	RunE: runApply,
}

var (
	terraformPath        string
	workingDir           string
	planArgs             []string
	applyArgs            []string
	nonInteractive       bool
	force                bool
	applyOutputFormat    string
	applyDangerThreshold int
)

func runApply(cmd *cobra.Command, args []string) error {
	// Load configuration from file
	cfg, err := loadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set defaults if not configured
	if cfg.Plan.DangerThreshold == 0 {
		cfg.Plan.DangerThreshold = applyDangerThreshold
	}
	if cfg.Terraform.Path == "" {
		cfg.Terraform.Path = terraformPath
	}
	if cfg.Terraform.DangerThreshold == 0 {
		cfg.Terraform.DangerThreshold = applyDangerThreshold
	}

	// Set plan defaults
	cfg.Plan.ShowDetails = false
	cfg.Plan.HighlightDangers = true
	cfg.Plan.ShowStatisticsSummary = true
	cfg.Plan.StatisticsSummaryFormat = "horizontal"
	cfg.Plan.AlwaysShowSensitive = true

	// Set terraform defaults
	cfg.Terraform.ShowDetails = false

	// Override with command-line flags
	if cmd.Flags().Changed("terraform-path") {
		cfg.Terraform.Path = terraformPath
	}
	if cmd.Flags().Changed("working-dir") {
		// This is not stored in config, but used directly
	}
	if cmd.Flags().Changed("plan-args") {
		cfg.Terraform.PlanArgs = planArgs
	}
	if cmd.Flags().Changed("apply-args") {
		cfg.Terraform.ApplyArgs = applyArgs
	}
	if cmd.Flags().Changed("danger-threshold") {
		cfg.Terraform.DangerThreshold = applyDangerThreshold
		cfg.Plan.DangerThreshold = applyDangerThreshold
	}
	if cmd.Flags().Changed("output") {
		// This is not stored in config, but used directly
	}

	// Create workflow manager
	workflowManager := workflow.NewWorkflowManager(cfg)

	// Create workflow options
	options := &workflow.WorkflowOptions{
		TerraformPath:   terraformPath,
		WorkingDir:      workingDir,
		PlanArgs:        planArgs,
		ApplyArgs:       applyArgs,
		NonInteractive:  nonInteractive,
		Force:           force,
		OutputFormat:    applyOutputFormat,
		DangerThreshold: applyDangerThreshold,
		Timeout:         30 * time.Minute,
		Environment:     make(map[string]string),
	}

	// Execute the workflow
	ctx := context.Background()
	err = workflowManager.Run(ctx, options)

	// Handle errors with proper exit codes and user-friendly messages
	if err != nil {
		if strataErr, ok := err.(*errors.StrataError); ok {
			// Display user-friendly error message
			fmt.Fprintln(os.Stderr, strataErr.FormatUserMessage())

			// Set appropriate exit code based on error type
			if strataErr.GetCode() == errors.ErrorCodeWorkflowCancelled {
				os.Exit(2) // User cancelled
			} else if strataErr.IsUserError() {
				os.Exit(1) // User error
			} else if strataErr.IsCritical() {
				os.Exit(3) // Critical system error
			} else {
				os.Exit(1) // General error
			}
		} else {
			// Fallback for non-StrataError errors
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Terraform path flag
	applyCmd.Flags().StringVar(&terraformPath, "terraform-path", "terraform",
		"Path to the Terraform binary")
	viper.BindPFlag("terraform.path", applyCmd.Flags().Lookup("terraform-path"))

	// Working directory flag
	applyCmd.Flags().StringVar(&workingDir, "working-dir", ".",
		"Working directory for Terraform commands")
	viper.BindPFlag("terraform.working-dir", applyCmd.Flags().Lookup("working-dir"))

	// Plan arguments flag
	applyCmd.Flags().StringSliceVar(&planArgs, "plan-args", []string{},
		"Additional arguments to pass to terraform plan")
	viper.BindPFlag("terraform.plan-args", applyCmd.Flags().Lookup("plan-args"))

	// Apply arguments flag
	applyCmd.Flags().StringSliceVar(&applyArgs, "apply-args", []string{},
		"Additional arguments to pass to terraform apply")
	viper.BindPFlag("terraform.apply-args", applyCmd.Flags().Lookup("apply-args"))

	// Non-interactive flag
	applyCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false,
		"Run in non-interactive mode (auto-approve)")
	viper.BindPFlag("terraform.non-interactive", applyCmd.Flags().Lookup("non-interactive"))

	// Force flag
	applyCmd.Flags().BoolVar(&force, "force", false,
		"Force apply in non-interactive mode even with destructive changes")
	viper.BindPFlag("terraform.force", applyCmd.Flags().Lookup("force"))

	// Output format flag (inherited from plan summary)
	applyCmd.Flags().StringVarP(&applyOutputFormat, "output", "o", "table",
		"Output format for plan summary (table, json, html, markdown)")
	viper.BindPFlag("output", applyCmd.Flags().Lookup("output"))

	// Danger threshold flag (inherited from plan summary)
	applyCmd.Flags().IntVar(&applyDangerThreshold, "danger-threshold", 3,
		"Number of destructive changes to trigger danger warning")
	viper.BindPFlag("plan.danger-threshold", applyCmd.Flags().Lookup("danger-threshold"))
}

// loadConfiguration loads configuration from file and returns a Config struct
func loadConfiguration() (*config.Config, error) {
	// Set configuration file name and paths
	viper.SetConfigName("strata")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	// Set default values
	viper.SetDefault("terraform.path", "terraform")
	viper.SetDefault("terraform.danger-threshold", 3)
	viper.SetDefault("terraform.show-details", false)
	viper.SetDefault("terraform.timeout", "30m")
	viper.SetDefault("plan.danger-threshold", 3)
	viper.SetDefault("plan.show-details", false)
	viper.SetDefault("plan.highlight-dangers", true)
	viper.SetDefault("plan.show-statistics-summary", true)
	viper.SetDefault("plan.statistics-summary-format", "horizontal")
	viper.SetDefault("plan.always-show-sensitive", true)

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, &errors.StrataError{
				Code:       errors.ErrorCodeConfigurationInvalid,
				Message:    "Error reading configuration file",
				Underlying: err,
				Context: map[string]interface{}{
					"config_paths": []string{"./strata.yaml", "$HOME/strata.yaml"},
				},
				Suggestions: []string{
					"Check if the configuration file has valid YAML syntax",
					"Verify file permissions are correct",
					"Remove the configuration file to use defaults",
				},
				RecoveryAction: "Fix configuration file syntax or remove it",
			}
		}
	}

	// Create config struct and unmarshal
	cfg := &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, &errors.StrataError{
			Code:       errors.ErrorCodeConfigurationInvalid,
			Message:    "Error parsing configuration",
			Underlying: err,
			Suggestions: []string{
				"Check configuration file syntax",
				"Verify all configuration values are valid",
				"Use 'strata --help' to see available options",
			},
			RecoveryAction: "Fix configuration syntax",
		}
	}

	// Validate configuration
	if err := validateConfiguration(cfg); err != nil {
		if strataErr, ok := err.(*errors.StrataError); ok {
			return nil, strataErr
		}
		return nil, &errors.StrataError{
			Code:       errors.ErrorCodeConfigurationInvalid,
			Message:    "Invalid configuration",
			Underlying: err,
			Suggestions: []string{
				"Check configuration values are within valid ranges",
				"Verify all required fields are set",
			},
			RecoveryAction: "Fix configuration values",
		}
	}

	return cfg, nil
}

// validateConfiguration validates the loaded configuration
func validateConfiguration(cfg *config.Config) error {
	// Validate terraform path is not empty
	if cfg.Terraform.Path == "" {
		return &errors.StrataError{
			Code:    errors.ErrorCodeConfigurationInvalid,
			Message: "Terraform path cannot be empty",
			Context: map[string]interface{}{
				"field": "terraform.path",
			},
			Suggestions: []string{
				"Set terraform.path in configuration file",
				"Use --terraform-path flag to specify path",
				"Ensure Terraform is installed and in PATH",
			},
			RecoveryAction: "Set terraform.path configuration",
		}
	}

	// Validate danger threshold is positive
	if cfg.Terraform.DangerThreshold < 0 {
		return &errors.StrataError{
			Code:    errors.ErrorCodeConfigurationInvalid,
			Message: "Terraform danger threshold must be non-negative",
			Context: map[string]interface{}{
				"field": "terraform.danger-threshold",
				"value": cfg.Terraform.DangerThreshold,
			},
			Suggestions: []string{
				"Set terraform.danger-threshold to 0 or higher",
				"Use --danger-threshold flag to override",
			},
			RecoveryAction: "Set danger threshold to 0 or higher",
		}
	}
	if cfg.Plan.DangerThreshold < 0 {
		return &errors.StrataError{
			Code:    errors.ErrorCodeConfigurationInvalid,
			Message: "Plan danger threshold must be non-negative",
			Context: map[string]interface{}{
				"field": "plan.danger-threshold",
				"value": cfg.Plan.DangerThreshold,
			},
			Suggestions: []string{
				"Set plan.danger-threshold to 0 or higher",
				"Use --danger-threshold flag to override",
			},
			RecoveryAction: "Set danger threshold to 0 or higher",
		}
	}

	// Validate timeout format if specified
	if cfg.Terraform.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Terraform.Timeout); err != nil {
			return &errors.StrataError{
				Code:       errors.ErrorCodeConfigurationInvalid,
				Message:    "Invalid timeout format",
				Underlying: err,
				Context: map[string]interface{}{
					"field": "terraform.timeout",
					"value": cfg.Terraform.Timeout,
				},
				Suggestions: []string{
					"Use valid duration format (e.g., '30m', '1h', '90s')",
					"Check Go duration format documentation",
				},
				RecoveryAction: "Fix timeout format",
			}
		}
	}

	// Validate statistics summary format
	validFormats := []string{"horizontal", "vertical", "compact"}
	if cfg.Plan.StatisticsSummaryFormat != "" {
		valid := false
		for _, format := range validFormats {
			if cfg.Plan.StatisticsSummaryFormat == format {
				valid = true
				break
			}
		}
		if !valid {
			return &errors.StrataError{
				Code:    errors.ErrorCodeConfigurationInvalid,
				Message: "Invalid statistics summary format",
				Context: map[string]interface{}{
					"field":         "plan.statistics-summary-format",
					"value":         cfg.Plan.StatisticsSummaryFormat,
					"valid_formats": validFormats,
				},
				Suggestions: []string{
					fmt.Sprintf("Use one of: %v", validFormats),
					"Check configuration documentation",
				},
				RecoveryAction: "Set valid statistics summary format",
			}
		}
	}

	return nil
}
