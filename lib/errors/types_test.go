package errors

import (
	"testing"
)

func TestStrataError_BasicFunctionality(t *testing.T) {
	err := &StrataError{
		Code:    ErrorCodeTerraformNotFound,
		Message: "Terraform not found",
		Context: map[string]interface{}{
			"path": "/usr/bin/terraform",
		},
		Suggestions: []string{
			"Install Terraform",
			"Check PATH",
		},
		RecoveryAction: "Install Terraform",
	}

	// Test basic error interface
	if err.Error() != "Terraform not found" {
		t.Errorf("Expected 'Terraform not found', got '%s'", err.Error())
	}

	// Test code retrieval
	if err.GetCode() != ErrorCodeTerraformNotFound {
		t.Errorf("Expected ErrorCodeTerraformNotFound, got %s", err.GetCode())
	}

	// Test context retrieval
	context := err.GetContext()
	if context["path"] != "/usr/bin/terraform" {
		t.Errorf("Expected path in context, got %v", context)
	}

	// Test suggestions
	suggestions := err.GetSuggestions()
	if len(suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(suggestions))
	}

	// Test recovery action
	if err.GetRecoveryAction() != "Install Terraform" {
		t.Errorf("Expected 'Install Terraform', got '%s'", err.GetRecoveryAction())
	}
}

func TestStrataError_ErrorClassification(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		isUser   bool
		isSystem bool
		isCrit   bool
	}{
		{
			name:     "User error",
			code:     ErrorCodeInvalidUserInput,
			isUser:   true,
			isSystem: false,
			isCrit:   false,
		},
		{
			name:     "System error",
			code:     ErrorCodeTerraformNotFound,
			isUser:   false,
			isSystem: true,
			isCrit:   false,
		},
		{
			name:     "Critical error",
			code:     ErrorCodeStateCorrupted,
			isUser:   false,
			isSystem: false,
			isCrit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &StrataError{Code: tt.code}

			if err.IsUserError() != tt.isUser {
				t.Errorf("IsUserError() = %v, want %v", err.IsUserError(), tt.isUser)
			}
			if err.IsSystemError() != tt.isSystem {
				t.Errorf("IsSystemError() = %v, want %v", err.IsSystemError(), tt.isSystem)
			}
			if err.IsCritical() != tt.isCrit {
				t.Errorf("IsCritical() = %v, want %v", err.IsCritical(), tt.isCrit)
			}
		})
	}
}

func TestStrataError_FluentInterface(t *testing.T) {
	err := &StrataError{
		Code:    ErrorCodePlanFailed,
		Message: "Plan failed",
	}

	// Test fluent interface
	err = err.WithContext("command", "terraform plan").
		WithSuggestion("Check configuration").
		WithRecoveryAction("Fix and retry")

	// Verify context was added
	if err.GetContext()["command"] != "terraform plan" {
		t.Error("Context not added correctly")
	}

	// Verify suggestion was added
	suggestions := err.GetSuggestions()
	if len(suggestions) != 1 || suggestions[0] != "Check configuration" {
		t.Error("Suggestion not added correctly")
	}

	// Verify recovery action was set
	if err.GetRecoveryAction() != "Fix and retry" {
		t.Error("Recovery action not set correctly")
	}
}

func TestStrataError_UserMessage(t *testing.T) {
	err := &StrataError{
		Code:    ErrorCodeTerraformNotFound,
		Message: "Terraform not found",
		Context: map[string]interface{}{
			"path": "/usr/bin/terraform",
		},
		Suggestions: []string{
			"Install Terraform",
		},
		RecoveryAction: "Install Terraform",
	}

	message := err.FormatUserMessage()

	// Should contain the main error message
	if !contains(message, "Terraform not found") {
		t.Error("User message should contain main error message")
	}

	// Should contain context
	if !contains(message, "path: /usr/bin/terraform") {
		t.Error("User message should contain context information")
	}

	// Should contain suggestions
	if !contains(message, "Install Terraform") {
		t.Error("User message should contain suggestions")
	}

	// Should contain recovery action
	if !contains(message, "Recovery Action") {
		t.Error("User message should contain recovery action")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
