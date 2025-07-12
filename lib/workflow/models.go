package workflow

import (
	"context"
	"time"

	"github.com/ArjenSchwarz/strata/lib/plan"
	"github.com/ArjenSchwarz/strata/lib/terraform"
)

// WorkflowManager handles the interactive workflow
type WorkflowManager interface {
	// Run executes the workflow
	Run(ctx context.Context, options *WorkflowOptions) error

	// PromptForAction prompts the user for action
	PromptForAction(summary *plan.PlanSummary) (Action, error)

	// DisplaySummary displays the plan summary
	DisplaySummary(summary *plan.PlanSummary) error

	// DisplayDetails displays detailed plan output
	DisplayDetails(planOutput string) error
}

// WorkflowOptions contains options for the Terraform workflow
type WorkflowOptions struct {
	// TerraformPath is the path to the Terraform binary
	TerraformPath string

	// WorkingDir is the directory to execute Terraform commands in
	WorkingDir string

	// PlanArgs are additional arguments for terraform plan
	PlanArgs []string

	// ApplyArgs are additional arguments for terraform apply
	ApplyArgs []string

	// NonInteractive indicates whether to run in non-interactive mode
	NonInteractive bool

	// Force indicates whether to force apply in non-interactive mode
	Force bool

	// OutputFormat is the format for output
	OutputFormat string

	// DangerThreshold is the threshold for dangerous changes
	DangerThreshold int

	// Timeout is the maximum time to wait for operations
	Timeout time.Duration

	// Environment variables to set for Terraform commands
	Environment map[string]string
}

// Action represents user actions in the workflow
type Action int

const (
	// ActionApply represents the action to apply the Terraform plan.
	ActionApply Action = iota
	// ActionViewDetails represents the action to view detailed information.
	ActionViewDetails
	// ActionCancel represents the action to cancel the workflow.
	ActionCancel
)

// String returns the string representation of an Action
func (a Action) String() string {
	switch a {
	case ActionApply:
		return "apply"
	case ActionViewDetails:
		return "view-details"
	case ActionCancel:
		return "cancel"
	default:
		return "unknown"
	}
}

// WorkflowResult contains the result of a workflow execution
type WorkflowResult struct {
	// Success indicates whether the workflow completed successfully
	Success bool

	// Action is the action that was taken
	Action Action

	// PlanSummary contains the plan summary if available
	PlanSummary *plan.PlanSummary

	// ApplyOutput contains the apply output if apply was executed
	ApplyOutput *terraform.ApplyOutput

	// Error contains any error that occurred
	Error error

	// Duration is the total time taken for the workflow
	Duration time.Duration
}

// WorkflowError represents errors that occur during workflow execution
// Deprecated: Use lib/errors.StrataError instead
type WorkflowError struct {
	Stage      string
	Message    string
	Underlying error
}

func (e *WorkflowError) Error() string {
	if e.Underlying != nil {
		return e.Message + ": " + e.Underlying.Error()
	}
	return e.Message
}

func (e *WorkflowError) Unwrap() error {
	return e.Underlying
}
