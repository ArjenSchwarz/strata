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
	notApplicable       = "N/A"
	formatTable         = "table"
	noPropertiesChanged = "No properties changed"
	truncatedIndicator  = " [truncated]"
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

	sort.Slice(sortedIndices, func(i, j int) bool {
		dangerI := isDangerousRow(dataRows[sortedIndices[i]])
		dangerJ := isDangerousRow(dataRows[sortedIndices[j]])

		// If one is dangerous and the other isn't, dangerous comes first
		if dangerI != dangerJ {
			return dangerI
		}

		// If both have same danger status, sort by action priority
		actionI := extractActionFromTableRow(dataRows[sortedIndices[i]])
		actionJ := extractActionFromTableRow(dataRows[sortedIndices[j]])

		priorityI := getActionSortPriority(actionI)
		priorityJ := getActionSortPriority(actionJ)

		return priorityI < priorityJ
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

// extractActionFromTableRow extracts the action from a table row
func extractActionFromTableRow(row string) string {
	// Use regex to find action words at the beginning of table cells
	re := regexp.MustCompile(`^\s*\|\s*(Add|Remove|Replace|Modify)\s*\|`)
	matches := re.FindStringSubmatch(row)
	if len(matches) > 1 {
		return matches[1]
	}
	// Also check for actions with emoji prefix (like "⚠️ Remove")
	re2 := regexp.MustCompile(`^\s*\|\s*[^|]*\s*(Add|Remove|Replace|Modify)\s*\|`)
	matches2 := re2.FindStringSubmatch(row)
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
	lowercaseFormat := strings.ToLower(outputFormat)
	for _, format := range supportedFormats {
		if lowercaseFormat == format {
			return nil
		}
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

	// Build the document using v2 builder pattern
	builder := output.New()

	// Re-enable all tables using the proven NewTableContent pattern
	// This fixes the multi-table rendering issue by using consistent table creation methods

	// Plan Information table - RE-ENABLED using NewTableContent pattern
	planData, err := f.createPlanInfoDataV2(summary)
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
	statsData, err := f.createStatisticsSummaryDataV2(summary)
	if err == nil && len(statsData) > 0 {
		statsTable, err := output.NewTableContent("Summary Statistics", statsData,
			output.WithKeys("TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED"))
		if err == nil {
			builder = builder.AddContent(statsTable)
		} else {
			// Log warning but continue operation - conservative error handling
			fmt.Printf("Warning: Failed to create summary statistics table: %v\n", err)
		}
	}

	// Resource Changes table - UNIFIED TABLE CREATION following go-output example pattern
	if err := f.handleResourceDisplay(summary, showDetails, outputConfig, builder); err != nil {
		return err
	}
	// If no conditions above are met, we show only Plan Information and Summary Statistics tables

	// Output Changes table - placed after resource changes section (requirement 2.1)
	if err := f.handleOutputDisplay(summary, builder); err != nil {
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
			afterValue := f.formatValueWithContext(change.After, change.Sensitive, true, "\u2002\u2002\u2002\u2002")
			line = fmt.Sprintf("  + %s = %s", change.Name, afterValue)
		} else {
			line = fmt.Sprintf("  + %s = %s",
				change.Name, f.formatValue(change.After, change.Sensitive))
		}
	case "remove":
		if isComplexValue(change.Before) {
			beforeValue := f.formatValueWithContext(change.Before, change.Sensitive, true, "\u2002\u2002\u2002\u2002")
			line = fmt.Sprintf("  - %s = %s", change.Name, beforeValue)
		} else {
			line = fmt.Sprintf("  - %s = %s",
				change.Name, f.formatValue(change.Before, change.Sensitive))
		}
	case "update":
		if isComplexValue(change.Before) || isComplexValue(change.After) {
			beforeValue := f.formatValueWithContext(change.Before, change.Sensitive, true, "\u2002\u2002\u2002\u2002")
			afterValue := f.formatValueWithContext(change.After, change.Sensitive, true, "\u2002\u2002\u2002\u2002")
			line = fmt.Sprintf("  ~ %s = %s -> %s",
				change.Name, beforeValue, afterValue)
		} else {
			line = fmt.Sprintf("  ~ %s = %s -> %s",
				change.Name,
				f.formatValue(change.Before, change.Sensitive),
				f.formatValue(change.After, change.Sensitive))
		}
	default:
		return ""
	}

	return line + replacementIndicator
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
		if v == "(known after apply)" {
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
		nextIndent := baseIndent + "\u2002\u2002\u2002\u2002"
		// Check if the value is complex (map or slice) to handle nested structures properly
		isValueNested := false
		switch v[key].(type) {
		case map[string]any, []any:
			isValueNested = true
		}
		value := f.formatValueWithContext(v[key], false, isValueNested, nextIndent)
		// Use Unicode En spaces for consistent indentation across all formats
		lines = append(lines, fmt.Sprintf("%s\u2002\u2002\u2002\u2002%s = %s", baseIndent, key, value))
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
		nextIndent := baseIndent + "\u2002\u2002"
		// Check if the item is complex (map or slice) to handle nested structures properly
		isItemNested := false
		switch item.(type) {
		case map[string]any, []any:
			isItemNested = true
		}
		value := f.formatValueWithContext(item, false, isItemNested, nextIndent)
		// Use Unicode En spaces for consistent indentation across all formats
		lines = append(lines, fmt.Sprintf("%s\u2002\u2002[%d] = %s", baseIndent, i, value))
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
			"action":           actionDisplay,
			"resource":         change.Address,
			"type":             change.Type,
			"id":               f.getDisplayID(change),
			"replacement":      f.getReplacementDisplay(change),
			"module":           change.ModulePath,
			"danger":           f.getDangerDisplay(change),
			"risk_level":       riskLevel,
			"property_changes": propertyChangesData, // Will be formatted by collapsible formatter
			// Add unknown value fields for JSON output
			"has_unknown_values": change.HasUnknownValues,
			"unknown_properties": change.UnknownProperties,
			// Add raw property changes for JSON structured access
			"property_change_details": propChanges.Changes,
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
		tableData := f.prepareResourceTableData(summary.ResourceChanges)

		// Use NewTableContent consistently to match working example pattern
		schema := f.getResourceTableSchema()
		resourceTable, err := output.NewTableContent("Resource Changes", tableData, output.WithSchema(schema...))
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
		output.WithKeys("TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED"))
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
		output.WithKeys("TOTAL CHANGES", "ADDED", "REMOVED", "MODIFIED", "REPLACEMENTS", "HIGH RISK", "UNMODIFIED"))
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

		// Prepare table data for this provider's resources
		tableData := f.prepareResourceTableData(resources)
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
			Name: "action",
			Type: "string",
		},
		{
			Name:      "resource",
			Type:      "string",
			Formatter: output.FilePathFormatter(50),
		},
		{
			Name: "type",
			Type: "string",
		},
		{
			Name: "id",
			Type: "string",
		},
		{
			Name: "replacement",
			Type: "string",
		},
		{
			Name: "module",
			Type: "string",
		},
		{
			Name: "danger",
			Type: "string",
		},
		{
			Name:      "property_changes",
			Type:      "object",
			Formatter: f.propertyChangesFormatterTerraform(),
		},
		{
			Name: "has_unknown_values",
			Type: "boolean",
		},
		{
			Name: "unknown_properties",
			Type: "array",
		},
		{
			Name: "property_change_details",
			Type: "array",
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
	groupData := f.prepareResourceTableData(resources)
	// Requirement 1.1: Only create table if data exists after filtering no-ops
	if len(groupData) > 0 {
		schema := f.getResourceTableSchema()
		// Create table for this provider group
		providerTable, err := output.NewTableContent(fmt.Sprintf("%s Resources", strings.ToUpper(providerName)), groupData, output.WithSchema(schema...))
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
	tableData := f.prepareResourceTableData(summary.ResourceChanges)
	// Requirement 1.1: Only create table if data exists after filtering no-ops
	if len(tableData) > 0 {
		schema := f.getResourceTableSchema()
		resourceTable, err := output.NewTableContent("Resource Changes", tableData, output.WithSchema(schema...))
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
			output.WithKeys("ACTION", "RESOURCE", "TYPE", "ID", "REPLACEMENT", "MODULE", "DANGER"))
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
			"NAME":      change.Name,
			"ACTION":    change.Action,
			"CURRENT":   currentValue,
			"PLANNED":   plannedValue,
			"SENSITIVE": sensitiveIndicator,
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
		return "(known after apply)" // requirement 2.3
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
			output.WithKeys("NAME", "ACTION", "CURRENT", "PLANNED", "SENSITIVE"))
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
