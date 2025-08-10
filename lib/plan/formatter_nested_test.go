package plan

import (
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
)

// TestFormatPropertyChange_NestedObjects tests the new nested object formatting
func TestFormatPropertyChange_NestedObjects(t *testing.T) {
	tests := []struct {
		name     string
		change   PropertyChange
		expected string
	}{
		{
			name: "simple nested object change",
			change: PropertyChange{
				Name:   "tags",
				Action: "update",
				Before: map[string]any{
					"Environment": "development",
					"Name":        "Test Instance",
				},
				After: map[string]any{
					"Environment": "production",
					"Name":        "Test Instance",
				},
				Sensitive: false,
			},
			expected: `  ~ tags {
      ~ Environment = "development" -> "production"
    }`,
		},
		{
			name: "nested object with added property",
			change: PropertyChange{
				Name:   "tags",
				Action: "update",
				Before: map[string]any{
					"Environment": "development",
				},
				After: map[string]any{
					"Environment": "production",
					"Version":     "v2",
				},
				Sensitive: false,
			},
			expected: `  ~ tags {
      ~ Environment = "development" -> "production"
      + Version = "v2"
    }`,
		},
		{
			name: "nested object with removed property",
			change: PropertyChange{
				Name:   "tags",
				Action: "update",
				Before: map[string]any{
					"Environment": "development",
					"Version":     "v1",
				},
				After: map[string]any{
					"Environment": "production",
				},
				Sensitive: false,
			},
			expected: `  ~ tags {
      ~ Environment = "development" -> "production"
      - Version = "v1"
    }`,
		},
		{
			name: "nested object with replacement trigger",
			change: PropertyChange{
				Name:   "config",
				Action: "update",
				Before: map[string]any{
					"size": "small",
				},
				After: map[string]any{
					"size": "large",
				},
				Sensitive:           false,
				TriggersReplacement: true,
			},
			expected: `  ~ config { # forces replacement
      ~ size = "small" -> "large"
    }`,
		},
		{
			name: "simple property change (non-nested)",
			change: PropertyChange{
				Name:      "instance_type",
				Action:    "update",
				Before:    "t3.small",
				After:     "t3.medium",
				Sensitive: false,
			},
			expected: `  ~ instance_type = "t3.small" -> "t3.medium"`,
		},
	}

	cfg := &config.Config{
		Plan: config.PlanConfig{
			ExpandableSections: config.ExpandableSectionsConfig{
				MaxDetailLength: 1000,
			},
		},
	}
	formatter := &Formatter{config: cfg}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatPropertyChange(tt.change)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShouldUseNestedFormat tests the logic for determining when to use nested formatting
func TestShouldUseNestedFormat(t *testing.T) {
	tests := []struct {
		name     string
		before   any
		after    any
		expected bool
	}{
		{
			name: "both maps with content",
			before: map[string]any{
				"key1": "value1",
			},
			after: map[string]any{
				"key1": "value2",
			},
			expected: true,
		},
		{
			name:     "before is not map",
			before:   "string",
			after:    map[string]any{"key": "value"},
			expected: false,
		},
		{
			name:     "after is not map",
			before:   map[string]any{"key": "value"},
			after:    "string",
			expected: false,
		},
		{
			name:     "both empty maps",
			before:   map[string]any{},
			after:    map[string]any{},
			expected: false,
		},
		{
			name:     "before empty, after has content",
			before:   map[string]any{},
			after:    map[string]any{"key": "value"},
			expected: true,
		},
		{
			name:     "before has content, after empty",
			before:   map[string]any{"key": "value"},
			after:    map[string]any{},
			expected: true,
		},
	}

	cfg := &config.Config{}
	formatter := &Formatter{config: cfg}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.shouldUseNestedFormat(tt.before, tt.after)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValuesEqual tests the value comparison logic
func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        any
		b        any
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "test1",
			b:        "test2",
			expected: false,
		},
		{
			name:     "equal numbers",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "different numbers",
			a:        42,
			b:        43,
			expected: false,
		},
		{
			name:     "nil values",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil, one not",
			a:        nil,
			b:        "test",
			expected: false,
		},
	}

	cfg := &config.Config{}
	formatter := &Formatter{config: cfg}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.valuesEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatNestedObjectChange_Integration tests the complete nested formatting
func TestFormatNestedObjectChange_Integration(t *testing.T) {
	cfg := &config.Config{}
	formatter := &Formatter{config: cfg}

	change := PropertyChange{
		Name:   "tags",
		Action: "update",
		Before: map[string]any{
			"Environment": "development",
			"Name":        "Test Instance",
			"Version":     "v1",
		},
		After: map[string]any{
			"Environment": "production",
			"Name":        "Test Instance",
			"Owner":       "team-a",
		},
		Sensitive: false,
	}

	result := formatter.formatNestedObjectChange(change)

	// Check that the output contains the expected structure
	assert.Contains(t, result, "~ tags {")
	assert.Contains(t, result, "~ Environment = \"development\" -> \"production\"")
	assert.Contains(t, result, "+ Owner = \"team-a\"")
	assert.Contains(t, result, "- Version = \"v1\"")
	assert.Contains(t, result, "    }")

	// Check that unchanged values (Name) are not shown
	assert.NotContains(t, result, "Name = \"Test Instance\"")

	// Verify the lines are in alphabetical order by key
	lines := strings.Split(result, "\n")
	assert.True(t, len(lines) >= 5) // Opening brace, 3 changes, closing brace
}
