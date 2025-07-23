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
	// Method 1: Check TF_WORKSPACE environment variable
	if workspace := os.Getenv("TF_WORKSPACE"); workspace != "" {
		return workspace
	}

	// Method 2: Try to execute terraform workspace show in the plan file's directory
	if workspace := p.getWorkspaceFromCLI(); workspace != "" {
		return workspace
	}

	// Method 3: Fallback to "default"
	return "default"
}

// getWorkspaceFromCLI attempts to get workspace information from terraform CLI
func (p *Parser) getWorkspaceFromCLI() string {
	// Get the directory containing the plan file
	planDir := filepath.Dir(p.planFile)

	// Execute terraform workspace show
	cmd := exec.Command("terraform", "workspace", "show")
	cmd.Dir = planDir

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Trim whitespace and return the workspace name
	workspace := strings.TrimSpace(string(output))
	if workspace == "" {
		return ""
	}

	return workspace
}

// extractBackendInfo extracts backend configuration from the plan
func (p *Parser) extractBackendInfo(_ *tfjson.Plan) BackendInfo {
	// Method 1: Try to read .terraform/terraform.tfstate
	if backend := p.getBackendFromTerraformDir(); backend.Type != "" {
		return backend
	}

	// Method 2: Fallback to local backend
	return BackendInfo{
		Type:     "local",
		Location: "terraform.tfstate",
		Config:   make(map[string]any),
	}
}

// getBackendFromTerraformDir attempts to read backend info from .terraform/terraform.tfstate
func (p *Parser) getBackendFromTerraformDir() BackendInfo {
	// Get the directory containing the plan file
	planDir := filepath.Dir(p.planFile)
	tfStateFile := filepath.Join(planDir, ".terraform", "terraform.tfstate")

	// Check if .terraform/terraform.tfstate exists
	if _, err := os.Stat(tfStateFile); os.IsNotExist(err) {
		return BackendInfo{}
	}

	// Read the file
	data, err := os.ReadFile(tfStateFile)
	if err != nil {
		return BackendInfo{}
	}

	// Parse the JSON
	var config struct {
		Backend struct {
			Type   string         `json:"type"`
			Config map[string]any `json:"config"`
		} `json:"backend"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return BackendInfo{}
	}

	// Extract location based on backend type
	location := p.extractBackendLocation(config.Backend.Type, config.Backend.Config)

	return BackendInfo{
		Type:     config.Backend.Type,
		Location: location,
		Config:   config.Backend.Config,
	}
}

// extractBackendLocation formats the backend location based on type and config
func (p *Parser) extractBackendLocation(backendType string, config map[string]any) string {
	switch backendType {
	case "s3":
		bucket, _ := config["bucket"].(string)
		key, _ := config["key"].(string)
		if bucket != "" && key != "" {
			return fmt.Sprintf("s3://%s/%s", bucket, key)
		}
	case "azurerm":
		account, _ := config["storage_account_name"].(string)
		container, _ := config["container_name"].(string)
		key, _ := config["key"].(string)
		if account != "" && container != "" && key != "" {
			return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", account, container, key)
		}
	case "gcs":
		bucket, _ := config["bucket"].(string)
		prefix, _ := config["prefix"].(string)
		if bucket != "" {
			if prefix != "" {
				return fmt.Sprintf("gs://%s/%s", bucket, prefix)
			}
			return fmt.Sprintf("gs://%s/default.tfstate", bucket)
		}
	case "remote":
		org, _ := config["organization"].(string)
		if workspaces, ok := config["workspaces"].(map[string]any); ok {
			if name, ok := workspaces["name"].(string); ok && org != "" {
				return fmt.Sprintf("app.terraform.io/%s/%s", org, name)
			}
		}
	case "local":
		if path, ok := config["path"].(string); ok && path != "" {
			return path
		}
		return "terraform.tfstate"
	}

	// Fallback for unknown or incomplete configurations
	return "terraform.tfstate"
}

// getPlanFileInfo gets file information including creation time
func (p *Parser) getPlanFileInfo(filePath string) (time.Time, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get file info: %w", err)
	}
	return fileInfo.ModTime(), nil
}
