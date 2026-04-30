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
