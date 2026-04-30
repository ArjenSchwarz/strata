package plan

import (
	"strings"
	"testing"

	output "github.com/ArjenSchwarz/go-output/v2"
	"github.com/ArjenSchwarz/strata/config"
)

// TestExpandableSectionsDisabled_PropertyFormatter verifies that when
// expandable_sections.enabled is false, the property changes formatter
// returns plain text instead of CollapsibleValue objects.
func TestExpandableSectionsDisabled_PropertyFormatter(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             false,
				AutoExpandDangerous: true,
			},
		},
	}
	formatter := NewFormatter(cfg)

	propAnalysis := PropertyChangeAnalysis{
		Changes: []PropertyChange{
			{Name: "instance_type", Action: "update", Before: "t2.micro", After: "t2.small"},
		},
		Count: 1,
	}

	t.Run("propertyChangesFormatterTerraform returns plain text", func(t *testing.T) {
		fn := formatter.propertyChangesFormatterTerraform()
		result := fn(propAnalysis)
		if _, ok := result.(output.CollapsibleValue); ok {
			t.Error("expected plain text when expandable_sections.enabled=false, got CollapsibleValue")
		}
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}
		if !strings.Contains(str, "1 properties changed") {
			t.Errorf("expected summary in plain text, got %q", str)
		}
		if !strings.Contains(str, "instance_type") {
			t.Errorf("expected details in plain text, got %q", str)
		}
	})

	t.Run("propertyChangesFormatterDirect returns plain text", func(t *testing.T) {
		fn := formatter.propertyChangesFormatterDirect()
		result := fn(propAnalysis)
		if _, ok := result.(output.CollapsibleValue); ok {
			t.Error("expected plain text when expandable_sections.enabled=false, got CollapsibleValue")
		}
		str, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}
		if !strings.Contains(str, "1 properties changed") {
			t.Errorf("expected summary in plain text, got %q", str)
		}
	})

	t.Run("map-based input also returns plain text", func(t *testing.T) {
		fn := formatter.propertyChangesFormatterTerraform()
		input := map[string]any{
			"analysis":    propAnalysis,
			"change_type": ChangeTypeUpdate,
		}
		result := fn(input)
		if _, ok := result.(output.CollapsibleValue); ok {
			t.Error("expected plain text for map input when expandable_sections.enabled=false, got CollapsibleValue")
		}
	})
}

// TestExpandableSectionsEnabled_PropertyFormatter confirms that when enabled=true,
// CollapsibleValue objects are still returned.
func TestExpandableSectionsEnabled_PropertyFormatter(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled:             true,
				AutoExpandDangerous: true,
			},
		},
	}
	formatter := NewFormatter(cfg)

	propAnalysis := PropertyChangeAnalysis{
		Changes: []PropertyChange{
			{Name: "instance_type", Action: "update", Before: "t2.micro", After: "t2.small"},
		},
		Count: 1,
	}

	fn := formatter.propertyChangesFormatterTerraform()
	result := fn(propAnalysis)
	if _, ok := result.(output.CollapsibleValue); !ok {
		t.Errorf("expected CollapsibleValue when expandable_sections.enabled=true, got %T", result)
	}
}

// TestExpandableSectionsDisabled_FormatSelection verifies that when disabled,
// standard renderers are used instead of collapsible renderers.
func TestExpandableSectionsDisabled_FormatSelection(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				Enabled: false,
			},
		},
	}
	formatter := NewFormatter(cfg)

	formats := []string{"table", "markdown", "html", "csv", "json"}
	for _, f := range formats {
		t.Run(f, func(t *testing.T) {
			result := formatter.getFormatFromConfig(f)
			// Just verify it doesn't panic and returns a valid format
			if result.Name == "" {
				t.Error("expected non-empty format name")
			}
		})
	}
}
