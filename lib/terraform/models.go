package terraform

import (
	"context"
	"fmt"
	"time"
)

// TerraformExecutor handles execution of Terraform commands
type TerraformExecutor interface {
	// Plan executes terraform plan and returns the path to the plan file
	Plan(ctx context.Context, args []string) (string, error)

	// Apply executes terraform apply with the given plan file
	Apply(ctx context.Context, planFile string, args []string) error

	// GetVersion returns the Terraform version
	GetVersion(ctx context.Context) (string, error)

	// CheckInstallation verifies that Terraform is installed and accessible
	CheckInstallation(ctx context.Context) error

	// DetectBackend detects the backend configuration from Terraform files
	DetectBackend(ctx context.Context) (*BackendConfig, error)

	// ValidateBackend validates the backend configuration
	ValidateBackend(ctx context.Context, config *BackendConfig) error
}

// TerraformOutputParser parses Terraform command output
type TerraformOutputParser interface {
	// ParsePlanOutput parses the output of terraform plan
	ParsePlanOutput(output string) (*PlanOutput, error)

	// ParseApplyOutput parses the output of terraform apply
	ParseApplyOutput(output string) (*ApplyOutput, error)
}

// ExecutorOptions contains options for the Terraform executor
type ExecutorOptions struct {
	// TerraformPath is the path to the Terraform binary
	TerraformPath string

	// WorkingDir is the directory to execute Terraform commands in
	WorkingDir string

	// Timeout is the maximum time to wait for commands to complete
	Timeout time.Duration

	// Environment variables to set for Terraform commands
	Environment map[string]string

	// BackendConfig contains backend-specific configuration
	BackendConfig *BackendConfig
}

// BackendConfig contains configuration for Terraform backends
type BackendConfig struct {
	// Type is the backend type (e.g., "s3", "gcs", "azurerm", "local")
	Type string

	// Config contains backend-specific configuration parameters
	Config map[string]interface{}

	// LockTimeout is the timeout for acquiring state locks
	LockTimeout time.Duration

	// DisableLocking disables state locking entirely
	DisableLocking bool
}

// PlanOutput contains parsed output from terraform plan
type PlanOutput struct {
	// HasChanges indicates whether the plan has changes
	HasChanges bool

	// ResourceChanges contains the number of resource changes
	ResourceChanges struct {
		Add     int
		Change  int
		Destroy int
	}

	// PlanFile is the path to the generated plan file
	PlanFile string

	// RawOutput is the raw output from terraform plan
	RawOutput string

	// ExitCode is the exit code from the terraform plan command
	ExitCode int
}

// ApplyOutput contains parsed output from terraform apply
type ApplyOutput struct {
	// Success indicates whether the apply was successful
	Success bool

	// ResourceChanges contains the number of resource changes applied
	ResourceChanges struct {
		Added     int
		Changed   int
		Destroyed int
	}

	// Error contains error information if the apply failed
	Error string

	// RawOutput is the raw output from terraform apply
	RawOutput string

	// ExitCode is the exit code from the terraform apply command
	ExitCode int
}

// Legacy error types - deprecated, use lib/errors package instead
// These are kept for backward compatibility

// CommandError represents an error from executing a Terraform command
// Deprecated: Use lib/errors.StrataError with appropriate error codes instead
type CommandError struct {
	Command    string
	ExitCode   int
	Output     string
	Stderr     string
	Underlying error
}

// StateError represents errors related to Terraform state operations
// Deprecated: Use lib/errors.StrataError with state-related error codes instead
type StateError struct {
	Type       StateErrorType
	Message    string
	Backend    string
	Underlying error
}

// StateErrorType represents different types of state-related errors
// Deprecated: Use lib/errors.ErrorCode instead
type StateErrorType string

const (
	StateErrorLockTimeout    StateErrorType = "lock_timeout"
	StateErrorLockConflict   StateErrorType = "lock_conflict"
	StateErrorBackendConfig  StateErrorType = "backend_config"
	StateErrorPermissions    StateErrorType = "permissions"
	StateErrorNetworkTimeout StateErrorType = "network_timeout"
)

func (e *CommandError) Error() string {
	if e.Underlying != nil {
		return e.Underlying.Error()
	}
	return e.Output
}

func (e *CommandError) Unwrap() error {
	return e.Underlying
}

func (e *StateError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("state error (%s): %s - %v", e.Type, e.Message, e.Underlying)
	}
	return fmt.Sprintf("state error (%s): %s", e.Type, e.Message)
}

func (e *StateError) Unwrap() error {
	return e.Underlying
}
