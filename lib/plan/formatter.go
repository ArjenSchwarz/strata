package plan

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	output "github.com/ArjenSchwarz/go-output/v2"
	"github.com/ArjenSchwarz/strata/config"
)

const (
	notApplicable       = "N/A"
	formatTable         = "table"
	noPropertiesChanged = "No properties changed"
	truncatedIndicator  = " [truncated]"
	// Unicode En space (U+2002) constants for consistent indentation across output formats
	indent       = "\u2002\u2002"             // 2 En spaces - basic indentation level
	nestedIndent = "\u2002\u2002\u2002\u2002" // 4 En spaces - nested indentation level

	// Action constants for table sorting
	tableActionRemove  = "Remove"
	tableActionReplace = "Replace"
	tableActionModify  = "Modify"
	tableActionAdd     = "Add"

	// Known after apply constant
	knownAfterApply = "(known after apply)"
)

// Cached regex patterns for ActionSortTransformer performance optimization
var (
	// Matches action words at the beginning of table cells
	actionStartRegex = regexp.MustCompile(`^\s*\|\s*(Add|Remove|Replace|Modify)\s*\|`)
	// Matches actions with emoji prefix (like "⚠️ Remove")
	actionEmojiRegex = regexp.MustCompile(`^\s*\|\s*[^|]*\s*(Add|Remove|Replace|Modify)\s*\|`)
	// Matches non-empty DANGER column (anything after the last | that's not just whitespace)
	dangerColumnRegex = regexp.MustCompile(`\|\s*[^|\s]+\s*$`)
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
	return format == output.Table.Name || format == output.Markdown.Name || format == output.HTML.Name || format == output.CSV.Name
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
	var tableStart = -1
	var dataRows []string
	var dataRowIndices []int
	headerFound := false
	inResourceTable := false

	for i, line := range lines {
		// Look for Resource Changes header
		if strings.Contains(line, "Resource Changes") {
			tableStart = i
			inResourceTable = true
			continue
		}

		// Look for table header with ACTION column
		if inResourceTable && !headerFound && strings.Contains(line, "| action") && strings.Contains(line, "| resource") {
			headerFound = true
			continue
		}

		// Skip separator line (| --- | --- | ...)
		if inResourceTable && headerFound && strings.Contains(line, "| ---") {
			continue
		}

		// Look for data rows in the Resource Changes table
		if inResourceTable && headerFound && strings.HasPrefix(strings.TrimSpace(line), "|") &&
			(strings.Contains(line, "Add") || strings.Contains(line, "Remove") ||
				strings.Contains(line, "Replace") || strings.Contains(line, "Modify")) {
			dataRows = append(dataRows, line)
			dataRowIndices = append(dataRowIndices, i)
		}

		// Check for end of table (empty line or new section header)
		if inResourceTable && headerFound && len(dataRows) > 0 {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "###") {
				break
			}
		}
	}

	if len(dataRows) == 0 || tableStart == -1 {
		return input, nil
	}

	// Sort the data rows by danger status first, then action priority
	sortedIndices := make([]int, len(dataRows))
	for i := range sortedIndices {
		sortedIndices[i] = i
	}

	// Enhanced sorting for Task 6.2 from Output Refinements feature
	// Sort by: 1) danger indicators, 2) action priority, 3) alphabetically
	sort.Slice(sortedIndices, func(i, j int) bool {
		rowI := dataRows[sortedIndices[i]]
		rowJ := dataRows[sortedIndices[j]]

		// First: Check for danger indicators using enhanced method
		dangerI := hasDangerIndicator(rowI)
		dangerJ := hasDangerIndicator(rowJ)

		// If one is dangerous and the other isn't, dangerous comes first
		if dangerI != dangerJ {
			return dangerI
		}

		// Second: Sort by action priority (delete > replace > update > create)
		actionI := t.extractAction(rowI)
		actionJ := t.extractAction(rowJ)

		priorityI := t.getActionPriority(actionI)
		priorityJ := t.getActionPriority(actionJ)

		if priorityI != priorityJ {
			return priorityI < priorityJ
		}

		// Third: Alphabetical sort by resource address
		// Extract resource address from the row (typically second column)
		addressI := t.extractResourceAddress(rowI)
		addressJ := t.extractResourceAddress(rowJ)

		return addressI < addressJ
	})

	// Create a new lines array with sorted data rows
	newLines := make([]string, len(lines))
	copy(newLines, lines)

	// Replace the data rows with sorted versions
	for i, sortedIdx := range sortedIndices {
		newLines[dataRowIndices[i]] = dataRows[sortedIdx]
	}

	return []byte(strings.Join(newLines, "\n")), nil
}

// hasDangerIndicator checks if a table row contains danger indicators
// Enhanced for Task 6.1 from Output Refinements feature
// Refactored from ActionSortTransformer method to package function for Task 11.2
func hasDangerIndicator(row string) bool {
	// First check for explicit danger indicators in content
	// Be careful not to match "Add" in words like "address"
	if strings.Contains(row, "⚠️") ||
		strings.Contains(row, "Sensitive") ||
		strings.Contains(row, "Dangerous") ||
		strings.Contains(row, "High Risk") ||
		strings.Contains(row, "Critical") {
		return true
	}

	// Check for non-empty DANGER column (last column)
	// A table row must have at least 4 columns to have a danger column:
	// | action | resource | properties | danger |
	// When split, this becomes: ["", "action", "resource", "properties", "danger", ""]
	// So we need at least 5 parts for a danger column to exist
	parts := strings.Split(row, "|")

	// Only check for danger column if there are enough columns
	// Minimum table is: | action | resource | so 4 parts when split
	// With danger column: | action | resource | props | danger | so 6 parts
	if len(parts) >= 5 {
		// Get the last non-empty column index
		lastColIndex := len(parts) - 1
		if strings.TrimSpace(parts[lastColIndex]) == "" && lastColIndex > 0 {
			lastColIndex--
		}

		// The danger column would be at index 4 or higher (after action, resource, props)
		// Only check if we have enough columns for a danger column
		if lastColIndex >= 3 {
			lastCol := strings.TrimSpace(parts[lastColIndex])
			// Column is dangerous if it has content other than "-" or empty
			if lastCol != "" && lastCol != "-" {
				return true
			}
		}
	}

	// Don't use regex fallback if there are too few columns
	// This prevents false positives on short rows
	return false
}

// extractAction extracts the action from a table row
// Enhanced for Task 6.2 from Output Refinements feature
func (t *ActionSortTransformer) extractAction(row string) string {
	// Use cached regex to find action words at the beginning of table cells
	matches := actionStartRegex.FindStringSubmatch(row)
	if len(matches) > 1 {
		return matches[1]
	}
	// Also check for actions with emoji prefix (like "⚠️ Remove")
	matches2 := actionEmojiRegex.FindStringSubmatch(row)
	if len(matches2) > 1 {
		return matches2[1]
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

// getActionPriority returns priority for sorting (lower = higher priority)
// Enhanced for Task 6.2 from Output Refinements feature
func (t *ActionSortTransformer) getActionPriority(action string) int {
	// Map action priority: delete=0, replace=1, update=2, create=3, noop=4
	switch action {
	case tableActionRemove, "Delete":
		return 0 // Highest priority
	case tableActionReplace:
		return 1
	case tableActionModify, "Update":
		return 2
	case tableActionAdd, "Create":
		return 3
	default:
		return 4 // Lowest priority (including no-op)
	}
}

// extractResourceAddress extracts the resource address from a table row
// Typically the second column in the table
func (t *ActionSortTransformer) extractResourceAddress(row string) string {
	// Split by | and get the second column (resource address)
	parts := strings.Split(row, "|")
	if len(parts) >= 3 {
		// Index 0 is empty (before first |), index 1 is action, index 2 is resource
		return strings.TrimSpace(parts[2])
	}
	return ""
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
	lowercaseFormat := strings.ToLower(outputFormat)
	if slices.Contains(supportedFormats, lowercaseFormat) {
		return nil
	}
	return fmt.Errorf("unsupported output format '%s'. Supported formats: %s", outputFormat, strings.Join(supportedFormats, ", "))
}

// OutputSummary outputs the plan summary using go-output v2 library
//
// ARCHITECTURE: Simplified Plan Rendering
// This function implements the simplified rendering architecture that fixes the multi-table
// rendering bug by following the proven pattern from go-output v2's collapsible-tables example.
//
// Key architectural decisions:
// 1. Unified Table Creation: Uses output.NewTableContent() exclusively for consistent table creation
// 2. Unified Document Building: Uses output.New().AddContent().AddContent().Build() pattern
// 3. Format Handling Delegation: All format-specific logic delegated to go-output library
// 4. Conservative Error Handling: Individual table failures logged but don't stop rendering
//
// ENHANCEMENT: No-Op Filtering Integration (Task 4.3)
// This enhanced version applies no-op filtering based on f.config.Plan.ShowNoOps configuration
// and displays "No changes detected" message when no actual changes exist (Requirement 3.5)
//
// The previous implementation incorrectly disabled tables due to a perceived "library bug"
// that didn't actually exist. This approach re-enables all tables while maintaining
// all existing functionality including collapsible content and provider grouping.
func (f *Formatter) OutputSummary(summary *PlanSummary, outputConfig *config.OutputConfiguration, showDetails bool) error {
	// Handle nil summary gracefully
	if summary == nil {
		return fmt.Errorf("plan summary cannot be nil")
	}

	ctx := context.Background()

	// Validate output format first
	if err := f.ValidateOutputFormat(outputConfig.Format); err != nil {
		return err
	}

	// TASK 4.3: Apply filtering based on f.config.Plan.ShowNoOps configuration
	// Make a copy of summary to avoid modifying the original
	filteredSummary := *summary
	filteredSummary.ResourceChanges = f.filterNoOps(summary.ResourceChanges)
	filteredSummary.OutputChanges = f.filterNoOpOutputs(summary.OutputChanges)

	// TASK 4.3: Display "No changes detected" message when no actual changes exist (Requirement 3.5)
	if len(filteredSummary.ResourceChanges) == 0 && len(filteredSummary.OutputChanges) == 0 {
		builder := output.New()
		builder = builder.Text("No changes detected")
		doc := builder.Build()

		stdoutFormat := f.getFormatFromConfig(outputConfig.Format)
		stdoutOptions := []output.OutputOption{
			output.WithFormat(stdoutFormat),
			output.WithWriter(output.NewStdoutWriter()),
		}
		stdoutOut := output.NewOutput(stdoutOptions...)
		return stdoutOut.Render(ctx, doc)
	}

	// Build the document using v2 builder pattern
	builder := output.New()

	// Re-enable all tables using the proven NewTableContent pattern
	// This fixes the multi-table rendering issue by using consistent table creation methods

	// Plan Information table - RE-ENABLED using NewTableContent pattern
	planData, err := f.createPlanInfoDataV2(&filteredSummary)
	if err == nil && len(planData) > 0 {
		planTable, err := output.NewTableContent("Plan Information", planData,
			output.WithKeys("Plan File", "Version", "Workspace", "Backend", "Created"))
		if err == nil {
			builder = builder.AddContent(planTable)
		} else {
			// Log warning but continue operation - conservative error handling
			fmt.Printf("Warning: Failed to create plan information table: %v\n", err)
		}
	}

	// Summary Statistics table - RE-ENABLED using NewTableContent pattern
	// TASK 4.3: Ensure statistics remain unchanged and count all resources including no-ops (Requirement 3.7)
	// Use original summary for statistics to maintain count of all resources
	statsData, err := f.createStatisticsSummaryDataV2(summary)
	if err == nil && len(statsData) > 0 {
		statsTable, err := output.NewTableContent("Summary Statistics", statsData,
			output.WithKeys("Total Changes", "Added", "Removed", "Modified", "Replacements", "High Risk", "Unmodified"))
		if err == nil {
			builder = builder.AddContent(statsTable)
		} else {
			// Log warning but continue operation - conservative error handling
			fmt.Printf("Warning: Failed to create summary statistics table: %v\n", err)
		}
	}

	// Resource Changes table - UNIFIED TABLE CREATION following go-output example pattern
	// Use filtered summary for display
	if err := f.handleResourceDisplay(&filteredSummary, showDetails, outputConfig, builder); err != nil {
		return err
	}
	// If no conditions above are met, we show only Plan Information and Summary Statistics tables

	// Output Changes table - placed after resource changes section (requirement 2.1)
	// Use filtered summary for display
	if err := f.handleOutputDisplay(&filteredSummary, builder); err != nil {
		return err
	}

	// Unified document building using output.New().AddContent().Build() pattern
	doc := builder.Build()

	// Render to stdout first - unified format handling delegated to go-output
	stdoutFormat := f.getFormatFromConfig(outputConfig.Format)
	if outputConfig.TableStyle != "" && outputConfig.Format == formatTable {
		stdoutFormat = f.getCollapsibleTableFormat(outputConfig.TableStyle)
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
	actionSortTransformer := &ActionSortTransformer{}
	if actionSortTransformer.CanTransform(stdoutFormat.Name) {
		stdoutOptions = append(stdoutOptions, output.WithTransformer(actionSortTransformer))
	}

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
			fileFormat = f.getCollapsibleTableFormat(outputConfig.TableStyle)
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
		actionSortTransformer := &ActionSortTransformer{}
		if actionSortTransformer.CanTransform(fileFormat.Name) {
			fileOptions = append(fileOptions, output.WithTransformer(actionSortTransformer))
		}

		fileOut := output.NewOutput(fileOptions...)
		if err := fileOut.Render(ctx, doc); err != nil {
			return fmt.Errorf("failed to render to file: %w", err)
		}
	}

	return nil
}

// getFormatFromConfig converts string format to v2 Format with collapsible support
func (f *Formatter) getFormatFromConfig(format string) output.Format {
	rendererConfig := f.getRendererConfig()

	switch strings.ToLower(format) {
	case "json":
		return output.JSON // JSON doesn't support collapsible content
	case "csv":
		return output.Format{
			Name:     output.CSV.Name,
			Renderer: output.NewCSVRendererWithCollapsible(rendererConfig),
		}
	case "html":
		return output.Format{
			Name:     output.HTML.Name,
			Renderer: output.NewHTMLRendererWithCollapsible(rendererConfig),
		}
	case "markdown":
		return output.Format{
			Name:     output.Markdown.Name,
			Renderer: output.NewMarkdownRendererWithCollapsible(rendererConfig),
		}
	case formatTable:
		return output.Format{
			Name:     output.Table.Name,
			Renderer: output.NewTableRendererWithCollapsible("Default", rendererConfig),
		}
	default:
		return output.Format{
			Name:     output.Table.Name,
			Renderer: output.NewTableRendererWithCollapsible("Default", rendererConfig),
		}
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
			"Total Changes": summary.Statistics.Total,
			"Added":         summary.Statistics.ToAdd,
			"Removed":       summary.Statistics.ToDestroy,
			"Modified":      summary.Statistics.ToChange,
			"Replacements":  summary.Statistics.Replacements,
			"High Risk":     summary.Statistics.HighRisk,
			"Unmodified":    summary.Statistics.Unmodified,
		},
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
			"Action":      getActionDisplay(change.ChangeType),
			"Resource":    change.Address,
			"Type":        change.Type,
			"ID":          displayID,
			"Replacement": replacementDisplay,
			"Module":      change.ModulePath,
			"Danger":      dangerInfo,
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

// formatPropertyChangeDetails formats property changes for collapsible display
func (f *Formatter) formatPropertyChangeDetails(changes []PropertyChange) string {
	var details []string
	for _, change := range changes {
		if change.Sensitive {
			// Mask sensitive values
			details = append(details, fmt.Sprintf("• %s: (sensitive value) → (sensitive value)", change.Name))
		} else {
			// Show actual values for non-sensitive properties
			details = append(details, fmt.Sprintf("• %s: %v → %v", change.Name, change.Before, change.After))
		}
	}
	return strings.Join(details, "\n")
}

// propertyChangesFormatterTerraform creates a collapsible formatter that displays property changes in Terraform's diff-style format
func (f *Formatter) propertyChangesFormatterTerraform() func(any) any {
	return func(val any) any {
		// Handle the new map-based data structure
		if dataMap, ok := val.(map[string]any); ok {
			if analysis, hasAnalysis := dataMap["analysis"]; hasAnalysis {
				if propAnalysis, isPropAnalysis := analysis.(PropertyChangeAnalysis); isPropAnalysis {
					if propAnalysis.Count == 0 {
						return noPropertiesChanged
					}

					// Create summary
					summary := fmt.Sprintf("%d properties changed", propAnalysis.Count)
					if f.hasSensitive(propAnalysis.Changes) {
						summary = fmt.Sprintf("⚠️ %s (includes sensitive)", summary)
					}
					if propAnalysis.Truncated {
						summary += truncatedIndicator
					}

					// Format details in Terraform style
					var details []string
					for _, change := range propAnalysis.Changes {
						line := f.formatPropertyChange(change)
						details = append(details, line)
					}

					shouldExpand := (f.config.Plan.ExpandableSections.AutoExpandDangerous && f.hasSensitive(propAnalysis.Changes)) ||
						f.config.ExpandAll

					return output.NewCollapsibleValue(summary,
						strings.Join(details, "\n"),
						output.WithExpanded(shouldExpand),
						output.WithMaxLength(f.config.Plan.ExpandableSections.MaxDetailLength))
				}
			}
		}

		// Fallback for backward compatibility with direct PropertyChangeAnalysis
		if propAnalysis, ok := val.(PropertyChangeAnalysis); ok {
			if propAnalysis.Count == 0 {
				return noPropertiesChanged
			}

			// Create summary
			summary := fmt.Sprintf("%d properties changed", propAnalysis.Count)
			if f.hasSensitive(propAnalysis.Changes) {
				summary = fmt.Sprintf("⚠️ %s (includes sensitive)", summary)
			}
			if propAnalysis.Truncated {
				summary += truncatedIndicator
			}

			// Format details in Terraform style - use standard formatting without context
			var details []string
			for _, change := range propAnalysis.Changes {
				line := f.formatPropertyChange(change)
				details = append(details, line)
			}

			shouldExpand := (f.config.Plan.ExpandableSections.AutoExpandDangerous && f.hasSensitive(propAnalysis.Changes)) ||
				f.config.ExpandAll

			return output.NewCollapsibleValue(summary,
				strings.Join(details, "\n"),
				output.WithExpanded(shouldExpand),
				output.WithMaxLength(f.config.Plan.ExpandableSections.MaxDetailLength))
		}
		return val
	}
}

// formatPropertyChange formats a single property change in Terraform's diff-style format with optional context
func (f *Formatter) formatPropertyChange(change PropertyChange) string {
	var line string
	replacementIndicator := ""

	// Add replacement indicator if this change triggers replacement
	if change.TriggersReplacement {
		replacementIndicator = " # forces replacement"
	}

	// Check if we're dealing with complex nested values that should use nested formatting
	isComplexValue := func(val any) bool {
		switch v := val.(type) {
		case map[string]any:
			return len(v) > 1 // Use nested formatting for maps with multiple keys
		case []any:
			return len(v) > 2 // Use nested formatting for arrays with multiple items
		default:
			return false
		}
	}

	// Format based on action and handle complex values
	switch change.Action {
	case "add":
		if isComplexValue(change.After) {
			afterValue := f.formatValueWithContext(change.After, change.Sensitive, true, nestedIndent)
			line = fmt.Sprintf("%s+ %s = %s", indent, change.Name, afterValue)
		} else {
			line = fmt.Sprintf("%s+ %s = %s",
				indent, change.Name, f.formatValue(change.After, change.Sensitive))
		}
	case "remove":
		if isComplexValue(change.Before) {
			beforeValue := f.formatValueWithContext(change.Before, change.Sensitive, true, nestedIndent)
			line = fmt.Sprintf("%s- %s = %s", indent, change.Name, beforeValue)
		} else {
			line = fmt.Sprintf("%s- %s = %s",
				indent, change.Name, f.formatValue(change.Before, change.Sensitive))
		}
	case "update":
		// Check if this is a nested object change that should use nested formatting
		switch {
		case f.shouldUseNestedFormat(change.Before, change.After):
			line = f.formatNestedObjectChange(change)
		case isComplexValue(change.Before) || isComplexValue(change.After):
			beforeValue := f.formatValueWithContext(change.Before, change.Sensitive, true, nestedIndent)
			afterValue := f.formatValueWithContext(change.After, change.Sensitive, true, nestedIndent)
			line = fmt.Sprintf("%s~ %s = %s -> %s",
				indent, change.Name, beforeValue, afterValue)
		default:
			line = fmt.Sprintf("%s~ %s = %s -> %s",
				indent, change.Name,
				f.formatValue(change.Before, change.Sensitive),
				f.formatValue(change.After, change.Sensitive))
		}
	default:
		return ""
	}

	// Only add replacement indicator for non-nested formats
	// (nested formats handle this internally)
	if change.Action != "update" || !f.shouldUseNestedFormat(change.Before, change.After) {
		line += replacementIndicator
	}

	return line
}

// formatValue formats a property value according to Terraform's formatting conventions
func (f *Formatter) formatValue(val any, sensitive bool) string {
	return f.formatValueWithContext(val, sensitive, false, "")
}

// formatValueWithContext formats a property value with context awareness for nested structures
func (f *Formatter) formatValueWithContext(val any, sensitive bool, isNested bool, indent string) string {
	if sensitive {
		return "(sensitive value)"
	}

	// Handle different value types
	switch v := val.(type) {
	case string:
		// Check if this is the unknown value marker (requirement 1.3)
		if v == knownAfterApply {
			return v // Return without quotes to match Terraform's display
		}
		return fmt.Sprintf("%q", v)
	case map[string]any:
		if isNested && len(v) > 1 {
			// Format maps with proper indentation for nested display
			return f.formatNestedMap(v, indent)
		} else {
			// Format maps inline with sorted keys for consistent output (backward compatibility)
			var keys []string
			for key := range v {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			var pairs []string
			for _, key := range keys {
				pairs = append(pairs, fmt.Sprintf("%s = %s", key, f.formatValueWithContext(v[key], false, false, "")))
			}
			return fmt.Sprintf("{ %s }", strings.Join(pairs, ", "))
		}
	case []any:
		if isNested && len(v) > 2 {
			// Format arrays with proper indentation for nested display
			return f.formatNestedArray(v, indent)
		} else {
			// Format lists inline (backward compatibility)
			var items []string
			for _, item := range v {
				items = append(items, f.formatValueWithContext(item, false, false, ""))
			}
			return fmt.Sprintf("[ %s ]", strings.Join(items, ", "))
		}
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatNestedMap formats a map with proper indentation and line breaks
func (f *Formatter) formatNestedMap(v map[string]any, baseIndent string) string {
	var keys []string
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var lines []string
	lines = append(lines, "{")
	for _, key := range keys {
		// Use Unicode En spaces for indentation (U+2002) - preserves spacing without HTML escaping issues
		nextIndent := baseIndent + nestedIndent
		// Check if the value is complex (map or slice) to handle nested structures properly
		isValueNested := false
		switch v[key].(type) {
		case map[string]any, []any:
			isValueNested = true
		}
		value := f.formatValueWithContext(v[key], false, isValueNested, nextIndent)
		// Use Unicode En spaces for consistent indentation across all formats
		lines = append(lines, fmt.Sprintf("%s%s%s = %s", baseIndent, nestedIndent, key, value))
	}
	lines = append(lines, baseIndent+"}")
	return strings.Join(lines, "\n")
}

// formatNestedArray formats an array with proper indentation and line breaks
func (f *Formatter) formatNestedArray(v []any, baseIndent string) string {
	var lines []string
	lines = append(lines, "[")
	for i, item := range v {
		// Use Unicode En spaces for indentation (U+2002) - preserves spacing without HTML escaping issues
		nextIndent := baseIndent + indent
		// Check if the item is complex (map or slice) to handle nested structures properly
		isItemNested := false
		switch item.(type) {
		case map[string]any, []any:
			isItemNested = true
		}
		value := f.formatValueWithContext(item, false, isItemNested, nextIndent)
		// Use Unicode En spaces for consistent indentation across all formats
		lines = append(lines, fmt.Sprintf("%s%s[%d] = %s", baseIndent, indent, i, value))
	}
	lines = append(lines, baseIndent+"]")
	return strings.Join(lines, "\n")
}

// hasSensitive checks if any changes in the list contain sensitive properties
func (f *Formatter) hasSensitive(changes []PropertyChange) bool {
	for _, change := range changes {
		if change.Sensitive {
			return true
		}
	}
	return false
}

// shouldUseNestedFormat determines if a property change should use the nested format
// instead of showing complete before/after values
func (f *Formatter) shouldUseNestedFormat(before, after any) bool {
	// Check if both values are maps (nested objects)
	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)

	// Only use nested format for map-to-map changes
	if !beforeIsMap || !afterIsMap {
		return false
	}

	// Use nested format if both maps have content (avoid for completely empty objects)
	return len(beforeMap) > 0 || len(afterMap) > 0
}

// formatNestedObjectChange formats a nested object change in Terraform-style diff format
func (f *Formatter) formatNestedObjectChange(change PropertyChange) string {
	beforeMap, _ := change.Before.(map[string]any)
	afterMap, _ := change.After.(map[string]any)

	// Get all unique keys from both maps
	allKeys := make(map[string]bool)
	for key := range beforeMap {
		allKeys[key] = true
	}
	for key := range afterMap {
		allKeys[key] = true
	}

	// Sort keys for consistent output
	var keys []string
	for key := range allKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var lines []string

	// Add the opening line with the property name
	replacementIndicator := ""
	if change.TriggersReplacement {
		replacementIndicator = " # forces replacement"
	}
	// Use Unicode En spaces (U+2002) for consistent spacing across formats
	lines = append(lines, fmt.Sprintf("%s~ %s {%s", indent, change.Name, replacementIndicator))

	// Process each key
	for _, key := range keys {
		beforeValue, hasBeforeValue := beforeMap[key]
		afterValue, hasAfterValue := afterMap[key]

		switch {
		case !hasBeforeValue && hasAfterValue:
			// Added property - use Unicode En spaces for indentation
			formattedValue := f.formatValue(afterValue, change.Sensitive)
			lines = append(lines, fmt.Sprintf("%s+ %s = %s", nestedIndent, key, formattedValue))
		case hasBeforeValue && !hasAfterValue:
			// Removed property - use Unicode En spaces for indentation
			formattedValue := f.formatValue(beforeValue, change.Sensitive)
			lines = append(lines, fmt.Sprintf("%s- %s = %s", nestedIndent, key, formattedValue))
		case hasBeforeValue && hasAfterValue:
			// Check if the value actually changed
			if !f.valuesEqual(beforeValue, afterValue) {
				// Modified property - use Unicode En spaces for indentation
				beforeFormatted := f.formatValue(beforeValue, change.Sensitive)
				afterFormatted := f.formatValue(afterValue, change.Sensitive)
				lines = append(lines, fmt.Sprintf("%s~ %s = %s -> %s", nestedIndent, key, beforeFormatted, afterFormatted))
			}
		}
	}

	// Add the closing brace with Unicode En spaces to match the opening
	lines = append(lines, indent+"}")

	return strings.Join(lines, "\n")
}

// valuesEqual compares two values for equality
func (f *Formatter) valuesEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// prepareResourceTableData transforms ResourceChange data for go-output v2 table display with collapsible content
// This function filters out no-op changes to implement empty table suppression (requirement 1)
func (f *Formatter) prepareResourceTableData(changes []ResourceChange) []map[string]any {
	tableData := make([]map[string]any, 0, len(changes))

	for _, change := range changes {
		// Skip no-op changes from details (requirement 1: Empty Table Suppression)
		if change.ChangeType == ChangeTypeNoOp {
			continue
		}

		// Use the property changes from the analyzer
		propChanges := change.PropertyChanges

		// Determine risk level based on existing danger flags
		riskLevel := "low"
		if change.IsDangerous {
			switch change.ChangeType {
			case ChangeTypeDelete:
				riskLevel = "critical"
			case ChangeTypeReplace:
				riskLevel = "high"
			default:
				riskLevel = "medium"
			}
		}

		// Determine action display
		actionDisplay := getActionDisplay(change.ChangeType)
		if change.IsDangerous {
			actionDisplay = "⚠️ " + actionDisplay
		}

		// Store change type alongside property changes for context-aware formatting
		propertyChangesData := map[string]any{
			"analysis":    propChanges,
			"change_type": change.ChangeType,
			"properties":  propChanges.Changes, // Include raw property changes for JSON access
		}

		row := map[string]any{
			"Action":           actionDisplay,
			"Resource":         change.Address,
			"Type":             change.Type,
			"ID":               f.getDisplayID(change),
			"Replacement":      f.getReplacementDisplay(change),
			"Module":           change.ModulePath,
			"Danger":           f.getDangerDisplay(change),
			"risk_level":       riskLevel,
			"Property Changes": propertyChangesData, // Will be formatted by collapsible formatter
		}

		// Add replacement reasons if available
		if len(change.ReplacementHints) > 0 {
			row["replacement_reasons"] = strings.Join(change.ReplacementHints, ", ")
		}

		tableData = append(tableData, row)
	}

	return tableData
}

// countChangedResources counts resources excluding no-ops for provider grouping threshold calculations
// This implements requirement 1.4: threshold comparison uses total changed resources, not total resources
func (f *Formatter) countChangedResources(changes []ResourceChange) int {
	count := 0
	for _, change := range changes {
		if change.ChangeType != ChangeTypeNoOp {
			count++
		}
	}
	return count
}

// getDisplayID returns the appropriate ID for display based on change type
func (f *Formatter) getDisplayID(change ResourceChange) string {
	switch change.ChangeType {
	case ChangeTypeCreate:
		return "-"
	case ChangeTypeDelete:
		return change.PhysicalID
	default:
		return change.PhysicalID
	}
}

// getReplacementDisplay returns the replacement display string
func (f *Formatter) getReplacementDisplay(change ResourceChange) string {
	if change.ChangeType == ChangeTypeDelete {
		return notApplicable
	}
	return string(change.ReplacementType)
}

// getDangerDisplay returns the danger information for display
func (f *Formatter) getDangerDisplay(change ResourceChange) string {
	if !change.IsDangerous {
		return ""
	}

	dangerInfo := "⚠️ " + change.DangerReason
	if len(change.DangerProperties) > 0 {
		dangerInfo += ": " + strings.Join(change.DangerProperties, ", ")
	}
	return dangerInfo
}

// addResourceChangesWithProgressiveDisclosure adds resource changes with collapsible features to existing builder
func (f *Formatter) addResourceChangesWithProgressiveDisclosure(builder *output.Builder, summary *PlanSummary) *output.Builder {
	// Create table with collapsible formatters for detailed information
	if len(summary.ResourceChanges) > 0 {
		// Apply priority sorting before preparing table data (Requirements 2.1, 2.2, 2.3)
		sortedResources := f.sortResourcesByPriority(summary.ResourceChanges)
		tableData := f.prepareResourceTableData(sortedResources)

		// Use NewTableContent consistently to match working example pattern
		schema := f.getResourceTableSchema()
		resourceTable, err := output.NewTableContent("Resource Changes", tableData,
			output.WithSchema(schema...))
		if err == nil {
			builder = builder.AddContent(resourceTable)
		} else {
			// Log warning but continue operation - conservative error handling
			fmt.Printf("Warning: Failed to create resource changes table: %v\n", err)
		}
	} else {
		builder = builder.Text("All resources are unchanged.")
	}
	return builder
}

// Backwards compatibility wrappers for tests
func (f *Formatter) formatResourceChangesWithProgressiveDisclosure(summary *PlanSummary) (*output.Document, error) {
	builder := output.New()
	// Add plan information section
	planInfoData, err := f.createPlanInfoDataV2(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan info data: %w", err)
	}
	// Use NewTableContent for consistency
	planTable, err := output.NewTableContent("Plan Information", planInfoData,
		output.WithKeys("Plan File", "Version", "Workspace", "Backend", "Created"))
	if err == nil {
		builder = builder.AddContent(planTable)
	}
	// Add statistics summary section
	statsData, err := f.createStatisticsSummaryDataV2(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to create statistics summary data: %w", err)
	}
	statsTable, err := output.NewTableContent(fmt.Sprintf("Summary for %s", summary.PlanFile), statsData,
		output.WithKeys("Total Changes", "Added", "Removed", "Modified", "Replacements", "High Risk", "Unmodified"))
	if err == nil {
		builder = builder.AddContent(statsTable)
	}
	builder = f.addResourceChangesWithProgressiveDisclosure(builder, summary)
	return builder.Build(), nil
}

func (f *Formatter) formatGroupedWithCollapsibleSections(summary *PlanSummary, groups map[string][]ResourceChange) (*output.Document, error) {
	builder := output.New()
	// Add plan information section
	planInfoData, err := f.createPlanInfoDataV2(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan info data: %w", err)
	}
	// Use NewTableContent for consistency
	planTable, err := output.NewTableContent("Plan Information", planInfoData,
		output.WithKeys("Plan File", "Version", "Workspace", "Backend", "Created"))
	if err == nil {
		builder = builder.AddContent(planTable)
	}
	// Add statistics summary section
	statsData, err := f.createStatisticsSummaryDataV2(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to create statistics summary data: %w", err)
	}
	statsTable, err := output.NewTableContent(fmt.Sprintf("Summary for %s", summary.PlanFile), statsData,
		output.WithKeys("Total Changes", "Added", "Removed", "Modified", "Replacements", "High Risk", "Unmodified"))
	if err == nil {
		builder = builder.AddContent(statsTable)
	}
	builder = f.addGroupedResourceChangesWithCollapsibleSections(builder, groups)
	return builder.Build(), nil
}

// addGroupedResourceChangesWithCollapsibleSections adds grouped resource changes with collapsible sections to existing builder
//
// TASK 5.2 FIX: This function now uses CollapsibleSection instead of Section to enable
// auto-expansion behavior for high-risk changes within provider groups (Requirement 6.4).
// Provider sections will auto-expand when they contain dangerous deletions or replacements.
func (f *Formatter) addGroupedResourceChangesWithCollapsibleSections(builder *output.Builder, groups map[string][]ResourceChange) *output.Builder {
	// Create collapsible sections for each provider with auto-expansion for high-risk changes
	for provider, resources := range groups {
		if len(resources) == 0 {
			continue
		}

		// Apply priority sorting within this provider group (Requirement 2.4)
		sortedResources := f.sortResourcesByPriority(resources)
		// Prepare table data for this provider's resources
		tableData := f.prepareResourceTableData(sortedResources)
		schema := f.getResourceTableSchema()

		// Determine if this provider section should auto-expand based on high-risk changes
		// Auto-expand when AutoExpandDangerous is enabled and provider has high-risk changes
		shouldExpandProvider := f.config.Plan.ExpandableSections.AutoExpandDangerous && f.hasHighRiskChanges(resources)

		// Override expansion if global ExpandAll is enabled
		if f.config.ExpandAll {
			shouldExpandProvider = true
		}

		// Add collapsible section using builder's CollapsibleSection method with NewTableContent pattern
		// This enables auto-expansion behavior for high-risk changes within provider groups
		builder = builder.CollapsibleSection(
			fmt.Sprintf("%s Provider (%d changes)", strings.ToUpper(provider), len(resources)),
			func(b *output.Builder) {
				providerTable, err := output.NewTableContent(
					fmt.Sprintf("%s Resources", strings.ToUpper(provider)),
					tableData,
					output.WithSchema(schema...),
				)
				if err == nil {
					b.AddContent(providerTable)
				} else {
					// Log warning but continue operation - conservative error handling
					fmt.Printf("Warning: Failed to create %s provider table: %v\n", provider, err)
				}
			},
			output.WithSectionExpanded(shouldExpandProvider),
		)
	}

	return builder
}

// getResourceTableSchema returns the schema configuration for resource tables
func (f *Formatter) getResourceTableSchema() []output.Field {
	return []output.Field{
		{
			Name: "Action",
			Type: "string",
		},
		{
			Name:      "Resource",
			Type:      "string",
			Formatter: output.FilePathFormatter(50),
		},
		{
			Name: "Type",
			Type: "string",
		},
		{
			Name: "ID",
			Type: "string",
		},
		{
			Name: "Replacement",
			Type: "string",
		},
		{
			Name: "Module",
			Type: "string",
		},
		{
			Name: "Danger",
			Type: "string",
		},
		{
			Name:      "Property Changes",
			Type:      "object",
			Formatter: f.propertyChangesFormatterTerraform(),
		},
	}
}

// propertyChangesFormatterDirect creates a collapsible formatter that returns NewCollapsibleValue directly
//
// DESIGN DECISION: This "Direct" version was kept during code consolidation because it properly
// integrates with the go-output v2 NewTableContent pattern. The previous non-Direct version
// had subtle differences that caused integration issues with the unified table creation approach.
func (f *Formatter) propertyChangesFormatterDirect() func(any) any {
	return func(val any) any {
		if propAnalysis, ok := val.(PropertyChangeAnalysis); ok {
			if propAnalysis.Count > 0 {
				// Create summary showing count and highlighting sensitive properties
				sensitiveCount := 0
				for _, change := range propAnalysis.Changes {
					if change.Sensitive {
						sensitiveCount++
					}
				}

				summary := fmt.Sprintf("%d properties changed", propAnalysis.Count)
				if sensitiveCount > 0 {
					summary = fmt.Sprintf("⚠️ %d properties changed (%d sensitive)", propAnalysis.Count, sensitiveCount)
				}
				if propAnalysis.Truncated {
					summary += truncatedIndicator
				}

				// Create detailed content
				details := f.formatPropertyChangeDetails(propAnalysis.Changes)

				// Auto-expand if sensitive properties are present and AutoExpandDangerous is enabled
				// Note: ExpandAll is handled by renderer's ForceExpansion, not here
				shouldExpand := f.config.Plan.ExpandableSections.AutoExpandDangerous && sensitiveCount > 0

				return output.NewCollapsibleValue(summary, details,
					output.WithExpanded(shouldExpand),
					output.WithMaxLength(f.config.Plan.ExpandableSections.MaxDetailLength))
			} else {
				// No properties changed - return simple string
				return noPropertiesChanged
			}
		}
		// Return input unchanged for non-PropertyChangeAnalysis types (required for test compatibility)
		return val
	}
}

// hasHighRiskChanges checks if any resources in the list are high risk
func (f *Formatter) hasHighRiskChanges(resources []ResourceChange) bool {
	for _, resource := range resources {
		// Auto-expand for critical or high risk changes
		if resource.IsDangerous && (resource.ChangeType == ChangeTypeDelete || resource.ChangeType == ChangeTypeReplace) {
			return true
		}
	}
	return false
}

// groupResourcesByProvider groups resources by their provider
// This function excludes no-ops from grouping (requirement 1.2: provider-specific tables don't include no-ops)
func (f *Formatter) groupResourcesByProvider(changes []ResourceChange) map[string][]ResourceChange {
	groups := make(map[string][]ResourceChange)
	for _, change := range changes {
		// Skip no-ops from grouping (requirement 1.2)
		if change.ChangeType == ChangeTypeNoOp {
			continue
		}

		provider := change.Provider
		if provider == "" {
			// Extract provider from resource type (e.g., "aws_instance" -> "aws")
			parts := strings.Split(change.Type, "_")
			if len(parts) > 0 {
				provider = parts[0]
			} else {
				provider = "unknown"
			}
		}
		groups[provider] = append(groups[provider], change)
	}
	return groups
}

// shouldAutoExpandProvider determines if a provider group should be auto-expanded based on risk level
func (f *Formatter) shouldAutoExpandProvider(resources []ResourceChange) bool {
	// Auto-expand if any resource in the group is dangerous or high-risk
	for _, resource := range resources {
		if resource.IsDangerous {
			return true
		}
	}
	return false
}

// filterSensitiveChanges returns only the changes marked as dangerous/sensitive
func (f *Formatter) filterSensitiveChanges(changes []ResourceChange) []ResourceChange {
	var sensitive []ResourceChange
	for _, change := range changes {
		if change.IsDangerous {
			sensitive = append(sensitive, change)
		}
	}
	return sensitive
}

// getCollapsibleTableFormat returns collapsible-enabled table format with specific style
func (f *Formatter) getCollapsibleTableFormat(style string) output.Format {
	rendererConfig := f.getRendererConfig()
	return output.Format{
		Name:     output.Table.Name,
		Renderer: output.NewTableRendererWithCollapsible(style, rendererConfig),
	}
}

// getRendererConfig creates renderer configuration with collapsible settings
func (f *Formatter) getRendererConfig() output.RendererConfig {
	return output.RendererConfig{
		ForceExpansion:       f.config.ExpandAll,
		MaxDetailLength:      f.config.Plan.ExpandableSections.MaxDetailLength,
		TruncateIndicator:    "... [truncated]",
		TableHiddenIndicator: "[expand for details]",
		HTMLCSSClasses: map[string]string{
			"details": "strata-collapsible",
			"summary": "strata-summary",
			"content": "strata-details",
		},
	}
}

// addResourceChangesTable handles the creation of resource changes tables with proper grouping logic
func (f *Formatter) addResourceChangesTable(summary *PlanSummary, builder *output.Builder) {
	// Check if provider grouping should be used (requirement 1.4: use changed resource count for threshold)
	changedResourceCount := f.countChangedResources(summary.ResourceChanges)
	shouldGroup := f.config.Plan.Grouping.Enabled && changedResourceCount >= f.config.Plan.Grouping.Threshold

	switch {
	case shouldGroup:
		f.addGroupedResourceTables(summary, builder)
	default:
		f.addStandardResourceTable(summary, builder)
	}
}

// addGroupedResourceTables creates provider-grouped resource tables
func (f *Formatter) addGroupedResourceTables(summary *PlanSummary, builder *output.Builder) {
	groups := f.groupResourcesByProvider(summary.ResourceChanges)
	if len(groups) > 1 {
		// Multiple providers: create provider-grouped sections
		for providerName, resources := range groups {
			f.addProviderGroupTable(providerName, resources, builder)
		}
	} else {
		// Single provider: create standard table
		f.addStandardResourceTable(summary, builder)
	}
}

// addProviderGroupTable creates a table for a specific provider group
func (f *Formatter) addProviderGroupTable(providerName string, resources []ResourceChange, builder *output.Builder) {
	// Apply priority sorting within this provider group (Requirement 2.4)
	sortedResources := f.sortResourcesByPriority(resources)
	groupData := f.prepareResourceTableData(sortedResources)
	// Requirement 1.1: Only create table if data exists after filtering no-ops
	if len(groupData) > 0 {
		schema := f.getResourceTableSchema()
		// Create table for this provider group
		providerTable, err := output.NewTableContent(fmt.Sprintf("%s Resources", strings.ToUpper(providerName)), groupData,
			output.WithSchema(schema...))
		if err == nil {
			// Create collapsible section for this provider (requirement 1.3: show only changed resources in count)
			changedCount := f.countChangedResources(resources)
			providerSection := output.NewCollapsibleSection(
				fmt.Sprintf("%s Provider (%d changes)", strings.ToUpper(providerName), changedCount),
				[]output.Content{providerTable},
				output.WithSectionExpanded(f.shouldAutoExpandProvider(resources)),
				output.WithSectionLevel(2),
			)
			builder.AddContent(providerSection)
		} else {
			fmt.Printf("Warning: Failed to create %s provider table: %v\n", providerName, err)
		}
	}
	// If groupData is empty, table is suppressed (requirement 1.2)
}

// addStandardResourceTable creates a standard resource changes table without grouping
func (f *Formatter) addStandardResourceTable(summary *PlanSummary, builder *output.Builder) {
	// Apply priority sorting before preparing table data (Requirements 2.1, 2.2, 2.3)
	sortedResources := f.sortResourcesByPriority(summary.ResourceChanges)
	tableData := f.prepareResourceTableData(sortedResources)
	// Requirement 1.1: Only create table if data exists after filtering no-ops
	if len(tableData) > 0 {
		schema := f.getResourceTableSchema()
		resourceTable, err := output.NewTableContent("Resource Changes", tableData,
			output.WithSchema(schema...))
		if err == nil {
			builder.AddContent(resourceTable)
		} else {
			fmt.Printf("Warning: Failed to create resource changes table: %v\n", err)
		}
	}
	// If tableData is empty, table is suppressed (requirement 1.1)
}

// handleResourceDisplay handles the different resource display scenarios based on showDetails and config
func (f *Formatter) handleResourceDisplay(summary *PlanSummary, showDetails bool, outputConfig *config.OutputConfiguration, builder *output.Builder) error {
	type displayMode int
	const (
		showAllResources displayMode = iota
		showNoChangesMessage
		showSensitiveOnly
		showNothing
	)

	// Determine which display mode to use
	mode := func() displayMode {
		switch {
		case showDetails && len(summary.ResourceChanges) > 0:
			return showAllResources
		case showDetails && len(summary.ResourceChanges) == 0:
			return showNoChangesMessage
		case f.config.Plan.AlwaysShowSensitive:
			return showSensitiveOnly
		default:
			return showNothing
		}
	}()

	// Handle the selected display mode
	switch mode {
	case showAllResources:
		f.addResourceChangesTable(summary, builder)
	case showNoChangesMessage:
		builder.Text("All resources are unchanged.")
	case showSensitiveOnly:
		return f.handleSensitiveResourceDisplay(summary, outputConfig, builder)
	case showNothing:
		// Show only Plan Information and Summary Statistics tables
	}

	return nil
}

// handleSensitiveResourceDisplay handles the display of sensitive resources when details are disabled
func (f *Formatter) handleSensitiveResourceDisplay(summary *PlanSummary, outputConfig *config.OutputConfiguration, builder *output.Builder) error {
	sensitiveChanges := f.filterSensitiveChanges(summary.ResourceChanges)
	if len(sensitiveChanges) > 0 {
		sensitiveData, err := f.createSensitiveResourceChangesDataV2(summary)
		if err != nil {
			return fmt.Errorf("failed to create sensitive resource changes data: %w", err)
		}
		sensitiveTable, err := output.NewTableContent("Sensitive Resource Changes", sensitiveData,
			output.WithKeys("Action", "Resource", "Type", "ID", "Replacement", "Module", "Danger"))
		if err == nil {
			builder.AddContent(sensitiveTable)
		} else {
			// Log warning but continue operation - conservative error handling
			fmt.Printf("Warning: Failed to create sensitive resource changes table: %v\n", err)
		}
	} else if outputConfig.OutputFile == "" {
		builder.Text("No sensitive resource changes detected.")
	}
	return nil
}

// createOutputChangesData creates the output changes data (requirement 2.2)
func (f *Formatter) createOutputChangesData(summary *PlanSummary) ([]map[string]any, error) {
	if summary == nil || len(summary.OutputChanges) == 0 {
		return nil, nil // Return nil for empty outputs to suppress section (requirement 2.8)
	}

	var data []map[string]any

	for _, change := range summary.OutputChanges {
		// Format current (before) value
		currentValue := formatOutputValue(change.Before, change.Sensitive, false) // Before values are typically not unknown

		// Format planned (after) value
		plannedValue := formatOutputValue(change.After, change.Sensitive, change.IsUnknown)

		// Format sensitive indicator (requirement 2.4)
		sensitiveIndicator := ""
		if change.Sensitive {
			sensitiveIndicator = "⚠️"
		}

		data = append(data, map[string]any{
			"Name":      change.Name,
			"Action":    change.Action,
			"Current":   currentValue,
			"Planned":   plannedValue,
			"Sensitive": sensitiveIndicator,
		})
	}

	return data, nil
}

// formatOutputValue formats an output value for display (requirements 2.3, 2.4)
func formatOutputValue(value any, sensitive bool, isUnknown bool) string {
	if sensitive {
		return "(sensitive value)" // requirement 2.4
	}
	if isUnknown {
		return knownAfterApply // requirement 2.3
	}
	if value == nil {
		return "-"
	}
	// Format as JSON for consistent display
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// handleOutputDisplay handles the display of output changes section (requirement 2.1)
func (f *Formatter) handleOutputDisplay(summary *PlanSummary, builder *output.Builder) error {
	// Create outputs data
	outputsData, err := f.createOutputChangesData(summary)
	if err != nil {
		return fmt.Errorf("failed to create output changes data: %w", err)
	}

	// Only add outputs section if there are output changes (requirement 2.8)
	if len(outputsData) > 0 {
		outputsTable, err := output.NewTableContent("Output Changes", outputsData,
			output.WithKeys("Name", "Action", "Current", "Planned", "Sensitive"))
		if err == nil {
			builder.AddContent(outputsTable)
		} else {
			// Log warning but continue operation - conservative error handling
			fmt.Printf("Warning: Failed to create output changes table: %v\n", err)
		}
	}
	// If outputsData is empty, section is suppressed (requirement 2.8)

	return nil
}

// filterNoOps filters out resources where ChangeType == ChangeTypeNoOp when ShowNoOps is false
// This implements Task 4.1 from the Output Refinements feature (Requirement 3.2)
func (f *Formatter) filterNoOps(resources []ResourceChange) []ResourceChange {
	if f.config.Plan.ShowNoOps {
		// Return original slice when ShowNoOps is true
		return resources
	}

	// Filter out resources where ChangeType == ChangeTypeNoOp
	filtered := make([]ResourceChange, 0, len(resources))
	for _, r := range resources {
		if r.ChangeType != ChangeTypeNoOp {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// filterNoOpOutputs filters out outputs where IsNoOp == true when ShowNoOps is false
// This implements Task 4.2 from the Output Refinements feature (Requirement 4.2, 4.3, 4.4)
func (f *Formatter) filterNoOpOutputs(outputs []OutputChange) []OutputChange {
	if f.config.Plan.ShowNoOps {
		// Include no-op outputs when --show-no-ops flag is enabled (Requirement 4.4)
		return outputs
	}

	// Filter out outputs where IsNoOp == true when ShowNoOps is false
	filtered := make([]OutputChange, 0, len(outputs))
	for _, o := range outputs {
		if !o.IsNoOp {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

// sortResourcesByPriority sorts resources by danger, action priority, and alphabetically
// This implements Task 5.1 from the Output Refinements feature (Requirements 2.1, 2.2, 2.3, 2.4)
func (f *Formatter) sortResourcesByPriority(resources []ResourceChange) []ResourceChange {
	// Make a copy to avoid modifying the original slice
	sorted := make([]ResourceChange, len(resources))
	copy(sorted, resources)

	sort.Slice(sorted, func(i, j int) bool {
		ri, rj := sorted[i], sorted[j]

		// First: Sort by danger/sensitivity (dangerous resources first - Requirement 2.1)
		if ri.IsDangerous != rj.IsDangerous {
			return ri.IsDangerous // Dangerous first
		}

		// Second: Sort by action type: delete > replace > update > create (Requirement 2.2)
		actionPriority := map[ChangeType]int{
			ChangeTypeDelete:  0, // Highest priority
			ChangeTypeReplace: 1,
			ChangeTypeUpdate:  2,
			ChangeTypeCreate:  3,
			ChangeTypeNoOp:    4, // Lowest priority
		}

		pi, pj := actionPriority[ri.ChangeType], actionPriority[rj.ChangeType]
		if pi != pj {
			return pi < pj
		}

		// Third: Alphabetical by resource address (Requirement 2.3)
		return ri.Address < rj.Address
	})

	return sorted
}

// sortResourceTableData sorts table data by danger, action priority, then alphabetically
// This implements the data-level sorting described in the design document
func sortResourceTableData(tableData []map[string]any) {
	sort.SliceStable(tableData, func(i, j int) bool {
		a, b := tableData[i], tableData[j]

		// 1. Compare danger status (using IsDangerous flag)
		dangerA, _ := a["IsDangerous"].(bool)
		dangerB, _ := b["IsDangerous"].(bool)
		if dangerA != dangerB {
			return dangerA // dangerous items first
		}

		// 2. Compare raw action priority (before decoration)
		actionA, _ := a["ActionType"].(string)
		actionB, _ := b["ActionType"].(string)
		priorityA := getActionPriority(actionA)
		priorityB := getActionPriority(actionB)
		if priorityA != priorityB {
			return priorityA < priorityB
		}

		// 3. Alphabetical by resource address
		resourceA, _ := a["Resource"].(string)
		resourceB, _ := b["Resource"].(string)
		return resourceA < resourceB
	})
}

// getActionPriority returns priority for sorting (lower = higher priority)
// This implements the action priority logic described in the design document
func getActionPriority(action string) int {
	switch action {
	case "Remove":
		return 0
	case "Replace":
		return 1
	case "Modify":
		return 2
	case "Add":
		return 3
	default:
		return 4
	}
}

// applyDecorations adds emoji and styling to sorted data
// This implements the decoration logic described in the design document
func applyDecorations(tableData []map[string]any) {
	for _, row := range tableData {
		actionType, _ := row["ActionType"].(string)
		isDangerous, _ := row["IsDangerous"].(bool)

		// Apply emoji decoration based on danger flag
		if isDangerous {
			row["Action"] = "⚠️ " + actionType
		} else {
			row["Action"] = actionType
		}

		// Remove internal fields used for sorting
		delete(row, "ActionType")
		delete(row, "IsDangerous")
	}
}
