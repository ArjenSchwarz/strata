package plan

import (
	"reflect"
	"testing"
)

// TestSortResourceTableData tests the sortResourceTableData function
// This tests danger sorting, action priority sorting, and alphabetical sorting
func TestSortResourceTableData(t *testing.T) {
	tests := []struct {
		name     string
		input    []map[string]any
		expected []map[string]any
	}{
		{
			name: "danger_sorting_first",
			input: []map[string]any{
				{"ActionType": "Add", "Resource": "aws_instance.safe", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "aws_rds.db", "IsDangerous": true},
				{"ActionType": "Modify", "Resource": "aws_s3.bucket", "IsDangerous": false},
			},
			expected: []map[string]any{
				{"ActionType": "Remove", "Resource": "aws_rds.db", "IsDangerous": true},
				{"ActionType": "Modify", "Resource": "aws_s3.bucket", "IsDangerous": false},
				{"ActionType": "Add", "Resource": "aws_instance.safe", "IsDangerous": false},
			},
		},
		{
			name: "action_priority_sorting",
			input: []map[string]any{
				{"ActionType": "Add", "Resource": "resource_a", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "resource_b", "IsDangerous": false},
				{"ActionType": "Replace", "Resource": "resource_c", "IsDangerous": false},
				{"ActionType": "Modify", "Resource": "resource_d", "IsDangerous": false},
			},
			expected: []map[string]any{
				{"ActionType": "Remove", "Resource": "resource_b", "IsDangerous": false},
				{"ActionType": "Replace", "Resource": "resource_c", "IsDangerous": false},
				{"ActionType": "Modify", "Resource": "resource_d", "IsDangerous": false},
				{"ActionType": "Add", "Resource": "resource_a", "IsDangerous": false},
			},
		},
		{
			name: "alphabetical_sorting_within_priority",
			input: []map[string]any{
				{"ActionType": "Remove", "Resource": "z_resource", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "a_resource", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "m_resource", "IsDangerous": false},
			},
			expected: []map[string]any{
				{"ActionType": "Remove", "Resource": "a_resource", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "m_resource", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "z_resource", "IsDangerous": false},
			},
		},
		{
			name: "combined_sorting_logic",
			input: []map[string]any{
				{"ActionType": "Add", "Resource": "safe_resource", "IsDangerous": false},
				{"ActionType": "Remove", "Resource": "z_dangerous", "IsDangerous": true},
				{"ActionType": "Replace", "Resource": "a_safe", "IsDangerous": false},
				{"ActionType": "Modify", "Resource": "a_dangerous", "IsDangerous": true},
			},
			expected: []map[string]any{
				{"ActionType": "Remove", "Resource": "z_dangerous", "IsDangerous": true},
				{"ActionType": "Modify", "Resource": "a_dangerous", "IsDangerous": true},
				{"ActionType": "Replace", "Resource": "a_safe", "IsDangerous": false},
				{"ActionType": "Add", "Resource": "safe_resource", "IsDangerous": false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of input to avoid modifying the original test data
			input := make([]map[string]any, len(tt.input))
			for i, item := range tt.input {
				input[i] = make(map[string]any)
				for k, v := range item {
					input[i][k] = v
				}
			}

			sortResourceTableData(input)

			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("sortResourceTableData() failed")
				t.Errorf("Expected: %+v", tt.expected)
				t.Errorf("Got:      %+v", input)
			}
		})
	}
}

// TestGetActionPriority tests the getActionPriority function with all action types
func TestGetActionPriority(t *testing.T) {
	tests := []struct {
		action   string
		expected int
	}{
		{"Remove", 0},
		{"Replace", 1},
		{"Modify", 2},
		{"Add", 3},
		{"Unknown", 4},
		{"", 4},
		{"Invalid", 4},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := getActionPriority(tt.action)
			if result != tt.expected {
				t.Errorf("getActionPriority(%q) = %d, expected %d", tt.action, result, tt.expected)
			}
		})
	}
}

// TestApplyDecorations tests the applyDecorations function
// Verifies emoji application and field cleanup
func TestApplyDecorations(t *testing.T) {
	tests := []struct {
		name     string
		input    []map[string]any
		expected []map[string]any
	}{
		{
			name: "dangerous_action_gets_emoji",
			input: []map[string]any{
				{"ActionType": "Remove", "IsDangerous": true, "Resource": "test"},
			},
			expected: []map[string]any{
				{"Action": "⚠️ Remove", "Resource": "test"},
			},
		},
		{
			name: "safe_action_no_emoji",
			input: []map[string]any{
				{"ActionType": "Add", "IsDangerous": false, "Resource": "test"},
			},
			expected: []map[string]any{
				{"Action": "Add", "Resource": "test"},
			},
		},
		{
			name: "mixed_dangerous_and_safe",
			input: []map[string]any{
				{"ActionType": "Remove", "IsDangerous": true, "Resource": "dangerous_db"},
				{"ActionType": "Add", "IsDangerous": false, "Resource": "safe_instance"},
			},
			expected: []map[string]any{
				{"Action": "⚠️ Remove", "Resource": "dangerous_db"},
				{"Action": "Add", "Resource": "safe_instance"},
			},
		},
		{
			name: "internal_fields_removed",
			input: []map[string]any{
				{"ActionType": "Modify", "IsDangerous": false, "Resource": "test", "OtherField": "keep"},
			},
			expected: []map[string]any{
				{"Action": "Modify", "Resource": "test", "OtherField": "keep"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a deep copy of input to avoid modifying the original test data
			input := make([]map[string]any, len(tt.input))
			for i, item := range tt.input {
				input[i] = make(map[string]any)
				for k, v := range item {
					input[i][k] = v
				}
			}

			applyDecorations(input)

			if !reflect.DeepEqual(input, tt.expected) {
				t.Errorf("applyDecorations() failed")
				t.Errorf("Expected: %+v", tt.expected)
				t.Errorf("Got:      %+v", input)
			}

			// Verify that ActionType and IsDangerous fields are removed
			for _, row := range input {
				if _, exists := row["ActionType"]; exists {
					t.Error("ActionType field should be removed after decoration")
				}
				if _, exists := row["IsDangerous"]; exists {
					t.Error("IsDangerous field should be removed after decoration")
				}
			}
		})
	}
}

// TestDataPipelineSortingEdgeCases tests edge cases: empty data, missing fields, null values
func TestDataPipelineSortingEdgeCases(t *testing.T) {
	t.Run("empty_data", func(t *testing.T) {
		var emptyData []map[string]any
		sortResourceTableData(emptyData)
		if len(emptyData) != 0 {
			t.Error("Empty data should remain empty")
		}

		applyDecorations(emptyData)
		if len(emptyData) != 0 {
			t.Error("Empty data should remain empty after decoration")
		}
	})

	t.Run("missing_fields", func(t *testing.T) {
		data := []map[string]any{
			{"Resource": "test1"}, // Missing ActionType and IsDangerous
			{"ActionType": "Add"}, // Missing Resource and IsDangerous
			{},                    // Empty map
		}

		// Should not panic
		sortResourceTableData(data)
		applyDecorations(data)

		// Verify graceful handling of missing fields
		if len(data) != 3 {
			t.Error("Should preserve all rows even with missing fields")
		}
	})

	t.Run("null_values", func(t *testing.T) {
		data := []map[string]any{
			{"ActionType": nil, "Resource": "test", "IsDangerous": nil},
			{"ActionType": "Add", "Resource": nil, "IsDangerous": false},
		}

		// Should not panic
		sortResourceTableData(data)
		applyDecorations(data)

		// Verify graceful handling of nil values
		if len(data) != 2 {
			t.Error("Should preserve all rows even with nil values")
		}
	})

	t.Run("wrong_types", func(t *testing.T) {
		data := []map[string]any{
			{"ActionType": 123, "Resource": true, "IsDangerous": "not_bool"},
		}

		// Should not panic
		sortResourceTableData(data)
		applyDecorations(data)

		// Verify graceful handling of wrong types
		if len(data) != 1 {
			t.Error("Should preserve row even with wrong types")
		}
	})

	t.Run("stable_sorting", func(t *testing.T) {
		// Test that sorting is stable - equal elements maintain relative order
		data := []map[string]any{
			{"ActionType": "Add", "Resource": "resource1", "IsDangerous": false, "Order": 1},
			{"ActionType": "Add", "Resource": "resource1", "IsDangerous": false, "Order": 2},
			{"ActionType": "Add", "Resource": "resource1", "IsDangerous": false, "Order": 3},
		}

		sortResourceTableData(data)

		// Verify order is preserved for equal elements
		if data[0]["Order"] != 1 || data[1]["Order"] != 2 || data[2]["Order"] != 3 {
			t.Error("Stable sort should preserve original order for equal elements")
		}
	})
}
