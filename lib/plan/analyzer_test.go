package plan

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestAnalyzer_analyzeReplacementNecessity(t *testing.T) {
	tests := []struct {
		name     string
		change   *tfjson.ResourceChange
		expected ReplacementType
	}{
		{
			name: "create action - never replacement",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
				},
			},
			expected: ReplacementNever,
		},
		{
			name: "update action - never replacement",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
				},
			},
			expected: ReplacementNever,
		},
		{
			name: "delete action - never replacement",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete},
				},
			},
			expected: ReplacementNever,
		},
		{
			name: "replace action - always replacement",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
				},
			},
			expected: ReplacementAlways,
		},
	}

	analyzer := &Analyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeReplacementNecessity(tt.change)
			if result != tt.expected {
				t.Errorf("analyzeReplacementNecessity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnalyzer_isConditionalReplacement(t *testing.T) {
	tests := []struct {
		name     string
		change   *tfjson.ResourceChange
		expected bool
	}{
		{
			name: "create action - not conditional",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
				},
			},
			expected: false,
		},
		{
			name: "replace action - not conditional (always)",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
				},
			},
			expected: false,
		},
	}

	analyzer := &Analyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isConditionalReplacement(tt.change)
			if result != tt.expected {
				t.Errorf("isConditionalReplacement() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnalyzer_calculateStatistics(t *testing.T) {
	tests := []struct {
		name     string
		changes  []ResourceChange
		expected ChangeStatistics
	}{
		{
			name:    "empty changes",
			changes: []ResourceChange{},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				Conditionals: 0,
				Total:        0,
			},
		},
		{
			name: "mixed changes",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeUpdate, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeDelete, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementConditional},
			},
			expected: ChangeStatistics{
				ToAdd:        1,
				ToChange:     1,
				ToDestroy:    1,
				Replacements: 1,
				Conditionals: 1,
				Total:        5,
			},
		},
		{
			name: "only replacements",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementConditional},
			},
			expected: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 2,
				Conditionals: 1,
				Total:        3,
			},
		},
	}

	analyzer := &Analyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateStatistics(tt.changes)
			if result != tt.expected {
				t.Errorf("calculateStatistics() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestAnalyzer_analyzeResourceChanges(t *testing.T) {
	plan := &tfjson.Plan{
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_instance.web",
				Type:    "aws_instance",
				Name:    "web",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
					Before:  nil,
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
			{
				Address: "aws_instance.old",
				Type:    "aws_instance",
				Name:    "old",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete, tfjson.ActionCreate},
					Before:  map[string]interface{}{"instance_type": "t2.small"},
					After:   map[string]interface{}{"instance_type": "t2.micro"},
				},
			},
		},
	}

	analyzer := NewAnalyzer(plan)
	changes := analyzer.analyzeResourceChanges()

	if len(changes) != 2 {
		t.Errorf("analyzeResourceChanges() returned %d changes, want 2", len(changes))
	}

	// Check first change (create)
	if changes[0].Address != "aws_instance.web" {
		t.Errorf("First change address = %s, want aws_instance.web", changes[0].Address)
	}
	if changes[0].ChangeType != ChangeTypeCreate {
		t.Errorf("First change type = %s, want %s", changes[0].ChangeType, ChangeTypeCreate)
	}
	if changes[0].ReplacementType != ReplacementNever {
		t.Errorf("First change replacement type = %s, want %s", changes[0].ReplacementType, ReplacementNever)
	}

	// Check second change (replace)
	if changes[1].Address != "aws_instance.old" {
		t.Errorf("Second change address = %s, want aws_instance.old", changes[1].Address)
	}
	if changes[1].ChangeType != ChangeTypeReplace {
		t.Errorf("Second change type = %s, want %s", changes[1].ChangeType, ChangeTypeReplace)
	}
	if changes[1].ReplacementType != ReplacementAlways {
		t.Errorf("Second change replacement type = %s, want %s", changes[1].ReplacementType, ReplacementAlways)
	}
}

func TestNewAnalyzer(t *testing.T) {
	plan := &tfjson.Plan{}
	analyzer := NewAnalyzer(plan)

	if analyzer.plan != plan {
		t.Error("NewAnalyzer() should set plan correctly")
	}
}
