package plan

import (
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

// Analyzer processes Terraform plan data and generates summaries
type Analyzer struct {
	plan *tfjson.Plan
}

// NewAnalyzer creates a new plan analyzer
func NewAnalyzer(plan *tfjson.Plan) *Analyzer {
	return &Analyzer{
		plan: plan,
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
		IsDryRun:         false, // TODO: Detect dry run mode
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
		// For now, treat all replacements as definite (Always)
		// In the future, we could analyze the change details to determine
		// if it's conditional based on computed values or unknown attributes
		return ReplacementAlways
	}

	// Delete operations are not replacements
	return ReplacementNever
}

// isConditionalReplacement checks if a resource change is a conditional replacement
func (a *Analyzer) isConditionalReplacement(change *tfjson.ResourceChange) bool {
	return a.analyzeReplacementNecessity(change) == ReplacementConditional
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
		switch change.ChangeType {
		case ChangeTypeCreate:
			stats.ToAdd++
		case ChangeTypeUpdate:
			stats.ToChange++
		case ChangeTypeDelete:
			stats.ToDestroy++
		case ChangeTypeReplace:
			// Count replacements and conditionals separately
			if change.ReplacementType == ReplacementConditional {
				stats.Conditionals++
			} else {
				stats.Replacements++
			}
		}
	}

	stats.Total = stats.ToAdd + stats.ToChange + stats.ToDestroy + stats.Replacements + stats.Conditionals
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
	if beforeMap, ok := change.Change.Before.(map[string]interface{}); ok {
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
	if afterMap, ok := change.Change.After.(map[string]interface{}); ok {
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
