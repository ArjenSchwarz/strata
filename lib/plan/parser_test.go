package plan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestParser_extractWorkspaceInfo(t *testing.T) {
	tests := []struct {
		name string
		plan *tfjson.Plan
		want string
	}{
		{
			name: "default workspace",
			plan: &tfjson.Plan{},
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			got := p.extractWorkspaceInfo(tt.plan)
			if got != tt.want {
				t.Errorf("extractWorkspaceInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_extractBackendInfo(t *testing.T) {
	tests := []struct {
		name string
		plan *tfjson.Plan
		want BackendInfo
	}{
		{
			name: "default backend",
			plan: &tfjson.Plan{},
			want: BackendInfo{
				Type:     "local",
				Location: "terraform.tfstate",
				Config:   make(map[string]any),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			got := p.extractBackendInfo(tt.plan)
			if got.Type != tt.want.Type {
				t.Errorf("extractBackendInfo().Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Location != tt.want.Location {
				t.Errorf("extractBackendInfo().Location = %v, want %v", got.Location, tt.want.Location)
			}
		})
	}
}

func TestParser_getPlanFileInfo(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tfplan")

	// Write some content to the file
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := &Parser{}
	result, err := p.getPlanFileInfo(tmpFile)
	if err != nil {
		t.Errorf("getPlanFileInfo() error = %v", err)
	}

	// Check that we got a reasonable timestamp (within the last minute)
	now := time.Now()
	if result.After(now) || result.Before(now.Add(-time.Minute)) {
		t.Errorf("getPlanFileInfo() returned unreasonable timestamp: %v", result)
	}
}

func TestParser_getPlanFileInfo_NonExistentFile(t *testing.T) {
	p := &Parser{}
	_, err := p.getPlanFileInfo("/non/existent/file")
	if err == nil {
		t.Error("getPlanFileInfo() should return error for non-existent file")
	}
}

func TestParser_LoadPlan_JSONFile(t *testing.T) {
	// Create a temporary JSON plan file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tfplan.json")

	// Create a minimal valid plan
	plan := &tfjson.Plan{
		FormatVersion:    "1.0",
		TerraformVersion: "1.6.0",
		ResourceChanges:  []*tfjson.ResourceChange{},
	}

	planJSON, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("Failed to marshal test plan: %v", err)
	}

	err = os.WriteFile(tmpFile, planJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	p := NewParser(tmpFile)
	result, err := p.LoadPlan()
	if err != nil {
		t.Errorf("LoadPlan() error = %v", err)
	}

	if result.FormatVersion != "1.0" {
		t.Errorf("LoadPlan().FormatVersion = %v, want %v", result.FormatVersion, "1.0")
	}
	if result.TerraformVersion != "1.6.0" {
		t.Errorf("LoadPlan().TerraformVersion = %v, want %v", result.TerraformVersion, "1.6.0")
	}
}

func TestParser_LoadPlan_NonExistentFile(t *testing.T) {
	p := NewParser("/non/existent/file")
	_, err := p.LoadPlan()
	if err == nil {
		t.Error("LoadPlan() should return error for non-existent file")
	}
}

func TestParser_ValidateStructure(t *testing.T) {
	tests := []struct {
		name    string
		plan    *tfjson.Plan
		wantErr bool
	}{
		{
			name:    "nil plan",
			plan:    nil,
			wantErr: true,
		},
		{
			name: "missing format version",
			plan: &tfjson.Plan{
				TerraformVersion: "1.6.0",
			},
			wantErr: true,
		},
		{
			name: "valid plan",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.6.0",
				ResourceChanges:  []*tfjson.ResourceChange{},
			},
			wantErr: false,
		},
		{
			name: "valid plan with nil resource changes",
			plan: &tfjson.Plan{
				FormatVersion:    "1.0",
				TerraformVersion: "1.6.0",
				ResourceChanges:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			err := p.ValidateStructure(tt.plan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewParser(t *testing.T) {
	planFile := "test.tfplan"
	p := NewParser(planFile)

	if p.planFile != planFile {
		t.Errorf("NewParser().planFile = %v, want %v", p.planFile, planFile)
	}
}
