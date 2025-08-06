/*
Copyright © 2025 Arjen Schwarz <developer@arjen.eu>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

func TestPlanSummaryFlagParsing(t *testing.T) {
	// Save original values
	originalShowNoOps := showNoOps
	defer func() {
		showNoOps = originalShowNoOps
	}()

	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "default value",
			args:     []string{},
			expected: false,
		},
		{
			name:     "show-no-ops flag set to true",
			args:     []string{"--show-no-ops"},
			expected: true,
		},
		{
			name:     "show-no-ops flag set to false explicitly",
			args:     []string{"--show-no-ops=false"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag to default
			showNoOps = false

			// Parse flags
			planSummaryCmd.ParseFlags(tt.args)

			// Check that the flag variable was set correctly
			if showNoOps != tt.expected {
				t.Errorf("Expected showNoOps to be %v, got %v", tt.expected, showNoOps)
			}
		})
	}
}

func TestPlanSummaryConfigPrecedence(t *testing.T) {
	// Save original values
	originalShowNoOps := showNoOps
	defer func() {
		showNoOps = originalShowNoOps
		viper.Reset()
	}()

	tests := []struct {
		name           string
		configValue    bool
		configSet      bool
		flagArgs       []string
		expectedFlag   bool
		expectedConfig bool
	}{
		{
			name:           "config only - true",
			configValue:    true,
			configSet:      true,
			flagArgs:       []string{},
			expectedFlag:   false, // Flag defaults to false
			expectedConfig: true,  // Config should be true
		},
		{
			name:           "config only - false",
			configValue:    false,
			configSet:      true,
			flagArgs:       []string{},
			expectedFlag:   false, // Flag defaults to false
			expectedConfig: false, // Config should be false
		},
		{
			name:           "flag set to true",
			configValue:    false,
			configSet:      true,
			flagArgs:       []string{"--show-no-ops"},
			expectedFlag:   true,  // Flag should be true
			expectedConfig: false, // Config value stays the same - precedence is handled in application logic
		},
		{
			name:           "flag set to false explicitly",
			configValue:    true,
			configSet:      true,
			flagArgs:       []string{"--show-no-ops=false"},
			expectedFlag:   false, // Flag should be false
			expectedConfig: true,  // Config value stays the same - precedence is handled in application logic
		},
		{
			name:           "no config, flag only",
			configValue:    false,
			configSet:      false,
			flagArgs:       []string{"--show-no-ops"},
			expectedFlag:   true,  // Flag should be true
			expectedConfig: false, // No config set, so should be false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset
			viper.Reset()
			showNoOps = false

			// Set config value if needed
			if tt.configSet {
				viper.Set("plan.show-no-ops", tt.configValue)
			}

			// Parse flags
			err := planSummaryCmd.ParseFlags(tt.flagArgs)
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Check flag variable
			if showNoOps != tt.expectedFlag {
				t.Errorf("Expected flag showNoOps to be %v, got %v", tt.expectedFlag, showNoOps)
			}

			// Check what Viper sees (should be config value, not flag value)
			viperValue := viper.GetBool("plan.show-no-ops")
			if viperValue != tt.expectedConfig {
				t.Errorf("Expected viper config to be %v, got %v", tt.expectedConfig, viperValue)
			}
		})
	}
}

func TestPlanSummaryFlagDefaults(t *testing.T) {
	// Test that the flag has the correct default value
	flag := planSummaryCmd.Flags().Lookup("show-no-ops")
	if flag == nil {
		t.Fatal("show-no-ops flag not found")
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected default value to be 'false', got %q", flag.DefValue)
	}

	// Test flag usage text
	expectedUsage := "Show no-op resources in the summary"
	if flag.Usage != expectedUsage {
		t.Errorf("Expected usage %q, got %q", expectedUsage, flag.Usage)
	}
}
