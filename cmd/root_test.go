package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMalformedConfigFile_ReturnsReadError(t *testing.T) {
	// Create a malformed YAML config file
	tmpDir := t.TempDir()
	malformedFile := filepath.Join(tmpDir, "bad.yaml")
	err := os.WriteFile(malformedFile, []byte(":\ninvalid: [yaml\n"), 0644)
	require.NoError(t, err)

	// Simulate what initConfig does when --config is explicitly provided
	v := viper.New()
	v.SetConfigFile(malformedFile)
	err = v.ReadInConfig()

	assert.Error(t, err, "ReadInConfig should return an error for malformed YAML")
	assert.Contains(t, err.Error(), "parsing", "error should mention parsing")
}

func TestValidConfigFile_NoError(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "good.yaml")
	err := os.WriteFile(validFile, []byte("output: json\n"), 0644)
	require.NoError(t, err)

	v := viper.New()
	v.SetConfigFile(validFile)
	err = v.ReadInConfig()

	assert.NoError(t, err, "ReadInConfig should succeed for valid YAML")
	assert.Equal(t, "json", v.GetString("output"))
}

// TestInitConfig_NestedEnvVarOverride is a regression test for T-1258.
//
// Bug: initConfig enabled viper.AutomaticEnv() without an env key replacer, so
// conventional environment variables for nested keys (those containing "." and
// "-") were ignored. As a result PLAN_HIGHLIGHT_DANGERS=false did not override
// plan.highlight-dangers, and only the impractical PLAN.HIGHLIGHT-DANGERS=false
// form worked.
//
// Expected: the underscore form of a nested key overrides the config value.
// Actual (before fix): the underscore form was ignored.
func TestInitConfig_NestedEnvVarOverride(t *testing.T) {
	// Each case sets the env var to "true" while the viper default (with no
	// config value and no bound flag) is the zero value false. The assertion
	// therefore only passes when the env var is actually read through the
	// nested key, which is the behaviour the bug broke.
	tests := []struct {
		name     string
		envKey   string
		envValue string
		viperKey string
		expected bool
	}{
		{
			name:     "PLAN_HIGHLIGHT_DANGERS overrides plan.highlight-dangers",
			envKey:   "PLAN_HIGHLIGHT_DANGERS",
			envValue: "true",
			viperKey: "plan.highlight-dangers",
			expected: true,
		},
		{
			name:     "PLAN_SHOW_DETAILS overrides plan.show-details",
			envKey:   "PLAN_SHOW_DETAILS",
			envValue: "true",
			viperKey: "plan.show-details",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)

			// Use an empty config file so the value can only come from the env var.
			tmpDir := t.TempDir()
			emptyFile := filepath.Join(tmpDir, "empty.yaml")
			require.NoError(t, os.WriteFile(emptyFile, []byte{}, 0644))

			// Point initConfig at the empty config via the package-level cfgFile.
			origCfgFile := cfgFile
			cfgFile = emptyFile
			t.Cleanup(func() { cfgFile = origCfgFile })

			t.Setenv(tt.envKey, tt.envValue)

			// Run the real initialization path used in production.
			initConfig()

			assert.Equal(t, tt.expected, viper.GetBool(tt.viperKey),
				"%s=%s should override %s", tt.envKey, tt.envValue, tt.viperKey)
		})
	}
}
