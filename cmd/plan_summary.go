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

Examples:
  # Generate summary from plan file
  strata plan summary terraform.tfplan

  # Generate summary with JSON output
  strata plan summary --output json terraform.tfplan

  # Generate summary with custom danger threshold
  strata plan summary --danger-threshold 5 terraform.tfplan

  # Generate summary with vertical statistics format
  strata plan summary --stats-format vertical terraform.tfplan

  # Generate summary without statistics table
  strata plan summary --show-statistics=false terraform.tfplan`,
	Args: cobra.ExactArgs(1),
	RunE: runPlanSummary,
}

var (
	outputFormat            string
	dangerThreshold         int
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

	// Create config for analyzer
	cfg := &config.Config{
		Plan: config.PlanConfig{
			DangerThreshold:         dangerThreshold,
			ShowDetails:             showDetails,
			HighlightDangers:        highlightDangers,
			ShowStatisticsSummary:   showStatisticsSummary,
			StatisticsSummaryFormat: statisticsSummaryFormat,
			AlwaysShowSensitive:     true, // Always show sensitive resources by default
		},
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

	// Create analyzer and generate summary
	analyzer := plan.NewAnalyzer(tfPlan, cfg)
	summary := analyzer.GenerateSummary(planFile)

	// Check for dangerous changes if threshold is set
	if highlightDangers && analyzer.GetDestructiveChangeCount(summary.ResourceChanges) >= dangerThreshold {
		fmt.Printf("⚠️  WARNING: %d destructive changes detected (threshold: %d)\n\n",
			analyzer.GetDestructiveChangeCount(summary.ResourceChanges), dangerThreshold)
	}

	// Create formatter and output summary
	formatter := plan.NewFormatter(cfg)

	return formatter.OutputSummary(summary, outputFormat, showDetails)
}

func init() {
	planCmd.AddCommand(planSummaryCmd)

	// Output format flag
	planSummaryCmd.Flags().StringVarP(&outputFormat, "output", "o", "table",
		"Output format (table, json, html, markdown)")
	viper.BindPFlag("output", planSummaryCmd.Flags().Lookup("output"))

	// Danger threshold flag
	planSummaryCmd.Flags().IntVar(&dangerThreshold, "danger-threshold", 3,
		"Number of destructive changes to trigger danger warning")
	viper.BindPFlag("plan.danger-threshold", planSummaryCmd.Flags().Lookup("danger-threshold"))

	// Show details flag
	planSummaryCmd.Flags().BoolVar(&showDetails, "details", false,
		"Show detailed change information")
	viper.BindPFlag("plan.show-details", planSummaryCmd.Flags().Lookup("details"))

	// Highlight dangers flag
	planSummaryCmd.Flags().BoolVar(&highlightDangers, "highlight-dangers", true,
		"Highlight potentially destructive changes")
	viper.BindPFlag("plan.highlight-dangers", planSummaryCmd.Flags().Lookup("highlight-dangers"))

	// Show statistics summary flag
	planSummaryCmd.Flags().BoolVar(&showStatisticsSummary, "show-statistics", true,
		"Show statistics summary table")
	viper.BindPFlag("plan.show-statistics-summary", planSummaryCmd.Flags().Lookup("show-statistics"))

	// Statistics summary format flag
	planSummaryCmd.Flags().StringVar(&statisticsSummaryFormat, "stats-format", "horizontal",
		"Statistics summary format (horizontal, vertical)")
	viper.BindPFlag("plan.statistics-summary-format", planSummaryCmd.Flags().Lookup("stats-format"))
}
