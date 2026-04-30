package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSensitiveNestedObjectLeak verifies that nested fields of a parent marked
// sensitive as a boolean (whole-object sensitivity) do not leak values.
// Regression test for T-548.
func TestSensitiveNestedObjectLeak(t *testing.T) {
	analyzer := &Analyzer{}

	t.Run("bool sensitive parent masks nested map children", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		before := map[string]any{
			"config": map[string]any{
				"host":     "old-host",
				"password": "old-secret",
			},
		}
		after := map[string]any{
			"config": map[string]any{
				"host":     "new-host",
				"password": "new-secret",
			},
		}
		// Terraform marks entire "config" as sensitive with a bool
		beforeSensitive := map[string]any{"config": true}
		afterSensitive := map[string]any{"config": true}

		analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, nil, analysis)

		// Should produce a single masked change for "config", not leak child values
		for _, change := range analysis.Changes {
			assert.NotContains(t, valStr(change.Before), "old-secret", "before value should not leak sensitive data")
			assert.NotContains(t, valStr(change.After), "new-secret", "after value should not leak sensitive data")
			assert.NotContains(t, valStr(change.Before), "old-host", "before value should not leak sensitive nested field")
			assert.NotContains(t, valStr(change.After), "new-host", "after value should not leak sensitive nested field")
			assert.True(t, change.Sensitive, "change should be marked sensitive")
		}
	})

	t.Run("bool sensitive parent masks nested slice children", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		before := map[string]any{
			"secrets": []any{"secret1", "secret2"},
		}
		after := map[string]any{
			"secrets": []any{"secret3", "secret4"},
		}
		beforeSensitive := map[string]any{"secrets": true}
		afterSensitive := map[string]any{"secrets": true}

		analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, nil, analysis)

		for _, change := range analysis.Changes {
			assert.NotContains(t, valStr(change.Before), "secret1", "before should not leak slice values")
			assert.NotContains(t, valStr(change.After), "secret3", "after should not leak slice values")
			assert.True(t, change.Sensitive, "change should be marked sensitive")
		}
	})

	t.Run("bool sensitive propagates through extractSensitiveIndex", func(t *testing.T) {
		// When a slice's sensitivity is bool true, each element should inherit sensitivity
		result := analyzer.extractSensitiveIndex(true, 0)
		assert.Equal(t, true, result)
		result = analyzer.extractSensitiveIndex(true, 5)
		assert.Equal(t, true, result)
	})

	t.Run("map sensitive per-field still works", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		before := map[string]any{
			"config": map[string]any{
				"host":     "old-host",
				"password": "old-secret",
			},
		}
		after := map[string]any{
			"config": map[string]any{
				"host":     "new-host",
				"password": "new-secret",
			},
		}
		// Per-field sensitivity: only password is sensitive
		beforeSensitive := map[string]any{"config": map[string]any{"password": true}}
		afterSensitive := map[string]any{"config": map[string]any{"password": true}}

		analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, nil, analysis)

		for _, change := range analysis.Changes {
			if change.Name == "password" {
				assert.True(t, change.Sensitive, "password should be sensitive")
				assert.Equal(t, sensitiveValue, change.Before)
				assert.Equal(t, sensitiveValue, change.After)
			}
			if change.Name == "host" {
				assert.False(t, change.Sensitive, "host should not be sensitive")
			}
		}
	})
}

// TestSensitiveAndUnknownNestedObject verifies that when a parent attribute is
// both sensitive=true and after_unknown=true, the unknown child tracking is
// preserved. Regression test for PR #61 review comment.
func TestSensitiveAndUnknownNestedObject(t *testing.T) {
	analyzer := &Analyzer{}

	t.Run("sensitive+unknown parent preserves unknown child tracking", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		before := map[string]any{
			"config": map[string]any{
				"host":     "old-host",
				"password": "old-secret",
			},
		}
		after := map[string]any{
			"config": map[string]any{
				"host":     "new-host",
				"password": "new-secret",
			},
		}
		// Parent is both sensitive and unknown
		beforeSensitive := map[string]any{"config": true}
		afterSensitive := map[string]any{"config": true}
		afterUnknown := map[string]any{"config": true}

		analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, afterUnknown, nil, analysis)

		// Should have unknown child properties collected
		hasUnknownChild := false
		for _, change := range analysis.Changes {
			if change.IsUnknown {
				hasUnknownChild = true
			}
			// Values must still be masked
			assert.NotContains(t, valStr(change.Before), "old-secret", "should not leak sensitive before value")
			assert.NotContains(t, valStr(change.After), "new-secret", "should not leak sensitive after value")
		}
		assert.True(t, hasUnknownChild, "should have at least one unknown child property tracked")
	})

	t.Run("sensitive+unknown map collects nested unknown properties", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		after := map[string]any{
			"db_config": map[string]any{
				"endpoint": "pending",
				"port":     "pending",
			},
		}
		afterSensitive := map[string]any{"db_config": true}
		afterUnknown := map[string]any{"db_config": true}

		analyzer.compareObjects("", nil, after, nil, afterSensitive, afterUnknown, nil, analysis)

		// collectNestedUnknownProperties should have walked the map children
		unknownNames := map[string]bool{}
		for _, change := range analysis.Changes {
			if change.IsUnknown {
				unknownNames[change.Name] = true
			}
		}
		assert.True(t, unknownNames["endpoint"] || unknownNames["port"] || unknownNames["db_config"],
			"should track unknown nested properties; got changes: %v", unknownNames)
	})
}

func valStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
