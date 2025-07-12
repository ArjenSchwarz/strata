package errors

import (
	"fmt"
	"strings"
)

// ErrorCode represents specific error codes for different scenarios
type ErrorCode string

const (
	// ErrorCodeTerraformNotFound indicates that the Terraform binary was not found in the system PATH.
	ErrorCodeTerraformNotFound ErrorCode = "TERRAFORM_NOT_FOUND"
	// ErrorCodeTerraformNotExecutable indicates that the Terraform binary exists but is not executable.
	ErrorCodeTerraformNotExecutable ErrorCode = "TERRAFORM_NOT_EXECUTABLE"
	// ErrorCodeInvalidVersion indicates that an invalid version was specified or detected.
	ErrorCodeInvalidVersion ErrorCode = "INVALID_VERSION"
	// ErrorCodeWorkingDirNotFound indicates that the specified working directory was not found.
	ErrorCodeWorkingDirNotFound ErrorCode = "WORKING_DIR_NOT_FOUND"
	// ErrorCodeConfigurationInvalid indicates that the configuration is invalid or malformed.
	ErrorCodeConfigurationInvalid ErrorCode = "CONFIGURATION_INVALID"

	// ErrorCodePlanFailed indicates that the Terraform plan command failed to execute successfully.
	ErrorCodePlanFailed ErrorCode = "PLAN_FAILED"
	// ErrorCodePlanFileNotCreated indicates that the plan file was not created successfully.
	ErrorCodePlanFileNotCreated ErrorCode = "PLAN_FILE_NOT_CREATED"
	// ErrorCodePlanFileCorrupted indicates that the plan file is corrupted or unreadable.
	ErrorCodePlanFileCorrupted ErrorCode = "PLAN_FILE_CORRUPTED"
	// ErrorCodePlanTimeout indicates that the Terraform plan command timed out.
	ErrorCodePlanTimeout ErrorCode = "PLAN_TIMEOUT"
	// ErrorCodePlanInterrupted indicates that the Terraform plan command was interrupted.
	ErrorCodePlanInterrupted ErrorCode = "PLAN_INTERRUPTED"
	// ErrorCodeInvalidPlanArgs indicates that invalid arguments were provided to the plan command.
	ErrorCodeInvalidPlanArgs ErrorCode = "INVALID_PLAN_ARGS"

	// ErrorCodeApplyFailed indicates that the Terraform apply command failed to execute successfully.
	ErrorCodeApplyFailed ErrorCode = "APPLY_FAILED"
	// ErrorCodeApplyTimeout indicates that the Terraform apply command timed out.
	ErrorCodeApplyTimeout ErrorCode = "APPLY_TIMEOUT"
	// ErrorCodeApplyInterrupted indicates that the Terraform apply command was interrupted.
	ErrorCodeApplyInterrupted ErrorCode = "APPLY_INTERRUPTED"
	// ErrorCodeInvalidApplyArgs indicates that invalid arguments were provided to the apply command.
	ErrorCodeInvalidApplyArgs ErrorCode = "INVALID_APPLY_ARGS"
	// ErrorCodeApplyRollbackFailed indicates that the rollback after a failed apply operation failed.
	ErrorCodeApplyRollbackFailed ErrorCode = "APPLY_ROLLBACK_FAILED"

	// ErrorCodeStateLockTimeout indicates that a timeout occurred while trying to acquire a state lock.
	ErrorCodeStateLockTimeout ErrorCode = "STATE_LOCK_TIMEOUT"
	// ErrorCodeStateLockConflict indicates that there is a conflict with the state lock.
	ErrorCodeStateLockConflict ErrorCode = "STATE_LOCK_CONFLICT"
	// ErrorCodeStateBackendConfig indicates that the state backend configuration is invalid.
	ErrorCodeStateBackendConfig ErrorCode = "STATE_BACKEND_CONFIG"
	// ErrorCodeStatePermissions indicates insufficient permissions to access the state.
	ErrorCodeStatePermissions ErrorCode = "STATE_PERMISSIONS"
	// ErrorCodeStateNetworkTimeout indicates a network timeout while accessing remote state.
	ErrorCodeStateNetworkTimeout ErrorCode = "STATE_NETWORK_TIMEOUT"
	// ErrorCodeStateCorrupted indicates that the state file is corrupted.
	ErrorCodeStateCorrupted ErrorCode = "STATE_CORRUPTED"
	// ErrorCodeStateVersionMismatch indicates a version mismatch in the state file.
	ErrorCodeStateVersionMismatch ErrorCode = "STATE_VERSION_MISMATCH"

	// ErrorCodeWorkflowCancelled indicates that the workflow was cancelled by the user.
	ErrorCodeWorkflowCancelled ErrorCode = "WORKFLOW_CANCELLED"
	// ErrorCodeUserInputFailed indicates that user input collection failed.
	ErrorCodeUserInputFailed ErrorCode = "USER_INPUT_FAILED"
	// ErrorCodeInvalidUserInput indicates that the user provided invalid input.
	ErrorCodeInvalidUserInput ErrorCode = "INVALID_USER_INPUT"
	// ErrorCodeNonInteractiveForced indicates that non-interactive mode was forced when interaction was required.
	ErrorCodeNonInteractiveForced ErrorCode = "NON_INTERACTIVE_FORCED"
	// ErrorCodeDestructiveChanges indicates that destructive changes were detected.
	ErrorCodeDestructiveChanges ErrorCode = "DESTRUCTIVE_CHANGES"

	// ErrorCodePlanAnalysisFailed indicates that the analysis of the Terraform plan failed.
	ErrorCodePlanAnalysisFailed ErrorCode = "PLAN_ANALYSIS_FAILED"
	// ErrorCodeInvalidPlanFormat indicates that the plan file format is invalid or unsupported.
	ErrorCodeInvalidPlanFormat ErrorCode = "INVALID_PLAN_FORMAT"
	// ErrorCodeParsingFailed indicates that parsing of the plan file failed.
	ErrorCodeParsingFailed ErrorCode = "PARSING_FAILED"

	// ErrorCodeInsufficientPermissions indicates that the operation failed due to insufficient permissions.
	ErrorCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	// ErrorCodeDiskSpaceFull indicates that the disk is full and the operation cannot continue.
	ErrorCodeDiskSpaceFull ErrorCode = "DISK_SPACE_FULL"
	// ErrorCodeNetworkUnavailable indicates that the network is unavailable.
	ErrorCodeNetworkUnavailable ErrorCode = "NETWORK_UNAVAILABLE"
	// ErrorCodeSystemResourceExhausted indicates that system resources are exhausted.
	ErrorCodeSystemResourceExhausted ErrorCode = "SYSTEM_RESOURCE_EXHAUSTED"
	// ErrorCodeTempFileCleanupFailed indicates that cleanup of temporary files failed.
	ErrorCodeTempFileCleanupFailed ErrorCode = "TEMP_FILE_CLEANUP_FAILED"
)

// StrataError is the base error type for all Strata errors
type StrataError struct {
	Code           ErrorCode
	Message        string
	Context        map[string]any
	Underlying     error
	Suggestions    []string
	RecoveryAction string
}

// Error implements the error interface
func (e *StrataError) Error() string {
	if e.Underlying != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Underlying.Error())
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *StrataError) Unwrap() error {
	return e.Underlying
}

// GetCode returns the error code
func (e *StrataError) GetCode() ErrorCode {
	return e.Code
}

// GetContext returns the error context
func (e *StrataError) GetContext() map[string]any {
	if e.Context == nil {
		return make(map[string]any)
	}
	return e.Context
}

// GetSuggestions returns suggested resolutions
func (e *StrataError) GetSuggestions() []string {
	return e.Suggestions
}

// GetRecoveryAction returns the recovery action
func (e *StrataError) GetRecoveryAction() string {
	return e.RecoveryAction
}

// WithContext adds context information to the error
func (e *StrataError) WithContext(key string, value any) *StrataError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithSuggestion adds a suggestion to the error
func (e *StrataError) WithSuggestion(suggestion string) *StrataError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// WithRecoveryAction sets the recovery action
func (e *StrataError) WithRecoveryAction(action string) *StrataError {
	e.RecoveryAction = action
	return e
}

// FormatUserMessage formats a user-friendly error message
func (e *StrataError) FormatUserMessage() string {
	var parts []string

	// Add the main error message
	parts = append(parts, fmt.Sprintf("❌ Error: %s", e.Message))

	// Add context information if available
	if len(e.Context) > 0 {
		parts = append(parts, "\n📋 Details:")
		for key, value := range e.Context {
			parts = append(parts, fmt.Sprintf("  • %s: %v", key, value))
		}
	}

	// Add suggestions if available
	if len(e.Suggestions) > 0 {
		parts = append(parts, "\n💡 Suggestions:")
		for _, suggestion := range e.Suggestions {
			parts = append(parts, fmt.Sprintf("  • %s", suggestion))
		}
	}

	// Add recovery action if available
	if e.RecoveryAction != "" {
		parts = append(parts, fmt.Sprintf("\n🔧 Recovery Action: %s", e.RecoveryAction))
	}

	// Add underlying error if available and different from main message
	if e.Underlying != nil && e.Underlying.Error() != e.Message {
		parts = append(parts, fmt.Sprintf("\n🔍 Technical Details: %s", e.Underlying.Error()))
	}

	return strings.Join(parts, "")
}

// IsRecoverable returns true if the error can potentially be recovered from
func (e *StrataError) IsRecoverable() bool {
	return e.RecoveryAction != "" || len(e.Suggestions) > 0
}

// IsCritical returns true if the error represents a critical failure
func (e *StrataError) IsCritical() bool {
	criticalCodes := []ErrorCode{
		ErrorCodeStateCorrupted,
		ErrorCodeApplyRollbackFailed,
		ErrorCodeSystemResourceExhausted,
		ErrorCodeDiskSpaceFull,
	}

	for _, code := range criticalCodes {
		if e.Code == code {
			return true
		}
	}
	return false
}

// IsUserError returns true if the error is likely caused by user input
func (e *StrataError) IsUserError() bool {
	userErrorCodes := []ErrorCode{
		ErrorCodeInvalidUserInput,
		ErrorCodeInvalidPlanArgs,
		ErrorCodeInvalidApplyArgs,
		ErrorCodeConfigurationInvalid,
		ErrorCodeWorkingDirNotFound,
	}

	for _, code := range userErrorCodes {
		if e.Code == code {
			return true
		}
	}
	return false
}

// IsSystemError returns true if the error is caused by system issues
func (e *StrataError) IsSystemError() bool {
	systemErrorCodes := []ErrorCode{
		ErrorCodeInsufficientPermissions,
		ErrorCodeDiskSpaceFull,
		ErrorCodeNetworkUnavailable,
		ErrorCodeSystemResourceExhausted,
		ErrorCodeTerraformNotFound,
		ErrorCodeTerraformNotExecutable,
	}

	for _, code := range systemErrorCodes {
		if e.Code == code {
			return true
		}
	}
	return false
}
