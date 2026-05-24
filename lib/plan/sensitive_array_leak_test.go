package plan

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSensitiveArrayRemovalLeak verifies that when a slice is removed and
// Terraform reports per-element sensitivity (e.g. before_sensitive:
// {"secrets": [true, true]}), the aggregate removal change is marked sensitive
// and the removed values are masked rather than leaked via PropertyChange.Before.
//
// Regression test for T-1348.
func TestSensitiveArrayRemovalLeak(t *testing.T) {
	analyzer := &Analyzer{}
	analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

	before := map[string]any{"secrets": []any{"old-secret-1", "old-secret-2"}}
	after := map[string]any{}
	beforeSensitive := map[string]any{"secrets": []any{true, true}}

	analyzer.compareObjects("", before, after, beforeSensitive, nil, nil, nil, analysis)

	if assert.Len(t, analysis.Changes, 1, "expected a single aggregate removal change") {
		assert.True(t, analysis.Changes[0].Sensitive, "removal of a per-element sensitive array should be marked sensitive")
		assert.NotContains(t, fmt.Sprint(analysis.Changes[0].Before), "old-secret", "removed secret values must be masked")
	}
}

// TestSensitiveArraySizeChangeLeak verifies that when an array changes size and
// Terraform reports per-element sensitivity, the aggregate size-change record is
// marked sensitive and both before and after values are masked.
//
// Regression test for T-1355.
func TestSensitiveArraySizeChangeLeak(t *testing.T) {
	analyzer := &Analyzer{}
	analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

	before := map[string]any{"secrets": []any{"old-secret-1", "old-secret-2"}}
	after := map[string]any{"secrets": []any{"new-secret-1"}}
	beforeSensitive := map[string]any{"secrets": []any{true, true}}
	afterSensitive := map[string]any{"secrets": []any{true}}

	analyzer.compareObjects("", before, after, beforeSensitive, afterSensitive, nil, nil, analysis)

	if assert.Len(t, analysis.Changes, 1, "expected a single aggregate size-change record") {
		assert.True(t, analysis.Changes[0].Sensitive, "size change of a per-element sensitive array should be marked sensitive")
		assert.NotContains(t, fmt.Sprint(analysis.Changes[0].Before), "old-secret", "before secret values must be masked")
		assert.NotContains(t, fmt.Sprint(analysis.Changes[0].After), "new-secret", "after secret values must be masked")
	}
}

// TestNonSensitiveArrayShowsValues is a control ensuring the fix does not
// regress non-sensitive arrays: real values must still be shown for both the
// removal and size-change aggregate sub-cases.
func TestNonSensitiveArrayShowsValues(t *testing.T) {
	analyzer := &Analyzer{}

	t.Run("non-sensitive removal shows real values", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		before := map[string]any{"items": []any{"value-1", "value-2"}}
		after := map[string]any{}

		analyzer.compareObjects("", before, after, nil, nil, nil, nil, analysis)

		if assert.Len(t, analysis.Changes, 1, "expected a single aggregate removal change") {
			assert.False(t, analysis.Changes[0].Sensitive, "non-sensitive removal should not be marked sensitive")
			assert.Contains(t, fmt.Sprint(analysis.Changes[0].Before), "value-1", "non-sensitive before values should be shown")
		}
	})

	t.Run("non-sensitive size change shows real values", func(t *testing.T) {
		analysis := &PropertyChangeAnalysis{Changes: []PropertyChange{}}

		before := map[string]any{"items": []any{"value-1", "value-2"}}
		after := map[string]any{"items": []any{"value-3"}}

		analyzer.compareObjects("", before, after, nil, nil, nil, nil, analysis)

		if assert.Len(t, analysis.Changes, 1, "expected a single aggregate size-change record") {
			assert.False(t, analysis.Changes[0].Sensitive, "non-sensitive size change should not be marked sensitive")
			assert.Contains(t, fmt.Sprint(analysis.Changes[0].Before), "value-1", "non-sensitive before values should be shown")
			assert.Contains(t, fmt.Sprint(analysis.Changes[0].After), "value-3", "non-sensitive after values should be shown")
		}
	})
}
