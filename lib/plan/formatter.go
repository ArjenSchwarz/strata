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
func (f *Formatter) OutputSummary(summary *PlanSummary, outputSettings *format.OutputSettings, showDetails bool) error {
	// Validate output format first
	if err := f.ValidateOutputFormat(outputSettings.OutputFormat); err != nil {
		return err
	}

	// Use the provided output settings directly
	settings := outputSettings
	settings.UseColors = true
	settings.UseEmoji = true
	settings.SeparateTables = true // Enable multi-table support

	output := format.OutputArray{
		Settings: settings,
		Contents: []format.OutputHolder{},
		Keys:     []string{},
	}

	// Add plan information section
	planInfoData, planInfoKeys, err := f.createPlanInfoData(summary, settings)
	if err != nil {
		return fmt.Errorf("failed to create plan info data: %w", err)
	}
	output.Contents = planInfoData
	output.Keys = planInfoKeys
	output.Settings.Title = "Plan Information"
	output.AddToBuffer()

	// Add statistics summary section
	statsData, statsKeys, err := f.createStatisticsSummaryData(summary, settings)
	if err != nil {
		return fmt.Errorf("failed to create statistics summary data: %w", err)
	}
	output.Contents = statsData
	output.Keys = statsKeys
	output.Settings.Title = fmt.Sprintf("Summary for %s", summary.PlanFile)
	output.AddToBuffer()

	// Add resource changes table if requested
	if showDetails {
		resourceData, resourceKeys, err := f.createResourceChangesData(summary, settings)
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
		sensitiveData, sensitiveKeys, err := f.createSensitiveResourceChangesData(summary, settings)
		if err != nil {
			return fmt.Errorf("failed to create sensitive resource changes data: %w", err)
		}
		if len(sensitiveData) > 0 {
			output.Contents = sensitiveData
			output.Keys = sensitiveKeys
			output.Settings.Title = "Sensitive Resource Changes"
			output.AddToBuffer()
		} else {
			// Handle no sensitive changes case - but only for stdout-only mode
			if settings.OutputFile == "" {
				fmt.Println("No sensitive resource changes detected.")
			}
		}
	}

	// Write all accumulated sections
	output.Write()
	return nil
}

// formatPlanInfo formats plan info to stdout (for testing compatibility)
func (f *Formatter) formatPlanInfo(summary *PlanSummary, outputSettings *format.OutputSettings) error {
	planInfoData, planInfoKeys, err := f.createPlanInfoData(summary, outputSettings)
	if err != nil {
		return err
	}

	outputSettings.Title = "Plan Information"

	output := format.OutputArray{
		Settings: outputSettings,
		Contents: planInfoData,
		Keys:     planInfoKeys,
	}

	output.Write()
	fmt.Println() // Add spacing
	return nil
}

// formatStatisticsSummary formats statistics to stdout (for testing compatibility)
func (f *Formatter) formatStatisticsSummary(summary *PlanSummary, outputSettings *format.OutputSettings) error {
	statsData, statsKeys, err := f.createStatisticsSummaryData(summary, outputSettings)
	if err != nil {
		return err
	}

	// Use the same settings but temporarily disable file output
	originalFile := outputSettings.OutputFile
	outputSettings.OutputFile = ""
	outputSettings.Title = fmt.Sprintf("Summary for %s", summary.PlanFile)

	output := format.OutputArray{
		Settings: outputSettings,
		Contents: statsData,
		Keys:     statsKeys,
	}

	output.Write()

	// Restore original file setting
	outputSettings.OutputFile = originalFile
	fmt.Println() // Add spacing
	return nil
}

// formatSensitiveResourceChanges formats sensitive resources to stdout (for testing compatibility)
func (f *Formatter) formatSensitiveResourceChanges(summary *PlanSummary, outputSettings *format.OutputSettings) error {
	sensitiveData, sensitiveKeys, err := f.createSensitiveResourceChangesData(summary, outputSettings)
	if err != nil {
		return err
	}

	if len(sensitiveData) == 0 {
		fmt.Println("No sensitive resource changes detected.")
		fmt.Println() // Add spacing
		return nil
	}

	// Use the same settings but temporarily disable file output
	originalFile := outputSettings.OutputFile
	outputSettings.OutputFile = ""
	outputSettings.Title = "Sensitive Resource Changes"

	output := format.OutputArray{
		Settings: outputSettings,
		Contents: sensitiveData,
		Keys:     sensitiveKeys,
	}

	output.Write()

	// Restore original file setting
	outputSettings.OutputFile = originalFile
	fmt.Println() // Add spacing
	return nil
}

// formatResourceChangesTable formats resource changes to stdout (for testing compatibility)
func (f *Formatter) formatResourceChangesTable(summary *PlanSummary, outputSettings *format.OutputSettings) error {
	resourceData, resourceKeys, err := f.createResourceChangesData(summary, outputSettings)
	if err != nil {
		return err
	}

	// Use the same settings but temporarily disable file output
	originalFile := outputSettings.OutputFile
	outputSettings.OutputFile = ""
	outputSettings.Title = "Resource Changes"

	output := format.OutputArray{
		Settings: outputSettings,
		Contents: resourceData,
		Keys:     resourceKeys,
	}

	output.Write()

	// Restore original file setting
	outputSettings.OutputFile = originalFile
	fmt.Println() // Add spacing
	return nil
}

// createStatisticsSummaryData creates the statistics summary data
func (f *Formatter) createStatisticsSummaryData(summary *PlanSummary, settings *format.OutputSettings) ([]format.OutputHolder, []string, error) {
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
func (f *Formatter) createResourceChangesData(summary *PlanSummary, settings *format.OutputSettings) ([]format.OutputHolder, []string, error) {
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
func (f *Formatter) createPlanInfoData(summary *PlanSummary, settings *format.OutputSettings) ([]format.OutputHolder, []string, error) {
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
func (f *Formatter) createSensitiveResourceChangesData(summary *PlanSummary, settings *format.OutputSettings) ([]format.OutputHolder, []string, error) {
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
