package plan

import (
	"strings"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
)

// Analyzer processes Terraform plan data and generates summaries
type Analyzer struct {
	plan   *tfjson.Plan
	config *config.Config
}

// NewAnalyzer creates a new plan analyzer
func NewAnalyzer(plan *tfjson.Plan, cfg *config.Config) *Analyzer {
	return &Analyzer{
		plan:   plan,
		config: cfg,
	}
}

// GenerateSummary creates a comprehensive summary of the plan
func (a *Analyzer) GenerateSummary(planFile string) *PlanSummary {
	parser := NewParser(planFile)

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
		}

		// Check if this is a sensitive resource
		if a.IsSensitiveResource(rc.Type) && changeType == ChangeTypeReplace {
			change.IsDangerous = true
			change.DangerReason = "Sensitive resource replacement"
		}

		// Check for sensitive properties
		dangerProps := a.checkSensitiveProperties(rc)
		if len(dangerProps) > 0 {
			change.IsDangerous = true
			change.DangerProperties = dangerProps
			if change.DangerReason == "" {
				change.DangerReason = "Sensitive property change"
			} else {
				change.DangerReason += " and sensitive property change"
			}
		}

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

		change := OutputChange{
			Name:       name,
			ChangeType: changeType,
			Sensitive:  false, // Default to false, will be updated when we find the correct field
			Before:     oc.Before,
			After:      oc.After,
		}

		changes = append(changes, change)
	}

	return changes
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
			moduleParts = append(moduleParts, parts[i+1])
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
