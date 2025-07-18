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

	// For testing compatibility, use individual formatter methods to avoid buffer state issues
	if f.config.NewOutputSettings().OutputFile == "" {
		// When no file output is specified, use individual methods for clean output
		err := f.formatPlanInfo(summary, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to format plan info: %w", err)
		}

		err = f.formatStatisticsSummary(summary, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to format statistics summary: %w", err)
		}

		if showDetails {
			err = f.formatResourceChangesTable(summary, outputFormat)
			if err != nil {
				return fmt.Errorf("failed to format resource changes: %w", err)
			}
		} else if f.config.Plan.AlwaysShowSensitive {
			err = f.formatSensitiveResourceChanges(summary, outputFormat)
			if err != nil {
				return fmt.Errorf("failed to format sensitive resources: %w", err)
			}
		}

		return nil
	}

	// Create output array for multi-section output (for file output)
	settings := f.config.NewOutputSettings()
	if settings == nil {
		return fmt.Errorf("failed to create output settings")
	}

	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.SeparateTables = true // Enable multi-table support

	output := format.OutputArray{
		Settings: settings,
		Contents: []format.OutputHolder{},
		Keys:     []string{},
	}

	// Add plan information section
	planInfoData, planInfoKeys, err := f.createPlanInfoData(summary, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to create plan info data: %w", err)
	}
	output.Contents = planInfoData
	output.Keys = planInfoKeys
	output.Settings.Title = "Plan Information"
	output.AddToBuffer()

	// Add statistics summary section
	statsData, statsKeys, err := f.createStatisticsSummaryData(summary, outputFormat)
	if err != nil {
		return fmt.Errorf("failed to create statistics summary data: %w", err)
	}
	output.Contents = statsData
	output.Keys = statsKeys
	output.Settings.Title = fmt.Sprintf("Summary for %s", summary.PlanFile)
	output.AddToBuffer()

	// Add resource changes table if requested
	if showDetails {
		resourceData, resourceKeys, err := f.createResourceChangesData(summary, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to create resource changes data: %w", err)
		}
		output.Contents = resourceData
		output.Keys = resourceKeys
		output.Settings.Title = "Resource Changes"
		output.AddToBuffer()
	} else if f.config.Plan.AlwaysShowSensitive {
		// When details are disabled but AlwaysShowSensitive is enabled,
		// show only the sensitive resource changes
		sensitiveData, sensitiveKeys, err := f.createSensitiveResourceChangesData(summary, outputFormat)
		if err != nil {
			return fmt.Errorf("failed to create sensitive resource changes data: %w", err)
		}
		if len(sensitiveData) > 0 {
			output.Contents = sensitiveData
			output.Keys = sensitiveKeys
			output.Settings.Title = "Sensitive Resource Changes"
			output.AddToBuffer()
		} else {
			// Handle no sensitive changes case
			fmt.Println("No sensitive resource changes detected.")
		}
	}

	// Write all accumulated sections
	output.Write()
	return nil
}

// formatPlanInfo formats plan info to stdout (for testing compatibility)
func (f *Formatter) formatPlanInfo(summary *PlanSummary, outputFormat string) error {
	planInfoData, planInfoKeys, err := f.createPlanInfoData(summary, outputFormat)
	if err != nil {
		return err
	}

	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Plan Information"
	settings.OutputFile = "" // No file output for individual methods

	output := format.OutputArray{
		Settings: settings,
		Contents: planInfoData,
		Keys:     planInfoKeys,
	}

	output.Write()
	fmt.Println() // Add spacing
	return nil
}

// formatStatisticsSummary formats statistics to stdout (for testing compatibility)
func (f *Formatter) formatStatisticsSummary(summary *PlanSummary, outputFormat string) error {
	statsData, statsKeys, err := f.createStatisticsSummaryData(summary, outputFormat)
	if err != nil {
		return err
	}

	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = fmt.Sprintf("Summary for %s", summary.PlanFile)
	settings.OutputFile = "" // No file output for individual methods

	output := format.OutputArray{
		Settings: settings,
		Contents: statsData,
		Keys:     statsKeys,
	}

	output.Write()
	fmt.Println() // Add spacing
	return nil
}

// formatSensitiveResourceChanges formats sensitive resources to stdout (for testing compatibility)
func (f *Formatter) formatSensitiveResourceChanges(summary *PlanSummary, outputFormat string) error {
	sensitiveData, sensitiveKeys, err := f.createSensitiveResourceChangesData(summary, outputFormat)
	if err != nil {
		return err
	}

	if len(sensitiveData) == 0 {
		fmt.Println("No sensitive resource changes detected.")
		fmt.Println() // Add spacing
		return nil
	}

	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Sensitive Resource Changes"
	settings.OutputFile = "" // No file output for individual methods

	output := format.OutputArray{
		Settings: settings,
		Contents: sensitiveData,
		Keys:     sensitiveKeys,
	}

	output.Write()
	fmt.Println() // Add spacing
	return nil
}

// formatResourceChangesTable formats resource changes to stdout (for testing compatibility)
func (f *Formatter) formatResourceChangesTable(summary *PlanSummary, outputFormat string) error {
	resourceData, resourceKeys, err := f.createResourceChangesData(summary, outputFormat)
	if err != nil {
		return err
	}

	settings := f.config.NewOutputSettings()
	settings.SetOutputFormat(outputFormat)
	settings.UseColors = true
	settings.UseEmoji = true
	settings.Title = "Resource Changes"
	settings.OutputFile = "" // No file output for individual methods

	output := format.OutputArray{
		Settings: settings,
		Contents: resourceData,
		Keys:     resourceKeys,
	}

	output.Write()
	fmt.Println() // Add spacing
	return nil
}

// createStatisticsSummaryData creates the statistics summary data
func (f *Formatter) createStatisticsSummaryData(summary *PlanSummary, outputFormat string) ([]format.OutputHolder, []string, error) {
	// Validate inputs
	if summary == nil {
		return nil, nil, fmt.Errorf("summary cannot be nil")
	}
	if summary.PlanFile == "" {
		return nil, nil, fmt.Errorf("plan file name is required")
	}

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

	keys := []string{"TOTAL", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "CONDITIONALS", "HIGH RISK"}
	return statsData, keys, nil
}

// createResourceChangesData creates the resource changes data for detailed view
func (f *Formatter) createResourceChangesData(summary *PlanSummary, outputFormat string) ([]format.OutputHolder, []string, error) {
	// Validate inputs
	if summary == nil {
		return nil, nil, fmt.Errorf("summary cannot be nil")
	}

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

	keys := []string{"ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"}
	return resourceData, keys, nil
}

// createResourceChangesDataOld converts the resource changes into OutputHolder format for detailed view (old format)
func (f *Formatter) createResourceChangesDataOld(summary *PlanSummary) []format.OutputHolder {
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

// createPlanInfoData creates the plan information data
func (f *Formatter) createPlanInfoData(summary *PlanSummary, outputFormat string) ([]format.OutputHolder, []string, error) {
	// Validate inputs
	if summary == nil {
		return nil, nil, fmt.Errorf("summary cannot be nil")
	}

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

	keys := []string{"Plan File", "Version", "Workspace", "Backend", "Created"}
	return planInfoData, keys, nil
}

// createSensitiveResourceChangesData creates data for sensitive resource changes only
func (f *Formatter) createSensitiveResourceChangesData(summary *PlanSummary, outputFormat string) ([]format.OutputHolder, []string, error) {
	// Validate inputs
	if summary == nil {
		return nil, nil, fmt.Errorf("summary cannot be nil")
	}

	// Filter for sensitive resources
	sensitiveChanges := []ResourceChange{}
	for _, change := range summary.ResourceChanges {
		if change.IsDangerous {
			sensitiveChanges = append(sensitiveChanges, change)
		}
	}

	// If no sensitive changes, return empty data
	if len(sensitiveChanges) == 0 {
		return []format.OutputHolder{}, []string{}, nil
	}

	// Create resource data for sensitive changes
	resourceData := []format.OutputHolder{}

	for _, change := range sensitiveChanges {
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

	keys := []string{"ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"}
	return resourceData, keys, nil
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
