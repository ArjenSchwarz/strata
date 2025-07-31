package plan

import (
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
)

func TestFormatterReplacementIndicators(t *testing.T) {
	// Create a sample ResourceChange with ReplacePaths
	resourceChange := &tfjson.ResourceChange{
		Address: "aws_instance.example",
		Type:    "aws_instance",
		Name:    "example",
		Change: &tfjson.Change{
			Actions: []tfjson.Action{tfjson.ActionDelete, tfjson.ActionCreate},
			Before: map[string]any{
				"ami":           "ami-12345",
				"instance_type": "t3.micro",
			},
			After: map[string]any{
				"ami":           "ami-67890",
				"instance_type": "t3.small",
			},
			ReplacePaths: []any{
				[]any{"ami"}, // Only AMI changes force replacement
			},
		},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{ShowDetails: true},
	}

	analyzer := NewAnalyzer(&tfjson.Plan{}, cfg)
	propAnalysis := analyzer.AnalyzePropertyChanges(resourceChange)

	// Test formatter
	formatter := NewFormatter(cfg)

	// Test formatPropertyChange directly
	for _, propChange := range propAnalysis.Changes {
		formatted := formatter.formatPropertyChange(propChange)
		t.Logf("Property: %s, Formatted: %s", propChange.Name, formatted)

		if propChange.Name == "ami" && propChange.TriggersReplacement {
			if !strings.Contains(formatted, "# forces replacement") {
				t.Errorf("Expected AMI change to show replacement indicator, got: %s", formatted)
			}
		}

		if propChange.Name == "instance_type" && !propChange.TriggersReplacement {
			if strings.Contains(formatted, "# forces replacement") {
				t.Errorf("Expected instance_type change to NOT show replacement indicator, got: %s", formatted)
			}
		}
	}

	t.Log("✅ Formatter replacement indicators working correctly!")
}
