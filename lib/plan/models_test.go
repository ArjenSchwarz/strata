package plan

import (
	"encoding/json"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestResourceChange_SerializationWithNewFields(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceChange
		wantJSON string
	}{
		{
			name: "resource change with enhanced fields",
			resource: ResourceChange{
				Address:          "aws_s3_bucket.example",
				Type:             "aws_s3_bucket",
				Name:             "example",
				ChangeType:       ChangeTypeCreate,
				IsDestructive:    false,
				ReplacementType:  ReplacementNever,
				Provider:         "aws",
				TopChanges:       []string{"bucket", "versioning", "encryption"},
				ReplacementHints: []string{"Bucket name changes require replacement"},
			},
			wantJSON: `{"address":"aws_s3_bucket.example","type":"aws_s3_bucket","name":"example","change_type":"create","is_destructive":false,"replacement_type":"Never","physical_id":"","planned_id":"","module_path":"","change_attributes":null,"is_dangerous":false,"danger_reason":"","danger_properties":null,"provider":"aws","top_changes":["bucket","versioning","encryption"],"replacement_hints":["Bucket name changes require replacement"]}`,
		},
		{
			name: "resource change with empty enhanced fields",
			resource: ResourceChange{
				Address:         "azurerm_resource_group.example",
				Type:            "azurerm_resource_group",
				Name:            "example",
				ChangeType:      ChangeTypeUpdate,
				IsDestructive:   false,
				ReplacementType: ReplacementNever,
				Provider:        "azurerm",
			},
			wantJSON: `{"address":"azurerm_resource_group.example","type":"azurerm_resource_group","name":"example","change_type":"update","is_destructive":false,"replacement_type":"Never","physical_id":"","planned_id":"","module_path":"","change_attributes":null,"is_dangerous":false,"danger_reason":"","danger_properties":null,"provider":"azurerm"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.resource)
			if err != nil {
				t.Fatalf("Failed to marshal ResourceChange: %v", err)
			}

			if string(jsonData) != tt.wantJSON {
				t.Errorf("JSON marshaling failed\nexpected: %s\ngot:      %s", tt.wantJSON, string(jsonData))
			}

			// Test unmarshaling
			var unmarshaled ResourceChange
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal ResourceChange: %v", err)
			}

			// Compare key fields
			if unmarshaled.Address != tt.resource.Address {
				t.Errorf("Address mismatch: expected %s, got %s", tt.resource.Address, unmarshaled.Address)
			}
			if unmarshaled.Provider != tt.resource.Provider {
				t.Errorf("Provider mismatch: expected %s, got %s", tt.resource.Provider, unmarshaled.Provider)
			}
			if len(unmarshaled.TopChanges) != len(tt.resource.TopChanges) {
				t.Errorf("TopChanges length mismatch: expected %d, got %d", len(tt.resource.TopChanges), len(unmarshaled.TopChanges))
			}
			if len(unmarshaled.ReplacementHints) != len(tt.resource.ReplacementHints) {
				t.Errorf("ReplacementHints length mismatch: expected %d, got %d", len(tt.resource.ReplacementHints), len(unmarshaled.ReplacementHints))
			}
		})
	}
}

func TestFromTerraformAction(t *testing.T) {
	tests := []struct {
		name     string
		actions  tfjson.Actions
		expected ChangeType
	}{
		{
			name:     "empty actions",
			actions:  tfjson.Actions{},
			expected: ChangeTypeNoOp,
		},
		{
			name:     "create action",
			actions:  tfjson.Actions{tfjson.ActionCreate},
			expected: ChangeTypeCreate,
		},
		{
			name:     "update action",
			actions:  tfjson.Actions{tfjson.ActionUpdate},
			expected: ChangeTypeUpdate,
		},
		{
			name:     "delete action",
			actions:  tfjson.Actions{tfjson.ActionDelete},
			expected: ChangeTypeDelete,
		},
		{
			name:     "delete and create actions (replace)",
			actions:  tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
			expected: ChangeTypeReplace,
		},
		{
			name:     "create and delete actions (replace)",
			actions:  tfjson.Actions{tfjson.ActionCreate, tfjson.ActionDelete},
			expected: ChangeTypeReplace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromTerraformAction(tt.actions)
			if result != tt.expected {
				t.Errorf("FromTerraformAction() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestChangeType_IsDestructive(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		expected   bool
	}{
		{
			name:       "create is not destructive",
			changeType: ChangeTypeCreate,
			expected:   false,
		},
		{
			name:       "update is not destructive",
			changeType: ChangeTypeUpdate,
			expected:   false,
		},
		{
			name:       "delete is destructive",
			changeType: ChangeTypeDelete,
			expected:   true,
		},
		{
			name:       "replace is destructive",
			changeType: ChangeTypeReplace,
			expected:   true,
		},
		{
			name:       "no-op is not destructive",
			changeType: ChangeTypeNoOp,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.changeType.IsDestructive()
			if result != tt.expected {
				t.Errorf("IsDestructive() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
