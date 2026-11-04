package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateSummarySensitivePropertyDanger covers T-1633 (nested
// sensitive_properties paths) and T-1372 (sensitive properties on whole
// resource creates/deletes) through the public GenerateSummary path.
func TestGenerateSummarySensitivePropertyDanger(t *testing.T) {
	tests := map[string]struct {
		actions    tfjson.Actions
		before     any
		after      any
		properties []config.SensitiveProperty
		wantDanger bool
		wantProps  []string
		wantReason string
	}{
		"T-1633 nested tag value changed": {
			actions:    tfjson.Actions{tfjson.ActionUpdate},
			before:     map[string]any{"instance_type": "t2.micro", "tags": map[string]any{"Name": "web", "Secret": "old"}},
			after:      map[string]any{"instance_type": "t2.micro", "tags": map[string]any{"Name": "web", "Secret": "new"}},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "tags.Secret"}},
			wantDanger: true,
			wantProps:  []string{"tags.Secret"},
			wantReason: "Credential change",
		},
		"T-1633 nested tag added": {
			actions:    tfjson.Actions{tfjson.ActionUpdate},
			before:     map[string]any{"tags": map[string]any{"Name": "web"}},
			after:      map[string]any{"tags": map[string]any{"Name": "web", "Secret": "new"}},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "tags.Secret"}},
			wantDanger: true,
			wantProps:  []string{"tags.Secret"},
		},
		"T-1633 sibling tag changed only is not dangerous": {
			actions:    tfjson.Actions{tfjson.ActionUpdate},
			before:     map[string]any{"tags": map[string]any{"Name": "web", "Secret": "same"}},
			after:      map[string]any{"tags": map[string]any{"Name": "api", "Secret": "same"}},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "tags.Secret"}},
			wantDanger: false,
			wantProps:  []string{},
		},
		"T-1633 list index path": {
			actions: tfjson.Actions{tfjson.ActionUpdate},
			before: map[string]any{"root_block_device": []any{
				map[string]any{"volume_size": float64(8)},
			}},
			after: map[string]any{"root_block_device": []any{
				map[string]any{"volume_size": float64(16)},
			}},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "root_block_device.0.volume_size"}},
			wantDanger: true,
			wantProps:  []string{"root_block_device.0.volume_size"},
		},
		"T-1633 key containing dots": {
			actions:    tfjson.Actions{tfjson.ActionUpdate},
			before:     map[string]any{"tags": map[string]any{"kubernetes.io/role": "a"}},
			after:      map[string]any{"tags": map[string]any{"kubernetes.io/role": "b"}},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "tags.kubernetes.io/role"}},
			wantDanger: true,
			wantProps:  []string{"tags.kubernetes.io/role"},
		},
		"top-level property still detected": {
			actions:    tfjson.Actions{tfjson.ActionUpdate},
			before:     map[string]any{"user_data": "old"},
			after:      map[string]any{"user_data": "new"},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "user_data"}},
			wantDanger: true,
			wantProps:  []string{"user_data"},
		},
		"T-1372 create with sensitive property": {
			actions:    tfjson.Actions{tfjson.ActionCreate},
			before:     nil,
			after:      map[string]any{"user_data": "new-data", "instance_type": "t2.micro"},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "user_data"}},
			wantDanger: true,
			wantProps:  []string{"user_data"},
			wantReason: "User data modification",
		},
		"T-1372 create with nested sensitive property": {
			actions:    tfjson.Actions{tfjson.ActionCreate},
			before:     nil,
			after:      map[string]any{"tags": map[string]any{"Secret": "x"}},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "tags.Secret"}},
			wantDanger: true,
			wantProps:  []string{"tags.Secret"},
		},
		"T-1372 create without sensitive property is not dangerous": {
			actions:    tfjson.Actions{tfjson.ActionCreate},
			before:     nil,
			after:      map[string]any{"instance_type": "t2.micro"},
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "user_data"}},
			wantDanger: false,
			wantProps:  []string{},
		},
		"T-1372 delete reports removed sensitive property": {
			actions:    tfjson.Actions{tfjson.ActionDelete},
			before:     map[string]any{"user_data": "old-data", "instance_type": "t2.micro"},
			after:      nil,
			properties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "user_data"}},
			wantDanger: true,
			wantProps:  []string{"user_data"},
			wantReason: "Resource deletion and User data modification",
		},
		"other resource type is ignored": {
			actions:    tfjson.Actions{tfjson.ActionCreate},
			before:     nil,
			after:      map[string]any{"user_data": "new-data"},
			properties: []config.SensitiveProperty{{ResourceType: "aws_launch_template", Property: "user_data"}},
			wantDanger: false,
			wantProps:  []string{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			plan := &tfjson.Plan{
				FormatVersion:    "1.2",
				TerraformVersion: "1.6.0",
				ResourceChanges: []*tfjson.ResourceChange{{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Name:    "web",
					Change: &tfjson.Change{
						Actions: tc.actions,
						Before:  tc.before,
						After:   tc.after,
					},
				}},
			}
			cfg := &config.Config{
				Plan:                config.PlanConfig{HighlightDangers: true},
				SensitiveProperties: tc.properties,
			}

			summary, err := NewAnalyzer(plan, cfg).GenerateSummary("")
			require.NoError(t, err)
			require.Len(t, summary.ResourceChanges, 1)

			got := summary.ResourceChanges[0]
			assert.Equal(t, tc.wantDanger, got.IsDangerous)
			if len(tc.wantProps) == 0 {
				assert.Empty(t, got.DangerProperties)
			} else {
				assert.Equal(t, tc.wantProps, got.DangerProperties)
			}
			if tc.wantReason != "" {
				assert.Equal(t, tc.wantReason, got.DangerReason)
			}
		})
	}
}

// TestGenerateSummarySensitivePropertyHighlightDisabled confirms the fix does
// not bypass the highlight-dangers switch.
func TestGenerateSummarySensitivePropertyHighlightDisabled(t *testing.T) {
	plan := &tfjson.Plan{
		ResourceChanges: []*tfjson.ResourceChange{{
			Address: "aws_instance.web",
			Type:    "aws_instance",
			Name:    "web",
			Change: &tfjson.Change{
				Actions: tfjson.Actions{tfjson.ActionCreate},
				After:   map[string]any{"tags": map[string]any{"Secret": "x"}},
			},
		}},
	}
	cfg := &config.Config{
		Plan:                config.PlanConfig{HighlightDangers: false},
		SensitiveProperties: []config.SensitiveProperty{{ResourceType: "aws_instance", Property: "tags.Secret"}},
	}

	summary, err := NewAnalyzer(plan, cfg).GenerateSummary("")
	require.NoError(t, err)
	require.Len(t, summary.ResourceChanges, 1)
	assert.False(t, summary.ResourceChanges[0].IsDangerous)
	assert.Empty(t, summary.ResourceChanges[0].DangerProperties)
}
