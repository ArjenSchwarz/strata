// Package errors provides error handling utilities and custom error types for the Strata application.
package errors

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewTerraformNotFoundError creates an error for when Terraform is not found
func NewTerraformNotFoundError(terraformPath string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodeTerraformNotFound,
		Message:    fmt.Sprintf("Terraform binary not found at path: %s", terraformPath),
		Underlying: err,
		Context: map[string]interface{}{
			"terraform_path": terraformPath,
			"search_path":    os.Getenv("PATH"),
		},
		Suggestions: []string{
			"Install Terraform from https://www.terraform.io/downloads.html",
			"Ensure Terraform is in your PATH environment variable",
			"Specify the correct path using --terraform-path flag",
			"Verify the binary has execute permissions",
		},
		RecoveryAction: "Install Terraform or update the path configuration",
	}
}

// NewTerraformNotExecutableError creates an error for when Terraform binary is not executable
func NewTerraformNotExecutableError(terraformPath string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodeTerraformNotExecutable,
		Message:    fmt.Sprintf("Terraform binary is not executable: %s", terraformPath),
		Underlying: err,
		Context: map[string]interface{}{
			"terraform_path": terraformPath,
		},
		Suggestions: []string{
			fmt.Sprintf("Run: chmod +x %s", terraformPath),
			"Verify the file is a valid Terraform binary",
			"Check file permissions and ownership",
		},
		RecoveryAction: fmt.Sprintf("Make the binary executable: chmod +x %s", terraformPath),
	}
}

// NewInvalidVersionError creates an error for unsupported Terraform versions
func NewInvalidVersionError(version string, minVersion string) *StrataError {
	return &StrataError{
		Code:    ErrorCodeInvalidVersion,
		Message: fmt.Sprintf("Unsupported Terraform version: %s (minimum required: %s)", version, minVersion),
		Context: map[string]interface{}{
			"current_version": version,
			"minimum_version": minVersion,
		},
		Suggestions: []string{
			fmt.Sprintf("Upgrade Terraform to version %s or higher", minVersion),
			"Use tfenv or similar tool to manage Terraform versions",
			"Check https://www.terraform.io/downloads.html for latest versions",
		},
		RecoveryAction: fmt.Sprintf("Upgrade Terraform to version %s or higher", minVersion),
	}
}

// NewWorkingDirNotFoundError creates an error for invalid working directories
func NewWorkingDirNotFoundError(workingDir string) *StrataError {
	absPath, _ := filepath.Abs(workingDir)
	return &StrataError{
		Code:    ErrorCodeWorkingDirNotFound,
		Message: fmt.Sprintf("Working directory not found: %s", workingDir),
		Context: map[string]interface{}{
			"working_dir":   workingDir,
			"absolute_path": absPath,
			"current_dir":   getCurrentDir(),
		},
		Suggestions: []string{
			"Verify the directory path is correct",
			"Create the directory if it doesn't exist",
			"Check directory permissions",
			"Use an absolute path instead of relative path",
		},
		RecoveryAction: fmt.Sprintf("Create the directory: mkdir -p %s", workingDir),
	}
}

// NewPlanFailedError creates an error for plan execution failures
func NewPlanFailedError(command string, exitCode int, output string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodePlanFailed,
		Message:    "Terraform plan execution failed",
		Underlying: err,
		Context: map[string]interface{}{
			"command":   command,
			"exit_code": exitCode,
			"output":    truncateOutput(output, 1000),
		},
		Suggestions: []string{
			"Review the Terraform configuration for syntax errors",
			"Check provider authentication and permissions",
			"Verify all required variables are set",
			"Run 'terraform validate' to check configuration",
			"Run 'terraform init' if providers need initialization",
		},
		RecoveryAction: "Fix configuration issues and retry the plan",
	}
}

// NewPlanTimeoutError creates an error for plan timeouts
func NewPlanTimeoutError(timeout string) *StrataError {
	return &StrataError{
		Code:    ErrorCodePlanTimeout,
		Message: fmt.Sprintf("Terraform plan timed out after %s", timeout),
		Context: map[string]interface{}{
			"timeout": timeout,
		},
		Suggestions: []string{
			"Increase the timeout using --timeout flag",
			"Check network connectivity to providers",
			"Verify provider endpoints are accessible",
			"Consider breaking down large configurations",
		},
		RecoveryAction: "Increase timeout or check network connectivity",
	}
}

// NewApplyFailedError creates an error for apply execution failures
func NewApplyFailedError(command string, exitCode int, output string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodeApplyFailed,
		Message:    "Terraform apply execution failed",
		Underlying: err,
		Context: map[string]interface{}{
			"command":   command,
			"exit_code": exitCode,
			"output":    truncateOutput(output, 1000),
		},
		Suggestions: []string{
			"Review the error output for specific failure reasons",
			"Check provider permissions and quotas",
			"Verify resource dependencies are correct",
			"Consider applying resources in smaller batches",
			"Check for resource conflicts or naming collisions",
		},
		RecoveryAction: "Review errors, fix issues, and retry the apply",
	}
}

// NewStateLockTimeoutError creates an error for state lock timeouts
func NewStateLockTimeoutError(backend string, timeout string) *StrataError {
	return &StrataError{
		Code:    ErrorCodeStateLockTimeout,
		Message: fmt.Sprintf("Timeout acquiring state lock for %s backend after %s", backend, timeout),
		Context: map[string]interface{}{
			"backend": backend,
			"timeout": timeout,
		},
		Suggestions: []string{
			"Wait for other Terraform operations to complete",
			"Check if another process is holding the lock",
			"Increase lock timeout if operations are expected to be long",
			"Force unlock if you're certain no other process is running (use with caution)",
		},
		RecoveryAction: "Wait for lock release or force unlock if safe",
	}
}

// NewStateLockConflictError creates an error for state lock conflicts
func NewStateLockConflictError(backend string, lockInfo string) *StrataError {
	return &StrataError{
		Code:    ErrorCodeStateLockConflict,
		Message: fmt.Sprintf("State is locked by another process on %s backend", backend),
		Context: map[string]interface{}{
			"backend":   backend,
			"lock_info": lockInfo,
		},
		Suggestions: []string{
			"Wait for the other Terraform operation to complete",
			"Check if the lock is stale (process no longer running)",
			"Contact team members who might be running Terraform",
			"Use 'terraform force-unlock' only if you're certain it's safe",
		},
		RecoveryAction: "Wait for lock release or coordinate with team",
	}
}

// NewStatePermissionsError creates an error for state access permission issues
func NewStatePermissionsError(backend string, operation string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodeStatePermissions,
		Message:    fmt.Sprintf("Insufficient permissions to %s state on %s backend", operation, backend),
		Underlying: err,
		Context: map[string]interface{}{
			"backend":   backend,
			"operation": operation,
		},
		Suggestions: []string{
			"Check your authentication credentials",
			"Verify you have the required permissions for the backend",
			"Contact your administrator to grant necessary permissions",
			"Ensure your credentials haven't expired",
		},
		RecoveryAction: "Update credentials or request appropriate permissions",
	}
}

// NewWorkflowCancelledError creates an error for cancelled workflows
func NewWorkflowCancelledError(reason string) *StrataError {
	return &StrataError{
		Code:    ErrorCodeWorkflowCancelled,
		Message: fmt.Sprintf("Workflow cancelled: %s", reason),
		Context: map[string]interface{}{
			"reason": reason,
		},
		Suggestions: []string{
			"Review the plan and try again if needed",
			"Use --force flag to bypass confirmations in non-interactive mode",
			"Address any concerns before retrying",
		},
		RecoveryAction: "Review and retry the operation if appropriate",
	}
}

// NewDestructiveChangesError creates an error for destructive changes without confirmation
func NewDestructiveChangesError(destructiveCount int, resources []string) *StrataError {
	return &StrataError{
		Code:    ErrorCodeDestructiveChanges,
		Message: fmt.Sprintf("Detected %d destructive changes that require confirmation", destructiveCount),
		Context: map[string]interface{}{
			"destructive_count": destructiveCount,
			"resources":         resources,
		},
		Suggestions: []string{
			"Review the destructive changes carefully",
			"Use interactive mode to confirm changes manually",
			"Use --force flag in non-interactive mode if you're certain",
			"Consider backing up affected resources first",
		},
		RecoveryAction: "Confirm destructive changes or use --force flag",
	}
}

// NewUserInputFailedError creates an error for user input failures
func NewUserInputFailedError(prompt string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodeUserInputFailed,
		Message:    fmt.Sprintf("Failed to read user input for: %s", prompt),
		Underlying: err,
		Context: map[string]interface{}{
			"prompt": prompt,
		},
		Suggestions: []string{
			"Ensure you're running in an interactive terminal",
			"Check if stdin is properly connected",
			"Use --non-interactive flag for automated environments",
		},
		RecoveryAction: "Run in interactive mode or use non-interactive flags",
	}
}

// NewPlanAnalysisFailedError creates an error for plan analysis failures
func NewPlanAnalysisFailedError(planFile string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodePlanAnalysisFailed,
		Message:    fmt.Sprintf("Failed to analyze plan file: %s", planFile),
		Underlying: err,
		Context: map[string]interface{}{
			"plan_file": planFile,
		},
		Suggestions: []string{
			"Verify the plan file exists and is readable",
			"Check if the plan file is corrupted",
			"Regenerate the plan file",
			"Ensure you have read permissions for the file",
		},
		RecoveryAction: "Regenerate the plan file and retry analysis",
	}
}

// NewInvalidPlanFormatError creates an error for invalid plan file formats
func NewInvalidPlanFormatError(planFile string, expectedFormat string) *StrataError {
	return &StrataError{
		Code:    ErrorCodeInvalidPlanFormat,
		Message: fmt.Sprintf("Invalid plan file format: %s (expected: %s)", planFile, expectedFormat),
		Context: map[string]interface{}{
			"plan_file":       planFile,
			"expected_format": expectedFormat,
		},
		Suggestions: []string{
			"Ensure the plan file was generated with a compatible Terraform version",
			"Regenerate the plan file with the current Terraform version",
			"Check if the file is corrupted or truncated",
		},
		RecoveryAction: "Regenerate the plan file with the current Terraform version",
	}
}

// NewSystemResourceExhaustedError creates an error for system resource exhaustion
func NewSystemResourceExhaustedError(resource string, err error) *StrataError {
	return &StrataError{
		Code:       ErrorCodeSystemResourceExhausted,
		Message:    fmt.Sprintf("System resource exhausted: %s", resource),
		Underlying: err,
		Context: map[string]interface{}{
			"resource": resource,
		},
		Suggestions: []string{
			"Free up system resources (memory, disk space, etc.)",
			"Close unnecessary applications",
			"Increase system limits if possible",
			"Run the operation on a machine with more resources",
		},
		RecoveryAction: "Free up system resources and retry",
	}
}

// Helper functions

// getCurrentDir returns the current working directory
func getCurrentDir() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "unknown"
}

// truncateOutput truncates output to a maximum length for context
func truncateOutput(output string, maxLength int) string {
	if len(output) <= maxLength {
		return output
	}
	return output[:maxLength] + "... (truncated)"
}
