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
		name  string
		input string
		want  func(string) bool // Function to validate the got
	}{
		{
			name:  "timestamp placeholder",
			input: "report-$TIMESTAMP.json",
			want: func(got string) bool {
				return strings.HasPrefix(got, "report-") &&
					strings.HasSuffix(got, ".json") &&
					strings.Contains(got, "T") // Timestamp format contains T
			},
		},
		{
			name:  "AWS region placeholder",
			input: "report-$AWS_REGION.json",
			want: func(got string) bool {
				return got == "report-us-east-1.json"
			},
		},
		{
			name:  "AWS account ID placeholder",
			input: "report-$AWS_ACCOUNTID.json",
			want: func(got string) bool {
				return got == "report-123456789012.json"
			},
		},

		{
			name:  "multiple placeholders",
			input: "$AWS_REGION-$AWS_ACCOUNTID.json",
			want: func(got string) bool {
				return got == "us-east-1-123456789012.json"
			},
		},
		{
			name:  "no placeholders",
			input: "simple-report.json",
			want: func(got string) bool {
				return got == "simple-report.json"
			},
		},
		{
			name:  "timestamp with other placeholders",
			input: "$TIMESTAMP-$AWS_REGION.json",
			want: func(got string) bool {
				return strings.Contains(got, "us-east-1.json") &&
					strings.Contains(got, "T") // Timestamp format
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.resolvePlaceholders(tt.input)
			if !tt.want(got) {
				t.Errorf("resolvePlaceholders() = %v, validation failed for input %v", got, tt.input)
			}
		})
	}
}
