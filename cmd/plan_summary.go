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
	"fmt"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/ArjenSchwarz/strata/lib/plan"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// planSummaryCmd represents the plan summary command
var planSummaryCmd = &cobra.Command{
	Use:   "summary [plan-file]",
	Short: "Generate a summary of Terraform plan changes",
	Long: `Generate a clear, concise summary of Terraform plan changes.

This command parses Terraform plan files and presents the changes in a
user-friendly format, highlighting potentially destructive operations
and providing statistical summaries.

Progressive Disclosure:
The summary uses collapsible sections to show essential information by default
while providing access to comprehensive details when needed. Property changes,
dependencies, and risk analysis are presented in expandable sections that can
be expanded individually or globally using the --expand-all flag.

Provider Grouping:
When analyzing plans with many resources, the summary automatically groups
resources by provider (aws, azurerm, google, etc.) to improve readability.
Grouping is enabled when resource count exceeds the configured threshold
(default: 10) and multiple providers are present.

Risk Analysis:
The summary automatically identifies and highlights potentially risky changes:
- Deletions and replacements are flagged based on resource sensitivity
- Sensitive resource types and properties are highlighted with warnings
- High-risk changes have their detail sections auto-expanded for visibility

File Output:
The --file and --file-format flags allow you to save output to a file in addition
to displaying it on stdout. The file format can be different from the stdout format.
File paths support placeholders for dynamic naming:
  $TIMESTAMP    - Current timestamp (2006-01-02T15-04-05 format)
  $AWS_REGION   - AWS region from context
  $AWS_ACCOUNTID - AWS account ID from context

Examples:
  # Generate summary from plan file
  strata plan summary terraform.tfplan

  # Generate summary with JSON output
  strata plan summary --output json terraform.tfplan

  # Expand all collapsible sections to see full details
  strata plan summary --expand-all terraform.tfplan

  # Generate summary with all details expanded in Markdown format
  strata plan summary --output markdown --expand-all terraform.tfplan

  # Save expanded output to file while displaying collapsed on stdout
  strata plan summary --file full-report.json --file-format json --expand-all terraform.tfplan

  # Use placeholders in filename for dynamic naming
  strata plan summary --file "report-$TIMESTAMP-$AWS_REGION.md" --file-format markdown terraform.tfplan

  # Generate summary with vertical statistics format
  strata plan summary --stats-format vertical terraform.tfplan

  # Generate summary without statistics table
  strata plan summary --show-statistics=false terraform.tfplan

Configuration:
The summary behavior can be customized through the strata.yaml configuration file:

  # Global expand control
  expand_all: false                    # Expand all collapsible sections

  plan:
    expandable_sections:
      enabled: true                    # Enable collapsible sections
      auto_expand_dangerous: true      # Auto-expand high-risk sections
      show_dependencies: true          # Show dependency information
    grouping:
      enabled: true                    # Enable provider grouping
      threshold: 10                    # Minimum resources to trigger grouping`,
	Args: cobra.ExactArgs(1),
	RunE: runPlanSummary,
}

var (
	showDetails             bool
	highlightDangers        bool
	showStatisticsSummary   bool
	statisticsSummaryFormat string
)

func runPlanSummary(cmd *cobra.Command, args []string) error {
	planFile := args[0]

	// Create parser and load plan
	parser := plan.NewParser(planFile)
	tfPlan, err := parser.LoadPlan()
	if err != nil {
		return fmt.Errorf("failed to load plan: %w", err)
	}

	// Validate plan structure
	if err := parser.ValidateStructure(tfPlan); err != nil {
		return fmt.Errorf("invalid plan structure: %w", err)
	}

	// Create config for analyzer with defaults
	cfg := config.GetDefaultConfig()
	cfg.Plan.ShowDetails = showDetails
	cfg.Plan.HighlightDangers = highlightDangers
	cfg.Plan.ShowStatisticsSummary = showStatisticsSummary
	cfg.Plan.StatisticsSummaryFormat = statisticsSummaryFormat

	// Read expand-all configuration from Viper (includes CLI flag override)
	cfg.ExpandAll = viper.GetBool("expand_all")

	// Load expandable sections configuration from config file if it exists
	if viper.IsSet("plan.expandable_sections") {
		if err := viper.UnmarshalKey("plan.expandable_sections", &cfg.Plan.ExpandableSections); err != nil {
			return fmt.Errorf("failed to parse expandable_sections config: %w", err)
		}
	}

	// Load grouping configuration from config file if it exists
	if viper.IsSet("plan.grouping") {
		if err := viper.UnmarshalKey("plan.grouping", &cfg.Plan.Grouping); err != nil {
			return fmt.Errorf("failed to parse grouping config: %w", err)
		}
	}

	// Load performance limits configuration from config file if it exists
	if viper.IsSet("plan.performance_limits") {
		if err := viper.UnmarshalKey("plan.performance_limits", &cfg.Plan.PerformanceLimits); err != nil {
			return fmt.Errorf("failed to parse performance_limits config: %w", err)
		}
	}

	// Load sensitive resources and properties from config file if they exist
	if viper.IsSet("sensitive_resources") {
		if err := viper.UnmarshalKey("sensitive_resources", &cfg.SensitiveResources); err != nil {
			return fmt.Errorf("failed to parse sensitive_resources config: %w", err)
		}
	}

	if viper.IsSet("sensitive_properties") {
		if err := viper.UnmarshalKey("sensitive_properties", &cfg.SensitiveProperties); err != nil {
			return fmt.Errorf("failed to parse sensitive_properties config: %w", err)
		}
	}

	// Handle configuration migration and show deprecation warnings
	warnings := cfg.MigrateDeprecatedConfig()
	config.PrintDeprecationWarnings(warnings)

	// Validate configuration
	if err := cfg.ValidateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create analyzer and generate summary
	analyzer := plan.NewAnalyzer(tfPlan, cfg)
	summary := analyzer.GenerateSummary(planFile)

	// Create formatter and output summary
	formatter := plan.NewFormatter(cfg)

	// Create output configuration for v2 API
	outputConfig := cfg.NewOutputConfiguration()

	// Validate file output settings before executing formatter
	if outputConfig.OutputFile != "" {
		validator := config.NewFileValidator(cfg)
		if err := validator.ValidateFileOutput(outputConfig); err != nil {
			return fmt.Errorf("file output validation failed: %w", err)
		}
	}

	return formatter.OutputSummary(summary, outputConfig, showDetails)
}

func init() {
	planCmd.AddCommand(planSummaryCmd)

	// Show details flag
	planSummaryCmd.Flags().BoolVar(&showDetails, "details", true,
		"Show detailed change information")
	if err := viper.BindPFlag("plan.show-details", planSummaryCmd.Flags().Lookup("details")); err != nil {
		panic(err)
	}

	// Highlight dangers flag
	planSummaryCmd.Flags().BoolVar(&highlightDangers, "highlight-dangers", true,
		"Highlight potentially destructive changes")
	if err := viper.BindPFlag("plan.highlight-dangers", planSummaryCmd.Flags().Lookup("highlight-dangers")); err != nil {
		panic(err)
	}

	// Show statistics summary flag
	planSummaryCmd.Flags().BoolVar(&showStatisticsSummary, "show-statistics", true,
		"Show statistics summary table")
	if err := viper.BindPFlag("plan.show-statistics-summary", planSummaryCmd.Flags().Lookup("show-statistics")); err != nil {
		panic(err)
	}

	// Statistics summary format flag
	planSummaryCmd.Flags().StringVar(&statisticsSummaryFormat, "stats-format", "horizontal",
		"Statistics summary format (horizontal, vertical)")
	if err := viper.BindPFlag("plan.statistics-summary-format", planSummaryCmd.Flags().Lookup("stats-format")); err != nil {
		panic(err)
	}
}
