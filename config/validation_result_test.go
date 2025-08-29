package config

import (
	"os"
	"testing"
)

func TestValidationResult(t *testing.T) {
	t.Parallel()
	t.Run("new validation result", func(t *testing.T) {
		t.Parallel()
		result := &ValidationResult{Valid: true}

		if result.HasErrors() {
			t.Errorf("HasErrors() = true, want false for new result")
		}

		if len(result.Errors) != 0 {
			t.Errorf("len(Errors) = %d, want 0", len(result.Errors))
		}
	})

	t.Run("add error", func(t *testing.T) {
		t.Parallel()
		result := &ValidationResult{Valid: true}
		testErr := &FileOutputError{
			Type:    "validation",
			Code:    "PATH_TRAVERSAL",
			Path:    "/test/path",
			Message: "test error",
		}

		result.AddError(testErr)

		if !result.HasErrors() {
			t.Errorf("HasErrors() = false, want true after adding error")
		}

		if result.Valid {
			t.Errorf("Valid = true, want false after adding error")
		}

		if len(result.Errors) != 1 {
			t.Errorf("len(Errors) = %d, want 1", len(result.Errors))
		}
	})

	t.Run("add warning and info", func(t *testing.T) {
		t.Parallel()
		result := &ValidationResult{Valid: true}

		result.AddWarning("test warning")
		result.AddInfo("test info")

		if len(result.Warnings) != 1 {
			t.Errorf("len(Warnings) = %d, want 1", len(result.Warnings))
		}

		if len(result.Infos) != 1 {
			t.Errorf("len(Infos) = %d, want 1", len(result.Infos))
		}

		if result.Warnings[0] != "test warning" {
			t.Errorf("Warnings[0] = %v, want 'test warning'", result.Warnings[0])
		}

		if result.Infos[0] != "test info" {
			t.Errorf("Infos[0] = %v, want 'test info'", result.Infos[0])
		}
	})
}

func TestFileOutputError(t *testing.T) {
	t.Parallel()
	t.Run("error without cause", func(t *testing.T) {
		t.Parallel()
		err := &FileOutputError{
			Type:    "validation",
			Code:    "PATH_TRAVERSAL",
			Path:    "/test/path",
			Message: "test error",
		}

		expected := "validation error [PATH_TRAVERSAL] for /test/path: test error"
		if err.Error() != expected {
			t.Errorf("Error() = %v, want %v", err.Error(), expected)
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		t.Parallel()
		cause := os.ErrNotExist
		err := &FileOutputError{
			Type:    "permission",
			Code:    "PERMISSION_DENIED",
			Path:    "/test/path",
			Message: "cannot access file",
			Cause:   cause,
		}

		expected := "permission error [PERMISSION_DENIED] for /test/path: cannot access file (caused by: file does not exist)"
		if err.Error() != expected {
			t.Errorf("Error() = %v, want %v", err.Error(), expected)
		}
	})

	t.Run("error code access", func(t *testing.T) {
		t.Parallel()
		err := &FileOutputError{
			Type:    "format",
			Code:    "UNSUPPORTED_FORMAT",
			Path:    "/test/path",
			Format:  "xml",
			Message: "unsupported format",
		}

		if err.Code != "UNSUPPORTED_FORMAT" {
			t.Errorf("Code = %v, want UNSUPPORTED_FORMAT", err.Code)
		}
	})
}

// Security Tests for Path Traversal Prevention
