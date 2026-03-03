package plan

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ArjenSchwarz/strata/config"
)

// Simple file output integration tests that work with the current system

func TestFormatter_FileOutput_ConfigIntegration(t *testing.T) {
	// Test that the config system can create output settings without errors
	cfg := &config.Config{}

	outputConfig := cfg.NewOutputConfiguration()
	if outputConfig == nil {
		t.Errorf("NewOutputConfiguration should not return nil")
		return
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

func TestFormatter_FileOutput_SensitiveOnlyNoSensitiveMessageStillShown(t *testing.T) {
	cfg := &config.Config{
		Plan: config.PlanConfig{
			AlwaysShowSensitive: true,
		},
	}
	formatter := NewFormatter(cfg)

	summary := &PlanSummary{
		PlanFile:         "test.tfplan",
		TerraformVersion: "1.6.0",
		Workspace:        "default",
		Backend: BackendInfo{
			Type:     "local",
			Location: "terraform.tfstate",
		},
		CreatedAt: time.Now(),
		Statistics: ChangeStatistics{
			Total: 1,
			ToAdd: 1,
		},
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_s3_bucket.test",
				Type:          "aws_s3_bucket",
				Name:          "test",
				ChangeType:    ChangeTypeCreate,
				IsDestructive: false,
				IsDangerous:   false,
			},
		},
	}

	outputFile := filepath.Join(t.TempDir(), "summary.md")
	outputConfig := &config.OutputConfiguration{
		Format:           "markdown",
		OutputFile:       outputFile,
		OutputFileFormat: "markdown",
		UseEmoji:         false,
		UseColors:        false,
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		_ = r.Close()
	}()

	err = formatter.OutputSummary(summary, outputConfig, false)
	_ = w.Close()
	if err != nil {
		t.Fatalf("OutputSummary() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	stdoutOutput := buf.String()

	if !strings.Contains(stdoutOutput, "No sensitive resource changes detected.") {
		t.Fatalf("expected stdout to contain no-sensitive message, got: %s", stdoutOutput)
	}

	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("expected output file to be created: %v", err)
	}
}

func TestFormatter_FileOutputColorDecision(t *testing.T) {
	if !shouldUseColorTransformer(true, "table") {
		t.Errorf("expected colors to be enabled for table output when UseColors is true")
	}

	if shouldUseColorTransformer(true, "markdown") {
		t.Errorf("expected colors to be disabled for markdown output even when UseColors is true")
	}

	if shouldUseColorTransformer(false, "table") {
		t.Errorf("expected colors to be disabled when UseColors is false")
	}
}
