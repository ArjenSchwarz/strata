package plan

import (
	"encoding/json"
	"strings"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// analyzeChangeJSON runs the production property-change analysis for a single
// update and returns the analysis together with its JSON encoding, which is
// what ends up in JSON output and PR comments.
func analyzeChangeJSON(t *testing.T, change *tfjson.Change) (PropertyChangeAnalysis, string) {
	t.Helper()
	analyzer := &Analyzer{}
	analysis := analyzer.AnalyzePropertyChanges(&tfjson.ResourceChange{
		Address: "aws_instance.example",
		Type:    "aws_instance",
		Name:    "example",
		Change:  change,
	})
	encoded, err := json.Marshal(analysis)
	require.NoError(t, err)
	return analysis, string(encoded)
}

func findChange(changes []PropertyChange, path string) *PropertyChange {
	for i := range changes {
		if strings.Join(changes[i].Path, ".") == path {
			return &changes[i]
		}
	}
	return nil
}

// TestSensitiveAggregateMapLeak verifies that common nested maps reported as a
// single aggregate PropertyChange honour per-field sensitivity markers below
// the map path instead of leaking the raw map.
//
// Regression test for T-1394.
func TestSensitiveAggregateMapLeak(t *testing.T) {
	tests := map[string]struct {
		property string
		change   *tfjson.Change
	}{
		"tags update with per-field marker": {
			property: "tags",
			change: &tfjson.Change{
				Actions:         tfjson.Actions{tfjson.ActionUpdate},
				Before:          map[string]any{"tags": map[string]any{"Name": "web", "secret": "old-secret"}},
				After:           map[string]any{"tags": map[string]any{"Name": "web", "secret": "new-secret"}},
				BeforeSensitive: map[string]any{"tags": map[string]any{"secret": true}},
				AfterSensitive:  map[string]any{"tags": map[string]any{"secret": true}},
			},
		},
		"environment update with after-only marker": {
			property: "environment",
			change: &tfjson.Change{
				Actions:        tfjson.Actions{tfjson.ActionUpdate},
				Before:         map[string]any{"environment": map[string]any{"API_KEY": "old-secret"}},
				After:          map[string]any{"environment": map[string]any{"API_KEY": "new-secret"}},
				AfterSensitive: map[string]any{"environment": map[string]any{"API_KEY": true}},
			},
		},
		"db_config removal with per-field marker": {
			property: "db_config",
			change: &tfjson.Change{
				Actions:         tfjson.Actions{tfjson.ActionUpdate},
				Before:          map[string]any{"db_config": map[string]any{"host": "db", "password": "old-secret"}},
				After:           map[string]any{},
				BeforeSensitive: map[string]any{"db_config": map[string]any{"password": true}},
			},
		},
		"metadata addition with nested list marker": {
			property: "metadata",
			change: &tfjson.Change{
				Actions:        tfjson.Actions{tfjson.ActionUpdate},
				Before:         map[string]any{},
				After:          map[string]any{"metadata": map[string]any{"keys": []any{"new-secret"}}},
				AfterSensitive: map[string]any{"metadata": map[string]any{"keys": []any{true}}},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			analysis, encoded := analyzeChangeJSON(t, tc.change)

			got := findChange(analysis.Changes, tc.property)
			require.NotNil(t, got, "expected an aggregate %s change; got %+v", tc.property, analysis.Changes)
			assert.True(t, got.Sensitive, "aggregate with a sensitive child must be marked sensitive")
			if got.Before != nil {
				assert.Equal(t, sensitiveValue, got.Before)
			}
			if got.After != nil {
				assert.Equal(t, sensitiveValue, got.After)
			}
			assert.NotContains(t, encoded, "old-secret")
			assert.NotContains(t, encoded, "new-secret")
		})
	}
}

// TestNonSensitiveAggregateMapShowsValues is a control: aggregate maps without
// any sensitive marker must still expose their values.
func TestNonSensitiveAggregateMapShowsValues(t *testing.T) {
	analysis, _ := analyzeChangeJSON(t, &tfjson.Change{
		Actions:         tfjson.Actions{tfjson.ActionUpdate},
		Before:          map[string]any{"tags": map[string]any{"Name": "old"}},
		After:           map[string]any{"tags": map[string]any{"Name": "new"}},
		BeforeSensitive: map[string]any{"tags": map[string]any{"Name": false}},
		AfterSensitive:  map[string]any{"tags": map[string]any{"Name": false}},
	})

	got := findChange(analysis.Changes, "tags")
	require.NotNil(t, got)
	assert.False(t, got.Sensitive)
	assert.Equal(t, map[string]any{"Name": "old"}, got.Before)
	assert.Equal(t, map[string]any{"Name": "new"}, got.After)
}

// TestUnknownParentSensitiveChildLeak verifies that when a parent object is
// unknown (after_unknown) and Terraform reports sensitivity only on a nested
// field, the per-child unknown changes mask the sensitive Before value while
// keeping non-sensitive siblings and update semantics intact.
//
// Regression test for T-1401.
func TestUnknownParentSensitiveChildLeak(t *testing.T) {
	analysis, encoded := analyzeChangeJSON(t, &tfjson.Change{
		Actions: tfjson.Actions{tfjson.ActionUpdate},
		Before: map[string]any{"settings": map[string]any{
			"password": "old-secret",
			"region":   "us-east-1",
		}},
		After: map[string]any{"settings": map[string]any{
			"password": "new-secret",
			"region":   "us-west-2",
		}},
		BeforeSensitive: map[string]any{"settings": map[string]any{"password": true}},
		AfterSensitive:  map[string]any{"settings": map[string]any{"password": true}},
		AfterUnknown:    map[string]any{"settings": true},
	})

	assert.NotContains(t, encoded, "old-secret")
	assert.NotContains(t, encoded, "new-secret")

	password := findChange(analysis.Changes, "settings.password")
	require.NotNil(t, password, "expected an unknown child change for settings.password; got %+v", analysis.Changes)
	assert.True(t, password.Sensitive)
	assert.True(t, password.IsUnknown)
	assert.Equal(t, sensitiveValue, password.Before)
	assert.Equal(t, actionUpdate, password.Action)

	region := findChange(analysis.Changes, "settings.region")
	require.NotNil(t, region)
	assert.False(t, region.Sensitive)
	assert.Equal(t, "us-east-1", region.Before, "non-sensitive siblings keep their before value")
	assert.Equal(t, actionUpdate, region.Action)
}

// TestUnknownParentSensitiveNestedListLeak covers the slice branch of unknown
// child collection: a sensitive list element below an unknown aggregate map
// must not be emitted in clear text.
//
// Regression test for T-1401.
func TestUnknownParentSensitiveNestedListLeak(t *testing.T) {
	analysis, encoded := analyzeChangeJSON(t, &tfjson.Change{
		Actions:         tfjson.Actions{tfjson.ActionUpdate},
		Before:          map[string]any{"app_config": map[string]any{"keys": []any{"old-secret", "public"}}},
		After:           map[string]any{"app_config": map[string]any{"keys": []any{"new-secret", "public"}}},
		BeforeSensitive: map[string]any{"app_config": map[string]any{"keys": []any{true, false}}},
		AfterSensitive:  map[string]any{"app_config": map[string]any{"keys": []any{true, false}}},
		AfterUnknown:    map[string]any{"app_config": true},
	})

	assert.NotContains(t, encoded, "old-secret")
	assert.NotContains(t, encoded, "new-secret")

	first := findChange(analysis.Changes, "app_config.keys.0")
	require.NotNil(t, first, "expected an unknown child change for app_config.keys[0]; got %+v", analysis.Changes)
	assert.True(t, first.Sensitive)
	assert.Equal(t, sensitiveValue, first.Before)

	second := findChange(analysis.Changes, "app_config.keys.1")
	require.NotNil(t, second)
	assert.False(t, second.Sensitive)
	assert.Equal(t, "public", second.Before)
}
