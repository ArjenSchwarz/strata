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
			wantJSON: `{"address":"aws_s3_bucket.example","type":"aws_s3_bucket","name":"example","change_type":"create","is_destructive":false,"replacement_type":"Never","physical_id":"","planned_id":"","module_path":"","change_attributes":null,"is_dangerous":false,"danger_reason":"","danger_properties":null,"provider":"aws","top_changes":["bucket","versioning","encryption"],"replacement_hints":["Bucket name changes require replacement"],"property_changes":{"changes":null,"count":0,"total_size_bytes":0,"truncated":false}}`,
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
			wantJSON: `{"address":"azurerm_resource_group.example","type":"azurerm_resource_group","name":"example","change_type":"update","is_destructive":false,"replacement_type":"Never","physical_id":"","planned_id":"","module_path":"","change_attributes":null,"is_dangerous":false,"danger_reason":"","danger_properties":null,"provider":"azurerm","property_changes":{"changes":null,"count":0,"total_size_bytes":0,"truncated":false}}`,
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

func TestResourceAnalysis_Serialization(t *testing.T) {
	tests := []struct {
		name     string
		analysis ResourceAnalysis
		wantJSON string
	}{
		{
			name: "complete resource analysis",
			analysis: ResourceAnalysis{
				PropertyChanges: PropertyChangeAnalysis{
					Changes: []PropertyChange{
						{
							Name:      "instance_type",
							Path:      []string{"instance_type"},
							Before:    "t3.micro",
							After:     "t3.small",
							Sensitive: false,
							Size:      20,
							Action:    "update",
						},
					},
					Count:     1,
					TotalSize: 20,
					Truncated: false,
				},
				ReplacementReasons: []string{"Instance type changes require replacement"},
				RiskLevel:          "medium",
			},
			wantJSON: `{"property_changes":{"changes":[{"name":"instance_type","path":["instance_type"],"before":"t3.micro","after":"t3.small","sensitive":false,"size":20,"action":"update"}],"count":1,"total_size_bytes":20,"truncated":false},"replacement_reasons":["Instance type changes require replacement"],"risk_level":"medium"}`,
		},
		{
			name: "analysis with truncated properties",
			analysis: ResourceAnalysis{
				PropertyChanges: PropertyChangeAnalysis{
					Changes:   []PropertyChange{},
					Count:     150,
					TotalSize: 2097152, // 2MB
					Truncated: true,
				},
				ReplacementReasons: []string{},
				RiskLevel:          "high",
			},
			wantJSON: `{"property_changes":{"changes":[],"count":150,"total_size_bytes":2097152,"truncated":true},"replacement_reasons":[],"risk_level":"high"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.analysis)
			if err != nil {
				t.Fatalf("Failed to marshal ResourceAnalysis: %v", err)
			}

			if string(jsonData) != tt.wantJSON {
				t.Errorf("JSON marshaling failed\nexpected: %s\ngot:      %s", tt.wantJSON, string(jsonData))
			}

			// Test unmarshaling
			var unmarshaled ResourceAnalysis
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal ResourceAnalysis: %v", err)
			}

			// Compare key fields
			if unmarshaled.RiskLevel != tt.analysis.RiskLevel {
				t.Errorf("RiskLevel mismatch: expected %s, got %s", tt.analysis.RiskLevel, unmarshaled.RiskLevel)
			}
			if unmarshaled.PropertyChanges.Count != tt.analysis.PropertyChanges.Count {
				t.Errorf("PropertyChanges.Count mismatch: expected %d, got %d", tt.analysis.PropertyChanges.Count, unmarshaled.PropertyChanges.Count)
			}
			if unmarshaled.PropertyChanges.Truncated != tt.analysis.PropertyChanges.Truncated {
				t.Errorf("PropertyChanges.Truncated mismatch: expected %v, got %v", tt.analysis.PropertyChanges.Truncated, unmarshaled.PropertyChanges.Truncated)
			}
		})
	}
}

func TestPropertyChange_SensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		change   PropertyChange
		expected bool
	}{
		{
			name: "sensitive property",
			change: PropertyChange{
				Name:      "password",
				Before:    "secret123",
				After:     "newsecret456",
				Sensitive: true,
			},
			expected: true,
		},
		{
			name: "non-sensitive property",
			change: PropertyChange{
				Name:      "instance_type",
				Before:    "t3.micro",
				After:     "t3.small",
				Sensitive: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.change.Sensitive != tt.expected {
				t.Errorf("Sensitive flag = %v, expected %v", tt.change.Sensitive, tt.expected)
			}
		})
	}
}
