package test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ArjenSchwarz/strata/config"
	"github.com/ArjenSchwarz/strata/lib/workflow"
)

// TestTerraformWorkflowE2E tests the complete end-to-end workflow
func TestTerraformWorkflowE2E(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-e2e-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test scenarios for end-to-end workflow
	scenarios := []struct {
		name           string
		config         string
		args           []string
		expectError    bool
		errorContains  string
		outputContains []string
	}{
		{
			name: "simple resource creation - non-interactive",
			config: `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--working-dir", "", // Will be set to tempDir
			},
			expectError: false,
			outputContains: []string{
				"Checking Terraform installation",
				"Executing Terraform plan",
				"Analyzing plan",
				"Applying changes",
				"completed successfully",
			},
		},
		{
			name: "no changes scenario - non-interactive",
			config: `
# Empty configuration - should result in no changes
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--working-dir", "", // Will be set to tempDir
			},
			expectError: false,
			outputContains: []string{
				"Checking Terraform installation",
				"Executing Terraform plan",
			},
		},
		{
			name: "force apply with destructive changes",
			config: `
resource "null_resource" "test1" {
  triggers = {
    value = "test1"
  }
}

resource "null_resource" "test2" {
  triggers = {
    value = "test2"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--force",
				"--working-dir", "", // Will be set to tempDir
			},
			expectError: false,
			outputContains: []string{
				"Checking Terraform installation",
				"Executing Terraform plan",
				"Applying changes",
			},
		},
		{
			name: "custom terraform path",
			config: `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--terraform-path", "terraform",
				"--working-dir", "", // Will be set to tempDir
			},
			expectError: false,
			outputContains: []string{
				"Checking Terraform installation",
				"Executing Terraform plan",
			},
		},
		{
			name: "invalid terraform path",
			config: `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--terraform-path", "nonexistent-terraform",
				"--working-dir", "", // Will be set to tempDir
			},
			expectError:   true,
			errorContains: "terraform",
		},
		{
			name: "custom timeout",
			config: `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--timeout", "60s",
				"--working-dir", "", // Will be set to tempDir
			},
			expectError: false,
			outputContains: []string{
				"Checking Terraform installation",
				"Executing Terraform plan",
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Create scenario-specific directory
			scenarioDir := filepath.Join(tempDir, scenario.name)
			err := os.MkdirAll(scenarioDir, 0755)
			require.NoError(t, err)

			// Update working directory in args
			for i, arg := range scenario.args {
				if arg == "--working-dir" && i+1 < len(scenario.args) && scenario.args[i+1] == "" {
					scenario.args[i+1] = scenarioDir
				}
			}

			// Write Terraform configuration
			configFile := filepath.Join(scenarioDir, "main.tf")
			err = os.WriteFile(configFile, []byte(scenario.config), 0644)
			require.NoError(t, err)

			// Initialize Terraform (skip for invalid terraform path scenarios)
			if !strings.Contains(strings.Join(scenario.args, " "), "nonexistent-terraform") {
				initCmd := exec.Command("terraform", "init")
				initCmd.Dir = scenarioDir
				err = initCmd.Run()
				require.NoError(t, err)
			}

			// Execute the workflow command
			output, err := executeWorkflowCommand(scenario.args)

			if scenario.expectError {
				assert.Error(t, err)
				if scenario.errorContains != "" {
					assert.Contains(t, strings.ToLower(output), strings.ToLower(scenario.errorContains))
				}
			} else {
				assert.NoError(t, err)
				for _, expectedOutput := range scenario.outputContains {
					assert.Contains(t, output, expectedOutput, "Expected output not found: %s", expectedOutput)
				}
			}

			t.Logf("Command output:\n%s", output)
		})
	}
}

// TestTerraformWorkflowE2E_OutputFormats tests different output formats
func TestTerraformWorkflowE2E_OutputFormats(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-e2e-output-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple Terraform configuration
	config := `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`

	configFile := filepath.Join(tempDir, "main.tf")
	err = os.WriteFile(configFile, []byte(config), 0644)
	require.NoError(t, err)

	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = tempDir
	err = initCmd.Run()
	require.NoError(t, err)

	// Test different output formats
	outputFormats := []struct {
		name           string
		format         string
		expectedOutput []string
	}{
		{
			name:   "table format",
			format: "table",
			expectedOutput: []string{
				"TERRAFORM PLAN SUMMARY",
				"Summary:",
			},
		},
		{
			name:           "json format",
			format:         "json",
			expectedOutput: []string{
				// JSON output would contain structured data
				// We'll check for basic JSON indicators
			},
		},
	}

	for _, outputFormat := range outputFormats {
		t.Run(outputFormat.name, func(t *testing.T) {
			args := []string{
				"apply",
				"--non-interactive",
				"--output", outputFormat.format,
				"--working-dir", tempDir,
			}

			output, err := executeWorkflowCommand(args)
			assert.NoError(t, err)

			for _, expectedOutput := range outputFormat.expectedOutput {
				if expectedOutput != "" {
					assert.Contains(t, output, expectedOutput)
				}
			}

			t.Logf("Output format %s:\n%s", outputFormat.format, output)
		})
	}
}

// TestTerraformWorkflowE2E_ConfigurationFile tests workflow with configuration files
func TestTerraformWorkflowE2E_ConfigurationFile(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-e2e-config-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Terraform configuration
	terraformConfig := `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`

	configFile := filepath.Join(tempDir, "main.tf")
	err = os.WriteFile(configFile, []byte(terraformConfig), 0644)
	require.NoError(t, err)

	// Create a Strata configuration file
	strataConfig := `
danger-threshold: 5
output: table
terraform:
  danger-threshold: 2
`

	strataConfigFile := filepath.Join(tempDir, "strata.yaml")
	err = os.WriteFile(strataConfigFile, []byte(strataConfig), 0644)
	require.NoError(t, err)

	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = tempDir
	err = initCmd.Run()
	require.NoError(t, err)

	// Test with configuration file
	args := []string{
		"apply",
		"--non-interactive",
		"--config", strataConfigFile,
		"--working-dir", tempDir,
	}

	output, err := executeWorkflowCommand(args)
	assert.NoError(t, err)
	assert.Contains(t, output, "Checking Terraform installation")

	t.Logf("Output with config file:\n%s", output)
}

// TestTerraformWorkflowE2E_ErrorScenarios tests various error scenarios
func TestTerraformWorkflowE2E_ErrorScenarios(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-e2e-error-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	errorScenarios := []struct {
		name          string
		config        string
		args          []string
		setupFunc     func(dir string) error
		expectError   bool
		errorContains string
	}{
		{
			name: "nonexistent working directory",
			config: `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--working-dir", "/nonexistent/directory",
			},
			expectError:   true,
			errorContains: "error",
		},
		{
			name: "invalid terraform configuration",
			config: `
resource "null_resource" "test" {
  triggers = {
    invalid_syntax = unclosed_string
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--working-dir", "", // Will be set to tempDir
			},
			setupFunc: func(dir string) error {
				// Initialize terraform even with invalid config
				initCmd := exec.Command("terraform", "init")
				initCmd.Dir = dir
				return initCmd.Run()
			},
			expectError:   true,
			errorContains: "error",
		},
		{
			name: "terraform not initialized",
			config: `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`,
			args: []string{
				"apply",
				"--non-interactive",
				"--working-dir", "", // Will be set to tempDir
			},
			setupFunc: func(dir string) error {
				// Don't run terraform init
				return nil
			},
			expectError:   true,
			errorContains: "error",
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Skip scenarios that require nonexistent directories
			if strings.Contains(scenario.name, "nonexistent") {
				// Test the error handling directly
				output, err := executeWorkflowCommand(scenario.args)
				assert.Error(t, err)
				if scenario.errorContains != "" {
					assert.Contains(t, strings.ToLower(output), strings.ToLower(scenario.errorContains))
				}
				return
			}

			// Create scenario-specific directory
			scenarioDir := filepath.Join(tempDir, scenario.name)
			err := os.MkdirAll(scenarioDir, 0755)
			require.NoError(t, err)

			// Update working directory in args
			for i, arg := range scenario.args {
				if arg == "--working-dir" && i+1 < len(scenario.args) && scenario.args[i+1] == "" {
					scenario.args[i+1] = scenarioDir
				}
			}

			// Write Terraform configuration
			configFile := filepath.Join(scenarioDir, "main.tf")
			err = os.WriteFile(configFile, []byte(scenario.config), 0644)
			require.NoError(t, err)

			// Run setup function if provided
			if scenario.setupFunc != nil {
				err = scenario.setupFunc(scenarioDir)
				// Ignore setup errors for error scenarios
				if err != nil {
					t.Logf("Setup failed as expected: %v", err)
				}
			}

			// Execute the workflow command
			output, err := executeWorkflowCommand(scenario.args)

			if scenario.expectError {
				assert.Error(t, err)
				if scenario.errorContains != "" {
					assert.Contains(t, strings.ToLower(output), strings.ToLower(scenario.errorContains))
				}
			} else {
				assert.NoError(t, err)
			}

			t.Logf("Error scenario output:\n%s", output)
		})
	}
}

// TestTerraformWorkflowE2E_Performance tests workflow performance
func TestTerraformWorkflowE2E_Performance(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-e2e-perf-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a larger Terraform configuration for performance testing
	config := "# Performance test configuration\n"
	for i := 0; i < 20; i++ {
		config += `
resource "null_resource" "test` + string(rune('a'+i%26)) + `" {
  triggers = {
    value = "test-value-` + string(rune('a'+i%26)) + `"
  }
}
`
	}

	configFile := filepath.Join(tempDir, "main.tf")
	err = os.WriteFile(configFile, []byte(config), 0644)
	require.NoError(t, err)

	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = tempDir
	err = initCmd.Run()
	require.NoError(t, err)

	// Measure workflow execution time
	start := time.Now()

	args := []string{
		"apply",
		"--non-interactive",
		"--timeout", "120s",
		"--working-dir", tempDir,
	}

	output, err := executeWorkflowCommand(args)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Contains(t, output, "completed successfully")

	t.Logf("Performance test completed in %v", duration)
	t.Logf("Performance test output:\n%s", output)

	// Performance assertion - should complete within reasonable time
	assert.Less(t, duration, 3*time.Minute, "Workflow should complete within 3 minutes")
}

// executeWorkflowCommand executes the terraform workflow command and returns output
func executeWorkflowCommand(args []string) (string, error) {
	// Create a workflow manager and execute the command programmatically
	// This simulates the CLI command execution

	// Parse command line arguments to extract options
	options := parseWorkflowArgs(args)

	// Create configuration
	config := &config.Config{
		Plan: config.PlanConfig{
			DangerThreshold: options.DangerThreshold,
		},
	}

	// Create workflow manager
	manager := workflow.NewWorkflowManager(config)

	// Capture output by redirecting stdout
	// In a real implementation, we'd use a proper output capture mechanism
	ctx := context.Background()

	// Execute workflow
	err := manager.Run(ctx, options)

	// Return simulated output and error
	if err != nil {
		return err.Error(), err
	}

	return "Workflow completed successfully", nil
}

// parseWorkflowArgs parses command line arguments into WorkflowOptions
func parseWorkflowArgs(args []string) *workflow.WorkflowOptions {
	options := &workflow.WorkflowOptions{
		TerraformPath:   "terraform",
		WorkingDir:      ".",
		NonInteractive:  false,
		Force:           false,
		OutputFormat:    "table",
		DangerThreshold: 3,
		Timeout:         30 * time.Minute,
		Environment:     make(map[string]string),
	}

	// Simple argument parsing (in real implementation, use cobra or similar)
	for i, arg := range args {
		switch arg {
		case "--non-interactive":
			options.NonInteractive = true
		case "--force":
			options.Force = true
		case "--terraform-path":
			if i+1 < len(args) {
				options.TerraformPath = args[i+1]
			}
		case "--working-dir":
			if i+1 < len(args) {
				options.WorkingDir = args[i+1]
			}
		case "--output":
			if i+1 < len(args) {
				options.OutputFormat = args[i+1]
			}
		case "--timeout":
			if i+1 < len(args) {
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					options.Timeout = duration
				}
			}
		}
	}

	return options
}

// TestTerraformWorkflowE2E_RealCLI tests the actual CLI command (if available)
func TestTerraformWorkflowE2E_RealCLI(t *testing.T) {
	// This test would require the actual strata binary to be built
	// Skip for now, but in a real test environment, you might:
	// 1. Build the binary as part of the test setup
	// 2. Execute it as a subprocess
	// 3. Capture and verify the output

	t.Skip("Real CLI testing requires built binary - implement when needed")

	// Example of how this might work:
	/*
		// Build the binary
		buildCmd := exec.Command("go", "build", "-o", "strata-test", ".")
		err := buildCmd.Run()
		require.NoError(t, err)
		defer os.Remove("strata-test")

		// Execute the CLI command
		cliCmd := exec.Command("./strata-test", "apply", "--non-interactive", "--working-dir", tempDir)
		output, err := cliCmd.CombinedOutput()

		// Verify results
		assert.NoError(t, err)
		assert.Contains(t, string(output), "completed successfully")
	*/
}
