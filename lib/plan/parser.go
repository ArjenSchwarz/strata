package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
)

// Parser handles Terraform plan file parsing
type Parser struct {
	planFile string
}

// NewParser creates a new plan parser instance
func NewParser(planFile string) *Parser {
	return &Parser{
		planFile: planFile,
	}
}

// LoadPlan loads and parses a Terraform plan file
func (p *Parser) LoadPlan() (*tfjson.Plan, error) {
	// Check if file exists
	if _, err := os.Stat(p.planFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("plan file does not exist: %s", p.planFile)
	}

	// Determine if we need to convert the plan to JSON
	var jsonData []byte
	var err error

	if strings.HasSuffix(p.planFile, ".json") {
		// Already a JSON file, read directly
		jsonData, err = os.ReadFile(p.planFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read plan file: %w", err)
		}
	} else {
		// Binary plan file, convert to JSON using terraform show
		jsonData, err = p.convertPlanToJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to convert plan to JSON: %w", err)
		}
	}

	// Parse the JSON
	var plan tfjson.Plan
	if err := json.Unmarshal(jsonData, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	return &plan, nil
}

// convertPlanToJSON converts a binary plan file to JSON using terraform show
func (p *Parser) convertPlanToJSON() ([]byte, error) {
	// Get the directory containing the plan file
	planDir := filepath.Dir(p.planFile)

	// Execute terraform show -json
	cmd := exec.Command("terraform", "show", "-json", p.planFile)
	cmd.Dir = planDir

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("terraform show failed: %s", string(exitError.Stderr))
		}
		return nil, fmt.Errorf("failed to execute terraform show: %w", err)
	}

	return output, nil
}

// ValidateStructure validates that the plan has the expected structure
func (p *Parser) ValidateStructure(plan *tfjson.Plan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}

	if plan.FormatVersion == "" {
		return fmt.Errorf("plan format version is missing")
	}

	// Check if we have resource changes (can be nil for no-op plans)
	if plan.ResourceChanges == nil {
		// This is valid for plans with no changes
		return nil
	}

	return nil
}

// extractWorkspaceInfo extracts workspace information from the plan
func (p *Parser) extractWorkspaceInfo(_ *tfjson.Plan) string {
	// Try to get workspace from plan metadata
	// For now, return "default" as a fallback
	// TODO: Extract actual workspace from plan when available
	return "default"
}

// extractBackendInfo extracts backend configuration from the plan
func (p *Parser) extractBackendInfo(_ *tfjson.Plan) BackendInfo {
	// Try to extract backend info from plan
	// For now, return local backend as fallback
	// TODO: Extract actual backend info from plan when available
	return BackendInfo{
		Type:     "local",
		Location: "terraform.tfstate",
		Config:   make(map[string]interface{}),
	}
}

// getPlanFileInfo gets file information including creation time
func (p *Parser) getPlanFileInfo(filePath string) (time.Time, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get file info: %w", err)
	}
	return fileInfo.ModTime(), nil
}
