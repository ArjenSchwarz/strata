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

	"github.com/ArjenSchwarz/strata/config"
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

func TestPlanSummaryViperConfigRespected(t *testing.T) {
	// This test verifies the fix for T-368: config values for plan flags
	// must be respected when CLI flags are not explicitly set.
	// Viper with BindPFlag handles precedence: CLI flag > config file > flag default.

	tests := []struct {
		name       string
		viperKey   string
		viperValue interface{}
		expected   interface{}
	}{
		{
			name:       "show-details from config overrides default",
			viperKey:   "plan.show-details",
			viperValue: false,
			expected:   false,
		},
		{
			name:       "highlight-dangers from config overrides default",
			viperKey:   "plan.highlight-dangers",
			viperValue: false,
			expected:   false,
		},
		{
			name:       "show-statistics-summary from config overrides default",
			viperKey:   "plan.show-statistics-summary",
			viperValue: false,
			expected:   false,
		},
		{
			name:       "statistics-summary-format from config overrides default",
			viperKey:   "plan.statistics-summary-format",
			viperValue: "vertical",
			expected:   "vertical",
		},
		{
			name:       "show-no-ops from config overrides default",
			viperKey:   "plan.show-no-ops",
			viperValue: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			// Rebind flags after reset so viper knows about them
			if err := viper.BindPFlag("plan.show-details", planSummaryCmd.Flags().Lookup("details")); err != nil {
				t.Fatal(err)
			}
			if err := viper.BindPFlag("plan.highlight-dangers", planSummaryCmd.Flags().Lookup("highlight-dangers")); err != nil {
				t.Fatal(err)
			}
			if err := viper.BindPFlag("plan.show-statistics-summary", planSummaryCmd.Flags().Lookup("show-statistics")); err != nil {
				t.Fatal(err)
			}
			if err := viper.BindPFlag("plan.statistics-summary-format", planSummaryCmd.Flags().Lookup("stats-format")); err != nil {
				t.Fatal(err)
			}
			if err := viper.BindPFlag("plan.show-no-ops", planSummaryCmd.Flags().Lookup("show-no-ops")); err != nil {
				t.Fatal(err)
			}

			// Simulate config file setting
			viper.Set(tt.viperKey, tt.viperValue)

			// Read value the same way runPlanSummary does
			switch expected := tt.expected.(type) {
			case bool:
				got := viper.GetBool(tt.viperKey)
				if got != expected {
					t.Errorf("viper.GetBool(%q) = %v, want %v", tt.viperKey, got, expected)
				}
			case string:
				got := viper.GetString(tt.viperKey)
				if got != expected {
					t.Errorf("viper.GetString(%q) = %q, want %q", tt.viperKey, got, expected)
				}
			}
		})
	}
}

func TestAlwaysShowSensitiveFlagRegistered(t *testing.T) {
	// Regression test for T-1182: docs/implementation/always-show-sensitive.md
	// documents `--always-show-sensitive=true`, but the flag was never registered.
	flag := planSummaryCmd.Flags().Lookup("always-show-sensitive")
	if flag == nil {
		t.Fatal("always-show-sensitive flag not found")
	}

	// Default must match GetDefaultConfig (true) so the flag does not change
	// behaviour when omitted.
	if flag.DefValue != "true" {
		t.Errorf("Expected default value to be 'true', got %q", flag.DefValue)
	}

	expectedUsage := "Show sensitive resource changes even when details are disabled"
	if flag.Usage != expectedUsage {
		t.Errorf("Expected usage %q, got %q", expectedUsage, flag.Usage)
	}
}

func TestDocumentedFlagsParse(t *testing.T) {
	// Regression test for T-1182: the exact CLI invocations from the docs must
	// parse without an "unknown flag" error.
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "details flag (README CLI example)",
			args: []string{"--details", "terraform.tfplan"},
		},
		{
			name: "always-show-sensitive flag (always-show-sensitive.md example)",
			args: []string{"--details=false", "--always-show-sensitive=true", "terraform.tfplan"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := planSummaryCmd.ParseFlags(tt.args); err != nil {
				t.Errorf("ParseFlags(%v) returned error: %v", tt.args, err)
			}
		})
	}
}

func TestAlwaysShowSensitiveFlagBinding(t *testing.T) {
	// Regression test for T-1182: when the --always-show-sensitive flag is set
	// explicitly, the bound viper key must reflect it so runPlanSummary picks it up.
	defer func() {
		viper.Reset()
		// Restore default flag state for other tests.
		_ = planSummaryCmd.Flags().Set("always-show-sensitive", "true")
	}()

	viper.Reset()
	if err := viper.BindPFlag("plan.always-show-sensitive", planSummaryCmd.Flags().Lookup("always-show-sensitive")); err != nil {
		t.Fatal(err)
	}

	if err := planSummaryCmd.ParseFlags([]string{"--always-show-sensitive=false"}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if !viper.IsSet("plan.always-show-sensitive") {
		t.Fatal("expected plan.always-show-sensitive to be set after explicit flag")
	}
	if viper.GetBool("plan.always-show-sensitive") {
		t.Error("expected plan.always-show-sensitive to be false after --always-show-sensitive=false")
	}
}

func TestAlwaysShowSensitiveDefaultPreserved(t *testing.T) {
	// Verify that when plan.always-show-sensitive is NOT set in viper,
	// the default value (true) from GetDefaultConfig is preserved.
	// This guards against viper.GetBool returning false for unset keys.
	viper.Reset()

	cfg := config.GetDefaultConfig()

	// Simulate the same logic as runPlanSummary
	if viper.IsSet("plan.always-show-sensitive") {
		cfg.Plan.AlwaysShowSensitive = viper.GetBool("plan.always-show-sensitive")
	}

	if !cfg.Plan.AlwaysShowSensitive {
		t.Error("AlwaysShowSensitive should default to true when not set in config")
	}

	// Now verify explicit false overrides the default
	viper.Set("plan.always-show-sensitive", false)
	if viper.IsSet("plan.always-show-sensitive") {
		cfg.Plan.AlwaysShowSensitive = viper.GetBool("plan.always-show-sensitive")
	}

	if cfg.Plan.AlwaysShowSensitive {
		t.Error("AlwaysShowSensitive should be false when explicitly set to false in config")
	}
}
