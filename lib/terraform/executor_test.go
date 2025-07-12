package terraform

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultExecutor_CleanupTempFile(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() string
		filePath string
		wantErr  bool
	}{
		{
			name: "cleanup existing file",
			setup: func() string {
				tmpFile, err := os.CreateTemp("", "test-cleanup-*.tmp")
				require.NoError(t, err)
				tmpFile.Close()
				return tmpFile.Name()
			},
			wantErr: false,
		},
		{
			name:     "cleanup non-existent file",
			filePath: "/tmp/non-existent-file.tmp",
			wantErr:  false, // Should not error for non-existent files
		},
		{
			name:     "cleanup empty path",
			filePath: "",
			wantErr:  false, // Should handle empty paths gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &DefaultExecutor{
				options: &ExecutorOptions{
					TerraformPath: "terraform",
					WorkingDir:    ".",
					Timeout:       30 * time.Minute,
					Environment:   make(map[string]string),
				},
			}

			filePath := tt.filePath
			if tt.setup != nil {
				filePath = tt.setup()
			}

			err := executor.cleanupTempFile(filePath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify file is actually removed if it existed
			if filePath != "" && tt.setup != nil {
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err), "File should be removed")
			}
		})
	}
}

func TestDefaultExecutor_ErrorRecoveryHelpers(t *testing.T) {
	executor := &DefaultExecutor{
		options: &ExecutorOptions{
			TerraformPath: "terraform",
			WorkingDir:    ".",
			Timeout:       30 * time.Minute,
			Environment:   make(map[string]string),
		},
	}

	t.Run("wrapPipeError", func(t *testing.T) {
		originalErr := assert.AnError
		err := executor.wrapPipeError("stdout", originalErr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stdout pipe")
		// The error message should contain information about the pipe failure
	})

	t.Run("enhancePlanStartError with permission denied", func(t *testing.T) {
		cmdArgs := []string{"plan", "-out=test.tfplan"}
		originalErr := &os.PathError{Op: "exec", Path: "terraform", Err: os.ErrPermission}

		err := executor.enhancePlanStartError(cmdArgs, originalErr)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Permission denied")
	})

	t.Run("enhancePlanStartError with not found", func(t *testing.T) {
		cmdArgs := []string{"plan", "-out=test.tfplan"}
		originalErr := &os.PathError{Op: "exec", Path: "terraform", Err: os.ErrNotExist}

		err := executor.enhancePlanStartError(cmdArgs, originalErr)

		assert.Error(t, err)
		// The error should be enhanced with recovery suggestions
		assert.Contains(t, err.Error(), "Terraform")
	})

	t.Run("truncateOutput", func(t *testing.T) {
		longOutput := "This is a very long output that should be truncated when it exceeds the maximum length limit"

		truncated := truncateOutput(longOutput, 20)
		expectedLen := 20 + len("... (truncated)") // 20 + 15 = 35
		assert.Len(t, truncated, expectedLen)
		assert.Contains(t, truncated, "... (truncated)")

		shortOutput := "Short"
		notTruncated := truncateOutput(shortOutput, 20)
		assert.Equal(t, shortOutput, notTruncated)
	})
}

func TestDefaultExecutor_EnhancedErrorAnalysis(t *testing.T) {
	executor := &DefaultExecutor{
		options: &ExecutorOptions{
			TerraformPath: "terraform",
			WorkingDir:    ".",
			Timeout:       30 * time.Minute,
			Environment:   make(map[string]string),
		},
	}

	t.Run("enhancePlanFailedError with authentication error", func(t *testing.T) {
		cmdArgs := []string{"plan", "-out=test.tfplan"}
		output := "Error: authentication failed - invalid credentials"

		err := executor.enhancePlanFailedError(cmdArgs, 1, output, assert.AnError)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terraform plan execution failed")
		// Should contain authentication-specific suggestions
		// We can't easily test the suggestions without exposing them, but we can verify the error is enhanced
	})

	t.Run("enhancePlanFailedError with permission error", func(t *testing.T) {
		cmdArgs := []string{"plan", "-out=test.tfplan"}
		output := "Error: access denied - insufficient permissions"

		err := executor.enhancePlanFailedError(cmdArgs, 1, output, assert.AnError)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terraform plan execution failed")
	})

	t.Run("enhanceApplyFailedError with quota error", func(t *testing.T) {
		cmdArgs := []string{"apply", "test.tfplan"}
		output := "Error: quota exceeded - cannot create more instances"

		err := executor.enhanceApplyFailedError(cmdArgs, 1, output, assert.AnError)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Terraform apply execution failed")
	})
}
