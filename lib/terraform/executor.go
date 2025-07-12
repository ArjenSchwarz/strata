package terraform

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ArjenSchwarz/strata/lib/errors"
)

// DefaultExecutor is the default implementation of TerraformExecutor
type DefaultExecutor struct {
	options *ExecutorOptions
}

// NewExecutor creates a new Terraform executor with the given options
func NewExecutor(options *ExecutorOptions) TerraformExecutor {
	if options == nil {
		options = &ExecutorOptions{
			TerraformPath: "terraform",
			WorkingDir:    ".",
			Timeout:       30 * time.Minute,
			Environment:   make(map[string]string),
		}
	}

	// Set default values if not provided
	if options.TerraformPath == "" {
		options.TerraformPath = "terraform"
	}
	if options.WorkingDir == "" {
		options.WorkingDir = "."
	}
	if options.Timeout == 0 {
		options.Timeout = 30 * time.Minute
	}
	if options.Environment == nil {
		options.Environment = make(map[string]string)
	}

	return &DefaultExecutor{
		options: options,
	}
}

// CheckInstallation verifies that Terraform is installed and accessible
func (e *DefaultExecutor) CheckInstallation(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, e.options.TerraformPath, "version")
	cmd.Dir = e.options.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range e.options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	_, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "executable file not found") ||
			strings.Contains(err.Error(), "no such file or directory") {
			return errors.NewTerraformNotFoundError(e.options.TerraformPath, err)
		}
		// Check if it's a permission error
		if strings.Contains(err.Error(), "permission denied") {
			return errors.NewTerraformNotExecutableError(e.options.TerraformPath, err)
		}
		// Fallback to generic error
		return errors.NewTerraformNotFoundError(e.options.TerraformPath, err)
	}

	return nil
}

// GetVersion returns the Terraform version
func (e *DefaultExecutor) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, e.options.TerraformPath, "version", "-json")
	cmd.Dir = e.options.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range e.options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	output, err := cmd.Output()
	if err != nil {
		return "", errors.NewPlanFailedError(
			"terraform version -json",
			cmd.ProcessState.ExitCode(),
			string(output),
			err,
		).WithContext("operation", "version_check")
	}

	// For now, return the raw output - we can parse JSON later if needed
	return strings.TrimSpace(string(output)), nil
}

// Plan executes terraform plan and returns the path to the plan file
func (e *DefaultExecutor) Plan(ctx context.Context, args []string) (string, error) {
	fmt.Println("Generating Terraform plan...")

	// Generate a unique plan file name
	planFile := filepath.Join(e.options.WorkingDir, fmt.Sprintf("terraform-%d.tfplan", time.Now().Unix()))

	// Set up cleanup for the plan file in case of failure
	var cleanupPlanFile bool
	defer func() {
		if cleanupPlanFile {
			if err := e.cleanupTempFile(planFile); err != nil {
				fmt.Printf("Warning: Failed to cleanup temporary plan file %s: %v\n", planFile, err)
			}
		}
	}()

	// Build the command arguments
	cmdArgs := []string{"plan", "-out=" + planFile, "-input=false"}
	cmdArgs = append(cmdArgs, args...)

	// Create the command with timeout
	ctx, cancel := context.WithTimeout(ctx, e.options.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.options.TerraformPath, cmdArgs...)
	cmd.Dir = e.options.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range e.options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Set up pipes for real-time output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cleanupPlanFile = true
		return "", e.wrapPipeError("stdout", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cleanupPlanFile = true
		return "", e.wrapPipeError("stderr", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		cleanupPlanFile = true
		return "", e.enhancePlanStartError(cmdArgs, err)
	}

	// Stream output in real-time
	var outputBuffer strings.Builder
	done := make(chan error, 1)

	go func() {
		defer close(done)

		// Combine stdout and stderr for monitoring
		combined := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(combined)

		for scanner.Scan() {
			line := scanner.Text()
			outputBuffer.WriteString(line + "\n")

			// Print to console for real-time feedback
			fmt.Println(line)
		}

		if err := scanner.Err(); err != nil {
			done <- err
			return
		}

		done <- nil
	}()

	// Wait for the command to complete
	cmdErr := cmd.Wait()
	streamErr := <-done

	if streamErr != nil {
		cleanupPlanFile = true
		return "", e.wrapStreamError(streamErr)
	}

	if cmdErr != nil {
		cleanupPlanFile = true

		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", e.enhancePlanTimeoutError(cmdArgs, outputBuffer.String())
		}

		// Enhanced error handling with recovery suggestions
		return "", e.enhancePlanFailedError(cmdArgs, cmd.ProcessState.ExitCode(), outputBuffer.String(), cmdErr)
	}

	// Verify the plan file was created
	if _, err := os.Stat(planFile); os.IsNotExist(err) {
		cleanupPlanFile = true
		return "", e.enhancePlanFileNotCreatedError(planFile, outputBuffer.String())
	}

	fmt.Println("Plan generated successfully")
	return planFile, nil
}

// Apply executes terraform apply with the given plan file
func (e *DefaultExecutor) Apply(ctx context.Context, planFile string, args []string) error {
	fmt.Println("Applying Terraform changes...")

	// Set up cleanup for the plan file after apply (success or failure)
	defer func() {
		if err := e.cleanupTempFile(planFile); err != nil {
			fmt.Printf("Warning: Failed to cleanup plan file %s: %v\n", planFile, err)
		}
	}()

	// Build the command arguments
	cmdArgs := []string{"apply", "-input=false"}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, planFile)

	// Create the command with timeout
	ctx, cancel := context.WithTimeout(ctx, e.options.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.options.TerraformPath, cmdArgs...)
	cmd.Dir = e.options.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range e.options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Set up pipes for real-time output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return e.wrapPipeError("stdout", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return e.wrapPipeError("stderr", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return e.enhanceApplyStartError(cmdArgs, err)
	}

	// Stream output in real-time
	var outputBuffer strings.Builder
	done := make(chan error, 1)

	go func() {
		defer close(done)

		// Combine stdout and stderr for monitoring
		combined := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(combined)

		for scanner.Scan() {
			line := scanner.Text()
			outputBuffer.WriteString(line + "\n")

			// Print to console for real-time feedback
			fmt.Println(line)
		}

		if err := scanner.Err(); err != nil {
			done <- err
			return
		}

		done <- nil
	}()

	// Wait for the command to complete
	cmdErr := cmd.Wait()
	streamErr := <-done

	if streamErr != nil {
		return e.wrapStreamError(streamErr)
	}

	if cmdErr != nil {
		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			return e.enhanceApplyTimeoutError(cmdArgs, outputBuffer.String())
		}

		// Enhanced error handling with recovery suggestions
		return e.enhanceApplyFailedError(cmdArgs, cmd.ProcessState.ExitCode(), outputBuffer.String(), cmdErr)
	}

	fmt.Println("Changes applied successfully")
	return nil
}

// DetectBackend detects the backend configuration from Terraform files
func (e *DefaultExecutor) DetectBackend(ctx context.Context) (*BackendConfig, error) {
	// Try to get backend configuration using terraform show -json
	cmd := exec.CommandContext(ctx, e.options.TerraformPath, "show", "-json")
	cmd.Dir = e.options.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range e.options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	output, err := cmd.Output()
	if err != nil {
		// If show fails, try to detect from configuration files
		return e.detectBackendFromConfig()
	}

	// Parse the JSON output to extract backend information
	// For now, return a basic detection - this can be enhanced with JSON parsing
	return e.parseBackendFromOutput(string(output))
}

// ValidateBackend validates the backend configuration
func (e *DefaultExecutor) ValidateBackend(ctx context.Context, config *BackendConfig) error {
	if config == nil {
		return nil // No backend config to validate
	}

	// Try to initialize the backend
	cmd := exec.CommandContext(ctx, e.options.TerraformPath, "init", "-backend=true", "-input=false")
	cmd.Dir = e.options.WorkingDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range e.options.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return e.parseStateError(string(output), err)
	}

	return nil
}

// detectBackendFromConfig attempts to detect backend from Terraform configuration files
func (e *DefaultExecutor) detectBackendFromConfig() (*BackendConfig, error) {
	// This is a simplified implementation
	// In a full implementation, we would parse .tf files to extract backend configuration
	return &BackendConfig{
		Type:           "local",
		Config:         make(map[string]interface{}),
		LockTimeout:    10 * time.Minute,
		DisableLocking: false,
	}, nil
}

// parseBackendFromOutput parses backend information from terraform show output
func (e *DefaultExecutor) parseBackendFromOutput(output string) (*BackendConfig, error) {
	// This is a simplified implementation
	// In a full implementation, we would parse the JSON output to extract backend details
	if strings.Contains(output, `"backend"`) {
		return &BackendConfig{
			Type:           "remote",
			Config:         make(map[string]interface{}),
			LockTimeout:    10 * time.Minute,
			DisableLocking: false,
		}, nil
	}

	return &BackendConfig{
		Type:           "local",
		Config:         make(map[string]interface{}),
		LockTimeout:    10 * time.Minute,
		DisableLocking: false,
	}, nil
}

// parseStateError parses Terraform output to identify specific state-related errors
func (e *DefaultExecutor) parseStateError(output string, originalErr error) error {
	output = strings.ToLower(output)

	if strings.Contains(output, "lock") && strings.Contains(output, "timeout") {
		return errors.NewStateLockTimeoutError("unknown", "unknown").
			WithContext("output", output).
			WithContext("original_error", originalErr.Error())
	}

	if strings.Contains(output, "lock") && (strings.Contains(output, "conflict") || strings.Contains(output, "already locked")) {
		return errors.NewStateLockConflictError("unknown", extractLockInfo(output)).
			WithContext("output", output).
			WithContext("original_error", originalErr.Error())
	}

	if strings.Contains(output, "backend") && strings.Contains(output, "configuration") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeStateBackendConfig,
			Message:    "Backend configuration error",
			Underlying: originalErr,
			Context: map[string]interface{}{
				"output": output,
			},
			Suggestions: []string{
				"Check backend configuration in your Terraform files",
				"Verify backend credentials and permissions",
				"Run 'terraform init' to reconfigure backend",
			},
			RecoveryAction: "Fix backend configuration and run 'terraform init'",
		}
	}

	if strings.Contains(output, "permission") || strings.Contains(output, "access denied") || strings.Contains(output, "unauthorized") {
		return errors.NewStatePermissionsError("unknown", "access", originalErr).
			WithContext("output", output)
	}

	if strings.Contains(output, "timeout") || strings.Contains(output, "network") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeStateNetworkTimeout,
			Message:    "Network timeout while accessing backend",
			Underlying: originalErr,
			Context: map[string]interface{}{
				"output": output,
			},
			Suggestions: []string{
				"Check network connectivity",
				"Verify backend endpoints are accessible",
				"Consider increasing network timeout settings",
			},
			RecoveryAction: "Check network connectivity and retry",
		}
	}

	// If we can't classify the error, return the original
	return originalErr
}

// extractLockInfo extracts lock information from Terraform output
func extractLockInfo(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "lock") && (strings.Contains(line, "id") || strings.Contains(line, "info")) {
			return strings.TrimSpace(line)
		}
	}
	return "Lock information not available"
}

// Error recovery helper methods

// cleanupTempFile safely removes a temporary file
func (e *DefaultExecutor) cleanupTempFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	// Check if file exists before attempting to remove
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clean up
	}

	// Attempt to remove the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove temporary file %s: %w", filePath, err)
	}

	return nil
}

// wrapPipeError wraps pipe creation errors with recovery suggestions
func (e *DefaultExecutor) wrapPipeError(pipeType string, err error) error {
	return &errors.StrataError{
		Code:       errors.ErrorCodeSystemResourceExhausted,
		Message:    fmt.Sprintf("Failed to create %s pipe for Terraform command", pipeType),
		Underlying: err,
		Context: map[string]interface{}{
			"pipe_type": pipeType,
		},
		Suggestions: []string{
			"Check if system has sufficient resources (memory, file descriptors)",
			"Close unnecessary applications to free up system resources",
			"Restart the terminal or shell session",
			"Check system limits with 'ulimit -a'",
		},
		RecoveryAction: "Free up system resources and retry",
	}
}

// enhancePlanStartError enhances plan command start errors with recovery suggestions
func (e *DefaultExecutor) enhancePlanStartError(cmdArgs []string, err error) error {
	errStr := strings.ToLower(err.Error())

	// Check for common error patterns and provide specific guidance
	if strings.Contains(errStr, "permission denied") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeInsufficientPermissions,
			Message:    "Permission denied when starting Terraform plan",
			Underlying: err,
			Context: map[string]interface{}{
				"command":        fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
				"terraform_path": e.options.TerraformPath,
				"working_dir":    e.options.WorkingDir,
			},
			Suggestions: []string{
				fmt.Sprintf("Check if Terraform binary is executable: chmod +x %s", e.options.TerraformPath),
				fmt.Sprintf("Verify write permissions in working directory: %s", e.options.WorkingDir),
				"Check if the working directory exists and is accessible",
				"Run with appropriate user permissions",
			},
			RecoveryAction: "Fix file permissions and retry",
		}
	}

	if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "not found") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeTerraformNotFound,
			Message:    "Terraform binary not found when starting plan",
			Underlying: err,
			Context: map[string]interface{}{
				"terraform_path": e.options.TerraformPath,
				"working_dir":    e.options.WorkingDir,
			},
			Suggestions: []string{
				"Install Terraform from https://www.terraform.io/downloads.html",
				"Verify Terraform is in your PATH environment variable",
				"Use --terraform-path flag to specify the correct path",
				"Check if the specified Terraform path is correct",
			},
			RecoveryAction: "Install Terraform or specify correct path",
		}
	}

	// Generic plan start error
	return errors.NewPlanFailedError(
		fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
		0,
		"",
		err,
	).WithContext("stage", "command_start").
		WithSuggestion("Check Terraform installation and permissions").
		WithSuggestion("Verify working directory is accessible").
		WithRecoveryAction("Fix system configuration and retry")
}

// wrapStreamError wraps output streaming errors with recovery suggestions
func (e *DefaultExecutor) wrapStreamError(err error) error {
	return &errors.StrataError{
		Code:       errors.ErrorCodeSystemResourceExhausted,
		Message:    "Error streaming Terraform command output",
		Underlying: err,
		Context: map[string]interface{}{
			"operation": "output_streaming",
		},
		Suggestions: []string{
			"Check system resources (memory, CPU)",
			"Reduce concurrent operations",
			"Check for system stability issues",
			"Restart the application if the problem persists",
		},
		RecoveryAction: "Check system resources and retry",
	}
}

// enhancePlanTimeoutError enhances plan timeout errors with specific recovery suggestions
func (e *DefaultExecutor) enhancePlanTimeoutError(cmdArgs []string, output string) error {
	// Analyze output to provide more specific suggestions
	outputLower := strings.ToLower(output)
	suggestions := []string{
		"Increase the timeout using --timeout flag",
		"Check network connectivity to provider endpoints",
	}

	if strings.Contains(outputLower, "downloading") || strings.Contains(outputLower, "initializing") {
		suggestions = append(suggestions,
			"Provider initialization may be slow - consider pre-downloading providers",
			"Check internet connectivity for provider downloads",
		)
	}

	if strings.Contains(outputLower, "refreshing") || strings.Contains(outputLower, "reading") {
		suggestions = append(suggestions,
			"State refresh may be slow - check backend connectivity",
			"Consider using -refresh=false if state is known to be current",
		)
	}

	if strings.Contains(outputLower, "aws") || strings.Contains(outputLower, "amazon") {
		suggestions = append(suggestions,
			"Check AWS credentials and region configuration",
			"Verify AWS service endpoints are accessible",
		)
	}

	return &errors.StrataError{
		Code:    errors.ErrorCodePlanTimeout,
		Message: fmt.Sprintf("Terraform plan timed out after %s", e.options.Timeout.String()),
		Context: map[string]interface{}{
			"timeout": e.options.Timeout.String(),
			"command": fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			"output":  truncateOutput(output, 500),
		},
		Suggestions:    suggestions,
		RecoveryAction: "Increase timeout or check connectivity issues",
	}
}

// enhancePlanFailedError enhances plan failure errors with specific recovery suggestions
func (e *DefaultExecutor) enhancePlanFailedError(cmdArgs []string, exitCode int, output string, err error) error {
	outputLower := strings.ToLower(output)
	suggestions := []string{
		"Review the Terraform configuration for syntax errors",
		"Run 'terraform validate' to check configuration",
	}
	recoveryAction := "Fix configuration issues and retry the plan"

	// Analyze output for specific error patterns
	if strings.Contains(outputLower, "authentication") || strings.Contains(outputLower, "credentials") {
		suggestions = append(suggestions,
			"Check provider authentication credentials",
			"Verify environment variables or credential files",
			"Ensure credentials have not expired",
		)
		recoveryAction = "Fix authentication credentials and retry"
	}

	if strings.Contains(outputLower, "permission") || strings.Contains(outputLower, "access denied") {
		suggestions = append(suggestions,
			"Check provider permissions and IAM policies",
			"Verify the credentials have sufficient permissions",
			"Review resource-specific access requirements",
		)
		recoveryAction = "Fix permissions and retry"
	}

	if strings.Contains(outputLower, "not found") || strings.Contains(outputLower, "does not exist") {
		suggestions = append(suggestions,
			"Check if referenced resources exist",
			"Verify resource names and identifiers",
			"Ensure resources are in the correct region/account",
		)
		recoveryAction = "Fix resource references and retry"
	}

	if strings.Contains(outputLower, "syntax") || strings.Contains(outputLower, "invalid") {
		suggestions = append(suggestions,
			"Check Terraform configuration syntax",
			"Verify HCL syntax is correct",
			"Use 'terraform fmt' to format configuration",
		)
		recoveryAction = "Fix syntax errors and retry"
	}

	if strings.Contains(outputLower, "variable") && strings.Contains(outputLower, "required") {
		suggestions = append(suggestions,
			"Set required variables using -var flags",
			"Create a terraform.tfvars file with variable values",
			"Set variables using environment variables (TF_VAR_*)",
		)
		recoveryAction = "Set required variables and retry"
	}

	if strings.Contains(outputLower, "provider") && strings.Contains(outputLower, "not found") {
		suggestions = append(suggestions,
			"Run 'terraform init' to download required providers",
			"Check provider configuration and version constraints",
			"Verify provider source and version are correct",
		)
		recoveryAction = "Run 'terraform init' and retry"
	}

	return &errors.StrataError{
		Code:       errors.ErrorCodePlanFailed,
		Message:    "Terraform plan execution failed",
		Underlying: err,
		Context: map[string]interface{}{
			"command":   fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			"exit_code": exitCode,
			"output":    truncateOutput(output, 1000),
		},
		Suggestions:    suggestions,
		RecoveryAction: recoveryAction,
	}
}

// enhancePlanFileNotCreatedError enhances plan file creation errors
func (e *DefaultExecutor) enhancePlanFileNotCreatedError(planFile string, output string) error {
	outputLower := strings.ToLower(output)
	suggestions := []string{
		"Check if Terraform has write permissions in the working directory",
		"Verify there's sufficient disk space",
	}

	// Analyze output for specific issues
	if strings.Contains(outputLower, "permission") || strings.Contains(outputLower, "access denied") {
		suggestions = append(suggestions,
			fmt.Sprintf("Fix write permissions for directory: %s", filepath.Dir(planFile)),
			"Run with appropriate user permissions",
		)
	}

	if strings.Contains(outputLower, "space") || strings.Contains(outputLower, "disk full") {
		suggestions = append(suggestions,
			"Free up disk space in the working directory",
			"Check disk usage with 'df -h'",
		)
	}

	if strings.Contains(outputLower, "no changes") {
		// This might not be an error - Terraform might not create a plan file if there are no changes
		return &errors.StrataError{
			Code:    errors.ErrorCodePlanFileNotCreated,
			Message: "No plan file created - possibly no changes detected",
			Context: map[string]interface{}{
				"plan_file":   planFile,
				"working_dir": e.options.WorkingDir,
				"output":      truncateOutput(output, 500),
			},
			Suggestions: []string{
				"This may be normal if there are no changes to apply",
				"Check the plan output to confirm no changes were detected",
				"Verify your configuration includes the expected resources",
			},
			RecoveryAction: "Review configuration if changes were expected",
		}
	}

	return &errors.StrataError{
		Code:    errors.ErrorCodePlanFileNotCreated,
		Message: fmt.Sprintf("Plan file was not created: %s", planFile),
		Context: map[string]interface{}{
			"plan_file":   planFile,
			"working_dir": e.options.WorkingDir,
			"output":      truncateOutput(output, 500),
		},
		Suggestions:    suggestions,
		RecoveryAction: "Ensure write permissions and sufficient disk space",
	}
}

// truncateOutput truncates output to a maximum length for context
func truncateOutput(output string, maxLength int) string {
	if len(output) <= maxLength {
		return output
	}
	return output[:maxLength] + "... (truncated)"
}

// enhanceApplyStartError enhances apply command start errors with recovery suggestions
func (e *DefaultExecutor) enhanceApplyStartError(cmdArgs []string, err error) error {
	errStr := strings.ToLower(err.Error())

	// Check for common error patterns and provide specific guidance
	if strings.Contains(errStr, "permission denied") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeInsufficientPermissions,
			Message:    "Permission denied when starting Terraform apply",
			Underlying: err,
			Context: map[string]interface{}{
				"command":        fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
				"terraform_path": e.options.TerraformPath,
				"working_dir":    e.options.WorkingDir,
			},
			Suggestions: []string{
				fmt.Sprintf("Check if Terraform binary is executable: chmod +x %s", e.options.TerraformPath),
				fmt.Sprintf("Verify write permissions in working directory: %s", e.options.WorkingDir),
				"Check if the working directory exists and is accessible",
				"Run with appropriate user permissions",
			},
			RecoveryAction: "Fix file permissions and retry",
		}
	}

	if strings.Contains(errStr, "no such file") || strings.Contains(errStr, "not found") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeTerraformNotFound,
			Message:    "Terraform binary not found when starting apply",
			Underlying: err,
			Context: map[string]interface{}{
				"terraform_path": e.options.TerraformPath,
				"working_dir":    e.options.WorkingDir,
			},
			Suggestions: []string{
				"Install Terraform from https://www.terraform.io/downloads.html",
				"Verify Terraform is in your PATH environment variable",
				"Use --terraform-path flag to specify the correct path",
				"Check if the specified Terraform path is correct",
			},
			RecoveryAction: "Install Terraform or specify correct path",
		}
	}

	// Generic apply start error
	return errors.NewApplyFailedError(
		fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
		0,
		"",
		err,
	).WithContext("stage", "command_start").
		WithSuggestion("Check Terraform installation and permissions").
		WithSuggestion("Verify working directory is accessible").
		WithRecoveryAction("Fix system configuration and retry")
}

// enhanceApplyTimeoutError enhances apply timeout errors with specific recovery suggestions
func (e *DefaultExecutor) enhanceApplyTimeoutError(cmdArgs []string, output string) error {
	// Analyze output to provide more specific suggestions
	outputLower := strings.ToLower(output)
	suggestions := []string{
		"Increase the timeout using --timeout flag",
		"Check network connectivity to provider endpoints",
		"Consider applying resources in smaller batches",
	}

	if strings.Contains(outputLower, "creating") || strings.Contains(outputLower, "provisioning") {
		suggestions = append(suggestions,
			"Resource creation may be slow - check provider service status",
			"Some resources (like RDS instances) can take 10+ minutes to create",
			"Consider using smaller resource configurations for faster provisioning",
		)
	}

	if strings.Contains(outputLower, "destroying") || strings.Contains(outputLower, "deleting") {
		suggestions = append(suggestions,
			"Resource deletion may be slow - check for dependencies",
			"Some resources may have protection or deletion policies",
			"Check if resources are being used by other services",
		)
	}

	if strings.Contains(outputLower, "modifying") || strings.Contains(outputLower, "updating") {
		suggestions = append(suggestions,
			"Resource updates may require recreation",
			"Check if the update requires downtime",
			"Some updates may trigger cascading changes",
		)
	}

	if strings.Contains(outputLower, "aws") || strings.Contains(outputLower, "amazon") {
		suggestions = append(suggestions,
			"Check AWS service health and region status",
			"Verify AWS credentials and region configuration",
			"Consider using a different AWS region if there are service issues",
		)
	}

	if strings.Contains(outputLower, "network") || strings.Contains(outputLower, "connection") {
		suggestions = append(suggestions,
			"Check internet connectivity and DNS resolution",
			"Verify firewall and proxy settings",
			"Consider using a more stable network connection",
		)
	}

	return &errors.StrataError{
		Code:    errors.ErrorCodeApplyTimeout,
		Message: fmt.Sprintf("Terraform apply timed out after %s", e.options.Timeout.String()),
		Context: map[string]interface{}{
			"timeout": e.options.Timeout.String(),
			"command": fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			"output":  truncateOutput(output, 500),
		},
		Suggestions:    suggestions,
		RecoveryAction: "Increase timeout, check connectivity, or break down the operation",
	}
}

// enhanceApplyFailedError enhances apply failure errors with specific recovery suggestions
func (e *DefaultExecutor) enhanceApplyFailedError(cmdArgs []string, exitCode int, output string, err error) error {
	outputLower := strings.ToLower(output)
	suggestions := []string{
		"Review the error output for specific failure reasons",
		"Check provider permissions and quotas",
	}
	recoveryAction := "Review errors, fix issues, and retry the apply"

	// Analyze output for specific error patterns
	if strings.Contains(outputLower, "authentication") || strings.Contains(outputLower, "credentials") {
		suggestions = append(suggestions,
			"Check provider authentication credentials",
			"Verify environment variables or credential files",
			"Ensure credentials have not expired",
			"Check if MFA or temporary credentials are required",
		)
		recoveryAction = "Fix authentication credentials and retry"
	}

	if strings.Contains(outputLower, "permission") || strings.Contains(outputLower, "access denied") || strings.Contains(outputLower, "unauthorized") {
		suggestions = append(suggestions,
			"Check provider permissions and IAM policies",
			"Verify the credentials have sufficient permissions for all operations",
			"Review resource-specific access requirements",
			"Check if additional permissions are needed for the specific resources",
		)
		recoveryAction = "Fix permissions and retry"
	}

	if strings.Contains(outputLower, "quota") || strings.Contains(outputLower, "limit") || strings.Contains(outputLower, "exceeded") {
		suggestions = append(suggestions,
			"Check provider service quotas and limits",
			"Request quota increases if needed",
			"Consider using different resource configurations",
			"Check if resources can be created in different regions",
		)
		recoveryAction = "Address quota limits and retry"
	}

	if strings.Contains(outputLower, "conflict") || strings.Contains(outputLower, "already exists") {
		suggestions = append(suggestions,
			"Check if resources already exist with the same name",
			"Consider importing existing resources into Terraform state",
			"Use unique resource names or identifiers",
			"Check for naming conflicts with existing infrastructure",
		)
		recoveryAction = "Resolve naming conflicts and retry"
	}

	if strings.Contains(outputLower, "dependency") || strings.Contains(outputLower, "depends on") {
		suggestions = append(suggestions,
			"Check resource dependencies and creation order",
			"Verify that dependent resources exist and are accessible",
			"Review resource references and data sources",
			"Consider explicit depends_on declarations",
		)
		recoveryAction = "Fix resource dependencies and retry"
	}

	if strings.Contains(outputLower, "timeout") && !strings.Contains(outputLower, "timed out after") {
		suggestions = append(suggestions,
			"Resource operations may be taking longer than expected",
			"Check provider service status and performance",
			"Consider increasing resource-specific timeouts",
			"Verify network connectivity to provider services",
		)
		recoveryAction = "Check service status and increase timeouts if needed"
	}

	if strings.Contains(outputLower, "state") && (strings.Contains(outputLower, "lock") || strings.Contains(outputLower, "locked")) {
		suggestions = append(suggestions,
			"Another Terraform operation may be in progress",
			"Check if state is locked by another process",
			"Wait for other operations to complete",
			"Use 'terraform force-unlock' only if you're certain it's safe",
		)
		recoveryAction = "Wait for state lock release or coordinate with team"
	}

	if strings.Contains(outputLower, "validation") || strings.Contains(outputLower, "invalid") {
		suggestions = append(suggestions,
			"Check resource configuration values",
			"Verify that all required fields are set correctly",
			"Review provider documentation for valid values",
			"Check for typos in resource configurations",
		)
		recoveryAction = "Fix configuration validation errors and retry"
	}

	if strings.Contains(outputLower, "network") || strings.Contains(outputLower, "connection") {
		suggestions = append(suggestions,
			"Check network connectivity to provider endpoints",
			"Verify DNS resolution for provider services",
			"Check firewall and proxy settings",
			"Consider network timeouts and retry policies",
		)
		recoveryAction = "Fix network connectivity issues and retry"
	}

	// Provider-specific suggestions
	if strings.Contains(outputLower, "aws") {
		suggestions = append(suggestions,
			"Check AWS service health dashboard",
			"Verify AWS region and availability zone settings",
			"Check AWS resource limits and service quotas",
		)
	}

	if strings.Contains(outputLower, "azure") {
		suggestions = append(suggestions,
			"Check Azure service health status",
			"Verify Azure subscription and resource group settings",
			"Check Azure resource limits and quotas",
		)
	}

	if strings.Contains(outputLower, "gcp") || strings.Contains(outputLower, "google") {
		suggestions = append(suggestions,
			"Check Google Cloud service status",
			"Verify GCP project and region settings",
			"Check GCP resource quotas and limits",
		)
	}

	return &errors.StrataError{
		Code:       errors.ErrorCodeApplyFailed,
		Message:    "Terraform apply execution failed",
		Underlying: err,
		Context: map[string]interface{}{
			"command":   fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			"exit_code": exitCode,
			"output":    truncateOutput(output, 1000),
		},
		Suggestions:    suggestions,
		RecoveryAction: recoveryAction,
	}
}
