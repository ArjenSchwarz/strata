package plan

import (
	"context"
	"fmt"
	"strings"

	output "github.com/ArjenSchwarz/go-output/v2"
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

// OutputSummary outputs the plan summary using go-output v2 library
func (f *Formatter) OutputSummary(summary *PlanSummary, outputConfig *config.OutputConfiguration, showDetails bool) error {
	ctx := context.Background()

	// Validate output format first
	if err := f.ValidateOutputFormat(outputConfig.Format); err != nil {
		return err
	}

	// Build the document using v2 builder pattern
	builder := output.New()

	// Add plan information section
	planInfoData, err := f.createPlanInfoDataV2(summary)
	if err != nil {
		return fmt.Errorf("failed to create plan info data: %w", err)
	}
	builder = builder.Table("Plan Information", planInfoData,
		output.WithKeys("Plan File", "Version", "Workspace", "Backend", "Created"))

	// Add statistics summary section
	statsData, err := f.createStatisticsSummaryDataV2(summary)
	if err != nil {
		return fmt.Errorf("failed to create statistics summary data: %w", err)
	}
	builder = builder.Table(fmt.Sprintf("Summary for %s", summary.PlanFile), statsData,
		output.WithKeys("TOTAL", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "CONDITIONALS", "HIGH RISK"))

	// Add resource changes table if requested
	if showDetails {
		resourceData, err := f.createResourceChangesDataV2(summary)
		if err != nil {
			return fmt.Errorf("failed to create resource changes data: %w", err)
		}
		builder = builder.Table("Resource Changes", resourceData,
			output.WithKeys("ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"))
	} else if f.config.Plan.AlwaysShowSensitive {
		// When details are disabled but AlwaysShowSensitive is enabled,
		// show only the sensitive resource changes
		sensitiveData, err := f.createSensitiveResourceChangesDataV2(summary)
		if err != nil {
			return fmt.Errorf("failed to create sensitive resource changes data: %w", err)
		}
		if len(sensitiveData) > 0 {
			builder = builder.Table("Sensitive Resource Changes", sensitiveData,
				output.WithKeys("ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"))
		} else {
			// Handle no sensitive changes case - but only for stdout-only mode
			if outputConfig.OutputFile == "" {
				fmt.Println("No sensitive resource changes detected.")
			}
		}
	}

	// Build the document
	doc := builder.Build()

	// Render to stdout first
	stdoutFormat := f.getFormatFromConfig(outputConfig.Format)
	if outputConfig.TableStyle != "" && outputConfig.Format == "table" {
		stdoutFormat = output.TableWithStyle(outputConfig.TableStyle)
	}

	stdoutOptions := []output.OutputOption{
		output.WithFormat(stdoutFormat),
		output.WithWriter(output.NewStdoutWriter()),
	}

	// Add transformers to stdout based on configuration
	if outputConfig.UseEmoji {
		stdoutOptions = append(stdoutOptions, output.WithTransformer(&output.EmojiTransformer{}))
	}
	if outputConfig.UseColors {
		stdoutOptions = append(stdoutOptions, output.WithTransformer(output.NewColorTransformer()))
	}

	stdoutOut := output.NewOutput(stdoutOptions...)
	if err := stdoutOut.Render(ctx, doc); err != nil {
		return fmt.Errorf("failed to render to stdout: %w", err)
	}

	// Render to file if configured
	if outputConfig.OutputFile != "" {
		fileWriter, err := output.NewFileWriter(".", outputConfig.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create file writer: %w", err)
		}
		
		// Enable absolute path support
		fileWriter = fileWriter.WithAbsolutePath()

		fileFormat := f.getFormatFromConfig(outputConfig.OutputFileFormat)
		fileOptions := []output.OutputOption{
			output.WithFormat(fileFormat),
			output.WithWriter(fileWriter),
		}

		// Add transformers to file output based on configuration
		if outputConfig.UseEmoji {
			fileOptions = append(fileOptions, output.WithTransformer(&output.EmojiTransformer{}))
		}
		if outputConfig.UseColors {
			fileOptions = append(fileOptions, output.WithTransformer(output.NewColorTransformer()))
		}

		fileOut := output.NewOutput(fileOptions...)
		if err := fileOut.Render(ctx, doc); err != nil {
			return fmt.Errorf("failed to render to file: %w", err)
		}
	}

	return nil
}

// getFormatFromConfig converts string format to v2 Format
func (f *Formatter) getFormatFromConfig(format string) output.Format {
	switch strings.ToLower(format) {
	case "json":
		return output.JSON
	case "csv":
		return output.CSV
	case "html":
		return output.HTML
	case "markdown":
		return output.Markdown
	case "table":
		return output.Table
	default:
		return output.Table
	}
}

// createPlanInfoDataV2 creates the plan information data for v2 API
func (f *Formatter) createPlanInfoDataV2(summary *PlanSummary) ([]map[string]any, error) {
	if summary == nil {
		return nil, fmt.Errorf("summary cannot be nil")
	}

	data := []map[string]any{
		{
			"Plan File": summary.PlanFile,
			"Version":   summary.TerraformVersion,
			"Workspace": summary.Workspace,
			"Backend":   fmt.Sprintf("%s (%s)", summary.Backend.Type, summary.Backend.Location),
			"Created":   summary.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}

	return data, nil
}

// createStatisticsSummaryDataV2 creates the statistics summary data for v2 API
func (f *Formatter) createStatisticsSummaryDataV2(summary *PlanSummary) ([]map[string]any, error) {
	if summary == nil {
		return nil, fmt.Errorf("summary cannot be nil")
	}
	if summary.PlanFile == "" {
		return nil, fmt.Errorf("plan file name is required")
	}

	data := []map[string]any{
		{
			"TOTAL":        summary.Statistics.Total,
			"ADDED":        summary.Statistics.ToAdd,
			"REMOVED":      summary.Statistics.ToDestroy,
			"MODIFIED":     summary.Statistics.ToChange,
			"REPLACEMENTS": summary.Statistics.Replacements,
			"CONDITIONALS": summary.Statistics.Conditionals,
			"HIGH RISK":    summary.Statistics.HighRisk,
		},
	}

	return data, nil
}

// createResourceChangesDataV2 creates the resource changes data for v2 API
func (f *Formatter) createResourceChangesDataV2(summary *PlanSummary) ([]map[string]any, error) {
	if summary == nil {
		return nil, fmt.Errorf("summary cannot be nil")
	}

	var data []map[string]any

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

		data = append(data, map[string]any{
			"ACTION":      getActionDisplay(change.ChangeType),
			"RESOURCE":    change.Address,
			"TYPE":        change.Type,
			"ID":          displayID,
			"REPLACEMENT": replacementDisplay,
			"MODULE":      change.ModulePath,
			"DANGER":      dangerInfo,
		})
	}

	return data, nil
}

// createSensitiveResourceChangesDataV2 creates data for sensitive resource changes only for v2 API
func (f *Formatter) createSensitiveResourceChangesDataV2(summary *PlanSummary) ([]map[string]any, error) {
	if summary == nil {
		return nil, fmt.Errorf("summary cannot be nil")
	}

	// Filter for sensitive resources
	var data []map[string]any

	for _, change := range summary.ResourceChanges {
		if !change.IsDangerous {
			continue
		}

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

		data = append(data, map[string]any{
			"ACTION":      getActionDisplay(change.ChangeType),
			"RESOURCE":    change.Address,
			"TYPE":        change.Type,
			"ID":          displayID,
			"REPLACEMENT": replacementDisplay,
			"MODULE":      change.ModulePath,
			"DANGER":      dangerInfo,
		})
	}

	return data, nil
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
