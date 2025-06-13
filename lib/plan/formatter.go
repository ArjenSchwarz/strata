package plan

import (
	"fmt"
	"strings"

	format "github.com/ArjenSchwarz/go-output"
	"github.com/ArjenSchwarz/strata/config"
)

// Formatter handles different output formats for plan summaries
type Formatter struct {
	config *config.Config
}

// NewFormatter creates a new formatter instance
func NewFormatter(cfg *config.Config) *Formatter {
	return &Formatter{
		config: cfg,
	}
}

// OutputSummary outputs the plan summary using go-output library
func (f *Formatter) OutputSummary(summary *PlanSummary, outputFormat string, showDetails bool) error {
	// Output plan information section
	if err := f.formatPlanInfo(summary, outputFormat); err != nil {
		return fmt.Errorf("failed to format plan info: %w", err)
	}

	// Output enhanced statistics summary
	if err := f.formatStatisticsSummary(summary, outputFormat); err != nil {
		return fmt.Errorf("failed to format statistics summary: %w", err)
	}

	// Output enhanced resource changes table if requested
	if showDetails {
		if err := f.formatResourceChangesTable(summary, outputFormat); err != nil {
			return fmt.Errorf("failed to format resource changes table: %w", err)
		}
	}

	return nil
}

// formatStatisticsSummary formats and outputs the horizontal statistics summary table
func (f *Formatter) formatStatisticsSummary(summary *PlanSummary, outputFormat string) error {
	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = fmt.Sprintf("Summary for %s", summary.PlanFile)

	// Create horizontal statistics data
	statsData := []format.OutputHolder{
		{
			Contents: map[string]interface{}{
				"TOTAL":        summary.Statistics.Total,
				"ADDED":        summary.Statistics.ToAdd,
				"REMOVED":      summary.Statistics.ToDestroy,
				"MODIFIED":     summary.Statistics.ToChange,
				"REPLACEMENTS": summary.Statistics.Replacements,
				"CONDITIONALS": summary.Statistics.Conditionals,
			},
		},
	}

	// Use go-output to handle all formatting
	output := format.OutputArray{
		Settings: settings,
		Contents: statsData,
		Keys:     []string{"TOTAL", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "CONDITIONALS"},
	}

	output.Write()
	fmt.Println() // Add spacing between sections
	return nil
}

// createResourceChangesData converts the resource changes into OutputHolder format for detailed view
func (f *Formatter) createResourceChangesData(summary *PlanSummary) []format.OutputHolder {
	var data []format.OutputHolder

	// Add resource changes
	for _, change := range summary.ResourceChanges {
		warning := ""
		if change.IsDestructive {
			warning = "⚠️ DESTRUCTIVE"
		}

		data = append(data, format.OutputHolder{
			Contents: map[string]interface{}{
				"Type":    "Resource Change",
				"Key":     change.Address,
				"Value":   string(change.ChangeType),
				"Details": change.Type,
				"Warning": warning,
				"Icon":    getChangeIcon(change.ChangeType),
			},
		})
	}

	// Add output changes
	for _, change := range summary.OutputChanges {
		sensitive := ""
		if change.Sensitive {
			sensitive = "🔒 SENSITIVE"
		}

		data = append(data, format.OutputHolder{
			Contents: map[string]interface{}{
				"Type":    "Output Change",
				"Key":     change.Name,
				"Value":   string(change.ChangeType),
				"Details": sensitive,
				"Warning": "",
				"Icon":    getChangeIcon(change.ChangeType),
			},
		})
	}

	return data
}

// formatPlanInfo formats and outputs the plan information section
func (f *Formatter) formatPlanInfo(summary *PlanSummary, outputFormat string) error {
	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Plan Information"

	// Create plan info data
	planInfoData := []format.OutputHolder{
		{
			Contents: map[string]interface{}{
				"Key":   "Plan File",
				"Value": summary.PlanFile,
			},
		},
		{
			Contents: map[string]interface{}{
				"Key":   "Terraform Version",
				"Value": summary.TerraformVersion,
			},
		},
		{
			Contents: map[string]interface{}{
				"Key":   "Workspace",
				"Value": summary.Workspace,
			},
		},
		{
			Contents: map[string]interface{}{
				"Key":   "Backend",
				"Value": fmt.Sprintf("%s (%s)", summary.Backend.Type, summary.Backend.Location),
			},
		},
		{
			Contents: map[string]interface{}{
				"Key":   "Created",
				"Value": summary.CreatedAt.Format("2006-01-02 15:04:05"),
			},
		},
		{
			Contents: map[string]interface{}{
				"Key": "Dry Run",
				"Value": func() string {
					if summary.IsDryRun {
						return "Yes"
					}
					return "No"
				}(),
			},
		},
	}

	// Use go-output to handle all formatting
	output := format.OutputArray{
		Settings: settings,
		Contents: planInfoData,
		Keys:     []string{"Key", "Value"},
	}

	output.Write()
	fmt.Println() // Add spacing between sections
	return nil
}

// formatResourceChangesTable formats and outputs the enhanced resource changes table
func (f *Formatter) formatResourceChangesTable(summary *PlanSummary, outputFormat string) error {
	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Resource Changes"

	// Create enhanced resource changes data
	resourceData := []format.OutputHolder{}

	for _, change := range summary.ResourceChanges {
		// Determine the display ID based on change type
		displayID := change.PhysicalID
		if change.ChangeType == ChangeTypeCreate {
			displayID = "-"
		} else if change.ChangeType == ChangeTypeDelete {
			displayID = change.PhysicalID
		}

		// Format replacement type for display
		replacementDisplay := string(change.ReplacementType)
		if change.ChangeType == ChangeTypeDelete {
			replacementDisplay = "N/A"
		}

		// Format danger information
		dangerInfo := ""
		if change.IsDangerous {
			dangerInfo = "⚠️ " + change.DangerReason
			if len(change.DangerProperties) > 0 {
				dangerInfo += ": " + strings.Join(change.DangerProperties, ", ")
			}
		}

		resourceData = append(resourceData, format.OutputHolder{
			Contents: map[string]interface{}{
				"ACTION":      getActionDisplay(change.ChangeType),
				"RESOURCE":    change.Address,
				"TYPE":        change.Type,
				"ID":          displayID,
				"REPLACEMENT": replacementDisplay,
				"MODULE":      change.ModulePath,
				"DANGER":      dangerInfo,
			},
		})
	}

	// Use go-output to handle all formatting
	output := format.OutputArray{
		Settings: settings,
		Contents: resourceData,
		Keys:     []string{"ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"},
	}

	output.Write()
	fmt.Println() // Add spacing between sections
	return nil
}

// getActionDisplay returns the display name for a change type
func getActionDisplay(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeCreate:
		return "Add"
	case ChangeTypeUpdate:
		return "Modify"
	case ChangeTypeDelete:
		return "Remove"
	case ChangeTypeReplace:
		return "Replace"
	default:
		return "No-op"
	}
}

// getChangeIcon returns the appropriate icon for a change type
func getChangeIcon(changeType ChangeType) string {
	switch changeType {
	case ChangeTypeCreate:
		return "+"
	case ChangeTypeUpdate:
		return "~"
	case ChangeTypeDelete:
		return "-"
	case ChangeTypeReplace:
		return "±"
	default:
		return " "
	}
}
