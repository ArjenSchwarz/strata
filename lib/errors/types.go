package errors

import (
	"fmt"
	"strings"
)

// ErrorCode represents specific error codes for different scenarios
type ErrorCode string

const (
	// Installation and setup errors
	ErrorCodeTerraformNotFound      ErrorCode = "TERRAFORM_NOT_FOUND"
	ErrorCodeTerraformNotExecutable ErrorCode = "TERRAFORM_NOT_EXECUTABLE"
	ErrorCodeInvalidVersion         ErrorCode = "INVALID_VERSION"
	ErrorCodeWorkingDirNotFound     ErrorCode = "WORKING_DIR_NOT_FOUND"
	ErrorCodeConfigurationInvalid   ErrorCode = "CONFIGURATION_INVALID"

	// Plan execution errors
	ErrorCodePlanFailed         ErrorCode = "PLAN_FAILED"
	ErrorCodePlanFileNotCreated ErrorCode = "PLAN_FILE_NOT_CREATED"
	ErrorCodePlanFileCorrupted  ErrorCode = "PLAN_FILE_CORRUPTED"
	ErrorCodePlanTimeout        ErrorCode = "PLAN_TIMEOUT"
	ErrorCodePlanInterrupted    ErrorCode = "PLAN_INTERRUPTED"
	ErrorCodeInvalidPlanArgs    ErrorCode = "INVALID_PLAN_ARGS"

	// Apply execution errors
	ErrorCodeApplyFailed         ErrorCode = "APPLY_FAILED"
	ErrorCodeApplyTimeout        ErrorCode = "APPLY_TIMEOUT"
	ErrorCodeApplyInterrupted    ErrorCode = "APPLY_INTERRUPTED"
	ErrorCodeInvalidApplyArgs    ErrorCode = "INVALID_APPLY_ARGS"
	ErrorCodeApplyRollbackFailed ErrorCode = "APPLY_ROLLBACK_FAILED"

	// State management errors
	ErrorCodeStateLockTimeout     ErrorCode = "STATE_LOCK_TIMEOUT"
	ErrorCodeStateLockConflict    ErrorCode = "STATE_LOCK_CONFLICT"
	ErrorCodeStateBackendConfig   ErrorCode = "STATE_BACKEND_CONFIG"
	ErrorCodeStatePermissions     ErrorCode = "STATE_PERMISSIONS"
	ErrorCodeStateNetworkTimeout  ErrorCode = "STATE_NETWORK_TIMEOUT"
	ErrorCodeStateCorrupted       ErrorCode = "STATE_CORRUPTED"
	ErrorCodeStateVersionMismatch ErrorCode = "STATE_VERSION_MISMATCH"

	// Workflow errors
	ErrorCodeWorkflowCancelled    ErrorCode = "WORKFLOW_CANCELLED"
	ErrorCodeUserInputFailed      ErrorCode = "USER_INPUT_FAILED"
	ErrorCodeInvalidUserInput     ErrorCode = "INVALID_USER_INPUT"
	ErrorCodeNonInteractiveForced ErrorCode = "NON_INTERACTIVE_FORCED"
	ErrorCodeDestructiveChanges   ErrorCode = "DESTRUCTIVE_CHANGES"

	// Analysis errors
	ErrorCodePlanAnalysisFailed ErrorCode = "PLAN_ANALYSIS_FAILED"
	ErrorCodeInvalidPlanFormat  ErrorCode = "INVALID_PLAN_FORMAT"
	ErrorCodeParsingFailed      ErrorCode = "PARSING_FAILED"

	// System errors
	ErrorCodeInsufficientPermissions ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrorCodeDiskSpaceFull           ErrorCode = "DISK_SPACE_FULL"
	ErrorCodeNetworkUnavailable      ErrorCode = "NETWORK_UNAVAILABLE"
	ErrorCodeSystemResourceExhausted ErrorCode = "SYSTEM_RESOURCE_EXHAUSTED"
	ErrorCodeTempFileCleanupFailed   ErrorCode = "TEMP_FILE_CLEANUP_FAILED"
)

// StrataError is the base error type for all Strata errors
type StrataError struct {
	Code           ErrorCode
	Message        string
	Context        map[string]interface{}
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
func (e *StrataError) GetContext() map[string]interface{} {
	if e.Context == nil {
		return make(map[string]interface{})
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
func (e *StrataError) WithContext(key string, value interface{}) *StrataError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
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
