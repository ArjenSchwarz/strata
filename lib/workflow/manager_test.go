package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/ArjenSchwarz/strata/lib/plan"
)

func TestNewWorkflowManager(t *testing.T) {
	config := &config.Config{
		Plan: config.PlanConfig{
			DangerThreshold: 3,
		},
	}

	manager := NewWorkflowManager(config)
	require.NotNil(t, manager)

	// Verify it implements the interface
	var _ WorkflowManager = manager
}

func TestDefaultWorkflowManager_hasDestructiveChanges(t *testing.T) {
	manager := &DefaultWorkflowManager{}

	tests := []struct {
		name     string
		summary  *plan.PlanSummary
		expected bool
	}{
		{
			name: "no destructive changes",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{
						Address:       "aws_instance.example",
						IsDestructive: false,
					},
				},
			},
			expected: false,
		},
		{
			name: "has destructive changes",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{
						Address:       "aws_instance.example",
						IsDestructive: false,
					},
					{
						Address:       "aws_db_instance.example",
						IsDestructive: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "empty resource changes",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.hasDestructiveChanges(tt.summary)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultWorkflowManager_hasDangerousChanges(t *testing.T) {
	manager := &DefaultWorkflowManager{}

	tests := []struct {
		name      string
		summary   *plan.PlanSummary
		threshold int
		expected  bool
	}{
		{
			name: "below threshold",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{IsDestructive: true},
					{IsDestructive: true},
				},
			},
			threshold: 3,
			expected:  false,
		},
		{
			name: "at threshold",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{IsDestructive: true},
					{IsDestructive: true},
					{IsDestructive: true},
				},
			},
			threshold: 3,
			expected:  true,
		},
		{
			name: "above threshold",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{IsDestructive: true},
					{IsDestructive: true},
					{IsDestructive: true},
					{IsDestructive: true},
				},
			},
			threshold: 3,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.hasDangerousChanges(tt.summary, tt.threshold)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultWorkflowManager_countDestructiveChanges(t *testing.T) {
	manager := &DefaultWorkflowManager{}

	tests := []struct {
		name     string
		summary  *plan.PlanSummary
		expected int
	}{
		{
			name: "no destructive changes",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{IsDestructive: false},
					{IsDestructive: false},
				},
			},
			expected: 0,
		},
		{
			name: "some destructive changes",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{IsDestructive: false},
					{IsDestructive: true},
					{IsDestructive: true},
				},
			},
			expected: 2,
		},
		{
			name: "all destructive changes",
			summary: &plan.PlanSummary{
				ResourceChanges: []plan.ResourceChange{
					{IsDestructive: true},
					{IsDestructive: true},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.countDestructiveChanges(tt.summary)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAction_String(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionApply, "apply"},
		{ActionViewDetails, "view-details"},
		{ActionCancel, "cancel"},
		{Action(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.action.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
