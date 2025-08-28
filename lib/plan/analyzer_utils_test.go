package plan

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

// TestSortPropertiesAlphabetically tests the property sorting functionality
func TestSortPropertiesAlphabetically(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		input    []PropertyChange
		expected []PropertyChange
	}{
		{
			name: "Basic alphabetical sorting case-insensitive",
			input: []PropertyChange{
				{Name: "zebra", Path: []string{"zebra"}, Action: "update"},
				{Name: "Apple", Path: []string{"Apple"}, Action: "update"},
				{Name: "banana", Path: []string{"banana"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "Apple", Path: []string{"Apple"}, Action: "update"},
				{Name: "banana", Path: []string{"banana"}, Action: "update"},
				{Name: "zebra", Path: []string{"zebra"}, Action: "update"},
			},
		},
		{
			name: "Same name properties sorted by path hierarchy",
			input: []PropertyChange{
				{Name: "config", Path: []string{"config", "nested", "deep"}, Action: "update"},
				{Name: "config", Path: []string{"config", "basic"}, Action: "update"},
				{Name: "config", Path: []string{"config"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "config", Path: []string{"config"}, Action: "update"},
				{Name: "config", Path: []string{"config", "basic"}, Action: "update"},
				{Name: "config", Path: []string{"config", "nested", "deep"}, Action: "update"},
			},
		},
		{
			name: "Natural sort ordering with numbers",
			input: []PropertyChange{
				{Name: "prop10", Path: []string{"prop10"}, Action: "update"},
				{Name: "prop2", Path: []string{"prop2"}, Action: "update"},
				{Name: "prop1", Path: []string{"prop1"}, Action: "update"},
				{Name: "prop20", Path: []string{"prop20"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "prop1", Path: []string{"prop1"}, Action: "update"},
				{Name: "prop2", Path: []string{"prop2"}, Action: "update"},
				{Name: "prop10", Path: []string{"prop10"}, Action: "update"},
				{Name: "prop20", Path: []string{"prop20"}, Action: "update"},
			},
		},
		{
			name: "Action type tiebreaker for identical names and paths",
			input: []PropertyChange{
				{Name: "prop", Path: []string{"prop"}, Action: "add"},
				{Name: "prop", Path: []string{"prop"}, Action: "remove"},
				{Name: "prop", Path: []string{"prop"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "prop", Path: []string{"prop"}, Action: "remove"},
				{Name: "prop", Path: []string{"prop"}, Action: "update"},
				{Name: "prop", Path: []string{"prop"}, Action: "add"},
			},
		},
		{
			name: "Mixed properties with special characters",
			input: []PropertyChange{
				{Name: "user_data", Path: []string{"user_data"}, Action: "update"},
				{Name: "user-name", Path: []string{"user-name"}, Action: "update"},
				{Name: "user.config", Path: []string{"user.config"}, Action: "update"},
				{Name: "user", Path: []string{"user"}, Action: "update"},
			},
			expected: []PropertyChange{
				{Name: "user", Path: []string{"user"}, Action: "update"},
				{Name: "user-name", Path: []string{"user-name"}, Action: "update"},
				{Name: "user.config", Path: []string{"user.config"}, Action: "update"},
				{Name: "user_data", Path: []string{"user_data"}, Action: "update"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := &PropertyChangeAnalysis{Changes: tc.input}
			analyzer.sortPropertiesAlphabetically(analysis)

			assert.Equal(t, len(tc.expected), len(analysis.Changes), "Length should match")

			for i, expected := range tc.expected {
				actual := analysis.Changes[i]
				assert.Equal(t, expected.Name, actual.Name, "Property name at index %d should match", i)
				assert.Equal(t, expected.Path, actual.Path, "Property path at index %d should match", i)
				assert.Equal(t, expected.Action, actual.Action, "Property action at index %d should match", i)
			}
		})
	}
}

// TestNaturalSort tests the natural sorting implementation
func TestNaturalSort(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		s1       string
		s2       string
		expected bool
	}{
		{
			name:     "Simple alphabetical order",
			s1:       "apple",
			s2:       "banana",
			expected: true,
		},
		{
			name:     "Numbers should sort numerically",
			s1:       "item2",
			s2:       "item10",
			expected: true,
		},
		{
			name:     "Mixed text and numbers",
			s1:       "version1.2.3",
			s2:       "version1.10.1",
			expected: true,
		},
		{
			name:     "Same prefix with numbers",
			s1:       "prop1",
			s2:       "prop2",
			expected: true,
		},
		{
			name:     "Equal strings",
			s1:       "same",
			s2:       "same",
			expected: false,
		},
		{
			name:     "Leading numbers",
			s1:       "2item",
			s2:       "10item",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.naturalSort(tc.s1, tc.s2)
			assert.Equal(t, tc.expected, result, "Natural sort result should match expected")
		})
	}
}

// TestSplitNatural tests the natural string splitting functionality
func TestSplitNatural(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple text",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "Simple number",
			input:    "123",
			expected: []string{"123"},
		},
		{
			name:     "Text with number",
			input:    "item123",
			expected: []string{"item", "123"},
		},
		{
			name:     "Number then text",
			input:    "123item",
			expected: []string{"123", "item"},
		},
		{
			name:     "Complex mixed",
			input:    "version1.2.3",
			expected: []string{"version", "1", ".", "2", ".", "3"},
		},
		{
			name:     "Multiple numbers",
			input:    "item123test456",
			expected: []string{"item", "123", "test", "456"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.splitNatural(tc.input)
			assert.Equal(t, tc.expected, result, "Split result should match expected")
		})
	}
}

// TestAnalyzePropertyChangesWithSorting tests the integration of sorting with property analysis
func TestAnalyzePropertyChangesWithSorting(t *testing.T) {
	analyzer := NewAnalyzer(nil, &config.Config{})

	// Create a test resource change with unsorted properties
	resourceChange := &tfjson.ResourceChange{
		Address: "test.resource",
		Type:    "test_resource",
		Name:    "resource",
		Change: &tfjson.Change{
			Before: map[string]any{
				"zebra_config":  "old_value",
				"apple_setting": "old_value",
				"banana_option": "old_value",
			},
			After: map[string]any{
				"zebra_config":  "new_value",
				"apple_setting": "new_value",
				"banana_option": "new_value",
			},
		},
	}

	result := analyzer.analyzePropertyChanges(resourceChange)

	// Verify that properties are sorted alphabetically
	assert.True(t, len(result.Changes) >= 3, "Should have at least 3 property changes")

	// Properties should be sorted alphabetically: apple_setting, banana_option, zebra_config
	propertyNames := make([]string, len(result.Changes))
	for i, change := range result.Changes {
		propertyNames[i] = change.Name
	}

	// Check that properties are in alphabetical order
	for i := 1; i < len(propertyNames); i++ {
		assert.True(t, strings.ToLower(propertyNames[i-1]) <= strings.ToLower(propertyNames[i]),
			"Properties should be in alphabetical order: %v", propertyNames)
	}
}

// TestMaskSensitiveValue tests the sensitive value masking functionality
func TestMaskSensitiveValue(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name        string
		value       any
		isSensitive bool
		expected    any
	}{
		{
			name:        "Non-sensitive value should not be masked",
			value:       "normal_value",
			isSensitive: false,
			expected:    "normal_value",
		},
		{
			name:        "Sensitive primitive string should be masked",
			value:       "secret_password",
			isSensitive: true,
			expected:    "(sensitive value)",
		},
		{
			name:        "Sensitive primitive number should be masked",
			value:       12345,
			isSensitive: true,
			expected:    "(sensitive value)",
		},
		{
			name:        "Sensitive primitive boolean should be masked",
			value:       true,
			isSensitive: true,
			expected:    "(sensitive value)",
		},
		{
			name:        "Sensitive map should preserve structure",
			value:       map[string]any{"key": "value"},
			isSensitive: true,
			expected:    map[string]any{"key": "value"},
		},
		{
			name:        "Sensitive slice should preserve structure",
			value:       []any{"item1", "item2"},
			isSensitive: true,
			expected:    []any{"item1", "item2"},
		},
		{
			name:        "Nil value should remain nil",
			value:       nil,
			isSensitive: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.maskSensitiveValue(tc.value, tc.isSensitive)
			assert.Equal(t, tc.expected, result, "Masked value should match expected")
		})
	}
}

// TestCompareObjectsWithSensitiveMasking tests the integration of sensitive masking with property comparison
func TestCompareObjectsWithSensitiveMasking(t *testing.T) {
	analyzer := &Analyzer{}

	// Test case: property change with sensitive values
	analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

	// Create test data with sensitive values
	before := map[string]any{
		"password": "old_secret",
		"username": "normal_user",
	}
	after := map[string]any{
		"password": "new_secret",
		"username": "normal_user",
	}

	// Sensitive flags indicating "password" is sensitive
	beforeSensitive := map[string]any{
		"password": true,
		"username": false,
	}
	afterSensitive := map[string]any{
		"password": true,
		"username": false,
	}

	// Call compareObjects with sensitive data
	analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, []string{}, analysis)

	// Verify that sensitive values are masked while non-sensitive values are preserved
	passwordFound := false
	usernameFound := false

	for _, change := range analysis.Changes {
		switch change.Name {
		case "password":
			passwordFound = true
			assert.True(t, change.Sensitive, "Password property should be marked as sensitive")
			assert.Equal(t, "(sensitive value)", change.Before, "Sensitive before value should be masked")
			assert.Equal(t, "(sensitive value)", change.After, "Sensitive after value should be masked")
		case "username":
			usernameFound = true
			assert.False(t, change.Sensitive, "Username property should not be marked as sensitive")
			assert.Equal(t, "normal_user", change.Before, "Non-sensitive before value should not be masked")
			assert.Equal(t, "normal_user", change.After, "Non-sensitive after value should not be masked")
		}
	}

	// Only password should change since username values are identical
	assert.True(t, passwordFound, "Should find password change")
	assert.False(t, usernameFound, "Should not find username change since values are identical")
}

// TestCompareObjectsWithNestedSensitiveValues tests sensitive masking in nested structures
func TestCompareObjectsWithNestedSensitiveValues(t *testing.T) {
	analyzer := &Analyzer{}
	analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

	// Test nested structure with some sensitive leaf values
	before := map[string]any{
		"config": map[string]any{
			"api_key":  "old_key",
			"endpoint": "https://api.example.com",
			"settings": map[string]any{
				"timeout": 30,
				"secret":  "old_secret",
			},
		},
	}
	after := map[string]any{
		"config": map[string]any{
			"api_key":  "new_key",
			"endpoint": "https://api.example.com",
			"settings": map[string]any{
				"timeout": 60,
				"secret":  "new_secret",
			},
		},
	}

	// Sensitive flags - api_key and secret are sensitive
	beforeSensitive := map[string]any{
		"config": map[string]any{
			"api_key":  true,
			"endpoint": false,
			"settings": map[string]any{
				"timeout": false,
				"secret":  true,
			},
		},
	}
	afterSensitive := map[string]any{
		"config": map[string]any{
			"api_key":  true,
			"endpoint": false,
			"settings": map[string]any{
				"timeout": false,
				"secret":  true,
			},
		},
	}

	analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, []string{}, analysis)

	// Check that sensitive leaf values are masked while structure is preserved
	changesByName := make(map[string]PropertyChange)
	for _, change := range analysis.Changes {
		changesByName[change.Name] = change
	}

	// api_key should be masked
	if apiKeyChange, exists := changesByName["api_key"]; exists {
		assert.True(t, apiKeyChange.Sensitive, "api_key should be marked as sensitive")
		assert.Equal(t, "(sensitive value)", apiKeyChange.Before, "Sensitive api_key before value should be masked")
		assert.Equal(t, "(sensitive value)", apiKeyChange.After, "Sensitive api_key after value should be masked")
	}

	// timeout should not be masked
	if timeoutChange, exists := changesByName["timeout"]; exists {
		assert.False(t, timeoutChange.Sensitive, "timeout should not be marked as sensitive")
		assert.Equal(t, 30, timeoutChange.Before, "Non-sensitive timeout before value should not be masked")
		assert.Equal(t, 60, timeoutChange.After, "Non-sensitive timeout after value should not be masked")
	}

	// secret should be masked
	if secretChange, exists := changesByName["secret"]; exists {
		assert.True(t, secretChange.Sensitive, "secret should be marked as sensitive")
		assert.Equal(t, "(sensitive value)", secretChange.Before, "Sensitive secret before value should be masked")
		assert.Equal(t, "(sensitive value)", secretChange.After, "Sensitive secret after value should be masked")
	}
}

// TestExtractPropertyName tests the property name extraction
func TestExtractPropertyName(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple property",
			path:     "name",
			expected: "name",
		},
		{
			name:     "nested property",
			path:     "tags.environment",
			expected: "environment",
		},
		{
			name:     "array property",
			path:     "items[0]",
			expected: "items",
		},
		{
			name:     "nested array property",
			path:     "tags[0].name",
			expected: "name",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "",
		},
		{
			name:     "deep nested property",
			path:     "config.database.settings.timeout",
			expected: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.extractPropertyName(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParsePath tests the path parsing functionality
func TestParsePath(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple property",
			path:     "name",
			expected: []string{"name"},
		},
		{
			name:     "nested property",
			path:     "tags.environment",
			expected: []string{"tags", "environment"},
		},
		{
			name:     "array property",
			path:     "items[0]",
			expected: []string{"items", "0"},
		},
		{
			name:     "nested array property",
			path:     "tags[0].name",
			expected: []string{"tags", "0", "name"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "multiple array indices",
			path:     "matrix[1][2]",
			expected: []string{"matrix", "1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.parsePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsSensitive tests the sensitive value detection
func TestIsSensitive(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name            string
		path            string
		sensitiveValues any
		expected        bool
	}{
		{
			name:            "simple sensitive property",
			path:            "password",
			sensitiveValues: map[string]any{"password": true},
			expected:        true,
		},
		{
			name:            "simple non-sensitive property",
			path:            "name",
			sensitiveValues: map[string]any{"password": true},
			expected:        false,
		},
		{
			name:            "nested sensitive property",
			path:            "config.password",
			sensitiveValues: map[string]any{"config": map[string]any{"password": true}},
			expected:        true,
		},
		{
			name:            "array sensitive property",
			path:            "secrets[0]",
			sensitiveValues: map[string]any{"secrets": []any{true, false}},
			expected:        true,
		},
		{
			name:            "array non-sensitive property",
			path:            "secrets[1]",
			sensitiveValues: map[string]any{"secrets": []any{true, false}},
			expected:        false,
		},
		{
			name:            "nil sensitive values",
			path:            "password",
			sensitiveValues: nil,
			expected:        false,
		},
		{
			name:            "path not found",
			path:            "nonexistent",
			sensitiveValues: map[string]any{"password": true},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isSensitive(tt.path, tt.sensitiveValues)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsOutputSensitive tests output sensitivity detection
func TestIsOutputSensitive(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		beforeSensitive   any
		afterSensitive    any
		expectedSensitive bool
	}{
		{
			name:              "non-sensitive output",
			beforeSensitive:   false,
			afterSensitive:    false,
			expectedSensitive: false,
		},
		{
			name:              "sensitive before value",
			beforeSensitive:   true,
			afterSensitive:    false,
			expectedSensitive: true,
		},
		{
			name:              "sensitive after value",
			beforeSensitive:   false,
			afterSensitive:    true,
			expectedSensitive: true,
		},
		{
			name:              "both sensitive",
			beforeSensitive:   true,
			afterSensitive:    true,
			expectedSensitive: true,
		},
		{
			name:              "nil values",
			beforeSensitive:   nil,
			afterSensitive:    nil,
			expectedSensitive: false,
		},
		{
			name:              "non-boolean sensitive values",
			beforeSensitive:   "not_boolean",
			afterSensitive:    123,
			expectedSensitive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := &tfjson.Change{
				BeforeSensitive: tt.beforeSensitive,
				AfterSensitive:  tt.afterSensitive,
			}
			result := analyzer.isOutputSensitive(change)
			if result != tt.expectedSensitive {
				t.Errorf("isOutputSensitive() = %v, expected %v", result, tt.expectedSensitive)
			}
		})
	}
}

// TestExtractSensitiveChild tests sensitive child extraction
func TestExtractSensitiveChild(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name            string
		sensitiveValues any
		key             string
		expected        any
	}{
		{
			name:            "extract child from map",
			sensitiveValues: map[string]any{"password": true, "name": false},
			key:             "password",
			expected:        true,
		},
		{
			name:            "extract missing child from map",
			sensitiveValues: map[string]any{"password": true},
			key:             "name",
			expected:        nil,
		},
		{
			name:            "extract from nil",
			sensitiveValues: nil,
			key:             "password",
			expected:        nil,
		},
		{
			name:            "extract from non-map",
			sensitiveValues: true,
			key:             "password",
			expected:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.extractSensitiveChild(tt.sensitiveValues, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractSensitiveIndex tests sensitive array index extraction
func TestExtractSensitiveIndex(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name            string
		sensitiveValues any
		index           int
		expected        any
	}{
		{
			name:            "extract valid index from array",
			sensitiveValues: []any{true, false, true},
			index:           1,
			expected:        false,
		},
		{
			name:            "extract out of bounds index from array",
			sensitiveValues: []any{true, false},
			index:           5,
			expected:        nil,
		},
		{
			name:            "extract negative index from array",
			sensitiveValues: []any{true, false},
			index:           -1,
			expected:        nil,
		},
		{
			name:            "extract from nil",
			sensitiveValues: nil,
			index:           0,
			expected:        nil,
		},
		{
			name:            "extract from non-array",
			sensitiveValues: map[string]any{"test": true},
			index:           0,
			expected:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.extractSensitiveIndex(tt.sensitiveValues, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAnalyzePropertyChangesWithNewCompareObjects tests the updated analyzePropertyChanges method
func TestAnalyzePropertyChangesWithNewCompareObjects(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		change            *tfjson.ResourceChange
		expectedChanges   int
		expectedTruncated bool
	}{
		{
			name: "simple property changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"name": "old-name",
						"size": 10,
					},
					After: map[string]any{
						"name": "new-name",
						"size": 20,
					},
				},
			},
			expectedChanges:   2,
			expectedTruncated: false,
		},
		{
			name: "no changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{"name": "same"},
					After:  map[string]any{"name": "same"},
				},
			},
			expectedChanges:   0,
			expectedTruncated: false,
		},
		{
			name: "nested property changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"tags":     map[string]any{"env": "dev", "team": "backend"},
						"settings": map[string]any{"timeout": 30},
					},
					After: map[string]any{
						"tags":     map[string]any{"env": "prod", "team": "backend"},
						"settings": map[string]any{"timeout": 60},
					},
				},
			},
			expectedChanges:   2, // env and timeout changed
			expectedTruncated: false,
		},
		{
			name: "sensitive property changes",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"password": "old-secret",
						"name":     "resource",
					},
					After: map[string]any{
						"password": "new-secret",
						"name":     "resource",
					},
					BeforeSensitive: map[string]any{
						"password": true,
						"name":     false,
					},
					AfterSensitive: map[string]any{
						"password": true,
						"name":     false,
					},
				},
			},
			expectedChanges:   1, // only password changed
			expectedTruncated: false,
		},
		{
			name: "nil change",
			change: &tfjson.ResourceChange{
				Change: nil,
			},
			expectedChanges:   0,
			expectedTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzePropertyChanges(tt.change)

			assert.Equal(t, tt.expectedChanges, result.Count, "Expected count should match")
			assert.Equal(t, tt.expectedChanges, len(result.Changes), "Expected changes length should match")
			assert.Equal(t, tt.expectedTruncated, result.Truncated, "Expected truncation status should match")

			// Verify that each change has the required Action field
			for i, change := range result.Changes {
				assert.NotEmpty(t, change.Action, "Change %d should have an action", i)
				assert.Contains(t, []string{"add", "remove", "update"}, change.Action, "Change %d should have valid action", i)
				assert.NotEmpty(t, change.Name, "Change %d should have a name", i)
			}
		})
	}
}

// TestCompareObjectsEnhanced tests the deep object comparison algorithm with specific scenarios
func TestCompareObjectsEnhanced(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name     string
		before   any
		after    any
		expected []PropertyChange
	}{
		{
			name:   "simple string change",
			before: map[string]any{"name": "old"},
			after:  map[string]any{"name": "new"},
			expected: []PropertyChange{{
				Name:   "name",
				Path:   []string{"name"},
				Action: "update",
				Before: "old",
				After:  "new",
			}},
		},
		{
			name: "nested object change",
			before: map[string]any{
				"tags": map[string]any{"env": "dev"},
			},
			after: map[string]any{
				"tags": map[string]any{"env": "prod"},
			},
			expected: []PropertyChange{{
				Name:   "tags",
				Path:   []string{"tags"},
				Action: "update",
				Before: map[string]any{"env": "dev"},
				After:  map[string]any{"env": "prod"},
			}},
		},
		{
			name:   "array length change",
			before: map[string]any{"items": []any{1, 2}},
			after:  map[string]any{"items": []any{1, 2, 3}},
			expected: []PropertyChange{{
				Name:   "items",
				Path:   []string{"items"},
				Action: "update",
				Before: []any{1, 2},
				After:  []any{1, 2, 3},
			}},
		},
		{
			name:   "property removal",
			before: map[string]any{"a": 1, "b": 2},
			after:  map[string]any{"a": 1},
			expected: []PropertyChange{{
				Name:   "b",
				Path:   []string{"b"},
				Action: "remove",
				Before: 2,
				After:  nil,
			}},
		},
		{
			name:   "property addition",
			before: map[string]any{"a": 1},
			after:  map[string]any{"a": 1, "b": 2},
			expected: []PropertyChange{{
				Name:   "b",
				Path:   []string{"b"},
				Action: "add",
				Before: nil,
				After:  2,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := PropertyChangeAnalysis{
				Changes: []PropertyChange{},
			}

			analyzer.compareObjects("", tt.before, tt.after, nil, nil, nil, nil, &analysis)

			assert.Equal(t, len(tt.expected), len(analysis.Changes), "Expected number of changes should match")

			for i, expectedChange := range tt.expected {
				if i < len(analysis.Changes) {
					actual := analysis.Changes[i]
					assert.Equal(t, expectedChange.Name, actual.Name, "Change %d name should match", i)
					assert.Equal(t, expectedChange.Path, actual.Path, "Change %d path should match", i)
					assert.Equal(t, expectedChange.Action, actual.Action, "Change %d action should match", i)
					assert.Equal(t, expectedChange.Before, actual.Before, "Change %d before value should match", i)
					assert.Equal(t, expectedChange.After, actual.After, "Change %d after value should match", i)
				}
			}
		})
	}
}

// TestEnforcePropertyLimits tests the performance limit enforcement
func TestEnforcePropertyLimits(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		initialChanges    []PropertyChange
		expectedCount     int
		expectedTruncated bool
		expectedTotalSize int
		testType          string
	}{
		{
			name: "under limits should not truncate",
			initialChanges: []PropertyChange{
				{Name: "prop1", Before: "small", After: "value", Action: "update"},
				{Name: "prop2", Before: "another", After: "small", Action: "update"},
			},
			expectedCount:     2,
			expectedTruncated: false,
			testType:          "normal",
		},
		{
			name: "property count limit should truncate",
			initialChanges: func() []PropertyChange {
				changes := make([]PropertyChange, MaxPropertiesPerResource+5)
				for i := range changes {
					changes[i] = PropertyChange{
						Name:   fmt.Sprintf("prop%d", i),
						Before: "value",
						After:  "new",
						Action: "update",
					}
				}
				return changes
			}(),
			expectedCount:     MaxPropertiesPerResource,
			expectedTruncated: true,
			testType:          "count_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := PropertyChangeAnalysis{
				Changes: tt.initialChanges,
			}

			analyzer.enforcePropertyLimits(&analysis)

			assert.Equal(t, tt.expectedCount, analysis.Count, "Count should match expected")
			assert.Equal(t, tt.expectedCount, len(analysis.Changes), "Changes length should match count")
			assert.Equal(t, tt.expectedTruncated, analysis.Truncated, "Truncated status should match expected")

			// Verify all remaining changes have Size set
			for i, change := range analysis.Changes {
				assert.GreaterOrEqual(t, change.Size, 0, "Change %d should have non-negative size", i)
			}
		})
	}
}

// TestIsValueUnknown tests the unknown value detection function
func TestIsValueUnknown(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name         string
		afterUnknown any
		path         string
		expected     bool
	}{
		{
			name:         "nil afterUnknown should return false",
			afterUnknown: nil,
			path:         "property",
			expected:     false,
		},
		{
			name:         "simple true unknown value should return true",
			afterUnknown: map[string]any{"id": true},
			path:         "id",
			expected:     true,
		},
		{
			name:         "simple false unknown value should return false",
			afterUnknown: map[string]any{"id": false},
			path:         "id",
			expected:     false,
		},
		{
			name:         "missing property should return false",
			afterUnknown: map[string]any{"id": true},
			path:         "name",
			expected:     false,
		},
		{
			name:         "nested object unknown property should return true",
			afterUnknown: map[string]any{"config": map[string]any{"timeout": true}},
			path:         "config.timeout",
			expected:     true,
		},
		{
			name:         "nested object known property should return false",
			afterUnknown: map[string]any{"config": map[string]any{"timeout": false}},
			path:         "config.timeout",
			expected:     false,
		},
		{
			name:         "array unknown element should return true",
			afterUnknown: map[string]any{"vpc_security_group_ids": []any{true, false}},
			path:         "vpc_security_group_ids[0]",
			expected:     true,
		},
		{
			name:         "array known element should return false",
			afterUnknown: map[string]any{"vpc_security_group_ids": []any{true, false}},
			path:         "vpc_security_group_ids[1]",
			expected:     false,
		},
		{
			name:         "out of bounds array index should return false",
			afterUnknown: map[string]any{"vpc_security_group_ids": []any{true}},
			path:         "vpc_security_group_ids[5]",
			expected:     false,
		},
		{
			name: "complex nested structure with unknown values",
			afterUnknown: map[string]any{
				"network_interface": []any{
					map[string]any{
						"subnet_id":    true,
						"device_index": false,
					},
				},
			},
			path:     "network_interface[0].subnet_id",
			expected: true,
		},
		{
			name: "complex nested structure with known values",
			afterUnknown: map[string]any{
				"network_interface": []any{
					map[string]any{
						"subnet_id":    true,
						"device_index": false,
					},
				},
			},
			path:     "network_interface[0].device_index",
			expected: false,
		},
		{
			name:         "invalid array index should return false",
			afterUnknown: map[string]any{"items": []any{true}},
			path:         "items[invalid]",
			expected:     false,
		},
		{
			name:         "negative array index should return false",
			afterUnknown: map[string]any{"items": []any{true}},
			path:         "items[-1]",
			expected:     false,
		},
		{
			name:         "boolean at intermediate level should return that boolean",
			afterUnknown: map[string]any{"config": true},
			path:         "config.anything",
			expected:     true,
		},
		{
			name:         "non-boolean non-map non-array value should return false",
			afterUnknown: map[string]any{"config": "invalid"},
			path:         "config.anything",
			expected:     false,
		},
		{
			name:         "empty path should handle root level",
			afterUnknown: true,
			path:         "",
			expected:     true,
		},
		{
			name: "deeply nested path with multiple unknown levels",
			afterUnknown: map[string]any{
				"level1": map[string]any{
					"level2": []any{
						map[string]any{
							"level3": map[string]any{
								"final": true,
							},
						},
					},
				},
			},
			path:     "level1.level2[0].level3.final",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.isValueUnknown(tc.afterUnknown, tc.path)
			assert.Equal(t, tc.expected, result, "isValueUnknown result should match expected")
		})
	}
}

// TestGetUnknownValueDisplay tests the unknown value display function
func TestGetUnknownValueDisplay(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "should return exact Terraform syntax",
			expected: "(known after apply)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.getUnknownValueDisplay()
			assert.Equal(t, tc.expected, result, "getUnknownValueDisplay should return exact Terraform syntax")
		})
	}

	// Additional test to ensure the string is exactly as required by requirement 1.3
	t.Run("exact string match requirement", func(t *testing.T) {
		result := analyzer.getUnknownValueDisplay()
		assert.Exactly(t, "(known after apply)", result, "Must return exact string '(known after apply)'")
		assert.NotEqual(t, "(known_after_apply)", result, "Should not have underscores")
		assert.NotEqual(t, "known after apply", result, "Should have parentheses")
		assert.NotEqual(t, "(Known After Apply)", result, "Should not be title case")
	})
}

// TestCompareObjectsWithUnknownValues tests enhanced compareObjects function with unknown values integration
func TestCompareObjectsWithUnknownValues(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name            string
		before          any
		after           any
		beforeSensitive any
		afterSensitive  any
		afterUnknown    any
		expectedChanges int
		expectedUnknown []bool
		expectedActions []string
		expectedNames   []string
		description     string
	}{
		{
			name: "simple property with unknown value should override after value",
			before: map[string]any{
				"id": "old-id",
			},
			after: map[string]any{
				"id": nil, // Would normally show as removal
			},
			afterUnknown: map[string]any{
				"id": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"update"},
			expectedNames:   []string{"id"},
			description:     "Unknown values should override standard before/after comparison logic (requirement 1.6)",
		},
		{
			name:   "newly created property with unknown value",
			before: map[string]any{},
			after: map[string]any{
				"id": nil,
			},
			afterUnknown: map[string]any{
				"id": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"add"},
			expectedNames:   []string{"id"},
			description:     "New properties with unknown values should show as additions (requirement 1.5)",
		},
		{
			name: "known to unknown transition",
			before: map[string]any{
				"ami": "ami-123",
			},
			after: map[string]any{
				"ami": nil,
			},
			afterUnknown: map[string]any{
				"ami": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"update"},
			expectedNames:   []string{"ami"},
			description:     "Known to unknown transitions should show before_value → (known after apply) (requirement 1.4)",
		},
		{
			name: "mixed known and unknown properties",
			before: map[string]any{
				"instance_type": "t2.micro",
				"id":            "old-id",
				"ami":           "ami-123",
			},
			after: map[string]any{
				"instance_type": "t2.small",
				"id":            nil,
				"ami":           "ami-456",
			},
			afterUnknown: map[string]any{
				"id": true, // Only id is unknown
			},
			expectedChanges: 3,
			expectedUnknown: []bool{false, true, false}, // Order may vary, so we'll check different in assertion
			expectedActions: []string{"update", "update", "update"},
			expectedNames:   []string{"instance_type", "id", "ami"},
			description:     "Mix of known and unknown properties should be handled correctly",
		},
		{
			name: "sensitive property with unknown value",
			before: map[string]any{
				"user_data": "old-script",
			},
			after: map[string]any{
				"user_data": nil,
			},
			beforeSensitive: map[string]any{
				"user_data": true,
			},
			afterSensitive: map[string]any{
				"user_data": true,
			},
			afterUnknown: map[string]any{
				"user_data": true,
			},
			expectedChanges: 1,
			expectedUnknown: []bool{true},
			expectedActions: []string{"update"},
			expectedNames:   []string{"user_data"},
			description:     "Unknown values should integrate with sensitive property detection (requirement 3.1)",
		},
		{
			name: "no unknown values should work normally",
			before: map[string]any{
				"instance_type": "t2.micro",
				"ami":           "ami-123",
			},
			after: map[string]any{
				"instance_type": "t2.small",
				"ami":           "ami-123",
			},
			afterUnknown:    map[string]any{}, // Empty unknown values
			expectedChanges: 1,
			expectedUnknown: []bool{false},
			expectedActions: []string{"update"},
			expectedNames:   []string{"instance_type"},
			description:     "Normal operation should not be affected when no unknown values present",
		},
		{
			name:            "nil after_unknown should work normally",
			before:          map[string]any{"id": "old-id"},
			after:           map[string]any{"id": "new-id"},
			afterUnknown:    nil,
			expectedChanges: 1,
			expectedUnknown: []bool{false},
			expectedActions: []string{"update"},
			expectedNames:   []string{"id"},
			description:     "Nil after_unknown should not break normal processing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := PropertyChangeAnalysis{
				Changes: []PropertyChange{},
			}

			analyzer.compareObjects("", tc.before, tc.after, tc.beforeSensitive, tc.afterSensitive, tc.afterUnknown, nil, &analysis)

			// Verify number of changes
			assert.Equal(t, tc.expectedChanges, len(analysis.Changes), tc.description)

			if tc.expectedChanges > 0 {
				// Count unknown changes
				unknownCount := 0
				for _, change := range analysis.Changes {
					if change.IsUnknown {
						unknownCount++
						// Verify unknown changes have correct display value
						assert.Equal(t, "(known after apply)", change.After, "Unknown values should display '(known after apply)' (%s)", tc.description)
						assert.Equal(t, "after", change.UnknownType, "Unknown values should have UnknownType 'after' (%s)", tc.description)
					}
				}

				// Check expected unknown count
				expectedUnknownCount := 0
				for _, expected := range tc.expectedUnknown {
					if expected {
						expectedUnknownCount++
					}
				}
				assert.Equal(t, expectedUnknownCount, unknownCount, "Number of unknown changes should match expected (%s)", tc.description)

				// Verify actions match expected (order-independent)
				if len(tc.expectedActions) > 0 {
					actualActions := make([]string, len(analysis.Changes))
					for i, change := range analysis.Changes {
						actualActions[i] = change.Action
					}
					assert.ElementsMatch(t, tc.expectedActions, actualActions, "Actions should match expected (%s)", tc.description)
				}

				// Verify names match expected (order-independent)
				if len(tc.expectedNames) > 0 {
					actualNames := make([]string, len(analysis.Changes))
					for i, change := range analysis.Changes {
						actualNames[i] = change.Name
					}
					assert.ElementsMatch(t, tc.expectedNames, actualNames, "Names should match expected (%s)", tc.description)
				}

				// Verify that unknown values don't appear as deletions (requirement 1.2)
				for i, change := range analysis.Changes {
					if change.IsUnknown && change.Before != nil {
						assert.NotEqual(t, "remove", change.Action, "Change %d: Unknown values should not appear as deletions (%s)", i, tc.description)
					}
				}

				// Verify sensitive property integration (requirement 3.1)
				for i, change := range analysis.Changes {
					if tc.beforeSensitive != nil || tc.afterSensitive != nil {
						// If we have sensitive values data, verify integration works
						assert.NotNil(t, change.Sensitive, "Change %d: Sensitive field should be set when sensitive data is present (%s)", i, tc.description)
					}
				}
			}
		})
	}
}

// TestAnalyzePropertyChangesWithUnknownValuesIntegration tests complete property analysis with unknown values
func TestAnalyzePropertyChangesWithUnknownValuesIntegration(t *testing.T) {
	analyzer := &Analyzer{}

	testCases := []struct {
		name                 string
		change               *tfjson.ResourceChange
		expectedUnknownCount int
		expectedTotalCount   int
		expectedUnknownProps []string
		description          string
	}{
		{
			name: "resource change with unknown values should populate unknown fields",
			change: &tfjson.ResourceChange{
				Type: "aws_instance",
				Change: &tfjson.Change{
					Before: map[string]any{
						"id":            nil,
						"ami":           "ami-123",
						"instance_type": "t2.micro",
					},
					After: map[string]any{
						"id":            nil,
						"ami":           "ami-123",
						"instance_type": "t2.small",
					},
					AfterUnknown: map[string]any{
						"id": true,
					},
				},
			},
			expectedUnknownCount: 1,
			expectedTotalCount:   2, // id and instance_type changes
			expectedUnknownProps: []string{"id"},
			description:          "Resource with unknown values should properly populate unknown tracking fields",
		},
		{
			name: "resource with multiple unknown properties",
			change: &tfjson.ResourceChange{
				Type: "aws_instance",
				Change: &tfjson.Change{
					Before: map[string]any{
						"id":                     nil,
						"vpc_security_group_ids": nil,
						"private_ip":             nil,
						"instance_type":          "t2.micro",
					},
					After: map[string]any{
						"id":                     nil,
						"vpc_security_group_ids": nil,
						"private_ip":             nil,
						"instance_type":          "t2.micro",
					},
					AfterUnknown: map[string]any{
						"id":                     true,
						"vpc_security_group_ids": true,
						"private_ip":             true,
					},
				},
			},
			expectedUnknownCount: 3,
			expectedTotalCount:   3,
			expectedUnknownProps: []string{"id", "vpc_security_group_ids", "private_ip"},
			description:          "Multiple unknown properties should all be tracked correctly",
		},
		{
			name: "resource with no unknown values should work normally",
			change: &tfjson.ResourceChange{
				Type: "aws_instance",
				Change: &tfjson.Change{
					Before: map[string]any{
						"instance_type": "t2.micro",
						"ami":           "ami-123",
					},
					After: map[string]any{
						"instance_type": "t2.small",
						"ami":           "ami-456",
					},
					AfterUnknown: map[string]any{}, // No unknown values
				},
			},
			expectedUnknownCount: 0,
			expectedTotalCount:   2,
			expectedUnknownProps: []string{},
			description:          "Resources without unknown values should work normally",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.analyzePropertyChanges(tc.change)

			// Count unknown properties
			unknownCount := 0
			var unknownProps []string
			for _, change := range result.Changes {
				if change.IsUnknown {
					unknownCount++
					unknownProps = append(unknownProps, change.Name)
				}
			}

			assert.Equal(t, tc.expectedUnknownCount, unknownCount, tc.description+" - unknown count")
			assert.Equal(t, tc.expectedTotalCount, result.Count, tc.description+" - total count")
			assert.ElementsMatch(t, tc.expectedUnknownProps, unknownProps, tc.description+" - unknown property names")

			// Verify all unknown changes have the correct display value
			for i, change := range result.Changes {
				if change.IsUnknown {
					assert.Equal(t, "(known after apply)", change.After, "Change %d should display '(known after apply)' for unknown values (%s)", i, tc.description)
					assert.Equal(t, "after", change.UnknownType, "Change %d should have UnknownType 'after' (%s)", i, tc.description)
				}
			}

			// Verify that all changes have required fields populated
			for i, change := range result.Changes {
				assert.NotEmpty(t, change.Action, "Change %d should have action (%s)", i, tc.description)
				assert.NotEmpty(t, change.Name, "Change %d should have name (%s)", i, tc.description)
				assert.NotNil(t, change.Path, "Change %d should have path (%s)", i, tc.description)
			}
		})
	}
}

// TestAnalyzePropertyChangesWithLimits tests the complete property analysis with performance limits
func TestAnalyzePropertyChangesWithLimits(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
		name              string
		change            *tfjson.ResourceChange
		expectedTruncated bool
		description       string
	}{
		{
			name: "normal change should not truncate",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: map[string]any{
						"name":    "old-name",
						"size":    10,
						"enabled": false,
					},
					After: map[string]any{
						"name":    "new-name",
						"size":    20,
						"enabled": true,
					},
				},
			},
			expectedTruncated: false,
			description:       "Small changes should not trigger truncation",
		},
		{
			name: "many properties should apply count limits",
			change: &tfjson.ResourceChange{
				Change: &tfjson.Change{
					Before: func() map[string]any {
						result := make(map[string]any)
						// Create more properties than the limit
						for i := range MaxPropertiesPerResource + 10 {
							result[fmt.Sprintf("prop_%d", i)] = fmt.Sprintf("value_%d", i)
						}
						return result
					}(),
					After: func() map[string]any {
						result := make(map[string]any)
						// Create more properties than the limit
						for i := range MaxPropertiesPerResource + 10 {
							result[fmt.Sprintf("prop_%d", i)] = fmt.Sprintf("new_value_%d", i)
						}
						return result
					}(),
				},
			},
			expectedTruncated: true,
			description:       "Many properties should trigger count limit truncation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzePropertyChanges(tt.change)

			assert.Equal(t, tt.expectedTruncated, result.Truncated, tt.description)
			assert.Equal(t, len(result.Changes), result.Count, "Count should match changes length")

			// Verify all changes have required fields
			for i, change := range result.Changes {
				assert.NotEmpty(t, change.Action, "Change %d should have action", i)
				assert.NotEmpty(t, change.Name, "Change %d should have name", i)
				assert.NotNil(t, change.Path, "Change %d should have path", i)
				assert.GreaterOrEqual(t, change.Size, 0, "Change %d should have non-negative size", i)
			}

			if tt.expectedTruncated {
				assert.LessOrEqual(t, result.TotalSize, MaxTotalPropertyMemory, "Total size should not exceed limit")
			}
		})
	}
}

func TestCrossFormatConsistencyForUnknownValuesAndOutputs(t *testing.T) {
	cfg := &config.Config{}
	analyzer := &Analyzer{config: cfg}

	// Create a comprehensive plan with various unknown values and outputs
	plan := &tfjson.Plan{
		FormatVersion:    "1.0",
		TerraformVersion: "1.5.0",
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_instance.test",
				Type:    "aws_instance",
				Name:    "test",
				Change: &tfjson.Change{
					Actions: []tfjson.Action{tfjson.ActionCreate},
					Before:  nil,
					After: map[string]any{
						"instance_type": "t3.micro",
						"ami":           "ami-12345678",
						"id":            nil,
						"public_ip":     nil,
					},
					AfterUnknown: map[string]any{
						"id":        true,
						"public_ip": true,
					},
				},
			},
		},
		OutputChanges: map[string]*tfjson.Change{
			"instance_ip": {
				Actions:      []tfjson.Action{tfjson.ActionCreate},
				Before:       nil,
				After:        nil,
				AfterUnknown: true,
			},
			"secret_value": {
				Actions:        []tfjson.Action{tfjson.ActionCreate},
				Before:         nil,
				After:          "top-secret",
				AfterSensitive: true,
			},
			"public_endpoint": {
				Actions: []tfjson.Action{tfjson.ActionCreate},
				Before:  nil,
				After:   "https://api.example.com",
			},
		},
	}

	analyzer.plan = plan
	summary := analyzer.GenerateSummary("")

	testCases := []struct {
		name            string
		validateContent func(t *testing.T, content string, format string)
		description     string
	}{
		{
			name: "unknown values display consistency across formats",
			validateContent: func(t *testing.T, content string, format string) {
				// Verify "(known after apply)" appears consistently (requirement 1.3)
				assert.Contains(t, content, "(known after apply)", format+" format should contain (known after apply)")

				// Count occurrences - should appear for both resource properties and outputs
				unknownCount := strings.Count(content, "(known after apply)")
				assert.GreaterOrEqual(t, unknownCount, 2, format+" format should have multiple (known after apply) instances")
			},
			description: "Unknown values should display (known after apply) consistently across all formats",
		},
		{
			name: "sensitive values display consistency across formats",
			validateContent: func(t *testing.T, content string, format string) {
				// Verify "(sensitive value)" appears consistently (requirement 2.4)
				assert.Contains(t, content, "(sensitive value)", format+" format should contain (sensitive value)")
			},
			description: "Sensitive values should display (sensitive value) consistently across all formats",
		},
		{
			name: "outputs section format consistency",
			validateContent: func(t *testing.T, content string, format string) {
				// Verify outputs section content appears
				outputNames := []string{"instance_ip", "secret_value", "public_endpoint"}
				for _, name := range outputNames {
					assert.Contains(t, content, name, format+" format should contain output "+name)
				}

				// Verify action indicators appear
				actionIndicators := []string{"+", "~", "-"}
				indicatorFound := false
				for _, indicator := range actionIndicators {
					if strings.Contains(content, indicator) {
						indicatorFound = true
						break
					}
				}
				assert.True(t, indicatorFound, format+" format should contain action indicators")
			},
			description: "Outputs section should appear consistently across all formats",
		},
	}

	// Test formats that are relevant for consistency checking
	formats := []struct {
		name       string
		getContent func() string
	}{
		{
			name: "JSON",
			getContent: func() string {
				// Test JSON serialization consistency
				jsonBytes, err := json.Marshal(summary)
				assert.NoError(t, err, "JSON marshaling should not error")
				return string(jsonBytes)
			},
		},
		{
			name: "Summary Structure",
			getContent: func() string {
				// Test the summary data structure directly
				var content strings.Builder

				// Add resource changes content
				for _, rc := range summary.ResourceChanges {
					content.WriteString(fmt.Sprintf("Resource: %s\n", rc.Address))
					if rc.HasUnknownValues {
						content.WriteString("Has unknown values\n")
						for _, prop := range rc.UnknownProperties {
							content.WriteString(fmt.Sprintf("Unknown property: %s\n", prop))
						}
					}
					for _, change := range rc.PropertyChanges.Changes {
						if change.IsUnknown {
							content.WriteString(fmt.Sprintf("Property %s: (known after apply)\n", change.Name))
						}
					}
				}

				// Add output changes content
				for _, oc := range summary.OutputChanges {
					content.WriteString(fmt.Sprintf("Output: %s %s\n", oc.Name, oc.Indicator))
					if oc.IsUnknown {
						content.WriteString("(known after apply)\n")
					} else if oc.Sensitive {
						content.WriteString("(sensitive value)\n")
					}
				}

				return content.String()
			},
		},
	}

	for _, format := range formats {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_%s", format.name, tc.name), func(t *testing.T) {
				content := format.getContent()
				tc.validateContent(t, content, format.name)
			})
		}
	}

	// Additional cross-format validation
	t.Run("data_consistency_across_processing", func(t *testing.T) {
		// Verify that the same data appears consistently in the summary structure

		// Check resource changes
		assert.Len(t, summary.ResourceChanges, 1, "should have 1 resource change")
		resource := summary.ResourceChanges[0]
		assert.True(t, resource.HasUnknownValues, "resource should have unknown values")
		assert.Contains(t, resource.UnknownProperties, "id", "resource should have unknown id")
		assert.Contains(t, resource.UnknownProperties, "public_ip", "resource should have unknown public_ip")

		// Check output changes
		assert.Len(t, summary.OutputChanges, 3, "should have 3 output changes")

		outputMap := make(map[string]OutputChange)
		for _, oc := range summary.OutputChanges {
			outputMap[oc.Name] = oc
		}

		// Verify unknown output
		if instanceIp, exists := outputMap["instance_ip"]; exists {
			assert.Equal(t, "(known after apply)", instanceIp.After, "unknown output should show (known after apply)")
			assert.True(t, instanceIp.IsUnknown, "unknown output should be marked as unknown")
		}

		// Verify sensitive output
		if secretValue, exists := outputMap["secret_value"]; exists {
			assert.Equal(t, "(sensitive value)", secretValue.After, "sensitive output should show (sensitive value)")
			assert.True(t, secretValue.Sensitive, "sensitive output should be marked as sensitive")
		}

		// Verify normal output
		if publicEndpoint, exists := outputMap["public_endpoint"]; exists {
			assert.Equal(t, "https://api.example.com", publicEndpoint.After, "normal output should show actual value")
			assert.False(t, publicEndpoint.Sensitive, "normal output should not be sensitive")
			assert.False(t, publicEndpoint.IsUnknown, "normal output should not be unknown")
		}
	})
}
