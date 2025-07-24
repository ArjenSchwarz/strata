package plan

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	output "github.com/ArjenSchwarz/go-output/v2"
	"github.com/ArjenSchwarz/strata/config"
)

const (
	notApplicable = "N/A"
	formatTable   = "table"
)

// Formatter handles different output formats for plan summaries
type Formatter struct {
	config *config.Config
}

// ActionSortTransformer sorts table data based on Terraform action priority
type ActionSortTransformer struct{}

// Name implements the output.Transformer interface
func (t *ActionSortTransformer) Name() string {
	return "ActionSort"
}

// Priority implements the output.Transformer interface
func (t *ActionSortTransformer) Priority() int {
	return 100
}

// CanTransform implements the output.Transformer interface
func (t *ActionSortTransformer) CanTransform(format string) bool {
	return format == output.FormatTable || format == output.FormatMarkdown || format == output.FormatHTML || format == output.FormatCSV
}

// Transform implements the output.Transformer interface
func (t *ActionSortTransformer) Transform(ctx context.Context, input []byte, format string) ([]byte, error) {
	content := string(input)

	// Check if this is a Resource Changes table by looking for the table headers
	if !strings.Contains(content, "Resource Changes") && !strings.Contains(content, "Sensitive Resource Changes") {
		return input, nil
	}

	// Find table rows using regex to match ACTION column patterns
	lines := strings.Split(content, "\n")
	var tableStart, tableEnd int
	var dataRows []string
	var headerSeparatorIndex int

	for i, line := range lines {
		if strings.Contains(line, "Resource Changes") {
			tableStart = i
		}
		if strings.Contains(line, "ACTION") && strings.Contains(line, "RESOURCE") {
			headerSeparatorIndex = i + 1 // Skip the header separator line
			continue
		}
		// Look for data rows (contain action verbs)
		if tableStart > 0 && i > headerSeparatorIndex &&
			(strings.Contains(line, "Add") || strings.Contains(line, "Remove") ||
				strings.Contains(line, "Replace") || strings.Contains(line, "Modify")) {
			dataRows = append(dataRows, line)
		}
		// Find end of table (empty line or next section)
		if tableStart > 0 && i > headerSeparatorIndex && strings.TrimSpace(line) == "" && len(dataRows) > 0 {
			tableEnd = i
			break
		}
	}

	if len(dataRows) == 0 {
		return input, nil
	}

	// Sort the data rows by danger status first, then action priority
	sort.Slice(dataRows, func(i, j int) bool {
		dangerI := isDangerousRow(dataRows[i])
		dangerJ := isDangerousRow(dataRows[j])

		// If one is dangerous and the other isn't, dangerous comes first
		if dangerI != dangerJ {
			return dangerI
		}

		// If both have same danger status, sort by action priority
		actionI := extractActionFromTableRow(dataRows[i])
		actionJ := extractActionFromTableRow(dataRows[j])

		priorityI := getActionSortPriority(actionI)
		priorityJ := getActionSortPriority(actionJ)

		return priorityI < priorityJ
	})

	// Reconstruct the content with sorted rows
	var result []string
	result = append(result, lines[:headerSeparatorIndex+1]...)
	result = append(result, dataRows...)
	if tableEnd > 0 {
		result = append(result, lines[tableEnd:]...)
	}

	return []byte(strings.Join(result, "\n")), nil
}

// extractActionFromTableRow extracts the action from a table row
func extractActionFromTableRow(row string) string {
	// Use regex to find action words at the beginning of table cells
	re := regexp.MustCompile(`^\s*\|\s*(Add|Remove|Replace|Modify)\s*\|`)
	matches := re.FindStringSubmatch(row)
	if len(matches) > 1 {
		return matches[1]
	}
	// Fallback: look for action words anywhere in the row
	actions := []string{"Remove", "Replace", "Modify", "Add"}
	for _, action := range actions {
		if strings.Contains(row, action) {
			return action
		}
	}
	return "Unknown"
}

// isDangerousRow checks if a table row represents a dangerous/sensitive resource
func isDangerousRow(row string) bool {
	// Look for danger indicators in the DANGER column (last column typically)
	// Check for warning emoji or danger-related text
	return strings.Contains(row, "⚠️") ||
		strings.Contains(row, "Sensitive") ||
		strings.Contains(row, "Dangerous") ||
		strings.Contains(row, "High Risk") ||
		// Look for non-empty DANGER column (anything after the last | that's not just whitespace)
		regexp.MustCompile(`\|\s*[^|\s]+\s*$`).MatchString(strings.TrimSpace(row))
}

// getActionSortPriority returns priority for sorting (lower = higher priority)
func getActionSortPriority(action string) int {
	switch action {
	case "Remove":
		return 1
	case "Replace":
		return 2
	case "Modify":
		return 3
	case "Add":
		return 4
	default:
		return 5
	}
}

// NewFormatter creates a new formatter instance
func NewFormatter(cfg *config.Config) *Formatter {
	return &Formatter{
		config: cfg,
	}
}

// ValidateOutputFormat validates that the output format is supported
func (f *Formatter) ValidateOutputFormat(outputFormat string) error {
	supportedFormats := []string{formatTable, "json", "html", "markdown"}
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
		output.WithKeys("TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED"))

	// Add resource changes table if requested
	if showDetails {
		resourceData, err := f.createResourceChangesDataV2(summary)
		if err != nil {
			return fmt.Errorf("failed to create resource changes data: %w", err)
		}
		if len(resourceData) > 0 {
			builder = builder.Table("Resource Changes", resourceData,
				output.WithKeys("ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"))
		} else {
			builder = builder.Text("All resources are unchanged.")
		}
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
		} else if outputConfig.OutputFile == "" {
			// Handle no sensitive changes case - but only for stdout-only mode
			builder = builder.Text("No sensitive resource changes detected.")
		}
	}

	// Build the document
	doc := builder.Build()

	// Render to stdout first
	stdoutFormat := f.getFormatFromConfig(outputConfig.Format)
	if outputConfig.TableStyle != "" && outputConfig.Format == formatTable {
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
	// Add action sorting transformer for supported formats
	stdoutOptions = append(stdoutOptions, output.WithTransformer(&ActionSortTransformer{}))

	stdoutOut := output.NewOutput(stdoutOptions...)
	if err := stdoutOut.Render(ctx, doc); err != nil {
		return fmt.Errorf("failed to render to stdout: %w", err)
	}

	// Render to file if configured
	if outputConfig.OutputFile != "" {
		fileWriter, err := output.NewFileWriterWithOptions(".", outputConfig.OutputFile, output.WithAbsolutePaths())
		if err != nil {
			return fmt.Errorf("failed to create file writer: %w", err)
		}

		fileFormat := f.getFormatFromConfig(outputConfig.OutputFileFormat)
		if outputConfig.TableStyle != "" && outputConfig.OutputFileFormat == formatTable {
			fileFormat = output.TableWithStyle(outputConfig.TableStyle)
		}
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
		// Add action sorting transformer for supported formats
		fileOptions = append(fileOptions, output.WithTransformer(&ActionSortTransformer{}))

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
	case formatTable:
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
			"TOTAL CHANGES": summary.Statistics.Total,
			"ADDED":         summary.Statistics.ToAdd,
			"REMOVED":       summary.Statistics.ToDestroy,
			"MODIFIED":      summary.Statistics.ToChange,
			"REPLACEMENTS":  summary.Statistics.Replacements,
			"HIGH RISK":     summary.Statistics.HighRisk,
			"UNMODIFIED":    summary.Statistics.Unmodified,
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
		// Skip no-op changes from details
		if change.ChangeType == ChangeTypeNoOp {
			continue
		}

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
			replacementDisplay = notApplicable
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

		// Skip no-op changes from details
		if change.ChangeType == ChangeTypeNoOp {
			continue
		}

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
			replacementDisplay = notApplicable
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
