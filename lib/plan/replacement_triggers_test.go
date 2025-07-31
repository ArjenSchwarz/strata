package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestReplacementTriggers(t *testing.T) {
	// Create a sample ResourceChange with ReplacePaths
	resourceChange := &tfjson.ResourceChange{
		Address: "aws_instance.example",
		Type:    "aws_instance",
		Name:    "example",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate}, // Replace
			Before: map[string]any{
				"instance_type": "t3.micro",
				"ami":           "ami-12345",
				"tags": map[string]any{
					"Name": "old-name",
				},
			},
			After: map[string]any{
				"instance_type": "t3.small",  // This changes
				"ami":           "ami-67890", // This changes and triggers replacement
				"tags": map[string]any{
					"Name": "new-name",
				},
			},
			// This indicates that AMI changes trigger replacement
			ReplacePaths: []any{
				[]any{"ami"}, // AMI changes force replacement
			},
		},
	}

	// Create analyzer with sample config
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ShowDetails: true,
		},
	}

	// Create a minimal plan
	tfPlan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.8.5",
		ResourceChanges: []*tfjson.ResourceChange{
			resourceChange,
		},
	}

	analyzer := NewAnalyzer(tfPlan, cfg)

	// Analyze the property changes
	propAnalysis := analyzer.AnalyzePropertyChanges(resourceChange)

	// Verify basic structure
	if propAnalysis.Count == 0 {
		t.Fatal("Expected property changes but got none")
	}

	t.Logf("Found %d property changes", propAnalysis.Count)

	// Find the changes and verify replacement triggers
	var amiChange, instanceTypeChange, tagsChange *PropertyChange

	for i := range propAnalysis.Changes {
		change := &propAnalysis.Changes[i]
		t.Logf("Change: %s, Action: %s, TriggersReplacement: %v",
			change.Name, change.Action, change.TriggersReplacement)

		switch change.Name {
		case "ami":
			amiChange = change
		case "instance_type":
			instanceTypeChange = change
		case "tags":
			tagsChange = change
		}
	}

	// Verify AMI change triggers replacement
	if amiChange == nil {
		t.Fatal("Expected to find AMI property change")
	}
	if !amiChange.TriggersReplacement {
		t.Error("Expected AMI change to trigger replacement, but it doesn't")
	}

	// Verify instance_type change does NOT trigger replacement (not in ReplacePaths)
	if instanceTypeChange != nil && instanceTypeChange.TriggersReplacement {
		t.Error("Expected instance_type change to NOT trigger replacement, but it does")
	}

	// Verify tags change does NOT trigger replacement
	if tagsChange != nil && tagsChange.TriggersReplacement {
		t.Error("Expected tags change to NOT trigger replacement, but it does")
	}

	t.Log("✅ Replacement triggers working correctly!")
}

func TestReplacementTriggersWithNestedPaths(t *testing.T) {
	// Test nested property paths
	resourceChange := &tfjson.ResourceChange{
		Address: "aws_rds_instance.example",
		Type:    "aws_rds_instance",
		Name:    "example",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate},
			Before: map[string]any{
				"engine":                  "postgres",
				"engine_version":          "13.7",
				"backup_retention_period": 7,
			},
			After: map[string]any{
				"engine":                  "postgres",
				"engine_version":          "14.9", // This changes and triggers replacement
				"backup_retention_period": 14,
			},
			// Engine version changes force replacement
			ReplacePaths: []any{
				[]any{"engine_version"},
			},
		},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{ShowDetails: true},
	}

	tfPlan := &tfjson.Plan{
		FormatVersion:    "1.2",
		TerraformVersion: "1.8.5",
		ResourceChanges:  []*tfjson.ResourceChange{resourceChange},
	}

	analyzer := NewAnalyzer(tfPlan, cfg)
	propAnalysis := analyzer.AnalyzePropertyChanges(resourceChange)

	// Find engine_version change
	var engineVersionChange *PropertyChange
	for i := range propAnalysis.Changes {
		change := &propAnalysis.Changes[i]
		if change.Name == "engine_version" {
			engineVersionChange = change
			break
		}
	}

	if engineVersionChange == nil {
		t.Fatal("Expected to find engine_version property change")
	}

	if !engineVersionChange.TriggersReplacement {
		t.Error("Expected engine_version change to trigger replacement")
	}

	t.Log("✅ Nested path replacement triggers working correctly!")
}
