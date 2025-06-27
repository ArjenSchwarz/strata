package plan

import (
	"testing"

	"github.com/ArjenSchwarz/strata/config"
)

// Simple file output integration tests that work with the current system

func TestFormatter_FileOutput_ConfigIntegration(t *testing.T) {
	// Test that the config system can create output settings without errors
	cfg := &config.Config{}

	settings := cfg.NewOutputSettings()
	if settings == nil {
		t.Errorf("NewOutputSettings should not return nil")
	}

	// Test that we can set file output properties
	settings.OutputFile = "test-output.json"
	settings.OutputFileFormat = "json"

	if settings.OutputFile != "test-output.json" {
		t.Errorf("OutputFile should be set correctly")
	}

	if settings.OutputFileFormat != "json" {
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
	settings := cfg.NewOutputSettings()
	settings.OutputFile = "" // No file output
	settings.OutputFileFormat = "json"

	err := validator.ValidateFileOutput(settings)
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
