package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/stretchr/testify/assert"
)

func TestCalculateStatistics(t *testing.T) {
	// Create a test config with sensitive resources
	cfg := &config.Config{
		SensitiveResources: []config.SensitiveResource{
			{ResourceType: "aws_rds_instance"},
			{ResourceType: "aws_dynamodb_table"},
		},
	}

	// Create analyzer with the config
	analyzer := &Analyzer{
		config: cfg,
	}

	testCases := []struct {
		name    string
		changes []ResourceChange
		want    ChangeStatistics
	}{
		{
			name:    "Empty changes should have all zeros",
			changes: []ResourceChange{},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        0,
			},
		},
		{
			name: "Single create should increment ToAdd and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate, ReplacementType: ReplacementNever},
			},
			want: ChangeStatistics{
				ToAdd:        1,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Single update should increment ToChange and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeUpdate, ReplacementType: ReplacementNever},
			},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     1,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Single delete should increment ToDestroy and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeDelete, ReplacementType: ReplacementNever},
			},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    1,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Replace with always replacement should increment Replacements and Total",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
			},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 1,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Dangerous sensitive resource should increment HighRisk",
			changes: []ResourceChange{
				{
					Type:        "aws_rds_instance",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 1,
				HighRisk:     1,
				Total:        1,
			},
		},
		{
			name: "Non-dangerous sensitive resource should not increment HighRisk",
			changes: []ResourceChange{
				{
					Type:        "aws_rds_instance",
					ChangeType:  ChangeTypeUpdate,
					IsDangerous: false,
				},
			},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     1,
				ToDestroy:    0,
				Replacements: 0,
				HighRisk:     0,
				Total:        1,
			},
		},
		{
			name: "Dangerous non-sensitive resource should increment HighRisk",
			changes: []ResourceChange{
				{
					Type:        "aws_s3_bucket",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			want: ChangeStatistics{
				ToAdd:        0,
				ToChange:     0,
				ToDestroy:    0,
				Replacements: 1,
				HighRisk:     1,
				Total:        1,
			},
		},
		{
			name: "Mixed changes should calculate correctly",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeUpdate, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeDelete, ReplacementType: ReplacementNever},
				{ChangeType: ChangeTypeReplace, ReplacementType: ReplacementAlways},
				{
					Type:        "aws_rds_instance",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
				{
					Type:        "aws_dynamodb_table",
					ChangeType:  ChangeTypeReplace,
					IsDangerous: true,
				},
			},
			want: ChangeStatistics{
				ToAdd:        1,
				ToChange:     1,
				ToDestroy:    1,
				Replacements: 3,
				HighRisk:     2,
				Total:        6,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzer.calculateStatistics(tc.changes, []OutputChange{})
			assert.Equal(t, tc.want.ToAdd, got.ToAdd, "ToAdd mismatch")
			assert.Equal(t, tc.want.ToChange, got.ToChange, "ToChange mismatch")
			assert.Equal(t, tc.want.ToDestroy, got.ToDestroy, "ToDestroy mismatch")
			assert.Equal(t, tc.want.Replacements, got.Replacements, "Replacements mismatch")
			assert.Equal(t, tc.want.HighRisk, got.HighRisk, "HighRisk mismatch")
			assert.Equal(t, tc.want.Total, got.Total, "Total mismatch")
			assert.Equal(t, tc.want.OutputChanges, got.OutputChanges, "OutputChanges mismatch")
		})
	}
}

// TestCalculateStatisticsWithOutputChanges tests statistics behavior with output changes, specifically verifying that no-op outputs are excluded
func TestCalculateStatisticsWithOutputChanges(t *testing.T) {
	// Create analyzer with empty config for this test
	analyzer := &Analyzer{
		config: &config.Config{},
	}

	testCases := []struct {
		name    string
		changes []ResourceChange
		outputs []OutputChange
		want    ChangeStatistics
	}{
		{
			name:    "No changes should have all zeros including OutputChanges",
			changes: []ResourceChange{},
			outputs: []OutputChange{},
			want: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    0,
				Total:         0,
				OutputChanges: 0,
			},
		},
		{
			name:    "Only output changes should count non-no-op outputs",
			changes: []ResourceChange{},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeUpdate, IsNoOp: false},
				{Name: "output2", ChangeType: ChangeTypeCreate, IsNoOp: false},
				{Name: "output3", ChangeType: ChangeTypeNoOp, IsNoOp: true}, // Should be excluded
			},
			want: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    0,
				Total:         0,
				OutputChanges: 2, // Only non-no-op outputs
			},
		},
		{
			name: "Mixed resource and output changes should count correctly",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeCreate},
				{ChangeType: ChangeTypeUpdate},
				{ChangeType: ChangeTypeNoOp}, // Should count in Unmodified but not Total
			},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeUpdate, IsNoOp: false},
				{Name: "output2", ChangeType: ChangeTypeNoOp, IsNoOp: true}, // Should be excluded from OutputChanges
				{Name: "output3", ChangeType: ChangeTypeDelete, IsNoOp: false},
			},
			want: ChangeStatistics{
				ToAdd:         1,
				ToChange:      1,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    1, // No-op resource counts in Unmodified
				Total:         2, // Only non-no-op resources
				OutputChanges: 2, // Only non-no-op outputs
			},
		},
		{
			name:    "All no-op outputs should result in zero OutputChanges",
			changes: []ResourceChange{},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeNoOp, IsNoOp: true},
				{Name: "output2", ChangeType: ChangeTypeNoOp, IsNoOp: true},
				{Name: "output3", ChangeType: ChangeTypeNoOp, IsNoOp: true},
			},
			want: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    0,
				Total:         0,
				OutputChanges: 0, // All outputs are no-ops, so excluded
			},
		},
		{
			name: "Resource no-ops should not affect output statistics",
			changes: []ResourceChange{
				{ChangeType: ChangeTypeNoOp}, // Resource no-op
				{ChangeType: ChangeTypeNoOp}, // Another resource no-op
			},
			outputs: []OutputChange{
				{Name: "output1", ChangeType: ChangeTypeUpdate, IsNoOp: false}, // Non-no-op output
			},
			want: ChangeStatistics{
				ToAdd:         0,
				ToChange:      0,
				ToDestroy:     0,
				Replacements:  0,
				HighRisk:      0,
				Unmodified:    2, // Resource no-ops count in Unmodified
				Total:         0, // No-op resources don't count in Total
				OutputChanges: 1, // Non-no-op output counts
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := analyzer.calculateStatistics(tc.changes, tc.outputs)
			assert.Equal(t, tc.want.ToAdd, got.ToAdd, "ToAdd mismatch")
			assert.Equal(t, tc.want.ToChange, got.ToChange, "ToChange mismatch")
			assert.Equal(t, tc.want.ToDestroy, got.ToDestroy, "ToDestroy mismatch")
			assert.Equal(t, tc.want.Replacements, got.Replacements, "Replacements mismatch")
			assert.Equal(t, tc.want.HighRisk, got.HighRisk, "HighRisk mismatch")
			assert.Equal(t, tc.want.Unmodified, got.Unmodified, "Unmodified mismatch")
			assert.Equal(t, tc.want.Total, got.Total, "Total mismatch")
			assert.Equal(t, tc.want.OutputChanges, got.OutputChanges, "OutputChanges mismatch")
		})
	}
}
