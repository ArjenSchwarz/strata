package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
)

func TestFormatValue_UnknownValue(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{
			name:     "unknown value marker string",
			value:    "(known after apply)",
			expected: "(known after apply)", // Should not have quotes
		},
		{
			name:     "regular string value",
			value:    "regular value",
			expected: `"regular value"`, // Should have quotes
		},
		{
			name:     "map with unknown value",
			value:    map[string]any{"key": "(known after apply)"},
			expected: `{ key = (known after apply) }`, // Unknown value without quotes inside map
		},
		{
			name:     "array with unknown value",
			value:    []any{"value1", "(known after apply)"},
			expected: `[ "value1", (known after apply) ]`, // Unknown value without quotes inside array
		},
		{
			name: "nested structure with unknown values",
			value: map[string]any{
				"normal":  "value",
				"unknown": "(known after apply)",
				"nested": map[string]any{
					"inner": "(known after apply)",
				},
			},
			expected: `{ nested = { inner = (known after apply) }, normal = "value", unknown = (known after apply) }`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.formatValue(tc.value, false)
			assert.Equal(t, tc.expected, result, "Formatting of %s should match expected output", tc.name)
		})
	}
}

func TestFormatPropertyChange_UnknownValue(t *testing.T) {
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	tests := []struct {
		name     string
		change   PropertyChange
		expected string
	}{
		{
			name: "property becoming unknown",
			change: PropertyChange{
				Name:      "instance_id",
				Before:    "i-1234567890",
				After:     "(known after apply)",
				Action:    "update",
				IsUnknown: true,
			},
			expected: `  ~ instance_id = "i-1234567890" -> (known after apply)`,
		},
		{
			name: "new property with unknown value",
			change: PropertyChange{
				Name:      "arn",
				After:     "(known after apply)",
				Action:    "add",
				IsUnknown: true,
			},
			expected: `  + arn = (known after apply)`,
		},
		{
			name: "property remaining unknown",
			change: PropertyChange{
				Name:      "vpc_id",
				Before:    "(known after apply)",
				After:     "(known after apply)",
				Action:    "update",
				IsUnknown: true,
			},
			expected: `  ~ vpc_id = (known after apply) -> (known after apply)`,
		},
		{
			name: "complex value becoming unknown",
			change: PropertyChange{
				Name:      "security_groups",
				Before:    []any{"sg-123", "sg-456"},
				After:     "(known after apply)",
				Action:    "update",
				IsUnknown: true,
			},
			expected: `  ~ security_groups = [ "sg-123", "sg-456" ] -> (known after apply)`,
		},
		{
			name: "unknown value with replacement",
			change: PropertyChange{
				Name:                "availability_zone",
				Before:              "us-east-1a",
				After:               "(known after apply)",
				Action:              "update",
				TriggersReplacement: true,
				IsUnknown:           true,
			},
			expected: `  ~ availability_zone = "us-east-1a" -> (known after apply) # forces replacement`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatter.formatPropertyChange(tc.change)
			assert.Equal(t, tc.expected, result, "Property change formatting should handle unknown values correctly")
		})
	}
}
