package plan

import (
	"time"

	tfjson "github.com/hashicorp/terraform-json"
)

// ChangeType represents the type of change being made to a resource
type ChangeType string

// ChangeType constants represent different types of Terraform resource changes
const (
	ChangeTypeCreate  ChangeType = "create"  // Resource is being created
	ChangeTypeUpdate  ChangeType = "update"  // Resource is being updated
	ChangeTypeDelete  ChangeType = "delete"  // Resource is being deleted
	ChangeTypeReplace ChangeType = "replace" // Resource is being replaced
	ChangeTypeNoOp    ChangeType = "no-op"   // No operation on resource
)

// ReplacementType represents whether a resource will be replaced
type ReplacementType string

// ReplacementType constants represent different replacement scenarios for Terraform resources
const (
	ReplacementNever  ReplacementType = "Never"  // Resource will not be replaced
	ReplacementAlways ReplacementType = "Always" // Resource will be replaced
)

// ResourceChange represents a change to a Terraform resource
type ResourceChange struct {
	Address          string          `json:"address"`
	Type             string          `json:"type"`
	Name             string          `json:"name"`
	ChangeType       ChangeType      `json:"change_type"`
	IsDestructive    bool            `json:"is_destructive"`
	ReplacementType  ReplacementType `json:"replacement_type"`
	PhysicalID       string          `json:"physical_id"`       // current physical resource ID
	PlannedID        string          `json:"planned_id"`        // planned physical resource ID
	ModulePath       string          `json:"module_path"`       // module hierarchy path
	ChangeAttributes []string        `json:"change_attributes"` // specific attributes changing
	Before           any             `json:"before,omitempty"`
	After            any             `json:"after,omitempty"`
	// New fields for danger highlights
	IsDangerous      bool     `json:"is_dangerous"`      // Whether this change is flagged as dangerous
	DangerReason     string   `json:"danger_reason"`     // Reason why this change is dangerous
	DangerProperties []string `json:"danger_properties"` // List of dangerous property changes
}

// PlanSummary contains the summarised information from a Terraform plan
type PlanSummary struct {
	FormatVersion    string           `json:"format_version"`
	TerraformVersion string           `json:"terraform_version"`
	PlanFile         string           `json:"plan_file"`
	Workspace        string           `json:"workspace"`
	Backend          BackendInfo      `json:"backend"`
	CreatedAt        time.Time        `json:"created_at"`
	ResourceChanges  []ResourceChange `json:"resource_changes"`
	OutputChanges    []OutputChange   `json:"output_changes"`
	Statistics       ChangeStatistics `json:"statistics"`
}

// OutputChange represents a change to a Terraform output
type OutputChange struct {
	Name       string     `json:"name"`
	ChangeType ChangeType `json:"change_type"`
	Sensitive  bool       `json:"sensitive"`
	Before     any        `json:"before,omitempty"`
	After      any        `json:"after,omitempty"`
}

// BackendInfo contains information about the Terraform backend
type BackendInfo struct {
	Type     string         `json:"type"`     // e.g., "s3", "local", "remote"
	Location string         `json:"location"` // bucket name, file path, etc.
	Config   map[string]any `json:"config"`   // additional backend config
}

// ChangeStatistics provides counts of different types of changes for the enhanced statistics summary table
type ChangeStatistics struct {
	ToAdd        int `json:"to_add"`       // ADDED: Resources to be created (new resources)
	ToChange     int `json:"to_change"`    // MODIFIED: Resources to be updated (existing resources with changes)
	ToDestroy    int `json:"to_destroy"`   // REMOVED: Resources to be destroyed (deleted resources)
	Replacements int `json:"replacements"` // REPLACEMENTS: Resources to be replaced (definite replacements)
	HighRisk     int `json:"high_risk"`    // HIGH RISK: Sensitive resources with danger flag
	Unmodified   int `json:"unmodified"`   // UNMODIFIED: Resources with no changes (no-op)
	Total        int `json:"total"`        // TOTAL: Total number of resource changes across all categories
}

// IsDestructive returns true if the change type is considered destructive
func (ct ChangeType) IsDestructive() bool {
	return ct == ChangeTypeDelete || ct == ChangeTypeReplace
}

// FromTerraformAction converts a Terraform action to our ChangeType
func FromTerraformAction(actions tfjson.Actions) ChangeType {
	if len(actions) == 0 {
		return ChangeTypeNoOp
	}

	// Handle multiple actions (e.g., delete + create = replace)
	if len(actions) > 1 {
		hasDelete := false
		hasCreate := false
		for _, action := range actions {
			if action == tfjson.ActionDelete {
				hasDelete = true
			}
			if action == tfjson.ActionCreate {
				hasCreate = true
			}
		}
		if hasDelete && hasCreate {
			return ChangeTypeReplace
		}
	}

	// Handle single actions
	switch actions[0] {
	case tfjson.ActionCreate:
		return ChangeTypeCreate
	case tfjson.ActionUpdate:
		return ChangeTypeUpdate
	case tfjson.ActionDelete:
		return ChangeTypeDelete
	case "replace":
		return ChangeTypeReplace
	case "no-op":
		return ChangeTypeNoOp
	default:
		return ChangeTypeNoOp
	}
}
