package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
)

// Simple file output integration tests that work with the current system

func TestFormatter_FileOutput_ConfigIntegration(t *testing.T) {
	// Test that the config system can create output settings without errors
	cfg := &config.Config{}

	outputConfig := cfg.NewOutputConfiguration()
	if outputConfig == nil {
		t.Errorf("NewOutputConfiguration should not return nil")
	}

	// Test that we can set file output properties
	outputConfig.OutputFile = "test-output.json"
	outputConfig.OutputFileFormat = "json"

	if outputConfig.OutputFile != "test-output.json" {
		t.Errorf("OutputFile should be set correctly")
	}

	if outputConfig.OutputFileFormat != "json" {
		t.Errorf("OutputFileFormat should be set correctly")
	}
}

func TestFormatter_FileOutput_ValidationIntegration(t *testing.T) {
	// Test that the validation system works with the config
	cfg := &config.Config{}
	validator := config.NewFileValidator(cfg)

	if validator == nil {
		t.Errorf("NewFileValidator should not return nil")
	}

	// Test with valid settings
	outputConfig := cfg.NewOutputConfiguration()
	outputConfig.OutputFile = "" // No file output
	outputConfig.OutputFileFormat = "json"

	err := validator.ValidateFileOutput(outputConfig)
	if err != nil {
		t.Errorf("Validation should pass for no file output: %v", err)
	}
}

func TestFormatter_FileOutput_FormatterIntegration(t *testing.T) {
	// Test that the formatter can be created and used without file output
	cfg := &config.Config{}
	formatter := NewFormatter(cfg)

	if formatter == nil {
		t.Errorf("NewFormatter should not return nil")
	}

	// Test that formatter can validate output formats
	err := formatter.ValidateOutputFormat("json")
	if err != nil {
		t.Errorf("ValidateOutputFormat should pass for json: %v", err)
	}

	err = formatter.ValidateOutputFormat("invalid")
	if err == nil {
		t.Errorf("ValidateOutputFormat should fail for invalid format")
	}
}
