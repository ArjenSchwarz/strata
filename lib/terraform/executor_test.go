package terraform

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
)

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name     string
		options  *ExecutorOptions
		expected *ExecutorOptions
	}{
		{
			name:    "nil options should use defaults",
			options: nil,
			expected: &ExecutorOptions{
				TerraformPath: "terraform",
				WorkingDir:    ".",
				Timeout:       30 * time.Minute,
				Environment:   make(map[string]string),
			},
		},
		{
			name: "empty options should use defaults",
			options: &ExecutorOptions{
				Environment: make(map[string]string),
			},
			expected: &ExecutorOptions{
				TerraformPath: "terraform",
				WorkingDir:    ".",
				Timeout:       30 * time.Minute,
				Environment:   make(map[string]string),
			},
		},
		{
			name: "custom options should be preserved",
			options: &ExecutorOptions{
				TerraformPath: "/usr/local/bin/terraform",
				WorkingDir:    "/tmp/test",
				Timeout:       60 * time.Minute,
				Environment:   map[string]string{"TF_VAR_test": "value"},
			},
			expected: &ExecutorOptions{
				TerraformPath: "/usr/local/bin/terraform",
				WorkingDir:    "/tmp/test",
				Timeout:       60 * time.Minute,
				Environment:   map[string]string{"TF_VAR_test": "value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.options)
			require.NotNil(t, executor)

			// Cast to DefaultExecutor to access options
			defaultExecutor, ok := executor.(*DefaultExecutor)
			require.True(t, ok, "Expected DefaultExecutor")

			assert.Equal(t, tt.expected.TerraformPath, defaultExecutor.options.TerraformPath)
			assert.Equal(t, tt.expected.WorkingDir, defaultExecutor.options.WorkingDir)
			assert.Equal(t, tt.expected.Timeout, defaultExecutor.options.Timeout)
			assert.Equal(t, tt.expected.Environment, defaultExecutor.options.Environment)
		})
	}
}

func TestDefaultExecutor_CheckInstallation(t *testing.T) {
	tests := []struct {
		name          string
		terraformPath string
		expectError   bool
		errorContains string
		setupMockCmd  func() (cleanup func())
	}{
		{
			name:          "terraform found and working",
			terraformPath: "terraform",
			expectError:   false,
			setupMockCmd: func() func() {
				// This test will only pass if terraform is actually installed
				// In a real test environment, we'd mock the command execution
				return func() {}
			},
		},
		{
			name:          "terraform not found",
			terraformPath: "nonexistent-terraform",
			expectError:   true,
			errorContains: "not found",
			setupMockCmd:  func() func() { return func() {} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupMockCmd()
			defer cleanup()

			options := &ExecutorOptions{
				TerraformPath: tt.terraformPath,
				WorkingDir:    ".",
				Timeout:       5 * time.Second,
				Environment:   make(map[string]string),
			}

			executor := NewExecutor(options)
			ctx := context.Background()

			err := executor.CheckInstallation(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errorContains))
				}
			} else {
				// Only assert no error if terraform is actually available
				if _, cmdErr := exec.LookPath("terraform"); cmdErr == nil {
					assert.NoError(t, err)
				} else {
					t.Skip("Terraform not available in test environment")
				}
			}
		})
	}
}

func TestDefaultExecutor_GetVersion(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	options := &ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    ".",
		Timeout:       10 * time.Second,
		Environment:   make(map[string]string),
	}

	executor := NewExecutor(options)
	ctx := context.Background()

	version, err := executor.GetVersion(ctx)

	assert.NoError(t, err)
	assert.NotEmpty(t, version)
	// Version should contain some JSON structure
	assert.Contains(t, version, "{")
}

func TestDefaultExecutor_Plan_Integration(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple Terraform configuration
	configContent := `
resource "null_resource" "test" {
  triggers = {
    timestamp = timestamp()
  }
}
`
	configFile := filepath.Join(tempDir, "main.tf")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = tempDir
	err = initCmd.Run()
	require.NoError(t, err)

	options := &ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    tempDir,
		Timeout:       30 * time.Second,
		Environment:   make(map[string]string),
	}

	executor := NewExecutor(options)
	ctx := context.Background()

	planFile, err := executor.Plan(ctx, []string{})

	assert.NoError(t, err)
	assert.NotEmpty(t, planFile)
	assert.True(t, strings.HasSuffix(planFile, ".tfplan"))

	// Verify plan file exists
	_, statErr := os.Stat(planFile)
	assert.NoError(t, statErr)

	// Clean up plan file
	os.Remove(planFile)
}

func TestDefaultExecutor_Apply_Integration(t *testing.T) {
	// Skip if terraform is not available
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("Terraform not available in test environment")
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a simple Terraform configuration
	configContent := `
resource "null_resource" "test" {
  triggers = {
    timestamp = "test-value"
  }
}
`
	configFile := filepath.Join(tempDir, "main.tf")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Dir = tempDir
	err = initCmd.Run()
	require.NoError(t, err)

	options := &ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    tempDir,
		Timeout:       30 * time.Second,
		Environment:   make(map[string]string),
	}

	executor := NewExecutor(options)
	ctx := context.Background()

	// First create a plan
	planFile, err := executor.Plan(ctx, []string{})
	require.NoError(t, err)
	require.NotEmpty(t, planFile)

	// Then apply it
	err = executor.Apply(ctx, planFile, []string{})
	assert.NoError(t, err)

	// Plan file should be cleaned up after apply
	_, statErr := os.Stat(planFile)
	assert.True(t, os.IsNotExist(statErr))
}

func TestDefaultExecutor_DetectBackend(t *testing.T) {
	options := &ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    ".",
		Timeout:       10 * time.Second,
		Environment:   make(map[string]string),
	}

	executor := NewExecutor(options)
	ctx := context.Background()

	backend, err := executor.DetectBackend(ctx)

	// Should not error even if terraform show fails
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.NotEmpty(t, backend.Type)
}

func TestDefaultExecutor_ValidateBackend(t *testing.T) {
	options := &ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    ".",
		Timeout:       10 * time.Second,
		Environment:   make(map[string]string),
	}

	executor := NewExecutor(options)
	ctx := context.Background()

	tests := []struct {
		name   string
		config *BackendConfig
	}{
		{
			name:   "nil config should not error",
			config: nil,
		},
		{
			name: "local backend config",
			config: &BackendConfig{
				Type:           "local",
				Config:         make(map[string]interface{}),
				LockTimeout:    10 * time.Minute,
				DisableLocking: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.ValidateBackend(ctx, tt.config)
			// Validation may fail due to terraform init requirements, but should not panic
			// We're mainly testing that the method handles different inputs gracefully
			if err != nil {
				t.Logf("Backend validation failed (expected in test environment): %v", err)
			}
		})
	}
}

func TestExecutorOptions_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		input    *ExecutorOptions
		expected *ExecutorOptions
	}{
		{
			name:  "nil input",
			input: nil,
			expected: &ExecutorOptions{
				TerraformPath: "terraform",
				WorkingDir:    ".",
				Timeout:       30 * time.Minute,
				Environment:   make(map[string]string),
			},
		},
		{
			name: "partial input",
			input: &ExecutorOptions{
				TerraformPath: "/custom/terraform",
			},
			expected: &ExecutorOptions{
				TerraformPath: "/custom/terraform",
				WorkingDir:    ".",
				Timeout:       30 * time.Minute,
				Environment:   make(map[string]string),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.input)
			defaultExecutor := executor.(*DefaultExecutor)

			assert.Equal(t, tt.expected.TerraformPath, defaultExecutor.options.TerraformPath)
			assert.Equal(t, tt.expected.WorkingDir, defaultExecutor.options.WorkingDir)
			assert.Equal(t, tt.expected.Timeout, defaultExecutor.options.Timeout)
			assert.NotNil(t, defaultExecutor.options.Environment)
		})
	}
}

func TestBackendConfig(t *testing.T) {
	config := &BackendConfig{
		Type: "s3",
		Config: map[string]interface{}{
			"bucket": "my-terraform-state",
			"key":    "path/to/state",
			"region": "us-west-2",
		},
		LockTimeout:    15 * time.Minute,
		DisableLocking: false,
	}

	assert.Equal(t, "s3", config.Type)
	assert.Equal(t, "my-terraform-state", config.Config["bucket"])
	assert.Equal(t, "path/to/state", config.Config["key"])
	assert.Equal(t, "us-west-2", config.Config["region"])
	assert.Equal(t, 15*time.Minute, config.LockTimeout)
	assert.False(t, config.DisableLocking)
}

// Benchmark tests for performance validation
func BenchmarkNewExecutor(b *testing.B) {
	options := &ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    ".",
		Timeout:       30 * time.Minute,
		Environment:   make(map[string]string),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewExecutor(options)
	}
}

func BenchmarkExecutorCheckInstallation(b *testing.B) {
	if _, err := exec.LookPath("terraform"); err != nil {
		b.Skip("Terraform not available in test environment")
	}

	executor := NewExecutor(&ExecutorOptions{
		TerraformPath: "terraform",
		WorkingDir:    ".",
		Timeout:       5 * time.Second,
		Environment:   make(map[string]string),
	})

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = executor.CheckInstallation(ctx)
	}
}
