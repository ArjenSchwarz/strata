// Package workflow provides workflow management and automation utilities.
package workflow

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/ArjenSchwarz/strata/lib/errors"
	"github.com/ArjenSchwarz/strata/lib/plan"
	"github.com/ArjenSchwarz/strata/lib/terraform"
)

// DefaultWorkflowManager is the default implementation of WorkflowManager
type DefaultWorkflowManager struct {
	executor terraform.TerraformExecutor
	parser   terraform.TerraformOutputParser
	config   *config.Config
}

// NewWorkflowManager creates a new workflow manager
func NewWorkflowManager(config *config.Config) WorkflowManager {
	executorOptions := &terraform.ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    ".",
		Timeout:       30 * time.Minute,
		Environment:   make(map[string]string),
	}

	return &DefaultWorkflowManager{
		executor: terraform.NewExecutor(executorOptions),
		parser:   terraform.NewOutputParser(),
		config:   config,
	}
}

// Run executes the workflow
func (w *DefaultWorkflowManager) Run(ctx context.Context, options *WorkflowOptions) error {
	startTime := time.Now()

	// Set up cleanup tracking for temporary resources
	var tempResources []string
	defer func() {
		// Clean up any temporary resources on exit
		w.cleanupTempResources(tempResources)
	}()

	// Detect CI/CD environment and adjust behavior
	cicdEnv := w.detectCICDEnvironment()
	if cicdEnv != "" {
		fmt.Printf("🔧 Detected CI/CD environment: %s\n", cicdEnv)
		// Force non-interactive mode in CI/CD environments
		if !options.NonInteractive {
			fmt.Println("🤖 Automatically enabling non-interactive mode for CI/CD")
			options.NonInteractive = true
		}
		// Apply CI/CD specific adjustments
		w.adjustForCICD(cicdEnv, options)
	}

	// Update executor options
	executorOptions := &terraform.ExecutorOptions{
		TerraformPath: options.TerraformPath,
		WorkingDir:    options.WorkingDir,
		Timeout:       options.Timeout,
		Environment:   options.Environment,
	}
	w.executor = terraform.NewExecutor(executorOptions)

	// Step 1: Check Terraform installation
	fmt.Println("Checking Terraform installation...")
	if err := w.executor.CheckInstallation(ctx); err != nil {
		// If it's already a StrataError, return it directly
		if strataErr, ok := err.(*errors.StrataError); ok {
			return strataErr
		}
		// Otherwise wrap it
		return &errors.StrataError{
			Code:       errors.ErrorCodeTerraformNotFound,
			Message:    "Terraform installation check failed",
			Underlying: err,
			Suggestions: []string{
				"Install Terraform from https://www.terraform.io/downloads.html",
				"Ensure Terraform is in your PATH",
			},
			RecoveryAction: "Install or configure Terraform",
		}
	}

	// Step 2: Execute terraform plan
	fmt.Println("Executing Terraform plan...")
	planFile, err := w.executor.Plan(ctx, options.PlanArgs)
	if err != nil {
		// Add plan file to cleanup list if it was created but plan failed
		if planFile != "" {
			tempResources = append(tempResources, planFile)
		}

		// Enhance error with recovery suggestions
		recoveredErr := w.recoverFromError(err, "terraform plan execution")

		// Provide user guidance for interactive sessions
		if !options.NonInteractive {
			w.provideUserGuidance(recoveredErr)
		}

		return recoveredErr
	}

	// Add plan file to cleanup list for later cleanup
	tempResources = append(tempResources, planFile)

	// Step 3: Analyze the plan using existing Strata functionality
	fmt.Println("Analyzing plan...")
	planSummary, err := w.analyzePlan(planFile)
	if err != nil {
		// If it's already a StrataError, return it directly
		if strataErr, ok := err.(*errors.StrataError); ok {
			return strataErr
		}
		// Otherwise wrap it
		return errors.NewPlanAnalysisFailedError(planFile, err)
	}

	// Step 4: Display summary
	if err := w.DisplaySummary(planSummary); err != nil {
		return &errors.StrataError{
			Code:       errors.ErrorCodePlanAnalysisFailed,
			Message:    "Failed to display plan summary",
			Underlying: err,
			Context: map[string]interface{}{
				"plan_file": planFile,
			},
			Suggestions: []string{
				"Check if the plan file is readable",
				"Verify the plan analysis completed successfully",
			},
			RecoveryAction: "Regenerate the plan and retry",
		}
	}

	// Step 5: Check for dangerous changes
	if w.hasDangerousChanges(planSummary, options.DangerThreshold) {
		fmt.Printf("⚠️  WARNING: Detected potentially destructive changes (threshold: %d)\n", options.DangerThreshold)
		if !options.NonInteractive && !options.Force {
			fmt.Println("Please review the changes carefully before proceeding.")
		}
	}

	// Step 6: Determine action
	var action Action
	if options.NonInteractive {
		w.logAuditEvent("NON_INTERACTIVE_MODE", "Workflow running in non-interactive mode", cicdEnv)

		// In non-interactive mode, check for destructive changes
		if w.hasDestructiveChanges(planSummary) {
			destructiveCount := w.countDestructiveChanges(planSummary)
			w.logAuditEvent("DESTRUCTIVE_CHANGES_DETECTED",
				fmt.Sprintf("Found %d destructive changes", destructiveCount), cicdEnv)

			if !options.Force {
				w.logAuditEvent("CANCELLED_NO_FORCE",
					"Cancelled due to destructive changes without --force flag", cicdEnv)
				fmt.Println("❌ Destructive changes detected in non-interactive mode.")
				fmt.Println("Use --force flag to proceed with destructive changes automatically.")
				action = ActionCancel
			} else {
				w.logAuditEvent("FORCED_APPLY",
					"Proceeding with destructive changes due to --force flag", cicdEnv)
				fmt.Println("⚠️  Proceeding with destructive changes due to --force flag.")
				action = ActionApply
			}
		} else {
			// No destructive changes, safe to apply
			w.logAuditEvent("SAFE_APPLY", "No destructive changes detected, proceeding with apply", cicdEnv)
			action = ActionApply
		}
	} else {
		// Interactive mode - prompt user
		w.logAuditEvent("INTERACTIVE_MODE", "Prompting user for action", cicdEnv)
		action, err = w.PromptForAction(planSummary)
		if err != nil {
			// If it's already a StrataError, return it directly
			if strataErr, ok := err.(*errors.StrataError); ok {
				return strataErr
			}
			// Otherwise wrap it
			return errors.NewUserInputFailedError("action selection", err)
		}
		w.logAuditEvent("USER_ACTION", fmt.Sprintf("User selected action: %s", action.String()), cicdEnv)
	}

	// Step 7: Execute action
	switch action {
	case ActionApply:
		fmt.Println("Applying changes...")
		if err := w.executor.Apply(ctx, planFile, options.ApplyArgs); err != nil {
			// Enhance error with recovery suggestions
			recoveredErr := w.recoverFromError(err, "terraform apply execution")

			// In CI/CD environments, provide detailed error information
			if cicdEnv != "" {
				fmt.Printf("❌ Apply failed in %s environment\n", cicdEnv)
				fmt.Printf("Error details: %v\n", recoveredErr)
			}

			// Provide user guidance for interactive sessions
			if !options.NonInteractive {
				w.provideUserGuidance(recoveredErr)
			}

			return recoveredErr
		}
		fmt.Printf("✅ Workflow completed successfully in %v\n", time.Since(startTime))

		// In CI/CD environments, provide additional success information
		if cicdEnv != "" {
			fmt.Printf("🎉 Deployment successful in %s environment\n", cicdEnv)
			w.generateMachineReadableOutput(planSummary, action, cicdEnv)
		}

	case ActionViewDetails:
		// This should be handled in the prompt loop, but if we get here, just display and exit
		fmt.Println("Detailed plan output was displayed. Workflow cancelled.")
		if cicdEnv != "" {
			// In CI/CD, this might indicate a configuration issue
			return &errors.StrataError{
				Code:    errors.ErrorCodeInvalidUserInput,
				Message: "Unexpected view details action in CI/CD environment",
				Context: map[string]interface{}{
					"cicd_env": cicdEnv,
					"action":   "view_details",
				},
				Suggestions: []string{
					"Use --non-interactive flag in CI/CD environments",
					"Configure the workflow for automated execution",
				},
				RecoveryAction: "Use non-interactive mode in CI/CD",
			}
		}

	case ActionCancel:
		fmt.Println("Workflow cancelled by user.")
		if cicdEnv != "" {
			fmt.Printf("🚫 Deployment cancelled in %s environment\n", cicdEnv)
		}
		// Return a specific error for cancellation to allow proper exit code handling
		return errors.NewWorkflowCancelledError("user cancelled the workflow")

	default:
		return &errors.StrataError{
			Code:    errors.ErrorCodeInvalidUserInput,
			Message: fmt.Sprintf("Unknown action: %v", action),
			Context: map[string]interface{}{
				"action": action.String(),
			},
			Suggestions: []string{
				"This is likely a bug in the application",
				"Please report this issue",
			},
			RecoveryAction: "Restart the workflow",
		}
	}

	return nil
}

// PromptForAction prompts the user for action
func (w *DefaultWorkflowManager) PromptForAction(summary *plan.PlanSummary) (Action, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nWhat would you like to do?")
		fmt.Println("  [a] Apply these changes")
		fmt.Println("  [d] View detailed plan output")
		fmt.Println("  [c] Cancel")
		fmt.Print("Enter your choice [a/d/c]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return ActionCancel, errors.NewUserInputFailedError("action selection", err)
		}

		choice := strings.ToLower(strings.TrimSpace(input))
		switch choice {
		case "a", "apply":
			// Check for destructive changes and require explicit confirmation
			if w.hasDestructiveChanges(summary) {
				confirmed, err := w.confirmDestructiveChanges(summary)
				if err != nil {
					return ActionCancel, err
				}
				if !confirmed {
					fmt.Println("Apply cancelled due to destructive changes.")
					continue
				}
			}
			return ActionApply, nil
		case "d", "details", "detail":
			// Display details and continue prompting
			if err := w.DisplayDetails(""); err != nil {
				fmt.Printf("Error displaying details: %v\n", err)
			}
			continue
		case "c", "cancel":
			return ActionCancel, nil
		default:
			fmt.Printf("Invalid choice '%s'. Please enter 'a', 'd', or 'c'.\n", choice)
			continue
		}
	}
}

// confirmDestructiveChanges prompts for explicit confirmation of destructive changes
func (w *DefaultWorkflowManager) confirmDestructiveChanges(summary *plan.PlanSummary) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	// Count destructive changes
	destructiveCount := 0
	var destructiveResources []string
	for _, change := range summary.ResourceChanges {
		if change.IsDestructive {
			destructiveCount++
			destructiveResources = append(destructiveResources, change.Address)
		}
	}

	fmt.Printf("\n⚠️  WARNING: This plan includes %d destructive changes:\n", destructiveCount)
	for _, resource := range destructiveResources {
		fmt.Printf("  - %s\n", resource)
	}

	fmt.Println("\nDestructive changes will permanently delete or replace resources.")
	fmt.Println("This action cannot be undone.")

	for {
		fmt.Print("\nDo you want to proceed with these destructive changes? [yes/no]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, errors.NewUserInputFailedError("destructive changes confirmation", err)
		}

		choice := strings.ToLower(strings.TrimSpace(input))
		switch choice {
		case "yes", "y":
			return true, nil
		case "no", "n":
			return false, nil
		default:
			fmt.Printf("Please enter 'yes' or 'no'.\n")
			continue
		}
	}
}

// hasDestructiveChanges checks if the plan has any destructive changes
func (w *DefaultWorkflowManager) hasDestructiveChanges(summary *plan.PlanSummary) bool {
	for _, change := range summary.ResourceChanges {
		if change.IsDestructive {
			return true
		}
	}
	return false
}

// DisplaySummary displays the plan summary with highlighting for dangerous changes
func (w *DefaultWorkflowManager) DisplaySummary(summary *plan.PlanSummary) error {
	// Display header
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("TERRAFORM PLAN SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	// Use existing formatter to display the summary
	formatter := plan.NewFormatter(w.config)
	if err := formatter.OutputSummary(summary, "table", false); err != nil {
		return err
	}

	// Highlight dangerous changes if present
	if w.hasDestructiveChanges(summary) {
		fmt.Println("\n⚠️  DESTRUCTIVE CHANGES DETECTED:")
		for _, change := range summary.ResourceChanges {
			if change.IsDestructive {
				fmt.Printf("  🔥 %s (%s)\n", change.Address, change.ChangeType)
				if change.IsDangerous && change.DangerReason != "" {
					fmt.Printf("     Reason: %s\n", change.DangerReason)
				}
			}
		}
	}

	// Display summary statistics
	fmt.Printf("\n📊 Summary: %d resources to be changed\n", summary.Statistics.Total)
	if summary.Statistics.ToAdd > 0 {
		fmt.Printf("  ➕ %d to add\n", summary.Statistics.ToAdd)
	}
	if summary.Statistics.ToChange > 0 {
		fmt.Printf("  🔄 %d to modify\n", summary.Statistics.ToChange)
	}
	if summary.Statistics.ToDestroy > 0 {
		fmt.Printf("  ❌ %d to destroy\n", summary.Statistics.ToDestroy)
	}
	if summary.Statistics.Replacements > 0 {
		fmt.Printf("  🔄 %d to replace\n", summary.Statistics.Replacements)
	}

	fmt.Println(strings.Repeat("=", 80))

	return nil
}

// DisplayDetails displays detailed plan output
func (w *DefaultWorkflowManager) DisplayDetails(planOutput string) error {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("DETAILED PLAN OUTPUT")
	fmt.Println(strings.Repeat("=", 80))

	if planOutput == "" {
		fmt.Println("Detailed plan output is not available in this context.")
		fmt.Println("The detailed output was already displayed during plan execution.")
	} else {
		fmt.Println(planOutput)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Press Enter to continue...")

	// Wait for user to press Enter
	reader := bufio.NewReader(os.Stdin)
	_, err := reader.ReadString('\n')
	if err != nil {
		return errors.NewUserInputFailedError("continue prompt", err)
	}

	return nil
}

// analyzePlan analyzes the plan file using existing Strata functionality
func (w *DefaultWorkflowManager) analyzePlan(planFile string) (*plan.PlanSummary, error) {
	// Create parser and load plan
	parser := plan.NewParser(planFile)
	tfPlan, err := parser.LoadPlan()
	if err != nil {
		return nil, errors.NewPlanAnalysisFailedError(planFile, err).
			WithSuggestion("Verify the plan file exists and is readable").
			WithSuggestion("Regenerate the plan file if it's corrupted")
	}

	// Validate plan structure
	if err := parser.ValidateStructure(tfPlan); err != nil {
		return nil, errors.NewInvalidPlanFormatError(planFile, "valid Terraform plan").
			WithContext("validation_error", err.Error()).
			WithSuggestion("Regenerate the plan with the current Terraform version")
	}

	// Create analyzer and generate summary
	analyzer := plan.NewAnalyzer(tfPlan, w.config)
	summary := analyzer.GenerateSummary(planFile)

	// Extract and apply danger information
	w.applyDangerAnalysis(summary, analyzer)

	return summary, nil
}

// applyDangerAnalysis applies danger analysis to the plan summary
func (w *DefaultWorkflowManager) applyDangerAnalysis(summary *plan.PlanSummary, analyzer *plan.Analyzer) {
	// Get destructive change count using the analyzer
	destructiveCount := analyzer.GetDestructiveChangeCount(summary.ResourceChanges)

	// Mark changes as dangerous based on various criteria
	for i := range summary.ResourceChanges {
		change := &summary.ResourceChanges[i]

		// Mark destructive changes as dangerous
		if change.IsDestructive {
			change.IsDangerous = true
			if change.DangerReason == "" {
				change.DangerReason = "Destructive operation"
			}
		}

		// Check for sensitive resources
		if w.isSensitiveResource(change.Type) {
			change.IsDangerous = true
			if change.DangerReason == "" {
				change.DangerReason = "Sensitive resource"
			} else {
				change.DangerReason += " (sensitive)"
			}
		}

		// Check for sensitive properties
		if sensitiveProps := w.getSensitiveProperties(change); len(sensitiveProps) > 0 {
			change.IsDangerous = true
			change.DangerProperties = sensitiveProps
			if change.DangerReason == "" {
				change.DangerReason = "Sensitive properties modified"
			}
		}
	}

	// Update high risk count in statistics
	highRiskCount := 0
	for _, change := range summary.ResourceChanges {
		if change.IsDangerous {
			highRiskCount++
		}
	}
	summary.Statistics.HighRisk = highRiskCount

	fmt.Printf("📊 Analysis complete: %d total changes, %d destructive, %d high-risk\n",
		summary.Statistics.Total, destructiveCount, highRiskCount)
}

// isSensitiveResource checks if a resource type is considered sensitive
func (w *DefaultWorkflowManager) isSensitiveResource(resourceType string) bool {
	if w.config == nil || len(w.config.SensitiveResources) == 0 {
		// Default sensitive resource types
		sensitiveTypes := []string{
			"aws_db_instance",
			"aws_rds_cluster",
			"aws_s3_bucket",
			"aws_iam_role",
			"aws_iam_policy",
			"aws_security_group",
			"aws_vpc",
			"aws_subnet",
		}
		for _, sensitive := range sensitiveTypes {
			if resourceType == sensitive {
				return true
			}
		}
		return false
	}

	// Use configured sensitive resources
	for _, sensitive := range w.config.SensitiveResources {
		if resourceType == sensitive.ResourceType {
			return true
		}
	}
	return false
}

// getSensitiveProperties returns sensitive properties that are being modified
func (w *DefaultWorkflowManager) getSensitiveProperties(change *plan.ResourceChange) []string {
	var sensitiveProps []string

	if w.config == nil || len(w.config.SensitiveProperties) == 0 {
		// Default sensitive properties
		defaultSensitive := []string{
			"password",
			"secret",
			"key",
			"token",
			"credential",
		}

		for _, attr := range change.ChangeAttributes {
			for _, sensitive := range defaultSensitive {
				if strings.Contains(strings.ToLower(attr), sensitive) {
					sensitiveProps = append(sensitiveProps, attr)
					break
				}
			}
		}
		return sensitiveProps
	}

	// Use configured sensitive properties
	for _, attr := range change.ChangeAttributes {
		for _, sensitive := range w.config.SensitiveProperties {
			if change.Type == sensitive.ResourceType &&
				strings.Contains(strings.ToLower(attr), strings.ToLower(sensitive.Property)) {
				sensitiveProps = append(sensitiveProps, attr)
				break
			}
		}
	}

	return sensitiveProps
}

// detectCICDEnvironment detects common CI/CD environments
func (w *DefaultWorkflowManager) detectCICDEnvironment() string {
	// Check for common CI/CD environment variables
	cicdEnvs := map[string]string{
		"GITHUB_ACTIONS":   "GitHub Actions",
		"GITLAB_CI":        "GitLab CI",
		"JENKINS_URL":      "Jenkins",
		"BUILDKITE":        "Buildkite",
		"CIRCLECI":         "CircleCI",
		"TRAVIS":           "Travis CI",
		"APPVEYOR":         "AppVeyor",
		"AZURE_PIPELINES":  "Azure Pipelines",
		"TEAMCITY_VERSION": "TeamCity",
		"BAMBOO_BUILD_KEY": "Bamboo",
		"TF_BUILD":         "Azure DevOps",
		"CI":               "Generic CI",
	}

	// Check in order of specificity (most specific first)
	for envVar, name := range cicdEnvs {
		if os.Getenv(envVar) != "" {
			return name
		}
	}

	// Check for additional indicators
	if os.Getenv("BUILD_NUMBER") != "" || os.Getenv("BUILD_ID") != "" {
		return "Generic CI (Build detected)"
	}

	return ""
}

// adjustForCICD adjusts workflow behavior for CI/CD environments
func (w *DefaultWorkflowManager) adjustForCICD(cicdEnv string, options *WorkflowOptions) {
	if cicdEnv == "" {
		return
	}

	// Set appropriate exit codes for CI/CD
	// This will be used by the calling code to set proper exit codes

	// Adjust output formatting for CI/CD
	if options.OutputFormat == "table" {
		// In CI/CD, prefer more machine-readable formats
		options.OutputFormat = "json"
		fmt.Println("📊 Switching to JSON output format for CI/CD compatibility")
	}

	// Extend timeout for CI/CD environments (they might be slower)
	if options.Timeout < 45*time.Minute {
		options.Timeout = 45 * time.Minute
		fmt.Println("⏱️  Extended timeout for CI/CD environment")
	}
}

// hasDangerousChanges checks if the plan has dangerous changes above the threshold
func (w *DefaultWorkflowManager) hasDangerousChanges(summary *plan.PlanSummary, threshold int) bool {
	destructiveCount := 0
	for _, change := range summary.ResourceChanges {
		if change.IsDestructive {
			destructiveCount++
		}
	}
	return destructiveCount >= threshold
}

// logAuditEvent logs events for audit trails, especially useful in CI/CD environments
func (w *DefaultWorkflowManager) logAuditEvent(eventType, message, cicdEnv string) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// In CI/CD environments, output structured logs
	if cicdEnv != "" {
		// Output as JSON for machine parsing
		fmt.Printf("AUDIT_LOG: %s | %s | %s\n", timestamp, eventType, message)
	} else {
		// Human-readable format for local development
		fmt.Printf("🔍 [%s] %s: %s\n", timestamp, eventType, message)
	}
}

// countDestructiveChanges counts the number of destructive changes
func (w *DefaultWorkflowManager) countDestructiveChanges(summary *plan.PlanSummary) int {
	count := 0
	for _, change := range summary.ResourceChanges {
		if change.IsDestructive {
			count++
		}
	}
	return count
}

// generateMachineReadableOutput generates machine-readable output for CI/CD systems
func (w *DefaultWorkflowManager) generateMachineReadableOutput(summary *plan.PlanSummary, action Action, cicdEnv string) {
	if cicdEnv == "" {
		return // Only generate for CI/CD environments
	}

	fmt.Println("MACHINE_READABLE_OUTPUT:")
	fmt.Printf("ACTION=%s\n", action.String())
	fmt.Printf("TOTAL_CHANGES=%d\n", summary.Statistics.Total)
	fmt.Printf("DESTRUCTIVE_CHANGES=%d\n", w.countDestructiveChanges(summary))
	fmt.Printf("HIGH_RISK_CHANGES=%d\n", summary.Statistics.HighRisk)
	fmt.Printf("CICD_ENV=%s\n", cicdEnv)
}

// Error recovery and cleanup methods

// cleanupTempResources cleans up temporary resources created during workflow execution
func (w *DefaultWorkflowManager) cleanupTempResources(tempResources []string) {
	if len(tempResources) == 0 {
		return
	}

	fmt.Printf("🧹 Cleaning up %d temporary resources...\n", len(tempResources))

	for _, resource := range tempResources {
		if err := w.cleanupSingleResource(resource); err != nil {
			fmt.Printf("Warning: Failed to cleanup resource %s: %v\n", resource, err)
		}
	}
}

// cleanupSingleResource cleans up a single temporary resource
func (w *DefaultWorkflowManager) cleanupSingleResource(resource string) error {
	// Check if it's a file path
	if strings.HasPrefix(resource, "/") || strings.Contains(resource, ".") {
		// Treat as file path
		if _, err := os.Stat(resource); err == nil {
			if err := os.Remove(resource); err != nil {
				return fmt.Errorf("failed to remove file %s: %w", resource, err)
			}
			fmt.Printf("  ✅ Removed temporary file: %s\n", resource)
		}
		return nil
	}

	// Add other resource types as needed (directories, etc.)
	return nil
}

// recoverFromError attempts to recover from common errors and provide user guidance
func (w *DefaultWorkflowManager) recoverFromError(err error, context string) error {
	if err == nil {
		return nil
	}

	// If it's already a StrataError, enhance it with recovery context
	if strataErr, ok := err.(*errors.StrataError); ok {
		return w.enhanceStrataError(strataErr, context)
	}

	// Convert generic errors to StrataErrors with recovery suggestions
	return w.convertToRecoverableError(err, context)
}

// enhanceStrataError enhances existing StrataErrors with additional recovery context
func (w *DefaultWorkflowManager) enhanceStrataError(strataErr *errors.StrataError, context string) *errors.StrataError {
	// Add context about where the error occurred
	strataErr = strataErr.WithContext("workflow_context", context)

	// Add workflow-specific recovery suggestions based on error type
	switch strataErr.GetCode() {
	case errors.ErrorCodeTerraformNotFound:
		strataErr = strataErr.WithSuggestion("Run 'which terraform' to check if Terraform is installed")
		strataErr = strataErr.WithSuggestion("Consider using a Terraform version manager like tfenv")

	case errors.ErrorCodePlanFailed:
		strataErr = strataErr.WithSuggestion("Try running 'terraform validate' first to check configuration")
		strataErr = strataErr.WithSuggestion("Check if 'terraform init' needs to be run")

	case errors.ErrorCodeApplyFailed:
		strataErr = strataErr.WithSuggestion("Review the plan output to understand what was being applied")
		strataErr = strataErr.WithSuggestion("Check if partial apply occurred - some resources may have been created")
		strataErr = strataErr.WithSuggestion("Consider running 'terraform refresh' to update state")

	case errors.ErrorCodeStateLockTimeout, errors.ErrorCodeStateLockConflict:
		strataErr = strataErr.WithSuggestion("Check if other team members are running Terraform")
		strataErr = strataErr.WithSuggestion("Wait a few minutes and try again")
		strataErr = strataErr.WithSuggestion("Use 'terraform force-unlock' only if you're certain it's safe")

	case errors.ErrorCodeUserInputFailed:
		strataErr = strataErr.WithSuggestion("Use --non-interactive flag for automated environments")
		strataErr = strataErr.WithSuggestion("Ensure you're running in a proper terminal environment")
	}

	return strataErr
}

// convertToRecoverableError converts generic errors to StrataErrors with recovery suggestions
func (w *DefaultWorkflowManager) convertToRecoverableError(err error, context string) *errors.StrataError {
	errStr := strings.ToLower(err.Error())

	// Analyze error message for common patterns
	if strings.Contains(errStr, "permission denied") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeInsufficientPermissions,
			Message:    fmt.Sprintf("Permission error in %s", context),
			Underlying: err,
			Context: map[string]interface{}{
				"workflow_context": context,
			},
			Suggestions: []string{
				"Check file and directory permissions",
				"Ensure you have the necessary access rights",
				"Try running with appropriate user permissions",
			},
			RecoveryAction: "Fix permissions and retry the operation",
		}
	}

	if strings.Contains(errStr, "no space") || strings.Contains(errStr, "disk full") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeDiskSpaceFull,
			Message:    fmt.Sprintf("Disk space error in %s", context),
			Underlying: err,
			Context: map[string]interface{}{
				"workflow_context": context,
			},
			Suggestions: []string{
				"Free up disk space in the working directory",
				"Check disk usage with 'df -h'",
				"Consider using a different directory with more space",
			},
			RecoveryAction: "Free up disk space and retry",
		}
	}

	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return &errors.StrataError{
			Code:       errors.ErrorCodeNetworkUnavailable,
			Message:    fmt.Sprintf("Network error in %s", context),
			Underlying: err,
			Context: map[string]interface{}{
				"workflow_context": context,
			},
			Suggestions: []string{
				"Check internet connectivity",
				"Verify DNS resolution",
				"Check firewall and proxy settings",
				"Try again after a few minutes",
			},
			RecoveryAction: "Fix network connectivity and retry",
		}
	}

	if strings.Contains(errStr, "timeout") {
		return &errors.StrataError{
			Code:       errors.ErrorCodePlanTimeout,
			Message:    fmt.Sprintf("Timeout error in %s", context),
			Underlying: err,
			Context: map[string]interface{}{
				"workflow_context": context,
			},
			Suggestions: []string{
				"Increase timeout using --timeout flag",
				"Check for network or service issues",
				"Consider breaking down the operation into smaller parts",
			},
			RecoveryAction: "Increase timeout or check for underlying issues",
		}
	}

	// Generic error with basic recovery suggestions
	return &errors.StrataError{
		Code:       errors.ErrorCodeSystemResourceExhausted,
		Message:    fmt.Sprintf("Error in %s: %s", context, err.Error()),
		Underlying: err,
		Context: map[string]interface{}{
			"workflow_context": context,
		},
		Suggestions: []string{
			"Check system resources and stability",
			"Try the operation again",
			"Review the error details for specific issues",
		},
		RecoveryAction: "Address the underlying issue and retry",
	}
}

// provideUserGuidance provides interactive guidance to help users recover from errors
func (w *DefaultWorkflowManager) provideUserGuidance(err error) {
	if strataErr, ok := err.(*errors.StrataError); ok {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🔧 ERROR RECOVERY GUIDANCE")
		fmt.Println(strings.Repeat("=", 60))

		fmt.Printf("Error: %s\n", strataErr.Message)

		if len(strataErr.GetSuggestions()) > 0 {
			fmt.Println("\n💡 Suggested actions:")
			for i, suggestion := range strataErr.GetSuggestions() {
				fmt.Printf("  %d. %s\n", i+1, suggestion)
			}
		}

		if strataErr.GetRecoveryAction() != "" {
			fmt.Printf("\n🔧 Recovery Action: %s\n", strataErr.GetRecoveryAction())
		}

		// Provide context-specific guidance
		if context, exists := strataErr.GetContext()["workflow_context"]; exists {
			fmt.Printf("\n📍 Context: Error occurred during %s\n", context)
		}

		fmt.Println(strings.Repeat("=", 60))
	}
}
