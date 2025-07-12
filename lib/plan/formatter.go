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

// ValidateOutputFormat validates that the output format is supported
func (f *Formatter) ValidateOutputFormat(outputFormat string) error {
	supportedFormats := []string{"table", "json", "html", "markdown"}
	for _, format := range supportedFormats {
		if strings.ToLower(outputFormat) == format {
			return nil
		}
	}
	return fmt.Errorf("unsupported output format '%s'. Supported formats: %s", outputFormat, strings.Join(supportedFormats, ", "))
}

// OutputSummary outputs the plan summary using go-output library
func (f *Formatter) OutputSummary(summary *PlanSummary, outputFormat string, showDetails bool) error {
	// Validate output format first
	if err := f.ValidateOutputFormat(outputFormat); err != nil {
		return err
	}

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
	} else if f.config.Plan.AlwaysShowSensitive {
		// When details are disabled but AlwaysShowSensitive is enabled,
		// show only the sensitive resource changes
		if err := f.formatSensitiveResourceChanges(summary, outputFormat); err != nil {
			return fmt.Errorf("failed to format sensitive resource changes: %w", err)
		}
	}

	return nil
}

// formatStatisticsSummary formats and outputs the horizontal statistics summary table
func (f *Formatter) formatStatisticsSummary(summary *PlanSummary, outputFormat string) error {
	// Validate inputs
	if summary == nil {
		return fmt.Errorf("summary cannot be nil")
	}
	if summary.PlanFile == "" {
		return fmt.Errorf("plan file name is required")
	}

	settings := f.config.NewOutputSettings()
	if settings == nil {
		return fmt.Errorf("failed to create output settings")
	}

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
				"HIGH RISK":    summary.Statistics.HighRisk,
			},
		},
	}

	// Use go-output to handle all formatting
	output := format.OutputArray{
		Settings: settings,
		Contents: statsData,
		Keys:     []string{"TOTAL", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "CONDITIONALS", "HIGH RISK"},
	}

	output.Write()
	fmt.Println() // Add spacing between sections
	return nil
}

// formatPlanInfo formats and outputs the plan information section using a horizontal layout
func (f *Formatter) formatPlanInfo(summary *PlanSummary, outputFormat string) error {
	// Validate inputs
	if summary == nil {
		return fmt.Errorf("summary cannot be nil")
	}

	settings := f.config.NewOutputSettings()
	if settings == nil {
		return fmt.Errorf("failed to create output settings")
	}

	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Plan Information"

	// Create horizontal plan info data with a single row for values
	planInfoData := []format.OutputHolder{
		{
			// Values row
			Contents: map[string]interface{}{
				"Plan File": summary.PlanFile,
				"Version":   summary.TerraformVersion,
				"Workspace": summary.Workspace,
				"Backend":   fmt.Sprintf("%s (%s)", summary.Backend.Type, summary.Backend.Location),
				"Created":   summary.CreatedAt.Format("2006-01-02 15:04:05"),
			},
		},
	}

	// Use go-output to handle all formatting
	output := format.OutputArray{
		Settings: settings,
		Contents: planInfoData,
		Keys:     []string{"Plan File", "Version", "Workspace", "Backend", "Created"},
	}

	output.Write()
	fmt.Println() // Add spacing between sections
	return nil
}

// formatSensitiveResourceChanges formats and outputs only sensitive resource changes
func (f *Formatter) formatSensitiveResourceChanges(summary *PlanSummary, outputFormat string) error {
	// Validate inputs
	if summary == nil {
		return fmt.Errorf("summary cannot be nil")
	}

	settings := f.config.NewOutputSettings()
	if settings == nil {
		return fmt.Errorf("failed to create output settings")
	}

	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Sensitive Resource Changes"

	// Filter for sensitive resources
	sensitiveChanges := []ResourceChange{}
	for _, change := range summary.ResourceChanges {
		if change.IsDangerous {
			sensitiveChanges = append(sensitiveChanges, change)
		}
	}

	// If no sensitive changes, return early
	if len(sensitiveChanges) == 0 {
		fmt.Println("No sensitive resource changes detected.")
		fmt.Println() // Add spacing
		return nil
	}

	// Create resource data for sensitive changes
	resourceData := []format.OutputHolder{}

	for _, change := range sensitiveChanges {
		// Determine the display ID based on change type
		displayID := change.PhysicalID
		switch change.ChangeType {
		case ChangeTypeCreate:
			displayID = "-"
		case ChangeTypeDelete:
			displayID = change.PhysicalID
		}

		// Format replacement type for display
		replacementDisplay := string(change.ReplacementType)
		if change.ChangeType == ChangeTypeDelete {
			replacementDisplay = notApplicableValue
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

// formatResourceChangesTable formats and outputs the enhanced resource changes table
func (f *Formatter) formatResourceChangesTable(summary *PlanSummary, outputFormat string) error {
	// Validate inputs
	if summary == nil {
		return fmt.Errorf("summary cannot be nil")
	}

	settings := f.config.NewOutputSettings()
	if settings == nil {
		return fmt.Errorf("failed to create output settings")
	}

	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Resource Changes"

	// Create enhanced resource changes data
	resourceData := []format.OutputHolder{}

	for _, change := range summary.ResourceChanges {
		// Determine the display ID based on change type
		displayID := change.PhysicalID
		switch change.ChangeType {
		case ChangeTypeCreate:
			displayID = "-"
		case ChangeTypeDelete:
			displayID = change.PhysicalID
		}

		// Format replacement type for display
		replacementDisplay := string(change.ReplacementType)
		if change.ChangeType == ChangeTypeDelete {
			replacementDisplay = notApplicableValue
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
