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
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", errors.NewPlanFailedError(
			fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			0,
			"",
			err,
		).WithContext("stage", "command_start")
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
		return "", fmt.Errorf("error streaming output: %w", streamErr)
	}

	if cmdErr != nil {
		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.NewPlanTimeoutError(e.options.Timeout.String()).
				WithContext("command", fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " "))).
				WithContext("output", outputBuffer.String())
		}
		return "", errors.NewPlanFailedError(
			fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			cmd.ProcessState.ExitCode(),
			outputBuffer.String(),
			cmdErr,
		)
	}

	// Verify the plan file was created
	if _, err := os.Stat(planFile); os.IsNotExist(err) {
		return "", &errors.StrataError{
			Code:    errors.ErrorCodePlanFileNotCreated,
			Message: fmt.Sprintf("Plan file was not created: %s", planFile),
			Context: map[string]interface{}{
				"plan_file":   planFile,
				"working_dir": e.options.WorkingDir,
			},
			Suggestions: []string{
				"Check if Terraform has write permissions in the working directory",
				"Verify there's sufficient disk space",
				"Check for filesystem errors",
			},
			RecoveryAction: "Ensure write permissions and sufficient disk space",
		}
	}

	fmt.Println("Plan generated successfully")
	return planFile, nil
}

// Apply executes terraform apply with the given plan file
func (e *DefaultExecutor) Apply(ctx context.Context, planFile string, args []string) error {
	fmt.Println("Applying Terraform changes...")

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
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return errors.NewApplyFailedError(
			fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			0,
			"",
			err,
		).WithContext("stage", "command_start")
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
		return fmt.Errorf("error streaming output: %w", streamErr)
	}

	if cmdErr != nil {
		// Check for timeout
		if ctx.Err() == context.DeadlineExceeded {
			return &errors.StrataError{
				Code:    errors.ErrorCodeApplyTimeout,
				Message: fmt.Sprintf("Terraform apply timed out after %s", e.options.Timeout.String()),
				Context: map[string]interface{}{
					"timeout": e.options.Timeout.String(),
					"command": fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
					"output":  outputBuffer.String(),
				},
				Suggestions: []string{
					"Increase the timeout using --timeout flag",
					"Check for network connectivity issues",
					"Consider applying resources in smaller batches",
				},
				RecoveryAction: "Increase timeout or break down the operation",
			}
		}
		return errors.NewApplyFailedError(
			fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
			cmd.ProcessState.ExitCode(),
			outputBuffer.String(),
			cmdErr,
		)
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
