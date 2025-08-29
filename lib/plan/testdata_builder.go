package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

const (
	testdataActionUpdate = "update"
)

// PlanBuilder provides a functional builder pattern for creating test Terraform plans
type PlanBuilder struct {
	formatVersion    string
	terraformVersion string
	resources        []tfjson.ResourceChange
	outputs          map[string]OutputChange
	variables        map[string]any
}

// NewPlanBuilder creates a new plan builder with sensible defaults
func NewPlanBuilder() *PlanBuilder {
	return &PlanBuilder{
		formatVersion:    "1.2",
		terraformVersion: "1.8.5",
		resources:        []tfjson.ResourceChange{},
		outputs:          make(map[string]OutputChange),
		variables:        make(map[string]any),
	}
}

// WithFormatVersion sets the plan format version
func (b *PlanBuilder) WithFormatVersion(version string) *PlanBuilder {
	b.formatVersion = version
	return b
}

// WithTerraformVersion sets the Terraform version
func (b *PlanBuilder) WithTerraformVersion(version string) *PlanBuilder {
	b.terraformVersion = version
	return b
}

// AddResource adds a resource change to the plan
func (b *PlanBuilder) AddResource(resource tfjson.ResourceChange) *PlanBuilder {
	b.resources = append(b.resources, resource)
	return b
}

// AddSimpleResource adds a simple resource change with basic configuration
func (b *PlanBuilder) AddSimpleResource(provider, resourceType, name, action string) *PlanBuilder {
	var actions []tfjson.Action
	switch action {
	case "create":
		actions = []tfjson.Action{tfjson.ActionCreate}
	case testdataActionUpdate:
		actions = []tfjson.Action{tfjson.ActionUpdate}
	case "delete":
		actions = []tfjson.Action{tfjson.ActionDelete}
	case "replace":
		actions = []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate}
	default:
		actions = []tfjson.Action{tfjson.ActionNoop}
	}

	resource := tfjson.ResourceChange{
		Address:      fmt.Sprintf("%s.%s", resourceType, name),
		Mode:         tfjson.ManagedResourceMode,
		Type:         resourceType,
		Name:         name,
		ProviderName: fmt.Sprintf("registry.terraform.io/hashicorp/%s", provider),
		Change: &tfjson.Change{
			Actions: actions,
			After: map[string]any{
				"name": name,
				"type": "test",
			},
		},
	}

	if action != "create" {
		resource.Change.Before = map[string]any{
			"name": name,
			"type": "old_test",
		}
	}

	return b.AddResource(resource)
}

// AddMultiProviderResources adds resources from multiple providers for realistic testing
func (b *PlanBuilder) AddMultiProviderResources(count int) *PlanBuilder {
	providers := []string{"aws", "azurerm", "google", "kubernetes"}
	actions := []string{"create", "update", "delete"}

	for i := range count {
		provider := providers[i%len(providers)]
		action := actions[i%len(actions)]
		resourceType := fmt.Sprintf("%s_instance", provider)
		if provider == "kubernetes" {
			resourceType = "kubernetes_deployment"
		}
		name := fmt.Sprintf("resource_%d", i)

		b.AddSimpleResource(provider, resourceType, name, action)
	}

	return b
}

// AddOutput adds an output change to the plan
func (b *PlanBuilder) AddOutput(name string, change OutputChange) *PlanBuilder {
	b.outputs[name] = change
	return b
}

// AddVariable adds a variable to the plan
func (b *PlanBuilder) AddVariable(name string, value any) *PlanBuilder {
	b.variables[name] = value
	return b
}

// Build creates the final terraform-json Plan
func (b *PlanBuilder) Build() *tfjson.Plan {
	// Convert resources to pointers
	resourceChanges := make([]*tfjson.ResourceChange, len(b.resources))
	for i, resource := range b.resources {
		resourceChanges[i] = &resource
	}

	// Create variables map with proper type
	variables := make(map[string]*tfjson.PlanVariable)
	for name, value := range b.variables {
		variables[name] = &tfjson.PlanVariable{
			Value: value,
		}
	}

	return &tfjson.Plan{
		FormatVersion:    b.formatVersion,
		TerraformVersion: b.terraformVersion,
		Variables:        variables,
		ResourceChanges:  resourceChanges,
		PlannedValues: &tfjson.StateValues{
			RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{},
			},
		},
		PriorState: &tfjson.State{
			FormatVersion:    "1.0",
			TerraformVersion: b.terraformVersion,
			Values: &tfjson.StateValues{
				RootModule: &tfjson.StateModule{},
			},
		},
	}
}

// BuildJSON creates the plan and marshals it to JSON
func (b *PlanBuilder) BuildJSON() ([]byte, error) {
	plan := b.Build()
	return json.MarshalIndent(plan, "", "  ")
}

// SaveToTestdata creates the plan and saves it to the testdata directory
func (b *PlanBuilder) SaveToTestdata(filename string) (string, error) {
	planJSON, err := b.BuildJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal plan JSON: %w", err)
	}

	// Ensure the filename has .json extension
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	testdataDir := "testdata"
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create testdata directory: %w", err)
	}

	planPath := filepath.Join(testdataDir, filename)
	if err := os.WriteFile(planPath, planJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write plan file: %w", err)
	}

	return planPath, nil
}

// SaveToTempFile creates the plan and saves it to a temporary file (for benchmarks)
func (b *PlanBuilder) SaveToTempFile(filename string) (string, error) {
	planJSON, err := b.BuildJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal plan JSON: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "strata-test-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	planPath := filepath.Join(tempDir, filename)
	if err := os.WriteFile(planPath, planJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write plan file: %w", err)
	}

	return planPath, nil
}

// Helper functions for common test scenarios

// CreateSimplePlan creates a basic plan with a few resources
func CreateSimplePlan() *PlanBuilder {
	return NewPlanBuilder().
		AddSimpleResource("aws", "aws_instance", "web", "create").
		AddSimpleResource("aws", "aws_s3_bucket", "data", "update")
}

// CreateMultiProviderPlan creates a plan with resources from multiple providers
func CreateMultiProviderPlan(resourceCount int) *PlanBuilder {
	return NewPlanBuilder().AddMultiProviderResources(resourceCount)
}

// CreateHighRiskPlan creates a plan with potentially dangerous changes
func CreateHighRiskPlan() *PlanBuilder {
	return NewPlanBuilder().
		AddSimpleResource("aws", "aws_db_instance", "main", "replace").
		AddSimpleResource("aws", "aws_instance", "web", "delete").
		AddSimpleResource("azurerm", "azurerm_sql_database", "prod", "replace")
}

// CreateEmptyPlan creates a plan with no changes
func CreateEmptyPlan() *PlanBuilder {
	return NewPlanBuilder()
}

// AddPropertyHeavyResource adds a resource with many properties for performance testing
func (b *PlanBuilder) AddPropertyHeavyResource(resourceType, name string, propertyCount, propertySize int) *PlanBuilder {
	before := make(map[string]any)
	after := make(map[string]any)

	for i := range propertyCount {
		value := strings.Repeat("x", propertySize)
		before[fmt.Sprintf("property_%d", i)] = value
		after[fmt.Sprintf("property_%d", i)] = value + "_updated"
	}

	resource := tfjson.ResourceChange{
		Address:      fmt.Sprintf("%s.%s", resourceType, name),
		Mode:         tfjson.ManagedResourceMode,
		Type:         resourceType,
		Name:         name,
		ProviderName: "registry.terraform.io/hashicorp/aws",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionUpdate},
			Before:  before,
			After:   after,
		},
	}

	return b.AddResource(resource)
}

// CreatePropertyBenchmarkPlan creates a plan optimized for property performance testing
func CreatePropertyBenchmarkPlan(propertyCount, propertySize int) *PlanBuilder {
	return NewPlanBuilder().AddPropertyHeavyResource("aws_instance", "property_heavy", propertyCount, propertySize)
}
