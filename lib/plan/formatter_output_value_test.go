package plan

import (
	"testing"
)

func TestFormatOutputValue(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		sensitive bool
		isUnknown bool
		want      string
	}{
		{"unknown", "anything", false, true, knownAfterApply},
		{"sensitive", "secret", true, false, "(sensitive value)"},
		{"nil", nil, false, false, "-"},
		{"string", "hello", false, false, `"hello"`},
		{"number", float64(42), false, false, "42"},
		{"bool", true, false, false, "true"},
		{"map_stable_order", map[string]any{"b": 2, "a": 1}, false, false, `{"a":1,"b":2}`},
		{"slice", []any{"x", "y"}, false, false, `["x","y"]`},
		{"nested_map", map[string]any{"key": map[string]any{"inner": "val"}}, false, false, `{"key":{"inner":"val"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOutputValue(tt.value, tt.sensitive, tt.isUnknown)
			if got != tt.want {
				t.Errorf("formatOutputValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
