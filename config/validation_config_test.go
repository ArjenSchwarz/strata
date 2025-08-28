package config

import (
	"os"
	"strings"
	"testing"
)

func TestConfig_ResolvePlaceholders(t *testing.T) {
	config := &Config{}

	// Mock environment variables for testing
	originalRegion := os.Getenv("AWS_REGION")
	originalAccountID := os.Getenv("AWS_ACCOUNT_ID")
	originalStackName := os.Getenv("STACK_NAME")

	// Set test environment variables
	if err := os.Setenv("AWS_REGION", "us-east-1"); err != nil {
		t.Fatalf("Failed to set AWS_REGION: %v", err)
	}
	if err := os.Setenv("AWS_ACCOUNT_ID", "123456789012"); err != nil {
		t.Fatalf("Failed to set AWS_ACCOUNT_ID: %v", err)
	}
	if err := os.Setenv("STACK_NAME", "test-stack"); err != nil {
		t.Fatalf("Failed to set STACK_NAME: %v", err)
	}

	// Restore original environment variables after test
	defer func() {
		if originalRegion != "" {
			if err := os.Setenv("AWS_REGION", originalRegion); err != nil {
				t.Logf("Failed to restore AWS_REGION: %v", err)
			}
		} else {
			if err := os.Unsetenv("AWS_REGION"); err != nil {
				t.Logf("Failed to unset AWS_REGION: %v", err)
			}
		}
		if originalAccountID != "" {
			if err := os.Setenv("AWS_ACCOUNT_ID", originalAccountID); err != nil {
				t.Logf("Failed to restore AWS_ACCOUNT_ID: %v", err)
			}
		} else {
			if err := os.Unsetenv("AWS_ACCOUNT_ID"); err != nil {
				t.Logf("Failed to unset AWS_ACCOUNT_ID: %v", err)
			}
		}
		if originalStackName != "" {
			if err := os.Setenv("STACK_NAME", originalStackName); err != nil {
				t.Logf("Failed to restore STACK_NAME: %v", err)
			}
		} else {
			if err := os.Unsetenv("STACK_NAME"); err != nil {
				t.Logf("Failed to unset STACK_NAME: %v", err)
			}
		}
	}()

	tests := []struct {
		name     string
		input    string
		expected func(string) bool // Function to validate the result
	}{
		{
			name:  "timestamp placeholder",
			input: "report-$TIMESTAMP.json",
			expected: func(result string) bool {
				return strings.HasPrefix(result, "report-") &&
					strings.HasSuffix(result, ".json") &&
					strings.Contains(result, "T") // Timestamp format contains T
			},
		},
		{
			name:  "AWS region placeholder",
			input: "report-$AWS_REGION.json",
			expected: func(result string) bool {
				return result == "report-us-east-1.json"
			},
		},
		{
			name:  "AWS account ID placeholder",
			input: "report-$AWS_ACCOUNTID.json",
			expected: func(result string) bool {
				return result == "report-123456789012.json"
			},
		},

		{
			name:  "multiple placeholders",
			input: "$AWS_REGION-$AWS_ACCOUNTID.json",
			expected: func(result string) bool {
				return result == "us-east-1-123456789012.json"
			},
		},
		{
			name:  "no placeholders",
			input: "simple-report.json",
			expected: func(result string) bool {
				return result == "simple-report.json"
			},
		},
		{
			name:  "timestamp with other placeholders",
			input: "$TIMESTAMP-$AWS_REGION.json",
			expected: func(result string) bool {
				return strings.Contains(result, "us-east-1.json") &&
					strings.Contains(result, "T") // Timestamp format
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.resolvePlaceholders(tt.input)
			if !tt.expected(result) {
				t.Errorf("resolvePlaceholders() = %v, validation failed for input %v", result, tt.input)
			}
		})
	}
}
