package plan

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
)

const (
	riskLevelHigh   = "high"
	riskLevelMedium = "medium"

	// MaxPropertiesPerResource limits the number of properties extracted per resource to prevent runaway extraction
	MaxPropertiesPerResource = 100
	// MaxPropertyValueSize limits individual property values to 10KB
	MaxPropertyValueSize = 10240
	// MaxTotalPropertyMemory limits total memory usage for all properties to 10MB
	MaxTotalPropertyMemory = 10485760
)

// Analyzer processes Terraform plan data and generates summaries
type Analyzer struct {
	plan          *tfjson.Plan
	config        *config.Config
	providerCache sync.Map // Cache for provider extraction results
}

// NewAnalyzer creates a new plan analyzer
func NewAnalyzer(plan *tfjson.Plan, cfg *config.Config) *Analyzer {
	return &Analyzer{
		plan:   plan,
		config: cfg,
	}
}

// compareObjects performs deep object comparison for property change extraction with optional replacement path checking
func (a *Analyzer) compareObjects(path string, before, after, beforeSensitive, afterSensitive any, replacePathStrings []string, analysis *PropertyChangeAnalysis) {
	// Handle nil cases
	if before == nil && after == nil {
		return
	}

	// Helper function to check if value is complex (map or slice)
	isComplexType := func(val any) bool {
		switch val.(type) {
		case map[string]any, []any:
			return true
		default:
			return false
		}
	}

	// Helper function to check if this is a nested object property that should be treated as a single change
	shouldTreatAsNestedObject := func(before, after any, path string) bool {
		// Never treat root-level objects as nested (empty path means root level)
		if path == "" {
			return false
		}

		// Check if both values are maps (nested objects)
		beforeMap, beforeIsMap := before.(map[string]any)
		afterMap, afterIsMap := after.(map[string]any)

		if !beforeIsMap && !afterIsMap {
			return false
		}

		// Handle cases where one is nil (complete addition/removal of nested object)
		if (before == nil && afterIsMap) || (after == nil && beforeIsMap) {
			return true
		}

		// Both are maps - check if this is a nested object that should be treated as single change
		// Common nested object property names that should be grouped
		propertyName := a.extractPropertyName(path)
		isCommonNestedProperty := propertyName == "tags" ||
			propertyName == "metadata" ||
			propertyName == "labels" ||
			propertyName == "environment" ||
			strings.HasSuffix(propertyName, "_config") ||
			strings.HasSuffix(propertyName, "_settings")

		// If it's a common nested property, treat as single change
		if isCommonNestedProperty {
			return true
		}

		// For other nested objects, check if they're relatively simple (not deeply nested)
		// If all values in both maps are primitive types, treat as single nested object
		allPrimitive := true
		for _, val := range beforeMap {
			if isComplexType(val) {
				allPrimitive = false
				break
			}
		}
		if allPrimitive {
			for _, val := range afterMap {
				if isComplexType(val) {
					allPrimitive = false
					break
				}
			}
		}

		return allPrimitive
	}

	// Determine action type
	determineAction := func(before, after any) string {
		if before == nil {
			return "add"
		}
		if after == nil {
			return "remove"
		}
		return "update"
	}

	// Handle nested objects first - check if this should be treated as a single nested property change
	_, beforeIsMap := before.(map[string]any)
	_, afterIsMap := after.(map[string]any)

	if (beforeIsMap || afterIsMap) && shouldTreatAsNestedObject(before, after, path) {
		// Only create a PropertyChange if the objects are actually different
		if !equals(before, after) {
			propertyPath := a.parsePath(path)
			triggersReplacement := false
			action := determineAction(before, after)

			// Check replacement paths if provided
			if len(replacePathStrings) > 0 {
				triggersReplacement = a.pathMatchesReplacePathString(propertyPath, replacePathStrings)
			}

			analysis.Changes = append(analysis.Changes, PropertyChange{
				Name:                a.extractPropertyName(path),
				Path:                propertyPath,
				Before:              before,
				After:               after,
				Action:              action,
				TriggersReplacement: triggersReplacement,
				Sensitive:           a.isSensitive(path, beforeSensitive) || a.isSensitive(path, afterSensitive),
			})
		}
		// Don't recurse further for nested objects we're treating as single changes
		return
	}

	// Property changes - only record for leaf values, not complex objects
	action := determineAction(before, after)
	if (before == nil || after == nil || !reflect.DeepEqual(before, after)) && !isComplexType(before) && !isComplexType(after) {
		propertyPath := a.parsePath(path)
		triggersReplacement := false

		// Check replacement paths if provided
		if len(replacePathStrings) > 0 {
			triggersReplacement = a.pathMatchesReplacePathString(propertyPath, replacePathStrings)
		}

		analysis.Changes = append(analysis.Changes, PropertyChange{
			Name:                a.extractPropertyName(path),
			Path:                propertyPath,
			Before:              before,
			After:               after,
			Action:              action,
			TriggersReplacement: triggersReplacement,
			Sensitive:           a.isSensitive(path, beforeSensitive) || a.isSensitive(path, afterSensitive),
		})

		// For leaf values, don't recurse further
		if !isComplexType(before) && !isComplexType(after) {
			return
		}
	}

	// Handle nested objects that should be recursively processed
	switch beforeVal := before.(type) {
	case map[string]any:
		var afterMap map[string]any
		var ok bool
		if after != nil {
			afterMap, ok = after.(map[string]any)
			if !ok {
				return
			}
		}

		// Get all unique keys from both maps
		allKeys := make(map[string]bool)
		for key := range beforeVal {
			allKeys[key] = true
		}
		for key := range afterMap {
			allKeys[key] = true
		}

		// Recursively compare each key
		for key := range allKeys {
			var newPath string
			if path != "" {
				newPath = fmt.Sprintf("%s.%s", path, key)
			} else {
				newPath = key
			}

			beforeChild := beforeVal[key]
			var afterChild any
			if afterMap != nil {
				afterChild = afterMap[key]
			}

			var beforeSensChild, afterSensChild any
			if beforeSensitive != nil {
				if beforeSensMap, ok := beforeSensitive.(map[string]any); ok {
					beforeSensChild = beforeSensMap[key]
				}
			}
			if afterSensitive != nil {
				if afterSensMap, ok := afterSensitive.(map[string]any); ok {
					afterSensChild = afterSensMap[key]
				}
			}

			a.compareObjects(newPath, beforeChild, afterChild, beforeSensChild, afterSensChild, replacePathStrings, analysis)
		}

	case []any:
		afterSlice, ok := after.([]any)
		if !ok || afterSlice == nil {
			return
		}

		// For different sized slices, treat as a single change if both have content
		if len(beforeVal) != len(afterSlice) && len(beforeVal) > 0 && len(afterSlice) > 0 {
			propertyPath := a.parsePath(path)
			triggersReplacement := false
			if len(replacePathStrings) > 0 {
				triggersReplacement = a.pathMatchesReplacePathString(propertyPath, replacePathStrings)
			}

			analysis.Changes = append(analysis.Changes, PropertyChange{
				Name:                a.extractPropertyName(path),
				Path:                propertyPath,
				Before:              before,
				After:               after,
				Action:              "update",
				TriggersReplacement: triggersReplacement,
			})
		} else {
			// Compare each element for same-sized arrays
			maxLen := len(beforeVal)
			if len(afterSlice) > maxLen {
				maxLen = len(afterSlice)
			}

			for i := 0; i < maxLen; i++ {
				newPath := fmt.Sprintf("%s[%d]", path, i)

				var beforeItem, afterItem any
				if i < len(beforeVal) {
					beforeItem = beforeVal[i]
				}
				if i < len(afterSlice) {
					afterItem = afterSlice[i]
				}

				a.compareObjects(newPath, beforeItem, afterItem,
					a.extractSensitiveIndex(beforeSensitive, i),
					a.extractSensitiveIndex(afterSensitive, i), replacePathStrings, analysis)
			}
		}

	case nil:
		// Handle nil to map/slice transitions
		switch afterVal := after.(type) {
		case map[string]any:
			// nil to map: treat all properties as additions
			for key := range afterVal {
				var newPath string
				if path != "" {
					newPath = fmt.Sprintf("%s.%s", path, key)
				} else {
					newPath = key
				}

				var afterSensChild any
				if afterSensitive != nil {
					if afterSensMap, ok := afterSensitive.(map[string]any); ok {
						afterSensChild = afterSensMap[key]
					}
				}

				a.compareObjects(newPath, nil, afterVal[key], nil, afterSensChild, replacePathStrings, analysis)
			}
		case []any:
			// nil to slice: treat all elements as additions
			for i, item := range afterVal {
				newPath := fmt.Sprintf("%s[%d]", path, i)
				a.compareObjects(newPath, nil, item, nil, a.extractSensitiveIndex(afterSensitive, i), replacePathStrings, analysis)
			}
		}
	}
}

// enforcePropertyLimits enforces performance limits on property analysis to prevent excessive memory usage
func (a *Analyzer) enforcePropertyLimits(analysis *PropertyChangeAnalysis) {
	// Limit the number of properties per resource
	if len(analysis.Changes) > MaxPropertiesPerResource {
		analysis.Changes = analysis.Changes[:MaxPropertiesPerResource]
		analysis.Truncated = true
	}

	// Calculate total size and enforce memory limits
	totalSize := 0
	for i, change := range analysis.Changes {
		size := a.estimateValueSize(change.Before) + a.estimateValueSize(change.After)
		if size > MaxPropertyValueSize {
			size = MaxPropertyValueSize // Cap individual property size
		}
		analysis.Changes[i].Size = size

		if totalSize+size > MaxTotalPropertyMemory {
			// Truncate at this point to stay within memory limits
			analysis.Changes = analysis.Changes[:i]
			analysis.Truncated = true
			break
		}
		totalSize += size
	}

	analysis.TotalSize = totalSize
	analysis.Count = len(analysis.Changes)
}

// extractPropertyName extracts the final property name from a path
func (a *Analyzer) extractPropertyName(path string) string {
	if path == "" {
		return ""
	}

	// Handle dot notation first to get the last component
	parts := strings.Split(path, ".")
	lastPart := parts[len(parts)-1]

	// Handle array indices in the last part
	if strings.Contains(lastPart, "[") {
		// For paths like "tags[0]", return the part before the array index
		beforeBracket := strings.Split(lastPart, "[")[0]
		if beforeBracket != "" {
			return beforeBracket
		}
	}

	return lastPart
}

// parsePath converts a dot-notation path to a slice of path components
func (a *Analyzer) parsePath(path string) []string {
	if path == "" {
		return []string{}
	}

	// Handle array indices by converting them to path components
	// e.g., "tags[0].name" becomes ["tags", "0", "name"]
	// e.g., "matrix[1][2]" becomes ["matrix", "1", "2"]
	result := []string{}
	parts := strings.Split(path, ".")

	for _, part := range parts {
		if strings.Contains(part, "[") {
			// Handle multiple array indices in one part like "matrix[1][2]"
			remaining := part

			// Extract the initial property name before any brackets
			firstBracket := strings.Index(remaining, "[")
			if firstBracket > 0 {
				propertyName := remaining[:firstBracket]
				result = append(result, propertyName)
				remaining = remaining[firstBracket:]
			}

			// Extract all array indices
			for strings.Contains(remaining, "[") {
				start := strings.Index(remaining, "[")
				end := strings.Index(remaining, "]")
				if start != -1 && end != -1 && end > start {
					index := remaining[start+1 : end]
					if index != "" {
						result = append(result, index)
					}
					remaining = remaining[end+1:]
				} else {
					break
				}
			}
		} else {
			result = append(result, part)
		}
	}

	return result
}

// isSensitive checks if a property at the given path is marked as sensitive
func (a *Analyzer) isSensitive(path string, sensitiveValues any) bool {
	if sensitiveValues == nil {
		return false
	}

	// Navigate the sensitive values structure following the path
	current := sensitiveValues
	pathParts := a.parsePath(path)

	for _, part := range pathParts {
		switch curr := current.(type) {
		case map[string]any:
			if val, exists := curr[part]; exists {
				current = val
			} else {
				return false
			}
		case []any:
			// Convert part to index
			if index, err := strconv.Atoi(part); err == nil && index >= 0 && index < len(curr) {
				current = curr[index]
			} else {
				return false
			}
		case bool:
			// If we encounter a boolean, it represents sensitivity for this level
			return curr
		default:
			return false
		}
	}

	// Final check - if we've navigated to a boolean, return it
	if sensitive, ok := current.(bool); ok {
		return sensitive
	}

	return false
}

// extractSensitiveChild extracts the sensitive values for a child property
func (a *Analyzer) extractSensitiveChild(sensitiveValues any, key string) any {
	if sensitiveValues == nil {
		return nil
	}

	if sensitiveMap, ok := sensitiveValues.(map[string]any); ok {
		return sensitiveMap[key]
	}

	return nil
}

// extractSensitiveIndex extracts the sensitive values for an array element
func (a *Analyzer) extractSensitiveIndex(sensitiveValues any, index int) any {
	if sensitiveValues == nil {
		return nil
	}

	if sensitiveSlice, ok := sensitiveValues.([]any); ok {
		if index >= 0 && index < len(sensitiveSlice) {
			return sensitiveSlice[index]
		}
	}

	return nil
}

// GenerateSummary creates a comprehensive summary of the plan
func (a *Analyzer) GenerateSummary(planFile string) *PlanSummary {
	parser := NewParser(planFile)

	// Load the plan if not already loaded
	if a.plan == nil {
		plan, err := parser.LoadPlan()
		if err != nil {
			return nil
		}
		a.plan = plan
	}

	summary := &PlanSummary{
		FormatVersion:    a.plan.FormatVersion,
		TerraformVersion: a.plan.TerraformVersion,
		PlanFile:         planFile,
		Workspace:        parser.extractWorkspaceInfo(a.plan),
		Backend:          parser.extractBackendInfo(a.plan),
		ResourceChanges:  a.analyzeResourceChanges(),
		OutputChanges:    a.analyzeOutputChanges(),
	}

	// Get file creation time
	if createdAt, err := parser.getPlanFileInfo(planFile); err == nil {
		summary.CreatedAt = createdAt
	}

	summary.Statistics = a.calculateStatistics(summary.ResourceChanges)
	return summary
}

// analyzeResourceChanges processes all resource changes in the plan
func (a *Analyzer) analyzeResourceChanges() []ResourceChange {
	if a.plan.ResourceChanges == nil {
		return []ResourceChange{}
	}

	changes := make([]ResourceChange, 0, len(a.plan.ResourceChanges))

	for _, rc := range a.plan.ResourceChanges {
		changeType := FromTerraformAction(rc.Change.Actions)
		replacementType := a.analyzeReplacementNecessity(rc)

		// Analyze property changes
		propertyChanges := a.analyzePropertyChanges(rc)

		change := ResourceChange{
			Address:          rc.Address,
			Type:             rc.Type,
			Name:             rc.Name,
			ChangeType:       changeType,
			IsDestructive:    changeType.IsDestructive(),
			ReplacementType:  replacementType,
			PhysicalID:       a.extractPhysicalID(rc),
			PlannedID:        a.extractPlannedID(rc),
			ModulePath:       a.extractModulePath(rc.Address),
			ChangeAttributes: a.getChangingAttributes(rc),
			Before:           rc.Change.Before,
			After:            rc.Change.After,
			// Check for sensitive resources and properties
			IsDangerous:      false, // Will be updated below
			DangerReason:     "",
			DangerProperties: []string{},
			// Enhanced summary visualization fields
			Provider:         a.extractProvider(rc.Type),
			ReplacementHints: a.extractReplacementHints(rc),
			TopChanges:       a.getTopChangedProperties(rc, 3),
			PropertyChanges:  propertyChanges,
		}

		// Enhanced danger reason logic
		change.IsDangerous, change.DangerReason = a.evaluateResourceDanger(rc, changeType)

		changes = append(changes, change)
	}

	return changes
}

// analyzeReplacementNecessity determines the replacement necessity for a resource change
func (a *Analyzer) analyzeReplacementNecessity(change *tfjson.ResourceChange) ReplacementType {
	// If it's not a destructive action, it's never a replacement
	changeType := FromTerraformAction(change.Change.Actions)
	if !changeType.IsDestructive() {
		return ReplacementNever
	}

	// Check if this is a replacement (delete + create)
	if changeType == ChangeTypeReplace {
		return ReplacementAlways
	}

	// Delete operations are not replacements
	return ReplacementNever
}

// analyzeOutputChanges processes all output changes in the plan
func (a *Analyzer) analyzeOutputChanges() []OutputChange {
	if a.plan.OutputChanges == nil {
		return []OutputChange{}
	}

	changes := make([]OutputChange, 0, len(a.plan.OutputChanges))

	for name, oc := range a.plan.OutputChanges {
		changeType := FromTerraformAction(oc.Actions)

		// Detect if output is sensitive by checking the sensitive flags
		// For outputs, if either BeforeSensitive or AfterSensitive is true, the output is sensitive
		isSensitive := a.isOutputSensitive(oc)

		change := OutputChange{
			Name:       name,
			ChangeType: changeType,
			Sensitive:  isSensitive,
			Before:     oc.Before,
			After:      oc.After,
		}

		// Mask sensitive values
		if isSensitive {
			if change.Before != nil {
				change.Before = nil // Don't expose sensitive before values
			}
			if change.After != nil {
				change.After = nil // Don't expose sensitive after values
			}
		}

		changes = append(changes, change)
	}

	return changes
}

// isOutputSensitive checks if an output change contains sensitive values
func (a *Analyzer) isOutputSensitive(oc *tfjson.Change) bool {
	// Check if BeforeSensitive or AfterSensitive indicate sensitive values
	// For outputs, these will be boolean values (true/false) if the output is sensitive
	if oc.BeforeSensitive != nil {
		if sensitive, ok := oc.BeforeSensitive.(bool); ok && sensitive {
			return true
		}
	}

	if oc.AfterSensitive != nil {
		if sensitive, ok := oc.AfterSensitive.(bool); ok && sensitive {
			return true
		}
	}

	return false
}

// calculateStatistics generates statistics from the resource changes
func (a *Analyzer) calculateStatistics(changes []ResourceChange) ChangeStatistics {
	stats := ChangeStatistics{}

	for _, change := range changes {
		// Count by change type
		switch change.ChangeType {
		case ChangeTypeCreate:
			stats.ToAdd++
		case ChangeTypeUpdate:
			stats.ToChange++
		case ChangeTypeDelete:
			stats.ToDestroy++
		case ChangeTypeReplace:
			stats.Replacements++
		case ChangeTypeNoOp:
			stats.Unmodified++
		}

		// Count high-risk changes (any resource with the dangerous flag set)
		if change.IsDangerous {
			stats.HighRisk++
		}
	}

	stats.Total = stats.ToAdd + stats.ToChange + stats.ToDestroy + stats.Replacements
	return stats
}

// GetDestructiveChanges returns only the changes that are considered destructive
func (a *Analyzer) GetDestructiveChanges(changes []ResourceChange) []ResourceChange {
	destructive := make([]ResourceChange, 0)

	for _, change := range changes {
		if change.IsDestructive {
			destructive = append(destructive, change)
		}
	}

	return destructive
}

// GetChangesByType returns changes filtered by type
func (a *Analyzer) GetChangesByType(changes []ResourceChange, changeType ChangeType) []ResourceChange {
	filtered := make([]ResourceChange, 0)

	for _, change := range changes {
		if change.ChangeType == changeType {
			filtered = append(filtered, change)
		}
	}

	return filtered
}

// GetChangesByResourceType returns changes filtered by Terraform resource type
func (a *Analyzer) GetChangesByResourceType(changes []ResourceChange, resourceType string) []ResourceChange {
	filtered := make([]ResourceChange, 0)

	for _, change := range changes {
		if strings.HasPrefix(change.Type, resourceType) {
			filtered = append(filtered, change)
		}
	}

	return filtered
}

// HasDestructiveChanges returns true if there are any destructive changes
func (a *Analyzer) HasDestructiveChanges(changes []ResourceChange) bool {
	for _, change := range changes {
		if change.IsDestructive {
			return true
		}
	}
	return false
}

// GetDestructiveChangeCount returns the count of destructive changes
func (a *Analyzer) GetDestructiveChangeCount(changes []ResourceChange) int {
	count := 0
	for _, change := range changes {
		if change.IsDestructive {
			count++
		}
	}
	return count
}

// extractPhysicalID extracts the current physical resource ID from a resource change
func (a *Analyzer) extractPhysicalID(change *tfjson.ResourceChange) string {
	// For new resources, there's no current physical ID
	if change.Change.Before == nil {
		return "-"
	}

	// Try to extract ID from the before state
	if beforeMap, ok := change.Change.Before.(map[string]any); ok {
		if id, exists := beforeMap["id"]; exists && id != nil {
			if idStr, ok := id.(string); ok && idStr != "" {
				return idStr
			}
		}
	}

	return "-"
}

// extractPlannedID extracts the planned physical resource ID from a resource change
func (a *Analyzer) extractPlannedID(change *tfjson.ResourceChange) string {
	// For deleted resources, there's no planned ID
	if change.Change.After == nil {
		return "N/A"
	}

	// Try to extract ID from the after state
	if afterMap, ok := change.Change.After.(map[string]any); ok {
		if id, exists := afterMap["id"]; exists && id != nil {
			if idStr, ok := id.(string); ok && idStr != "" {
				return idStr
			}
		}
	}

	return "-"
}

// extractModulePath extracts the module hierarchy path from a resource address
func (a *Analyzer) extractModulePath(address string) string {
	// Check if the address contains module information
	if !strings.Contains(address, "module.") {
		return "-"
	}

	// Extract module path from address
	// Example: module.app.module.storage.aws_s3_bucket.data -> app/storage
	parts := strings.Split(address, ".")
	var moduleParts []string

	for i, part := range parts {
		if part == "module" && i+1 < len(parts) {
			moduleName := parts[i+1]
			// Strip iterator brackets if present (e.g., "s3_module[0]" -> "s3_module")
			if idx := strings.Index(moduleName, "["); idx != -1 {
				moduleName = moduleName[:idx]
			}
			moduleParts = append(moduleParts, moduleName)
		}
	}

	if len(moduleParts) == 0 {
		return "-"
	}

	return strings.Join(moduleParts, "/")
}

// getChangingAttributes identifies specific attributes that are changing in a resource
func (a *Analyzer) getChangingAttributes(change *tfjson.ResourceChange) []string {
	var attributes []string

	// For now, return a basic set of attributes
	// In the future, we could analyze the before/after states to identify specific changing attributes
	switch FromTerraformAction(change.Change.Actions) {
	case ChangeTypeCreate:
		attributes = append(attributes, "all")
	case ChangeTypeDelete:
		attributes = append(attributes, "all")
	case ChangeTypeReplace:
		attributes = append(attributes, "all")
	case ChangeTypeUpdate:
		// For updates, we could analyze the diff to find specific attributes
		// For now, just indicate it's an update
		attributes = append(attributes, "modified")
	}

	return attributes
}

// IsSensitiveResource checks if a resource type is in the sensitive resources list
func (a *Analyzer) IsSensitiveResource(resourceType string) bool {
	if a.config == nil || len(a.config.SensitiveResources) == 0 {
		return false
	}

	for _, sr := range a.config.SensitiveResources {
		if sr.ResourceType == resourceType {
			return true
		}
	}

	return false
}

// IsSensitiveProperty checks if a property is sensitive for a given resource type
func (a *Analyzer) IsSensitiveProperty(resourceType string, propertyName string) bool {
	if a.config == nil || len(a.config.SensitiveProperties) == 0 {
		return false
	}

	for _, sp := range a.config.SensitiveProperties {
		if sp.ResourceType == resourceType && sp.Property == propertyName {
			return true
		}
	}

	return false
}

// checkSensitiveProperties checks if any properties in the change match sensitive properties
func (a *Analyzer) checkSensitiveProperties(change *tfjson.ResourceChange) []string {
	var sensitiveProps []string

	// If there's no change or no config, return empty
	if change.Change.Before == nil || change.Change.After == nil || a.config == nil {
		return sensitiveProps
	}

	// Extract before and after as maps
	beforeMap, beforeOk := change.Change.Before.(map[string]any)
	afterMap, afterOk := change.Change.After.(map[string]any)

	if !beforeOk || !afterOk {
		return sensitiveProps
	}

	// Check each property to see if it's changed and if it's sensitive
	for propName := range afterMap {
		// Skip if property doesn't exist in before (new property)
		beforeVal, exists := beforeMap[propName]
		if !exists {
			continue
		}

		afterVal := afterMap[propName]

		// If property has changed and is sensitive, add to list
		if !equals(beforeVal, afterVal) && a.IsSensitiveProperty(change.Type, propName) {
			sensitiveProps = append(sensitiveProps, propName)
		}
	}

	return sensitiveProps
}

// extractProvider extracts provider from resource type (e.g., "aws" from "aws_s3_bucket")
// Uses thread-safe caching for performance
func (a *Analyzer) extractProvider(resourceType string) string {
	// Check cache first
	if cached, ok := a.providerCache.Load(resourceType); ok {
		return cached.(string)
	}

	// Extract provider from resource type
	parts := strings.Split(resourceType, "_")
	provider := "unknown"
	if len(parts) > 0 && parts[0] != "" {
		provider = parts[0]
	}

	// Cache the result
	a.providerCache.Store(resourceType, provider)
	return provider
}

// extractReplacementHints extracts human-readable reasons for resource replacements
func (a *Analyzer) extractReplacementHints(change *tfjson.ResourceChange) []string {
	hints := make([]string, 0)

	// Only show replacement hints for replacement operations
	if change.Change == nil || change.Change.ReplacePaths == nil || len(change.Change.ReplacePaths) == 0 {
		return hints
	}

	// Convert ReplacePaths to human-readable strings
	for _, replacePath := range change.Change.ReplacePaths {
		hint := a.formatReplacePath(replacePath)
		if hint != "" {
			hints = append(hints, hint)
		}
	}

	return hints
}

// formatReplacePath converts a replacement path to a human-readable string
func (a *Analyzer) formatReplacePath(path any) string {
	switch p := path.(type) {
	case []any:
		// Handle nested paths like ["network_interface", 0, "subnet_id"]
		var parts []string
		for _, part := range p {
			switch partValue := part.(type) {
			case string:
				parts = append(parts, partValue)
			case int:
				parts = append(parts, "["+strconv.Itoa(partValue)+"]")
			case float64:
				parts = append(parts, "["+strconv.Itoa(int(partValue))+"]")
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, ".")
		}
	case string:
		// Handle simple string paths
		return p
	}

	return ""
}

// evaluateResourceDanger determines if a resource change is dangerous and provides a descriptive reason
func (a *Analyzer) evaluateResourceDanger(change *tfjson.ResourceChange, changeType ChangeType) (bool, string) {
	isDangerous := false
	reasonParts := make([]string, 0)

	// All deletion operations are considered risky by default
	if changeType == ChangeTypeDelete {
		isDangerous = true
		if a.IsSensitiveResource(change.Type) {
			reasonParts = append(reasonParts, "Sensitive resource deletion")
		} else {
			reasonParts = append(reasonParts, "Resource deletion")
		}
	}

	// Sensitive resource replacements are higher risk
	if a.IsSensitiveResource(change.Type) && changeType == ChangeTypeReplace {
		isDangerous = true
		reasonParts = append(reasonParts, a.getSensitiveResourceReason(change.Type))
	}

	// Check for sensitive property changes (only if we have the necessary data)
	if change.Change != nil {
		dangerProps := a.checkSensitiveProperties(change)
		if len(dangerProps) > 0 {
			isDangerous = true
			reasonParts = append(reasonParts, a.getSensitivePropertyReason(dangerProps))
		}
	}

	// Join all reasons with "and"
	reason := ""
	if len(reasonParts) > 0 {
		reason = strings.Join(reasonParts, " and ")
	}

	return isDangerous, reason
}

// getSensitiveResourceReason returns a descriptive reason for sensitive resource changes
func (a *Analyzer) getSensitiveResourceReason(resourceType string) string {
	// Provide specific reasons based on common resource types
	switch {
	case strings.Contains(resourceType, "rds") || strings.Contains(resourceType, "database"):
		return "Database replacement"
	case strings.Contains(resourceType, "instance") || strings.Contains(resourceType, "vm") || strings.Contains(resourceType, "virtual_machine"):
		return "Compute instance replacement"
	case strings.Contains(resourceType, "bucket") || strings.Contains(resourceType, "storage"):
		return "Storage replacement"
	case strings.Contains(resourceType, "security_group") || strings.Contains(resourceType, "firewall"):
		return "Security rule replacement"
	case strings.Contains(resourceType, "network") || strings.Contains(resourceType, "vpc"):
		return "Network infrastructure replacement"
	default:
		return "Sensitive resource replacement"
	}
}

// getSensitivePropertyReason returns a descriptive reason for sensitive property changes
func (a *Analyzer) getSensitivePropertyReason(properties []string) string {
	if len(properties) == 1 {
		// Provide specific reasons for common sensitive properties
		prop := properties[0]
		switch {
		case strings.Contains(strings.ToLower(prop), "password") || strings.Contains(strings.ToLower(prop), "secret"):
			return "Credential change"
		case strings.Contains(strings.ToLower(prop), "key") || strings.Contains(strings.ToLower(prop), "token"):
			return "Authentication key change"
		case strings.Contains(strings.ToLower(prop), "userdata") || strings.Contains(strings.ToLower(prop), "user_data"):
			return "User data modification"
		case strings.Contains(strings.ToLower(prop), "security") || strings.Contains(strings.ToLower(prop), "policy"):
			return "Security configuration change"
		default:
			return "Sensitive property change: " + prop
		}
	} else {
		return "Multiple sensitive properties changed"
	}
}

// getTopChangedProperties returns the first N properties that are changing for update operations
func (a *Analyzer) getTopChangedProperties(change *tfjson.ResourceChange, limit int) []string {
	properties := make([]string, 0)

	// Only show property changes for update operations when ShowContext is enabled
	if a.config == nil || !a.config.Plan.ShowContext || change.Change == nil ||
		FromTerraformAction(change.Change.Actions) != ChangeTypeUpdate {
		return properties
	}

	// Skip if we don't have both before and after states
	if change.Change.Before == nil || change.Change.After == nil {
		return properties
	}

	// Convert to maps for comparison
	beforeMap, beforeOk := change.Change.Before.(map[string]any)
	afterMap, afterOk := change.Change.After.(map[string]any)

	if !beforeOk || !afterOk {
		return properties
	}

	// Find changed properties
	count := 0
	for propName := range afterMap {
		if count >= limit {
			break
		}

		beforeVal, existsBefore := beforeMap[propName]
		afterVal := afterMap[propName]

		// If property exists in before and values differ, it's changed
		if existsBefore && !equals(beforeVal, afterVal) {
			properties = append(properties, propName)
			count++
		}
	}

	// Also check for properties that were removed (exist in before but not after)
	for propName := range beforeMap {
		if count >= limit {
			break
		}

		if _, existsAfter := afterMap[propName]; !existsAfter {
			properties = append(properties, propName+" (removed)")
			count++
		}
	}

	return properties
}

// AnalyzePropertyChanges extracts property changes with performance safeguards using the new compareObjects method (exported for testing)
func (a *Analyzer) AnalyzePropertyChanges(change *tfjson.ResourceChange) PropertyChangeAnalysis {
	return a.analyzePropertyChanges(change)
}

// analyzePropertyChanges extracts property changes with performance safeguards using the new compareObjects method
func (a *Analyzer) analyzePropertyChanges(change *tfjson.ResourceChange) PropertyChangeAnalysis {
	analysis := PropertyChangeAnalysis{
		Changes: []PropertyChange{},
	}

	if change.Change == nil {
		return analysis
	}

	// Extract replacement paths as strings for simpler matching
	var replacePathStrings []string
	if change.Change.ReplacePaths != nil {
		for _, replacePath := range change.Change.ReplacePaths {
			pathStr := a.convertReplacePathToString(replacePath)
			if pathStr != "" {
				replacePathStrings = append(replacePathStrings, pathStr)
			}
		}
	}

	// Use deep comparison with sensitive values and replacement paths
	a.compareObjects("", change.Change.Before, change.Change.After,
		change.Change.BeforeSensitive, change.Change.AfterSensitive, replacePathStrings, &analysis)

	// Note: Deduplication removed - improved comparison logic prevents duplicates at source

	// Apply performance limits using the new dedicated function
	a.enforcePropertyLimits(&analysis)
	return analysis
}

// convertReplacePathToString converts a replacement path from Terraform to a dot-notation string
func (a *Analyzer) convertReplacePathToString(replacePath any) string {
	switch p := replacePath.(type) {
	case []any:
		var parts []string
		for _, part := range p {
			switch partValue := part.(type) {
			case string:
				parts = append(parts, partValue)
			case int:
				parts = append(parts, strconv.Itoa(partValue))
			case float64:
				parts = append(parts, strconv.Itoa(int(partValue)))
			default:
				parts = append(parts, fmt.Sprintf("%v", partValue))
			}
		}
		return strings.Join(parts, ".")
	case string:
		return p
	default:
		return fmt.Sprintf("%v", p)
	}
}

// pathMatchesReplacePathString checks if a property path matches any of the replacement paths using simplified string comparison
func (a *Analyzer) pathMatchesReplacePathString(propertyPath []string, replacePathStrings []string) bool {
	if len(replacePathStrings) == 0 {
		return false
	}

	propertyPathStr := strings.Join(propertyPath, ".")

	for _, replacePathStr := range replacePathStrings {
		// Check if property path starts with replacement path (prefix match)
		if strings.HasPrefix(propertyPathStr, replacePathStr) {
			// Ensure it's a complete path component match, not a partial match
			if len(propertyPathStr) == len(replacePathStr) ||
				(len(propertyPathStr) > len(replacePathStr) && propertyPathStr[len(replacePathStr)] == '.') {
				return true
			}
		}
	}
	return false
}

// compareValues recursively compares two values and calls the callback for each difference
func (a *Analyzer) compareValues(before, after any, path []string, depth, maxDepth int, callback func(PropertyChange) bool) error {
	// Prevent infinite recursion
	if depth > maxDepth {
		return nil
	}

	// Handle nil cases
	if before == nil && after == nil {
		return nil
	}

	// If values are equal, no change
	if equals(before, after) {
		return nil
	}

	// Handle maps specially
	beforeMap, beforeIsMap := before.(map[string]any)
	afterMap, afterIsMap := after.(map[string]any)

	if beforeIsMap && afterIsMap {
		// Compare map keys
		allKeys := make(map[string]bool)
		for k := range beforeMap {
			allKeys[k] = true
		}
		for k := range afterMap {
			allKeys[k] = true
		}

		for key := range allKeys {
			beforeVal, beforeExists := beforeMap[key]
			afterVal, afterExists := afterMap[key]

			newPath := make([]string, len(path)+1)
			copy(newPath, path)
			newPath[len(path)] = key

			switch {
			case !beforeExists:
				// New property
				pc := PropertyChange{
					Name:      strings.Join(newPath, "."),
					Path:      newPath,
					Before:    nil,
					After:     afterVal,
					Sensitive: false, // Will be updated if needed
				}
				if !callback(pc) {
					return nil // Stop processing
				}
			case !afterExists:
				// Removed property
				pc := PropertyChange{
					Name:      strings.Join(newPath, "."),
					Path:      newPath,
					Before:    beforeVal,
					After:     nil,
					Sensitive: false,
				}
				if !callback(pc) {
					return nil
				}
			default:
				// Compare nested values
				err := a.compareValues(beforeVal, afterVal, newPath, depth+1, maxDepth, callback)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Handle slices specially
	beforeSlice, beforeIsSlice := before.([]any)
	afterSlice, afterIsSlice := after.([]any)

	if beforeIsSlice && afterIsSlice {
		maxLen := len(beforeSlice)
		if len(afterSlice) > maxLen {
			maxLen = len(afterSlice)
		}

		for i := 0; i < maxLen; i++ {
			var beforeVal, afterVal any
			indexPath := make([]string, len(path)+1)
			copy(indexPath, path)
			indexPath[len(path)] = strconv.Itoa(i)

			if i < len(beforeSlice) {
				beforeVal = beforeSlice[i]
			}
			if i < len(afterSlice) {
				afterVal = afterSlice[i]
			}

			if !equals(beforeVal, afterVal) {
				err := a.compareValues(beforeVal, afterVal, indexPath, depth+1, maxDepth, callback)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	// For primitive values or different types, record the change
	pc := PropertyChange{
		Name:      strings.Join(path, "."),
		Path:      path,
		Before:    before,
		After:     after,
		Sensitive: false,
	}

	// Check if this property is sensitive
	// For now, we'll skip sensitive property detection in this function
	// and handle it at a higher level as we need more context

	callback(pc)
	return nil
}

// estimateValueSize estimates the memory size of a value in bytes
func (a *Analyzer) estimateValueSize(value any) int {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case string:
		return len(v)
	case int, int8, int16, int32, int64:
		return 8
	case uint, uint8, uint16, uint32, uint64:
		return 8
	case float32:
		return 4
	case float64:
		return 8
	case bool:
		return 1
	case map[string]any:
		size := 0
		for k, val := range v {
			size += len(k) + a.estimateValueSize(val)
		}
		return size
	case []any:
		size := 0
		for _, val := range v {
			size += a.estimateValueSize(val)
		}
		return size
	default:
		// For unknown types, use JSON marshaling size as approximation
		// This is expensive but gives a reasonable estimate
		return len(fmt.Sprintf("%v", v))
	}
}

// assessRiskLevel provides simplified risk assessment
func (a *Analyzer) assessRiskLevel(change *tfjson.ResourceChange) string {
	// Simple risk assessment based on change type and resource sensitivity
	changeType := FromTerraformAction(change.Change.Actions)

	if changeType == ChangeTypeDelete {
		if a.IsSensitiveResource(change.Type) {
			return "critical"
		}
		return riskLevelHigh
	}

	if changeType == ChangeTypeReplace {
		if a.IsSensitiveResource(change.Type) {
			return riskLevelHigh
		}
		return riskLevelMedium
	}

	if a.IsSensitiveResource(change.Type) && changeType == ChangeTypeUpdate {
		return riskLevelMedium
	}

	return "low"
}

// AnalyzeResource performs comprehensive analysis with performance limits
func (a *Analyzer) AnalyzeResource(change *tfjson.ResourceChange) (*ResourceAnalysis, error) {
	analysis := &ResourceAnalysis{}

	// Extract property changes with limits for performance
	propAnalysis := a.analyzePropertyChanges(change)
	analysis.PropertyChanges = propAnalysis

	// Get replacement reasons (existing functionality)
	analysis.ReplacementReasons = a.extractReplacementHints(change)

	// Perform simple risk assessment
	analysis.RiskLevel = a.assessRiskLevel(change)

	return analysis, nil
}

// groupByProvider groups resource changes by provider with smart grouping logic
func (a *Analyzer) groupByProvider(changes []ResourceChange) map[string][]ResourceChange {
	groups := make(map[string][]ResourceChange)

	// Check if grouping should be applied
	if a.config == nil || !a.config.Plan.Grouping.Enabled {
		return groups
	}

	// Check threshold - only group if we have enough resources
	threshold := a.config.Plan.Grouping.Threshold
	if threshold == 0 {
		threshold = 10 // Default threshold
	}
	if len(changes) < threshold {
		return groups
	}

	// Count providers to check diversity
	providerCounts := make(map[string]int)
	for _, change := range changes {
		provider := a.extractProvider(change.Type)
		providerCounts[provider]++
	}

	// Skip grouping if all resources are from the same provider
	if len(providerCounts) <= 1 {
		return groups
	}

	// Group resources by provider
	for _, change := range changes {
		provider := a.extractProvider(change.Type)
		groups[provider] = append(groups[provider], change)
	}

	return groups
}

// equals is a helper to compare two values, handling maps and slices specially
func equals(a, b any) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle maps specially since they're not directly comparable
	// Check if both are maps
	aMap, aIsMap := a.(map[string]any)
	bMap, bIsMap := b.(map[string]any)

	if aIsMap && bIsMap {
		// If maps have different lengths, they're not equal
		if len(aMap) != len(bMap) {
			return false
		}

		// Check each key-value pair
		for k, aVal := range aMap {
			bVal, exists := bMap[k]
			if !exists {
				return false
			}

			// Recursively compare values
			if !equals(aVal, bVal) {
				return false
			}
		}
		return true
	}

	// Handle slices specially since they're not directly comparable
	aSlice, aIsSlice := a.([]any)
	bSlice, bIsSlice := b.([]any)

	if aIsSlice && bIsSlice {
		// If slices have different lengths, they're not equal
		if len(aSlice) != len(bSlice) {
			return false
		}

		// Check each element
		for i, aVal := range aSlice {
			bVal := bSlice[i]
			// Recursively compare values
			if !equals(aVal, bVal) {
				return false
			}
		}
		return true
	}

	// For non-map and non-slice types, use direct comparison
	return a == b
}
